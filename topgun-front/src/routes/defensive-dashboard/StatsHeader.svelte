<script lang="ts">
	import { goto } from "$app/navigation";

	interface Props {
		selectedCamerasCount: number;
		detectionsCount: number;
		activeConnectionsCount: number;
		onDisconnectAll: () => void;
	}

	let { selectedCamerasCount, detectionsCount, activeConnectionsCount, onDisconnectAll }: Props =
		$props();

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
</script>

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
				<h1 class="m-0 text-xl font-bold">Defensive Dashboard</h1>
				<p class="m-0 opacity-90 text-xs">Real-time Detection Monitoring</p>
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
			<div class="flex items-center gap-2 px-4 py-3 bg-white/20 rounded-lg text-sm font-medium">
				<span class="text-lg">👨‍💼</span>
				<span>commander</span>
			</div>
		</div>
	</div>
</header>

<style>
</style>
