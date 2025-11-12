#!/usr/bin/env python3
"""
Unified YOLO Detection System
- Real-time video streaming with YOLO detection
- Auto-update model via MQTT (hot-reload without restart)
- WebSocket streaming to Go server
- Optimized for Raspberry Pi (targeting 5+ FPS)
"""

import os
import cv2
import base64
import json
import time
import argparse
import threading
from datetime import datetime
from pathlib import Path
from queue import Queue, Empty

import websocket
import paho.mqtt.client as mqtt
from ultralytics import YOLO
from model_utils import load_latest_model, get_latest_model

# ==================== Configuration ====================
GO_SERVER_URL = "ws://192.168.8.185:8080"
WEBSOCKET_PATH = "/ws/video-input"
TARGET_FPS = 5
JPEG_QUALITY = 70
CONF_THRESHOLD = 0.6
IOU_THRESHOLD = 0.4
TARGET_WIDTH = 640  # Changed from 854 to 640 (faster YOLO processing)
TARGET_HEIGHT = 480
YOLO_IMG_SIZE = 640  # YOLO inference size (lower = faster)
MQTT_BROKER = "192.168.8.185"
MQTT_PORT = 1883
MQTT_TOPIC = "topgun/ai"

# ==================== Global Variables ====================
model = None
current_model_path = None
ws = None
mqtt_client = None
is_ws_connected = False
is_running = True
model_reload_lock = threading.Lock()
frame_buffer = Queue(maxsize=2)  # Small buffer to prevent lag

# Performance stats
stats = {
    'frame_count': 0,
    'total_detections': 0,
    'start_time': None,
    'last_print_time': None,
    'dropped_frames': 0,
    'total_inference_time': 0,
    'total_encode_time': 0,
}

# ==================== MQTT Callbacks ====================
def on_connect_mqtt(client, userdata, flags, rc):
    """MQTT connection callback"""
    if rc == 0:
        print(f"✅ MQTT Connected")
        client.subscribe(MQTT_TOPIC)
        print(f"📡 Subscribed to: {MQTT_TOPIC}")
    else:
        print(f"❌ MQTT connection failed: {rc}")

def on_message_mqtt(client, userdata, msg):
    """MQTT message callback - hot reload model"""
    global current_model_path, model
    
    try:
        data = json.loads(msg.payload)
        
        if isinstance(data, dict) and 'metadata' in data and 'data' in data:
            metadata = data['metadata']
            original_filename = metadata.get('filename', 'model.pt')
            file_size = metadata.get('size', 0)
            
            print(f"\n{'='*60}")
            print(f"📥 New model received: {original_filename} ({file_size/1024/1024:.2f} MB)")
            
            # Decode and save
            file_content = base64.b64decode(data['data'])
            os.makedirs('models', exist_ok=True)
            
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            file_ext = Path(original_filename).suffix
            timestamped_filename = f"model_{timestamp}{file_ext}"
            file_path = os.path.join('models', timestamped_filename)
            
            with open(file_path, 'wb') as f:
                f.write(file_content)
            
            print(f"💾 Saved as: {timestamped_filename}")
            
            # Hot reload model
            if file_ext in ['.pt', '.pth']:
                try:
                    print(f"🔄 Hot-reloading model...")
                    
                    # Load new model
                    new_model = YOLO(file_path)
                    
                    # Warm up
                    dummy = new_model(cv2.imread(file_path) if os.path.exists(file_path) else None, verbose=False)
                    
                    # Atomic swap with lock
                    with model_reload_lock:
                        old_path = current_model_path
                        model = new_model
                        current_model_path = file_path
                    
                    print(f"✅ Model updated! {os.path.basename(old_path)} → {timestamped_filename}")
                    print(f"{'='*60}\n")
                except Exception as e:
                    print(f"❌ Model load failed: {e}")
                    print(f"{'='*60}\n")
        
    except Exception as e:
        print(f"⚠️  MQTT message error: {e}")

def on_disconnect_mqtt(client, userdata, rc):
    """MQTT disconnect callback"""
    if rc != 0:
        print(f"⚠️  MQTT disconnected, auto-reconnecting...")

# ==================== WebSocket Callbacks ====================
def on_open_ws(ws_conn):
    """WebSocket connection callback"""
    global is_ws_connected
    is_ws_connected = True
    print(f"✅ WebSocket Connected: {GO_SERVER_URL}{WEBSOCKET_PATH}")

def on_close_ws(ws_conn, close_status_code, close_msg):
    """WebSocket disconnect callback"""
    global is_ws_connected
    is_ws_connected = False
    print(f"⚠️  WebSocket Disconnected")

def on_error_ws(ws_conn, error):
    """WebSocket error callback"""
    print(f"❌ WebSocket Error: {error}")

def on_message_ws(ws_conn, message):
    """WebSocket message callback"""
    try:
        data = json.loads(message)
        if 'status' in data:
            print(f"📊 Server: {data}")
    except:
        pass

# ==================== Model Initialization ====================
def initialize_model():
    """Initialize YOLO model"""
    global model, current_model_path
    
    print("="*60)
    print("🚀 Initializing YOLO Model")
    print("="*60)
    
    model = load_latest_model()
    
    if model is None:
        # Fallback to default
        default_path = "models/best_drone_detector.pt"
        if os.path.exists(default_path):
            print(f"📝 Using default: {default_path}")
            model = YOLO(default_path)
            current_model_path = default_path
        else:
            print(f"❌ No model found!")
            return False
    else:
        current_model_path = get_latest_model()
        print(f"✅ Loaded: {os.path.basename(current_model_path)}")
    
    # Warm up model
    print(f"🔥 Warming up model...")
    dummy_frame = cv2.imread(current_model_path) if os.path.exists(current_model_path) else None
    _ = model(dummy_frame, verbose=False)
    
    print("="*60 + "\n")
    return True

# ==================== MQTT Setup ====================
def setup_mqtt():
    """Setup MQTT client"""
    global mqtt_client
    
    print("="*60)
    print("📡 Setting up MQTT")
    print("="*60)
    
    mqtt_client = mqtt.Client(client_id="unified-yolo-stream", clean_session=True)
    mqtt_client.on_connect = on_connect_mqtt
    mqtt_client.on_message = on_message_mqtt
    mqtt_client.on_disconnect = on_disconnect_mqtt
    mqtt_client.reconnect_delay_set(min_delay=1, max_delay=120)
    
    try:
        print(f"🔌 Connecting to {MQTT_BROKER}:{MQTT_PORT}...")
        mqtt_client.connect(MQTT_BROKER, MQTT_PORT, keepalive=60)
        mqtt_client.loop_start()
        time.sleep(1)
        print(f"✅ MQTT setup complete")
        print("="*60 + "\n")
        return True
    except Exception as e:
        print(f"⚠️  MQTT failed: {e}")
        print(f"💡 Continuing without MQTT...")
        print("="*60 + "\n")
        return False

# ==================== WebSocket Setup ====================
def setup_websocket():
    """Setup WebSocket connection"""
    global ws, is_ws_connected
    
    ws_url = f"{GO_SERVER_URL}{WEBSOCKET_PATH}"
    print(f"🔌 Connecting to WebSocket: {ws_url}...")
    
    try:
        ws = websocket.WebSocketApp(
            ws_url,
            on_open=on_open_ws,
            on_message=on_message_ws,
            on_error=on_error_ws,
            on_close=on_close_ws
        )
        
        ws_thread = threading.Thread(target=ws.run_forever, daemon=True)
        ws_thread.start()
        
        time.sleep(2)
        
        if not is_ws_connected:
            print(f"❌ WebSocket connection failed")
            print(f"💡 Make sure Go server is running at {GO_SERVER_URL}")
            return False
        
        return True
    except Exception as e:
        print(f"❌ WebSocket setup failed: {e}")
        return False

# ==================== Video Processing (Optimized) ====================
def process_video_optimized(video_path, display=False, skip_frames=0):
    """
    Process video with optimized YOLO detection
    
    Optimizations:
    1. Resize before YOLO (640x480 instead of 854x480)
    2. Smaller YOLO imgsz (640 instead of 1920)
    3. Skip frames if processing too slow
    4. Reuse buffers
    5. Thread-safe model reload
    
    Args:
        video_path: Path to video file
        display: Show local window
        skip_frames: Skip N frames if falling behind
    """
    global model, is_running, is_ws_connected, stats
    
    if not os.path.exists(video_path):
        print(f"❌ Video not found: {video_path}")
        return
    
    print("="*60)
    print("🎥 Starting Optimized Video Stream")
    print("="*60)
    print(f"📁 Video: {video_path}")
    print(f"🎯 Model: {os.path.basename(current_model_path)}")
    print(f"📊 Target: {TARGET_FPS} FPS @ {TARGET_WIDTH}x{TARGET_HEIGHT}")
    print(f"🖼️  JPEG Quality: {JPEG_QUALITY}%")
    print(f"🔍 YOLO imgsz: {YOLO_IMG_SIZE} (smaller = faster)")
    print(f"🚀 Optimizations: Enabled")
    print(f"🛑 Press Ctrl+C to stop")
    print("="*60 + "\n")
    
    cap = cv2.VideoCapture(video_path)
    
    if not cap.isOpened():
        print(f"❌ Failed to open video")
        return
    
    # Video info
    fps = cap.get(cv2.CAP_PROP_FPS)
    width = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
    height = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    
    print(f"📺 Original: {width}x{height} @ {fps:.1f} FPS ({total_frames} frames)\n")
    
    # Reset stats
    stats = {
        'frame_count': 0,
        'total_detections': 0,
        'start_time': time.time(),
        'last_print_time': time.time(),
        'dropped_frames': 0,
        'total_inference_time': 0,
        'total_encode_time': 0,
    }
    
    frame_interval = 1.0 / TARGET_FPS
    frames_to_skip = 0
    
    # Pre-allocate buffer for resize (optimization)
    resize_buffer = None
    
    try:
        while is_running:
            loop_start = time.time()
            
            # Read frame
            ret, frame = cap.read()
            if not ret:
                print("\n📹 End of video")
                
                # Loop video
                cap.set(cv2.CAP_PROP_POS_FRAMES, 0)
                print("🔄 Looping video...")
                continue
            
            # Skip frames if falling behind
            if frames_to_skip > 0:
                frames_to_skip -= 1
                stats['dropped_frames'] += 1
                continue
            
            stats['frame_count'] += 1
            
            # Resize frame (OPTIMIZATION: smaller = faster YOLO)
            resize_start = time.time()
            frame_resized = cv2.resize(frame, (TARGET_WIDTH, TARGET_HEIGHT))
            
            # YOLO detection with thread-safe model access
            inference_start = time.time()
            with model_reload_lock:
                results = model(
                    frame_resized, 
                    conf=CONF_THRESHOLD, 
                    iou=IOU_THRESHOLD, 
                    imgsz=YOLO_IMG_SIZE,  # OPTIMIZATION: 640 is much faster than 1920
                    verbose=False,
                    half=False  # Set to True if Raspberry Pi supports FP16
                )
            inference_time = time.time() - inference_start
            stats['total_inference_time'] += inference_time
            
            # Get annotated frame
            annotated_frame = results[0].plot()
            
            # Count detections
            detections = len(results[0].boxes) if results[0].boxes is not None else 0
            stats['total_detections'] += detections
            
            # Encode to JPEG
            encode_start = time.time()
            _, buffer = cv2.imencode('.jpg', annotated_frame, 
                                    [cv2.IMWRITE_JPEG_QUALITY, JPEG_QUALITY])
            frame_base64 = base64.b64encode(buffer).decode('utf-8')
            encode_time = time.time() - encode_start
            stats['total_encode_time'] += encode_time
            
            # Prepare data
            data = {
                'frame': frame_base64,
                'timestamp': time.time(),
                'frame_number': stats['frame_count'],
                'detections': detections,
                'width': TARGET_WIDTH,
                'height': TARGET_HEIGHT,
                'model': os.path.basename(current_model_path)
            }
            
            # Send via WebSocket (non-blocking)
            if is_ws_connected:
                try:
                    ws.send(json.dumps(data))
                except Exception as e:
                    if stats['frame_count'] % 30 == 0:  # Print occasionally
                        print(f"⚠️  Send failed: {e}")
            
            # Display locally
            if display:
                cv2.imshow('YOLO Stream', annotated_frame)
                if cv2.waitKey(1) & 0xFF == ord('q'):
                    print("\n🛑 Stopped by user")
                    break
            
            # Print stats every 2 seconds
            current_time = time.time()
            if current_time - stats['last_print_time'] >= 2.0:
                elapsed = current_time - stats['start_time']
                actual_fps = stats['frame_count'] / elapsed
                avg_detections = stats['total_detections'] / stats['frame_count']
                avg_inference = (stats['total_inference_time'] / stats['frame_count']) * 1000
                avg_encode = (stats['total_encode_time'] / stats['frame_count']) * 1000
                
                print(f"📊 Frame: {stats['frame_count']:5d} | "
                      f"FPS: {actual_fps:4.1f} | "
                      f"Det: {detections:2d} (avg:{avg_detections:4.1f}) | "
                      f"Inf: {inference_time*1000:5.1f}ms (avg:{avg_inference:5.1f}ms) | "
                      f"Enc: {encode_time*1000:4.1f}ms | "
                      f"Drop: {stats['dropped_frames']:4d}")
                
                stats['last_print_time'] = current_time
            
            # Frame rate control with adaptive skip
            loop_time = time.time() - loop_start
            
            if loop_time > frame_interval * 1.5:
                # Falling behind, skip next frame
                frames_to_skip = 1
            
            sleep_time = max(0, frame_interval - loop_time)
            if sleep_time > 0:
                time.sleep(sleep_time)
    
    except KeyboardInterrupt:
        print("\n\n🛑 Stopped by user (Ctrl+C)")
        is_running = False
    
    finally:
        cap.release()
        if display:
            cv2.destroyAllWindows()
        
        # Final stats
        elapsed = time.time() - stats['start_time']
        avg_fps = stats['frame_count'] / elapsed if elapsed > 0 else 0
        avg_detections = stats['total_detections'] / stats['frame_count'] if stats['frame_count'] > 0 else 0
        avg_inference = (stats['total_inference_time'] / stats['frame_count']) * 1000 if stats['frame_count'] > 0 else 0
        avg_encode = (stats['total_encode_time'] / stats['frame_count']) * 1000 if stats['frame_count'] > 0 else 0
        
        print("\n" + "="*60)
        print("📊 Final Statistics")
        print("="*60)
        print(f"⏱️  Duration: {elapsed:.2f}s")
        print(f"🎞️  Frames: {stats['frame_count']} (dropped: {stats['dropped_frames']})")
        print(f"📈 Average FPS: {avg_fps:.2f}")
        print(f"🎯 Total detections: {stats['total_detections']}")
        print(f"📊 Avg detections/frame: {avg_detections:.2f}")
        print(f"⚡ Avg inference time: {avg_inference:.1f}ms")
        print(f"🖼️  Avg encode time: {avg_encode:.1f}ms")
        print(f"💾 Model: {os.path.basename(current_model_path)}")
        print("="*60)

# ==================== Main ====================
def main():
    """Main entry point"""
    global GO_SERVER_URL, TARGET_FPS, JPEG_QUALITY, CONF_THRESHOLD
    global TARGET_WIDTH, TARGET_HEIGHT, YOLO_IMG_SIZE, MQTT_BROKER, is_running
    
    parser = argparse.ArgumentParser(
        description='Unified YOLO Detection System with MQTT model update and WebSocket streaming',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Basic usage with default settings (optimized for Raspberry Pi)
  python3 unified_stream.py video.mp4
  
  # With local display
  python3 unified_stream.py video.mp4 --display
  
  # Custom FPS and quality
  python3 unified_stream.py video.mp4 --fps 3 --quality 60
  
  # Custom server
  python3 unified_stream.py video.mp4 --server ws://192.168.1.100:8080
  
  # Ultra-fast mode (lower quality, higher FPS)
  python3 unified_stream.py video.mp4 --fps 8 --quality 50 --imgsz 416
  
  # High quality mode (slower)
  python3 unified_stream.py video.mp4 --fps 3 --quality 85 --imgsz 640
        """
    )
    
    parser.add_argument('video', type=str, help='Path to video file')
    parser.add_argument('--display', action='store_true', help='Display video locally')
    parser.add_argument('--server', type=str, default=GO_SERVER_URL,
                       help=f'Go server URL (default: {GO_SERVER_URL})')
    parser.add_argument('--fps', type=int, default=TARGET_FPS,
                       help=f'Target FPS (default: {TARGET_FPS})')
    parser.add_argument('--quality', type=int, default=JPEG_QUALITY,
                       help=f'JPEG quality 0-100 (default: {JPEG_QUALITY})')
    parser.add_argument('--conf', type=float, default=CONF_THRESHOLD,
                       help=f'Confidence threshold (default: {CONF_THRESHOLD})')
    parser.add_argument('--width', type=int, default=TARGET_WIDTH,
                       help=f'Target width (default: {TARGET_WIDTH})')
    parser.add_argument('--height', type=int, default=TARGET_HEIGHT,
                       help=f'Target height (default: {TARGET_HEIGHT})')
    parser.add_argument('--imgsz', type=int, default=YOLO_IMG_SIZE,
                       help=f'YOLO inference size (default: {YOLO_IMG_SIZE}, lower=faster)')
    parser.add_argument('--mqtt-broker', type=str, default=MQTT_BROKER,
                       help=f'MQTT broker address (default: {MQTT_BROKER})')
    parser.add_argument('--no-mqtt', action='store_true',
                       help='Disable MQTT model updates')
    
    args = parser.parse_args()
    
    # Update config
    GO_SERVER_URL = args.server
    TARGET_FPS = args.fps
    JPEG_QUALITY = args.quality
    CONF_THRESHOLD = args.conf
    TARGET_WIDTH = args.width
    TARGET_HEIGHT = args.height
    YOLO_IMG_SIZE = args.imgsz
    MQTT_BROKER = args.mqtt_broker
    
    print("\n")
    print("="*60)
    print("🚀 Unified YOLO Detection System")
    print("="*60)
    print(f"📅 Started: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"🎯 Features: Video Stream + MQTT Model Update")
    print(f"⚡ Optimized for: Raspberry Pi")
    print("="*60 + "\n")
    
    # 1. Initialize model
    if not initialize_model():
        print("❌ Model initialization failed")
        return 1
    
    # 2. Setup MQTT (optional)
    if not args.no_mqtt:
        setup_mqtt()
    else:
        print("⚠️  MQTT disabled by --no-mqtt flag\n")
    
    # 3. Setup WebSocket
    if not setup_websocket():
        print("❌ WebSocket setup failed")
        if mqtt_client:
            mqtt_client.loop_stop()
            mqtt_client.disconnect()
        return 1
    
    # 4. Process video
    try:
        process_video_optimized(args.video, display=args.display)
    finally:
        # Cleanup
        is_running = False
        
        if mqtt_client:
            print("\n🔌 Disconnecting MQTT...")
            mqtt_client.loop_stop()
            mqtt_client.disconnect()
        
        if ws and is_ws_connected:
            print("🔌 Disconnecting WebSocket...")
            ws.close()
        
        print("\n")
        print("="*60)
        print("👋 Program terminated")
        print(f"📅 Ended: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("="*60)
    
    return 0

if __name__ == "__main__":
    exit(main())
