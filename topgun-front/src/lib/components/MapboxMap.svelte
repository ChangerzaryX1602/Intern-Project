<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import mapboxgl from 'mapbox-gl';
	import 'mapbox-gl/dist/mapbox-gl.css';

	type MarkerConfig = {
		lngLat: [number, number];
		popup?: string;
		color?: string;
		id?: string; // For grouping markers with same ID
	};

	type Props = {
		accessToken: string;
		center?: [number, number];
		zoom?: number;
		style?: string;
		class?: string;
		markers?: MarkerConfig[];
		drawLines?: boolean; // Enable drawing lines between markers with same ID
	};

	let {
		accessToken,
		center = [100.5018, 13.7563], 
		zoom = 12,
		style = 'mapbox://styles/mapbox/satellite-streets-v12',
		class: className = '',
		markers = [],
		drawLines = false
	}: Props = $props();

	let mapContainer: HTMLDivElement;
	let map: mapboxgl.Map | null = null;
	let mapboxMarkers: mapboxgl.Marker[] = [];
	let mapLoaded = $state(false);
	const LINE_SOURCE_ID = 'drone-paths';
	const LINE_LAYER_ID = 'drone-paths-layer';

	// Create arrow icon for direction indication
	function createArrowIcon(color: string = '#10b981') {
		const size = 80; // Increased from 40 to 80 for bigger arrows
		const canvas = document.createElement('canvas');
		canvas.width = size;
		canvas.height = size;
		const ctx = canvas.getContext('2d')!;

		// Draw arrow pointing right (will be rotated by map)
		ctx.fillStyle = color;
		ctx.strokeStyle = color;
		ctx.lineWidth = 3; // Thicker border
		
		// Draw a triangular arrow
		ctx.beginPath();
		ctx.moveTo(size * 0.2, size * 0.3); // Top left
		ctx.lineTo(size * 0.7, size * 0.5); // Point (right)
		ctx.lineTo(size * 0.2, size * 0.7); // Bottom left
		ctx.closePath();
		ctx.fill();
		ctx.stroke();

		return {
			width: size,
			height: size,
			data: ctx.getImageData(0, 0, size, size).data
		};
	}

	// Function to draw lines between markers with same ID
	function drawPathLines() {
		if (!map || !mapLoaded || !drawLines) return;

		console.log('Drawing path lines for markers with same ID');

		// Remove existing arrow layer first
		if (map.getLayer(LINE_LAYER_ID + '-arrows')) {
			map.removeLayer(LINE_LAYER_ID + '-arrows');
		}
		// Remove existing line layer
		if (map.getLayer(LINE_LAYER_ID)) {
			map.removeLayer(LINE_LAYER_ID);
		}
		// Remove source
		if (map.getSource(LINE_SOURCE_ID)) {
			map.removeSource(LINE_SOURCE_ID);
		}

		// Group markers by ID
		const markerGroups = new Map<string, MarkerConfig[]>();
		markers.forEach((marker) => {
			if (marker.id) {
				if (!markerGroups.has(marker.id)) {
					markerGroups.set(marker.id, []);
				}
				markerGroups.get(marker.id)!.push(marker);
			}
		});

		console.log('Marker groups:', Array.from(markerGroups.entries()));

		// Create line features for groups with more than 1 marker
		const lineFeatures: any[] = [];
		markerGroups.forEach((group, id) => {
			if (group.length > 1) {
				// Sort by timestamp if available, or maintain order
				const coordinates = group.map((m) => m.lngLat);
				const color = group[0].color || '#3b82f6';
				
				// Determine icon name based on color
				let iconName = 'arrow-icon-blue';
				if (color === '#10b981') iconName = 'arrow-icon-green';
				else if (color === '#ef4444') iconName = 'arrow-icon-red';
				
				lineFeatures.push({
					type: 'Feature',
					properties: {
						id: id,
						color: color,
						iconName: iconName
					},
					geometry: {
						type: 'LineString',
						coordinates: coordinates
					}
				});
				
				console.log(`Created line for ID ${id} with ${coordinates.length} points, color: ${color}, icon: ${iconName}`);
			}
		});

		// Add lines to map if any exist
		if (lineFeatures.length > 0) {
			map.addSource(LINE_SOURCE_ID, {
				type: 'geojson',
				data: {
					type: 'FeatureCollection',
					features: lineFeatures
				}
			});

			// Add the line layer
			map.addLayer({
				id: LINE_LAYER_ID,
				type: 'line',
				source: LINE_SOURCE_ID,
				layout: {
					'line-join': 'round',
					'line-cap': 'round'
				},
				paint: {
					'line-color': ['get', 'color'],
					'line-width': 3,
					'line-opacity': 0.8
				}
			});

			// Add arrow symbols layer on top of the line
			map.addLayer({
				id: LINE_LAYER_ID + '-arrows',
				type: 'symbol',
				source: LINE_SOURCE_ID,
				layout: {
					'symbol-placement': 'line',
					'symbol-spacing': 120, // Increased spacing for bigger arrows
					'icon-image': ['get', 'iconName'], // Use icon based on line color
					'icon-size': 1.2, // Increased from 0.6 to 1.2 (doubled)
					'icon-rotation-alignment': 'map',
					'icon-allow-overlap': true,
					'icon-ignore-placement': true
				},
				paint: {
					'icon-opacity': 0.9
				}
			});

			console.log(`Added ${lineFeatures.length} path lines with directional arrows to map`);
		} else {
			console.log('No groups with multiple markers found');
		}
	}

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

		// Draw path lines if enabled
		drawPathLines();

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
			// Create arrow icons for different colors
			const colors = [
				{ name: 'green', value: '#10b981' },
				{ name: 'red', value: '#ef4444' },
				{ name: 'blue', value: '#3b82f6' }
			];

			colors.forEach(({ name, value }) => {
				const arrowIcon = createArrowIcon(value);
				if (map && !map.hasImage(`arrow-icon-${name}`)) {
					map.addImage(`arrow-icon-${name}`, arrowIcon, { pixelRatio: 2 });
				}
			});

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
