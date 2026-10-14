<script lang="ts">
  import { onMount } from "svelte";
  import { fetchPrebuiltDecks, installPrebuiltDeck, type UploadDeckResponse } from "../api";
  import { LobbyApiError, session, type ApiViolation } from "../session";
  import { preselectFor, rememberGuestLastDeck, type LastDeck } from "../lastDeck";
  import { signedInUserID } from "../myGames";
  import { imperfectCardLine, summariseCoverage, type PrebuiltDeck } from "../prebuiltDecks";
  import { prebuiltRow } from "../deckPicker";
  import DeckRadioList from "./DeckRadioList.svelte";
  import Icon from "./Icon.svelte";
  import ShopLinks from "./ShopLinks.svelte";

  // PrebuiltDeckPicker is the "I just want to play" path: decks whose
  // cards this engine actually implements, installed with one press. It sits above the paste-a-decklist box in
  // DeckUploadForm.svelte, so both the lobby panel and the in-game
  // deck-import modal get it without either having to know it exists.
  //
  // The coverage line under each deck is not decoration. The decks are
  // build-tested to contain no card the engine fails to REGISTER —
  // which is not the same as "every card works fully", and the gap
  // between those two claims is exactly where a player's trust gets
  // spent. So the picker states the real profile, and the caveated
  // cards are one click away rather than hidden. The rows carry the
  // short figure (deckPicker.ts); the full sentence, the summary and
  // the caveats are said once, for the selected deck, under the list.

  interface Props {
    gameID: string;
    playerID: string;
    /** The deck to preselect when it is a pre-built deck still offered. */
    lastDeck?: LastDeck | null;
    onSuccess?: (res: UploadDeckResponse) => void;
    /**
     * Start as a one-line toggle (the player has saved decks of their
     * own). null holds the section back until the caller knows.
     */
    collapsed?: boolean | null;
    /**
     * Offer "Play this deck in paper" (ADR 0179 §7). Off by default: the
     * in-game deck modal sits on the table route, where no shop link
     * belongs. The lobby turns it on.
     */
    shopLinks?: boolean;
  }

  const {
    gameID,
    playerID,
    lastDeck = null,
    onSuccess,
    collapsed = false,
    shopLinks = false,
  }: Props = $props();

  let expanded = $state(false);
  const open = $derived(collapsed === false || expanded);

  let decks = $state<PrebuiltDeck[]>([]);
  // null until the probe settles. A server without the route (or a
  // logged-out caller) leaves this empty and the whole section never
  // renders — the paste box below is always a working fallback, so a
  // missing picker costs a convenience, not the feature.
  let loaded = $state(false);
  let selected = $state<string>("");
  let busy = $state(false);
  let errorMessage = $state("");
  let violations = $state<ApiViolation[]>([]);
  let installed = $state("");

  const current = $derived(decks.find((d) => d.id === selected));
  const rows = $derived(decks.map(prebuiltRow));

  // Preselect the last deck once both the list and the last deck are
  // in, unless the player has already picked one themselves.
  let touched = false;
  $effect(() => {
    const pick = preselectFor(
      "prebuilt",
      decks.map((d) => d.id),
      lastDeck,
    );
    if (pick && !touched) selected = pick;
  });
  const coverage = $derived(current ? summariseCoverage(current.coverage) : null);
  const imperfect = $derived(current?.coverage.imperfect ?? []);

  // One shot on mount, not $effect: the catalog is fixed for the life
  // of the build, and re-fetching it because a prop changed would be
  // a request per keystroke somewhere upstream.
  onMount(() => {
    void fetchPrebuiltDecks()
      .then((res) => {
        decks = res.decks ?? [];
        if (!selected && decks.length > 0) selected = decks[0].id;
      })
      .catch(() => {
        // A server without the route, or a caller whose session has
        // expired. Either way the section stays hidden and the paste
        // box below still works.
        decks = [];
      })
      .finally(() => {
        loaded = true;
      });
  });

  function violationLabel(v: ApiViolation): string {
    return v.card ? `${v.card}: ${v.message}` : v.message;
  }

  async function install(): Promise<void> {
    if (!selected || busy) return;
    busy = true;
    errorMessage = "";
    violations = [];
    installed = "";
    try {
      const res = await installPrebuiltDeck(gameID, playerID, selected);
      installed = res.deck_name || current?.name || "deck";
      // A signed-in person's last deck is recorded by the server; a
      // guest's lives in this browser (ADR 0110 §5 item 5).
      if (signedInUserID($session) === null) rememberGuestLastDeck(res.deck_id || selected);
      onSuccess?.(res);
    } catch (err) {
      if (err instanceof LobbyApiError) {
        errorMessage = err.message;
        if (err.violations && err.violations.length > 0) violations = err.violations;
      } else {
        errorMessage = "could not install that deck";
      }
    } finally {
      busy = false;
    }
  }
</script>

{#if loaded && decks.length > 0 && collapsed !== null}
  <div class="prebuilt">
    {#if collapsed}
      <button
        type="button"
        class="toggle"
        class:open
        aria-expanded={open}
        onclick={() => (expanded = !expanded)}
      >
        <Icon name="chevronRight" size={12} />
        or play a pre-built deck ({decks.length})
      </button>
    {/if}
    {#if open}
      <p class="lede">
        Pick one and press play. Every card in these decks is implemented by this engine — each deck
        says exactly how completely.
      </p>

      <DeckRadioList
        {rows}
        name="prebuilt-deck"
        bind:selected
        filterLabel="filter the pre-built decks"
        onpick={() => (touched = true)}
      />

      {#if current && coverage}
        <div class="chosen">
          {#if current.summary}<p class="summary">{current.summary}</p>{/if}
          <p class="cov cov-{coverage.tone}">
            <i class="dot" class:ok={coverage.tone === "full"}></i>{coverage.headline}
          </p>
          {#if coverage.detail}
            <p class="detail"><b>{current.name}</b>: {coverage.detail}.</p>
          {/if}
        </div>
      {/if}

      {#if imperfect.length > 0}
        <!-- Same register as DeckUploadForm's `unimplemented` block, and
           for the same reason: this deck is fine and installed-able.
           Collapsed because the count is what sets the expectation and
           the names are for whoever wants to check a specific card. -->
        <details class="caveats">
          <summary>
            which {imperfect.length}
            {imperfect.length === 1 ? "card" : "cards"}, and what changes?
          </summary>
          <ul>
            {#each imperfect as card (card.name)}
              <li>
                <b>{card.name}</b>
                <span>{imperfectCardLine(card)}</span>
              </li>
            {/each}
          </ul>
        </details>
      {/if}

      {#if shopLinks && current?.cards && current.cards.length > 0}
        <div class="paper">
          <ShopLinks list={current.cards} label="Play this deck in paper" />
        </div>
      {/if}

      <div class="row-actions">
        {#if installed}
          <p class="success"><Icon name="check" size={13} /> playing {installed}</p>
        {/if}
        <button class="primary" onclick={install} disabled={busy || !selected}>
          {busy ? "installing…" : current ? `play ${current.name}` : "play this deck"}
        </button>
      </div>

      {#if errorMessage}
        <pre class="deck-error">{errorMessage}</pre>
      {/if}
      {#if violations.length}
        <ul class="viol">
          {#each violations as v (v.code + (v.card ?? "") + v.message)}
            <li class="vi"><span class="code">{v.code}</span>{violationLabel(v)}</li>
          {/each}
        </ul>
      {/if}
    {/if}
  </div>
{/if}

<style>
  .prebuilt {
    margin-top: 10px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--border);
  }
  /* The collapsed section: a quiet text button, not a second primary. */
  .toggle {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin: 0;
    padding: 2px 0;
    border: 0;
    background: none;
    color: var(--fg-muted);
    font: inherit;
    font-size: 12.5px;
    cursor: pointer;
  }
  .toggle:hover {
    color: var(--fg);
  }
  .toggle :global(svg) {
    transition: transform 120ms var(--ease);
  }
  .toggle.open {
    margin-bottom: 8px;
  }
  .toggle.open :global(svg) {
    transform: rotate(90deg);
  }
  .lede {
    margin: 0 0 10px;
    font-size: 12.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .chosen {
    margin-top: 8px;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .summary {
    margin: 0;
    font-size: 12px;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .cov {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin: 0;
    font-size: 11.5px;
    color: var(--fg-muted);
  }
  .cov .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--accent);
    flex: 0 0 auto;
  }
  /* Green only for a deck that really does play as printed, so the
     colour keeps meaning something. */
  .cov .dot.ok {
    background: var(--mint);
  }
  .cov-gap .dot {
    background: var(--danger);
  }
  .detail b {
    color: var(--fg-muted);
    font-weight: 600;
  }
  .detail {
    margin: 0;
    font-size: 11.5px;
    line-height: 1.45;
    color: var(--fg-dim);
  }
  .caveats {
    margin-top: 6px;
    font-size: 11.5px;
    color: var(--fg-muted);
  }
  .caveats summary {
    cursor: pointer;
  }
  .caveats ul {
    list-style: none;
    margin: 8px 0 0;
    padding: 0;
    max-height: 200px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .caveats li {
    display: flex;
    flex-direction: column;
    gap: 1px;
    line-height: 1.4;
    color: var(--fg-dim);
  }
  .caveats b {
    color: var(--fg-muted);
    font-weight: 600;
  }
  .row-actions {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: flex-end;
    margin-top: 10px;
  }
  /* "play <deck name>": a long name ellipsises rather than widening
     the button past the panel. */
  .row-actions button {
    max-width: min(100%, 22rem);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .success {
    margin: 0;
    margin-right: auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--mint);
    font-size: 12.5px;
  }
  .deck-error {
    white-space: pre-wrap;
    margin: 10px 0 0;
    padding: 8px 10px;
    border-radius: 8px;
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 30%, transparent);
    color: var(--fg);
    font-family: var(--font-ui);
    font-size: 12.5px;
  }
  .viol {
    list-style: none;
    margin: 8px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .vi {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border-radius: 8px;
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 30%, transparent);
    font-size: 12.5px;
    color: var(--fg);
  }
  .vi .code {
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.08em;
    color: var(--danger);
    font-weight: 700;
    flex: 0 0 auto;
  }
</style>
