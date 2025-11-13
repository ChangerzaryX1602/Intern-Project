# MQTT Detection Integration

## Overview
ระบบ MQTT Detection Integration รับข้อมูลการตรวจจับจาก Raspberry PI ผ่าน MQTT และบันทึกลงฐานข้อมูลพร้อมแคปรูปจาก video stream

## Architecture

```
Raspberry PI → MQTT Broker (topic: topgun/ai) → Go Server → Database + File Storage
                                                    ↓
                                              Video Stream → Frame Cache → Capture
```

## Data Flow

1. **Raspberry PI** ส่งข้อมูลการตรวจจับผ่าน MQTT topic `topgun/ai`:
```json
{
  "x": 0.28023433685302734,
  "y": 0.7684724926948547,
  "w": 0.21453095972537994,
  "h": 0.46305423974990845,
  "lat": 14.30436542994377,
  "lon": 101.17195665882147,
  "alt": 43.07347682209543,
  "confidence": 1.0,
  "track_id": 256,
  "timestamp": 1762984799.0192025
}
```

2. **MQTT Handler** รับข้อมูลและ:
   - แคปรูปจาก video stream cache ล่าสุด
   - บันทึกรูปลง `./upload/mqtt_capture_<timestamp>_track_<track_id>.jpg`
   - บันทึกข้อมูลลงฐานข้อมูล (ตาราง `detects`)
   - Broadcast ไปยัง WebSocket clients

3. **Video Stream** อัพเดท frame cache:
   - ทุกครั้งที่ broadcast video frame
   - เก็บ frame ล่าสุดไว้ใน memory (base64 decoded)
   - พร้อมให้ MQTT handler แคปภาพได้ทันที

## Components

### 1. Video Frame Cache (`websocket.go`)
- `VideoFrameCache` - struct สำหรับเก็บ video frame ล่าสุด
- `UpdateVideoFrameCache()` - อัพเดท cache เมื่อมี frame ใหม่
- `GetLatestVideoFrame()` - ดึง frame ล่าสุดสำหรับแคปรูป

### 2. MQTT Handler (`mqtt_handler.go`)
- `MQTTDetectHandler` - handler สำหรับประมวลผล MQTT messages
- `HandleMessage()` - ประมวลผลข้อมูลจาก Raspberry PI
- `saveFrameToFile()` - บันทึกรูปลง upload directory
- `StartMQTTSubscription()` - เริ่ม MQTT subscription

### 3. Data Models
- `RaspberryPIDetection` - struct สำหรับข้อมูลจาก Raspberry PI
- `Detect.Objects` - JSONB array เก็บข้อมูลการตรวจจับ

## Configuration

ตั้งค่าใน `configs/dev.yaml`:

```yaml
mqtt:
  broker: "tcp://localhost:1883"
  client_id: "topgun-services"
  topic: "topgun/ai"                    # Topic สำหรับรับข้อมูล
  detect_topic: "topgun/ai"             # Optional: Override detect topic
  camera_id: "00000000-0000-0000-0000-000000000001"  # Optional: Camera UUID
```

## File Storage

รูปที่แคปจะถูกบันทึกที่:
- Path: `./upload/mqtt_capture_<timestamp>_track_<track_id>.jpg`
- Format: JPEG
- Naming: `mqtt_capture_20060102_150405_track_256.jpg`

## Database Schema

```sql
CREATE TABLE detects (
  id SERIAL PRIMARY KEY,
  camera_id UUID NOT NULL,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  path VARCHAR(255),                    -- Path to captured image
  objects JSONB                         -- Detection data array
);
```

Example `objects` field:
```json
[{
  "x": 0.280,
  "y": 0.768,
  "w": 0.214,
  "h": 0.463,
  "lat": 14.304,
  "lon": 101.172,
  "alt": 43.073,
  "confidence": 1.0,
  "track_id": 256,
  "timestamp": 1762984799.0192025
}]
```

## Testing

### 1. Start MQTT Broker
```bash
cd topgun-services
docker-compose up mosquitto
```

### 2. Start Go Server
```bash
make go-run
# หรือ
cd cmd/server && go run main.go
```

### 3. Start Video Stream (Python)
```bash
cd submitGeardindaeng2025_sub_1
python3 stream_with_mqtt.py
```

### 4. Publish Test MQTT Message
```bash
mosquitto_pub -h localhost -t "topgun/ai" -m '{
  "x": 0.5,
  "y": 0.5,
  "w": 0.2,
  "h": 0.3,
  "lat": 14.304,
  "lon": 101.172,
  "alt": 43.0,
  "confidence": 0.95,
  "track_id": 100,
  "timestamp": 1762984799.0
}'
```

## Logs

ตัวอย่าง logs เมื่อระบบทำงาน:

```
[INFO] Successfully subscribed to MQTT topic: topgun/ai
[INFO] Received MQTT message on topic topgun/ai
[INFO] RaspberryPI Detection: TrackID=256, Lat=14.304365, Lon=101.171957, Alt=43.07, Confidence=1.00
[INFO] Saved captured frame to: ./upload/mqtt_capture_20251113_143052_track_256.jpg
[INFO] Successfully saved detection ID=42 with 1 objects to database
[INFO] Broadcasted detection to WebSocket clients
```

## Error Handling

- หาก video frame ไม่พร้อม: บันทึก detection โดยไม่มีรูป (path = "")
- หากบันทึกรูปล้มเหลว: log error แต่ยังบันทึกข้อมูลลง database
- หาก MQTT connection ขาด: auto reconnect (5-30 วินาที)

## WebSocket Broadcasting

หลังบันทึกสำเร็จ จะ broadcast ไปยัง WebSocket clients ที่ subscribe camera_id นั้น:

```json
{
  "id": 42,
  "camera_id": "00000000-0000-0000-0000-000000000001",
  "timestamp": "2025-11-13T14:30:52+07:00",
  "path": "./upload/mqtt_capture_20251113_143052_track_256.jpg",
  "objects": [{...}],
  "image_data": "base64_encoded_image...",
  "mime_type": "image/jpeg"
}
```

## Performance

- Video frame cache ใช้ sync.RWMutex เพื่อ thread-safety
- MQTT handler ทำงานแบบ asynchronous
- Database write เป็น blocking operation (รอจนบันทึกสำเร็จ)
- File write ใช้ buffered I/O

## Security

- MQTT broker ควรใช้ authentication
- Upload directory ควรจำกัด permissions (0755)
- File permissions: 0644 (read for all, write for owner only)
