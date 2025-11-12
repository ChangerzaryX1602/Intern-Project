<script lang="ts">
	import { onMount } from 'svelte';
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import { env } from '$env/dynamic/public';

	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';

	// Mock data for drones
	interface Drone {
		id: string; // Unique ID for each record
		trackingId: string; // Tracking ID for grouping (can be duplicate)
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
			id: 'A01-1', // Unique ID
			trackingId: 'A01', // Tracking ID for grouping
			name: 'Drone ID: A01',
			status: 'connected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: { lat: 13.7563, lng: 100.5018 }, // กรุงเทพ
			gpsStatus: 'good',
			lastUpdate: '11/11/2025 18:16 น.'
		},
		{
			id: 'A01-2',
			trackingId: 'A01',
			name: 'Drone ID: A01',
			status: 'connected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: { lat: 13.7663, lng: 100.5218 }, // เคลื่อนไปทางเหนือ-ตะวันออก
			gpsStatus: 'good',
			lastUpdate: '11/11/2025 18:17 น.'
		},
		{
			id: 'A01-3',
			trackingId: 'A01',
			name: 'Drone ID: A01',
			status: 'connected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: { lat: 13.7463, lng: 100.5418 }, // เคลื่อนไปทางตะวันออก
			gpsStatus: 'good',
			lastUpdate: '11/11/2025 18:18 น.'
		},
		{
			id: 'A02-1',
			trackingId: 'A02',
			name: 'Drone ID: A02',
			status: 'connected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: { lat: 13.7463, lng: 100.4918 },
			gpsStatus: 'good',
			lastUpdate: '11/11/2025 18:16 น.'
		},
		{
			id: 'A03-1',
			trackingId: 'A03',
			name: 'Drone ID: A03',
			status: 'disconnected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: { lat: 13.7263, lng: 100.5118 },
			gpsStatus: 'loss',
			lastUpdate: '11/11/2025 20:57 น.'
		}
	]);

	let selectedDrone = $state<Drone | null>(null);
	let searchQuery = $state('');
	let mapCenter: [number, number] = $state([100.5018, 13.7563]); // [lng, lat] - กรุงเทพ

	// Generate markers for connected drones
	let markers = $derived(
		drones
			.filter((d) => d.status === 'connected')
			.map((d) => ({
				id: d.trackingId, // Use trackingId for line drawing
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

<div class="w-screen h-screen flex flex-col bg-gray-100 overflow-hidden">
	<!-- Header -->
	<header class="bg-gradient-to-br from-red-500 to-red-600 text-white px-8 py-6 shadow-lg z-10">
		<div class="flex justify-between items-center">
			<div class="flex items-center gap-4">
				<div>
					<h1 class="m-0 text-3xl font-bold">Offensive Dashboard</h1>
					<p class="m-0 opacity-90 text-sm">Drone Control & Monitoring</p>
				</div>
			</div>

			<div class="flex items-center gap-4 px-6 py-2 bg-white/20 rounded-lg">
				<span class="text-sm opacity-90">Server Time:</span>
				<span class="text-lg font-bold">11/11/2025 19:02</span>
				<span class="px-3 py-1 bg-white/30 rounded-xl text-xs font-semibold">🔴 LIVE</span>
			</div>
		</div>
	</header>

	<!-- Main Content -->
	<main class="flex gap-6 px-6 pt-6 flex-1 overflow-hidden">
		<!-- Left Sidebar - Drone List -->
		<aside class="w-[30%] bg-white rounded-xl shadow-md flex flex-col overflow-hidden">
			<!-- Search -->
			<SearchBox
				bind:value={searchQuery}
				placeholder="ค้นหา Drone"
				label="🔍 ค้นหา Drone"
				onSearch={() => {}}
			/>

			<!-- Drone List Header -->
			<div class="px-5 py-4 border-b border-gray-200">
				<h2 class="m-0 text-lg text-gray-800 font-bold">ทั้งหมด {drones.length} ตัว</h2>
			</div>

			<!-- Drone Cards -->
			<div class="flex-1 overflow-y-auto p-3 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 hover:scrollbar-thumb-gray-400">
				{#each filteredDrones as drone (drone.id)}
					<div
						class="flex flex-col p-4 rounded-lg border-2 bg-white mb-3 cursor-pointer transition-all duration-200 hover:bg-gray-50 hover:border-gray-300 hover:translate-x-0.5"
						class:drone-selected={selectedDrone?.id === drone.id}
						class:border-gray-200={selectedDrone?.id !== drone.id}
						class:border-l-4={drone.status === 'connected' || drone.status === 'disconnected'}
						class:border-l-green-500={drone.status === 'connected'}
						class:border-l-red-500={drone.status === 'disconnected'}
						class:opacity-70={drone.status === 'disconnected'}
						onclick={() => selectDrone(drone)}
						role="button"
						tabindex="0"
						onkeydown={(e) => e.key === 'Enter' && selectDrone(drone)}
					>
						<div class="flex justify-between items-center mb-3">
							<span class="text-base font-bold text-gray-800">{drone.name}</span>
							<button
								class="px-3 py-1 rounded-xl text-xs font-semibold flex items-center gap-1.5 border-none cursor-pointer transition-all duration-200 hover:scale-105"
								class:bg-green-100={drone.status === 'connected'}
								class:text-green-800={drone.status === 'connected'}
								class:bg-red-100={drone.status === 'disconnected'}
								class:text-red-800={drone.status === 'disconnected'}
								onclick={(e) => {
									e.stopPropagation();
									toggleDroneStatus(drone.id);
								}}
							>
								<span
									class="w-2 h-2 rounded-full"
									class:bg-green-500={drone.status === 'connected'}
									class:animate-pulse-slow={drone.status === 'connected'}
									class:bg-red-500={drone.status === 'disconnected'}
								></span>
								{#if drone.status === 'connected'}
									Connect
								{:else}
									Disconnect
								{/if}
							</button>
						</div>

						<div class="flex flex-col gap-1">
							<div class="flex gap-2 text-sm">
								<span class="text-gray-600 min-w-[80px]">ปลายทาง:</span>
								<span class="text-gray-800 font-medium">{drone.location}</span>
							</div>
							<div class="flex gap-2 text-sm">
								<span class="text-gray-600 min-w-[80px]">ชุดบุกค้า:</span>
								<span class="text-gray-800 font-medium">ภาคกลาง, กรุงเทพ</span>
							</div>
							<div class="flex gap-2 text-sm">
								<span class="text-gray-600 min-w-[80px]">เวลาทรง:</span>
								<span class="text-gray-800 font-medium">{drone.lastUpdate}</span>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</aside>

		<!-- Right Map Section -->
		<section class="w-[70%] bg-white rounded-xl shadow-md overflow-hidden relative">
			<div class="w-full h-full relative">
				<MapboxMap accessToken={mapboxToken} center={mapCenter} zoom={13} {markers} drawLines={true} />

				<!-- Map Legend -->
				<div class="absolute bottom-6 left-6 bg-white px-4 py-3 rounded-lg shadow-md z-[5]">
					<div class="flex items-center gap-2 mb-2">
						<span class="w-3 h-3 rounded-full border-2 border-white shadow-md bg-green-500"></span>
						<span class="text-sm text-gray-700">สัญญาณ GPS</span>
					</div>
					<div class="flex items-center gap-2 mb-2">
						<span class="w-3 h-3 rounded-full border-2 border-white shadow-md bg-red-500"></span>
						<span class="text-sm text-gray-700">ไม่มีสัญญาณ GPS</span>
					</div>
					<div class="flex items-center gap-2">
						<span class="w-[30px] h-[3px] rounded bg-green-500"></span>
						<span class="text-sm text-gray-700">เส้นทางเคลื่อนที่ (ID ซ้ำ)</span>
					</div>
				</div>
			</div>

			<!-- Selected Drone Info Popup -->
			{#if selectedDrone && selectedDrone.status === 'connected'}
				<div class="absolute top-6 right-6 w-[300px] bg-white rounded-xl shadow-xl p-6 z-10 border-2 border-red-500">
					<button
						class="absolute top-3 right-3 w-7 h-7 border-none bg-gray-100 rounded-full cursor-pointer flex items-center justify-center text-base text-gray-600 transition-all duration-200 hover:bg-gray-200 hover:text-gray-800"
						onclick={() => (selectedDrone = null)}
					>
						✕
					</button>
					<h3 class="m-0 mb-4 text-2xl text-red-500 font-bold">{selectedDrone.id}</h3>
					<div class="flex flex-col gap-2">
						<div class="flex items-center gap-2 text-sm">
							<span class="text-gray-600">ภาคกลาง</span>
						</div>
						<div class="flex items-center gap-2 text-sm">
							<span class="text-gray-600">กรุงเทพ</span>
						</div>
						<div class="flex items-center gap-2 text-sm">
							<span class="text-gray-600">เวลาทรงหนารถอาวุธ:</span>
						</div>
						<div class="flex items-center gap-2 text-sm">
							<span class="text-gray-600">หมิงตอแผน:</span>
						</div>
						<div class="flex items-center gap-2 text-sm">
							<span class="text-gray-600">กรุงเทพ</span>
						</div>
						<div class="flex items-center gap-2 text-sm">
							<span class="text-xl font-bold text-gray-800">{selectedDrone.id}</span>
							<span
								class="px-3 py-1 rounded-xl text-xs font-semibold ml-auto"
								class:bg-green-100={selectedDrone.gpsStatus === 'good'}
								class:text-green-800={selectedDrone.gpsStatus === 'good'}
								class:bg-red-100={selectedDrone.gpsStatus !== 'good'}
								class:text-red-800={selectedDrone.gpsStatus !== 'good'}
							>
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
	.drone-selected {
		background: linear-gradient(135deg, #ffeded 0%, #ffe5e5 100%);
		border-color: #ef4444 !important;
		box-shadow: 0 2px 8px rgba(239, 68, 68, 0.2);
	}

	@keyframes pulse-slow {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	.animate-pulse-slow {
		animation: pulse-slow 2s ease-in-out infinite;
	}
</style>
