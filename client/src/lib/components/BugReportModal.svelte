<script lang="ts">
  // BugReportModal — the in-app "report a bug" form (ADR 0017).
  // Collects a title + description, optionally snapshots game
  // context (game ID, turn, phase/step, seq, connection status),
  // and POSTs /bugreport; the server renders the issue body and
  // files it in the project's GitHub repo. On success the modal
  // links straight to the created issue.
  //
  // Unlike the server-driven prompt modals (discard / choice), this
  // one is user-opened and user-dismissable: Cancel closes it and
  // nothing is sent.

  import { submitBugReport } from "../api";
  import { BUG_DESC_MAX, BUG_TITLE_MAX, buildBugContext, validateBugReport } from "../bugReport";
  import type { GameView } from "../protocol";

  interface Props {
    gameID: string;
    view: GameView | null;
    seq: number;
    connection: string;
    onclose: () => void;
  }

  const { gameID, view, seq, connection, onclose }: Props = $props();

  let title = $state("");
  let description = $state("");
  let includeContext = $state(true);
  let submitting = $state(false);
  let error = $state<string | null>(null);
  let filedURL = $state<string | null>(null);
  let filedNumber = $state<number | null>(null);

  // Context is snapshotted at submit time, but summarised live so
  // the reporter can see what the checkbox will attach.
  const contextSummary = $derived.by(() => {
    const ctx = buildBugContext(gameID, view, seq, connection);
    const bits = [`game ${gameID.slice(0, 8)}`];
    if (ctx.turn) bits.push(`turn ${ctx.turn}, ${ctx.phase}/${ctx.step}`);
    if (ctx.seq) bits.push(`seq ${ctx.seq}`);
    if (ctx.connection) bits.push(`ws ${ctx.connection}`);
    return bits.join(" · ");
  });

  async function submit(): Promise<void> {
    const problem = validateBugReport(title, description);
    if (problem) {
      error = problem;
      return;
    }
    submitting = true;
    error = null;
    try {
      const ctx = includeContext ? buildBugContext(gameID, view, seq, connection) : undefined;
      const res = await submitBugReport(title.trim(), description, ctx);
      filedURL = res.url;
      filedNumber = res.number;
    } catch (e) {
      error = e instanceof Error ? e.message : "filing the issue failed";
    } finally {
      submitting = false;
    }
  }
</script>

<div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="bug-report-title">
  <div class="modal">
    {#if filedURL}
      <h2 id="bug-report-title">Bug filed — thank you</h2>
      <p class="hint">
        Your report is now
        <a href={filedURL} target="_blank" rel="noreferrer">issue #{filedNumber}</a> in the project tracker.
      </p>
      <div class="footer">
        <span></span>
        <button type="button" class="submit" onclick={onclose}>Done</button>
      </div>
    {:else}
      <h2 id="bug-report-title">Report a bug</h2>
      <p class="hint">
        What happened, and what did you expect instead? This files an issue in the project's GitHub
        repo — no account needed on your side.
      </p>
      <label class="field">
        <span class="field-label">Title</span>
        <input
          type="text"
          maxlength={BUG_TITLE_MAX}
          placeholder="e.g. cast dialog ignored my X value"
          bind:value={title}
          disabled={submitting}
        />
      </label>
      <label class="field">
        <span class="field-label">
          Details <span class="muted">({description.length}/{BUG_DESC_MAX})</span>
        </span>
        <textarea
          rows="6"
          maxlength={BUG_DESC_MAX}
          placeholder="Steps to reproduce, what you saw, what you expected…"
          bind:value={description}
          disabled={submitting}
        ></textarea>
      </label>
      <label class="context">
        <input type="checkbox" bind:checked={includeContext} disabled={submitting} />
        <span>attach game context <span class="muted">({contextSummary})</span></span>
      </label>
      {#if error}
        <p class="error" role="alert">{error}</p>
      {/if}
      <div class="footer">
        <button type="button" class="cancel" onclick={onclose} disabled={submitting}>
          Cancel
        </button>
        <button
          type="button"
          class="submit"
          onclick={submit}
          disabled={submitting || title.trim().length === 0}
        >
          {submitting ? "Filing…" : "File issue"}
        </button>
      </div>
    {/if}
  </div>
</div>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(4, 8, 16, 0.7);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    animation: fade-in 160ms var(--ease);
  }
  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  .modal {
    background: linear-gradient(180deg, var(--surface) 0%, var(--bg-2) 100%);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius-xl);
    padding: 22px 26px;
    width: min(560px, 92vw);
    max-height: 86vh;
    overflow: auto;
    box-shadow:
      0 30px 80px rgba(0, 0, 0, 0.7),
      0 0 0 1px rgba(0, 0, 0, 0.4),
      inset 0 1px 0 rgba(255, 255, 255, 0.05);
    animation: modal-in 220ms var(--ease);
  }
  @keyframes modal-in {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  h2 {
    margin: 0 0 6px;
    font-size: 18px;
    letter-spacing: -0.01em;
    color: var(--gold);
    text-transform: none;
    font-weight: 700;
  }
  .hint {
    color: var(--fg-muted);
    font-size: 13px;
    line-height: 1.4;
    margin: 0 0 14px;
  }
  .hint a {
    color: var(--accent);
  }
  .field {
    display: block;
    margin-bottom: 12px;
  }
  .field-label {
    display: block;
    font-size: 12px;
    color: var(--fg-muted);
    margin-bottom: 4px;
    letter-spacing: 0.02em;
  }
  input[type="text"],
  textarea {
    width: 100%;
    box-sizing: border-box;
    background: rgba(0, 0, 0, 0.3);
    color: var(--fg);
    border: 1px solid rgba(122, 167, 255, 0.22);
    border-radius: var(--radius);
    padding: 8px 10px;
    font-size: 13px;
    font-family: inherit;
    resize: vertical;
  }
  input[type="text"]:focus,
  textarea:focus {
    outline: none;
    border-color: var(--accent);
  }
  .context {
    display: flex;
    align-items: baseline;
    gap: 8px;
    font-size: 12px;
    color: var(--fg-muted);
    margin: 2px 0 0;
    cursor: pointer;
  }
  .muted {
    color: var(--fg-muted);
    opacity: 0.8;
  }
  .error {
    color: #ff8a8a;
    font-size: 12px;
    margin: 10px 0 0;
  }
  .footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 16px;
    padding-top: 14px;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    gap: 12px;
  }
  .cancel {
    padding: 8px 18px;
    border-radius: 999px;
    background: transparent;
    color: var(--fg-muted);
    border: 1px solid rgba(255, 255, 255, 0.14);
    cursor: pointer;
  }
  .cancel:hover:not(:disabled) {
    color: var(--fg);
    border-color: rgba(255, 255, 255, 0.3);
  }
  .submit {
    padding: 8px 22px;
    border-radius: 999px;
    background: linear-gradient(180deg, #ffe59a 0%, #e6b85f 100%);
    color: #231806;
    border: 1px solid rgba(255, 230, 160, 0.6);
    font-weight: 800;
    letter-spacing: 0.02em;
    cursor: pointer;
    box-shadow:
      0 6px 18px rgba(255, 208, 122, 0.25),
      inset 0 1px 0 rgba(255, 255, 255, 0.4);
  }
  .submit:hover:not(:disabled) {
    filter: brightness(1.04);
  }
  .submit:disabled,
  .cancel:disabled {
    opacity: 0.4;
    cursor: not-allowed;
    box-shadow: none;
  }
</style>
