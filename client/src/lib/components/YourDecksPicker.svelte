<script lang="ts">
  import { onMount } from "svelte";
  import { fetchMyDecks, seatLibraryDeck, type UploadDeckResponse } from "../api";
  import { LobbyApiError, session, type ApiViolation } from "../session";
  import { deckSubtitle, isSignedIn, type MyDeckInfo } from "../myDecks";
  import Icon from "./Icon.svelte";

  // YourDecksPicker is the deck-library half of the deck panel (ADR
  // 0051 decision 7, S34 sub-PR 5): pick a deck you've pasted before
  // and seat it again without re-pasting. Sits inside DeckUploadForm,
  // above the paste box, next to PrebuiltDeckPicker — same shape
  // (radio list + one button), on purpose, so a player who has used
  // one already knows how to use the other.
  //
  // Renders nothing for anyone without a user_id (see myDecks.
  // isSignedIn for why that's not a plain truthiness check) or with
  // an empty library — the paste box below is always a working
  // fallback either way.

  interface Props {
    gameID: string;
    playerID: string;
    onSuccess?: (res: UploadDeckResponse) => void;
  }

  const { gameID, playerID: _playerID, onSuccess }: Props = $props();

  const sess = $derived($session);
  const signedIn = $derived(isSignedIn(sess?.principal.user_id));

  let decks = $state<MyDeckInfo[]>([]);
  let loaded = $state(false);
  let selected = $state<string>("");
  let busy = $state(false);
  let errorMessage = $state("");
  let violations = $state<ApiViolation[]>([]);
  let seated = $state("");

  const current = $derived(decks.find((d) => d.id === selected));

  // One shot on mount — the library doesn't change while this panel
  // is open (a save happens through the paste box below, in the same
  // component tree, not here), so there's nothing to re-fetch on.
  onMount(() => {
    if (!signedIn) {
      loaded = true;
      return;
    }
    void fetchMyDecks()
      .then((res) => {
        decks = res.decks ?? [];
        if (!selected && decks.length > 0) selected = decks[0].id;
      })
      .catch(() => {
        // Session expired between render and fetch, or the route
        // isn't there. Either way the section stays hidden.
        decks = [];
      })
      .finally(() => {
        loaded = true;
      });
  });

  function violationLabel(v: ApiViolation): string {
    return v.card ? `${v.card}: ${v.message}` : v.message;
  }

  async function seat(): Promise<void> {
    if (!selected || busy) return;
    busy = true;
    errorMessage = "";
    violations = [];
    seated = "";
    try {
      const res = await seatLibraryDeck(gameID, selected);
      seated = res.deck_name || current?.name || "deck";
      onSuccess?.(res);
    } catch (err) {
      if (err instanceof LobbyApiError) {
        errorMessage = err.message;
        if (err.violations && err.violations.length > 0) violations = err.violations;
      } else {
        errorMessage = "could not seat that deck";
      }
    } finally {
      busy = false;
    }
  }
</script>

{#if loaded && signedIn && decks.length > 0}
  <div class="your-decks">
    <p class="lede">Your decks — seat one without pasting it again.</p>

    <ul class="decks">
      {#each decks as deck (deck.id)}
        <li>
          <label class="deck" class:sel={selected === deck.id}>
            <input type="radio" name="your-deck" value={deck.id} bind:group={selected} />
            <span class="body">
              <span class="name">{deck.name}</span>
              <span class="sub">{deckSubtitle(deck)}</span>
            </span>
          </label>
        </li>
      {/each}
    </ul>

    <div class="row-actions">
      {#if seated}
        <p class="success"><Icon name="check" size={13} /> playing {seated}</p>
      {/if}
      <button class="primary" onclick={seat} disabled={busy || !selected}>
        {busy ? "seating…" : "play this deck"}
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
  /* Deliberately smaller than PrebuiltDeckPicker's styling (no colour
     pips, no coverage detail, no caveats disclosure) — this is a list
     of the player's own decks, not a catalogue to be sold on. */
  .your-decks {
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
  .name {
    font-weight: 600;
    font-size: 13.5px;
    color: var(--fg);
  }
  .sub {
    font-size: 11.5px;
    color: var(--fg-dim);
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
