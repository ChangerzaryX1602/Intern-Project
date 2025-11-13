<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import VideoStream from '$lib/components/VideoStream.svelte';
	import { env } from '$env/dynamic/public';
	import StatsHeader from './StatsHeader.svelte';
	import CameraSelector from './CameraSelector.svelte';
	import DetectionCard from './DetectionCard.svelte';
	import SearchResultCard from './SearchResultCard.svelte';
	import type { Camera, Detection, Pagination } from './types';
	import { goto } from '$app/navigation';

	// Environment variables
	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';
	const wsUrl = env.PUBLIC_WS_URL || 'ws://localhost:8080/api/v1/detect/ws';
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080/api/v1';
	const videoServerUrl = env.PUBLIC_VIDEO_SERVER_URL || 'ws://localhost:8080';

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
	let markers = $state<Array<{ lngLat: [number, number]; popup?: string; color?: string; icon?: string; kind?: 'start' | 'latest' }>>([]);
	
	// Track latest position for each drone (by track_id) with timestamp
	type DronePosition = { lngLat: [number, number]; popup?: string; color?: string; icon?: string; kind?: 'start' | 'latest'; timestamp: number };
	let latestDronePositions = $state<Map<number, DronePosition>>(new Map());
	
	// Track full path for each drone
	type DronePathPoint = { lngLat: [number, number]; timestamp: number; color: string };
	let dronePaths = $state<Map<number, DronePathPoint[]>>(new Map());
	
	// Maximum drones to display on map
	const MAX_DRONES_ON_MAP = 2;

	// Generate path lines for drones
	let pathLines = $derived.by(() => {
		const lines: Array<any> = [];

		for (const [trackId, pathPoints] of dronePaths.entries()) {
			if (pathPoints.length < 2) continue;

			// Get the color from the most recent point
			const color = pathPoints[pathPoints.length - 1].color;

			// Build coordinates array
			const coordinates = pathPoints.map(p => p.lngLat);

			lines.push({
				id: `drone-${trackId}`,
				coordinates,
				color
			});
		}

		return lines;
	});

	// Search history state
	let startDate = $state<string>('');
	let endDate = $state<string>('');
	let filteredDetections = $state<Detection[]>([]);
	let searchHistory = $state<Array<{ startDate: string; endDate: string; count: number }>>([]);
	let isSearching = $state(false);
	let searchError = $state('');
	let showSearchModal = $state(false);

	// WebSocket connections map (camera_id -> WebSocket)
	let wsConnections = $state<Map<string, WebSocket>>(new Map());
	let reconnectTimeouts = new Map<string, any>();
	let searchDebounceTimeout: any = null;

	// Model upload state
	let isUploadingModel = $state(false);
	let uploadProgress = $state('');
	let fileInputRef: HTMLInputElement;

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
						lng: o.lng ?? o.lon ?? undefined,
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
						objects: rawObjects, // Add original objects for DetectionCard
						image_base64: data.image_data,
						mime_type: data.mime_type
					};

					console.log('New detection added:', newDetection);

				// Add markers for any objects that have lat/lng
				console.log('Raw objects received:', rawObjects);
				rawObjects.forEach((o: any, idx: number) => {
					const latRaw = o.lat ?? o.latitude ?? null;
					const lngRaw = o.lng ?? o.longitude ?? o.lon ?? null;

					console.log(`Object ${idx}: lat=${latRaw}, lng=${lngRaw}`, o);

					if (latRaw != null && lngRaw != null) {
						const lat = typeof latRaw === 'string' ? parseFloat(latRaw) : Number(latRaw);
						const lng = typeof lngRaw === 'string' ? parseFloat(lngRaw) : Number(lngRaw);

						console.log(`Parsed: lat=${lat}, lng=${lng}`);

						if (!Number.isNaN(lat) && !Number.isNaN(lng)) {
							const cameraName = getCameraName(data.camera_id);
							const trackId = o.track_id ?? idx;
							const objType = o.type ?? o.class_name ?? 'drone';
							const objective = o.objective ?? '';
							const size = o.size ?? '';

							const popup = `
								<div style="font-size:13px">
									<strong>${cameraName}</strong><br/>
									Track ID: ${trackId}<br/>
									${objective ? `Objective: ${objective}<br/>` : ''}
									${size ? `Size: ${size}<br/>` : ''}
									${new Date(data.timestamp).toLocaleString()}
								</div>`;

						const color = (o.objective && String(o.objective).toLowerCase() === 'our') ? '#10b981' : '#ef4444';

						// Update latest position for this track_id with timestamp
						latestDronePositions.set(trackId, {
							lngLat: [lng, lat] as [number, number], 
							popup, 
							color,
							icon: 'drone',
							kind: 'latest',
							timestamp: Date.now()
						});

						// Add to path history
						if (!dronePaths.has(trackId)) {
							dronePaths.set(trackId, []);
						}
						const path = dronePaths.get(trackId)!;
						path.push({
							lngLat: [lng, lat] as [number, number],
							timestamp: Date.now(),
							color
						});
						}
					}
				});

				// Limit to MAX_DRONES_ON_MAP (keep the most recent ones)
				if (latestDronePositions.size > MAX_DRONES_ON_MAP) {
					// Sort by timestamp (most recent first)
					const sortedDrones = Array.from(latestDronePositions.entries())
						.sort((a, b) => b[1].timestamp - a[1].timestamp);
					
					// Get IDs to keep
					const idsToKeep = new Set(sortedDrones.slice(0, MAX_DRONES_ON_MAP).map(([id]) => id));
					
					// Remove old drones from positions and paths
					latestDronePositions.clear();
					sortedDrones.slice(0, MAX_DRONES_ON_MAP).forEach(([trackId, drone]) => {
						latestDronePositions.set(trackId, drone);
					});
					
					// Clean up paths for drones no longer tracked
					for (const trackId of dronePaths.keys()) {
						if (!idsToKeep.has(trackId)) {
							dronePaths.delete(trackId);
						}
					}
					
					console.log(`Limited drones to ${MAX_DRONES_ON_MAP} most recent`);
				}

				// Convert latest positions to markers array (without timestamp)
				markers = Array.from(latestDronePositions.values()).map(({ timestamp, ...drone }) => drone);
				console.log('Total unique drones on map:', markers.length);					// Prepend detection to list
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

	// Search detections by date range from API
	async function searchDetectionsByDate() {
		if (!startDate || !endDate) {
			searchError = 'Please select both start and end dates';
			return;
		}

		// Validate date range
		const start = new Date(startDate);
		const end = new Date(endDate);
		if (start > end) {
			searchError = 'Start date must be before end date';
			return;
		}

		isSearching = true;
		searchError = '';

		try {
			// Build query params
			const params = new URLSearchParams({
				start_date: startDate,
				end_date: endDate,
				page: '1',
				limit: '100' // Get more results for history search
			});

			const response = await fetch(`${apiUrl}/detect?${params}`);
			const result = await response.json();

			console.log('API Response:', result);

			if (!response.ok || !result.success) {
				throw new Error(result.error || 'Search failed');
			}

			// Handle different response structures
			let detectionData = [];
			if (result.data) {
				// Check if data is array or object with detects property
				if (Array.isArray(result.data)) {
					detectionData = result.data;
				} else if (result.data.detects && Array.isArray(result.data.detects)) {
					detectionData = result.data.detects;
				} else if (result.data.data && Array.isArray(result.data.data)) {
					detectionData = result.data.data;
				}
			}

			filteredDetections = detectionData;

			console.log('Filtered detections:', filteredDetections);
			console.log('Number of detections:', filteredDetections.length);

			// Add to search history
			const historyEntry = {
				startDate,
				endDate,
				count: filteredDetections.length
			};

			searchHistory = [historyEntry, ...searchHistory.slice(0, 9)]; // Keep last 10 searches

			console.log(`Found ${filteredDetections.length} detections between ${startDate} and ${endDate}`);
			
			// Open modal to show results
			showSearchModal = true;
		} catch (error) {
			console.error('Failed to search detections:', error);
			searchError = error instanceof Error ? error.message : 'Failed to search detections';
			filteredDetections = [];
		} finally {
			isSearching = false;
		}
	}

	// Upload model file via MQTT
	async function uploadModel(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];

		if (!file) {
			return;
		}

		// Validate file extension
		const validExtensions = ['.pt', '.onnx', '.pb', '.tflite', '.pth'];
		const fileExt = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
		if (!validExtensions.includes(fileExt)) {
			alert(`Invalid file type. Supported: ${validExtensions.join(', ')}`);
			input.value = ''; // Reset input
			return;
		}

		isUploadingModel = true;
		uploadProgress = `Uploading ${file.name}...`;

		try {
			// Create FormData
			const formData = new FormData();
			formData.append('file', file);
			formData.append('encode_base64', 'true');

			// Upload to API
			const response = await fetch(`${apiUrl}/mqtt/upload-file`, {
				method: 'POST',
				body: formData
			});

			const result = await response.json();

			if (!response.ok || !result.success) {
				throw new Error(result.error || 'Upload failed');
			}

			uploadProgress = `Successfully uploaded ${file.name}`;
			alert(`Model uploaded successfully!\n\nFile: ${result.filename}\nSize: ${(result.size / 1024).toFixed(2)} KB\nTopic: ${result.topic}`);

			// Reset input
			input.value = '';

			// Clear success message after 3 seconds
			setTimeout(() => {
				uploadProgress = '';
			}, 3000);
		} catch (error) {
			console.error('Failed to upload model:', error);
			uploadProgress = '';
			alert(`Failed to upload model: ${error instanceof Error ? error.message : 'Unknown error'}`);
			input.value = ''; // Reset input
		} finally {
			isUploadingModel = false;
		}
	}

	// Format date for display
	function formatDateDisplay(dateString: string): string {
		const date = new Date(dateString);
		const day = String(date.getDate()).padStart(2, '0');
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const year = date.getFullYear();
		return `${day}/${month}/${year}`;
	}

	// Clear search
	function clearSearch() {
		startDate = '';
		endDate = '';
		filteredDetections = [];
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

	<!-- Main Content Layout: Camera Selector -->
	<main class="flex gap-6 px-6 pt-6 pb-0 overflow-hidden flex-1">
		<section class="flex flex-col w-[80%]">
			<div class="flex gap-6 h-[60%]">
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
				
				<div class="h-full bg-gray-300 w-[80%] rounded-xl shadow-md">
					<MapboxMap accessToken={mapboxToken} center={mapCenter} zoom={12} {markers} />
				</div>
			</div>

			<div class="flex justify-between h-[35%] w-full mt-6 gap-6">
				<div class="bg-white rounded-xl shadow-md w-[37%]">
					เอาไว้ทำไรเอาไว้ทำไรเอาไว้ทำไรเอ
				</div>

				<div class="h-full bg-black w-full rounded-xl shadow-md">
					<VideoStream serverUrl={videoServerUrl} showStats={true} autoReconnect={true} />
				</div>
			</div>
		</section>
		
		<section class="bg-white rounded-xl shadow-md flex">
			<div class="bg-white flex flex-col justify-between items-center px-4 py-4 gap-4 border-b border-gray-200">
				<!-- Model Upload Section -->
				<div class="w-full">
					<div class="flex items-center gap-2 mb-4">
						<span class="text-2xl">🔍</span>
						<h1 class="text-base font-bold text-gray-800">Search History</h1>
					</div>
					<div class="flex flex-col gap-3">
						<div class="flex gap-3">
							<div class="flex-1">
								<label for="start-date" class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
								<input 
									id="start-date" 
									type="date" 
									bind:value={startDate}
									class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm" 
								/>
							</div>
							<div class="flex-1">
								<label for="end-date" class="block text-sm font-medium text-gray-700 mb-1">End Date</label>
								<input 
									id="end-date" 
									type="date" 
									bind:value={endDate}
									class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm" 
								/>
							</div>
						</div>
						<button 
							onclick={searchDetectionsByDate}
							disabled={isSearching || !startDate || !endDate}
							class="w-full px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors font-medium text-sm"
						>
							{#if isSearching}
								กำลังค้นหา...
							{:else}
								ค้นหา
							{/if}
						</button>
						{#if searchError}
							<p class="text-xs text-red-600 text-center">{searchError}</p>
						{/if}
						{#if filteredDetections.length > 0}
							<p class="text-sm text-green-600 text-center font-medium">
								พบ {filteredDetections.length} รายการ
							</p>
						{/if}
					</div>

					<div class="flex justify-between items-center mb-3 border-t border-gray-200 mt-3">
						<h2 class="mt-2 text-lg flex items-center gap-2 text-gray-800 font-bold">
							<span class="inline-block">📋</span>
							All Detections
							<span class="bg-indigo-500 text-white px-2 py-0.5 rounded-xl text-xs font-semibold">
								{detections.length}
							</span>
						</h2>
					</div>

					<div class="flex flex-col gap-4 overflow-y-auto py-2 flex-1 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 hover:scrollbar-thumb-gray-400">
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
				</div>
				
				<!-- Upload New Model -->
				<div class="w-full border-t border-gray-200 pt-2">
					<div class="bg-gradient-to-r from-indigo-50 to-blue-50 rounded-lg p-2 mb-3">
						<div class="flex justify-center items-center gap-2">
							<span class="text-2xl">🤖</span>
							<div>
								<h3 class="font-bold text-gray-800 text-sm">Current Model</h3>
								<p class="text-xs text-gray-600">YOLO v8 NCNN Model 960</p>
							</div>
						</div>
					</div>
					<label for="file-upload" class="block text-sm font-semibold text-gray-700 mb-2">Upload New Model</label>
					<div class="relative">
						<input 
							id="file-upload" 
							type="file" 
							accept=".pt,.onnx,.pb,.tflite,.pth"
							class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
							onchange={uploadModel}
							bind:this={fileInputRef}
							disabled={isUploadingModel}
						/>
						<div class="flex items-center justify-center gap-2 px-4 py-3 bg-indigo-50 border-2 border-dashed border-indigo-300 rounded-lg hover:bg-indigo-100 hover:border-indigo-400 transition-all cursor-pointer {isUploadingModel ? 'opacity-50 cursor-not-allowed' : ''}">
							{#if isUploadingModel}
								<span class="text-lg animate-spin">⌛</span>
								<span class="text-sm font-medium text-indigo-700">Uploading...</span>
							{:else}
								<span class="text-lg">⏫</span>
								<span class="text-sm font-medium text-indigo-700">Select model file</span>
							{/if}
						</div>
						<p class="text-xs text-gray-500 mt-1.5 text-center">Supported: .pt, .onnx, .pb, .tflite</p>
						{#if uploadProgress}
							<p class="text-xs text-green-600 mt-1.5 text-center font-medium">{uploadProgress}</p>
						{/if}
					</div>
				</div>
			</div>
		</section>
	</main>
</div>

<!-- Search Results Modal -->
{#if showSearchModal}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50" onclick={() => showSearchModal = false} onkeydown={(e) => e.key === 'Escape' && (showSearchModal = false)} role="button" tabindex="0">
		<div class="bg-white rounded-xl shadow-2xl w-[90vw] h-[85vh] flex flex-col" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
			<!-- Modal Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
				<div>
					<h2 class="text-xl font-bold text-gray-800">ผลการค้นหา</h2>
					<p class="text-sm text-gray-600 mt-1">
						{formatDateDisplay(startDate)} - {formatDateDisplay(endDate)}
						<span class="ml-2 bg-indigo-100 text-indigo-700 px-2 py-0.5 rounded-full text-xs font-semibold">
							พบ {filteredDetections.length} รายการ
						</span>
					</p>
				</div>
				<button 
					onclick={() => showSearchModal = false}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Close modal"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Modal Content -->
			<div class="flex-1 overflow-y-auto p-6">
				{#if filteredDetections.length === 0}
					<div class="flex flex-col items-center justify-center h-full text-gray-400">
						<div class="text-6xl mb-4 opacity-50">🔍</div>
						<p class="text-lg font-medium text-gray-600">ไม่พบรายการในช่วงเวลาที่เลือก</p>
						<p class="text-sm text-gray-500 mt-2">ลองเปลี่ยนช่วงเวลาในการค้นหา</p>
					</div>
				{:else}
					<div class="grid grid-cols-3 gap-4">
						{#each filteredDetections as detection (detection.id)}
							<SearchResultCard
								{detection}
								cameraName={detection.camera?.name || 'GearDinDaeng2025'}
								onClick={() => {
									selectedDetection = detection;
									showSearchModal = false;
								}}
							/>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Modal Footer -->
			<div class="flex items-center justify-between px-6 py-4 border-t border-gray-200 bg-gray-50">
				<div class="text-sm text-gray-600">
					แสดง {filteredDetections.length} จากทั้งหมด {filteredDetections.length} รายการ
				</div>
				<button 
					onclick={() => showSearchModal = false}
					class="px-4 py-2 bg-gray-600 text-white rounded-md hover:bg-gray-700 transition-colors font-medium"
				>
					ปิด
				</button>
			</div>
		</div>
	</div>
{/if}
