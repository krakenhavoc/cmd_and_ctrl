<script lang="ts">
  import { onMount } from "svelte";
  import { fetchPrebuiltDecks, installPrebuiltDeck, type UploadDeckResponse } from "../api";
  import { LobbyApiError, type ApiViolation } from "../session";
  import {
    deckSubtitle,
    imperfectCardLine,
    orderedColors,
    summariseCoverage,
    type PrebuiltDeck,
  } from "../prebuiltDecks";
  import Icon from "./Icon.svelte";

  // PrebuiltDeckPicker is the "I just want to play" path: four decks
  // whose cards this engine actually implements, installed with one
  // press. It sits above the paste-a-decklist box in
  // DeckUploadForm.svelte, so both the lobby panel and the in-game
  // deck-import modal get it without either having to know it exists.
  //
  // The coverage line under each deck is not decoration. The decks are
  // build-tested to contain no card the engine fails to REGISTER —
  // which is not the same as "every card works fully", and the gap
  // between those two claims is exactly where a player's trust gets
  // spent. So the picker states the real profile, and the caveated
  // cards are one click away rather than hidden.

  interface Props {
    gameID: string;
    playerID: string;
    onSuccess?: (res: UploadDeckResponse) => void;
  }

  const { gameID, playerID, onSuccess }: Props = $props();

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

{#if loaded && decks.length > 0}
  <div class="prebuilt">
    <p class="lede">
      Pick one and press play. Every card in these decks is implemented by this engine — each deck
      says exactly how completely.
    </p>

    <ul class="decks">
      {#each decks as deck (deck.id)}
        {@const cov = summariseCoverage(deck.coverage)}
        <li>
          <label class="deck" class:sel={selected === deck.id}>
            <input type="radio" name="prebuilt-deck" value={deck.id} bind:group={selected} />
            <span class="body">
              <span class="top">
                <span class="name">{deck.name}</span>
                <span class="pips">
                  {#each orderedColors(deck.colors) as c (c)}
                    <i class="pip pip-{c}">{c}</i>
                  {/each}
                </span>
              </span>
              <span class="sub">{deckSubtitle(deck)}</span>
              {#if deck.summary}<span class="summary">{deck.summary}</span>{/if}
              <span class="cov cov-{cov.tone}">
                <i class="dot" class:ok={cov.tone === "full"}></i>{cov.headline}
              </span>
            </span>
          </label>
        </li>
      {/each}
    </ul>

    {#if current && coverage && coverage.detail}
      <p class="detail"><b>{current.name}</b>: {coverage.detail}.</p>
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

    <div class="row-actions">
      {#if installed}
        <p class="success"><Icon name="check" size={13} /> playing {installed}</p>
      {/if}
      <button class="primary" onclick={install} disabled={busy || !selected}>
        {busy ? "installing…" : "play this deck"}
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
  </div>
{/if}

<style>
  .prebuilt {
    margin-top: 10px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--border);
  }
  .lede {
    margin: 0 0 10px;
    font-size: 12.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  .decks {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
  }
  .deck {
    display: flex;
    gap: 10px;
    align-items: flex-start;
    padding: 10px 12px;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--surface-sunken);
    cursor: pointer;
  }
  .deck.sel {
    border-color: var(--accent);
    background: var(--accent-soft);
  }
  .deck input {
    margin-top: 3px;
    flex: 0 0 auto;
  }
  .body {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }
  .top {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  .name {
    font-weight: 600;
    font-size: 13.5px;
    color: var(--fg);
  }
  .pips {
    display: inline-flex;
    gap: 3px;
  }
  /* Same tints and the same "colour on the text, not the fill"
     treatment as the catalogue page's colour filter, so a player who
     has seen one recognises the other. */
  .pip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    border: 1px solid currentColor;
    background: color-mix(in srgb, currentColor 16%, transparent);
    font-size: 9px;
    font-style: normal;
    font-weight: 700;
  }
  .pip-W {
    color: #e8dcc0;
  }
  .pip-U {
    color: #7fb4e8;
  }
  .pip-B {
    color: #b09ac4;
  }
  .pip-R {
    color: #e88b7f;
  }
  .pip-G {
    color: #7fc79a;
  }
  .sub {
    font-size: 11.5px;
    color: var(--fg-dim);
    text-transform: lowercase;
  }
  .summary {
    font-size: 12px;
    line-height: 1.45;
    color: var(--fg-muted);
  }
  .cov {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 2px;
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
    margin: 8px 0 0;
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
    background: rgba(255, 107, 107, 0.08);
    border: 1px solid rgba(255, 107, 107, 0.3);
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
    background: rgba(255, 107, 107, 0.08);
    border: 1px solid rgba(255, 107, 107, 0.3);
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
