<script lang="ts">
	import { onMount } from 'svelte';
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import { env } from '$env/dynamic/public';

	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';

	// Mock data for drones
	interface Drone {
		id: string;
		name: string;
		status: 'connected' | 'disconnected';
		location: string;
		coordinates: {
			lat: number;
			lng: number;
		};
		gpsStatus: 'good' | 'loss';
		lastUpdate: string;
	}

	let drones = $state<Drone[]>([
		{
			id: 'A01',
			name: 'Drone ID: A01',
			status: 'connected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: { lat: 14.286169, lng: 101.171044 },
			gpsStatus: 'good',
			lastUpdate: '11/11/2025 18:16 น.'
		},
		{
			id: 'A02',
			name: 'Drone ID: A02',
			status: 'connected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: { lat: 14.287215, lng: 101.1716 },
			gpsStatus: 'good',
			lastUpdate: '11/11/2025 18:16 น.'
		},
		{
			id: 'A03',
			name: 'Drone ID: A03',
			status: 'disconnected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: { lat: 14.286958, lng: 101.170515 },
			gpsStatus: 'loss',
			lastUpdate: '11/11/2025 20:57 น.'
		}
	]);

	let selectedDrone = $state<Drone | null>(null);
	let searchQuery = $state('');
	let mapCenter: [number, number] = $state([101.171, 14.287]);

	// Generate markers for connected drones
	let markers = $derived(
		drones
			.filter((d) => d.status === 'connected')
			.map((d) => ({
				lngLat: [d.coordinates.lng, d.coordinates.lat] as [number, number],
				popup: `
					<div style="font-size:13px">
						<strong>${d.name}</strong><br/>
						${d.location}<br/>
						GPS: ${d.gpsStatus === 'good' ? '✓ Connected' : '✗ Loss'}<br/>
						${d.lastUpdate}
					</div>`,
				color: d.gpsStatus === 'good' ? '#10b981' : '#ef4444'
			}))
	);

	// Filter drones based on search
	let filteredDrones = $derived(
		drones.filter((d) => d.name.toLowerCase().includes(searchQuery.toLowerCase()))
	);

	function selectDrone(drone: Drone) {
		selectedDrone = drone;
		// Center map on selected drone
		if (drone.status === 'connected') {
			mapCenter = [drone.coordinates.lng, drone.coordinates.lat];
		}
	}

	function toggleDroneStatus(droneId: string) {
		const drone = drones.find((d) => d.id === droneId);
		if (drone) {
			drone.status = drone.status === 'connected' ? 'disconnected' : 'connected';
			if (drone.status === 'disconnected') {
				drone.gpsStatus = 'loss';
			}
		}
	}

	onMount(() => {
		console.log('Offensive Dashboard mounted');
	});
</script>

<svelte:head>
	<title>Offensive Dashboard - Drone Control</title>
</svelte:head>

<div class="dashboard">
	<!-- Header -->
	<header class="header">
		<div class="header-content">
			<div class="logo-section">
				<div class="logo">🎯</div>
				<div>
					<h1>Offensive Dashboard</h1>
					<p class="subtitle">Drone Control & Monitoring</p>
				</div>
			</div>

			<div class="server-time">
				<span class="time-label">Server Time:</span>
				<span class="time-value">11/11/2025 19:02</span>
				<span class="status-badge">🔴 LIVE</span>
			</div>
		</div>
	</header>

	<!-- Main Content -->
	<main class="main-content">
		<!-- Left Sidebar - Drone List -->
		<aside class="sidebar">
			<!-- Search -->
			<div class="search-section">
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="ค้นหา Drone"
					class="search-input"
				/>
				<button class="search-btn">🔍</button>
			</div>

			<!-- Drone List Header -->
			<div class="list-header">
				<h2>ทั้งหมด {drones.length} ตัว</h2>
			</div>

			<!-- Drone Cards -->
			<div class="drone-list">
				{#each filteredDrones as drone (drone.id)}
					<div
						class="drone-card"
						class:selected={selectedDrone?.id === drone.id}
						class:connected={drone.status === 'connected'}
						class:disconnected={drone.status === 'disconnected'}
						onclick={() => selectDrone(drone)}
						role="button"
						tabindex="0"
						onkeydown={(e) => e.key === 'Enter' && selectDrone(drone)}
					>
						<div class="drone-header">
							<span class="drone-id">{drone.name}</span>
							<button
								class="status-badge"
								class:connected={drone.status === 'connected'}
								class:disconnected={drone.status === 'disconnected'}
								onclick={(e) => {
									e.stopPropagation();
									toggleDroneStatus(drone.id);
								}}
							>
								{#if drone.status === 'connected'}
									<span class="status-dot connected"></span>
									Connect
								{:else}
									<span class="status-dot disconnected"></span>
									Disconnect
								{/if}
							</button>
						</div>

						<div class="drone-info">
							<div class="info-row">
								<span class="label">ปลายทาง:</span>
								<span class="value">{drone.location}</span>
							</div>
							<div class="info-row">
								<span class="label">ชุดบุกค้า:</span>
								<span class="value">ภาคกลาง, กรุงเทพ</span>
							</div>
							<div class="info-row">
								<span class="label">เวลาทรง:</span>
								<span class="value">{drone.lastUpdate}</span>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</aside>

		<!-- Right Map Section -->
		<section class="map-section">
			<div class="map-container">
				<MapboxMap accessToken={mapboxToken} center={mapCenter} zoom={13} {markers} />

				<!-- Map Legend -->
				<div class="map-legend">
					<div class="legend-item">
						<span class="legend-line gps"></span>
						<span class="legend-text">สัญญาณ GPS</span>
					</div>
					<div class="legend-item">
						<span class="legend-line gps-loss"></span>
						<span class="legend-text">ไม่มีสัญญาณ GPS</span>
					</div>
				</div>
			</div>

			<!-- Selected Drone Info Popup -->
			{#if selectedDrone && selectedDrone.status === 'connected'}
				<div class="drone-popup">
					<button class="close-btn" onclick={() => (selectedDrone = null)}>✕</button>
					<h3>{selectedDrone.id}</h3>
					<div class="popup-info">
						<div class="popup-row">
							<span class="popup-label">ภาคกลาง</span>
						</div>
						<div class="popup-row">
							<span class="popup-label">กรุงเทพ</span>
						</div>
						<div class="popup-row">
							<span class="popup-label">เวลาทรงหนารถอาวุธ:</span>
						</div>
						<div class="popup-row">
							<span class="popup-label">หมิงตอแผน:</span>
						</div>
						<div class="popup-row">
							<span class="popup-label">กรุงเทพ</span>
						</div>
						<div class="popup-row">
							<span class="popup-value-large">{selectedDrone.id}</span>
							<span class="gps-badge" class:good={selectedDrone.gpsStatus === 'good'}>
								GPS {selectedDrone.gpsStatus === 'good' ? 'Connect' : 'Loss'}
							</span>
						</div>
					</div>
				</div>
			{/if}
		</section>
	</main>
</div>

<style>
	:global(body) {
		margin: 0;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
			sans-serif;
	}

	.dashboard {
		width: 100vw;
		height: 100vh;
		display: flex;
		flex-direction: column;
		background: #f5f7fa;
		overflow: hidden;
	}

	/* Header */
	.header {
		background: linear-gradient(135deg, #ff6b6b 0%, #ee5a6f 100%);
		color: white;
		padding: 1.5rem 2rem;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		z-index: 10;
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.logo-section {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.logo {
		font-size: 3rem;
		animation: pulse-rotate 3s ease-in-out infinite;
	}

	@keyframes pulse-rotate {
		0%,
		100% {
			transform: scale(1) rotate(0deg);
		}
		50% {
			transform: scale(1.1) rotate(5deg);
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

	.server-time {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.5rem 1.5rem;
		background: rgba(255, 255, 255, 0.2);
		border-radius: 8px;
	}

	.time-label {
		font-size: 0.85rem;
		opacity: 0.9;
	}

	.time-value {
		font-size: 1.1rem;
		font-weight: 700;
	}

	.status-badge {
		padding: 0.25rem 0.75rem;
		background: rgba(255, 255, 255, 0.3);
		border-radius: 12px;
		font-size: 0.8rem;
		font-weight: 600;
	}

	/* Main Content */
	.main-content {
		display: flex;
		gap: 1.5rem;
		padding: 1.5rem;
		flex: 1;
		overflow: hidden;
	}

	/* Left Sidebar */
	.sidebar {
		width: 30%;
		background: white;
		border-radius: 12px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.search-section {
		padding: 1.25rem;
		border-bottom: 1px solid #e5e7eb;
		display: flex;
		gap: 0.5rem;
	}

	.search-input {
		flex: 1;
		padding: 0.75rem 1rem;
		border: 2px solid #e5e7eb;
		border-radius: 8px;
		font-size: 0.9rem;
		transition: all 0.2s;
	}

	.search-input:focus {
		outline: none;
		border-color: #ff6b6b;
		box-shadow: 0 0 0 3px rgba(255, 107, 107, 0.1);
	}

	.search-btn {
		padding: 0.75rem 1.25rem;
		border: none;
		border-radius: 8px;
		background: #ff6b6b;
		color: white;
		font-size: 1.2rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.search-btn:hover {
		background: #ee5a6f;
	}

	.list-header {
		padding: 1rem 1.25rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.list-header h2 {
		margin: 0;
		font-size: 1.1rem;
		color: #1f2937;
		font-weight: 700;
	}

	/* Drone List */
	.drone-list {
		flex: 1;
		overflow-y: auto;
		padding: 0.75rem;
	}

	.drone-list::-webkit-scrollbar {
		width: 6px;
	}

	.drone-list::-webkit-scrollbar-track {
		background: #f3f4f6;
	}

	.drone-list::-webkit-scrollbar-thumb {
		background: #d1d5db;
		border-radius: 3px;
	}

	/* Drone Card */
	.drone-card {
		display: flex;
		flex-direction: column;
		padding: 1rem;
		border-radius: 8px;
		border: 2px solid #e5e7eb;
		background: white;
		margin-bottom: 0.75rem;
		cursor: pointer;
		transition: all 0.2s;
		text-align: left;
		width: 100%;
	}

	.drone-card:hover {
		background: #f9fafb;
		border-color: #d1d5db;
		transform: translateX(2px);
	}

	.drone-card.selected {
		background: linear-gradient(135deg, #ffeded 0%, #ffe5e5 100%);
		border-color: #ff6b6b;
		box-shadow: 0 2px 8px rgba(255, 107, 107, 0.2);
	}

	.drone-card.connected {
		border-left: 4px solid #10b981;
	}

	.drone-card.disconnected {
		border-left: 4px solid #ef4444;
		opacity: 0.7;
	}

	.drone-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.75rem;
	}

	.drone-id {
		font-size: 1rem;
		font-weight: 700;
		color: #1f2937;
	}

	.drone-header .status-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 0.375rem;
		border: none;
		cursor: pointer;
		transition: all 0.2s;
	}

	.drone-header .status-badge.connected {
		background: #d1fae5;
		color: #065f46;
	}

	.drone-header .status-badge.disconnected {
		background: #fee2e2;
		color: #991b1b;
	}

	.drone-header .status-badge:hover {
		transform: scale(1.05);
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.status-dot.connected {
		background: #10b981;
		animation: pulse-dot 2s ease-in-out infinite;
	}

	.status-dot.disconnected {
		background: #ef4444;
	}

	@keyframes pulse-dot {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	.drone-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.info-row {
		display: flex;
		gap: 0.5rem;
		font-size: 0.85rem;
	}

	.info-row .label {
		color: #6b7280;
		min-width: 80px;
	}

	.info-row .value {
		color: #1f2937;
		font-weight: 500;
	}

	/* Right Map Section */
	.map-section {
		width: 70%;
		background: white;
		border-radius: 12px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
		overflow: hidden;
		position: relative;
	}

	.map-container {
		width: 100%;
		height: 100%;
		position: relative;
	}

	/* Map Legend */
	.map-legend {
		position: absolute;
		bottom: 1.5rem;
		left: 1.5rem;
		background: white;
		padding: 1rem;
		border-radius: 8px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		z-index: 5;
	}

	.legend-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.5rem;
	}

	.legend-item:last-child {
		margin-bottom: 0;
	}

	.legend-line {
		width: 30px;
		height: 3px;
		border-radius: 2px;
	}

	.legend-line.gps {
		background: #10b981;
	}

	.legend-line.gps-loss {
		background: #ef4444;
		border: 2px dashed #ef4444;
		height: 0;
	}

	.legend-text {
		font-size: 0.85rem;
		color: #374151;
	}

	/* Drone Popup */
	.drone-popup {
		position: absolute;
		top: 1.5rem;
		right: 1.5rem;
		width: 300px;
		background: white;
		border-radius: 12px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
		padding: 1.5rem;
		z-index: 10;
		border: 2px solid #ff6b6b;
	}

	.close-btn {
		position: absolute;
		top: 0.75rem;
		right: 0.75rem;
		width: 28px;
		height: 28px;
		border: none;
		background: #f3f4f6;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1rem;
		color: #6b7280;
		transition: all 0.2s;
	}

	.close-btn:hover {
		background: #e5e7eb;
		color: #1f2937;
	}

	.drone-popup h3 {
		margin: 0 0 1rem 0;
		font-size: 1.5rem;
		color: #ff6b6b;
		font-weight: 700;
	}

	.popup-info {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.popup-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.9rem;
	}

	.popup-label {
		color: #6b7280;
	}

	.popup-value-large {
		font-size: 1.25rem;
		font-weight: 700;
		color: #1f2937;
	}

	.gps-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		margin-left: auto;
	}

	.gps-badge.good {
		background: #d1fae5;
		color: #065f46;
	}

	.gps-badge:not(.good) {
		background: #fee2e2;
		color: #991b1b;
	}

	/* Responsive */
	@media (max-width: 1400px) {
		.sidebar {
			width: 35%;
		}
		.map-section {
			width: 65%;
		}
	}

	@media (max-width: 1024px) {
		.main-content {
			flex-direction: column;
		}
		.sidebar,
		.map-section {
			width: 100%;
		}
		.map-section {
			height: 400px;
		}
	}
</style>
