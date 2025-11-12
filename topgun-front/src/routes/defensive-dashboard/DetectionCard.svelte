<script lang="ts">
	import type { Detection } from './types';

	interface Props {
		detection: Detection;
		isSelected: boolean;
		cameraName: string;
		onClick: () => void;
	}

	let { detection, isSelected, cameraName, onClick }: Props = $props();

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
</script>

<button class="flex gap-3 p-3.5 rounded-xl border-2 cursor-pointer transition-all duration-200 min-w-[280px] shrink-0 text-left hover:bg-gray-50 hover:border-gray-300 hover:-translate-y-0.5 hover:shadow-[0_4px_12px_rgba(0,0,0,0.08)]" 
	class:detection-selected={isSelected}
	class:border-gray-200={!isSelected}
	class:bg-white={!isSelected}
	onclick={onClick}>
	<div class="relative w-20 h-20 shrink-0 rounded-lg overflow-hidden bg-gray-100">
		{#if detection.image_base64}
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
		{#if detection.detected_objects && detection.detected_objects.length > 0}
			<div class="absolute top-1 right-1 bg-red-500/95 text-white px-2 py-0.5 rounded-xl text-xs font-semibold flex items-center gap-1 shadow-[0_2px_4px_rgba(0,0,0,0.2)]">
				<span class="text-[0.7rem]">🎯</span>
				{detection.detected_objects.length}
			</div>
		{/if}
	</div>
	<div class="flex-1 min-w-0 flex flex-col justify-center gap-1">
		<div class="text-sm font-semibold text-gray-800 flex items-center gap-1.5 whitespace-nowrap overflow-hidden text-ellipsis">
			<span class="text-[0.85rem] shrink-0">📹</span>
			{cameraName}
		</div>
		<div class="text-[0.8rem] text-gray-600">{formatDate(detection.detected_at)}</div>
		{#if detection.detected_objects && detection.detected_objects.length > 0}
			<div class="flex flex-wrap gap-1.5 mt-1">
				{#each detection.detected_objects.slice(0, 3) as obj}
					<span class="inline-block px-2 py-0.5 bg-gray-100 text-gray-700 rounded text-[0.7rem] font-medium whitespace-nowrap">{obj.class_name}</span>
				{/each}
				{#if detection.detected_objects.length > 3}
					<span class="inline-block px-2 py-0.5 bg-indigo-500 text-white rounded text-[0.7rem] font-medium whitespace-nowrap">+{detection.detected_objects.length - 3}</span>
				{/if}
			</div>
		{/if}
	</div>
</button>

<style>
	.detection-selected {
		background: linear-gradient(135deg, #faf5ff 0%, #e9d5ff 100%);
		border-color: #6366f1 !important;
		box-shadow: 0 4px 16px rgba(99, 102, 241, 0.25);
	}
</style>
