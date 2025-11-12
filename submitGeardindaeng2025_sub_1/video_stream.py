#!/usr/bin/env python3
"""
Real-time Video Detection with YOLO and WebSocket Streaming
Sends annotated frames to Go server via WebSocket
"""

import os
import cv2
import base64
import json
import time
import argparse
from datetime import datetime
from pathlib import Path
import threading

import websocket
from ultralytics import YOLO
from model_utils import load_latest_model, get_latest_model

# ==================== Configuration ====================
GO_SERVER_URL = "ws://192.168.8.201:8080"
WEBSOCKET_PATH = "/ws/video-input"
TARGET_FPS = 5  # Reduced to 5 FPS
JPEG_QUALITY = 70  # Reduced quality for smaller size
CONF_THRESHOLD = 0.6
IOU_THRESHOLD = 0.4
TARGET_WIDTH = 854  # 480p width
TARGET_HEIGHT = 480  # 480p height

# ==================== Global Variables ====================
model = None
current_model_path = None
ws = None
frame_count = 0
start_time = None
is_connected = False

# ==================== WebSocket Events ====================
def on_open(ws_conn):
    """Callback when connected to server"""
    global is_connected
    is_connected = True
    print(f"✅ Connected to Go server: {GO_SERVER_URL}{WEBSOCKET_PATH}")

def on_close(ws_conn, close_status_code, close_msg):
    """Callback when disconnected from server"""
    global is_connected
    is_connected = False
    print(f"⚠️  Disconnected from server")

def on_error(ws_conn, error):
    """Callback on connection error"""
    print(f"❌ Connection error: {error}")

def on_message(ws_conn, message):
    """Callback when receiving message from server"""
    try:
        data = json.loads(message)
        if 'status' in data:
            print(f"📊 Server: {data}")
    except:
        pass

# ==================== Model Functions ====================
def initialize_model():
    """Initialize YOLO model"""
    global model, current_model_path
    
    print("="*60)
    print("🚀 Initializing YOLO Model")
    print("="*60)
    
    model = load_latest_model()
    
    if model is None:
        # Try default model
        default_path = "best_drone_detector.pt"
        if os.path.exists(default_path):
            print(f"📝 Using default model: {default_path}")
            model = YOLO(default_path)
            current_model_path = default_path
        else:
            print(f"❌ No model found!")
            return False
    else:
        current_model_path = get_latest_model()
        print(f"✅ Loaded: {os.path.basename(current_model_path)}")
    
    print("="*60 + "\n")
    return True

# ==================== Video Processing ====================
def process_video(video_path, display=False):
    """
    Process video with YOLO detection and stream to server
    
    Args:
        video_path: Path to video file
        display: Show video window locally
    """
    global model, frame_count, start_time, is_connected
    
    if not os.path.exists(video_path):
        print(f"❌ Video file not found: {video_path}")
        return
    
    print("="*60)
    print("🎥 Starting Video Stream")
    print("="*60)
    print(f"📁 Video: {video_path}")
    print(f"🎯 Model: {os.path.basename(current_model_path)}")
    print(f"📊 Target FPS: {TARGET_FPS}")
    print(f"🖼️  JPEG Quality: {JPEG_QUALITY}%")
    print(f"🔍 Confidence: {CONF_THRESHOLD}")
    print(f"🛑 Press 'q' to quit")
    print("="*60 + "\n")
    
    # Open video
    cap = cv2.VideoCapture(video_path)
    
    if not cap.isOpened():
        print(f"❌ Failed to open video: {video_path}")
        return
    
    # Get video properties
    fps = cap.get(cv2.CAP_PROP_FPS)
    width = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
    height = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    
    print(f"📺 Video info:")
    print(f"   Original Resolution: {width}x{height}")
    print(f"   Target Resolution: {TARGET_WIDTH}x{TARGET_HEIGHT} (480p)")
    print(f"   FPS: {fps:.2f}")
    print(f"   Total frames: {total_frames}")
    print(f"   Duration: {total_frames/fps:.2f}s\n")
    
    frame_count = 0
    start_time = time.time()
    frame_interval = 1.0 / TARGET_FPS
    last_print_time = start_time
    total_detections = 0
    
    try:
        while True:
            loop_start = time.time()
            
            ret, frame = cap.read()
            if not ret:
                print("\n📹 End of video reached")
                break
            
            frame_count += 1
            
            # Resize frame to 480p
            frame_resized = cv2.resize(frame, (TARGET_WIDTH, TARGET_HEIGHT))
            
            # Run YOLO detection on resized frame
            results = model(frame_resized, conf=CONF_THRESHOLD, iou=IOU_THRESHOLD, verbose=False)
            
            # Get annotated frame with bounding boxes
            annotated_frame = results[0].plot()
            
            # Count detections
            detections = len(results[0].boxes) if results[0].boxes is not None else 0
            total_detections += detections
            
            # Encode frame as JPEG
            encode_start = time.time()
            _, buffer = cv2.imencode('.jpg', annotated_frame, 
                                    [cv2.IMWRITE_JPEG_QUALITY, JPEG_QUALITY])
            frame_base64 = base64.b64encode(buffer).decode('utf-8')
            encode_time = time.time() - encode_start
            
            # Prepare data
            data = {
                'frame': frame_base64,
                'timestamp': time.time(),
                'frame_number': frame_count,
                'detections': detections,
                'width': TARGET_WIDTH,
                'height': TARGET_HEIGHT,
                'model': os.path.basename(current_model_path)
            }
            
            # Send to server via WebSocket
            if is_connected:
                try:
                    ws.send(json.dumps(data))
                except Exception as e:
                    print(f"⚠️  Failed to send frame: {e}")
            else:
                print(f"⚠️  Not connected, skipping frame {frame_count}")
            
            # Display locally if requested
            if display:
                cv2.imshow('YOLO Detection Stream', annotated_frame)
                if cv2.waitKey(1) & 0xFF == ord('q'):
                    print("\n🛑 Stopped by user")
                    break
            
            # Print stats every 2 seconds
            current_time = time.time()
            if current_time - last_print_time >= 2.0:
                elapsed = current_time - start_time
                actual_fps = frame_count / elapsed
                avg_detections = total_detections / frame_count
                
                print(f"📊 Frame: {frame_count}/{total_frames} | "
                      f"FPS: {actual_fps:.1f} | "
                      f"Detections: {detections} (avg: {avg_detections:.1f}) | "
                      f"Encode: {encode_time*1000:.1f}ms")
                
                last_print_time = current_time
            
            # Control frame rate
            loop_time = time.time() - loop_start
            sleep_time = max(0, frame_interval - loop_time)
            if sleep_time > 0:
                time.sleep(sleep_time)
    
    except KeyboardInterrupt:
        print("\n\n🛑 Stopped by user (Ctrl+C)")
    
    finally:
        # Cleanup
        cap.release()
        if display:
            cv2.destroyAllWindows()
        
        # Print final stats
        elapsed = time.time() - start_time
        avg_fps = frame_count / elapsed if elapsed > 0 else 0
        avg_detections = total_detections / frame_count if frame_count > 0 else 0
        
        print("\n" + "="*60)
        print("📊 Final Statistics")
        print("="*60)
        print(f"⏱️  Duration: {elapsed:.2f}s")
        print(f"🎞️  Frames processed: {frame_count}")
        print(f"📈 Average FPS: {avg_fps:.2f}")
        print(f"🎯 Total detections: {total_detections}")
        print(f"📊 Average detections/frame: {avg_detections:.2f}")
        print("="*60)

# ==================== Main ====================
def main():
    """Main entry point"""
    # Declare globals first
    global GO_SERVER_URL, TARGET_FPS, JPEG_QUALITY, CONF_THRESHOLD, TARGET_WIDTH, TARGET_HEIGHT
    
    parser = argparse.ArgumentParser(
        description='Stream video with YOLO detection to Go server via WebSocket'
    )
    parser.add_argument('video', type=str, help='Path to video file')
    parser.add_argument('--display', action='store_true', 
                       help='Display video locally')
    parser.add_argument('--server', type=str, default=GO_SERVER_URL,
                       help=f'Go server URL (default: {GO_SERVER_URL})')
    parser.add_argument('--fps', type=int, default=TARGET_FPS,
                       help=f'Target FPS (default: {TARGET_FPS}, max recommended: 5 for bandwidth)')
    parser.add_argument('--quality', type=int, default=JPEG_QUALITY,
                       help=f'JPEG quality 0-100 (default: {JPEG_QUALITY}, lower = smaller size)')
    parser.add_argument('--conf', type=float, default=CONF_THRESHOLD,
                       help=f'Confidence threshold (default: {CONF_THRESHOLD})')
    parser.add_argument('--width', type=int, default=TARGET_WIDTH,
                       help=f'Target width (default: {TARGET_WIDTH})')
    parser.add_argument('--height', type=int, default=TARGET_HEIGHT,
                       help=f'Target height (default: {TARGET_HEIGHT})')
    
    args = parser.parse_args()
    
    # Update config from arguments
    GO_SERVER_URL = args.server
    TARGET_FPS = args.fps
    JPEG_QUALITY = args.quality
    CONF_THRESHOLD = args.conf
    TARGET_WIDTH = args.width
    TARGET_HEIGHT = args.height
    
    print("\n")
    print("="*60)
    print("🎥 YOLO Video Stream to WebSocket")
    print("="*60)
    print(f"📅 Started: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("="*60 + "\n")
    
    # Initialize model
    if not initialize_model():
        print("❌ Failed to initialize model")
        return
    
    # Connect to server
    global ws
    ws_url = f"{GO_SERVER_URL}{WEBSOCKET_PATH}"
    print(f"🔌 Connecting to server: {ws_url}...")
    try:
        ws = websocket.WebSocketApp(
            ws_url,
            on_open=on_open,
            on_message=on_message,
            on_error=on_error,
            on_close=on_close
        )
        
        # Start WebSocket in background thread
        ws_thread = threading.Thread(target=ws.run_forever, daemon=True)
        ws_thread.start()
        
        # Wait for connection
        time.sleep(2)
        
        if not is_connected:
            print(f"❌ Failed to connect to server")
            print(f"💡 Make sure Go server is running at {GO_SERVER_URL}")
            return
            
    except Exception as e:
        print(f"❌ Failed to connect: {e}")
        print(f"💡 Make sure Go server is running at {GO_SERVER_URL}")
        return
    
    # Process video
    try:
        process_video(args.video, display=args.display)
    finally:
        # Disconnect
        if ws and is_connected:
            print("\n🔌 Disconnecting from server...")
            ws.close()
        
        print("\n")
        print("="*60)
        print("👋 Program terminated")
        print(f"📅 Ended: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("="*60)

if __name__ == "__main__":
    main()
