package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fermyon/spin-go-sdk/variables"
	"net/http"
	"time"

	spinhttp "github.com/fermyon/spin-go-sdk/http"
)

type DadJokeEvent struct {
	Timestamp string `json:"timestamp"`
}

func postToLocalEndpoint() error {
	payload := DadJokeEvent{
		Timestamp: time.Now().String(),
	}

	// Convert the payload to JSON
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	pubSubName, _ := variables.Get("pubsub_name")
	topicName, _ := variables.Get("topic_name")
	daprPort, _ := variables.Get("dapr_http_port")

	// Send the POST request
	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://localhost:%s/v1.0/publish/%s/%s", daprPort, pubSubName, topicName),
		bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := spinhttp.NewClient()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to post to local endpoint, status: %s", resp.Status)
	}

	fmt.Println("Successfully posted to local endpoint.")
	return nil
}

func NewDadJoke(w http.ResponseWriter, _ *http.Request, ps spinhttp.Params) {
	if err := postToLocalEndpoint(); err != nil {
		fmt.Println("Error posting to local endpoint:", err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Dad Joke Request</title>
        <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
    </head>
    <body class="bg-gray-100 flex items-center justify-center h-screen">
        <div class="bg-white p-8 rounded-lg shadow-md text-center max-w-md">
            <h1 class="text-2xl font-bold text-blue-600">Thank you for requesting a new dad joke.</h1>
        </div>
    </body>
    </html>
`))
}

func init() {
	// call the Handle function
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		router := spinhttp.NewRouter()
		router.GET("/", func(w http.ResponseWriter, r *http.Request, ps spinhttp.Params) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dad Joke Request</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 flex items-center justify-center h-screen">
<div class="bg-white p-8 rounded-lg shadow-md text-center max-w-md">
    <h1 class="text-2xl font-bold text-blue-600">Dad Joke Request API. Head to <a href="/new" class="text-blue-600">/new</a> to request a new dad joke.</h1>
</div>
</body>
</html>`))
		})
		router.GET("/new", NewDadJoke)

		router.ServeHTTP(w, r)
	})
}

func main() {}
