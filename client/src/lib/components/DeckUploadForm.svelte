<script lang="ts">
  import { uploadDeck, type UploadDeckResponse } from "../api";
  import { LobbyApiError, type ApiViolation } from "../session";
  import Icon from "./Icon.svelte";

  // DeckUploadForm is the shared deck-import textarea + URL field +
  // submit button + violation/warning rendering used by both
  // Lobby.svelte and the in-game DeckImportModal in Game.svelte
  // (S08.5 wave 1). Owns its own input + feedback state so callers
  // don't have to thread a Record<gameID, ...> shape through.
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
</div>

<style>
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
