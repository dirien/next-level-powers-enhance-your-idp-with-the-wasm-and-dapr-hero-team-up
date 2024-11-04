const express = require('express');
const http = require('http');
const WebSocket = require('ws');
const {DaprServer, CommunicationProtocolEnum} = require('@dapr/dapr');

// Server configuration
const daprHost = "127.0.0.1";
const daprPort = process.env.DAPR_HTTP_PORT || "3602"; // Dapr HTTP port
const serverHost = "127.0.0.1";
const serverPort = "6002";
const pubsubName =   process.env.PUBSUB_NAME || "pubsub";
const topicName = process.env.TOPIC_NAME || "dadjokes";

const app = express();
const server = http.createServer(app);
const wss = new WebSocket.Server({server});

// Serve static files from the "public" directory
app.use(express.static('public'));

// Set up WebSocket connection
wss.on('connection', (ws) => {
    console.log('WebSocket client connected');

    // Send a welcome message or any initial data if needed
    ws.send(JSON.stringify({message: 'Connected to WebSocket server'}));

    // Handle WebSocket disconnections
    ws.on('close', () => console.log('WebSocket client disconnected'));
});

// Function to send data to all connected WebSocket clients
function broadcastToClients(data) {
    wss.clients.forEach((client) => {
        if (client.readyState === WebSocket.OPEN) {
            client.send(JSON.stringify(data));
        }
    });
}

// Set up Dapr server for pub/sub
const daprServer = new DaprServer({
    serverHost,
    serverPort,
    communicationProtocol: CommunicationProtocolEnum.HTTP,
    clientOptions: {
        daprHost,
        daprPort,
    },
});

// Subscribe to Dapr pub/sub topic and handle incoming messages
daprServer.pubsub.subscribe(pubsubName, topicName, async (data) => {
    console.log(`Received message from pubsub: ${JSON.stringify(data)}`);

    // Broadcast the received joke and image URL to WebSocket clients
    broadcastToClients(data);
}, "events");

// Start both the HTTP server and Dapr server
server.listen(3000, () => {
    daprServer.start().then(() => {
        console.log(`Dapr pub/sub server started on port ${serverPort}`);
    }).catch((err) => {
        console.error(`Error starting Dapr pub/sub server: ${err}`);
    });
    console.log(`HTTP server is listening on http://localhost:${serverPort}`);
});


