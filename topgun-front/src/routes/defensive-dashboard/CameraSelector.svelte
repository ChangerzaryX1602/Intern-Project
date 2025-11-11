<script lang="ts">
	import type { Camera, Pagination } from './types';

	interface Props {
		cameras: Camera[];
		selectedCameraIds: Set<string>;
		searchName: string;
		isLoading: boolean;
		pagination: Pagination;
		wsConnections: Map<string, WebSocket>;
		onToggleCamera: (cameraId: string) => void;
		onSearch: () => void;
		onSearchChange: (value: string) => void;
		onNextPage: () => void;
		onPrevPage: () => void;
	}

	let {
		cameras,
		selectedCameraIds,
		searchName = $bindable(),
		isLoading,
		pagination,
		wsConnections,
		onToggleCamera,
		onSearch,
		onSearchChange,
		onNextPage,
		onPrevPage
	}: Props = $props();
</script>

<aside class="sidebar">
	<!-- Camera Search -->
	<div class="sidebar-search">
		<label for="search-camera">🔍 Search Camera</label>
		<div class="search-box">
			<input
				id="search-camera"
				type="text"
				bind:value={searchName}
				placeholder="Search by name, location..."
				class="search-input"
				onkeydown={(e) => e.key === 'Enter' && onSearch()}
			/>
			<button class="btn-search" onclick={onSearch}>Search</button>
		</div>
	</div>

	<!-- Camera List -->
	<div class="camera-list">
		<div class="camera-list-header">
			<h2 class="section-title">
				<span class="icon">📹</span>
				Select Cameras
				{#if selectedCameraIds.size > 0}
					<span class="badge">{selectedCameraIds.size} selected</span>
				{/if}
			</h2>
		</div>

		{#if isLoading}
			<div class="loading-state">
				<div class="spinner"></div>
				<p>Loading cameras...</p>
			</div>
		{:else if cameras.length === 0}
			<div class="empty-state">
				<div class="empty-icon">📹</div>
				<p>No cameras found</p>
				<small>Try adjusting your search</small>
			</div>
		{:else}
			<div class="cameras-scroll">
				{#each cameras as camera (camera.id)}
					<button
						class="camera-card"
						class:selected={selectedCameraIds.has(camera.id)}
						onclick={() => onToggleCamera(camera.id)}
					>
						<div class="camera-checkbox">
							{#if selectedCameraIds.has(camera.id)}
								<span class="check-icon">✓</span>
							{:else}
								<span class="uncheck-icon">○</span>
							{/if}
						</div>
						<div class="camera-info">
							<div class="camera-name">{camera.name}</div>
							<div class="camera-location">
								<span class="icon-small">📍</span>
								{camera.location}
							</div>
							<div class="camera-institute">
								<span class="icon-small">🏛️</span>
								{camera.institute}
							</div>
							{#if wsConnections.has(camera.id)}
								<div class="camera-status connected">
									<span class="status-dot"></span>
									Connected
								</div>
							{/if}
						</div>
					</button>
				{/each}
			</div>

			<!-- Pagination -->
			{#if pagination.totalPages > 1}
				<div class="pagination">
					<button class="btn-page" onclick={onPrevPage} disabled={pagination.page === 1}>
						← Prev
					</button>
					<span class="page-info"> Page {pagination.page} of {pagination.totalPages} </span>
					<button
						class="btn-page"
						onclick={onNextPage}
						disabled={pagination.page >= pagination.totalPages}
					>
						Next →
					</button>
				</div>
			{/if}
		{/if}
	</div>
</aside>

<style>
	.sidebar {
		width: 30%;
		background: white;
		border-radius: 12px;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.sidebar-search {
		padding: 1.25rem;
		border-bottom: 1px solid #e5e7eb;
		background: linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%);
	}

	.sidebar-search label {
		display: block;
		font-size: 0.9rem;
		font-weight: 600;
		color: #374151;
		margin-bottom: 0.5rem;
	}

	.search-box {
		display: flex;
		gap: 0.5rem;
	}

	.search-input {
		flex: 1;
		padding: 0.75rem 1rem;
		border: 2px solid #e5e7eb;
		border-radius: 8px;
		font-size: 0.9rem;
		transition: all 0.2s;
		background: white;
	}

	.search-input:focus {
		outline: none;
		border-color: #667eea;
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
	}

	.btn-search {
		padding: 0.75rem 1.25rem;
		border: none;
		border-radius: 8px;
		background: #667eea;
		color: white;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
		flex-shrink: 0;
	}

	.btn-search:hover {
		background: #5568d3;
	}

	.camera-list {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.camera-list-header {
		padding: 1rem 1.25rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.section-title {
		margin: 0;
		font-size: 1.1rem;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: #1f2937;
		font-weight: 700;
	}

	.badge {
		background: #667eea;
		color: white;
		padding: 0.125rem 0.5rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
	}

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		color: #6b7280;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid #e5e7eb;
		border-top-color: #667eea;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		text-align: center;
		color: #9ca3af;
	}

	.empty-icon {
		font-size: 3rem;
		margin-bottom: 1rem;
		opacity: 0.5;
	}

	.empty-state p {
		margin: 0.5rem 0;
		font-weight: 500;
		color: #6b7280;
	}

	.empty-state small {
		font-size: 0.85rem;
	}

	.cameras-scroll {
		flex: 1;
		overflow-y: auto;
		padding: 0.75rem;
	}

	.cameras-scroll::-webkit-scrollbar {
		width: 6px;
	}

	.cameras-scroll::-webkit-scrollbar-track {
		background: #f3f4f6;
	}

	.cameras-scroll::-webkit-scrollbar-thumb {
		background: #d1d5db;
		border-radius: 3px;
	}

	.cameras-scroll::-webkit-scrollbar-thumb:hover {
		background: #9ca3af;
	}

	.camera-card {
		display: flex;
		gap: 0.75rem;
		padding: 1rem;
		border-radius: 8px;
		border: 2px solid #e5e7eb;
		background: white;
		margin-bottom: 0.75rem;
		cursor: pointer;
		transition: all 0.2s;
		text-align: left;
		width: 100%;
	}

	.camera-card:hover {
		background: #f9fafb;
		border-color: #d1d5db;
		transform: translateX(2px);
	}

	.camera-card.selected {
		background: linear-gradient(135deg, #ede9fe 0%, #ddd6fe 100%);
		border-color: #667eea;
		box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
	}

	.camera-checkbox {
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 4px;
		font-size: 1rem;
		font-weight: 700;
	}

	.check-icon {
		color: #667eea;
	}

	.uncheck-icon {
		color: #d1d5db;
	}

	.camera-info {
		flex: 1;
		min-width: 0;
	}

	.camera-name {
		font-size: 1rem;
		font-weight: 600;
		color: #1f2937;
		margin-bottom: 0.25rem;
	}

	.camera-location,
	.camera-institute {
		font-size: 0.85rem;
		color: #6b7280;
		margin-bottom: 0.125rem;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.icon-small {
		font-size: 0.75rem;
	}

	.camera-status {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
		font-weight: 600;
		margin-top: 0.5rem;
	}

	.camera-status.connected {
		color: #10b981;
	}

	.status-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #10b981;
		animation: pulse-dot 2s ease-in-out infinite;
	}

	@keyframes pulse-dot {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}

	.pagination {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 1.25rem;
		border-top: 1px solid #e5e7eb;
		background: #f9fafb;
	}

	.btn-page {
		padding: 0.5rem 1rem;
		border: 1px solid #e5e7eb;
		border-radius: 6px;
		background: white;
		color: #374151;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn-page:hover:not(:disabled) {
		background: #f3f4f6;
		border-color: #d1d5db;
	}

	.btn-page:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.page-info {
		font-size: 0.875rem;
		color: #6b7280;
		font-weight: 500;
	}

	.icon {
		display: inline-block;
	}
</style>
