<script lang="ts">
	import { goto } from '$app/navigation';
</script>

<div class="container">
	<div class="header">
		<h1>TOPGUN - UAV Surveillance Dashboard</h1>
		<p>Unmanned aerial vehicle inspection system</p>
	</div>

	<div class="cards-grid">
		<!-- Defense Card -->
		<button
			class="card defense-card"
			onclick={() => goto('/defensive-dashboard')}
		>
			<div class="card-icon">🛡️</div>
			<h2>Defense</h2>
			<p>Real-time detection and monitoring</p>
			<p class="subtitle">Track detected objects from cameras</p>
			<div class="card-arrow">→</div>
		</button>

		<!-- Offense Card -->
		<button
			class="card offense-card"
			onclick={() => goto('/offensive-dashboard')}
		>
			<div class="card-icon">⚔️</div>
			<h2>Offense</h2>
			<p>Strategic operations tracking</p>
			<p class="subtitle">Manage offensive campaigns</p>
			<div class="card-arrow">→</div>
		</button>

		<!-- Battle Card -->
		<button
			class="card battle-card"
			onclick={() => goto('/battle')}
		>
			<div class="card-icon">⚡</div>
			<h2>Battle</h2>
			<p>Defense & Offense coordination</p>
			<p class="subtitle">Unified battle management</p>
			<div class="card-arrow">→</div>
		</button>
	</div>
    <p class="subtitle mt-20">by GearDinDaeng2025</p>
</div>

<style>
	:global(body) {
		margin: 0;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
		min-height: 100vh;
	}

	.container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 100vh;
		min-width: 170vh;
		padding: 2rem;
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	}

	.header {
		text-align: center;
		margin-bottom: 4rem;
		color: white;
	}

	.header h1 {
		margin: 0;
		font-size: 2.5rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
	}

	.header p {
		margin: 0;
		font-size: 1.4rem;
		opacity: 0.9;
		font-weight: 500;
	}

	.cards-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 2rem;
		max-width: 1200px;
		width: 100%;
	}

	.card {
		position: relative;
		padding: 2rem;
		border: none;
		border-radius: 16px;
		background: white;
		cursor: pointer;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
		overflow: hidden;
		text-align: center;
	}

	.card::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		height: 4px;
		transition: all 0.3s ease;
	}

	.defense-card::before {
		background: linear-gradient(90deg, #10b981, #34d399);
	}

	.offense-card::before {
		background: linear-gradient(90deg, #f97316, #fb923c);
	}

	.battle-card::before {
		background: linear-gradient(90deg, #3b82f6, #60a5fa);
	}

	.card:hover {
		transform: translateY(-8px);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
	}

	.card:hover::before {
		height: 6px;
	}

	.card-icon {
		font-size: 3.5rem;
		margin-bottom: 1rem;
		display: inline-block;
	}

	.card h2 {
		margin: 1rem 0 0.5rem 0;
		font-size: 1.75rem;
		color: #1f2937;
		font-weight: 700;
	}

	.card p {
		margin: 0.5rem 0;
		color: #6b7280;
		font-size: 1rem;
	}

	.card p.subtitle {
		font-size: 0.85rem;
		color: #9ca3af;
		margin-bottom: 1.5rem;
	}

    .subtitle {
		font-size: 1rem;
		color: white;
        font-weight: 700;
	}

	.card-arrow {
		font-size: 1.5rem;
		color: #667eea;
		transition: all 0.3s ease;
		margin-top: 1rem;
	}

	.card:hover .card-arrow {
		transform: translateX(4px);
		color: #5568d3;
	}

	/* Responsive */
	@media (max-width: 768px) {
		.header h1 {
			font-size: 2rem;
		}

		.cards-grid {
			grid-template-columns: 1fr;
			gap: 1.5rem;
		}

		.card {
			padding: 1.5rem;
		}

		.card-icon {
			font-size: 2.5rem;
		}

		.card h2 {
			font-size: 1.5rem;
		}
	}
</style>
