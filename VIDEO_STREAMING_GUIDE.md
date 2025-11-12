# YOLO Video Streaming System

## 📦 Python Dependencies

```bash
pip install websocket-client opencv-python ultralytics
```

Or install all at once:

```bash
pip install websocket-client opencv-python ultralytics paho-mqtt
```

## 🚀 Quick Start

### 1. Start Go Server

```bash
cd topgun-services
make run
# Or: go run cmd/server/main.go
```

Server will be running at `http://localhost:8080`

### 2. Start Frontend (Svelte)

```bash
cd topgun-front
npm install
npm run dev
```

Frontend will be at `http://localhost:5173`

### 3. Stream Video with Python

```bash
cd submitGeardindaeng2025_sub_1
python3 video_stream.py P1_VIDEO_1.mp4
```

#### Options:

```bash
# Display video locally while streaming
python3 video_stream.py P1_VIDEO_1.mp4 --display

# Custom FPS
python3 video_stream.py P1_VIDEO_1.mp4 --fps 20

# Custom JPEG quality (0-100)
python3 video_stream.py P1_VIDEO_1.mp4 --quality 70

# Custom confidence threshold
python3 video_stream.py P1_VIDEO_1.mp4 --conf 0.7

# Custom server URL
python3 video_stream.py P1_VIDEO_1.mp4 --server http://192.168.1.100:8080
```

### 4. View Stream

Open browser: `http://localhost:5173/video-stream`

## 🏗️ Architecture

```
┌─────────────────┐
│ Python (YOLO)   │
│ video_stream.py │
│ - Read video    │
│ - Detect objects│
│ - Encode JPEG   │
└────────┬────────┘
         │ WebSocket
         │ /ws/video-input
         ▼
┌─────────────────────────┐
│ Go Server (Fiber)       │
│ - Receive frames        │
│ - Broadcast to clients  │
│ pkg/detect/websocket.go │
└────────┬────────────────┘
         │ WebSocket
         │ /ws/video-stream
         ▼
┌─────────────────────┐
│ Svelte Frontend     │
│ VideoStream.svelte  │
│ - Display frames    │
│ - Show stats        │
│ - Auto-reconnect    │
└─────────────────────┘
```

## 📡 WebSocket Endpoints

### 1. `/ws/video-input` - Python to Go Server

**Purpose:** Receive video frames from Python

**Message Format:**
```json
{
  "frame": "base64_encoded_jpeg",
  "timestamp": 1699876543.123,
  "frame_number": 1234,
  "detections": 5,
  "width": 1920,
  "height": 1080,
  "model": "model_20241112_150000.pt"
}
```

### 2. `/ws/video-stream` - Go Server to Clients

**Purpose:** Broadcast video frames to all viewers

**Same message format as above**

## 🎯 Features

### Python Side (`video_stream.py`)
- ✅ YOLO object detection
- ✅ Annotated frames with bounding boxes
- ✅ Configurable FPS, quality, confidence
- ✅ Auto-reconnect to server
- ✅ Real-time stats logging
- ✅ Optional local display

### Go Server (`pkg/detect/websocket.go`)
- ✅ Receive frames from Python
- ✅ Broadcast to multiple clients
- ✅ Connection management
- ✅ Non-blocking broadcast (drops frames if client slow)
- ✅ Automatic cleanup

### Svelte Frontend (`VideoStream.svelte`)
- ✅ Real-time video display
- ✅ Stats overlay (FPS, detections, model)
- ✅ Connection status indicator
- ✅ Auto-reconnect with exponential backoff
- ✅ Responsive design

## 🔧 Configuration

### Python
Edit `video_stream.py` or use command-line args:
```python
GO_SERVER_URL = "http://localhost:8080"
TARGET_FPS = 30
JPEG_QUALITY = 80
CONF_THRESHOLD = 0.6
```

### Go Server
No additional config needed. WebSocket routes are auto-registered.

### Svelte
Edit `+page.svelte`:
```typescript
let serverUrl = 'ws://localhost:8080';
```

## 📊 Performance Tips

1. **Reduce FPS**: Lower `--fps` for less CPU usage
2. **Lower Quality**: Use `--quality 60-70` for smaller frames
3. **Higher Confidence**: Use `--conf 0.7` to reduce false positives
4. **Network**: Use wired connection for best results

## 🐛 Troubleshooting

### "Failed to connect to MQTT broker"
This is for model updates, not video streaming. Video streaming uses WebSocket, not MQTT.

### "Connection error" in browser
- Make sure Go server is running
- Check server URL in Svelte component
- Check browser console for errors

### Low FPS / Laggy
- Reduce `--fps` value
- Lower `--quality` value
- Check network bandwidth
- Check CPU usage

### "No model found"
Place a YOLO model file in `submitGeardindaeng2025_sub_1/models/` folder.

## 📝 Example Workflow

```bash
# Terminal 1: Start Go Server
cd topgun-services && make run

# Terminal 2: Start Svelte Frontend
cd topgun-front && npm run dev

# Terminal 3: Stream Video
cd submitGeardindaeng2025_sub_1
python3 video_stream.py P1_VIDEO_1.mp4 --fps 25 --quality 75

# Open browser: http://localhost:5173/video-stream
```

## 🎬 Demo

The video stream page shows:
- Live video with YOLO detections (bounding boxes)
- Real-time FPS counter
- Number of objects detected
- Model name being used
- Connection status
- Frame number

All in real-time! 🚀
