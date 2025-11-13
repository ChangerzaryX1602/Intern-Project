<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import MapboxMap from '$lib/components/MapboxMap.svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import { env } from '$env/dynamic/public';
	import { goto } from '$app/navigation';

	const mapboxToken = env.PUBLIC_MAPBOX_TOKEN || '';
	const WS_URL = 'ws://localhost:8080/api/v1/detect/attack-ws';
	const API_URL = 'http://localhost:8080/api/v1/attack';
	const DUMMY_CAMERA_ID = '00000000-0000-0000-0000-000000000000'; // UUID placeholder for attack data (not used for attack-ws)

	// Attack data from WebSocket (updated to match new Go model)
	interface Vector3 {
		x: number;
		y: number;
		z: number;
	}

	interface TargetLocation {
		lat: number;
		lng: number;
		target_destcription: string;
	}

	interface AttackData {
		id: number;
		drone_id: string;
		lat: number;
		lng: number;
		height: number;
		function: string;
		acceleration: Vector3; // Changed from number to Vector3
		velocity: Vector3;     // Changed from number to Vector3
		distance: number;
		status: string;
		time_left: number;          // New field
		target?: TargetLocation;    // New field (optional)
		land?: TargetLocation;      // New field (optional)
		created_at: string;
	}

	// Transformed drone data for UI
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
		velocity: Vector3;      // Changed to Vector3
		acceleration: Vector3;  // Changed to Vector3
		distance: number;
		timeLeft: number;       // New field
		target?: TargetLocation;  // New field
		land?: TargetLocation;    // New field
	}

	// Grouped drone data
	interface DroneGroup {
		droneId: string;
		name: string;
		paths: Drone[];
		isExpanded: boolean;
		lastStatus: 'connected' | 'disconnected';
		lastGpsStatus: 'good' | 'loss';
		latestUpdate: string;
	}

	// Color palette for random drone colors
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
	let isConnected = $state(false);
	let isLoadingInitial = $state(true);
	let ws: WebSocket | null = null;
	let reconnectTimeout: any = null;
	let error = $state<string | null>(null);

	let selectedDrone = $state<Drone | null>(null);
	let searchQuery = $state('');
	let mapCenter: [number, number] = $state([100.5018, 13.7563]); // [lng, lat] - กรุงเทพ

	// Fetch initial attack data
	async function fetchInitialData() {
		try {
			isLoadingInitial = true;
			error = null;

			const response = await fetch(API_URL);
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const result = await response.json();

			if (result.success && result.data.attacks) {
				// Transform and add initial data
				const drones = result.data.attacks.map((attack: AttackData) => transformAttackToDrone(attack));
				allDrones = drones;
				updateDroneGroups();

				// Set map center to first drone
				if (allDrones.length > 0) {
					mapCenter = [allDrones[0].coordinates.lng, allDrones[0].coordinates.lat];
				}
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to fetch initial data';
			console.error('Error fetching initial data:', e);
		} finally {
			isLoadingInitial = false;
		}
	}

	// Transform attack data to drone
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
			land: attack.land
		};
	}

	// Update drone groups (group by drone_id)
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

	// Connect to WebSocket
	function connectWebSocket() {
		if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
			return;
		}

		try {
			ws = new WebSocket(WS_URL);

			ws.onopen = () => {
				console.log('✅ WebSocket connected for attack data');
				isConnected = true;
				error = null;

				// No need to send camera_id for attack-ws
				// Server will automatically send confirmation
			};

			ws.onmessage = (event) => {
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
							mapCenter = [latest.coordinates.lng, latest.coordinates.lat];
						}
					}
				} catch (e) {
					console.error('Error parsing WebSocket message:', e);
				}
			};

			ws.onerror = (e) => {
				console.error('❌ WebSocket error:', e);
				error = 'WebSocket connection error';
			};

			ws.onclose = () => {
				console.log('🔌 WebSocket disconnected');
				isConnected = false;

				// Auto-reconnect after 3 seconds
				reconnectTimeout = setTimeout(() => {
					console.log('🔄 Attempting to reconnect...');
					connectWebSocket();
				}, 3000);
			};
		} catch (e) {
			console.error('Failed to create WebSocket:', e);
			error = 'Failed to connect to WebSocket';
		}
	}

	// Disconnect WebSocket
	function disconnectWebSocket() {
		if (reconnectTimeout) {
			clearTimeout(reconnectTimeout);
			reconnectTimeout = null;
		}

		if (ws) {
			ws.close();
			ws = null;
		}

		isConnected = false;
	}

	// Toggle group expansion
	function toggleGroupExpansion(droneId: string) {
		const group = droneGroups.find((g) => g.droneId === droneId);
		if (group) {
			group.isExpanded = !group.isExpanded;
		}
	}

	// Generate markers for map: only start point (small colored marker) + latest point (drone sticker)
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
				id: group.droneId,
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

			// Latest marker (latest path) - rendered as drone sticker
			const latest = group.paths[group.paths.length - 1];
			const velocityMag = Math.sqrt(latest.velocity.x**2 + latest.velocity.y**2 + latest.velocity.z**2);
			const accelerationMag = Math.sqrt(latest.acceleration.x**2 + latest.acceleration.y**2 + latest.acceleration.z**2);
			out.push({
				id: group.droneId,
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
						${latest.target ? `🎯 ${latest.target.target_destcription}<br/>` : ''}
						${latest.lastUpdate}
					</div>`,
				color,
				icon: 'drone'
			});
		});

		return out;
	});

	// Generate path lines with ALL coordinates (not just start and latest)
	let pathLines = $derived.by(() => {
		const lines: Array<any> = [];

		droneGroups.forEach((group) => {
			if (!group.isExpanded) return;
			if (!group.paths || group.paths.length < 2) return; // Need at least 2 points for a line

			const color = getDroneColor(group.droneId);
			
			// Build full path coordinates array
			const coordinates = group.paths.map(drone => 
				[drone.coordinates.lng, drone.coordinates.lat] as [number, number]
			);

			lines.push({
				id: group.droneId,
				coordinates,
				color
			});
		});

		return lines;
	});

	// Filter drone groups based on search
	let filteredGroups = $derived(
		droneGroups.filter((g) => g.name.toLowerCase().includes(searchQuery.toLowerCase()))
	);

	function selectDrone(drone: Drone) {
		selectedDrone = drone;
		// Center map on selected drone
		if (drone.status === 'connected') {
			mapCenter = [drone.coordinates.lng, drone.coordinates.lat];
		}
	}

	function toggleDroneStatus(droneId: string) {
		const drone = allDrones.find((d) => d.id === droneId);
		if (drone) {
			drone.status = drone.status === 'connected' ? 'disconnected' : 'connected';
			if (drone.status === 'disconnected') {
				drone.gpsStatus = 'loss';
			}
			updateDroneGroups();
		}
	}

	onMount(async () => {
		console.log('Offensive Dashboard mounted');
		// Fetch initial data first
		await fetchInitialData();
		// Then connect to WebSocket for real-time updates
		connectWebSocket();
	});

	onDestroy(() => {
		console.log('Offensive Dashboard unmounted');
		// Disconnect WebSocket
		disconnectWebSocket();
	});

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

	// Notification settings
	let notificationSettings = $state({
		email: false,
		line: false
	});
	let showNotificationMenu = $state(false);

	$effect(() => {
		const interval = setInterval(() => {
			currentTime = new Date();
		}, 1000);

		return () => clearInterval(interval);
	});

	function toggleNotification(type: 'email' | 'line') {
		notificationSettings[type] = !notificationSettings[type];
	}

	function getNotificationCount(): number {
		return Object.values(notificationSettings).filter(Boolean).length;
	}
</script>

<svelte:head>
	<title>Offensive Dashboard - Drone Control</title>
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
					<h1 class="m-0 text-xl font-bold">Offensive Dashboard</h1>
					<p class="m-0 opacity-90 text-xs">Drone Control & Monitoring</p>
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
			<!-- Notification Settings -->
			<div class="relative">
				<button
					onclick={() => (showNotificationMenu = !showNotificationMenu)}
					class="flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors cursor-pointer"
					title="Notification settings"
				>
					<span
						class="text-lg transition-all"
						class:pulse={detectionsCount > 0}
					>🔔</span>

					<span>Notifications</span>
					{#if getNotificationCount() > 0}
						<span class="ml-1 inline-block px-1 bg-white/30 rounded-full">ON</span>
					{:else}
						<span class="ml-1 inline-block px-1 bg-white/30 rounded-full">OFF</span>
					{/if}
				</button>

				<!-- Notification Menu -->
				{#if showNotificationMenu}
					<div class="absolute right-0 mt-2 w-56 bg-white rounded-lg shadow-xl z-50 p-4">
						<h3 class="text-sm font-bold text-gray-800 mb-3">Choose notification channels:</h3>
						<div class="space-y-3">
							<!-- Email Notification -->
							<label class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 cursor-pointer transition-colors">
								<input
									type="checkbox"
									checked={notificationSettings.email}
									onchange={() => toggleNotification('email')}
									class="w-4 h-4 text-indigo-600 rounded cursor-pointer"
								/>
								<div class="flex-1">
									<div class="text-sm font-medium text-gray-800">📧 Email</div>
									<div class="text-xs text-gray-500">Receive alerts via email</div>
								</div>
								{#if notificationSettings.email}
									<span class="text-lg">✓</span>
								{/if}
							</label>

							<!-- Line Notification -->
							<label class="flex items-center gap-3 p-3 rounded-lg hover:bg-gray-50 cursor-pointer transition-colors">
								<input
									type="checkbox"
									checked={notificationSettings.line}
									onchange={() => toggleNotification('line')}
									class="w-4 h-4 text-indigo-600 rounded cursor-pointer"
								/>
								<div class="flex-1">
									<div class="text-sm font-medium text-gray-800">💬 Line</div>
									<div class="text-xs text-gray-500">Receive alerts via Line</div>
								</div>
								{#if notificationSettings.line}
									<span class="text-lg">✓</span>
								{/if}
							</label>
						</div>

						<!-- Active Channels Summary -->
						{#if getNotificationCount() > 0}
							<div class="mt-4 pt-3 border-t border-gray-200">
								<div class="text-xs text-gray-600">
									<span class="font-medium">{getNotificationCount()} channel{getNotificationCount() > 1 ? 's' : ''} active:</span>
									<div class="mt-1 flex gap-2">
										{#if notificationSettings.email}
											<span class="inline-block px-2 py-1 bg-blue-100 text-blue-700 text-xs rounded">📧 Email</span>
										{/if}
										{#if notificationSettings.line}
											<span class="inline-block px-2 py-1 bg-green-100 text-green-700 text-xs rounded">💬 Line</span>
										{/if}
									</div>
								</div>
							</div>
						{:else}
							<div class="mt-4 pt-3 border-t border-gray-200 text-center">
								<div class="text-xs text-gray-500">No notifications enabled</div>
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Commander Info -->
			<div class="flex items-center gap-2 px-4 py-2 bg-white/20 rounded-4xl text-sm font-medium">
				<span class="text-lg">👨‍💼</span>
				<span>commander</span>
			</div>
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
				placeholder="🔍 ค้นหา Drone"
				label=""
				onSearch={() => {}}
			/>

			<!-- Drone List Header -->
			<div class="px-5 py-4 border-b border-gray-200">
				<h2 class="m-0 text-lg text-gray-800 font-bold">
					{#if isLoadingInitial}
						🔄 กำลังโหลดข้อมูล...
					{:else if !isConnected}
						⚠️ กำลังเชื่อมต่อ WebSocket...
					{:else if error}
						❌ เกิดข้อผิดพลาด
					{:else}
						ทั้งหมด {droneGroups.length} Drone ({allDrones.length} paths)
					{/if}
				</h2>
				{#if error}
					<p class="text-sm text-red-500 mt-1">{error}</p>
					<button
						class="mt-2 px-3 py-1 bg-red-500 text-white text-sm rounded hover:bg-red-600 transition-colors"
						onclick={connectWebSocket}
					>
						เชื่อมต่อใหม่
					</button>
				{/if}
			</div>

			<!-- Drone Cards (Grouped) -->
			<div class="flex-1 overflow-y-auto p-3 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 hover:scrollbar-thumb-gray-400">
				{#if isLoadingInitial}
					<div class="flex items-center justify-center h-full">
						<div class="text-center">
							<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-red-500 mx-auto mb-4"></div>
							<p class="text-gray-500">กำลังโหลดข้อมูลเริ่มต้น...</p>
						</div>
					</div>
				{:else if !isConnected && allDrones.length === 0}
					<div class="flex items-center justify-center h-full">
						<div class="text-center">
							<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-red-500 mx-auto mb-4"></div>
							<p class="text-gray-500">กำลังเชื่อมต่อ WebSocket...</p>
						</div>
					</div>
				{:else if filteredGroups.length === 0}
					<div class="flex items-center justify-center h-full">
						<div class="text-center">
							<p class="text-gray-500 mb-2">📡 {searchQuery ? 'ไม่พบข้อมูลที่ค้นหา' : 'รอข้อมูล Drone...'}</p>
							{#if !searchQuery}
								<p class="text-xs text-gray-400">เชื่อมต่อสำเร็จ รอข้อมูลจาก Server</p>
							{/if}
						</div>
					</div>
				{:else}
					{#each filteredGroups as group (group.droneId)}
						<!-- Group Header (Collapsible) -->
						<div
							class="flex flex-col p-4 rounded-lg border-2 bg-white mb-3 cursor-pointer transition-all duration-200 hover:bg-gray-50"
							style="border-left: 4px solid {getDroneColor(group.droneId)};"
							class:border-gray-200={!group.isExpanded}
							class:shadow-md={group.isExpanded}
							style:background-color={group.isExpanded ? `${getDroneColor(group.droneId)}10` : 'white'}
							onclick={() => toggleGroupExpansion(group.droneId)}
							role="button"
							tabindex="0"
							onkeydown={(e) => e.key === 'Enter' && toggleGroupExpansion(group.droneId)}
						>
							<div class="flex justify-between items-center">
								<div class="flex items-center gap-3">
									<!-- Color indicator circle -->
									<div 
										class="w-3 h-3 rounded-full" 
										style="background-color: {getDroneColor(group.droneId)};"
									></div>
									<span class="text-base">
										{group.isExpanded ? '🔽' : '▶️'}
									</span>
									<div>
										<span class="text-base font-bold text-gray-800">{group.name}</span>
										<span class="ml-2 text-xs text-gray-500">({group.paths.length} paths)</span>
									</div>
								</div>
								<button
									class="px-3 py-1 rounded-xl text-xs font-semibold flex items-center gap-1.5 border-none cursor-pointer transition-all duration-200 hover:scale-105"
									class:bg-green-100={group.lastStatus === 'connected'}
									class:text-green-800={group.lastStatus === 'connected'}
									class:bg-red-100={group.lastStatus === 'disconnected'}
									class:text-red-800={group.lastStatus === 'disconnected'}
									onclick={(e) => {
										e.stopPropagation();
									}}
								>
									<span
										class="w-2 h-2 rounded-full"
										class:bg-green-500={group.lastStatus === 'connected'}
										class:animate-pulse-slow={group.lastStatus === 'connected'}
										class:bg-red-500={group.lastStatus === 'disconnected'}
									></span>
									{#if group.lastStatus === 'connected'}
										Connect
									{:else}
										Disconnect
									{/if}
								</button>
							</div>

							<!-- Latest info (always visible) -->
							<div class="mt-2 flex gap-4 text-xs text-gray-600">
								<span>📍 {group.paths[group.paths.length - 1].coordinates.lat.toFixed(4)}, {group.paths[group.paths.length - 1].coordinates.lng.toFixed(4)}</span>
								<span>⏰ {group.latestUpdate}</span>
							</div>
						</div>

						<!-- Expanded Paths -->
						{#if group.isExpanded}
							<div class="ml-4 mb-3 space-y-2">
								{#each group.paths as drone, index (drone.id)}
									<div
										class="flex flex-col p-3 rounded-lg border-2 bg-gray-50 hover:bg-gray-100 transition-all duration-200 cursor-pointer"
										style="border-left: 3px solid {getDroneColor(group.droneId)};"
										class:border-blue-400={selectedDrone?.id === drone.id}
										class:bg-blue-50={selectedDrone?.id === drone.id}
										class:shadow={selectedDrone?.id === drone.id}
										onclick={(e) => {
											e.stopPropagation();
											selectDrone(drone);
										}}
										onkeydown={(e) => {
											if (e.key === 'Enter' || e.key === ' ') {
												e.stopPropagation();
												e.preventDefault();
												selectDrone(drone);
											}
										}}
										role="button"
										tabindex="0"
									>
										<div class="flex justify-between items-center mb-2">
											<span class="text-sm font-semibold text-gray-700">
												📍 Path #{index + 1}
											</span>
											<span
												class="px-2 py-0.5 rounded text-xs"
												class:bg-green-100={drone.gpsStatus === 'good'}
												class:text-green-700={drone.gpsStatus === 'good'}
												class:bg-red-100={drone.gpsStatus !== 'good'}
												class:text-red-700={drone.gpsStatus !== 'good'}
											>
												GPS {drone.gpsStatus === 'good' ? '✓' : '✗'}
											</span>
										</div>
										<div class="grid grid-cols-2 gap-x-4 gap-y-1 text-xs">
											<div>
												<span class="text-gray-600">ตำแหน่ง:</span>
												<span class="font-medium text-gray-800 ml-1">{drone.coordinates.lat.toFixed(4)}, {drone.coordinates.lng.toFixed(4)}</span>
											</div>
											<div>
												<span class="text-gray-600">ความสูง:</span>
												<span class="font-medium text-gray-800 ml-1">{drone.height.toFixed(2)} ม.</span>
											</div>
											<div>
												<span class="text-gray-600">ความเร็ว:</span>
												<span class="font-medium text-gray-800 ml-1">{Math.sqrt(drone.velocity.x**2 + drone.velocity.y**2 + drone.velocity.z**2).toFixed(2)} m/s</span>
											</div>
											<div>
												<span class="text-gray-600">ความเร่ง:</span>
												<span class="font-medium text-gray-800 ml-1">{Math.sqrt(drone.acceleration.x**2 + drone.acceleration.y**2 + drone.acceleration.z**2).toFixed(2)} m/s²</span>
											</div>
											<div>
												<span class="text-gray-600">ระยะทาง:</span>
												<span class="font-medium text-gray-800 ml-1">{drone.distance.toFixed(2)} ม.</span>
											</div>
											<div>
												<span class="text-gray-600">เวลาคงเหลือ:</span>
												<span class="font-medium text-gray-800 ml-1">{drone.timeLeft.toFixed(0)} วินาที</span>
											</div>
											{#if drone.target}
												<div class="col-span-2 mt-1 pt-1 border-t border-gray-200">
													<span class="text-gray-600">🎯 Target:</span>
													<span class="font-medium text-gray-800 ml-1">{drone.target.target_destcription}</span>
													<span class="text-gray-500 ml-1 text-xs">({drone.target.lat.toFixed(4)}, {drone.target.lng.toFixed(4)})</span>
												</div>
											{/if}
											{#if drone.land}
												<div class="col-span-2">
													<span class="text-gray-600">🛬 Landing:</span>
													<span class="font-medium text-gray-800 ml-1">{drone.land.target_destcription}</span>
													<span class="text-gray-500 ml-1 text-xs">({drone.land.lat.toFixed(4)}, {drone.land.lng.toFixed(4)})</span>
												</div>
											{/if}
											<div class="col-span-2">
												<span class="text-gray-600">เวลา:</span>
												<span class="font-medium text-gray-800 ml-1">{drone.lastUpdate}</span>
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					{/each}
				{/if}
			</div>
		</aside>

		<!-- Right Map Section -->
		<section class="w-[70%] bg-white rounded-xl shadow-md overflow-hidden relative">
			<div class="w-full h-full relative">
				<MapboxMap accessToken={mapboxToken} center={mapCenter} zoom={17} {markers} {pathLines} />

				<!-- Map Legend -->
				<div class="absolute bottom-6 left-6 bg-white px-4 py-3 rounded-lg shadow-md z-5">
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
				<div class="absolute top-6 right-6 w-[320px] bg-white rounded-xl shadow-xl p-6 z-10 border-2 border-red-500">
					<button
						class="absolute top-3 right-3 w-7 h-7 border-none bg-gray-100 rounded-full cursor-pointer flex items-center justify-center text-base text-gray-600 transition-all duration-200 hover:bg-gray-200 hover:text-gray-800"
						onclick={() => (selectedDrone = null)}
					>
						✕
					</button>
					<h3 class="m-0 mb-4 text-2xl text-red-500 font-bold">{selectedDrone.name}</h3>
					<div class="flex flex-col gap-3">
						<div class="flex justify-between items-center text-sm pb-2 border-b border-gray-200">
							<span class="text-gray-600">สถานะ GPS:</span>
							<span
								class="px-3 py-1 rounded-xl text-xs font-semibold"
								class:bg-green-100={selectedDrone.gpsStatus === 'good'}
								class:text-green-800={selectedDrone.gpsStatus === 'good'}
								class:bg-red-100={selectedDrone.gpsStatus !== 'good'}
								class:text-red-800={selectedDrone.gpsStatus !== 'good'}
							>
								{selectedDrone.gpsStatus === 'good' ? '✓ Connected' : '✗ Loss'}
							</span>
						</div>
						<div class="flex justify-between items-center text-sm">
							<span class="text-gray-600">พิกัด Lat:</span>
							<span class="text-gray-800 font-mono font-semibold">{selectedDrone.coordinates.lat.toFixed(6)}</span>
						</div>
						<div class="flex justify-between items-center text-sm">
							<span class="text-gray-600">พิกัด Lng:</span>
							<span class="text-gray-800 font-mono font-semibold">{selectedDrone.coordinates.lng.toFixed(6)}</span>
						</div>
						<div class="flex justify-between items-center text-sm">
							<span class="text-gray-600">ความสูง:</span>
							<span class="text-gray-800 font-semibold">{selectedDrone.height} เมตร</span>
						</div>
						<div class="flex justify-between items-center text-sm">
							<span class="text-gray-600">ความเร็ว:</span>
							<span class="text-gray-800 font-semibold">{selectedDrone.velocity} km/h</span>
						</div>
						<div class="flex justify-between items-center text-sm">
							<span class="text-gray-600">ความเร่ง:</span>
							<span class="text-gray-800 font-semibold">{selectedDrone.acceleration} m/s²</span>
						</div>
						<div class="flex justify-between items-center text-sm">
							<span class="text-gray-600">ระยะทาง:</span>
							<span class="text-gray-800 font-semibold">{selectedDrone.distance} เมตร</span>
						</div>
						<div class="flex justify-between items-center text-sm pt-2 border-t border-gray-200">
							<span class="text-gray-600">อัพเดทล่าสุด:</span>
							<span class="text-gray-800 text-xs">{selectedDrone.lastUpdate}</span>
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

	@keyframes pulse {
		0% { transform: scale(1); }
		50% { transform: scale(1.3); }
		100% { transform: scale(1); }
	}

	.pulse {
		animation: pulse 1s infinite ease-in-out;
		filter: drop-shadow(0 0 8px rgba(255, 255, 0, 0.6));
	}
</style>
