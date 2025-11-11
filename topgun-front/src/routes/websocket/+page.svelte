<script lang="ts">
	import { useWebSocket } from '$lib/hooks/useWebSocket';
	import { env } from '$env/dynamic/public';

	// Example WebSocket URL - replace with your actual WebSocket server
	const WS_URL = env.PUBLIC_WS_URL || 'ws://localhost:8080/ws';

	// Initialize WebSocket connection
	const { data, status, error, send, connect, disconnect } = useWebSocket(WS_URL, {
		reconnect: true,
		reconnectInterval: 3000,
		reconnectAttempts: 5,
		onOpen: () => console.log('WebSocket opened'),
		onClose: () => console.log('WebSocket closed'),
		onMessage: (event) => console.log('Message received:', event.data)
	});

	let message = $state('');
	let messageHistory: any[] = $state([]);

	// Track message history
	$effect(() => {
		if ($data) {
			messageHistory = [...messageHistory, { timestamp: new Date(), data: $data }];
		}
	});

	function sendMessage() {
		if (message.trim()) {
			send({ type: 'message', content: message });
			message = '';
		}
	}

	function handleKeyPress(event: KeyboardEvent) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			sendMessage();
		}
	}

	function clearHistory() {
		messageHistory = [];
	}

	function getStatusColor(currentStatus: string) {
		switch (currentStatus) {
			case 'connected':
				return '#22c55e';
			case 'connecting':
				return '#f59e0b';
			case 'error':
				return '#ef4444';
			default:
				return '#6b7280';
		}
	}
</script>

<svelte:head>
	<title>WebSocket Demo - Topgun</title>
</svelte:head>

<div class="websocket-page">
	<div class="header">
		<h1>WebSocket Connection</h1>
		<div class="status-bar">
			<div class="status-indicator" style="background-color: {getStatusColor($status)}"></div>
			<span class="status-text">Status: {$status}</span>
		</div>
	</div>

	<div class="content">
		<div class="info-section">
			<h2>Connection Info</h2>
			<div class="info-grid">
				<div class="info-item">
					<span class="info-label">URL:</span>
					<code>{WS_URL}</code>
				</div>
				<div class="info-item">
					<span class="info-label">Status:</span>
					<span class="status-badge" style="background-color: {getStatusColor($status)}">
						{$status}
					</span>
				</div>
				{#if $error}
					<div class="info-item error">
						<span class="info-label">Error:</span>
						<span>{$error}</span>
					</div>
				{/if}
			</div>
			<div class="button-group">
				<button onclick={connect} disabled={$status === 'connected'}>Connect</button>
				<button onclick={disconnect} disabled={$status === 'disconnected'}>Disconnect</button>
			</div>
		</div>

		<div class="message-section">
			<h2>Send Message</h2>
			<div class="message-input-group">
				<input
					type="text"
					bind:value={message}
					onkeypress={handleKeyPress}
					placeholder="Type your message..."
					disabled={$status !== 'connected'}
				/>
				<button onclick={sendMessage} disabled={$status !== 'connected' || !message.trim()}>
					Send
				</button>
			</div>
		</div>

		<div class="history-section">
			<div class="history-header">
				<h2>Message History ({messageHistory.length})</h2>
				<button onclick={clearHistory} class="clear-btn">Clear</button>
			</div>
			<div class="message-list">
				{#if messageHistory.length === 0}
					<div class="empty-state">No messages yet</div>
				{:else}
					{#each messageHistory.slice().reverse() as item}
						<div class="message-item">
							<div class="message-time">
								{item.timestamp.toLocaleTimeString()}
							</div>
							<div class="message-content">
								<pre>{JSON.stringify(item.data, null, 2)}</pre>
							</div>
						</div>
					{/each}
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.websocket-page {
		width: 100%;
		min-height: 100vh;
		background: #f3f4f6;
	}

	.header {
		background: white;
		padding: 1.5rem 2rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	h1 {
		margin: 0;
		font-size: 1.875rem;
		color: #1f2937;
	}

	.status-bar {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.status-indicator {
		width: 12px;
		height: 12px;
		border-radius: 50%;
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

	.status-text {
		font-weight: 500;
		color: #4b5563;
	}

	.content {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.info-section,
	.message-section,
	.history-section {
		background: white;
		padding: 1.5rem;
		border-radius: 0.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	h2 {
		margin: 0 0 1rem 0;
		font-size: 1.25rem;
		color: #1f2937;
	}

	.info-grid {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	.info-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.info-item .info-label {
		font-weight: 600;
		color: #4b5563;
		min-width: 80px;
	}

	.info-item code {
		background: #f3f4f6;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.875rem;
	}

	.info-item.error {
		color: #ef4444;
	}

	.status-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 9999px;
		font-size: 0.875rem;
		font-weight: 500;
		color: white;
	}

	.button-group {
		display: flex;
		gap: 0.5rem;
	}

	button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-weight: 500;
		cursor: pointer;
		background: #3b82f6;
		color: white;
		transition: background 0.2s;
	}

	button:hover:not(:disabled) {
		background: #2563eb;
	}

	button:disabled {
		background: #9ca3af;
		cursor: not-allowed;
	}

	.message-input-group {
		display: flex;
		gap: 0.5rem;
	}

	input {
		flex: 1;
		padding: 0.5rem 1rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		font-size: 1rem;
	}

	input:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	input:disabled {
		background: #f3f4f6;
		cursor: not-allowed;
	}

	.history-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.clear-btn {
		background: #ef4444;
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
	}

	.clear-btn:hover:not(:disabled) {
		background: #dc2626;
	}

	.message-list {
		max-height: 400px;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.empty-state {
		text-align: center;
		color: #9ca3af;
		padding: 2rem;
	}

	.message-item {
		padding: 1rem;
		background: #f9fafb;
		border-radius: 0.375rem;
		border-left: 3px solid #3b82f6;
	}

	.message-time {
		font-size: 0.75rem;
		color: #6b7280;
		margin-bottom: 0.5rem;
	}

	.message-content pre {
		margin: 0;
		font-size: 0.875rem;
		white-space: pre-wrap;
		word-break: break-word;
	}
</style>
