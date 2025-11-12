#!/usr/bin/env python3
"""
YOLO Detection with Auto Model Update via MQTT
- Predict every 3 seconds
- Auto-update model when new model received via MQTT
- Save new models with timestamp
"""

import os
import csv
import json
import base64
import time
import threading
from datetime import datetime
from pathlib import Path

import paho.mqtt.client as mqtt
from ultralytics import YOLO
from model_utils import load_latest_model, get_latest_model

# ==================== Global Variables ====================
model = None
current_model_path = None
mqtt_client = None
is_running = True
prediction_interval = 10  # วินาที

# ==================== MQTT Callbacks ====================
def on_connect(client, userdata, flags, rc):
    """Callback when connected to MQTT broker"""
    if rc == 0:
        print(f"✅ Connected to MQTT broker successfully")
        client.subscribe("topgun/ai")
        print(f"📡 Subscribed to topic: topgun/ai")
    else:
        print(f"❌ Failed to connect, return code {rc}")

def on_message(client, userdata, msg):
    """Callback when receiving a message - auto-reload model with timestamp"""
    global current_model_path, model
    
    try:
        # Try to parse as JSON (for file uploads)
        data = json.loads(msg.payload)
        
        if isinstance(data, dict) and 'metadata' in data and 'data' in data:
            # This is a file upload
            metadata = data['metadata']
            original_filename = metadata.get('filename', 'model.pt')
            file_size = metadata.get('size', 0)
            
            print(f"\n{'='*60}")
            print(f"📥 Receiving new model file: {original_filename}")
            print(f"📦 Size: {file_size:,} bytes ({file_size/1024/1024:.2f} MB)")
            
            # Decode base64 data
            file_content = base64.b64decode(data['data'])
            
            # Create models directory if it doesn't exist
            os.makedirs('models', exist_ok=True)
            
            # Generate timestamped filename
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            file_ext = Path(original_filename).suffix
            timestamped_filename = f"model_{timestamp}{file_ext}"
            file_path = os.path.join('models', timestamped_filename)
            
            # Save the file with timestamp
            with open(file_path, 'wb') as f:
                f.write(file_content)
            
            print(f"💾 Saved as: {timestamped_filename}")
            
            # If it's a .pt or .pth file, reload the model
            if file_ext in ['.pt', '.pth']:
                try:
                    print(f"🔄 Loading new model...")
                    new_model = YOLO(file_path)
                    
                    # Update global variables
                    model = new_model
                    current_model_path = file_path
                    
                    print(f"✅ Model updated successfully!")
                    print(f"🎯 Now using: {timestamped_filename}")
                    print(f"{'='*60}\n")
                except Exception as e:
                    print(f"❌ Failed to load model: {e}")
                    print(f"⚠️  Keeping current model: {current_model_path}")
                    print(f"{'='*60}\n")
            else:
                print(f"⚠️  File extension {file_ext} is not a model file")
                print(f"{'='*60}\n")
        else:
            # Regular JSON message
            print(f"📨 Received JSON: {data}")
            
    except json.JSONDecodeError:
        # Not JSON, just a text message
        message = msg.payload.decode('utf-8', errors='ignore')
        print(f"💬 Received message: {message[:100]}")
    except Exception as e:
        print(f"❌ Error processing message: {e}")

def on_disconnect(client, userdata, rc):
    """Callback when disconnected from broker"""
    if rc != 0:
        print(f"⚠️  Connection lost. Reconnecting...")

# ==================== Model Functions ====================
def initialize_model():
    """Initialize YOLO model - load latest available"""
    global model, current_model_path
    
    print("="*60)
    print("🚀 Initializing YOLO Model")
    print("="*60)
    print("🔍 Looking for latest model...")
    
    model = load_latest_model()
    
    if model is None:
        # No model in models/ folder, use default
        default_path = "models/eiei1.pt"
        if os.path.exists(default_path):
            print(f"📝 Using default model: {default_path}")
            model = YOLO(default_path)
            current_model_path = default_path
        else:
            print(f"❌ No model found!")
            print(f"💡 Please place a model file in ./models/ directory")
            return False
    else:
        current_model_path = get_latest_model()
        print(f"✅ Loaded latest model: {os.path.basename(current_model_path)}")
    
    print(f"🎯 Current model: {current_model_path}")
    print("="*60 + "\n")
    return True

def predict_bbox(image_path):
    """
    ใช้ YOLO ตรวจจับวัตถุจากภาพ
    Returns: list of bounding boxes [(center_x, center_y, width, height), ...]
    """
    global model
    
    if model is None:
        print("❌ Model not loaded!")
        return [[0, 0, 0, 0]]
    
    try:
        results = model.track(source=image_path, conf=0.6, iou=0.4, imgsz=1920, verbose=False)
        all_boxes = []

        for result in results:
            boxes = result.boxes
            if boxes is not None:
                for box in boxes:
                    xywh = box.xywh[0].cpu().numpy().tolist()  # [center_x, center_y, width, height]
                    all_boxes.append(tuple(xywh))

        if len(all_boxes) == 0:
            return [[0, 0, 0, 0]]

        return all_boxes
        
    except Exception as e:
        print(f"❌ Error during prediction: {e}")
        return [[0, 0, 0, 0]]

def run_prediction_loop():
    """
    Main prediction loop - runs every 10 seconds
    """
    global is_running, current_model_path
    
    folder = "TEST_DATA"
    
    # Check if TEST_DATA folder exists
    if not os.path.exists(folder):
        print(f"❌ Folder '{folder}' not found!")
        print(f"💡 Please create the folder and add test images")
        return
    
    print("="*60)
    print("🎯 Starting Continuous Prediction Loop")
    print("="*60)
    print(f"📁 Input folder: {folder}/")
    print(f"📄 Output file: output.csv")
    print(f"⏱️  Interval: {prediction_interval} seconds")
    print(f"🛑 Press Ctrl+C to stop")
    print("="*60 + "\n")
    
    iteration = 0
    
    try:
        while is_running:
            iteration += 1
            start_time = time.time()
            
            print(f"\n{'─'*60}")
            print(f"🔄 Iteration #{iteration} - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
            print(f"🎯 Using model: {os.path.basename(current_model_path)}")
            print(f"{'─'*60}")
            
            # Get all images
            files = os.listdir(folder)
            files = [x for x in files if x.endswith(('.jpg', '.jpeg', '.png'))]
            
            if len(files) == 0:
                print(f"⚠️  No images found in {folder}/")
                time.sleep(prediction_interval)
                continue
            
            print(f"📷 Found {len(files)} images")
            
            # Create/overwrite CSV file
            with open("output.csv", "w", newline="") as csvfile:
                writer = csv.writer(csvfile)
                writer.writerow(["image_name", "center_x", "center_y", "width", "height"])
            
            # Process each image
            processed = 0
            total_detections = 0
            
            for idx, fname in enumerate(files, 1):
                image_path = os.path.join(folder, fname)
                
                # Predict
                bboxs = predict_bbox(image_path)
                
                if bboxs == [[0, 0, 0, 0]]:
                    print(f"  [{idx}/{len(files)}] {fname}: No detections")
                    continue
                
                # Save to CSV
                for center_x, center_y, width, height in bboxs:
                    row = [fname, center_x, center_y, width, height]
                    with open("output.csv", "a", newline="") as csvfile:
                        writer = csv.writer(csvfile)
                        writer.writerow(row)
                
                print(f"  [{idx}/{len(files)}] {fname}: ✅ {len(bboxs)} objects")
                processed += 1
                total_detections += len(bboxs)
            
            elapsed = time.time() - start_time
            
            print(f"\n📊 Results:")
            print(f"  ✅ Processed: {processed}/{len(files)} images")
            print(f"  🎯 Detections: {total_detections} objects")
            print(f"  ⏱️  Time taken: {elapsed:.2f}s")
            print(f"  📄 Saved to: output.csv")
            
            # Wait for next iteration
            wait_time = max(0, prediction_interval - elapsed)
            if wait_time > 0:
                print(f"\n⏳ Waiting {wait_time:.1f}s until next prediction...")
                time.sleep(wait_time)
            
    except KeyboardInterrupt:
        print(f"\n\n{'='*60}")
        print("🛑 Stopping prediction loop...")
        print(f"📊 Total iterations completed: {iteration}")
        print("="*60)
        is_running = False

# ==================== MQTT Setup ====================
def setup_mqtt():
    """Setup MQTT client for receiving model updates"""
    global mqtt_client
    
    print("="*60)
    print("📡 Setting up MQTT Client")
    print("="*60)
    
    # Create MQTT client
    mqtt_client = mqtt.Client(client_id="python-yolo-main", clean_session=True)
    
    # Set callbacks
    mqtt_client.on_connect = on_connect
    mqtt_client.on_message = on_message
    mqtt_client.on_disconnect = on_disconnect
    
    # Configure auto-reconnect
    mqtt_client.reconnect_delay_set(min_delay=1, max_delay=120)
    
    # Connect to broker
    broker = "localhost"
    port = 1883
    
    try:
        print(f"🔌 Connecting to MQTT broker at {broker}:{port}...")
        mqtt_client.connect(broker, port, keepalive=60)
        
        # Start background loop
        mqtt_client.loop_start()
        
        time.sleep(2)  # Wait for connection
        
        if mqtt_client.is_connected():
            print(f"✅ MQTT client connected successfully")
            print("="*60 + "\n")
            return True
        else:
            print(f"⚠️  MQTT connection is establishing...")
            print("="*60 + "\n")
            return True
            
    except Exception as e:
        print(f"❌ Failed to connect to MQTT broker: {e}")
        print(f"💡 Make sure Mosquitto is running: docker-compose up -d")
        print("="*60 + "\n")
        return False

# ==================== Main ====================
def main():
    """Main entry point"""
    print("\n")
    print("="*60)
    print("🤖 YOLO Auto-Prediction System with MQTT Model Update")
    print("="*60)
    print(f"📅 Started at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print("="*60 + "\n")
    
    # 1. Initialize model
    if not initialize_model():
        print("❌ Failed to initialize model. Exiting...")
        return
    
    # 2. Setup MQTT
    mqtt_ok = setup_mqtt()
    if not mqtt_ok:
        print("⚠️  MQTT setup failed, but continuing with local model...")
    
    # 3. Start prediction loop
    try:
        run_prediction_loop()
    finally:
        # Cleanup
        global mqtt_client, is_running
        is_running = False
        
        if mqtt_client is not None:
            print("\n🔌 Disconnecting MQTT client...")
            mqtt_client.loop_stop()
            mqtt_client.disconnect()
        
        print("\n")
        print("="*60)
        print("👋 Program terminated gracefully")
        print(f"📅 Ended at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print("="*60)

if __name__ == "__main__":
    main()
