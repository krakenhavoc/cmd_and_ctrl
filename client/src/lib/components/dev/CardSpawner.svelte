<script lang="ts">
  // Dev-only card spawner (ADR 0023).
  //
  // Search the Scryfall index by name, pick a seat and a zone, spawn.
  // The point is to reach a board state in five seconds instead of
  // twenty turns — the workflow this was built for is "implement a
  // card, put it on the battlefield next to the thing it interacts
  // with, watch what the engine does".
  //
  // Everything here is gated twice over: the component only mounts
  // when /config reports the card_spawn feature, and the routes it
  // calls 404 in production regardless of what this component
  // believes.
  import { searchDevCards, spawnDevCard, type DevCardResult, type DevSpawnZone } from "../../api";
  import type { GameView, PlayerView } from "../../protocol";

  interface Props {
    gameID: string;
    snapshot: GameView | null;
  }
  const { gameID, snapshot }: Props = $props();

  const ZONES: DevSpawnZone[] = ["battlefield", "hand", "graveyard", "exile", "library", "command"];

  let query = $state("");
  let results: DevCardResult[] = $state([]);
  let selected: DevCardResult | null = $state(null);
  let zone: DevSpawnZone = $state("battlefield");
  let count = $state(1);
  let commander = $state(false);
  let seatID = $state("");
  let busy = $state(false);
  let error: string | null = $state(null);
  let lastResult: string | null = $state(null);

  const seats: PlayerView[] = $derived(snapshot?.seats ?? []);

  // Default the seat to the first one, and re-default if the seat we
  // had disappears (game swapped under us).
  $effect(() => {
    if (seats.length === 0) return;
    if (!seats.some((p) => p.id === seatID)) {
      seatID = seats[0].id;
    }
  });

  // Debounced search. The AbortController matters more than the
  // debounce: without it a slow early query can resolve after a fast
  // later one and overwrite the results with stale rows.
  let searchTimer: ReturnType<typeof setTimeout> | null = null;
  let inflight: AbortController | null = null;

  $effect(() => {
    const q = query.trim();
    if (searchTimer) clearTimeout(searchTimer);
    if (q.length < 2) {
      results = [];
      return;
    }
    searchTimer = setTimeout(() => {
      inflight?.abort();
      const ctrl = new AbortController();
      inflight = ctrl;
      searchDevCards(q, ctrl.signal).then((r) => {
        if (ctrl.signal.aborted) return;
        results = r;
      });
    }, 200);
    return () => {
      if (searchTimer) clearTimeout(searchTimer);
    };
  });

  function pick(c: DevCardResult) {
    selected = c;
    // Spawning to the command zone almost always means "treat this as
    // a commander"; pre-tick it rather than making it a second step
    // people forget and then wonder why the card is inert.
    if (zone === "command") commander = true;
  }

  async function spawn() {
    if (!selected || !seatID || busy) return;
    busy = true;
    error = null;
    lastResult = null;
    try {
      const res = await spawnDevCard(gameID, {
        scryfallID: selected.id,
        playerID: seatID,
        zone,
        count,
        commander,
      });
      lastResult = `${res.count}× ${res.name} → ${res.zone}`;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      busy = false;
    }
  }
</script>

<div class="spawner">
  <div class="search">
    <input
      type="search"
      bind:value={query}
      placeholder="Card name…"
      aria-label="Search cards by name"
      autocomplete="off"
    />
    <ol class="results" aria-label="Search results">
      {#each results as c (c.id)}
        <li>
          <button class="result" class:sel={selected?.id === c.id} onclick={() => pick(c)}>
            <span class="name">{c.name}</span>
            <span class="cost">{c.mana_cost}</span>
            <span class="type">{c.type_line}</span>
            <span class="set">{c.set.toUpperCase()}</span>
          </button>
        </li>
      {:else}
        <li class="hint">
          {query.trim().length < 2 ? "Type at least two characters." : "No matches."}
        </li>
      {/each}
    </ol>
  </div>

  <div class="controls">
    <div class="picked">
      {#if selected}
        <strong>{selected.name}</strong>
        <span class="type">{selected.type_line}</span>
      {:else}
        <span class="hint">Select a card.</span>
      {/if}
    </div>

    <label>
      Seat
      <select bind:value={seatID}>
        {#each seats as p (p.id)}
          <option value={p.id}>{p.name}</option>
        {/each}
      </select>
    </label>

    <label>
      Zone
      <select bind:value={zone}>
        {#each ZONES as z (z)}<option value={z}>{z}</option>{/each}
      </select>
    </label>

    <label>
      Count
      <input type="number" min="1" max="20" bind:value={count} />
    </label>

    <label class="check">
      <input type="checkbox" bind:checked={commander} />
      Commander
    </label>

    <button class="go" onclick={spawn} disabled={!selected || !seatID || busy}>
      {busy ? "spawning…" : "spawn"}
    </button>

    {#if zone === "battlefield"}
      <p class="note">Battlefield spawns fire ETB triggers.</p>
    {/if}
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    {#if lastResult}<p class="ok" role="status">{lastResult}</p>{/if}
  </div>
</div>

<style>
  .spawner {
    display: grid;
    grid-template-columns: 1fr 15rem;
    height: 100%;
    min-height: 0;
  }

  .search {
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-right: 1px solid var(--border, #273049);
  }
  .search input[type="search"] {
    margin: 0.4rem;
    padding: 0.3rem 0.5rem;
    border: 1px solid var(--border, #273049);
    border-radius: var(--radius-sm, 4px);
    background: var(--surface-sunken, #0a1122);
    color: var(--fg, #e6ecff);
    font: inherit;
  }

  .results {
    flex: 1;
    margin: 0;
    padding: 0;
    list-style: none;
    overflow-y: auto;
  }
  .result {
    display: grid;
    grid-template-columns: 1fr auto 9rem 2.5rem;
    gap: 0.5em;
    align-items: baseline;
    width: 100%;
    padding: 0.2rem 0.5rem;
    border: none;
    background: none;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }
  .result:hover {
    background: var(--surface, #111a2e);
  }
  .result.sel {
    background: var(--accent-soft, rgba(122, 167, 255, 0.18));
  }
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .cost {
    color: var(--gold, #ffd07a);
  }
  .type,
  .set {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--fg-dim, #6c7a99);
  }

  .controls {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    padding: 0.5rem;
    overflow-y: auto;
  }
  .picked {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    min-height: 2.4em;
    padding-bottom: 0.3rem;
    border-bottom: 1px solid var(--border, #273049);
  }
  .controls label {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.5rem;
    color: var(--fg-muted, #9aa5cd);
  }
  .controls label.check {
    justify-content: flex-start;
  }
  .controls select,
  .controls input[type="number"] {
    flex: 1;
    max-width: 8rem;
    padding: 0.15rem 0.3rem;
    border: 1px solid var(--border, #273049);
    border-radius: var(--radius-sm, 4px);
    background: var(--surface, #111a2e);
    color: var(--fg, #e6ecff);
    font: inherit;
  }
  .go {
    margin-top: 0.2rem;
    padding: 0.35rem;
    border: 1px solid var(--gold, #ffd07a);
    border-radius: var(--radius-sm, 4px);
    background: var(--gold-soft, rgba(255, 208, 122, 0.18));
    color: var(--gold, #ffd07a);
    font: inherit;
    font-weight: 700;
    letter-spacing: 0.08em;
    cursor: pointer;
  }
  .go:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .note,
  .hint {
    margin: 0;
    color: var(--fg-dim, #6c7a99);
  }
  .error {
    margin: 0;
    color: var(--danger, #ff7a7a);
  }
  .ok {
    margin: 0;
    color: var(--mint, #7aff9a);
  }
</style>
