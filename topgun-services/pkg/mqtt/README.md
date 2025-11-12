# MQTT File Transfer API

This package provides MQTT integration for the topgun-services, including the ability to send files (like .pt PyTorch model files) through MQTT.

## Configuration

Add to your config file (`configs/dev.yaml` or `configs/uat.yaml`):

```yaml
mqtt:
  broker: "tcp://localhost:1883"
  client_id: "topgun-services"
  topic: "topgun/ai"
```

## API Endpoints

### 1. Check MQTT Status
```bash
GET /api/v1/mqtt/status
```

**Response:**
```json
{
  "connected": true,
  "topic": "topgun/ai"
}
```

### 2. Publish Text Message
```bash
POST /api/v1/mqtt/publish
Content-Type: application/json

{
  "message": "Hello from Go server"
}
```

### 3. Publish JSON Data
```bash
POST /api/v1/mqtt/publish-json
Content-Type: application/json

{
  "detection": "person",
  "confidence": 0.95,
  "timestamp": "2025-11-12T10:30:00Z"
}
```

### 4. Upload and Send File via MQTT
```bash
POST /api/v1/mqtt/upload-file
Content-Type: multipart/form-data

Form Data:
- file: [your .pt file]
- encode_base64: true (optional, default: true)
```

**Example with curl:**
```bash
curl -X POST http://localhost:8080/api/v1/mqtt/upload-file \
  -F "file=@/path/to/model.pt" \
  -F "encode_base64=true"
```

**Response:**
```json
{
  "success": true,
  "message": "File uploaded and published successfully",
  "topic": "topgun/ai",
  "filename": "model.pt",
  "size": 14567890,
  "encoded": true
}
```

### 5. Send File from Server Path
```bash
POST /api/v1/mqtt/send-file
Content-Type: application/json

{
  "file_path": "./submitGeardindaeng2025_sub_1/eiei1.pt",
  "encode_base64": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "File sent successfully",
  "topic": "topgun/ai",
  "filename": "eiei1.pt",
  "size": 14567890,
  "encoded": true
}
```

## File Encoding Options

### Base64 Encoding (encode_base64: true)
- **Pros:** Safe for all MQTT brokers, works with text-based protocols
- **Cons:** 33% size increase
- **Use case:** Smaller files, compatibility
- **Format:** JSON with metadata and base64-encoded data

Example message structure:
```json
{
  "metadata": {
    "type": "file_upload",
    "filename": "model.pt",
    "size": 14567890,
    "ext": ".pt",
    "encoded": true
  },
  "data": "base64_encoded_content_here..."
}
```

### Raw Binary (encode_base64: false)
- **Pros:** No size increase, more efficient
- **Cons:** Requires binary-safe MQTT setup
- **Use case:** Large files, optimized transmission
- **Format:** Metadata sent first (JSON), then raw bytes

## Python Client Example

### Receiving Files via MQTT

```python
import paho.mqtt.client as mqtt
import json
import base64

def on_message(client, userdata, msg):
    try:
        data = json.loads(msg.payload)
        
        if 'metadata' in data and 'data' in data:
            # Base64 encoded file
            metadata = data['metadata']
            filename = metadata['filename']
            
            # Decode and save
            file_content = base64.b64decode(data['data'])
            with open(filename, 'wb') as f:
                f.write(file_content)
            
            print(f"Received file: {filename} ({metadata['size']} bytes)")
    except:
        print(f"Received message: {msg.payload[:100]}")

client = mqtt.Client()
client.on_message = on_message
client.connect("localhost", 1883, 60)
client.subscribe("topgun/ai")
client.loop_forever()
```

### Sending Files from Python

```python
import paho.mqtt.client as mqtt
import json
import base64

def send_pt_file(filename, topic="topgun/ai"):
    with open(filename, 'rb') as f:
        file_content = f.read()
    
    # Encode as base64
    encoded = base64.b64encode(file_content).decode('utf-8')
    
    # Prepare message
    message = {
        "metadata": {
            "type": "file_upload",
            "filename": filename,
            "size": len(file_content),
            "ext": ".pt"
        },
        "data": encoded
    }
    
    client = mqtt.Client()
    client.connect("localhost", 1883, 60)
    client.publish(topic, json.dumps(message))
    client.disconnect()
    
    print(f"Sent {filename} ({len(file_content)} bytes)")

# Usage
send_pt_file("model.pt")
```

## JavaScript/Node.js Client Example

```javascript
const mqtt = require('mqtt');
const fs = require('fs');

const client = mqtt.connect('mqtt://localhost:1883');

client.on('connect', () => {
    console.log('Connected to MQTT');
    
    // Subscribe to receive files
    client.subscribe('topgun/ai', (err) => {
        if (!err) console.log('Subscribed to topgun/ai');
    });
});

client.on('message', (topic, message) => {
    try {
        const data = JSON.parse(message.toString());
        
        if (data.metadata && data.data) {
            const filename = data.metadata.filename;
            const fileBuffer = Buffer.from(data.data, 'base64');
            
            fs.writeFileSync(filename, fileBuffer);
            console.log(`Received file: ${filename} (${data.metadata.size} bytes)`);
        }
    } catch (e) {
        console.log('Received:', message.toString().substring(0, 100));
    }
});

// Send a file
function sendFile(filepath) {
    const fileBuffer = fs.readFileSync(filepath);
    const encoded = fileBuffer.toString('base64');
    
    const message = {
        metadata: {
            type: 'file_upload',
            filename: filepath,
            size: fileBuffer.length,
            ext: '.pt'
        },
        data: encoded
    };
    
    client.publish('topgun/ai', JSON.stringify(message));
    console.log(`Sent ${filepath} (${fileBuffer.length} bytes)`);
}
```

## Testing

### 1. Start MQTT broker
```bash
cd topgun-services
docker-compose up -d mosquitto
```

### 2. Start the server
```bash
make go-run
```

### 3. Test file upload
```bash
# Upload a .pt file
curl -X POST http://localhost:8080/api/v1/mqtt/upload-file \
  -F "file=@./submitGeardindaeng2025_sub_1/eiei1.pt"

# Or send an existing file from server
curl -X POST http://localhost:8080/api/v1/mqtt/send-file \
  -H "Content-Type: application/json" \
  -d '{"file_path": "./submitGeardindaeng2025_sub_1/eiei1.pt", "encode_base64": true}'
```

### 4. Monitor MQTT messages
```bash
docker exec -it topgun-mosquitto mosquitto_sub -t "topgun/ai" -v
```

## Notes

- Maximum file size depends on MQTT broker configuration (default Mosquitto: 10MB)
- For larger files, consider implementing chunking
- Base64 encoding increases payload size by ~33%
- QoS level is set to 1 (at least once delivery)
