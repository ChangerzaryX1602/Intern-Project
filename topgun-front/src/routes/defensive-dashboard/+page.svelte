<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import { env } from '$env/dynamic/public';
	import StatsHeader from './StatsHeader.svelte';
	import CameraSelector from './CameraSelector.svelte';
	import DetectionCard from './DetectionCard.svelte';
	import type { Camera, Detection, Pagination } from './types';

	// Environment variables
	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';
	const wsUrl = env.PUBLIC_WS_URL || 'ws://localhost:8080/api/v1/detect/ws';
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080/api/v1';

	// State
	let cameras = $state<Camera[]>([]);
	let selectedCameraIds = $state<Set<string>>(new Set());
	let searchCameraName = $state('');
	let isLoadingCameras = $state(false);
	let cameraPagination = $state<Pagination>({
		page: 1,
		limit: 20,
		total: 0,
		totalPages: 0
	});

	let detections = $state<Detection[]>([]);
	let selectedDetection = $state<Detection | null>(null);
	let mapCenter: [number, number] = $state([100.5018, 13.7563]);
	let markers = $state<Array<{ lngLat: [number, number]; popup?: string; color?: string }>>([]);

	// WebSocket connections map (camera_id -> WebSocket)
	let wsConnections = $state<Map<string, WebSocket>>(new Map());
	let reconnectTimeouts = new Map<string, any>();
	let searchDebounceTimeout: any = null;

	// Fetch cameras from API
	async function fetchCameras(page: number = 1, search: string = '') {
		isLoadingCameras = true;
		try {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: cameraPagination.limit.toString()
			});

			if (search.trim()) {
				params.append('keyword', search.trim());
				params.append('column', 'name,location,institute');
			}

			const response = await fetch(`${apiUrl}/camera?${params}`);
			const result = await response.json();

			if (result.success && result.data) {
				cameras = result.data.cameras || [];
				if (result.data.pagination) {
					cameraPagination = {
						page: result.data.pagination.page || 1,
						limit: result.data.pagination.limit || 20,
						total: result.data.pagination.total || 0,
						totalPages: result.data.pagination.total_pages || 0
					};
				}
			}
		} catch (error) {
			console.error('Failed to fetch cameras:', error);
			alert('Failed to load cameras');
		} finally {
			isLoadingCameras = false;
		}
	}

	// Toggle camera selection
	function toggleCameraSelection(cameraId: string) {
		const newSet = new Set(selectedCameraIds);
		if (newSet.has(cameraId)) {
			newSet.delete(cameraId);
			disconnectCamera(cameraId);
		} else {
			newSet.add(cameraId);
			connectCamera(cameraId);
		}
		selectedCameraIds = newSet;
	}

	// Connect WebSocket for a specific camera
	function connectCamera(cameraId: string) {
		if (wsConnections.has(cameraId)) {
			return;
		}

		const ws = new WebSocket(wsUrl);

		ws.onopen = () => {
			console.log(`WebSocket connected for camera: ${cameraId}`);
			ws.send(JSON.stringify({ camera_id: cameraId }));
		};

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data);
			console.log(`Received from camera ${cameraId}:`, data);

				if (data.status === 'subscribed') {
					console.log('Subscribed to camera:', data.camera_id);
				} else if (data.id && data.camera_id) {
					const rawObjects = data.objects || [];

					// Normalize detected objects into our DetectedObject[] shape
					const detected_objects = rawObjects.map((o: any) => ({
						class_name: o.class_name ?? o.class ?? undefined,
						confidence: o.confidence ?? undefined,

						obj_id: o.obj_id ?? undefined,
						type: o.type ?? undefined,
						lat: o.lat ?? undefined,
						lng: o.lng ?? undefined,
						objective: o.objective ?? undefined,
						size: o.size ?? undefined,
						details: o.details ?? undefined
					}));

					const newDetection: Detection = {
						id: data.id,
						camera_id: data.camera_id,
						detected_at: data.timestamp,
						path: data.path,
						detected_objects,
						image_base64: data.image_data,
						mime_type: data.mime_type
					};

					console.log('New detection added:', newDetection);

				// Add markers for any objects that have lat/lng
				console.log('Raw objects received:', rawObjects);
				const newMarkers = [...markers];
				rawObjects.forEach((o: any, idx: number) => {
					const latRaw = o.lat ?? o.latitude ?? null;
					const lngRaw = o.lng ?? o.longitude ?? o.lon ?? null;

					console.log(`Object ${idx}: lat=${latRaw}, lng=${lngRaw}`, o);

					if (latRaw != null && lngRaw != null) {
						const lat = typeof latRaw === 'string' ? parseFloat(latRaw) : Number(latRaw);
						const lng = typeof lngRaw === 'string' ? parseFloat(lngRaw) : Number(lngRaw);

						console.log(`Parsed: lat=${lat}, lng=${lng}`);

						if (!Number.isNaN(lat) && !Number.isNaN(lng)) {
							// popup: show camera name, obj id/type and timestamp
							const cameraName = getCameraName(data.camera_id);
							const objId = o.obj_id ?? o.id ?? `obj_${idx}`;
							const objType = o.type ?? o.class_name ?? 'object';
							const objective = o.objective ?? '';
							const size = o.size ?? '';

							const popup = `
								<div style="font-size:13px">
									<strong>${cameraName}</strong><br/>
									${objType} ${objId}<br/>
									${objective ? `objective: ${objective}<br/>` : ''}
									${size ? `size: ${size}<br/>` : ''}
									${new Date(data.timestamp).toLocaleString()}
								</div>`;

						const color = (o.objective && String(o.objective).toLowerCase() === 'our') ? '#10b981' : '#ef4444';

						const marker: { lngLat: [number, number]; popup?: string; color?: string } = { 
							lngLat: [lng, lat] as [number, number], 
							popup, 
							color 
						};
						console.log('Adding marker:', marker);
						newMarkers.push(marker);
						}
					}
				});

				console.log('Total markers after processing:', newMarkers.length);
				markers = newMarkers;					// Prepend detection to list
					detections = [newDetection, ...detections];

					if (!selectedDetection) {
						selectedDetection = newDetection;
					}
				}
		};

		ws.onerror = (error) => {
			console.error(`WebSocket error for camera ${cameraId}:`, error);
		};

		ws.onclose = () => {
			console.log(`WebSocket disconnected for camera ${cameraId}`);
			wsConnections.delete(cameraId);

			const timeout = setTimeout(() => {
				if (selectedCameraIds.has(cameraId)) {
					console.log(`Reconnecting camera ${cameraId}...`);
					connectCamera(cameraId);
				}
			}, 3000);

			reconnectTimeouts.set(cameraId, timeout);
		};

		wsConnections.set(cameraId, ws);
	}

	// Disconnect WebSocket for a specific camera
	function disconnectCamera(cameraId: string) {
		const ws = wsConnections.get(cameraId);
		if (ws) {
			ws.close();
			wsConnections.delete(cameraId);
		}

		const timeout = reconnectTimeouts.get(cameraId);
		if (timeout) {
			clearTimeout(timeout);
			reconnectTimeouts.delete(cameraId);
		}
	}

	// Disconnect all cameras
	function disconnectAllCameras() {
		wsConnections.forEach((ws, cameraId) => {
			disconnectCamera(cameraId);
		});
		selectedCameraIds = new Set();
		detections = [];
		selectedDetection = null;
	}

	// Search cameras
	function handleSearch() {
		fetchCameras(1, searchCameraName);
	}

	// Debounced search - auto search after 0.5 seconds
	function handleSearchInput() {
		if (searchDebounceTimeout) {
			clearTimeout(searchDebounceTimeout);
		}
		searchDebounceTimeout = setTimeout(() => {
			fetchCameras(1, searchCameraName);
		}, 500);
	}

	// Load more cameras (pagination)
	function loadNextPage() {
		if (cameraPagination.page < cameraPagination.totalPages) {
			fetchCameras(cameraPagination.page + 1, searchCameraName);
		}
	}

	function loadPrevPage() {
		if (cameraPagination.page > 1) {
			fetchCameras(cameraPagination.page - 1, searchCameraName);
		}
	}

	function selectDetection(detection: Detection) {
		selectedDetection = detection;
	}

	// Get camera name by ID
	function getCameraName(cameraId: string): string {
		const camera = cameras.find((c) => c.id === cameraId);
		return camera ? camera.name : cameraId.substring(0, 8);
	}

	// Highlight matching text
	function highlightText(text: string, keyword: string): string {
		if (!keyword.trim()) return text;
		const regex = new RegExp(`(${keyword.trim()})`, 'gi');
		return text.replace(regex, '<mark>$1</mark>');
	}

	// Debug: Log marker changes
	$effect(() => {
		console.log('Markers state updated:', markers.length, markers);
	});

	// Load cameras on mount
	onMount(() => {
		fetchCameras();
	});

	onDestroy(() => {
		disconnectAllCameras();
		if (searchDebounceTimeout) {
			clearTimeout(searchDebounceTimeout);
		}
	});
</script>

<svelte:head>
	<title>Defensive Dashboard - Real-time Detection</title>
</svelte:head>

<div class="w-screen h-screen flex flex-col bg-gray-100 overflow-hidden">
	<!-- Header with Stats -->
	<StatsHeader
		selectedCamerasCount={selectedCameraIds.size}
		detectionsCount={detections.length}
		activeConnectionsCount={wsConnections.size}
		onDisconnectAll={disconnectAllCameras}
	/>

	<!-- Main Content Layout: Camera Selector (30%) + Map (70%) -->
	<main class="flex gap-6 px-6 pt-6 pb-0 overflow-hidden flex-1">
		<CameraSelector
			{cameras}
			{selectedCameraIds}
			bind:searchName={searchCameraName}
			isLoading={isLoadingCameras}
			pagination={cameraPagination}
			{wsConnections}
			onToggleCamera={toggleCameraSelection}
			onSearch={handleSearch}
			onSearchChange={(value) => (searchCameraName = value)}
			onSearchInput={handleSearchInput}
			onNextPage={loadNextPage}
			onPrevPage={loadPrevPage}
		/>

		<section class="w-[70%] bg-white rounded-xl shadow-md overflow-hidden relative xl:w-[65%]">
			<div class="w-full h-full bg-gray-300">
				<MapboxMap accessToken={mapboxToken} center={mapCenter} zoom={12} {markers} />
			</div>
		</section>
	</main>

	<!-- Bottom Horizontal Detection List -->
	<footer class="bg-white border-t-2 border-gray-200 px-6 py-4 max-h-[220px] flex flex-col">
		<div class="flex justify-between items-center mb-3">
			<h2 class="m-0 text-lg flex items-center gap-2 text-gray-800 font-bold">
				<span class="inline-block">📋</span>
				All Detections
				<span class="bg-indigo-500 text-white px-2 py-0.5 rounded-xl text-xs font-semibold">
					{detections.length}
				</span>
			</h2>
		</div>

		<div class="flex gap-4 overflow-x-auto overflow-y-hidden py-2 flex-1 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 hover:scrollbar-thumb-gray-400">
			{#if detections.length === 0}
				<div class="flex flex-col items-center justify-center w-full p-8 text-center text-gray-400">
					<div class="text-4xl mb-2 opacity-50">📷</div>
					<p class="my-1 font-medium text-gray-600">No detections yet</p>
					<small class="text-sm">Select cameras to start monitoring...</small>
				</div>
			{:else}
				{#each detections as detection (detection.id)}
					<DetectionCard
						{detection}
						isSelected={selectedDetection?.id === detection.id}
						cameraName={getCameraName(detection.camera_id)}
						onClick={() => selectDetection(detection)}
					/>
				{/each}
			{/if}
		</div>
	</footer>
</div>
