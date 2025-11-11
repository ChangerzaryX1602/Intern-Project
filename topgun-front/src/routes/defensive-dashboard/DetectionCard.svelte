<script lang="ts">
	import type { Detection } from './types';

	interface Props {
		detection: Detection;
		isSelected: boolean;
		cameraName: string;
		onClick: () => void;
	}

	let { detection, isSelected, cameraName, onClick }: Props = $props();

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		const hours = String(date.getHours()).padStart(2, '0');
		const minutes = String(date.getMinutes()).padStart(2, '0');
		const seconds = String(date.getSeconds()).padStart(2, '0');
		const day = String(date.getDate()).padStart(2, '0');
		const month = String(date.getMonth() + 1).padStart(2, '0');
		const year = date.getFullYear();
		return `${hours}:${minutes}:${seconds} ${day}/${month}/${year}`;
	}
</script>

<button class="detection-card" class:selected={isSelected} onclick={onClick}>
	<div class="detection-thumbnail">
		{#if detection.image_base64}
			<img
				src="data:image/jpeg;base64,{detection.image_base64}"
				alt="Detection {detection.id}"
				class="thumbnail-img"
			/>
		{:else}
			<div class="thumbnail-placeholder">
				<span>📷</span>
			</div>
		{/if}
		{#if detection.detected_objects && detection.detected_objects.length > 0}
			<div class="object-badge">
				<span class="badge-icon">🎯</span>
				{detection.detected_objects.length}
			</div>
		{/if}
	</div>
	<div class="detection-info">
		<div class="detection-camera">
			<span class="camera-icon">📹</span>
			{cameraName}
		</div>
		<div class="detection-time">{formatDate(detection.detected_at)}</div>
		{#if detection.detected_objects && detection.detected_objects.length > 0}
			<div class="detection-objects">
				{#each detection.detected_objects.slice(0, 3) as obj}
					<span class="object-tag">{obj.class_name}</span>
				{/each}
				{#if detection.detected_objects.length > 3}
					<span class="object-tag more">+{detection.detected_objects.length - 3}</span>
				{/if}
			</div>
		{/if}
	</div>
</button>

<style>
	.detection-card {
		display: flex;
		gap: 0.75rem;
		padding: 0.875rem;
		border-radius: 10px;
		border: 2px solid #e5e7eb;
		background: white;
		cursor: pointer;
		transition: all 0.2s;
		min-width: 280px;
		flex-shrink: 0;
		text-align: left;
	}

	.detection-card:hover {
		background: #f9fafb;
		border-color: #d1d5db;
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
	}

	.detection-card.selected {
		background: linear-gradient(135deg, #ede9fe 0%, #ddd6fe 100%);
		border-color: #667eea;
		box-shadow: 0 4px 16px rgba(102, 126, 234, 0.25);
	}

	.detection-thumbnail {
		position: relative;
		width: 80px;
		height: 80px;
		flex-shrink: 0;
		border-radius: 8px;
		overflow: hidden;
		background: #f3f4f6;
	}

	.thumbnail-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.thumbnail-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		color: #d1d5db;
	}

	.object-badge {
		position: absolute;
		top: 4px;
		right: 4px;
		background: rgba(239, 68, 68, 0.95);
		color: white;
		padding: 0.125rem 0.5rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 0.25rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
	}

	.badge-icon {
		font-size: 0.7rem;
	}

	.detection-info {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: 0.25rem;
	}

	.detection-camera {
		font-size: 0.9rem;
		font-weight: 600;
		color: #1f2937;
		display: flex;
		align-items: center;
		gap: 0.375rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.camera-icon {
		font-size: 0.85rem;
		flex-shrink: 0;
	}

	.detection-time {
		font-size: 0.8rem;
		color: #6b7280;
	}

	.detection-objects {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem;
		margin-top: 0.25rem;
	}

	.object-tag {
		display: inline-block;
		padding: 0.125rem 0.5rem;
		background: #f3f4f6;
		color: #374151;
		border-radius: 4px;
		font-size: 0.7rem;
		font-weight: 500;
		white-space: nowrap;
	}

	.object-tag.more {
		background: #667eea;
		color: white;
	}
</style>
