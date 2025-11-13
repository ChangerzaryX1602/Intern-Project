import { untrack } from 'svelte';

export function useImageZoom() {
	let isOpen = $state(false);
	let imageSrc = $state('');

	function openZoom(src: string) {
		imageSrc = src;
		isOpen = true;
	}

	function closeZoom() {
		isOpen = false;
		// Clear imageSrc after animation
		setTimeout(() => {
			untrack(() => {
				imageSrc = '';
			});
		}, 200);
	}

	function setupImageClickHandlers(containerElement: HTMLElement) {
		const images = containerElement.querySelectorAll('img');
		
		images.forEach((img) => {
			img.style.cursor = 'pointer';
			img.addEventListener('click', () => {
				if (img.src) {
					openZoom(img.src);
				}
			});
		});

		// Return cleanup function
		return () => {
			images.forEach((img) => {
				img.style.cursor = '';
				// Remove event listeners by cloning (removes all listeners)
				const newImg = img.cloneNode(true);
				img.parentNode?.replaceChild(newImg, img);
			});
		};
	}

	return {
		get isOpen() {
			return isOpen;
		},
		get imageSrc() {
			return imageSrc;
		},
		openZoom,
		closeZoom,
		setupImageClickHandlers
	};
}
