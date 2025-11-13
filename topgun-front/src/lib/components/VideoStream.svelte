<script lang="ts">
	import { onMount, onDestroy } from 'svelte';

	// Props
	export let serverUrl: string = 'ws://localhost:8080';
	export let autoReconnect: boolean = true;
	export let showStats: boolean = true;

	// State
	let ws: WebSocket | null = null;
	let connected = false;
	let frameNumber = 0;
	let detections = 0;
	let fps = 0;
	let modelName = '';
	let lastFrameTime = 0;
	let frameCount = 0;
	let fpsInterval: number;

	// Image data
	let imageSrc = '';

	// Reconnection
	let reconnectAttempts = 0;
	let reconnectTimer: number;
	const MAX_RECONNECT_ATTEMPTS = 999; // Essentially unlimited
	const RECONNECT_BASE_DELAY = 1000; // 1 second
	const RECONNECT_MAX_DELAY = 10000; // 10 seconds max

	// Heartbeat / Ping-Pong
	let heartbeatInterval: number;
	let lastHeartbeat = 0;
	const HEARTBEAT_INTERVAL = 5000; // Send ping every 5 seconds
	const HEARTBEAT_TIMEOUT = 15000; // Consider dead if no response in 15 seconds
	let missedHeartbeats = 0;
	const MAX_MISSED_HEARTBEATS = 3;

	function connect() {
		try {
			console.log(`Connecting to ${serverUrl}/ws/video-stream...`);
			ws = new WebSocket(`${serverUrl}/ws/video-stream`);

			ws.onopen = () => {
				console.log('✅ WebSocket connected');
				connected = true;
				reconnectAttempts = 0;
				missedHeartbeats = 0;
				lastHeartbeat = Date.now();

				// Start heartbeat
				startHeartbeat();
			};

			ws.onmessage = (event) => {
				try {
					// Update last heartbeat on any message received
					lastHeartbeat = Date.now();
					missedHeartbeats = 0;

					const data = JSON.parse(event.data);

					// Check if it's a status message
					if (data.status === 'connected') {
						console.log('📡 Subscribed to video stream');
						return;
					}

					// Check if it's a pong response
					if (data.type === 'pong') {
						return; // Just reset heartbeat timer
					}

					// Update frame data
					if (data.frame) {
						imageSrc = `data:image/jpeg;base64,${data.frame}`;
						frameNumber = data.frame_number || 0;
						detections = data.detections || 0;
						modelName = data.model || '';

						// Calculate FPS
						const now = performance.now();
						if (lastFrameTime > 0) {
							frameCount++;
							if (now - lastFrameTime > 1000) {
								fps = Math.round((frameCount * 1000) / (now - lastFrameTime));
								frameCount = 0;
								lastFrameTime = now;
							}
						} else {
							lastFrameTime = now;
						}
					}
				} catch (err) {
					console.error('Error parsing message:', err);
				}
			};

			ws.onerror = (error) => {
				console.error('❌ WebSocket error:', error);
			};

			ws.onclose = () => {
				console.log('⚠️ WebSocket disconnected');
				connected = false;
				imageSrc = '';
				
				// Stop heartbeat
				stopHeartbeat();

				// Auto reconnect
				if (autoReconnect) {
					scheduleReconnect();
				}
			};
		} catch (err) {
			console.error('Failed to create WebSocket:', err);
			if (autoReconnect) {
				scheduleReconnect();
			}
		}
	}

	function scheduleReconnect() {
		if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
			console.error('Max reconnection attempts reached');
			return;
		}

		reconnectAttempts++;
		const delay = Math.min(
			RECONNECT_BASE_DELAY * Math.pow(1.5, reconnectAttempts),
			RECONNECT_MAX_DELAY
		);
		
		console.log(`Reconnecting in ${(delay / 1000).toFixed(1)}s... (attempt ${reconnectAttempts})`);
		
		reconnectTimer = setTimeout(() => {
			if (!connected) {
				connect();
			}
		}, delay);
	}

	function startHeartbeat() {
		// Clear existing interval
		if (heartbeatInterval) {
			clearInterval(heartbeatInterval);
		}

		// Send ping periodically
		heartbeatInterval = setInterval(() => {
			if (!ws || ws.readyState !== WebSocket.OPEN) {
				return;
			}

			const timeSinceLastHeartbeat = Date.now() - lastHeartbeat;

			// Check if we missed too many heartbeats
			if (timeSinceLastHeartbeat > HEARTBEAT_TIMEOUT) {
				missedHeartbeats++;
				console.warn(`⚠️  Missed heartbeat (${missedHeartbeats}/${MAX_MISSED_HEARTBEATS})`);

				if (missedHeartbeats >= MAX_MISSED_HEARTBEATS) {
					console.error('❌ Connection seems dead, forcing reconnect...');
					ws?.close();
					return;
				}
			}

			// Send ping
			try {
				ws.send(JSON.stringify({ type: 'ping', timestamp: Date.now() }));
			} catch (err) {
				console.error('Failed to send ping:', err);
			}
		}, HEARTBEAT_INTERVAL);
	}

	function stopHeartbeat() {
		if (heartbeatInterval) {
			clearInterval(heartbeatInterval);
			heartbeatInterval = 0;
		}
	}

	function disconnect() {
		// Stop heartbeat
		stopHeartbeat();
		
		// Clear reconnect timer
		if (reconnectTimer) {
			clearTimeout(reconnectTimer);
		}
		
		// Clear FPS interval
		if (fpsInterval) {
			clearInterval(fpsInterval);
		}
		
		// Close WebSocket
		if (ws) {
			// Disable auto-reconnect temporarily
			const wasAutoReconnect = autoReconnect;
			autoReconnect = false;
			
			ws.close();
			ws = null;
			
			// Restore auto-reconnect setting
			autoReconnect = wasAutoReconnect;
		}
	}

	onMount(() => {
		connect();

		// Update FPS every second
		fpsInterval = setInterval(() => {
			// FPS calculation is done in onmessage
		}, 1000);
	});

	onDestroy(() => {
		disconnect();
	});
</script>

<div class="video-stream-container">
	<div class="video-wrapper">
		{#if imageSrc}
			<img src={imageSrc} alt="Video Stream" class="video-frame" />
		{:else}
			<div class="no-stream">
				<div class="spinner"></div>
				<p>
					{#if connected}
						Waiting for video frames...
					{:else}
						Connecting to server...
					{/if}
				</p>
			</div>
		{/if}

		
	</div>
</div>

<style>
	.video-stream-container {
		width: 100%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
		background: #000;
		border-radius: 8px;
		overflow: hidden;
	}

	.video-wrapper {
		position: relative;
		width: 100%;
		height: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.video-frame {
		max-width: 100%;
		max-height: 100%;
		object-fit: contain;
		display: block;
	}

	.no-stream {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		color: #fff;
		padding: 2rem;
	}

	.spinner {
		width: 50px;
		height: 50px;
		border: 4px solid rgba(255, 255, 255, 0.1);
		border-top-color: #fff;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.stats-overlay {
		position: absolute;
		top: 1rem;
		right: 1rem;
		background: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(10px);
		padding: 1rem;
		border-radius: 8px;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		font-family: 'Courier New', monospace;
		font-size: 0.9rem;
		min-width: 200px;
	}

	.stat {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
	}

	.stat-label {
		color: #aaa;
		font-weight: 600;
	}

	.stat-value {
		color: #fff;
		font-weight: bold;
	}

	.stat-value.detections {
		color: #4ade80;
	}

	.stat-value.model {
		color: #60a5fa;
		font-size: 0.8rem;
	}

	.connection-status {
		padding: 0.25rem 0.75rem;
		border-radius: 4px;
		font-weight: bold;
		text-align: center;
		width: 100%;
	}

	.connection-status.connected {
		background: rgba(74, 222, 128, 0.2);
		color: #4ade80;
	}

	.connection-status.disconnected {
		background: rgba(248, 113, 113, 0.2);
		color: #f87171;
	}

	.connection-status.reconnecting {
		background: rgba(251, 191, 36, 0.2);
		color: #fbbf24;
		animation: pulse 2s ease-in-out infinite;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	@media (max-width: 768px) {
		.stats-overlay {
			font-size: 0.8rem;
			padding: 0.75rem;
			min-width: 150px;
		}
	}
</style>
