<script lang="ts">
  import { uploadDeck, type UploadDeckResponse } from "../api";
  import { LobbyApiError, type ApiViolation } from "../session";

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
    rows="8"
    placeholder={"https://moxfield.com/decks/abc123\n\n— or —\n\nCommander:\n1 Atraxa, Praetors' Voice\n\nMainboard:\n1 Sol Ring\n..."}
    bind:value={source}
    aria-label="decklist source"
  ></textarea>
  <div class="row-actions">
    <button onclick={submit} disabled={busy}>
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
    <pre class="error deck-error">{errorMessage}</pre>
  {/if}
  {#if violations.length}
    <ul class="violations">
      {#each violations as v (v.code + (v.card ?? "") + v.message)}
        <li><span class="code">{v.code}</span> · {violationLabel(v)}</li>
      {/each}
    </ul>
  {/if}
  {#if warnings.length}
    <ul class="warnings">
      {#each warnings as v (v.code + (v.card ?? "") + v.message)}
        <li><span class="code">{v.code}</span> · {violationLabel(v)}</li>
      {/each}
    </ul>
  {/if}
  {#if successMessage}
    <p class="success">{successMessage}</p>
  {/if}
</div>

<style>
  textarea {
    width: 100%;
    font-family: monospace;
    font-size: 0.9em;
    padding: 0.5rem;
    margin-top: 0.5rem;
    box-sizing: border-box;
  }
  .row-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }
  .error {
    color: #c00;
  }
  .deck-error {
    white-space: pre-wrap;
    margin-top: 0.5rem;
    background: #fee;
    padding: 0.5rem;
    border-radius: 3px;
  }
  .violations,
  .warnings {
    margin-top: 0.5rem;
    padding: 0.5rem 0.75rem 0.5rem 1.5rem;
    border-radius: 3px;
    font-size: 0.85em;
  }
  .violations {
    background: #fee;
    color: #900;
  }
  .warnings {
    background: #ffb;
    color: #660;
  }
  .violations li,
  .warnings li {
    margin: 0.15rem 0;
  }
  .code {
    font-family: monospace;
    font-size: 0.9em;
    opacity: 0.75;
  }
  .success {
    color: #060;
    margin-top: 0.5rem;
  }
</style>
