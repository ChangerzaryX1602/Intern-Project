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
		distance?: number; // in km
		avgSpeed?: number; // in km/h
		avgAcceleration?: number; // in m/s²
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
			endCoordinates: { lat: 13.8621, lng: 100.5144 },
			distance: 12.5,
			avgSpeed: 25.0,
			avgAcceleration: 1.5
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
			endCoordinates: { lat: 13.5990, lng: 100.5998 },
			distance: 18.3,
			avgSpeed: 21.5,
			avgAcceleration: 1.2
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
	let showHistory = $state(false);

	let droneMapCenter: [number, number] = $state([100.5018, 13.7563]);
	let cameraMapCenter: [number, number] = $state([100.5018, 13.7563]);

	// Real-time server time
	let currentTime = $state(new Date());

	$effect(() => {
		const interval = setInterval(() => {
			currentTime = new Date();
		}, 1000);

		return () => clearInterval(interval);
	});

	function formatDateTime(date: Date): string {
		return date.toLocaleString('th-TH', {
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit',
			hour12: false
		});
	}

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

		// keep the displayed time up-to-date
		const t = setInterval(() => {
			currentTime = new Date();
		}, 1000);

		return () => clearInterval(t);
	});
</script>

<svelte:head>
	<title>Battle Dashboard - Drone & Camera Monitor</title>
</svelte:head>

<div class="w-screen h-screen flex flex-col bg-gray-100 overflow-hidden">
	<!-- Header -->
    <header class="bg-gradient-to-br from-indigo-500 to-purple-600 text-white px-8 py-2 shadow-lg z-10">
        <div class="flex justify-between items-center gap-8 max-w-full">
            <div class="flex items-center gap-4">
                <!-- <div class="text-5xl animate-bounce-slow">🛡️</div> -->
                <div>
                    <h1 class="m-0 text-xl font-bold">Battle Dashboard</h1>
                    <p class="m-0 opacity-90 text-xs">Drone & Camera Monitoring System</p>
                </div>
            </div>

            <div class="flex items-center gap-6">
                <!-- Server Time - Real-time display -->
                <div class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium">
                    <span class="text-lg">🕐</span>
                    <div>
                        <div class="text-xs opacity-75">Server Time</div>
                        <div class="font-mono font-bold">{formatDateTime(currentTime)}</div>
                    </div>
                </div>

                <!-- Commander Info -->
                <div class="flex items-center gap-2 px-4 py-3 bg-white/20 rounded-lg text-sm font-medium">
                    <span class="text-lg">👨‍💼</span>
                    <span>commander</span>
                </div>
            </div>
        </div>
    </header>

	<!-- Main Content -->
	<main class="flex gap-3 px-4 py-3 flex-1 overflow-hidden">
		<!-- Left Side - Drones -->
		<div class="w-1/2 flex flex-col gap-2 overflow-hidden">
			<!-- Drone Search & List -->
			<div class="bg-white rounded-xl shadow-md flex flex-col overflow-hidden">
				<div class="px-4 py-2 border-b border-gray-200 bg-gradient-to-br from-purple-50 to-purple-100">
					<div class="flex justify-between items-center mb-2">
						<h2 class="m-0 text-base text-gray-800 font-bold flex items-center gap-2">
							<span>🚁</span>
							Offense
							<span class="bg-purple-500 text-white px-2 py-0.5 rounded-xl text-xs font-semibold">
								Drones {drones.length} ตัว
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

				<div class="flex-1 flex gap-2 overflow-y-auto p-2 list-scrollbar">
					{#each filteredDrones as drone (drone.id)}
						<div
							class="flex flex-col text-nowrap justify-between items-center p-2 rounded-lg border-2 border-gray-200 bg-white mb-1 cursor-pointer transition-all duration-200 hover:bg-gray-50"
							role="button"
							tabindex="0"
							onclick={() => selectDrone(drone)}
							onkeydown={(e) => e.key === 'Enter' && selectDrone(drone)}
						>
							<div class="flex items-center mb-2">
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
			<div class="bg-white rounded-xl shadow-md overflow-hidden flex-1 relative">
				<MapboxMap
					accessToken={mapboxToken}
					center={droneMapCenter}
					zoom={12}
					markers={droneMarkers}
					drawLines={true}
				/>
				<div class="absolute bottom-3 right-3 bg-white px-3 py-2 rounded-lg shadow-md flex gap-3 text-xs">
					<span class="text-blue-500">● มี GPS</span>
					<span class="text-red-500">● ไม่มี GPS</span>
				</div>
			</div>

			<!-- Drone History -->
			<div class="bg-white rounded-xl shadow-md p-3 overflow-y-auto scrollbar-thin" style="height: 35vh;">
				<h2 class="m-0 mb-2 text-sm text-gray-800 font-bold flex items-center gap-2">
					<span>📜</span>
					{selectedDrone ? `ประวัติการเดินทาง - ${selectedDrone.name}` : 'ประวัติการเดินทาง'}
				</h2>
				{#if selectedDrone}
					<div class="p-2 mb-2 border-2 border-gray-200 rounded-lg">
						<h3 class="m-0 mb-2 text-xs font-bold text-purple-600">{selectedDrone.name}</h3>
						<div class="flex mb-2">
							<div class="w-1/4 text-xs text-gray-600 font-semibold">เริ่มบิน:</div>
							<div class="w-3/4 flex gap-8">
								<div class="text-xs text-gray-800 font-medium">{selectedDrone.startDate}</div>
								<div class="text-xs text-gray-500">📍 {selectedDrone.startCoordinates.lat.toFixed(4)}, {selectedDrone.startCoordinates.lng.toFixed(4)}</div>
							</div>
						</div>
						<div class="flex mb-2">
							<div class="w-1/4 text-xs text-gray-600 font-semibold">ปลายทาง:</div>
							<div class="w-3/4 flex gap-8">
								<div class="text-xs text-gray-800 font-medium">{selectedDrone.endDate}</div>
								<div class="text-xs text-gray-500">📍 {selectedDrone.endCoordinates.lat.toFixed(4)}, {selectedDrone.endCoordinates.lng.toFixed(4)}</div>
							</div>
						</div>
                        <div class="flex mb-2">
							<div class="w-1/4 text-xs text-gray-600 font-semibold">ระยะทาง:</div>
							<div class="w-3/4">
								<div class="text-xs text-gray-600">{selectedDrone.distance ?? 'N/A'} km</div>
							</div>
						</div>
                        <div class="flex mb-2">
							<div class="w-1/4 text-xs text-gray-600 font-semibold">ความเร็วเฉลี่ย:</div>
							<div class="w-3/4">
								<div class="text-xs text-gray-600">{selectedDrone.avgSpeed ?? 'N/A'} km/h</div>
							</div>
						</div>
                        <div class="flex">
							<div class="w-1/4 text-xs text-gray-600 font-semibold">ความเร่งเฉลี่ย:</div>
							<div class="w-3/4">
								<div class="text-xs text-gray-600">{selectedDrone.avgAcceleration ?? 'N/A'} m/s²</div>
							</div>
						</div>
					</div>
				{:else}
					<div class="flex items-center justify-center h-32 text-center text-gray-400">
						<p class="text-sm">เลือก Drone เพื่อดูประวัติการเดินทาง</p>
					</div>
				{/if}
			</div>
		</div>

		<!-- Right Side - Cameras -->
		<div class="w-1/2 flex flex-col gap-2 overflow-hidden">
			<!-- Camera Search & List -->
			<div class="bg-white rounded-xl shadow-md flex flex-col overflow-hidden">
				<div class="px-4 py-2 border-b border-gray-200 bg-gradient-to-br from-blue-50 to-blue-100">
					<div class="flex justify-between items-center mb-2">
						<h2 class="m-0 text-base text-gray-800 font-bold flex items-center gap-2">
							<span>📹</span>
							Defense
							<span class="bg-blue-500 text-white px-2 py-0.5 rounded-xl text-xs font-semibold">
								Cameras {cameras.length} ตัว
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

				<div class="flex-1 flex gap-2 text-nowrap overflow-y-auto p-2 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 list-scrollbar">
					{#each filteredCameras as camera (camera.id)}
						<div
							class="flex flex-col p-2 rounded-lg border-2 border-gray-200 bg-white mb-1.5 cursor-pointer transition-all duration-200 hover:bg-gray-50"
							role="button"
							tabindex="0"
							onclick={() => selectCamera(camera)}
							onkeydown={(e) => e.key === 'Enter' && selectCamera(camera)}
						>
							<div class="flex justify-between items-center mb-2">
								<span class="text-sm font-bold text-gray-800">{camera.id}</span>
								<span
									class="w-1.5 h-1.5 rounded-full"
									class:bg-green-500={camera.status === 'online'}
									class:bg-red-500={camera.status === 'offline'}
								></span>
							</div>
							<div class="text-xs text-gray-600 flex items-center gap-1">
								<span>📍</span>
								{camera.coordinates.lat.toFixed(4)}, {camera.coordinates.lng.toFixed(4)}
							</div>
						</div>
					{/each}
				</div>
			</div>

            <div class="flex w-full gap-2 overflow-hidden flex-1">
                <div class="w-2/3 flex flex-col gap-2">
                    <!-- Camera Map -->
                    <div class="bg-white rounded-xl shadow-md overflow-hidden flex-1">
                        <MapboxMap
                            accessToken={mapboxToken}
                            center={cameraMapCenter}
                            zoom={12}
                            markers={cameraMarkers}
                        />
                    </div>
                    <div class="bg-white rounded-xl shadow-md overflow-hidden flex-1">
                        Stream
                    </div>
                </div>
                <!-- Detection History -->
                <div class="w-1/3 bg-white rounded-xl shadow-md p-3">
                    <!-- Model Upload Section -->
                    <div class="w-full">
                        <div class="mb-2">
                            <div class="flex items-center gap-2">
                                <span class="text-2xl">🤖</span>
                                <div>
                                    <h3 class="font-bold text-gray-800 text-sm">Current Model</h3>
                                    <p class="text-xs text-gray-600">YOLO v8 NCNN Model 960</p>
                                </div>
                            </div>
                        </div>
                        <div class="relative">
                            <input 
                                id="file-upload" 
                                type="file" 
                                accept=".pt,.onnx,.pb,.tflite"
                                class="absolute inset-0 w-full h-full opacity-0 cursor-pointer" 
                            />
                            <div class="flex items-center justify-center gap-2 px-2 py-2 bg-indigo-50 border-2 border-dashed border-indigo-300 rounded-lg hover:bg-indigo-100 hover:border-indigo-400 transition-all cursor-pointer">
                                <span class="text-lg">📦</span>
                                <span class="text-sm font-medium text-indigo-700">Upload New Model</span>
                            </div>
                            <p class="text-xs text-gray-500 mt-1 text-center">Supported: .pt, .onnx, .pb, .tflite</p>
                        </div>
                    </div>

                    <div class="flex flex-col gap-1 border-y border-gray-200 pb-4 my-2">
                        <div class="flex-1 mt-1">
                            <label for="start-date" class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
                            <input id="start-date" type="date" class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm" />
                        </div>
                        <div class="flex-1">
                            <label for="end-date" class="block text-sm font-medium text-gray-700 mb-1">End Date</label>
                            <input id="end-date" type="date" class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm" />
                        </div>
                        <button 
                            onclick={() => showHistory = !showHistory}
                            class="w-full mt-2 px-4 py-2.5 bg-gradient-to-r from-green-400 to-emerald-400 text-white font-semibold text-sm rounded-lg shadow-md hover:from-green-600 hover:to-emerald-700 transition-all duration-200 active:scale-95"
                        >
                            {showHistory ? '📜 Back to Feed' : '🔍 Search History'}
                        </button>
                    </div>

                    {#if !showHistory}
                        <h2 class="m-0 mb-2 text-sm text-gray-800 font-bold flex items-center gap-2">
                            <span>🎯</span>
                            Detection Feed
                        </h2>
                        {#each detections as detection (detection.id)}
                            <div class="flex gap-2 p-2 mb-2 border-2 border-gray-200 rounded-lg">
                                <div class="w-16 h-16 shrink-0 bg-gray-200 rounded-lg flex items-center justify-center">
                                    <span class="text-2xl">📷</span>
                                </div>
                                <div class="flex-1">
                                    <h3 class="m-0 mb-1 text-xs font-bold text-blue-600">{detection.cameraName}</h3>
                                    <h3 class="text-xs text-gray-500 font-normal ml-1">{detection.detectedAt}</h3>
                                    <h3>ตรวจพบโดรน 3</h3>
                                </div>
                            </div>
                        {/each}
                    {:else}
                        <h2 class="m-0 mb-2 text-sm text-gray-800 font-bold flex items-center gap-2">
                            <span>📜</span>
                            Detection History
                        </h2>
                        <div class="text-center py-8 text-gray-500">
                            <p class="text-sm">ประวัติการตรวจจับจากวันที่ที่เลือก</p>
                        </div>
                    {/if}
                </div>
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

	/* Always show scrollbar for camera list */
	.list-scrollbar::-webkit-scrollbar {
        height: 6px;
	}

	.list-scrollbar::-webkit-scrollbar-track {
		background: #f3f4f6;
		border-radius: 2px;
	}

	.list-scrollbar::-webkit-scrollbar-thumb {
		background: #d1d5db;
		border-radius: 2px;
	}

	.list-scrollbar::-webkit-scrollbar-thumb:hover {
		background: #9ca3af;
	}
</style>