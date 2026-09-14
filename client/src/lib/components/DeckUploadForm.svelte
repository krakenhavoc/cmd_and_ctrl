<script lang="ts">
  import { uploadDeck, type UploadDeckResponse } from "../api";
  import { LobbyApiError, type ApiViolation } from "../session";
  import Icon from "./Icon.svelte";
  import PrebuiltDeckPicker from "./PrebuiltDeckPicker.svelte";

  // DeckUploadForm is the shared deck panel used by both Lobby.svelte
  // and the in-game DeckImportModal in Game.svelte (S08.5 wave 1):
  // the pre-built deck picker, then the deck-import textarea + URL
  // field + submit button + violation/warning rendering. Owns its own
  // input + feedback state so callers don't have to thread a
  // Record<gameID, ...> shape through.
  //
  // The picker is mounted HERE rather than beside each call site so
  // that both places get it and neither can drift. It is first
  // because it is the answer for a player who has no decklist and
  // wants a working game — until it existed, that player's only path
  // was to go and build one, and take their chances on catalog
  // coverage when they came back. It hides itself when the server
  // offers no pre-built decks, and the paste box below is unchanged
  // either way.
  //
  // The submit path posts to POST /games/{id}/decks via api.uploadDeck;
  // the server's StateLobby gate rejects post-Start uploads, so the
  // caller is responsible for only mounting this component while the
  // game is still in lobby.

  interface Props {
    gameID: string;
    playerID: string;
    // onSuccess fires after a successful upload so the caller can
    // refresh listings, dismiss the modal, etc. The response payload
    // is forwarded so the caller can surface a richer toast if it
    // wants to (default rendering inside the component is fine).
    onSuccess?: (res: UploadDeckResponse) => void;
  }

  const { gameID, playerID, onSuccess }: Props = $props();

  let source: string = $state("");
  let busy = $state(false);
  let errorMessage = $state("");
  let violations = $state<ApiViolation[]>([]);
  let warnings = $state<ApiViolation[]>([]);
  let successMessage = $state("");
  // Cards in the accepted deck whose printed rules the engine does
  // not carry out. Rendered collapsed: the count is the part that
  // sets expectations, and on a typical Commander deck the list is
  // most of the 100 — worth having, not worth unfurling by default.
  let unimplemented = $state<string[]>([]);

  function violationLabel(v: ApiViolation): string {
    return v.card ? `${v.card}: ${v.message}` : v.message;
  }

  function looksLikeURL(s: string): boolean {
    const t = s.trim();
    return t.startsWith("http://") || t.startsWith("https://");
  }

  function sourceHostname(s: string): string {
    const t = s.trim();
    if (!looksLikeURL(t)) return "";
    try {
      return new URL(t).hostname.replace(/^www\./, "");
    } catch {
      return t;
    }
  }

  function clearFeedback(): void {
    errorMessage = "";
    violations = [];
    warnings = [];
    successMessage = "";
    unimplemented = [];
  }

  async function submit(): Promise<void> {
    if (!source.trim()) {
      clearFeedback();
      errorMessage = "paste a decklist first";
      return;
    }
    busy = true;
    clearFeedback();
    try {
      const res = await uploadDeck(gameID, playerID, source);
      successMessage = `uploaded ${res.deck_name || "deck"}: ${res.card_count} cards, commander: ${res.commanders.join(", ")}`;
      if (res.warnings && res.warnings.length > 0) warnings = res.warnings;
      if (res.unimplemented && res.unimplemented.length > 0) unimplemented = res.unimplemented;
      source = "";
      onSuccess?.(res);
    } catch (err) {
      if (err instanceof LobbyApiError) {
        errorMessage = err.message;
        if (err.violations && err.violations.length > 0) violations = err.violations;
        if (err.warnings && err.warnings.length > 0) warnings = err.warnings;
      } else {
        errorMessage = "upload failed";
      }
    } finally {
      busy = false;
    }
  }
</script>

<div class="deck-upload-form">
  <PrebuiltDeckPicker {gameID} {playerID} {onSuccess} />
  <p class="own-list">
    Or bring your own list — a Moxfield or Archidekt deck URL, a Moxfield JSON export, or a
    plain-text decklist.
  </p>
  <textarea
    rows="7"
    placeholder={"https://moxfield.com/decks/abc123\n\n— or —\n\nCommander:\n1 Atraxa, Praetors' Voice\n\nMainboard:\n1 Sol Ring\n..."}
    bind:value={source}
    aria-label="decklist source"
  ></textarea>
  <div class="row-actions">
    {#if successMessage}
      <p class="success"><Icon name="check" size={13} /> {successMessage}</p>
    {/if}
    <button class="primary" onclick={submit} disabled={busy}>
      {#if busy}
        {#if looksLikeURL(source)}
          fetching from {sourceHostname(source)}…
        {:else}
          uploading…
        {/if}
      {:else if looksLikeURL(source)}
        import deck
      {:else}
        upload deck
      {/if}
    </button>
  </div>
  {#if errorMessage}
    <pre class="deck-error">{errorMessage}</pre>
  {/if}
  {#if violations.length || warnings.length}
    <ul class="viol">
      {#each violations as v (v.code + (v.card ?? "") + v.message)}
        <li class="vi"><span class="code">{v.code}</span>{violationLabel(v)}</li>
      {/each}
      {#each warnings as v (v.code + (v.card ?? "") + v.message)}
        <li class="vi warn"><span class="code">{v.code}</span>{violationLabel(v)}</li>
      {/each}
    </ul>
  {/if}
  {#if unimplemented.length}
    <!-- Not a violation and not a warning: the deck is legal and
         installed. This is the one honest sentence about what the
         engine will and won't do for it, delivered before anyone has
         cast anything. Collapsed by default — the count is what sets
         the expectation; the names are for the player who wants to
         know which ones. -->
    <details class="unimpl">
      <summary>
        {unimplemented.length} of these cards {unimplemented.length === 1 ? "has" : "have"} rules the
        engine doesn't implement yet — {unimplemented.length === 1 ? "it" : "they"} behave as manual sandbox
        cards
      </summary>
      <ul>
        {#each unimplemented as name (name)}
          <li>{name}</li>
        {/each}
      </ul>
    </details>
  {/if}
</div>

<style>
  .own-list {
    margin: 14px 0 0;
    font-size: 12.5px;
    line-height: 1.5;
    color: var(--fg-muted);
  }
  textarea {
    width: 100%;
    min-height: 128px;
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.5;
    padding: 10px 12px;
    margin-top: 10px;
    box-sizing: border-box;
    resize: vertical;
  }
  .row-actions {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: flex-end;
    margin-top: 8px;
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
  .vi.warn {
    background: var(--gold-soft);
    border-color: rgba(217, 180, 92, 0.4);
  }
  .vi.warn .code {
    color: var(--gold-strong);
  }
  /* Muted on purpose. Red is for a deck that was rejected and gold
     for one that lost something; this deck is fine, and the note is
     information rather than a problem. Styling it like a warning
     would teach players to dismiss it. */
  .unimpl {
    margin-top: 8px;
    padding: 8px 10px;
    border-radius: 8px;
    background: var(--surface-sunken);
    border: 1px solid var(--border);
    font-size: 12.5px;
    color: var(--fg-muted);
  }
  .unimpl summary {
    cursor: pointer;
    line-height: 1.45;
  }
  .unimpl ul {
    list-style: none;
    margin: 8px 0 0;
    padding: 0;
    max-height: 180px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-family: var(--font-mono);
    font-size: 11.5px;
    color: var(--fg-dim);
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
</style>
