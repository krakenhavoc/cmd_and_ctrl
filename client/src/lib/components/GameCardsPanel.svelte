<script lang="ts">
  // GameCardsPanel — "Cards from this game" (ADR 0179 §7): the
  // commanders and cards the opponents cast in a game that has ended,
  // each with shop links, and one "Buy these cards" cart for the lot.
  //
  // Only ever mounted for an ended game, from History in the lobby: the
  // server refuses the read (409) while a game runs, so it cannot show
  // on the table. Opponents are not named (ADR 0165 §9).
  //
  // Renders nothing unless the viewer sees shop links.
  import { onMount } from "svelte";
  import { fetchGameCards } from "../api";
  import { LobbyApiError } from "../session";
  import { boughtList, opponentLabel, type GameCards } from "../gameCards";
  import { visibleShops } from "../shopLinksStore";
  import ShopLinks from "./ShopLinks.svelte";
  import ShopNote from "./ShopNote.svelte";

  interface Props {
    gameID: string;
  }
  const { gameID }: Props = $props();

  let cards = $state<GameCards | null>(null);
  let error = $state("");
  let loading = $state(true);

  onMount(() => {
    void fetchGameCards(gameID)
      .then((c) => (cards = c))
      .catch((err) => {
        error = err instanceof LobbyApiError ? err.message : "could not load this game's cards";
      })
      .finally(() => (loading = false));
  });

  const list = $derived(cards ? boughtList(cards) : []);
</script>

{#if $visibleShops.length > 0}
  <section class="gc" aria-label="Cards from this game" data-testid="game-cards">
    <h3>Cards from this game</h3>
    {#if loading}
      <p class="muted" role="status">Loading…</p>
    {:else if error}
      <p class="muted" role="alert">{error}</p>
    {:else if !cards || !cards.ready}
      <p class="muted" role="status">
        This game's card list is not ready yet. Try again in a moment.
      </p>
    {:else if list.length === 0}
      <p class="muted">No cards were recorded for the other players.</p>
    {:else}
      <p class="muted">What the other players played. Opponents are not named.</p>
      <ShopLinks {list} label="Buy these cards" note={false} />
      <ShopNote />
      <ul class="seats">
        {#each cards.opponents as o, i (i)}
          <li class="seat">
            <p class="who">
              <b>{opponentLabel(i)}</b>
              {#if o.commanders.length > 0}
                <span class="cmd">{o.commanders.join(" / ")}</span>
              {/if}
            </p>
            {#each o.commanders as name (name)}
              <ShopLinks card={name} compact note={false} />
            {/each}
            {#if o.cards.length > 0}
              <details>
                <summary>{o.cards.length} {o.cards.length === 1 ? "card" : "cards"} cast</summary>
                <ul class="cards">
                  {#each o.cards as c (c.name)}
                    <li>
                      <span class="cname">{c.name}</span>
                      <ShopLinks card={c.name} compact note={false} />
                    </li>
                  {/each}
                </ul>
              </details>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </section>
{/if}

<style>
  .gc {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
    padding: 12px 14px;
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--surface);
  }
  h3 {
    margin: 0;
    font-size: 0.95rem;
  }
  .muted {
    margin: 0;
    color: var(--fg-dim);
    font-size: 0.85rem;
  }
  .seats,
  .cards {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 10px;
  }
  .cards {
    gap: 6px;
    margin-top: 6px;
  }
  .cards li {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 2px 10px;
    min-width: 0;
  }
  .cname {
    overflow-wrap: anywhere;
  }
  .who {
    margin: 0 0 2px;
    display: flex;
    flex-wrap: wrap;
    gap: 2px 8px;
  }
  .cmd {
    color: var(--fg-muted);
    overflow-wrap: anywhere;
  }
  summary {
    cursor: pointer;
    color: var(--fg-muted);
    font-size: 0.85rem;
  }
</style>
