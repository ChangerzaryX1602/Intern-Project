<script lang="ts">
	import SearchBox from '$lib/components/SearchBox.svelte';
	import type { Camera, Pagination } from './types';

	interface Props {
		cameras: Camera[];
		selectedCameraIds: Set<string>;
		searchName: string;
		isLoading: boolean;
		pagination: Pagination;
		wsConnections: Map<string, WebSocket>;
		onToggleCamera: (cameraId: string) => void;
		onSearch: () => void;
		onSearchChange: (value: string) => void;
		onSearchInput: () => void;
		onNextPage: () => void;
		onPrevPage: () => void;
	}

	let {
		cameras,
		selectedCameraIds,
		searchName = $bindable(),
		isLoading,
		pagination,
		wsConnections,
		onToggleCamera,
		onSearch,
		onSearchChange,
		onSearchInput,
		onNextPage,
		onPrevPage
	}: Props = $props();

	// Highlight matching text
	function highlightText(text: string, keyword: string): string {
		if (!keyword.trim()) return text;
		const regex = new RegExp(`(${keyword.trim()})`, 'gi');
		return text.replace(regex, '<mark>$1</mark>');
	}
</script>

<aside class="w-[30%] bg-white rounded-xl shadow-md flex flex-col overflow-hidden">
	<!-- Camera Search -->
	<SearchBox
		bind:value={searchName}
		placeholder="Search by name, location..."
		label="🔍 Search Camera"
		{onSearch}
		onInput={onSearchInput}
	/>

	<!-- Camera List -->
	<div class="flex-1 flex flex-col overflow-hidden">
		<div class="px-5 py-4 border-b border-gray-200">
			<h2 class="m-0 text-lg flex items-center gap-2 text-gray-800 font-bold">
				<span class="inline-block">📹</span>
				Select Cameras
				{#if selectedCameraIds.size > 0}
					<span class="bg-indigo-500 text-white px-2 py-0.5 rounded-xl text-xs font-semibold">
						{selectedCameraIds.size} selected
					</span>
				{/if}
			</h2>
		</div>

		{#if isLoading}
			<div class="flex flex-col items-center justify-center p-12 text-gray-600">
				<div class="w-10 h-10 border-4 border-gray-200 border-t-indigo-500 rounded-full animate-spin mb-4"></div>
				<p>Loading cameras...</p>
			</div>
		{:else if cameras.length === 0}
			<div class="flex flex-col items-center justify-center p-12 text-center text-gray-400">
				<div class="text-5xl mb-4 opacity-50">📹</div>
				<p class="my-2 font-medium text-gray-600">No cameras found</p>
				<small class="text-sm">Try adjusting your search</small>
			</div>
		{:else}
			<div class="flex-1 overflow-y-auto p-3 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 hover:scrollbar-thumb-gray-400">
				{#each cameras as camera (camera.id)}
					<button
						class="flex gap-3 p-4 rounded-lg border-2 mb-3 cursor-pointer transition-all duration-200 text-left w-full hover:bg-gray-50 hover:border-gray-300 hover:translate-x-0.5"
						class:camera-selected={selectedCameraIds.has(camera.id)}
						class:border-gray-200={!selectedCameraIds.has(camera.id)}
						class:bg-white={!selectedCameraIds.has(camera.id)}
						onclick={() => onToggleCamera(camera.id)}
					>
						<div class="flex-shrink-0 w-6 h-6 flex items-center justify-center rounded text-base font-bold">
							{#if selectedCameraIds.has(camera.id)}
								<span class="text-indigo-500">✓</span>
							{:else}
								<span class="text-gray-300">○</span>
							{/if}
						</div>
						<div class="flex-1 min-w-0">
							<div class="text-base font-semibold text-gray-800 mb-1">
								{@html highlightText(camera.name, searchName)}
							</div>
							<div class="text-sm text-gray-600 mb-0.5 flex items-center gap-1">
								<span class="text-xs">📍</span>
								{@html highlightText(camera.location, searchName)}
							</div>
							<div class="text-sm text-gray-600 mb-0.5 flex items-center gap-1">
								<span class="text-xs">🏛️</span>
								{@html highlightText(camera.institute, searchName)}
							</div>
							{#if wsConnections.has(camera.id)}
								<div class="flex items-center gap-1.5 text-xs font-semibold mt-2 text-green-500">
									<span class="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse-slow"></span>
									Connected
								</div>
							{/if}
						</div>
					</button>
				{/each}
			</div>

			<!-- Pagination -->
			{#if pagination.totalPages > 1}
				<div class="flex justify-between items-center px-5 py-4 border-t border-gray-200 bg-gray-50">
					<button
						class="px-4 py-2 border border-gray-200 rounded-md bg-white text-gray-700 font-medium cursor-pointer transition-all duration-200 hover:bg-gray-100 hover:border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
						onclick={onPrevPage}
						disabled={pagination.page === 1}
					>
						← Prev
					</button>
					<span class="text-sm text-gray-600 font-medium">
						Page {pagination.page} of {pagination.totalPages}
					</span>
					<button
						class="px-4 py-2 border border-gray-200 rounded-md bg-white text-gray-700 font-medium cursor-pointer transition-all duration-200 hover:bg-gray-100 hover:border-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
						onclick={onNextPage}
						disabled={pagination.page >= pagination.totalPages}
					>
						Next →
					</button>
				</div>
			{/if}
		{/if}
	</div>
</aside>

<style>
	.camera-selected {
		background: linear-gradient(135deg, #faf5ff 0%, #e9d5ff 100%);
		border-color: #6366f1 !important;
		box-shadow: 0 2px 8px rgba(99, 102, 241, 0.2);
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

	:global(mark) {
		background-color: #fef08a;
		color: #1f2937;
		padding: 0.125rem 0.25rem;
		border-radius: 0.25rem;
		font-weight: 600;
	}
</style>
