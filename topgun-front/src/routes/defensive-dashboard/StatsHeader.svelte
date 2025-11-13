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

	function toggleNotification(type: 'email' | 'line') {
		notificationSettings[type] = !notificationSettings[type];
	}

	function getNotificationCount(): number {
		return Object.values(notificationSettings).filter(Boolean).length;
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
			<!-- Notification Settings -->
			<div class="relative">
				<button
					onclick={() => (showNotificationMenu = !showNotificationMenu)}
					class="flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors cursor-pointer"
					title="Notification settings"
				>
					<span class="text-lg">🔔</span>
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

<style>
</style>
