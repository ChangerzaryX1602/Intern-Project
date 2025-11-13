<script lang="ts">
	interface Props {
		isOpen?: boolean;
		imageSrc?: string;
		onClose: () => void;
	}

	let { isOpen = $bindable(false), imageSrc = $bindable(''), onClose }: Props = $props();

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			onClose();
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			onClose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen && imageSrc}
	<div
		class="fixed inset-0 bg-black/80 flex items-center justify-center z-[999999] p-4"
		onclick={handleBackdropClick}
		role="dialog"
		aria-modal="true"
	>
		<div class="relative max-w-[95vw] max-h-[95vh]">
			<button
				onclick={onClose}
				class="absolute -top-12 right-0 text-white hover:text-gray-300 transition-colors text-4xl font-bold w-10 h-10 flex items-center justify-center"
				aria-label="Close"
			>
				×
			</button>
			<img
				src={imageSrc}
				alt="Zoomed view"
				class="max-w-full max-h-[90vh] object-contain rounded-lg"
			/>
		</div>
	</div>
{/if}
