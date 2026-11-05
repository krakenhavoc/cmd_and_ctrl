<script lang="ts">
  // The opt-in card on #/stats (ADR 0217 §2 and §8): while advanced
  // stats are off for the person, one card that turns them on, the same
  // switch as Settings → Stats. Once on, a quiet line says so and where to
  // switch it off. Drawn only for a signed-in person on a server that can
  // keep the setting.

  import { L } from "../labels";
  import { setAdvancedStats, statsPrefs } from "../statsPrefs";

  let busy = $state(false);
  let error = $state("");

  async function turnOn(): Promise<void> {
    busy = true;
    error = "";
    try {
      await setAdvancedStats(true);
    } catch (e) {
      error = e instanceof Error ? e.message : "could not turn them on; try again";
    } finally {
      busy = false;
    }
  }
</script>

{#if $statsPrefs.available}
  {#if $statsPrefs.advanced}
    <p class="on-note" role="status" data-testid="advanced-stats-on">
      Advanced stats are on. You can switch them off in Settings › Stats.
    </p>
  {:else}
    <section
      class="card opt-in"
      aria-labelledby="advanced-stats-title"
      data-testid="advanced-stats-opt-in"
    >
      <div class="copy">
        <h2 id="advanced-stats-title">Advanced stats</h2>
        <p>
          See more about how your games and decks play, from every game you have already finished.
          Free, and only you see them. You can switch them off at any time.
        </p>
      </div>
      <button type="button" class="btn primary" disabled={busy} onclick={() => void turnOn()}>
        {L.advancedStatsOptIn}
      </button>
      {#if error}
        <p class="error" role="alert">{error}</p>
      {/if}
    </section>
  {/if}
{/if}

<style>
  .on-note {
    margin: 0 0 16px;
    color: var(--fg-dim);
    font-size: 0.85rem;
  }
  .opt-in {
    margin-bottom: 16px;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px 20px;
  }
  .copy {
    flex: 1 1 260px;
    min-width: 0;
  }
  .copy h2 {
    margin: 0 0 4px;
  }
  .copy p {
    margin: 0;
    color: var(--fg-muted);
  }
  .error {
    flex-basis: 100%;
    margin: 0;
    color: var(--danger);
  }
</style>
