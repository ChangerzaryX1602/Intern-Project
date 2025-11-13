<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import mapboxgl from 'mapbox-gl';
	import 'mapbox-gl/dist/mapbox-gl.css';

	type MarkerConfig = {
		lngLat: [number, number];
		popup?: string;
		color?: string;
		id?: string; // For grouping markers with same ID
		kind?: 'start' | 'latest'; // start = small colored marker, latest = drone sticker
		icon?: string; // optional icon hint, e.g. 'drone'
	};

	type PathLine = {
		id: string;
		coordinates: [number, number][];
		color: string;
	};

	type Props = {
		accessToken: string;
		center?: [number, number];
		zoom?: number;
		style?: string;
		class?: string;
		markers?: MarkerConfig[];
		pathLines?: PathLine[]; // Full path coordinates for drawing lines
	};

	let {
		accessToken,
		center = [100.5018, 13.7563], 
		zoom = 12,
		style = 'mapbox://styles/mapbox/satellite-streets-v12',
		class: className = '',
		markers = [],
		pathLines = []
	}: Props = $props();

	let mapContainer: HTMLDivElement;
	let map: mapboxgl.Map | null = null;
	let mapboxMarkers: mapboxgl.Marker[] = [];
	let mapLoaded = $state(false);
	const LINE_SOURCE_ID = 'drone-paths';
	const LINE_LAYER_ID = 'drone-paths-layer';

	// Create arrow icon for direction indication
	function createArrowIcon(color: string = '#10b981') {
		const size = 70; // Reduced from 80 to ~27 (1/3 of original)
		const canvas = document.createElement('canvas');
		canvas.width = size;
		canvas.height = size;
		const ctx = canvas.getContext('2d')!;

		// Draw arrow pointing right (will be rotated by map)
		ctx.fillStyle = color;
		ctx.strokeStyle = color;
		ctx.lineWidth = 1; // Thinner border (was 3)
		
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
		if (!map || !mapLoaded || !pathLines || pathLines.length === 0) return;

		console.log('Drawing path lines:', pathLines.length);

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

		// Create line features from pathLines
		const lineFeatures: any[] = pathLines.map((path) => {
			const colorHex = path.color.replace('#', '');
			const iconName = `arrow-icon-${colorHex}`;

			return {
				type: 'Feature',
				properties: {
					id: path.id,
					color: path.color,
					iconName: iconName
				},
				geometry: {
					type: 'LineString',
					coordinates: path.coordinates
				}
			};
		});

		console.log(`Created ${lineFeatures.length} line features with full paths`);

		// Add lines to map
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
				'symbol-spacing': 50, // Adjusted spacing for smaller arrows
				'icon-image': ['get', 'iconName'],
				'icon-size': 0.8, // Reduced from 1.2 to 0.8 (smaller arrows)
				'icon-rotation-alignment': 'map',
				'icon-allow-overlap': true,
				'icon-ignore-placement': true
			},
			paint: {
				'icon-opacity': 0.9
			}
		});

		console.log(`Added ${lineFeatures.length} path lines with directional arrows to map`);
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
			let mapboxMarker: mapboxgl.Marker;

			// Latest marker -> render as drone sticker (GIF/SVG) using DOM element
			if (marker.kind === 'latest' || marker.icon === 'drone') {
				const color = marker.color || '#3b82f6';
				const el = document.createElement('div');
				el.className = 'drone-sticker';
				el.style.width = '48px';
				el.style.height = '48px';
				el.style.display = 'flex';
				el.style.alignItems = 'center';
				el.style.justifyContent = 'center';
				el.style.pointerEvents = 'auto';
				
				// Use drone.gif image
				const img = document.createElement('img');
				img.src = '/drone.gif';
				img.alt = 'Drone';
				img.style.width = '100%';
				img.style.height = '100%';
				img.style.objectFit = 'contain';
				
				el.appendChild(img);

				mapboxMarker = new mapboxgl.Marker({ element: el, anchor: 'center' })
					.setLngLat(marker.lngLat);

				if (marker.popup) {
					mapboxMarker.setPopup(new mapboxgl.Popup({ anchor: 'top' }).setHTML(marker.popup));
				}
			} else {
				// Default small colored marker for start points
				mapboxMarker = new mapboxgl.Marker({ color: marker.color || '#FF0000' })
					.setLngLat(marker.lngLat);
				if (marker.popup) {
					mapboxMarker.setPopup(new mapboxgl.Popup({ anchor: 'left' }).setHTML(marker.popup));
				}
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
					zoom: 18, // Increased from 15 to 18 for closer zoom
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
					maxZoom: 18, // Increased from 17 to 18 for closer zoom
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

	// Watch for pathLines changes and redraw
	$effect(() => {
		console.log('$effect triggered for pathLines:', pathLines?.length || 0, 'mapLoaded:', mapLoaded);
		if (pathLines && mapLoaded) {
			drawPathLines();
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
			zoom,
			attributionControl: false
		});

		// Add navigation controls
		map.addControl(new mapboxgl.NavigationControl(), 'top-right');

		// Wait for map to load before adding markers
		map.on('load', () => {
			// Create arrow icons for all possible colors
			const colors = [
				{ name: 'ef4444', value: '#ef4444' }, // red
				{ name: 'f59e0b', value: '#f59e0b' }, // orange
				{ name: 'eab308', value: '#eab308' }, // yellow
				{ name: '84cc16', value: '#84cc16' }, // lime
				{ name: '10b981', value: '#10b981' }, // green
				{ name: '14b8a6', value: '#14b8a6' }, // teal
				{ name: '06b6d4', value: '#06b6d4' }, // cyan
				{ name: '3b82f6', value: '#3b82f6' }, // blue
				{ name: '6366f1', value: '#6366f1' }, // indigo
				{ name: '8b5cf6', value: '#8b5cf6' }, // violet
				{ name: 'a855f7', value: '#a855f7' }, // purple
				{ name: 'ec4899', value: '#ec4899' }, // pink
				{ name: 'f43f5e', value: '#f43f5e' }  // rose
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

	/* Drone sticker marker styling */
	.drone-sticker {
		background: rgba(255,255,255,0.9);
		border-radius: 50%;
		padding: 4px;
		box-shadow: 0 2px 8px rgba(0,0,0,0.3);
	}
</style>
