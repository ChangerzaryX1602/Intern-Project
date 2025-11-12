<script lang="ts">
	interface Props {
		value: string;
		placeholder?: string;
		label?: string;
		icon?: string;
		onSearch: () => void;
		onInput?: () => void;
	}

	let {
		value = $bindable(),
		placeholder = 'Search...',
		label = '🔍 Search',
		icon = '🔍',
		onSearch,
		onInput
	}: Props = $props();
</script>

<div class="px-5 py-5 border-b border-gray-200 bg-gradient-to-br from-gray-50 to-gray-100">
	<label for="search-input" class="block text-sm font-semibold text-gray-700 mb-2">
		{label}
	</label>
	<div class="flex gap-2">
		<input
			id="search-input"
			type="text"
			bind:value
			{placeholder}
			class="flex-1 px-4 py-3 border-2 border-gray-200 rounded-lg text-sm transition-all duration-200 bg-white focus:outline-none focus:border-indigo-500 focus:shadow-[0_0_0_3px_rgba(102,126,234,0.1)]"
			oninput={onInput}
			onkeydown={(e) => e.key === 'Enter' && onSearch()}
		/>
		<!-- <button
			class="px-5 py-3 border-none rounded-lg bg-indigo-500 text-white font-semibold cursor-pointer transition-all duration-200 shrink-0 hover:bg-indigo-600"
			onclick={onSearch}
		>
			Search
		</button> -->
	</div>
</div>
