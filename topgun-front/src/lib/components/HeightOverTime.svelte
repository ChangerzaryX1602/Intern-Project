<script lang="ts">
	import { onMount } from 'svelte';
	import Chart from 'chart.js/auto';
	import { env } from '$env/dynamic/public';

	const apiUrl = env.PUBLIC_API_URL || 'http://localhost:8080/api/v1';
	let chartCanvas: HTMLCanvasElement;
	let chart: Chart | null = null;
	let isLoading = true;

	// Store height data points
	let heightData: { timestamp: string; height: number }[] = [];

	// Fetch data from the API
	async function fetchHeightData() {
		try {
			const res = await fetch(`${apiUrl}/detections`);
			if (!res.ok) throw new Error('Failed to fetch detection data');
			const detections = await res.json();

			// Example structure: { latitude, longitude, height, timestamp }
			heightData = detections.map((d: any) => ({
				timestamp: new Date(d.timestamp * 1000).toLocaleTimeString('en-US', {
					hour: '2-digit',
					minute: '2-digit',
					second: '2-digit'
				}),
				height: parseFloat(d.height) || 0
			}));

			updateChart();
		} catch (err) {
			console.error('Error fetching data:', err);
		} finally {
			isLoading = false;
		}
	}

	function updateChart() {
		if (!chart) return;
		chart.data.labels = heightData.map(d => d.timestamp);
		chart.data.datasets[0].data = heightData.map(d => d.height);
		chart.update();
	}

	onMount(() => {
		const ctx = chartCanvas.getContext('2d');
		if (!ctx) return; // Type-safe null check

		chart = new Chart(ctx, {
			type: 'line',
			data: {
				labels: [],
				datasets: [
					{
						label: 'Height (meters)',
						data: [],
						borderColor: '#6366f1',
						backgroundColor: 'rgba(99,102,241,0.2)',
						fill: true,
						tension: 0.3,
						pointRadius: 4,
						pointBackgroundColor: '#4f46e5'
					}
				]
			},
			options: {
				responsive: true,
				plugins: {
					title: {
						display: true,
						text: 'Height Variation Over Time',
						font: { size: 16, weight: 'bold' }
					}
				},
				scales: {
					x: {
						title: { display: true, text: 'Timestamp' },
						grid: { color: '#eee' }
					},
					y: {
						title: { display: true, text: 'Height (m)' },
						grid: { color: '#eee' }
					}
				}
			}
		});

		// Initial fetch
		fetchHeightData();

		// Refresh data every 5 seconds
		const interval = setInterval(fetchHeightData, 5000);
		return () => clearInterval(interval);
	});
</script>

<div class="p-4 bg-white rounded-2xl shadow-md">
	{#if isLoading}
		<div class="text-center text-gray-500 py-8">Loading height data...</div>
	{:else}
		<canvas bind:this={chartCanvas}></canvas>
	{/if}
</div>

<style>
	div {
		width: 100%;
		height: auto;
		margin: auto;
	}
</style>