<script lang="ts">
	import { onMount } from 'svelte';
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import { env } from '$env/dynamic/public';

	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';

	// Mock data for drones
	interface Drone {
		id: string;
		name: string;
		status: 'connected' | 'disconnected';
		startDate: string;
		startLocation: string;
		startCoordinates: { lat: number; lng: number };
		endDate: string;
		endLocation: string;
		endCoordinates: { lat: number; lng: number };
	}

	// Mock data for cameras
	interface Camera {
		id: string;
		name: string;
		status: 'online' | 'offline';
		location: string;
		coordinates: { lat: number; lng: number };
	}

	// Mock data for detections
	interface Detection {
		id: string;
		cameraId: string;
		cameraName: string;
		droneId: string;
		detectedAt: string;
		coordinates: { lat: number; lng: number };
		imageUrl?: string;
	}

	let drones = $state<Drone[]>([
		{
			id: 'A01',
			name: 'Drone ID: A01',
			status: 'connected',
			startDate: '11/11/2025 18:00 น.',
			startLocation: 'ภาคกลาง, กรุงเทพ',
			startCoordinates: { lat: 13.7563, lng: 100.5018 },
			endDate: '11/11/2025 19:30 น.',
			endLocation: 'ภาคกลาง, นนทบุรี',
			endCoordinates: { lat: 13.8621, lng: 100.5144 }
		},
		{
			id: 'A02',
			name: 'Drone ID: A02',
			status: 'disconnected',
			startDate: '11/11/2025 17:00 น.',
			startLocation: 'ภาคกลาง, ปทุมธานี',
			startCoordinates: { lat: 13.9564, lng: 100.5265 },
			endDate: '11/11/2025 18:45 น.',
			endLocation: 'ภาคกลาง, สมุทรปราการ',
			endCoordinates: { lat: 13.5990, lng: 100.5998 }
		}
	]);

	let cameras = $state<Camera[]>([
		{
			id: 'CAM-001',
			name: 'Camera #001',
			status: 'online',
			location: 'ภาคกลาง, กรุงเทพ - ถนนสุขุมวิท',
			coordinates: { lat: 13.7563, lng: 100.5018 }
		},
		{
			id: 'CAM-002',
			name: 'Camera #002',
			status: 'online',
			location: 'ภาคกลาง, กรุงเทพ - สยามสแควร์',
			coordinates: { lat: 13.7465, lng: 100.5348 }
		},
		{
			id: 'CAM-003',
			name: 'Camera #003',
			status: 'offline',
			location: 'ภาคกลาง, นนทบุรี',
			coordinates: { lat: 13.8621, lng: 100.5144 }
		}
	]);

	let detections = $state<Detection[]>([
		{
			id: 'DET-001',
			cameraId: 'CAM-001',
			cameraName: 'Camera #001',
			droneId: 'A01',
			detectedAt: '11/11/2025 18:30 น.',
			coordinates: { lat: 13.7563, lng: 100.5018 }
		},
		{
			id: 'DET-002',
			cameraId: 'CAM-002',
			cameraName: 'Camera #002',
			droneId: 'A01',
			detectedAt: '11/11/2025 19:00 น.',
			coordinates: { lat: 13.7465, lng: 100.5348 }
		}
	]);

	let droneSearchQuery = $state('');
	let cameraSearchQuery = $state('');
	let selectedDrone = $state<Drone | null>(null);
	let selectedCamera = $state<Camera | null>(null);

	let droneMapCenter: [number, number] = $state([100.5018, 13.7563]);
	let cameraMapCenter: [number, number] = $state([100.5018, 13.7563]);

	// Generate markers for drones
	let droneMarkers = $derived(
		drones
			.filter((d) => d.status === 'connected')
			.map((d) => ({
				id: d.id,
				lngLat: [d.startCoordinates.lng, d.startCoordinates.lat] as [number, number],
				popup: `<div style="font-size:13px"><strong>${d.name}</strong><br/>${d.startLocation}</div>`,
				color: '#10b981'
			}))
	);

	// Generate markers for cameras
	let cameraMarkers = $derived(
		cameras.map((c) => ({
			id: c.id,
			lngLat: [c.coordinates.lng, c.coordinates.lat] as [number, number],
			popup: `<div style="font-size:13px"><strong>${c.name}</strong><br/>${c.location}</div>`,
			color: c.status === 'online' ? '#3b82f6' : '#ef4444'
		}))
	);

	// Filter drones and cameras based on search
	let filteredDrones = $derived(
		drones.filter((d) => d.name.toLowerCase().includes(droneSearchQuery.toLowerCase()))
	);

	let filteredCameras = $derived(
		cameras.filter((c) =>
			c.name.toLowerCase().includes(cameraSearchQuery.toLowerCase()) ||
			c.location.toLowerCase().includes(cameraSearchQuery.toLowerCase())
		)
	);

	function toggleDroneStatus(droneId: string) {
		const drone = drones.find((d) => d.id === droneId);
		if (drone) {
			drone.status = drone.status === 'connected' ? 'disconnected' : 'connected';
		}
	}

	function selectDrone(drone: Drone) {
		selectedDrone = drone;
		droneMapCenter = [drone.startCoordinates.lng, drone.startCoordinates.lat];
	}

	function selectCamera(camera: Camera) {
		selectedCamera = camera;
		cameraMapCenter = [camera.coordinates.lng, camera.coordinates.lat];
	}

	onMount(() => {
		console.log('Battle Dashboard mounted');
	});
</script>

<svelte:head>
	<title>Battle Dashboard - Drone & Camera Monitor</title>
</svelte:head>

<div class="w-screen h-screen flex flex-col bg-gray-100 overflow-hidden">
	<!-- Header -->
	<header class="bg-gradient-to-br from-purple-600 to-indigo-600 text-white px-6 py-3 shadow-lg z-10">
		<div class="flex justify-between items-center">
			<div>
				<h1 class="m-0 text-xl font-bold">Battle Dashboard</h1>
				<p class="m-0 opacity-90 text-xs">Drone & Camera Monitoring System</p>
			</div>
			<div class="flex items-center gap-3 px-4 py-1.5 bg-white/20 rounded-lg">
				<span class="text-xs opacity-90">Server Time:</span>
				<span class="text-sm font-bold">11/11/2025 19:02</span>
				<span class="px-2 py-0.5 bg-white/30 rounded-xl text-xs font-semibold">🟢 LIVE</span>
			</div>
		</div>
	</header>

	<!-- Main Content -->
	<main class="flex gap-3 px-4 py-3 flex-1 overflow-hidden">
		<!-- Left Side - Drones -->
		<div class="w-1/2 flex flex-col gap-2 overflow-hidden">
			<!-- Drone Search & List -->
			<div class="bg-white rounded-xl shadow-md flex flex-col overflow-hidden" style="height: 22vh;">
				<div class="px-4 py-2 border-b border-gray-200 bg-gradient-to-br from-purple-50 to-purple-100">
					<div class="flex justify-between items-center mb-2">
						<h2 class="m-0 text-base text-gray-800 font-bold flex items-center gap-2">
							<span>🚁</span>
							Offense - Drones
							<span class="bg-purple-500 text-white px-2 py-0.5 rounded-xl text-xs font-semibold">
								{drones.length}
							</span>
						</h2>
    
                        <input
						type="text"
						bind:value={droneSearchQuery}
						placeholder="ค้นหา Drone..."
						class="w-1/2 px-3 py-1.5 border-2 border-gray-200 rounded-lg text-xs transition-all duration-200 bg-white focus:outline-none focus:border-purple-500"
					/>
					</div>
			
				</div>

				<div class="flex-1 flex gap-2 overflow-y-auto p-2 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100">
					{#each filteredDrones as drone (drone.id)}
						<div
							class="flex flex-col justify-between items-center p-2 rounded-lg border-2 border-gray-200 bg-white mb-1.5 cursor-pointer transition-all duration-200 hover:bg-gray-50"
							role="button"
							tabindex="0"
							onclick={() => selectDrone(drone)}
							onkeydown={(e) => e.key === 'Enter' && selectDrone(drone)}
						>
							<div class="flex items-center">
								<span class="text-sm font-bold text-gray-800">{drone.name}</span>
							</div>
							<button
								class="px-2 py-0.5 rounded-xl text-xs font-semibold"
								class:bg-green-100={drone.status === 'connected'}
								class:text-green-800={drone.status === 'connected'}
								class:bg-red-100={drone.status === 'disconnected'}
								class:text-red-800={drone.status === 'disconnected'}
								onclick={(e) => {
									e.stopPropagation();
									toggleDroneStatus(drone.id);
								}}
							>
								{drone.status === 'connected' ? '🟢 Connected' : '🔴 Disconnected'}
							</button>
						</div>
					{/each}
				</div>
			</div>

			<!-- Drone Map -->
			<div class="bg-white rounded-xl shadow-md overflow-hidden flex-1">
				<MapboxMap
					accessToken={mapboxToken}
					center={droneMapCenter}
					zoom={12}
					markers={droneMarkers}
					drawLines={true}

				/>
			</div>
            <div class="flex gap-3 text-xs">
                <span class="text-blue-500">● มี GPS</span>
                <span class="text-red-500">● ไม่มี GPS</span>
            </div>

			<!-- Drone History -->
			<div class="bg-white rounded-xl shadow-md p-3 overflow-y-auto scrollbar-thin" style="height: 18vh;">
				<h2 class="m-0 mb-2 text-sm text-gray-800 font-bold flex items-center gap-2">
					<span>📜</span>
					ประวัติการเดินทาง
				</h2>
				{#each drones as drone (drone.id)}
					<div class="p-2 mb-2 border-2 border-gray-200 rounded-lg">
						<h3 class="m-0 mb-1.5 text-xs font-bold text-purple-600">{drone.name}</h3>
						<div class="flex mb-1.5">
							<div class="w-1/4 text-xs text-gray-600 font-semibold">เริ่มบิน:</div>
							<div class="w-3/4">
								<div class="text-xs text-gray-800 font-medium">{drone.startDate}</div>
								<div class="text-xs text-gray-600">{drone.startLocation}</div>
							</div>
						</div>
						<div class="flex">
							<div class="w-1/4 text-xs text-gray-600 font-semibold">ปลายทาง:</div>
							<div class="w-3/4">
								<div class="text-xs text-gray-800 font-medium">{drone.endDate}</div>
																<div class="text-xs text-gray-600">{drone.endLocation}</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Right Side - Cameras -->
		<div class="w-1/2 flex flex-col gap-2 overflow-hidden">
			<!-- Camera Search & List -->
			 <h1 class="text-center text-5xl">Deffense</h1>
			<div class="bg-white rounded-xl shadow-md flex flex-col overflow-hidden" style="height: 22vh;">
				<div class="px-4 py-2 border-b border-gray-200 bg-gradient-to-br from-blue-50 to-blue-100">
					<div class="flex justify-between items-center mb-2">
						<h2 class="m-0 text-base text-gray-800 font-bold flex items-center gap-2">
							<span>📹</span>
							Defense - Cameras
							<span class="bg-blue-500 text-white px-2 py-0.5 rounded-xl text-xs font-semibold">
								{cameras.length}
							</span>
						</h2>
                        <input
						type="text"
						bind:value={cameraSearchQuery}
						placeholder="ค้นหากล้อง..."
						class="w-1/2 px-3 py-1.5 border-2 border-gray-200 rounded-lg text-xs transition-all duration-200 bg-white focus:outline-none focus:border-blue-500"
					/>
					</div>
					
				</div>

				<div class="flex-1 flex gap-2 overflow-y-auto p-2 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100">
					{#each filteredCameras as camera (camera.id)}
						<div
							class="flex flex-col p-2 rounded-lg border-2 border-gray-200 bg-white mb-1.5 cursor-pointer transition-all duration-200 hover:bg-gray-50"
							role="button"
							tabindex="0"
							onclick={() => selectCamera(camera)}
							onkeydown={(e) => e.key === 'Enter' && selectCamera(camera)}
						>
							<div class="flex justify-between items-center mb-0.5">
								<span class="text-sm font-bold text-gray-800">{camera.id}</span>
								<span
									class="w-1.5 h-1.5 rounded-full"
									class:bg-green-500={camera.status === 'online'}
									class:bg-red-500={camera.status === 'offline'}
								></span>
							</div>
							<div class="text-xs text-gray-600 flex items-center gap-1">
								<span>📍</span>
								{camera.location}
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- Camera Map -->
			<div class="bg-white rounded-xl shadow-md overflow-hidden flex-1">
				<MapboxMap
					accessToken={mapboxToken}
					center={cameraMapCenter}
					zoom={12}
					markers={cameraMarkers}
				/>
			</div>

			<!-- Detection History -->
			<div class="bg-white rounded-xl shadow-md p-3 overflow-y-auto scrollbar-thin" style="height: 18vh;">
				<h2 class="m-0 mb-2 text-sm text-gray-800 font-bold flex items-center gap-2">
					<span>🎯</span>
					ประวัติการตรวจจับ
				</h2>
				{#each detections as detection (detection.id)}
					<div class="flex gap-2 p-2 mb-2 border-2 border-gray-200 rounded-lg">
						<div class="w-16 h-16 shrink-0 bg-gray-200 rounded-lg flex items-center justify-center">
							<span class="text-2xl">📷</span>
						</div>
						<div class="flex-1">
							<h3 class="m-0 mb-1 text-xs font-bold text-blue-600">
								{detection.cameraName}
								<span class="text-xs text-gray-500 font-normal ml-1">พบเมื่อ {detection.detectedAt}</span>
							</h3>
							<div class="flex mb-0.5">
								<div class="w-1/4 text-xs text-gray-600 font-semibold">Drone ID:</div>
								<div class="w-3/4">
									<div class="text-xs text-gray-800 font-medium">{detection.droneId}</div>
									<div class="text-xs text-gray-500">
										{detection.coordinates.lat.toFixed(4)}, {detection.coordinates.lng.toFixed(4)}
									</div>
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</main>
</div>

<style>
	@keyframes pulse-slow {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

</style>