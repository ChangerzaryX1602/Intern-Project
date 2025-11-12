<script lang="ts">
	interface Props {
		selectedCamerasCount: number;
		detectionsCount: number;
		activeConnectionsCount: number;
		onDisconnectAll: () => void;
	}

	let { selectedCamerasCount, detectionsCount, activeConnectionsCount, onDisconnectAll }: Props =
		$props();
</script>

<header class="bg-gradient-to-br from-indigo-500 to-purple-600 text-white px-8 py-6 shadow-lg z-10">
	<div class="flex justify-between items-center gap-8 max-w-full">
		<div class="flex items-center gap-4">
			<!-- <div class="text-5xl animate-bounce-slow">🛡️</div> -->
			<div>
				<h1 class="m-0 text-[1.75rem] font-bold">Defensive Dashboard</h1>
				<p class="m-0 opacity-90 text-sm">Real-time Detection Monitoring</p>
			</div>
		</div>

		<div class="flex items-center gap-4">
			<div class="flex gap-6 px-4 py-2 bg-white/20 rounded-lg">
				<div class="flex flex-col items-center gap-1">
					<span class="text-xs opacity-90 uppercase tracking-wide">Cameras</span>
					<span class="text-2xl font-bold">{selectedCamerasCount}</span>
				</div>
				<div class="flex flex-col items-center gap-1">
					<span class="text-xs opacity-90 uppercase tracking-wide">Detections</span>
					<span class="text-2xl font-bold">{detectionsCount}</span>
				</div>
			</div>

			{#if selectedCamerasCount > 0}
				<button
					class="px-6 py-2.5 rounded-lg border-none font-semibold cursor-pointer flex items-center gap-2 transition-all duration-200 text-sm bg-red-500 text-white hover:bg-red-600 hover:-translate-y-0.5 hover:shadow-md"
					onclick={onDisconnectAll}
				>
					<span class="inline-block">⏹</span>
					Disconnect All
				</button>
			{/if}

			<div
				class="flex items-center gap-2 px-4 py-2 bg-white/20 rounded-full text-sm font-medium"
				class:active={activeConnectionsCount > 0}
			>
				<span
					class="w-2 h-2 rounded-full transition-colors duration-300"
					class:bg-gray-400={activeConnectionsCount === 0}
					class:bg-green-400={activeConnectionsCount > 0}
					class:animate-pulse-slow={activeConnectionsCount > 0}
				></span>
				<span
					>{activeConnectionsCount > 0
						? `${activeConnectionsCount} Connected`
						: 'Disconnected'}</span
				>
			</div>
		</div>
	</div>
</header>

<style>
	@keyframes bounce-slow {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-10px);
		}
	}

	.animate-bounce-slow {
		animation: bounce-slow 3s ease-in-out infinite;
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
</style>
