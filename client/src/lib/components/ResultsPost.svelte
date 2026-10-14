<script lang="ts">
  // The table's results post in the CMD & CTRL Discord (ADR 0174), as a
  // section of the lobby's table settings. Everyone who can see the
  // settings sees whether the result will be posted, which matters
  // because the post names the seats; the host (or an admin) also gets
  // the box. The server enforces all of it: a disabled box is only so
  // the click never happens, and its hint says why.

  import type { GameMeta } from "../api";
  import { resultsBlocked, resultsStatusLine } from "../resultsPost";

  interface Props {
    game: GameMeta;
    // Host or admin (canManageTableFor).
    canManage: boolean;
    ontoggle: (enabled: boolean) => void;
    busy?: boolean;
    error?: string | null;
  }

  const { game, canManage, ontoggle, busy = false, error = null }: Props = $props();

  const status = $derived(resultsStatusLine(game));
  const blocked = $derived(resultsBlocked(game));
</script>

<section class="results-post" aria-labelledby={`results-title-${game.id}`}>
  <h3 id={`results-title-${game.id}`}>Game result in Discord</h3>
  <p class="results-status" data-testid="results-status">{status}</p>

  {#if canManage}
    <label class="results-toggle">
      <input
        type="checkbox"
        checked={game.results === true}
        disabled={busy || blocked !== null}
        onchange={(e) => ontoggle((e.currentTarget as HTMLInputElement).checked)}
      />
      <span>
        Post the result to the CMD & CTRL Discord
        <span class="results-sum"
          >the table's name, each seat's name and commanders, who won, the turns and how long it
          took. Nobody is pinged.</span
        >
      </span>
    </label>
    {#if blocked}
      <p class="results-hint" data-testid="results-hint">{blocked}</p>
    {/if}
  {/if}
  {#if error}
    <p class="error" role="alert">{error}</p>
  {/if}
</section>

<style>
  .results-post {
    margin-top: 1rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--border, rgba(127, 127, 127, 0.3));
  }
  h3 {
    margin: 0 0 0.35rem;
    font-size: 0.95rem;
  }
  .results-status,
  .results-hint {
    margin: 0.25rem 0;
  }
  .results-hint {
    opacity: 0.85;
    font-size: 0.9em;
  }
  .results-toggle {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }
  .results-sum {
    display: block;
    opacity: 0.75;
    font-size: 0.85em;
  }
</style>
