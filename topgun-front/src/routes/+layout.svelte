<script lang="ts">
	import '../app.css';
	import favicon from '$lib/assets/favicon.svg';
	import { ImageZoomModal, useImageZoom } from '$lib';
	import { onMount } from 'svelte';
	
	let { children } = $props();

	const imageZoom = useImageZoom();

	onMount(() => {
		// Setup global image click handlers
		const cleanup = imageZoom.setupImageClickHandlers(document.body);
		
		// Re-setup on navigation or dynamic content changes
		const observer = new MutationObserver(() => {
			imageZoom.setupImageClickHandlers(document.body);
		});
		
		observer.observe(document.body, {
			childList: true,
			subtree: true
		});

		return () => {
			cleanup();
			observer.disconnect();
		};
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{@render children()}

<ImageZoomModal 
	bind:isOpen={imageZoom.isOpen} 
	bind:imageSrc={imageZoom.imageSrc}
	onClose={imageZoom.closeZoom}
/>