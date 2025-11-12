<script lang="ts">
	interface Props {
		selectedCamerasCount: number;
		detectionsCount: number;
		activeConnectionsCount: number;
		onDisconnectAll: () => void;
	}

	let { selectedCamerasCount, detectionsCount, activeConnectionsCount, onDisconnectAll }: Props =
		$props();
</script>

<header class="header">
	<div class="header-content">
		<div class="logo-section">
			<!-- <div class="logo">🛡️</div> -->
			<div>
				<h1>Defensive Dashboard</h1>
				<p class="subtitle">Real-time Detection Monitoring</p>
			</div>
		</div>

		<div class="connection-section">
			<div class="status-info">
				<div class="stat-item">
					<span class="stat-label">Cameras</span>
					<span class="stat-value">{selectedCamerasCount}</span>
				</div>
				<div class="stat-item">
					<span class="stat-label">Detections</span>
					<span class="stat-value">{detectionsCount}</span>
				</div>
			</div>

			{#if selectedCamerasCount > 0}
				<button class="btn btn-danger" onclick={onDisconnectAll}>
					<span class="icon">⏹</span>
					Disconnect All
				</button>
			{/if}

			<div class="status-indicator" class:active={activeConnectionsCount > 0}>
				<span class="pulse"></span>
				<span
					>{activeConnectionsCount > 0
						? `${activeConnectionsCount} Connected`
						: 'Disconnected'}</span
				>
			</div>
		</div>
	</div>
</header>

<style>
	.header {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		color: white;
		padding: 1.5rem 2rem;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		z-index: 10;
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 2rem;
		max-width: 100%;
	}

	.logo-section {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.logo {
		font-size: 3rem;
		animation: float 3s ease-in-out infinite;
	}

	@keyframes float {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-10px);
		}
	}

	.logo-section h1 {
		margin: 0;
		font-size: 1.75rem;
		font-weight: 700;
	}

	.subtitle {
		margin: 0;
		opacity: 0.9;
		font-size: 0.9rem;
	}

	.connection-section {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.status-info {
		display: flex;
		gap: 1.5rem;
		padding: 0.5rem 1rem;
		background: rgba(255, 255, 255, 0.2);
		border-radius: 8px;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.25rem;
	}

	.stat-label {
		font-size: 0.75rem;
		opacity: 0.9;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
	}

	.btn {
		padding: 0.625rem 1.5rem;
		border-radius: 8px;
		border: none;
		font-weight: 600;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.2s;
		font-size: 0.9rem;
	}

	.btn-danger {
		background: #ef4444;
		color: white;
	}

	.btn-danger:hover {
		background: #dc2626;
		transform: translateY(-1px);
		box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
	}

	.status-indicator {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: rgba(255, 255, 255, 0.2);
		border-radius: 20px;
		font-size: 0.85rem;
		font-weight: 500;
	}

	.pulse {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #9ca3af;
		transition: background 0.3s;
	}

	.status-indicator.active .pulse {
		background: #10b981;
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

	.icon {
		display: inline-block;
	}
</style>
