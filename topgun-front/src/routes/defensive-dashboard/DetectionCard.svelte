<script lang="ts">
	import { onMount } from 'svelte';
	import type { Detection } from './types';
	import { env } from '$env/dynamic/public';

	interface Props {
		detection: Detection;
		isSelected: boolean;
		cameraName: string;
		onClick: () => void;
	}

	let { detection, isSelected, cameraName, onClick }: Props = $props();
	let showModal = $state(false);
	let imageData = $state<string>('');
	let isLoadingImage = $state(false);

	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080/api/v1';

	// Parse objects data
	function parseObjects() {
		// Try new format first (objects)
		const objectsArray = detection.objects || detection.detected_objects;
		
		console.log('Detection data:', detection);
		console.log('Objects array:', objectsArray);
		
		if (!objectsArray || objectsArray.length === 0) {
			console.log('No objects found');
			return [];
		}
		
		const parsed = objectsArray.map((obj: any) => {
			// If obj is string, try to parse it
			if (typeof obj === 'string') {
				try {
					return JSON.parse(obj);
				} catch {
					return obj;
				}
			}
			return obj;
		});
		
		console.log('Parsed objects:', parsed);
		return parsed;
	}

	// Load image lazily
	async function loadImage() {
		if (imageData || isLoadingImage || !detection.id) return;
		
		isLoadingImage = true;
		try {
			const response = await fetch(`${apiUrl}/detect/${detection.id}/file`);
			if (response.ok) {
				const blob = await response.blob();
				imageData = URL.createObjectURL(blob);
			}
		} catch (error) {
			console.error('Failed to load image:', error);
		} finally {
			isLoadingImage = false;
		}
	}

	// Start loading image when card mounts
	onMount(() => {
		loadImage();
	});

	function formatDate(dateString?: string): string {
		if (!dateString) return '-';
		const date = new Date(dateString);
		if (isNaN(date.getTime())) return '-';
		const hours = String(date.getHours()).padStart(2, '0');
		const minutes = String(date.getMinutes()).padStart(2, '0');
		const seconds = String(date.getSeconds()).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const year = date.getFullYear();
		return `${hours}:${minutes}:${seconds} ${day}/${month}/${year}`;
	}

	function handleCardClick() {
		showModal = true;
		onClick();
	}

	function closeModal() {
		showModal = false;
	}

	function formatLatLng(value: string | number | undefined): string {
		if (value === undefined || value === null) return '-';
		const num = typeof value === 'string' ? parseFloat(value) : value;
		if (Number.isNaN(num) || !Number.isFinite(num)) return '-';
		return num.toFixed(6);
	}

	function formatNumber(value: string | number | undefined, decimals: number = 2): string {
		if (value === undefined || value === null) return '-';
		const num = typeof value === 'string' ? parseFloat(value) : value;
		if (Number.isNaN(num) || !Number.isFinite(num)) return '-';
		return num.toFixed(decimals);
	}

	function formatTimestamp(timestamp: number | undefined): string {
		if (!timestamp || timestamp === 0) return '-';
		try {
			const date = new Date(timestamp * 1000);
			if (isNaN(date.getTime())) return '-';
			return date.toLocaleString('th-TH');
		} catch {
			return '-';
		}
	}

	const objects = $derived(parseObjects());

</script>

<button 
	class="flex gap-3 p-3.5 rounded-xl border-2 cursor-pointer transition-all duration-200 min-w-[280px] shrink-0 text-left hover:-translate-y-0.5 hover:shadow-[0_4px_12px_rgba(0,0,0,0.08)]" 
	class:border-gray-100={!isSelected}
	class:bg-white={!isSelected}
	onclick={handleCardClick}>
	<div class="relative w-20 h-20 shrink-0 rounded-lg overflow-hidden bg-gray-100">
		{#if isLoadingImage}
			<div class="w-full h-full flex items-center justify-center">
				<div class="animate-spin text-2xl">⏳</div>
			</div>
		{:else if imageData}
			<img
				src={imageData}
				alt="Detection {detection.id}"
				class="w-full h-full object-cover"
			/>
		{:else if detection.image_base64}
			<img
				src="data:image/jpeg;base64,{detection.image_base64}"
				alt="Detection {detection.id}"
				class="w-full h-full object-cover"
			/>
		{:else}
			<div class="w-full h-full flex items-center justify-center text-3xl text-gray-300">
				<span>📷</span>
			</div>
		{/if}
	</div>
	<div class="flex-1 min-w-0 flex flex-col justify-between gap-2">
		<div>
			<div class="text-sm font-bold text-gray-900 flex items-center gap-1.5 whitespace-nowrap overflow-hidden text-ellipsis">
				<span class="text-base">📹</span>
				<span class="truncate">{cameraName}</span>
			</div>
			<div class="text-xs text-gray-500 mt-1 flex items-center gap-1">
				<span>🕐</span>
				<span>{formatDate(detection.detected_at)}</span>
			</div>
		</div>
		{#if objects.length > 0}
			<div class="flex items-center gap-2 pt-1 border-t border-gray-100">				
<span class="text-xs font-semibold text-red-600">({objects[0].lat !== undefined ? formatLatLng(objects[0].lat) : '-'},{objects[0].lon !== undefined ? formatLatLng(objects[0].lon) : '-'}) Attitude: {objects[0].alt !== undefined ? formatNumber(objects[0].alt) : '-'}</span>
			</div>
		{/if}
	</div>
</button>

<!-- Modal Popup -->
{#if showModal}
	<div 
		class="fixed inset-0 bg-black bg-opacity-20 backdrop-blur-sm flex items-center justify-center z-50" 
		role="dialog" 
		aria-modal="true"
		tabindex="-1"
		onclick={closeModal}
		onkeydown={(e) => e.key === 'Escape' && closeModal()}>
		<div 
			class="bg-white rounded-xl shadow-2xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto" 
			role="presentation"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}>
			<!-- Header -->
			<div class="sticky top-0 bg-gradient-to-r from-indigo-600 to-blue-600 text-white px-6 py-4 flex justify-between items-center shadow-md">
				<div class="flex items-center gap-2">
					<span class="text-2xl">📹</span>
					<div>
						<h2 class="text-xl font-bold">{cameraName}</h2>
						<p class="text-sm text-indigo-100">{formatDate(detection.detected_at)}</p>
					</div>
				</div>
				<button 
					type="button"
					onclick={closeModal} 
					class="text-2xl hover:bg-white hover:bg-opacity-20 rounded-full w-8 h-8 flex items-center justify-center transition"
					aria-label="ปิด">
					✕
				</button>
			</div>

			<!-- Content -->
			<div class="p-6 space-y-6">
				<!-- Detection Image -->
				{#if isLoadingImage}
					<div class="rounded-lg overflow-hidden bg-gray-200 h-96 flex items-center justify-center">
						<div class="text-center">
							<div class="text-6xl animate-spin mb-4">⏳</div>
							<p class="text-gray-600">กำลังโหลดรูปภาพ...</p>
						</div>
					</div>
				{:else if imageData}
					<div class="rounded-lg overflow-hidden bg-gray-200">
						<img
							src={imageData}
							alt="Detection"
							class="w-full h-auto max-h-96 object-cover"
						/>
					</div>
				{:else if detection.image_base64}
					<div class="rounded-lg overflow-hidden bg-gray-200">
						<img
							src="data:image/jpeg;base64,{detection.image_base64}"
							alt="Detection"
							class="w-full h-auto max-h-96 object-cover"
						/>
					</div>
				{/if}

				<!-- Drone Details -->
				{#if objects.length > 0}
					<div>
						<h3 class="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
							<span class="text-xl">🚁</span>
							ข้อมูลโดรนที่ตรวจพบ ({objects.length} ตำแหน่ง)
						</h3>
						
						<div class="space-y-4">
							{#each objects as obj, index}
								<div class="border-l-4 border-indigo-500 bg-indigo-50 p-4 rounded-lg">
									<div class="font-bold text-gray-900 mb-3 text-base">
										🆔 Track ID: {obj.track_id !== undefined ? obj.track_id : `Object_${index + 1}`}
									</div>
									
									<div class="grid grid-cols-2 gap-3 text-sm">
										{#if obj.lat !== undefined}
											<div class="flex flex-col gap-1">
												<span class="text-gray-600 font-medium">📍 Latitude</span>
												<span class="text-gray-900 font-mono text-xs">{formatLatLng(obj.lat)}</span>
											</div>
										{/if}

										{#if obj.lon !== undefined}
											<div class="flex flex-col gap-1">
												<span class="text-gray-600 font-medium">📍 Longitude</span>
												<span class="text-gray-900 font-mono text-xs">{formatLatLng(obj.lon)}</span>
											</div>
										{/if}

										{#if obj.alt !== undefined}
											<div class="flex flex-col gap-1">
												<span class="text-gray-600 font-medium">✈️ ความสูง</span>
												<span class="text-gray-900">{formatNumber(obj.alt, 2)} m</span>
											</div>
										{/if}

										{#if obj.timestamp !== undefined}
											<div class="flex flex-col gap-1">
												<span class="text-gray-600 font-medium">🕐 เวลาตรวจจับ</span>
												<span class="text-gray-900 text-xs">{formatTimestamp(obj.timestamp)}</span>
											</div>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					</div>
				{:else}
					<div class="bg-gray-100 p-6 rounded-lg text-center text-gray-500">
						ไม่พบข้อมูลโดรน
					</div>
				{/if}
			</div>

			<!-- Footer -->
			<div class="bg-gray-50 px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
				<button 
					type="button"
					onclick={closeModal} 
					class="px-4 py-2 bg-gray-300 text-gray-800 font-semibold rounded-lg hover:bg-gray-400 transition">
					ปิด
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(body) {
		overflow: hidden;
	}
</style>
