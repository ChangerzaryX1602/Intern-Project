<script lang="ts">
	import { onMount } from 'svelte';
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import { env } from '$env/dynamic/public';

	// Get Mapbox token from environment or use placeholder
	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';

	// State for current location
	let currentLocation: [number, number] = $state([100.5018, 13.7563]); // Default: Bangkok
	let locationError = $state<string | null>(null);
	let isLoadingLocation = $state(true);

	// Define markers - will update with current location
	let markers = $state([
		{
			lngLat: [100.5018, 13.7563] as [number, number],
			popup: '<h3>Your Location</h3><p>Current position</p>',
			color: '#3b82f6'
		}
	]);

	function getCurrentLocation() {
		if (!('geolocation' in navigator)) {
			isLoadingLocation = false;
			locationError = 'Geolocation not supported by your browser.';
			return;
		}

		isLoadingLocation = true;
		locationError = null;

		navigator.geolocation.getCurrentPosition(
			(position) => {
				const { latitude, longitude, accuracy } = position.coords;
				console.log('✅ Location obtained:', { latitude, longitude, accuracy });
				
				currentLocation = [longitude, latitude];
				
				// Update marker with current location
				markers = [
					{
						lngLat: [longitude, latitude] as [number, number],
						popup: `<h3>Your Location</h3><p>Lat: ${latitude.toFixed(6)}<br/>Lng: ${longitude.toFixed(6)}<br/>Accuracy: ±${accuracy.toFixed(0)}m</p>`,
						color: '#3b82f6'
					}
				];
				
				isLoadingLocation = false;
			},
			(error) => {
				isLoadingLocation = false;
				console.error('❌ Geolocation error:', error);
				
				switch (error.code) {
					case error.PERMISSION_DENIED:
						locationError = 'Location access denied. Please allow location access in your browser settings.';
						break;
					case error.POSITION_UNAVAILABLE:
						locationError = 'Location unavailable. Please check your device settings.';
						break;
					case error.TIMEOUT:
						locationError = 'Location request timeout. Please try again.';
						break;
					default:
						locationError = `Error getting location: ${error.message}`;
				}
			},
			{
				enableHighAccuracy: true,
				timeout: 10000,
				maximumAge: 0
			}
		);
	}

	onMount(() => {
		getCurrentLocation();
	});
</script>

<svelte:head>
	<title>Map - Topgun</title>
</svelte:head>

<div class="map-page">
	<div class="header">
		<div class="header-left">
			<h1>Map</h1>
			<button class="refresh-btn" onclick={getCurrentLocation} disabled={isLoadingLocation}>
				{#if isLoadingLocation}
					<span class="spinner-small"></span>
				{:else}
					🔄
				{/if}
				Refresh Location
			</button>
		</div>
		{#if isLoadingLocation}
			<div class="location-status loading">
				<span class="spinner"></span>
				<span>Getting your location...</span>
			</div>
		{:else if locationError}
			<div class="location-status error">
				<span>⚠️</span>
				<span>{locationError}</span>
			</div>
		{:else}
			<div class="location-status success">
				<span>📍</span>
				<span>Lat: {currentLocation[1].toFixed(4)}, Lng: {currentLocation[0].toFixed(4)}</span>
			</div>
		{/if}
	</div>
	<div class="map-wrapper">
		{#if mapboxToken}
			<MapboxMap
				accessToken={mapboxToken}
				center={currentLocation}
				zoom={15}
				style="mapbox://styles/mapbox/streets-v12"
				{markers}
			/>
		{:else}
			<div class="error-message">
				<p>⚠️ Mapbox token is missing!</p>
				<p>Please add PUBLIC_MAPBOX_TOKEN to your .env file</p>
			</div>
		{/if}
	</div>
</div>

<style>
	.map-page {
		width: 100%;
		height: 100vh;
		display: flex;
		flex-direction: column;
	}

	.header {
		padding: 1rem;
		background: #1a1a1a;
		color: white;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	h1 {
		margin: 0;
		font-size: 1.5rem;
	}

	.refresh-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: #3b82f6;
		color: white;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		cursor: pointer;
		transition: background 0.2s;
	}

	.refresh-btn:hover:not(:disabled) {
		background: #2563eb;
	}

	.refresh-btn:disabled {
		background: #6b7280;
		cursor: not-allowed;
	}

	.spinner-small {
		width: 12px;
		height: 12px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	.location-status {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
	}

	.location-status.loading {
		background: rgba(59, 130, 246, 0.2);
		color: #93c5fd;
	}

	.location-status.success {
		background: rgba(34, 197, 94, 0.2);
		color: #86efac;
	}

	.location-status.error {
		background: rgba(239, 68, 68, 0.2);
		color: #fca5a5;
	}

	.spinner {
		width: 14px;
		height: 14px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.map-wrapper {
		flex: 1;
		position: relative;
		overflow: hidden;
	}

	.error-message {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		background: #fff3cd;
		color: #856404;
		padding: 2rem;
		text-align: center;
	}

	.error-message p {
		margin: 0.5rem 0;
		font-size: 1.1rem;
	}
</style>
