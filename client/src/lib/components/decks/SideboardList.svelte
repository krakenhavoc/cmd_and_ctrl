<script lang="ts">
  // SideboardList — a saved deck's sideboard, read only (ADR 0191, amended
  // 2026-10-19, #3833 and #4054). The cards a deck owns outside the game
  // are not a pile on the table; this is where a player checks them, from
  // the deck itself. Opponents never see a sideboard anywhere.

  import { onMount } from "svelte";
  import { fetchDeckPrintings } from "../../api";
  import { sideboardLines, sideboardTotal, type SideboardLine } from "../../deckSideboard";
  import type { MyDeckInfo } from "../../myDecks";

  const { deck, onclose }: { deck: MyDeckInfo; onclose: () => void } = $props();

  let lines = $state<SideboardLine[]>([]);
  let loading = $state(true);
  let error = $state("");
  const total = $derived(sideboardTotal(lines));

  onMount(() => {
    let disposed = false;
    void (async () => {
      try {
        const result = await fetchDeckPrintings(deck.id);
        if (!disposed) lines = sideboardLines(result.cards);
      } catch (err) {
        if (!disposed) {
          error =
            err instanceof Error
              ? err.message
              : "Could not load this deck's sideboard. Close it and try again.";
        }
      } finally {
        if (!disposed) loading = false;
      }
    })();
    return () => {
      disposed = true;
    };
  });
</script>

<section class="sideboard" aria-label={`${deck.name} sideboard`} data-testid="deck-sideboard">
  <h4>
    Sideboard{#if !loading && !error}<span class="count">({total})</span>{/if}
  </h4>
  <p class="note">
    Cards this deck owns outside the game. At the table they come into play only through an effect
    that reaches outside the game, such as a Wish. Nobody else sees them.
  </p>
  {#if loading}
    <p class="note" role="status">Loading the sideboard…</p>
  {:else if error}
    <p class="error" role="alert">{error}</p>
  {:else if lines.length === 0}
    <p class="note">This deck has no sideboard.</p>
  {:else}
    <ul class="lines">
      {#each lines as l (l.name)}
        <li><span class="n">{l.count}×</span> {l.name}</li>
      {/each}
    </ul>
  {/if}
  <div class="actions">
    <button type="button" class="btn sm ghost" onclick={onclose}>Close</button>
  </div>
</section>

<style>
  .sideboard {
    border-top: 1px solid var(--border-strong);
    padding-top: 14px;
    min-width: 0;
  }
  h4 {
    margin: 0 0 6px;
    color: var(--fg);
    font-size: 15px;
  }
  .count {
    margin-inline-start: 0.35em;
    color: var(--fg-muted);
    font-weight: 400;
    font-variant-numeric: tabular-nums;
  }
  .note {
    color: var(--fg-muted);
    font-size: 12px;
    line-height: 1.5;
    margin: 6px 0 12px;
  }
  .lines {
    list-style: none;
    margin: 0 0 12px;
    padding: 0;
    columns: 2 180px;
    column-gap: 16px;
    font-size: 13px;
  }
  .lines li {
    break-inside: avoid;
    padding: 2px 0;
    overflow-wrap: anywhere;
  }
  .n {
    color: var(--fg-muted);
    font-variant-numeric: tabular-nums;
  }
  .error {
    color: var(--danger);
    font-size: 12px;
    overflow-wrap: anywhere;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 12px;
  }
</style>
