<script lang="ts">
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import { useWebSocket } from '$lib/hooks/useWebSocket';
	import { env } from '$env/dynamic/public';

	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';
	const wsUrl = env.PUBLIC_WS_URL || 'ws://localhost:8080/ws';

	// WebSocket connection for real-time location updates
	const { data, status, send } = useWebSocket(wsUrl, {
		reconnect: true,
		reconnectInterval: 3000
	});

	// Initial center
	let mapCenter: [number, number] = $state([100.5018, 13.7563]);
	let markers = $state([
		{
			lngLat: [100.5018, 13.7563] as [number, number],
			popup: '<h3>Bangkok</h3><p>Starting point</p>',
			color: '#3b82f6'
		}
	]);

	// Update map when WebSocket data arrives
	$effect(() => {
		if ($data && typeof $data === 'object' && 'lat' in $data && 'lng' in $data) {
			const newLocation: [number, number] = [$data.lng, $data.lat];
			mapCenter = newLocation;

			// Add new marker
			markers = [
				...markers,
				{
					lngLat: newLocation,
					popup: `<h3>Update</h3><p>Lat: ${$data.lat}, Lng: ${$data.lng}</p>`,
					color: '#ef4444'
				}
			];
		}
	});

	function sendLocation() {
		// Example: Send current location to server
		send({
			type: 'location_update',
			lat: mapCenter[1],
			lng: mapCenter[0],
			timestamp: new Date().toISOString()
		});
	}
</script>

<svelte:head>
	<title>Map with WebSocket - Topgun</title>
</svelte:head>

<div class="page">
	<div class="header">
		<h1>Real-time Map</h1>
		<div class="controls">
			<div class="status">
				<span class="status-dot" class:connected={$status === 'connected'}></span>
				<span>WebSocket: {$status}</span>
			</div>
			<button onclick={sendLocation} disabled={$status !== 'connected'}>
				Send Location
			</button>
		</div>
	</div>

	<div class="map-container">
		{#if mapboxToken}
			<MapboxMap
				accessToken={mapboxToken}
				center={mapCenter}
				zoom={12}
				style="mapbox://styles/mapbox/streets-v12"
				{markers}
			/>
		{:else}
			<div class="error">
				<p>⚠️ Mapbox token is missing!</p>
				<p>Add PUBLIC_MAPBOX_TOKEN to your .env file</p>
			</div>
		{/if}
	</div>
</div>

<style>
	.page {
		width: 100%;
		height: 100vh;
		display: flex;
		flex-direction: column;
	}

	.header {
		background: white;
		padding: 1rem 1.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		display: flex;
		justify-content: space-between;
		align-items: center;
		z-index: 10;
	}

	h1 {
		margin: 0;
		font-size: 1.5rem;
		color: #1f2937;
	}

	.controls {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.status {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: #6b7280;
	}

	.status-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background: #9ca3af;
		transition: background 0.3s;
	}

	.status-dot.connected {
		background: #22c55e;
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

	button {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		background: #3b82f6;
		color: white;
		font-weight: 500;
		cursor: pointer;
		transition: background 0.2s;
	}

	button:hover:not(:disabled) {
		background: #2563eb;
	}

	button:disabled {
		background: #9ca3af;
		cursor: not-allowed;
	}

	.map-container {
		flex: 1;
		position: relative;
	}

	.error {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		background: #fff3cd;
		color: #856404;
		text-align: center;
	}

	.error p {
		margin: 0.5rem 0;
	}
</style>
