<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import VideoStream from '$lib/components/VideoStream.svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import { env } from '$env/dynamic/public';
	
	// Import defensive dashboard components
	import CameraSelector from '../defensive-dashboard/CameraSelector.svelte';
	import DetectionCard from '../defensive-dashboard/DetectionCard.svelte';
	import SearchResultCard from '../defensive-dashboard/SearchResultCard.svelte';
	import type { Camera, Detection, Pagination } from '../defensive-dashboard/types';
	import { goto } from '$app/navigation';
	import { LngLat } from 'mapbox-gl';

	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';
	const wsUrl = env.PUBLIC_WS_URL || 'ws://localhost:8080/api/v1/detect/ws';
	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080/api/v1';
	const videoServerUrl = env.PUBLIC_VIDEO_SERVER_URL || 'ws://localhost:8080';
	const ATTACK_WS_URL = 'ws://localhost:8080/api/v1/detect/attack-ws';
	const ATTACK_API_URL = 'http://localhost:8080/api/v1/attack';

	// Offensive drone data (from Attack API) - EXACT SAME AS OFFENSIVE DASHBOARD
	interface Vector3 {
		x: number;
		y: number;
		z: number;
	}

	interface TargetLocation {
		lat: number;
		lng: number;
		description: string;
	}

	interface AttackData {
		id: number;
		drone_id: string;
		lat: number;
		lng: number;
		height: number;
		function: string;
		acceleration: Vector3;
		velocity: Vector3;
		distance: number;
		status: string;
		time_left: number;
		target?: TargetLocation;
		landing?: TargetLocation;
		created_at: string;
	}

	// Transformed drone data for UI (EACH PATH POINT)
	interface Drone {
		id: string; // Unique ID (from attack.id)
		trackingId: string; // Tracking ID for grouping (drone_id)
		name: string;
		status: 'connected' | 'disconnected';
		location: string;
		coordinates: {
			lat: number;
			lng: number;
		};
		gpsStatus: 'good' | 'loss';
		lastUpdate: string;
		height: number;
		velocity: Vector3;
		acceleration: Vector3;
		distance: number;
		timeLeft: number;
		target?: TargetLocation;
		landing?: TargetLocation;
	}

	// Grouped drone data (SAME AS OFFENSIVE DASHBOARD)
	interface DroneGroup {
		droneId: string;
		name: string;
		paths: Drone[];
		isExpanded: boolean;
		lastStatus: 'connected' | 'disconnected';
		lastGpsStatus: 'good' | 'loss';
		latestUpdate: string;
	}

	// Defensive camera data (from Camera API)
	interface CameraData {
		id: string;
		name: string;
		location?: string;
		institute?: string;
		latitude?: number;
		longitude?: number;
	}

	interface Camera {
		id: string;
		name: string;
		status: 'online' | 'offline';
		location: string;
		coordinates: { lat: number; lng: number };
	}

	// Detection data (from Detection WebSocket)
	interface DetectionData {
		id: number;
		camera_id: string;
		timestamp: string;
		path: string;
		objects?: any[];
	}

	interface Detection {
		id: string;
		cameraId: string;
		cameraName: string;
		droneId: string;
		detectedAt: string;
		coordinates: { lat: number; lng: number };
		imageUrl?: string;
		objectCount?: number;
		objects?: any[];
		image_base64?: string;
	}

	// Color palette for random drone colors (SAME AS OFFENSIVE DASHBOARD)
	const DRONE_COLORS = [
		'#ef4444', // red
		'#f59e0b', // orange
		'#eab308', // yellow
		'#84cc16', // lime
		'#10b981', // green
		'#14b8a6', // teal
		'#06b6d4', // cyan
		'#3b82f6', // blue
		'#6366f1', // indigo
		'#8b5cf6', // violet
		'#a855f7', // purple
		'#ec4899', // pink
		'#f43f5e'  // rose
	];

	// Store assigned colors for each drone_id
	const droneColors = new Map<string, string>();

	// Get or assign color for drone
	function getDroneColor(droneId: string): string {
		if (!droneColors.has(droneId)) {
			// Assign random color
			const color = DRONE_COLORS[Math.floor(Math.random() * DRONE_COLORS.length)];
			droneColors.set(droneId, color);
		}
		return droneColors.get(droneId)!;
	}

	let droneGroups = $state<DroneGroup[]>([]);
	let allDrones = $state<Drone[]>([]); // Keep all drones for map
	let cameras = $state<Camera[]>([]);
	let detections = $state<Detection[]>([]);

	// WebSocket connections
	let attackWs: WebSocket | null = null;
	let detectionWsConnections = new Map<string, WebSocket>();
	let selectedCameraIds = $state<Set<string>>(new Set());
	let isLoadingCameras = $state(false);
	let selectedCameraId = $state<string | null>(null);

	// Track latest drone positions on defensive map
	type DronePosition = { lngLat: [number, number]; popup?: string; color?: string; icon?: string; kind?: 'start' | 'latest'; timestamp: number };
	let latestDronePositions = $state<Map<number, DronePosition>>(new Map());
	let defensiveMarkers = $state<Array<{ lngLat: [number, number]; popup?: string; color?: string; icon?: string; kind?: 'start' | 'latest' }>>([]);
	const MAX_DRONES_ON_MAP = 2;

	// Track full path for each defensive drone
	type DefensiveDronePathPoint = { lngLat: [number, number]; timestamp: number; color: string };
	let defensiveDronePaths = $state<Map<number, DefensiveDronePathPoint[]>>(new Map());

	// Generate path lines for defensive drones
	let defensivePathLines = $derived.by(() => {
		const lines: Array<any> = [];

		for (const [trackId, pathPoints] of defensiveDronePaths.entries()) {
			if (pathPoints.length < 2) continue;

			// Get the color from the most recent point
			const color = pathPoints[pathPoints.length - 1].color;

			// Build coordinates array
			const coordinates = pathPoints.map(p => p.lngLat);

			lines.push({
				id: `defensive-drone-${trackId}`,
				coordinates,
				color
			});
		}

		return lines;
	});

	let droneSearchQuery = $state('');
	let cameraSearchQuery = $state('');
	let selectedDrone = $state<Drone | null>(null);
	let selectedCamera = $state<Camera | null>(null);
	let showHistory = $state(false);

	// Search history state
	let startDate = $state<string>('');
	let endDate = $state<string>('');
	let filteredDetections = $state<Detection[]>([]);
	let isSearching = $state(false);
	let searchError = $state('');
	let showSearchModal = $state(false);
	
	// Model upload state
	let isUploadingModel = $state(false);
	let uploadProgress = $state('');
	let fileInputRef: HTMLInputElement;

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

	// Generate markers for map: only start point (small colored marker) + latest point (drone sticker)
	// EXACT SAME AS OFFENSIVE DASHBOARD
	let markers = $derived.by(() => {
		const out: Array<any> = [];

		// For each group, if expanded, create two markers: start and latest
		droneGroups.forEach((group) => {
			if (!group.isExpanded) return;
			if (!group.paths || group.paths.length === 0) return;

			const color = getDroneColor(group.droneId);

			// Start marker (first recorded path)
			const start = group.paths[0];
			out.push({
				id: `${group.droneId}-start`,
				kind: 'start',
				lngLat: [start.coordinates.lng, start.coordinates.lat] as [number, number],
				popup: `
					<div style="font-size:13px">
						<strong>${group.name} (start)</strong><br/>
						${start.location}<br/>
						${start.lastUpdate}
					</div>`,
				color
			});

			// Target marker (if exists in first path)
			if (start.target) {
				out.push({
					id: `${group.droneId}-target`,
					kind: 'target',
					lngLat: [start.target.lng, start.target.lat] as [number, number],
					popup: `
						<div style="font-size:13px">
							<strong>🎯 Target</strong><br/>
							${start.target.description}<br/>
							Drone: ${group.name}
						</div>`,
					color: '#ef4444',
					icon: 'target',
					label: '🎯 ' + start.target.description
				});
			}

			// Landing marker (if exists in first path)
			if (start.landing) {
				out.push({
					id: `${group.droneId}-landing`,
					kind: 'landing',
					lngLat: [start.landing.lng, start.landing.lat] as [number, number],
					popup: `
						<div style="font-size:13px">
							<strong>🛬 Landing Point</strong><br/>
							${start.landing.description}<br/>
							Drone: ${group.name}
						</div>`,
					color: '#10b981',
					icon: 'landing',
					label: '🛬 ' + start.landing.description
				});
			}

			// Latest marker (latest path) - rendered as drone sticker
			const latest = group.paths[group.paths.length - 1];
			const velocityMag = Math.sqrt(latest.velocity.x**2 + latest.velocity.y**2 + latest.velocity.z**2);
			const accelerationMag = Math.sqrt(latest.acceleration.x**2 + latest.acceleration.y**2 + latest.acceleration.z**2);
			out.push({
				id: `${group.droneId}-latest`,
				kind: 'latest',
				lngLat: [latest.coordinates.lng, latest.coordinates.lat] as [number, number],
				popup: `
					<div style="font-size:13px">
						<strong>${group.name} (latest)</strong><br/>
						${latest.location}<br/>
						Height: ${latest.height.toFixed(2)}m<br/>
						Velocity: ${velocityMag.toFixed(2)} m/s<br/>
						Acceleration: ${accelerationMag.toFixed(2)} m/s²<br/>
						Distance: ${latest.distance.toFixed(2)} m<br/>
						Time Left: ${latest.timeLeft.toFixed(0)}s<br/>
						GPS: ${latest.gpsStatus === 'good' ? '✓ Connected' : '✗ Loss'}<br/>
						${latest.target ? `🎯 ${latest.target.description}<br/>` : ''}
						${latest.lastUpdate}
					</div>`,
				color,
				icon: 'drone'
			});
		});

		return out;
	});

	// Generate path lines with ALL coordinates (not just start and latest)
	// EXACT SAME AS OFFENSIVE DASHBOARD
	// Split into segments based on GPS status changes
	let pathLines = $derived.by(() => {
		const lines: Array<any> = [];

		droneGroups.forEach((group) => {
			if (!group.isExpanded) return;
			if (!group.paths || group.paths.length < 2) return; // Need at least 2 points for a line

			const originalColor = getDroneColor(group.droneId);
			
			// Split path into segments based on GPS status
			let currentSegment: [number, number][] = [];
			let currentStatus = group.paths[0].gpsStatus;
			let segmentIndex = 0;

			group.paths.forEach((drone, index) => {
				const coord: [number, number] = [drone.coordinates.lng, drone.coordinates.lat];
				
				// If status changed, save current segment and start new one
				if (drone.gpsStatus !== currentStatus && currentSegment.length > 0) {
					// Add current point to close the segment
					currentSegment.push(coord);
					
					// Save segment with appropriate color
					lines.push({
						id: `${group.droneId}-segment-${segmentIndex++}`,
						coordinates: [...currentSegment],
						color: currentStatus === 'loss' ? '#ef4444' : originalColor
					});
					
					// Start new segment with this point
					currentSegment = [coord];
					currentStatus = drone.gpsStatus;
				} else {
					currentSegment.push(coord);
				}
			});

			// Add final segment
			if (currentSegment.length > 1) {
				lines.push({
					id: `${group.droneId}-segment-${segmentIndex}`,
					coordinates: currentSegment,
					color: currentStatus === 'loss' ? '#ef4444' : originalColor
				});
			}
		});

		return lines;
	});

	// Generate camera markers for defensive map - SAME AS DEFENSIVE DASHBOARD
	let cameraMarkers = $derived(
		cameras.map((c) => ({
			id: c.id,
			lngLat: [c.coordinates.lng, c.coordinates.lat] as [number, number],
			popup: `<div style="font-size:13px"><strong>${c.name}</strong><br/>${c.location}</div>`,
			color: c.status === 'online' ? '#3b82f6' : '#ef4444'
		}))
	);

	// Filter drone groups based on search
	let filteredGroups = $derived(
		droneGroups.filter((g) => g.name.toLowerCase().includes(droneSearchQuery.toLowerCase()))
	);

	let filteredCameras = $derived(
		cameras.filter((c) =>
			c.name.toLowerCase().includes(cameraSearchQuery.toLowerCase()) ||
			c.location.toLowerCase().includes(cameraSearchQuery.toLowerCase())
		)
	);

	// Toggle group expansion
	function toggleGroupExpansion(droneId: string) {
		const group = droneGroups.find((g) => g.droneId === droneId);
		if (group) {
			group.isExpanded = !group.isExpanded;
		}
	}

	function selectDrone(drone: Drone) {
		selectedDrone = drone;
		// Center map on selected drone
		if (drone.status === 'connected') {
			droneMapCenter = [drone.coordinates.lng, drone.coordinates.lat];
		}
	}

	function selectCamera(camera: Camera) {
		selectedCamera = camera;
		selectedCameraId = camera.id;
		cameraMapCenter = [camera.coordinates.lng, camera.coordinates.lat];
	}

	function getCameraName(cameraId: string): string {
		const camera = cameras.find(c => c.id === cameraId);
		return camera?.name || 'Unknown Camera';
	}

	function disconnectCamera(cameraId: string) {
		const ws = detectionWsConnections.get(cameraId);
		if (ws) {
			ws.close();
			detectionWsConnections.delete(cameraId);
		}
	}

	// Upload model file via MQTT - SAME AS DEFENSIVE DASHBOARD
	async function uploadModel(event: Event) {
		console.log('uploadModel called', event);
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		console.log('Selected file:', file);

		if (!file) return;

		// Validate file extension
		const validExtensions = ['.pt', '.onnx', '.pb', '.tflite', '.pth'];
		const fileExt = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
		if (!validExtensions.includes(fileExt)) {
			alert(`Invalid file type. Supported: ${validExtensions.join(', ')}`);
			input.value = '';
			return;
		}

		isUploadingModel = true;
		uploadProgress = `Uploading ${file.name}...`;

		try {
			const formData = new FormData();
			formData.append('file', file);
			formData.append('encode_base64', 'true');

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

			input.value = '';

			setTimeout(() => {
				uploadProgress = '';
			}, 3000);
		} catch (error) {
			console.error('Failed to upload model:', error);
			uploadProgress = '';
			alert(`Failed to upload model: ${error instanceof Error ? error.message : 'Unknown error'}`);
			input.value = '';
		} finally {
			isUploadingModel = false;
		}
	}

	// Search detection history - SAME AS DEFENSIVE DASHBOARD
	async function searchDetectionHistory() {
		if (!startDate || !endDate) {
			searchError = 'กรุณาเลือกวันที่เริ่มต้นและสิ้นสุด';
			return;
		}

		isSearching = true;
		searchError = '';
		try {
			const params = new URLSearchParams({
				start_date: startDate,
				end_date: endDate,
				page: '1',
				limit: '50'
			});

			// Add selected camera IDs if any
			if (selectedCameraIds.size > 0) {
				Array.from(selectedCameraIds).forEach(id => {
					params.append('camera_ids', id);
				});
			}

			const response = await fetch(`${apiUrl}/detect?${params}`);
			const result = await response.json();

			if (result.success && result.data) {
				// Handle different response structures
				let detectionData = [];
				if (Array.isArray(result.data)) {
					detectionData = result.data;
				} else if (result.data.detects && Array.isArray(result.data.detects)) {
					detectionData = result.data.detects;
				} else if (result.data.data && Array.isArray(result.data.data)) {
					detectionData = result.data.data;
				}

				filteredDetections = detectionData;
				showSearchModal = true;
			} else {
				searchError = 'ไม่พบข้อมูล';
			}
		} catch (error) {
			console.error('Failed to search detection history:', error);
			searchError = 'เกิดข้อผิดพลาดในการค้นหา';
		} finally {
			isSearching = false;
		}
	}

	// Format date for display - SAME AS DEFENSIVE DASHBOARD
	function formatDateDisplay(dateString: string): string {
		const date = new Date(dateString);
		const day = String(date.getDate()).padStart(2, '0');
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const year = date.getFullYear();
		return `${day}/${month}/${year}`;
	}

	// Transform attack data to drone (SAME AS OFFENSIVE DASHBOARD)
	function transformAttackToDrone(attack: AttackData): Drone {
		return {
			id: `${attack.drone_id}-${attack.id}`,
			trackingId: attack.drone_id,
			name: `Drone ID: ${attack.drone_id}`,
			status: attack.status.toLowerCase() === 'good' || attack.status.toLowerCase() === 'done' ? 'connected' : 'disconnected',
			location: 'ภาคกลาง, กรุงเทพ',
			coordinates: {
				lat: attack.lat,
				lng: attack.lng
			},
			gpsStatus: attack.status.toLowerCase() === 'good' || attack.status.toLowerCase() === 'done' ? 'good' : 'loss',
			lastUpdate: new Date(attack.created_at).toLocaleString('th-TH', {
				year: 'numeric',
				month: '2-digit',
				day: '2-digit',
				hour: '2-digit',
				minute: '2-digit',
				hour12: false
			}),
			height: attack.height,
			velocity: attack.velocity,
			acceleration: attack.acceleration,
			distance: attack.distance,
			timeLeft: attack.time_left || 0,
			target: attack.target,
			landing: attack.landing
		};
	}

	// Update drone groups (group by drone_id) - SAME AS OFFENSIVE DASHBOARD
	function updateDroneGroups() {
		const grouped = new Map<string, Drone[]>();

		// Group drones by trackingId
		allDrones.forEach((drone) => {
			if (!grouped.has(drone.trackingId)) {
				grouped.set(drone.trackingId, []);
			}
			grouped.get(drone.trackingId)!.push(drone);
		});

		// Convert to DroneGroup array
		droneGroups = Array.from(grouped.entries()).map(([droneId, paths]) => {
			// Sort by time (oldest to newest)
			paths.sort((a, b) => {
				const timeA = new Date(a.lastUpdate).getTime();
				const timeB = new Date(b.lastUpdate).getTime();
				return timeA - timeB;
			});

			const latestDrone = paths[paths.length - 1];
			const existingGroup = droneGroups.find((g) => g.droneId === droneId);

			return {
				droneId,
				name: `Drone ID: ${droneId}`,
				paths,
				isExpanded: existingGroup?.isExpanded ?? true, // Default: expanded (show all drones)
				lastStatus: latestDrone.status,
				lastGpsStatus: latestDrone.gpsStatus,
				latestUpdate: latestDrone.lastUpdate
			};
		});
	}

	// Fetch offensive drones from Attack API - SAME AS OFFENSIVE DASHBOARD
	async function fetchAttackData() {
		try {
			const response = await fetch(ATTACK_API_URL);
			const result = await response.json();

			if (result.success && result.data.attacks) {
				// Transform and add initial data
				const drones = result.data.attacks.map((attack: AttackData) => transformAttackToDrone(attack));
				allDrones = drones;
				updateDroneGroups();

				// Set map center to first drone
				if (allDrones.length > 0) {
					droneMapCenter = [allDrones[0].coordinates.lng, allDrones[0].coordinates.lat];
				}
			}
		} catch (error) {
			console.error('Failed to fetch attack data:', error);
		}
	}

	// Connect to Attack WebSocket - SAME AS OFFENSIVE DASHBOARD
	function connectAttackWebSocket() {
		if (attackWs) return;

		attackWs = new WebSocket(ATTACK_WS_URL);

		attackWs.onopen = () => {
			console.log('✅ Attack WebSocket connected');
		};

		attackWs.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				console.log('📨 Received attack data:', data);

				// Handle confirmation message
				if (data.status === 'connected') {
					console.log('✅ Subscribed to attack data stream');
					return;
				}

				// Handle attack data format (single attack from broadcast)
				if (data.drone_id) {
					// Single attack data format - append
					const newDrone = transformAttackToDrone(data);
					const exists = allDrones.some((d) => d.id === newDrone.id);
					if (!exists) {
						allDrones = [...allDrones, newDrone];
						updateDroneGroups();

						// Update map center to latest drone
						const latest = allDrones[allDrones.length - 1];
						droneMapCenter = [latest.coordinates.lng, latest.coordinates.lat];
					}
				}
			} catch (error) {
				console.error('Error parsing attack WebSocket message:', error);
			}
		};

		attackWs.onclose = () => {
			console.log('🔌 Attack WebSocket disconnected');
			attackWs = null;
			setTimeout(connectAttackWebSocket, 3000);
		};
	}

	// Fetch defensive cameras from Camera API
	async function fetchCameras() {
		isLoadingCameras = true;
		try {
			const response = await fetch(`${apiUrl}/camera?page=1&limit=50`);
			const result = await response.json();

			if (result.success && result.data) {
				cameras = (result.data.cameras || []).map((cam: CameraData) => ({
					id: cam.id,
					name: cam.name,
					status: 'online' as const,
					location: cam.location || cam.institute || 'ไม่ระบุตำแหน่ง',
					coordinates: {
						lat: cam.latitude || 13.7563,
						lng: cam.longitude || 100.5018
					}
				}));

				// Auto-select first 3 cameras for detection feed
				if (cameras.length > 0) {
					cameras.slice(0, 3).forEach(cam => {
						selectedCameraIds.add(cam.id);
						connectDetectionWebSocket(cam.id);
					});
				}
			}
		} catch (error) {
			console.error('Failed to fetch cameras:', error);
		} finally {
			isLoadingCameras = false;
		}
	}

	// Connect to Detection WebSocket for a camera
	function connectDetectionWebSocket(cameraId: string) {
		if (detectionWsConnections.has(cameraId)) return;

		const ws = new WebSocket(wsUrl);

		ws.onopen = () => {
			console.log(`Detection WebSocket connected for camera: ${cameraId}`);
			ws.send(JSON.stringify({ camera_id: cameraId }));
		};

		ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (data.status === 'subscribed') return;

				if (data.id && data.camera_id) {
					const camera = cameras.find(c => c.id === data.camera_id);
					const objects = data.objects || [];
					
					const newDetection: Detection = {
						id: String(data.id),
						cameraId: data.camera_id,
						cameraName: camera?.name || 'Unknown',
						droneId: objects[0]?.track_id || 'Unknown',
						detectedAt: new Date(data.timestamp).toLocaleString('th-TH'),
						coordinates: {
							lat: objects[0]?.lat || 13.7563,
							lng: objects[0]?.lon || 100.5018
						},
						objectCount: objects.length,
						objects: objects, // Add raw objects for display
						image_base64: data.image_data // Add image data
					};

					detections = [newDetection, ...detections].slice(0, 10);

					// Update drone markers on defensive map
					objects.forEach((o: any) => {
						const lat = o.lat ?? o.latitude;
						const lng = o.lng ?? o.lon ?? o.longitude;
						const trackId = o.track_id;

						if (lat != null && lng != null && trackId != null) {
							const latNum = typeof lat === 'string' ? parseFloat(lat) : Number(lat);
							const lngNum = typeof lng === 'string' ? parseFloat(lng) : Number(lng);

							if (!Number.isNaN(latNum) && !Number.isNaN(lngNum)) {
								const cameraName = getCameraName(data.camera_id);
								const popup = `
									<div style="font-size:13px">
										<strong>${cameraName}</strong><br/>
										Track ID: ${trackId}<br/>
										${new Date(data.timestamp).toLocaleString()}
									</div>`;

								latestDronePositions.set(trackId, {
									lngLat: [lngNum, latNum],
									popup,
									color: '#ef4444',
									icon: 'drone',
									kind: 'latest',
									timestamp: Date.now()
								});
							}
						}
					});

				// Limit drones on map
				if (latestDronePositions.size > MAX_DRONES_ON_MAP) {
					const sorted = Array.from(latestDronePositions.entries())
						.sort((a, b) => b[1].timestamp - a[1].timestamp);
					
					// Get IDs to keep
					const idsToKeep = new Set(sorted.slice(0, MAX_DRONES_ON_MAP).map(([id]) => id));
					
					latestDronePositions.clear();
					sorted.slice(0, MAX_DRONES_ON_MAP).forEach(([id, pos]) => {
						latestDronePositions.set(id, pos);
					});
					
					// Clean up paths for drones no longer tracked
					for (const trackId of defensiveDronePaths.keys()) {
						if (!idsToKeep.has(trackId)) {
							defensiveDronePaths.delete(trackId);
						}
					}
				}					// Update defensiveMarkers array - SAME AS DEFENSIVE DASHBOARD
					defensiveMarkers = Array.from(latestDronePositions.values()).map(({ timestamp, ...drone }) => drone);
					console.log('Total unique drones on defensive map:', defensiveMarkers.length);
				}
			} catch (error) {
				console.error('Error parsing detection WebSocket message:', error);
			}
		};

		ws.onclose = () => {
			console.log(`Detection WebSocket disconnected for camera: ${cameraId}`);
			detectionWsConnections.delete(cameraId);
			if (selectedCameraIds.has(cameraId)) {
				setTimeout(() => connectDetectionWebSocket(cameraId), 3000);
			}
		};

		detectionWsConnections.set(cameraId, ws);
	}

	onMount(() => {
		console.log('Battle Dashboard mounted');

		// Fetch initial data
		Promise.all([
			fetchAttackData(),
			fetchCameras()
		]);

		// Connect WebSockets
		connectAttackWebSocket();

		// Update time
		const t = setInterval(() => {
			currentTime = new Date();
		}, 1000);

		return () => {
			clearInterval(t);
			// Cleanup WebSockets
			attackWs?.close();
			detectionWsConnections.forEach(ws => ws.close());
		};
	});

	onDestroy(() => {
		attackWs?.close();
		detectionWsConnections.forEach(ws => ws.close());
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

                <button
                    onclick={() => goto("/")}
                    class="p-2 rounded-lg bg-white/20 hover:bg-white/30 transition-colors cursor-pointer"
                    title="Back to home"
                    aria-label="Back to home"
                >
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                    </svg>
                </button>

                <div>
                    <h1 class="m-0 text-xl font-bold">Battle Dashboard</h1>
                    <p class="m-0 opacity-90 text-xs">Drone & Camera Monitoring System</p>
                </div>

                <!-- Server Time - Real-time display -->
                <div class="flex items-center gap-2 px-4 py-2 text-sm font-medium border-l border-white/50">
                    <span class="text-lg">🕐</span>
                    <div>
                        <div class="text-xs opacity-75">Server Time</div>
                        <div class="font-mono font-bold">{formatDateTime(currentTime)}</div>
                    </div>
                </div>
            </div>

            <div class="flex items-center gap-6">
                <!-- Commander Info -->
                <div class="flex items-center gap-2 px-4 py-2 bg-white/20 rounded-4xl text-sm font-medium">
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
					<div class="flex justify-between items-center">
						<h2 class="m-0 text-base text-gray-800 font-bold flex items-center gap-2">
							<span>🚁</span>
							Offense
							<span class="bg-purple-500 text-white px-2 py-0.5 rounded-xl text-xs font-semibold">
								{droneGroups.length} Drones ({allDrones.length} paths)
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

				<div class="flex-1 flex flex-col gap-2 overflow-y-auto p-2 list-scrollbar">
					{#each filteredGroups as group (group.droneId)}
						<div class="mb-1">
							<!-- Group Header -->
							<div
								class="flex flex-col text-nowrap justify-between items-center p-2 rounded-lg border-2 bg-white cursor-pointer transition-all duration-200 hover:bg-gray-50"
								style="border-left: 3px solid {getDroneColor(group.droneId)};"
								role="button"
								tabindex="0"
								onclick={() => toggleGroupExpansion(group.droneId)}
								onkeydown={(e) => e.key === 'Enter' && toggleGroupExpansion(group.droneId)}
							>
								<div class="flex items-center mb-1 gap-2">
									<div 
										class="w-2 h-2 rounded-full" 
										style="background-color: {getDroneColor(group.droneId)};"
									></div>
									<span class="text-xs">{group.isExpanded ? '🔽' : '▶️'}</span>
									<span class="text-sm font-bold text-gray-800">{group.name}</span>
									<span class="text-xs text-gray-500">({group.paths.length})</span>
								</div>
								<button
									class="px-2 py-0.5 rounded-xl text-xs font-semibold"
									class:bg-green-100={group.lastStatus === 'connected'}
									class:text-green-800={group.lastStatus === 'connected'}
									class:bg-red-100={group.lastStatus === 'disconnected'}
									class:text-red-800={group.lastStatus === 'disconnected'}
									onclick={(e) => e.stopPropagation()}
								>
									{group.lastStatus === 'connected' ? '🟢 Connect' : '🔴 Disconnect'}
								</button>
							</div>

				
						</div>
					{/each}
					
				</div>
			</div>

			<!-- Drone Map -->
			<div class="bg-white rounded-xl shadow-md overflow-hidden flex-1 relative">
				<MapboxMap
					accessToken={mapboxToken}
					center={droneMapCenter}
					zoom={17}
					{markers}
					{pathLines}
				/>
				<div class="absolute bottom-3 right-3 bg-white px-3 py-2 rounded-lg shadow-md flex gap-3 text-xs">
					<span class="text-green-500">● สัญญาณ GPS</span>
					<span class="text-red-500">● ไม่มีสัญญาณ GPS</span>
				</div>
			</div>

			<!-- Drone History -->
			<div class="bg-white rounded-xl shadow-md p-3 overflow-y-auto scrollbar-thin" style="height: 35vh;">
				<div class="flex items-center justify-between mb-2">
					<h2 class="m-0 text-sm text-gray-800 font-bold flex items-center gap-2">
						<span>📜</span>
						{selectedDrone ? `ประวัติการเดินทาง - ${selectedDrone.name}` : 'ประวัติการเดินทาง - ทั้งหมด'}
					</h2>
					{#if selectedDrone}
						<button
							onclick={() => selectedDrone = null}
							class="px-3 py-1 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg text-xs font-medium transition-colors flex items-center gap-1"
						>
							<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
							</svg>
							ย้อนกลับ
						</button>
					{/if}
				</div>
				{#if selectedDrone}
					<!-- Single Selected Drone Detail -->
					<div class="p-2 mb-2 border-2 border-gray-200 rounded-lg">
						<h3 class="m-0 mb-2 text-xs font-bold text-purple-600">{selectedDrone.name}</h3>
						<div class="grid grid-cols-2 gap-x-4 gap-y-1 text-xs">
							<div>
								<span class="text-gray-600">ตำแหน่ง:</span>
								<span class="font-medium text-gray-800 ml-1">{selectedDrone.coordinates.lat.toFixed(4)}, {selectedDrone.coordinates.lng.toFixed(4)}</span>
							</div>
							<div>
								<span class="text-gray-600">ความสูง:</span>
								<span class="font-medium text-gray-800 ml-1">{selectedDrone.height.toFixed(2)} ม.</span>
							</div>
							<div>
								<span class="text-gray-600">ความเร็ว:</span>
								<span class="font-medium text-gray-800 ml-1">{Math.sqrt(selectedDrone.velocity.x**2 + selectedDrone.velocity.y**2 + selectedDrone.velocity.z**2).toFixed(2)} m/s</span>
							</div>
							<div>
								<span class="text-gray-600">ความเร่ง:</span>
								<span class="font-medium text-gray-800 ml-1">{Math.sqrt(selectedDrone.acceleration.x**2 + selectedDrone.acceleration.y**2 + selectedDrone.acceleration.z**2).toFixed(2)} m/s²</span>
							</div>
							<div>
								<span class="text-gray-600">ระยะทาง:</span>
								<span class="font-medium text-gray-800 ml-1">{selectedDrone.distance.toFixed(2)} ม.</span>
							</div>
							<div>
								<span class="text-gray-600">เวลาคงเหลือ:</span>
								<span class="font-medium text-gray-800 ml-1">{selectedDrone.timeLeft.toFixed(0)} วินาที</span>
							</div>
							{#if selectedDrone.target}
								<div class="col-span-2 mt-1 pt-1 border-t border-gray-200">
									<span class="text-gray-600">🎯 Target:</span>
									<span class="font-medium text-gray-800 ml-1">{selectedDrone.target.description}</span>
									<span class="text-gray-500 ml-1 text-xs">({selectedDrone.target.lat.toFixed(4)}, {selectedDrone.target.lng.toFixed(4)})</span>
								</div>
							{/if}
							{#if selectedDrone.landing}
								<div class="col-span-2">
									<span class="text-gray-600">🛬 Landing:</span>
									<span class="font-medium text-gray-800 ml-1">{selectedDrone.landing.description}</span>
									<span class="text-gray-500 ml-1 text-xs">({selectedDrone.landing.lat.toFixed(4)}, {selectedDrone.landing.lng.toFixed(4)})</span>
								</div>
							{/if}
							<div class="col-span-2">
								<span class="text-gray-600">เวลา:</span>
								<span class="font-medium text-gray-800 ml-1">{selectedDrone.lastUpdate}</span>
							</div>
						</div>
					</div>
				{:else}
					<!-- All Drones Overview -->
					<div class="space-y-2">
						{#each droneGroups as group (group.droneId)}
							<div class="border-2 border-gray-200 rounded-lg p-2">
								<div class="flex items-center gap-2 mb-2">
									<div 
										class="w-3 h-3 rounded-full" 
										style="background-color: {getDroneColor(group.droneId)};"
									></div>
									<h3 class="text-xs font-bold text-gray-800">{group.name}</h3>
									<span class="text-xs text-gray-500">({group.paths.length} paths)</span>
								</div>
								<div class="space-y-1">
									{#each group.paths as drone, index (drone.id)}
										<div 
											class="p-2 rounded bg-gray-50 border-l-2 hover:bg-gray-100 cursor-pointer transition-all"
											style="border-left-color: {getDroneColor(group.droneId)};"
											onclick={() => selectDrone(drone)}
											onkeydown={(e) => e.key === 'Enter' && selectDrone(drone)}
											role="button"
											tabindex="0"
										>
											<div class="flex justify-between items-center text-xs">
												<span class="font-semibold text-gray-700">Path #{index + 1}</span>
												<span class="text-gray-500">{drone.lastUpdate}</span>
											</div>
											<div class="text-xs text-gray-600 mt-1">
												lat. {drone.coordinates.lat.toFixed(4)}, lng. {drone.coordinates.lng.toFixed(4)} 
												| alt {drone.height.toFixed(2)}m 
												| velocity {Math.sqrt(drone.velocity.x**2 + drone.velocity.y**2 + drone.velocity.z**2).toFixed(2)}m/s
												| acceleration {Math.sqrt(drone.acceleration.x**2 + drone.acceleration.y**2 + drone.acceleration.z**2).toFixed(2)}m/s²
											</div>
										</div>
									{/each}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Right Side - Cameras -->
		<div class="w-1/2 flex flex-col gap-2 overflow-hidden">
			<!-- Camera Search & List -->
			<div class="bg-white rounded-xl shadow-md flex flex-col overflow-hidden">
				<div class="px-4 py-2 border-b border-gray-200 bg-gradient-to-br from-blue-50 to-blue-100">
					<div class="flex justify-between items-center">
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
								<span class="text-sm font-bold text-gray-800 mr-2">{camera.name}</span>
								<span
									class="w-1.5 h-1.5 rounded-full"
									class:bg-green-500={camera.status === 'online'}
									class:bg-red-500={camera.status === 'offline'}
								></span>
							</div>
							<div class="text-xs text-gray-600 flex items-center gap-1">
								<span>📍</span>
								lat. {camera.coordinates.lat.toFixed(4)}, long. {camera.coordinates.lng.toFixed(4)}
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
                            markers={defensiveMarkers}
                            pathLines={defensivePathLines}
                        />
                    </div>
                    <!-- Video Stream -->
                    <div class="bg-white rounded-xl shadow-md overflow-hidden flex-1">
                        {#if selectedCameraId}
                            <VideoStream serverUrl={videoServerUrl} />
                        {:else}
                            <div class="w-full h-full flex items-center justify-center text-gray-400">
                                <p>เลือกกล้องเพื่อดู Stream</p>
                            </div>
                        {/if}
                    </div>
                </div>
                <!-- Detection History -->
                <div class="w-1/3 bg-white rounded-xl shadow-md p-3">
					<!-- Model Upload Section - SAME AS DEFENSIVE DASHBOARD -->
					<div class="w-full">
						<div class="bg-gradient-to-r from-indigo-50 to-blue-50 rounded-lg p-2 mb-3">
							<div class="flex items-center gap-2">
								<span class="text-2xl">🤖</span>
								<div>
									<h3 class="font-bold text-gray-800 text-sm">Current Model</h3>
									<p class="text-xs text-gray-600">YOLO v8 NCNN Model 960</p>
								</div>
							</div>
						</div>

						<!-- <label for="file-upload-defensive" class="block text-sm font-semibold text-gray-700 mb-2">Upload New Model</label> -->
						<div class="relative">
							<input 
								id="file-upload-defensive" 
								type="file" 
								accept=".pt,.onnx,.pb,.tflite,.pth"
								class="absolute inset-0 w-full h-full opacity-0 cursor-pointer z-10"
								onchange={uploadModel}
								bind:this={fileInputRef}
								disabled={isUploadingModel}
							/>
							<div class="flex items-center justify-center gap-2 px-4 py-3 bg-indigo-50 border-2 border-dashed border-indigo-300 rounded-lg hover:bg-indigo-100 hover:border-indigo-400 transition-all cursor-pointer {isUploadingModel ? 'opacity-50 cursor-not-allowed' : ''}">
								{#if isUploadingModel}
									<span class="text-lg animate-spin">⌛</span>
									<span class="text-sm font-medium text-indigo-700">Uploading...</span>
								{:else}
									<span class="text-lg">📦</span>
									<span class="text-sm font-medium text-indigo-700">Upload New Model</span>
								{/if}
							</div>
							<p class="text-xs text-gray-500 mt-1.5 text-center">Supported: .pt, .onnx, .pb, .tflite</p>
							{#if uploadProgress}
								<p class="text-xs text-green-600 mt-1.5 text-center font-medium">{uploadProgress}</p>
							{/if}
						</div>
					</div>					
                    <div class="flex flex-col gap-1 border-y border-gray-200 pb-4 my-2">
                        <h2 class="m-0 text-lg flex items-center gap-2 text-gray-800 font-bold mt-1">Search History</h2>
						<div class="flex-1 mt-1">
							<label for="start-date" class="block text-sm font-medium text-gray-700 mb-1">Start Date</label>
							<input id="start-date" type="date" bind:value={startDate} class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm" />
						</div>
						<div class="flex-1">
							<label for="end-date" class="block text-sm font-medium text-gray-700 mb-1">End Date</label>
							<input id="end-date" type="date" bind:value={endDate} class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm" />
						</div>
						{#if searchError}
							<p class="text-xs text-red-500 mt-1">{searchError}</p>
						{/if}
						<button 
							onclick={searchDetectionHistory}
							disabled={isSearching || !startDate || !endDate}
							class="w-full px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors font-medium text-sm"
						>
							{#if isSearching}
								กำลังค้นหา...
							{:else}
								ค้นหา
							{/if}
						</button>
						{#if filteredDetections.length > 0}
							<p class="text-sm text-green-600 text-center font-medium">
								พบ {filteredDetections.length} รายการ
							</p>
						{/if}
					</div>
					
					<!-- Detection Feed - SAME AS DEFENSIVE DASHBOARD -->
					<div class="w-full pt-2">
						<div class="flex justify-between items-center mb-3">
							<h2 class="m-0 text-lg flex items-center gap-2 text-gray-800 font-bold">
								<span class="inline-block">📋</span>
								All Detections
								<span class="bg-indigo-500 text-white px-2 rounded-xl text-xs font-semibold">
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
								{#each detections.slice(0, 10) as detection (detection.id)}
									<div class="p-2 border-2 border-gray-200 rounded-lg hover:bg-gray-50 transition-colors cursor-pointer">
										<div class="flex items-start gap-2">
											<!-- Image Preview -->
											<div class="w-16 h-16 shrink-0 rounded-lg overflow-hidden bg-gray-100">
												{#if detection.image_base64}
													<img
														src="data:image/jpeg;base64,{detection.image_base64}"
														alt="Detection {detection.id}"
														class="w-full h-full object-cover"
													/>
												{:else}
													<img
														src="{apiUrl}/detect/{detection.id}/file"
														alt="Detection {detection.id}"
														class="w-full h-full object-cover"
														onerror={(e) => { e.currentTarget.src = 'data:image/svg+xml,%3Csvg xmlns=\"http://www.w3.org/2000/svg\" width=\"64\" height=\"64\"%3E%3Ctext x=\"50%\" y=\"50%\" font-size=\"32\" text-anchor=\"middle\" dy=\".3em\"%3E📷%3C/text%3E%3C/svg%3E'; }}
													/>
												{/if}
											</div>
											<div class="flex-1 min-w-0">
												<div class="text-xs font-bold text-gray-800 truncate">
													📹 {getCameraName(detection.cameraId)}
												</div>
												<div class="text-xs text-gray-500 mt-1">
													🕐 {detection.detectedAt}
												</div>
												{#if detection.objects && detection.objects[0]}
													<div class="text-xs text-red-600 font-semibold mt-1">
														({detection.objects[0].lat !== undefined ? detection.objects[0].lat.toFixed(6) : '-'},{detection.objects[0].lon !== undefined ? detection.objects[0].lon.toFixed(6) : '-'}) Alt: {detection.objects[0].alt !== undefined ? detection.objects[0].alt.toFixed(2) : '-'}m
													</div>
												{/if}
											</div>
										</div>
									</div>
								{/each}
							{/if}
						</div>
					</div>
                </div>
            </div>
		</div>
	</main>
</div>

<!-- Search Results Modal - SAME AS DEFENSIVE DASHBOARD -->
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
								cameraName={detection.cameraName || 'GearDinDaeng2025'}
								onClick={() => {
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