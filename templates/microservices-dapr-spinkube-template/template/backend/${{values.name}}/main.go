package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	_ "golang.org/x/crypto/x509roots/fallback"
)

type DadJokeEvent struct {
	Timestamp string `json:"timestamp"`
}

// ImageRequest Define a struct for the request payload
type ImageRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

// ImageResponse Define a struct for the response data
type ImageResponse struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

// MessagePayload Define a struct for the message payload to post to the local server
type MessagePayload struct {
	Joke     string `json:"joke"`
	ImageURL string `json:"image_url"`
}

var daprClient dapr.Client

func fetchDadJoke() (string, error) {
	req, err := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse the JSON response to get the joke
	var jokeData struct {
		Joke string `json:"joke"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jokeData); err != nil {
		return "", err
	}

	return jokeData.Joke, nil
}
func generateImage(prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("openai_api_key is not set")
	}

	requestBody := ImageRequest{
		Prompt: fmt.Sprintf("generate a picture from this dad joke: \"%s\"", prompt),
		N:      1,
		Model:  "dall-e-3",
		Size:   "1024x1024",
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse the response
	var imageResponse ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&imageResponse); err != nil {
		return "", err
	}

	// Return the URL of the generated image
	if len(imageResponse.Data) > 0 {
		return imageResponse.Data[0].URL, nil
	}
	return "", fmt.Errorf("no image URL returned from OpenAI")
}

var dadjoke = &common.Subscription{
	PubsubName: os.Getenv("PUBSUB_NAME"),
	Topic:      os.Getenv("TOPIC_NAME"),
	Route:      "/events",
}

func main() {

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	r := chi.NewRouter()
	r.HandleFunc("/health/{endpoints:liveness|readiness}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	s := daprd.NewServiceWithMux(fmt.Sprintf(":%s", port), r)

	if err := s.AddTopicEventHandler(dadjoke, eventHandler); err != nil {
		log.Fatalf("error adding topic subscription: %v", err)
	}

	dc, err := dapr.NewClient()
	if err != nil {
		log.Fatalf("order proc: dapr client: %s", err)
	}
	daprClient = dc
	defer daprClient.Close()

	if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("error listenning: %v", err)
	}
}

func eventHandler(ctx context.Context, e *common.TopicEvent) (bool, error) {
	fmt.Printf("DadJoke request received with ID %s\n", e.ID)

	// Step 1: Fetch a dad joke
	joke, err := fetchDadJoke()
	if err != nil {
		fmt.Println("Error fetching dad joke:", err)
	}
	fmt.Println("Dad Joke:", joke)

	// Step 2: Use the joke as a prompt to generate an image
	imageURL, err := generateImage(joke)
	if err != nil {
		fmt.Println("Error generating image:", err)
	}
	fmt.Println("Generated Image URL:", imageURL)

	pubsubComponentName := os.Getenv("PUBSUB_NAME")
	pubsubTopic := os.Getenv("TOPIC_RESULT_NAME")

	payload := MessagePayload{
		Joke:     joke,
		ImageURL: imageURL,
	}
	err = daprClient.PublishEvent(context.Background(), pubsubComponentName, pubsubTopic, payload)
	if err != nil {
		log.Println("Error publishing event:", err)
	}
	return false, nil
}
