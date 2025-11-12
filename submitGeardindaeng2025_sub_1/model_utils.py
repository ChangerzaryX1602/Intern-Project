"""
Helper module for automatically using the latest YOLO model
"""

import os
from datetime import datetime
from ultralytics import YOLO


def get_latest_model(models_dir="./models"):
    """
    Get the path to the latest model file based on modification time
    
    Args:
        models_dir: Directory containing model files
        
    Returns:
        str: Path to the latest model file, or None if no models found
    """
    if not os.path.exists(models_dir):
        return None
    
    models = [f for f in os.listdir(models_dir) if f.endswith(('.pt', '.pth'))]
    if not models:
        return None
    
    # Get full paths and sort by modification time (newest first)
    model_paths = [os.path.join(models_dir, m) for m in models]
    model_paths.sort(key=lambda x: os.path.getmtime(x), reverse=True)
    
    return model_paths[0]


def load_latest_model(models_dir="./models"):
    """
    Load the latest YOLO model from the models directory
    
    Args:
        models_dir: Directory containing model files
        
    Returns:
        YOLO: Loaded YOLO model object, or None if no models found
    """
    latest_model_path = get_latest_model(models_dir)
    
    if latest_model_path is None:
        print(f"❌ No models found in {models_dir}")
        return None
    
    try:
        print(f"🔄 Loading latest model: {latest_model_path}")
        model = YOLO(latest_model_path)
        
        # Show modification time
        mod_time = os.path.getmtime(latest_model_path)
        mod_datetime = datetime.fromtimestamp(mod_time).strftime("%Y-%m-%d %H:%M:%S")
        print(f"✅ Model loaded successfully!")
        print(f"📅 Last updated: {mod_datetime}")
        
        return model
    except Exception as e:
        print(f"❌ Failed to load model: {e}")
        return None


def list_all_models(models_dir="./models"):
    """
    List all available models sorted by modification time
    
    Args:
        models_dir: Directory containing model files
        
    Returns:
        list: List of tuples (filename, full_path, mod_time)
    """
    if not os.path.exists(models_dir):
        print(f"📦 Directory not found: {models_dir}")
        return []
    
    models = [f for f in os.listdir(models_dir) if f.endswith(('.pt', '.pth'))]
    if not models:
        print(f"📦 No models found in {models_dir}")
        return []
    
    # Get full paths and info
    model_info = []
    for name in models:
        path = os.path.join(models_dir, name)
        mod_time = os.path.getmtime(path)
        size = os.path.getsize(path)
        model_info.append((name, path, mod_time, size))
    
    # Sort by modification time (newest first)
    model_info.sort(key=lambda x: x[2], reverse=True)
    
    # Print info
    print(f"\n📦 Available models in {models_dir}:")
    for i, (name, path, mod_time, size) in enumerate(model_info, 1):
        mod_datetime = datetime.fromtimestamp(mod_time).strftime("%Y-%m-%d %H:%M:%S")
        size_mb = size / (1024 * 1024)
        marker = "🎯" if i == 1 else "  "
        print(f"{marker} {i}. {name}")
        print(f"      Updated: {mod_datetime} | Size: {size_mb:.2f} MB")
    print()
    
    return model_info


# Example usage:
if __name__ == "__main__":
    # List all models
    list_all_models()
    
    # Load the latest model
    model = load_latest_model()
    
    if model:
        print(f"🎯 Ready to use the latest model!")
