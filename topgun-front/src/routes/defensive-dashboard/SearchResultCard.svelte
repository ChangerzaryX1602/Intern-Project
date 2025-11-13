<script lang="ts">
	import { env } from '$env/dynamic/public';

	interface Props {
		detection: any;
		cameraName: string;
		onClick: () => void;
	}

	let { detection, cameraName, onClick }: Props = $props();
	let showImageModal = $state(false);

	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080/api/v1';

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		const hours = String(date.getHours()).padStart(2, '0');
		const minutes = String(date.getMinutes()).padStart(2, '0');
		const seconds = String(date.getSeconds()).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const year = date.getFullYear();
		return `${hours}:${minutes}:${seconds} ${day}/${month}/${year}`;
	}

	function formatTimestamp(timestamp: number): string {
		const date = new Date(timestamp * 1000);
		return date.toLocaleString('th-TH');
	}

	function handleCardClick() {
		showImageModal = true;
		onClick();
	}

	const imageUrl = `${apiUrl}/detect/${detection.id}/file`;
	const objects = detection.objects || [];
</script>

<button
	class="flex flex-col gap-3 p-4 rounded-xl border-2 border-gray-200 bg-white cursor-pointer transition-all duration-200 hover:shadow-lg hover:-translate-y-1 text-left"
	onclick={handleCardClick}
>
	<!-- Header -->
	<div class="flex items-start justify-between gap-2">
		<div class="flex-1 min-w-0">
			<h3 class="font-bold text-gray-800 text-sm truncate flex items-center gap-2">
				<span>📹</span>
				{cameraName}
			</h3>
			<p class="text-xs text-gray-500 flex items-center gap-1 mt-1">
				<span>🕐</span>
				{formatDate(detection.timestamp)}
			</p>
		</div>
		{#if objects.length > 0}
			<span class="bg-red-100 text-red-700 px-2 py-1 rounded-full text-xs font-bold shrink-0">
				🚁 {objects.length}
			</span>
		{/if}
	</div>

	<!-- Image Preview -->
	<div class="relative w-full h-32 rounded-lg overflow-hidden bg-gray-100">
		<img src={imageUrl} alt="Detection {detection.id}" class="w-full h-full object-cover" />
	</div>

	<!-- Objects Details -->
	{#if objects.length > 0}
		<div class="space-y-2">
			{#each objects as obj, idx}
				<div class="bg-indigo-50 border-l-4 border-indigo-500 p-2 rounded text-xs">
					<div class="font-bold text-gray-800 mb-1">
						🆔 Track ID: {obj.track_id ?? idx + 1}
					</div>
					<div class="grid grid-cols-2 gap-x-3 gap-y-1 text-gray-700">
						<div>
							<span class="text-gray-500">📍 Lat:</span>
							<span class="font-mono ml-1">{obj.lat?.toFixed(6) ?? '-'}</span>
						</div>
						<div>
							<span class="text-gray-500">📍 Lon:</span>
							<span class="font-mono ml-1">{obj.lon?.toFixed(6) ?? '-'}</span>
						</div>
						<div>
							<span class="text-gray-500">✈️ Alt:</span>
							<span class="ml-1">{obj.alt?.toFixed(2) ?? '-'} m</span>
						</div>
						<div class="col-span-2">
							<span class="text-gray-500">🕐 Time:</span>
							<span class="ml-1">{obj.timestamp ? formatTimestamp(obj.timestamp) : '-'}</span>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</button>

<!-- Image Zoom Modal -->
{#if showImageModal}
	<div 
		class="fixed inset-0 z-[100] flex items-center justify-center bg-black bg-opacity-90 p-8"
		onclick={() => showImageModal = false}
	>
		<div 
			class="relative max-w-7xl max-h-full flex flex-col"
			onclick={(e) => e.stopPropagation()}
		>
			<!-- Close Button -->
			<button
				onclick={() => showImageModal = false}
				class="absolute -top-12 right-0 text-white hover:text-gray-300 transition-colors"
			>
				<svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>

			<!-- Header Info -->
			<div class="bg-gradient-to-r from-indigo-600 to-blue-600 text-white px-6 py-4 rounded-t-xl">
				<h2 class="text-xl font-bold">📹 {cameraName}</h2>
				<p class="text-sm opacity-90 mt-1">🕐 {formatDate(detection.timestamp)}</p>
			</div>

			<!-- Image -->
			<div class="bg-white p-4 max-h-[70vh] overflow-auto">
				<img 
					src={imageUrl} 
					alt="Detection {detection.id}"
					class="w-full h-auto rounded-lg"
				/>
			</div>

			<!-- Objects Details -->
			{#if objects.length > 0}
				<div class="bg-white px-6 py-4 rounded-b-xl border-t border-gray-200">
					<h3 class="text-lg font-bold text-gray-900 mb-3 flex items-center gap-2">
						<span>🚁</span>
						ข้อมูลโดรนที่ตรวจพบ ({objects.length} ตำแหน่ง)
					</h3>
					<div class="grid grid-cols-2 gap-3 max-h-48 overflow-y-auto">
						{#each objects as obj, idx}
							<div class="bg-indigo-50 border-l-4 border-indigo-500 p-3 rounded text-sm">
								<div class="font-bold text-gray-800 mb-2">
									🆔 Track ID: {obj.track_id ?? idx + 1}
								</div>
								<div class="space-y-1 text-gray-700">
									<div class="flex justify-between">
										<span class="text-gray-600">📍 Latitude:</span>
										<span class="font-mono">{obj.lat?.toFixed(6) ?? '-'}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">📍 Longitude:</span>
										<span class="font-mono">{obj.lon?.toFixed(6) ?? '-'}</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">✈️ ความสูง:</span>
										<span>{obj.alt?.toFixed(2) ?? '-'} m</span>
									</div>
									<div class="flex justify-between">
										<span class="text-gray-600">🕐 เวลาตรวจจับ:</span>
										<span class="text-xs">{obj.timestamp ? formatTimestamp(obj.timestamp) : '-'}</span>
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}
