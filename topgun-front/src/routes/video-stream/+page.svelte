<script lang="ts">
	import VideoStream from '$lib/components/VideoStream.svelte';

	let serverUrl = 'ws://192.168.8.201:8080';
</script>

<svelte:head>
	<title>Live Video Stream | YOLO Detection</title>
</svelte:head>

<div class="page-container">
	<header class="page-header">
		<h1>🎥 Live Video Stream</h1>
		<p>Real-time YOLO object detection streaming</p>
	</header>

	<div class="stream-container">
		<VideoStream {serverUrl} showStats={true} autoReconnect={true} />
	</div>

	<div class="info-panel">
		<div class="info-card">
			<h3>📡 Connection Info</h3>
			<div class="info-item">
				<span class="label">Server:</span>
				<code>{serverUrl}</code>
			</div>
			<div class="info-item">
				<span class="label">Endpoint:</span>
				<code>/ws/video-stream</code>
			</div>
		</div>

		<div class="info-card">
			<h3>ℹ️ How it works</h3>
			<ol>
				<li>Python script processes video with YOLO</li>
				<li>Frames sent to Go server via WebSocket</li>
				<li>Go server broadcasts to all connected viewers</li>
				<li>Real-time display with detection stats</li>
			</ol>
		</div>

		<div class="info-card">
			<h3>🚀 Start Streaming</h3>
			<p>Run the Python video streamer:</p>
			<code class="command">
				cd submitGeardindaeng2025_sub_1<br />
				python3 video_stream.py P1_VIDEO_1.mp4
			</code>
		</div>
	</div>
</div>

<style>
	.page-container {
		min-height: 100vh;
		padding: 2rem;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	}

	.page-header {
		text-align: center;
		margin-bottom: 2rem;
		color: white;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: bold;
		margin-bottom: 0.5rem;
	}

	.page-header p {
		font-size: 1.1rem;
		opacity: 0.9;
	}

	.stream-container {
		max-width: 1400px;
		margin: 0 auto 2rem;
		height: 70vh;
		background: #000;
		border-radius: 12px;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
		overflow: hidden;
	}

	.info-panel {
		max-width: 1400px;
		margin: 0 auto;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.info-card {
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(10px);
		padding: 1.5rem;
		border-radius: 12px;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
	}

	.info-card h3 {
		margin: 0 0 1rem 0;
		color: #667eea;
		font-size: 1.2rem;
		font-weight: 600;
	}

	.info-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.75rem;
	}

	.info-item .label {
		font-weight: 600;
		color: #555;
		min-width: 80px;
	}

	code {
		background: #f3f4f6;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-family: 'Courier New', monospace;
		font-size: 0.9rem;
		color: #667eea;
	}

	.command {
		display: block;
		background: #1f2937;
		color: #10b981;
		padding: 1rem;
		border-radius: 8px;
		margin-top: 0.5rem;
		line-height: 1.6;
		overflow-x: auto;
	}

	ol {
		margin: 0;
		padding-left: 1.5rem;
	}

	ol li {
		margin-bottom: 0.5rem;
		color: #374151;
	}

	@media (max-width: 768px) {
		.page-container {
			padding: 1rem;
		}

		.page-header h1 {
			font-size: 1.8rem;
		}

		.stream-container {
			height: 50vh;
		}

		.info-panel {
			grid-template-columns: 1fr;
		}
	}
</style>
