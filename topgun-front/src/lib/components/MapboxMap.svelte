<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import mapboxgl from 'mapbox-gl';
	import 'mapbox-gl/dist/mapbox-gl.css';

	type MarkerConfig = {
		lngLat: [number, number];
		popup?: string;
		color?: string;
	};

	type Props = {
		accessToken: string;
		center?: [number, number];
		zoom?: number;
		style?: string;
		class?: string;
		markers?: MarkerConfig[];
	};

	let {
		accessToken,
		center = [100.5018, 13.7563], 
		zoom = 12,
		style = 'mapbox://styles/mapbox/streets-v12',
		class: className = '',
		markers = []
	}: Props = $props();

	let mapContainer: HTMLDivElement;
	let map: mapboxgl.Map | null = null;
	let mapboxMarkers: mapboxgl.Marker[] = [];
	let mapLoaded = $state(false);

	// Function to update markers
	function updateMarkers() {
		if (!map || !mapLoaded) {
			console.log('Cannot update markers: map not ready', { map: !!map, mapLoaded });
			return;
		}

		console.log('Updating markers on map:', markers.length, markers);

		// Remove old markers
		mapboxMarkers.forEach((marker) => marker.remove());
		mapboxMarkers = [];

		// Add new markers
		markers.forEach((marker, index) => {
			console.log(`Adding marker ${index}:`, marker);
			const mapboxMarker = new mapboxgl.Marker({ color: marker.color || '#FF0000' })
				.setLngLat(marker.lngLat);

			if (marker.popup) {
				mapboxMarker.setPopup(new mapboxgl.Popup().setHTML(marker.popup));
			}

			mapboxMarker.addTo(map!);
			mapboxMarkers.push(mapboxMarker);
		});

		console.log('Markers updated on map:', markers.length, 'Total mapbox markers:', mapboxMarkers.length);

		// Auto-fit map to show all markers
		if (markers.length > 0) {
			if (markers.length === 1) {
				// Single marker: center on it with fixed zoom
				map!.flyTo({
					center: markers[0].lngLat,
					zoom: 15,
					duration: 1000
				});
			} else {
				// Multiple markers: fit bounds to show all
				const bounds = new mapboxgl.LngLatBounds();
				markers.forEach((marker) => {
					bounds.extend(marker.lngLat);
				});
				map!.fitBounds(bounds, {
					padding: { top: 50, bottom: 50, left: 50, right: 50 },
					maxZoom: 15,
					duration: 1000
				});
			}
		}
	}

	// Watch for center changes and update map
	$effect(() => {
		if (map && center && mapLoaded) {
			map.setCenter(center);
		}
	});

	// Watch for zoom changes and update map
	$effect(() => {
		if (map && zoom && mapLoaded) {
			map.setZoom(zoom);
		}
	});

	// Watch for marker changes and update map
	$effect(() => {
		console.log('$effect triggered for markers:', markers.length, 'mapLoaded:', mapLoaded);
		if (markers && mapLoaded) {
			updateMarkers();
		}
	});

	onMount(() => {
		if (!accessToken) {
			console.error('Mapbox access token is required');
			return;
		}

		mapboxgl.accessToken = accessToken;

		map = new mapboxgl.Map({
			container: mapContainer,
			style,
			center,
			zoom
		});

		// Add navigation controls
		map.addControl(new mapboxgl.NavigationControl(), 'top-right');

		// Wait for map to load before adding markers
		map.on('load', () => {
			mapLoaded = true;
			console.log('Map loaded, adding initial markers');
			updateMarkers();
		});
	});

	onDestroy(() => {
		// Remove all markers
		mapboxMarkers.forEach((marker) => marker.remove());
		mapboxMarkers = [];
		// Remove map
		map?.remove();
	});
</script>

<div bind:this={mapContainer} class="mapbox-container {className}"></div>

<style>
	.mapbox-container {
		width: 100%;
		height: 100%;
	}
</style>
