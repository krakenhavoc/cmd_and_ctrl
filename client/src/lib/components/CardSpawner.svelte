<script lang="ts">
  // The card and token spawner.
  //
  // It has two lives (ADR 0075 §2.4, which amends ADR 0023):
  //
  // - **dev** (`managed={false}`, the default) is the original tool.
  //   It lives in the dev dock, it calls the dev routes, and those
  //   404 in production regardless of what this component believes.
  //   On a preview box anyone at the table is a tester, so there is
  //   no per-caller gate at all.
  // - **managed** (`managed`) is the production spawner. Its gates
  //   are the table's, not the deployment's: the caller is the host
  //   or the admin, and the table has switched spawning on. Every
  //   spawn it makes is named in the public game log and is undoable.
  //
  // The mode picks the routes and nothing else, because the panel is
  // the same panel. This component does NOT decide who may open it —
  // its callers do, with devFeature("card_spawn") on one side and
  // canSpawn() on the other — and the server refuses either way.
  //
  // The Tokens tab is offered only in managed mode, because a token
  // has no Scryfall printing and the dev route can only resolve a
  // printing. That gap is what put this feature in production in the
  // first place: a Treasure is the commonest thing a table needs and
  // the one thing the old spawner could never make.
  import {
    searchDevCards,
    searchSpawnCards,
    spawnDevCard,
    spawnOnTable,
    fetchSpawnTokens,
    type DevCardResult,
  } from "../api";
  import { SPAWN_ZONES, spawnZonesFor, type SpawnZone } from "../tableSettings";
  import type { GameView, PlayerView } from "../protocol";

  interface Props {
    gameID: string;
    snapshot: GameView | null;
    // Use the production routes (and offer tokens). False is the dev
    // dock's tool, unchanged.
    managed?: boolean;
  }
  const { gameID, snapshot, managed = false }: Props = $props();

  type Tab = "cards" | "tokens";
  let tab = $state<Tab>("cards");

  let query = $state("");
  let results = $state<DevCardResult[]>([]);
  let selected = $state<DevCardResult | null>(null);
  let tokens = $state<string[]>([]);
  let tokensLoaded = $state(false);
  let tokenQuery = $state("");
  let selectedToken = $state<string | null>(null);
  let zone = $state<SpawnZone>("battlefield");
  let count = $state(1);
  let commander = $state(false);
  let seatID = $state("");
  let busy = $state(false);
  let error = $state<string | null>(null);
  let lastResult = $state<string | null>(null);

  const seats: PlayerView[] = $derived(snapshot?.seats ?? []);
  const kind = $derived<"card" | "token">(tab === "tokens" ? "token" : "card");
  const zones = $derived(managed ? spawnZonesFor(kind) : SPAWN_ZONES);
  const picked = $derived(tab === "tokens" ? selectedToken : (selected?.id ?? null));

  // Default the seat to the first one, and re-default if the seat we
  // had disappears (game swapped under us).
  $effect(() => {
    if (seats.length === 0) return;
    if (!seats.some((p) => p.id === seatID)) {
      seatID = seats[0].id;
    }
  });

  // A token can only go to the battlefield (CR 704.5d — it ceases to
  // exist anywhere else at the next check). Switching to the Tokens
  // tab with "hand" selected would otherwise leave a zone picked that
  // the list no longer offers, and a select bound to an absent option
  // reads as blank.
  $effect(() => {
    if (!zones.includes(zone)) zone = zones[0];
  });

  // The token list is fetched once, lazily, when the tab is first
  // opened: it is ~160 strings and the same for every table, so
  // fetching it on mount would cost a request for a tab most opens
  // never touch.
  $effect(() => {
    if (tab !== "tokens" || !managed || tokensLoaded) return;
    tokensLoaded = true;
    fetchSpawnTokens(gameID).then((t) => (tokens = t));
  });

  const tokenMatches = $derived.by(() => {
    const q = tokenQuery.trim().toLowerCase();
    if (!q) return tokens;
    return tokens.filter((t) => t.toLowerCase().includes(q));
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
      const find = managed
        ? searchSpawnCards(gameID, q, ctrl.signal)
        : searchDevCards(q, ctrl.signal);
      find.then((r) => {
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
    if (!picked || !seatID || busy) return;
    busy = true;
    error = null;
    lastResult = null;
    try {
      if (managed) {
        const res = await spawnOnTable(gameID, {
          token: tab === "tokens" ? (selectedToken ?? undefined) : undefined,
          scryfallID: tab === "cards" ? selected?.id : undefined,
          playerID: seatID,
          zone,
          count,
          commander,
        });
        lastResult = `${res.count}× ${res.name} → ${res.zone}`;
      } else {
        const res = await spawnDevCard(gameID, {
          scryfallID: selected?.id,
          playerID: seatID,
          zone,
          count,
          commander,
        });
        lastResult = `${res.count}× ${res.name} → ${res.zone}`;
      }
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      busy = false;
    }
  }
</script>

<div class="spawner">
  <div class="search">
    {#if managed}
      <div class="tabs" role="tablist" aria-label="what to spawn">
        <button
          role="tab"
          class:sel={tab === "cards"}
          aria-selected={tab === "cards"}
          onclick={() => (tab = "cards")}>Cards</button
        >
        <button
          role="tab"
          class:sel={tab === "tokens"}
          aria-selected={tab === "tokens"}
          onclick={() => (tab = "tokens")}>Tokens</button
        >
      </div>
    {/if}

    {#if tab === "tokens"}
      <input
        type="search"
        bind:value={tokenQuery}
        placeholder="Filter tokens…"
        aria-label="Filter token templates"
        autocomplete="off"
      />
      <ol class="results" aria-label="Token templates">
        {#each tokenMatches as t (t)}
          <li>
            <button
              class="result token"
              class:sel={selectedToken === t}
              onclick={() => (selectedToken = t)}
            >
              <span class="name">{t}</span>
            </button>
          </li>
        {:else}
          <li class="hint">
            {tokens.length === 0
              ? "No token templates on this server."
              : "No template matches that."}
          </li>
        {/each}
      </ol>
    {:else}
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
    {/if}
  </div>

  <div class="controls">
    <div class="picked">
      {#if tab === "tokens"}
        {#if selectedToken}
          <strong>{selectedToken}</strong>
          <span class="type">token</span>
        {:else}
          <span class="hint">Select a token.</span>
        {/if}
      {:else if selected}
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
      <select bind:value={zone} disabled={zones.length === 1}>
        {#each zones as z (z)}<option value={z}>{z}</option>{/each}
      </select>
    </label>

    <label>
      Count
      <input type="number" min="1" max="20" bind:value={count} />
    </label>

    {#if tab !== "tokens"}
      <label class="check">
        <input type="checkbox" bind:checked={commander} />
        Commander
      </label>
    {/if}

    <button class="go" onclick={spawn} disabled={!picked || !seatID || busy}>
      {busy ? "spawning…" : "spawn"}
    </button>

    {#if zone === "battlefield"}
      <p class="note">Battlefield spawns fire ETB triggers.</p>
    {/if}
    {#if tab === "tokens"}
      <p class="note">
        Tokens can only go to the battlefield — anywhere else they vanish at the next check.
      </p>
    {/if}
    {#if managed}
      <p class="note">Every spawn is named in the game log, and undo takes it back for free.</p>
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

  .tabs {
    display: flex;
    gap: 1px;
    padding: 0.4rem 0.4rem 0;
  }
  .tabs button {
    flex: 1;
    padding: 0.25rem 0.5rem;
    border: 1px solid var(--border, #273049);
    border-radius: var(--radius-sm, 4px);
    background: var(--surface, #111a2e);
    color: var(--fg-muted, #9aa5cd);
    font: inherit;
    cursor: pointer;
  }
  .tabs button.sel {
    border-color: var(--gold, #ffd07a);
    color: var(--gold, #ffd07a);
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
  .result.token {
    grid-template-columns: 1fr;
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
