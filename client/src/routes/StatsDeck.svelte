<script lang="ts">
  // StatsDeck — one deck's advanced stats (#/stats/deck/<deck_id>, ADR
  // 0217 §5 and §8), and the same page inside a shared view
  // (#/stats/shared/<token>/deck/<deck_id>, `token` set), where it is
  // read-only, needs no sign-in, shows the owner's name and leaves out
  // card images.
  //
  // The head leads with the commanders' art, then the version picker (§5:
  // since the last edit by default, all time, or one earlier list) and the
  // table makeup and date range filters. The body is DeckStats. While the
  // person's advanced stats are off the page says what they are and offers
  // the switch, the same card as #/stats. A deck that is not theirs, or
  // was deleted, is "not found"; a revoked shared link says so.
  import { onMount, untrack } from "svelte";
  import Page from "../lib/components/Page.svelte";
  import AdvancedStatsOptIn from "../lib/components/AdvancedStatsOptIn.svelte";
  import CommanderArt from "../lib/components/decks/CommanderArt.svelte";
  import DeckStats from "../lib/components/stats/DeckStats.svelte";
  import StatsFilters from "../lib/components/stats/StatsFilters.svelte";
  import { L } from "../lib/labels";
  import { loadStatsPrefs, statsPrefs } from "../lib/statsPrefs";
  import { signedInUserID } from "../lib/myGames";
  import { LobbyApiError, session } from "../lib/session";
  import {
    fetchAdvancedDeck,
    fetchSharedDeck,
    noGamesSinceEdit,
    VERSION_ALL,
    VERSION_CURRENT,
    versionNote,
    versionText,
    type AdvancedDeck,
    type Makeup,
    type SharedDeck,
    type StatsRange,
  } from "../lib/advancedStats";

  interface Props {
    deckId: string;
    /** A shared view's token: the page is then read-only and public. */
    token?: string;
  }
  const { deckId, token }: Props = $props();

  const shared = $derived(token !== undefined);
  const isPerson = $derived(signedInUserID($session) !== null);
  const advancedOn = $derived($statsPrefs.available && $statsPrefs.advanced);

  let data = $state<AdvancedDeck | SharedDeck | null>(null);
  let loading = $state(true);
  let error = $state("");
  let notFound = $state(false);
  let off = $state(false);
  let makeup = $state<Makeup>("all");
  let range = $state<StatsRange>("all");
  let version = $state<string>(VERSION_CURRENT);
  let failedArt = $state<string[]>([]);

  async function load(): Promise<void> {
    loading = true;
    error = "";
    notFound = false;
    off = false;
    try {
      const got = token
        ? await fetchSharedDeck(token, deckId, version, makeup, range)
        : await fetchAdvancedDeck(deckId, version, makeup, range);
      data = got;
      // The server may read "current" as "all" (a list that resolves no card).
      version = got.version;
    } catch (err) {
      if (err instanceof LobbyApiError && err.status === 404) {
        notFound = true;
        data = null;
      } else if (err instanceof LobbyApiError && err.status === 403 && !token) {
        off = true;
        data = null;
        void loadStatsPrefs();
      } else {
        error = err instanceof LobbyApiError ? err.message : "could not load this deck's stats";
      }
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if (!shared && isPerson) void loadStatsPrefs();
  });

  // Another deck or link: start again from "since last edit". Only the
  // route is tracked; a filter or version change reads on its own.
  $effect(() => {
    void deckId;
    void token;
    untrack(() => {
      version = VERSION_CURRENT;
      data = null;
      failedArt = [];
      if (shared || isPerson) void load();
      else loading = false;
    });
  });

  // The person switched advanced stats on from this page's card.
  $effect(() => {
    if (advancedOn && off) untrack(() => void load());
  });

  function onFilters(m: Makeup, r: StatsRange): void {
    makeup = m;
    range = r;
    void load();
  }

  function onVersion(e: Event): void {
    version = (e.currentTarget as HTMLSelectElement).value;
    void load();
  }

  function showAll(): void {
    version = VERSION_ALL;
    void load();
  }

  const art = $derived(
    (data?.deck.commander_cards ?? [])
      .filter((c) => !!c.scryfall_id && !failedArt.includes(c.scryfall_id))
      .map((c) => ({ name: c.name, scryfallId: c.scryfall_id! })),
  );
  const earlier = $derived((data?.versions ?? []).filter((v) => !v.current));
  const backHref = $derived(token ? `#/stats/shared/${encodeURIComponent(token)}` : "#/stats");
  const owner = $derived(data && "display_name" in data ? data.display_name : "");
  const title = $derived(data ? data.deck.name : "Deck stats");
  const lede = $derived(
    shared
      ? "A read-only copy of one deck's stats, shared by its player. Other players are never named."
      : "How this deck does at your tables. Only your own seats are counted.",
  );
</script>

<Page eyebrow={owner ? `${owner}'s deck stats` : "Deck stats"} {title} {lede}>
  {#snippet actions()}
    <a class="btn" href={backHref} data-testid="deck-stats-back">All stats</a>
  {/snippet}

  {#if !shared && !isPerson}
    <p class="card note" role="status">
      Statistics belong to a signed-in account. Sign in with Discord to see how your decks do.
    </p>
  {:else if notFound}
    <p class="card note" role="status" data-testid="deck-stats-gone">
      {shared
        ? "This stats link or this deck doesn't work. The link may have been revoked, or the deck removed. Ask the person who shared it."
        : "This deck has no stats page. It may have been deleted, or it isn't one of your saved decks."}
    </p>
  {:else if off}
    <p class="muted">
      Per-deck stats are part of advanced stats, which are off for you. Turn them on to see this
      page.
    </p>
    <AdvancedStatsOptIn />
  {:else if error}
    <p class="card note error" role="alert">{error}</p>
    <button type="button" class="btn" onclick={() => void load()}>Try again</button>
  {:else if !data}
    <p class="muted" role="status">Loading this deck's stats…</p>
  {:else}
    <div data-testid="deck-stats">
      <div class="head">
        {#if !shared && art.length > 0}
          <div class="band" aria-hidden="true">
            {#each art as a (a.scryfallId)}
              <CommanderArt
                scryfallId={a.scryfallId}
                name={a.name}
                onfail={() => (failedArt = [...failedArt, a.scryfallId])}
              />
            {/each}
          </div>
        {/if}
        <div class="who">
          <p class="cmdrs">{data.deck.commanders.join(" and ") || "No commander"}</p>
          <div class="picker">
            <label for="deck-version">{L.statsVersionPicker}</label>
            <select
              id="deck-version"
              name="deck-version"
              autocomplete="off"
              value={version}
              onchange={onVersion}
              disabled={loading}
              data-testid="deck-version"
            >
              <option value={VERSION_CURRENT}>Since your last edit</option>
              <option value={VERSION_ALL}>All time</option>
              {#each earlier as v (v.hash)}
                <option value={v.hash}>{versionText(v)}</option>
              {/each}
            </select>
            <p class="note-line" role="status" data-testid="deck-version-note">
              {versionNote(data)}
            </p>
          </div>
        </div>
      </div>

      <StatsFilters {makeup} {range} onchange={onFilters} disabled={loading} />

      {#if data.no_games_since_edit}
        <div class="card note" role="status" data-testid="deck-no-games-since-edit">
          <p>{noGamesSinceEdit(data)}</p>
          <button type="button" class="btn" onclick={showAll}>{L.statsAllDeckGames}</button>
        </div>
      {/if}

      <div class:stale={loading} class="body">
        <DeckStats deck={data} {shared} />
      </div>
    </div>
  {/if}
</Page>

<style>
  .muted {
    color: var(--fg-muted);
    margin: 0 0 12px;
  }
  .note.error {
    color: var(--danger);
    border-color: var(--danger);
    margin-bottom: 12px;
  }
  .head {
    display: grid;
    gap: 14px;
    margin-bottom: 14px;
  }
  .band {
    display: flex;
    height: 120px;
    border-radius: var(--radius-lg);
    overflow: hidden;
    border: 1px solid var(--border);
    background: var(--surface-sunken);
  }
  .who {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: 12px 24px;
  }
  .cmdrs {
    margin: 0;
    font-weight: 600;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .picker {
    display: grid;
    gap: 4px;
    flex: 1 1 280px;
    min-width: 0;
    max-width: 100%;
  }
  .picker label {
    font: 600 0.72rem/1 var(--font-mono);
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .picker select {
    width: 100%;
    min-width: 0;
    text-overflow: ellipsis;
    min-height: 40px;
    padding: 6px 10px;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    background: var(--surface);
    color: var(--fg);
    font: inherit;
  }
  .note-line {
    margin: 0;
    color: var(--fg-muted);
    font-size: 0.85rem;
  }
  .card.note {
    display: grid;
    gap: 10px;
    justify-items: start;
    margin-top: 12px;
  }
  .card.note p {
    margin: 0;
  }
  .body {
    margin-top: 16px;
  }
  .stale {
    opacity: 0.6;
    transition: opacity 120ms var(--ease);
  }
</style>
