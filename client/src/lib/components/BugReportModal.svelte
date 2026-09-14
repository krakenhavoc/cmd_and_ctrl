<script lang="ts">
  // BugReportModal — the in-app report form (ADR 0017).
  //
  // Collects a KIND, a title, a description, screenshots, and two
  // kinds of automatic context: the game coordinates (game ID, turn,
  // phase/step, seq, connection) and the client log — the last 200
  // protocol frames and JS errors this browser saw. POSTs the lot to
  // /bugreport, where the server renders the issue body and files it
  // in the project repo.
  //
  // The kind picker is first because it changes everything below it:
  // the wording, and what the report attaches. It is an explicit
  // choice rather than something inferred from the prose (ADR 0017
  // §8) — one tap, never wrong, and no round trip in front of someone
  // trying to report a broken game.
  //
  // Both automatic attachments are opt-out and both are shown before
  // sending, in full, in a disclosure. That is deliberate: a report
  // that quietly ships a log is a report people stop filing once they
  // notice. The log holds nothing the reporter hasn't already seen —
  // including their own chat — but "trust me" is not the same as
  // "here it is".
  //
  // Unlike the server-driven prompt modals (discard / choice), this one
  // is user-opened and user-dismissable: Cancel closes it and nothing
  // is sent.

  import { onDestroy, untrack } from "svelte";
  import { submitBugReport } from "../api";
  import {
    BUG_DESC_MAX,
    BUG_KINDS,
    BUG_MAX_IMAGES,
    BUG_MAX_TOTAL_IMAGE_BYTES,
    BUG_TITLE_MAX,
    BUG_IMAGE_TYPES,
    DEFAULT_BUG_KIND,
    acceptableImages,
    buildBugContext,
    bugKindSpec,
    collectBugLog,
    describeBugAttachments,
    formatBytes,
    joinPhrases,
    validateAttachments,
    validateBugReport,
    type BugLogEntry,
    type BugReportKind,
  } from "../bugReport";
  import { recentClientErrors } from "../clientErrors";
  import type { LogEntry } from "../ws";
  import type { GameView } from "../protocol";

  interface Props {
    gameID: string;
    view: GameView | null;
    seq: number;
    connection: string;
    // wsLog is GameClient's protocol ring buffer, passed in rather than
    // subscribed to here so this component stays renderable in
    // isolation and the modal can't outlive the client it read from.
    wsLog: LogEntry[];
    // attachments is false when the server has a GitHub token but
    // nowhere to store files; the picker hides rather than offering an
    // upload that would 503.
    attachments: boolean;
    onclose: () => void;
  }

  const { gameID, view, seq, connection, wsLog, attachments, onclose }: Props = $props();

  let kind = $state<BugReportKind>(DEFAULT_BUG_KIND);
  let title = $state("");
  let description = $state("");
  let includeContext = $state(true);
  let includeLog = $state(true);
  let showLog = $state(false);
  let images = $state<File[]>([]);
  let submitting = $state(false);
  let error = $state<string | null>(null);
  let filedURL = $state<string | null>(null);
  let filedNumber = $state<number | null>(null);
  let filedLabel = $state<string | null>(null);
  let fileInput = $state<HTMLInputElement | null>(null);

  // The kind drives the copy and the attachments. The server keeps
  // its own copy of this table and re-derives everything from the
  // kind name — this one exists so the form can tell the reporter
  // what it is about to send.
  const spec = $derived(bugKindSpec(kind));
  const autoAttached = $derived(describeBugAttachments(spec, gameID.length > 0));

  // Snapshotted once when the modal opens, not recomputed as frames
  // keep arriving: the log the reporter reviews in the disclosure has
  // to be the log that gets sent, and a live buffer would mean those
  // two differ by however long they spent typing.
  // untrack makes the one-shot read explicit: wsLog keeps growing while
  // the modal is open, and subscribing here would let the reviewed log
  // and the submitted log drift apart.
  const capturedLog: BugLogEntry[] = untrack(() => collectBugLog(wsLog, recentClientErrors()));

  // Context is snapshotted at submit time, but summarised live so the
  // reporter can see what the checkbox will attach.
  const contextSummary = $derived.by(() => {
    const ctx = buildBugContext(gameID, view, seq, connection);
    const bits = [`game ${gameID.slice(0, 8)}`];
    if (ctx.turn) bits.push(`turn ${ctx.turn}, ${ctx.phase}/${ctx.step}`);
    if (ctx.seq) bits.push(`seq ${ctx.seq}`);
    if (ctx.connection) bits.push(`ws ${ctx.connection}`);
    return bits.join(" \u00b7 ");
  });

  const totalImageBytes = $derived(images.reduce((n, f) => n + f.size, 0));

  // Object URLs for the thumbnails, revoked when the set changes or the
  // modal closes — a leaked blob URL pins the whole image in memory for
  // the life of the tab.
  let previews = $state<string[]>([]);
  $effect(() => {
    const urls = images.map((f) => URL.createObjectURL(f));
    previews = urls;
    return () => urls.forEach((u) => URL.revokeObjectURL(u));
  });
  onDestroy(() => previews.forEach((u) => URL.revokeObjectURL(u)));

  // addFiles vets and appends. Non-image entries are dropped silently
  // when they arrived alongside images (a clipboard paste carries
  // text/html parts too) but reported when they were the whole
  // selection — otherwise picking a PDF looks like the button is broken.
  function addFiles(incoming: File[]): void {
    if (incoming.length === 0) return;
    const usable = acceptableImages(incoming) as File[];
    if (usable.length === 0) {
      error = "attachments must be PNG, JPEG, GIF, or WebP images";
      return;
    }
    const next = [...images, ...usable];
    const problem = validateAttachments(next);
    if (problem) {
      error = problem;
      return;
    }
    error = null;
    images = next;
  }

  function onPick(ev: Event): void {
    const input = ev.currentTarget as HTMLInputElement;
    addFiles(Array.from(input.files ?? []));
    // Clear the input so re-picking the same file fires change again.
    input.value = "";
  }

  // Paste is the fast path: screenshot to clipboard, Ctrl+V into the
  // modal. Bound on the modal container rather than the textarea so it
  // works wherever the caret happens to be.
  function onPaste(ev: ClipboardEvent): void {
    const files = Array.from(ev.clipboardData?.files ?? []);
    if (files.length === 0) return;
    ev.preventDefault();
    addFiles(files);
  }

  function onDrop(ev: DragEvent): void {
    const files = Array.from(ev.dataTransfer?.files ?? []);
    if (files.length === 0) return;
    ev.preventDefault();
    addFiles(files);
  }

  // Without preventDefault on dragover the browser refuses the drop and
  // navigates to the file instead. Narrowed to file drags so an
  // in-app drag (a card, say) is untouched.
  function onDragOver(ev: DragEvent): void {
    if (ev.dataTransfer?.types.includes("Files")) ev.preventDefault();
  }

  function removeImage(i: number): void {
    images = images.filter((_, j) => j !== i);
    error = null;
  }

  async function submit(): Promise<void> {
    const problem = validateBugReport(title, description) ?? validateAttachments(images);
    if (problem) {
      error = problem;
      return;
    }
    submitting = true;
    error = null;
    try {
      const res = await submitBugReport({
        title: title.trim(),
        kind,
        description,
        context: includeContext ? buildBugContext(gameID, view, seq, connection) : undefined,
        // The server drops a log the kind doesn't take; not sending
        // one in the first place keeps "what the form showed" and
        // "what left the browser" the same thing.
        log: spec.log && includeLog ? capturedLog : undefined,
        images,
      });
      filedURL = res.url;
      filedNumber = res.number;
      filedLabel = res.label ?? null;
    } catch (e) {
      error = e instanceof Error ? e.message : "filing the issue failed";
    } finally {
      submitting = false;
    }
  }

  // Same fixed-width shape the server renders, so what the reporter
  // reviews here matches what lands in the issue.
  function logLine(e: BugLogEntry): string {
    const t = e.at > 0 ? new Date(e.at).toISOString().slice(11, 23) : "--:--:--.---";
    return `${t}  ${e.kind.padEnd(8)}  ${e.text}`;
  }
</script>

<!-- Paste and drop listen at the window rather than on an element:
     the modal is modal, so nothing else should be receiving either, and
     hanging drop handlers off the dialog div would demand a tabindex on
     a container that has no business taking focus. Both handlers no-op
     unless the event actually carries files, so pasting text into the
     textarea behaves normally. -->
<svelte:window onpaste={onPaste} ondrop={onDrop} ondragover={onDragOver} />

<div class="backdrop" role="dialog" aria-modal="true" aria-labelledby="bug-report-title">
  <div class="modal">
    {#if filedURL}
      <h2 id="bug-report-title">Sent — thank you</h2>
      <p class="hint">
        Your {spec.kind === "bug" ? "report" : spec.kind} is now
        <a href={filedURL} target="_blank" rel="noreferrer">issue #{filedNumber}</a>
        in the project tracker{#if filedLabel}, labelled <code>{filedLabel}</code>{/if}.
      </p>
      <div class="footer">
        <span></span>
        <button type="button" class="submit" onclick={onclose}>Done</button>
      </div>
    {:else}
      <h2 id="bug-report-title">{spec.heading}</h2>

      <!-- The kind comes first: it decides the wording of everything
           below and what the report attaches. Real radios rather than
           styled buttons, so arrow keys and screen readers work. -->
      <fieldset class="kinds" disabled={submitting}>
        <legend class="field-label">What kind of report is this?</legend>
        <!-- The chips live in their own flex row rather than making
             the fieldset itself flex: a <legend> inside a flex
             container is laid out differently across browsers. -->
        <div class="kind-row">
          {#each BUG_KINDS as k (k.kind)}
            <label class="kind" class:picked={kind === k.kind}>
              <input type="radio" name="bug-kind" value={k.kind} bind:group={kind} />
              <span>{k.chip}</span>
            </label>
          {/each}
        </div>
      </fieldset>

      <!-- What the choice costs, in one line. Switching from broken
           to missing quietly stops sending a replay, and the reporter
           should watch that happen rather than find out in an issue. -->
      <p class="attaches">
        Filed as <code>{spec.label}</code>.
        {#if autoAttached.length > 0}
          Sends {joinPhrases(autoAttached)} along with your screenshots.
        {:else}
          Sends nothing but your words and any screenshots you attach.
        {/if}
      </p>

      <p class="hint">{spec.hint}</p>
      <label class="field">
        <span class="field-label">Title</span>
        <input
          type="text"
          maxlength={BUG_TITLE_MAX}
          placeholder={spec.titlePlaceholder}
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
          placeholder={spec.detailsPlaceholder}
          bind:value={description}
          disabled={submitting}
        ></textarea>
      </label>

      {#if attachments}
        <div class="field">
          <span class="field-label">
            Screenshots
            <span class="muted">
              — paste, drop, or pick · {images.length}/{BUG_MAX_IMAGES}
              {#if images.length > 0}
                · {formatBytes(totalImageBytes)} of {formatBytes(BUG_MAX_TOTAL_IMAGE_BYTES)}
              {/if}
            </span>
          </span>
          {#if images.length > 0}
            <ul class="shots">
              {#each images as file, i (file.name + file.size + i)}
                <li>
                  <img src={previews[i]} alt={file.name} />
                  <button
                    type="button"
                    class="remove"
                    title="remove {file.name}"
                    aria-label="remove {file.name}"
                    onclick={() => removeImage(i)}
                    disabled={submitting}>×</button
                  >
                  <span class="shot-size">{formatBytes(file.size)}</span>
                </li>
              {/each}
            </ul>
          {/if}
          <div class="pick-row">
            <button
              type="button"
              class="pick"
              onclick={() => fileInput?.click()}
              disabled={submitting || images.length >= BUG_MAX_IMAGES}
            >
              Add image…
            </button>
            <span class="muted small">or press Ctrl/⌘+V to paste one</span>
          </div>
          <input
            bind:this={fileInput}
            type="file"
            accept={BUG_IMAGE_TYPES.join(",")}
            multiple
            hidden
            onchange={onPick}
          />
        </div>
      {/if}

      <label class="context">
        <input type="checkbox" bind:checked={includeContext} disabled={submitting} />
        <span>attach game context <span class="muted">({contextSummary})</span></span>
      </label>

      {#if spec.log && capturedLog.length > 0}
        <label class="context">
          <input type="checkbox" bind:checked={includeLog} disabled={submitting} />
          <span>
            attach recent activity log
            <span class="muted">({capturedLog.length} entries)</span>
            <button
              type="button"
              class="peek"
              onclick={(e) => {
                e.preventDefault();
                showLog = !showLog;
              }}>{showLog ? "hide" : "show"}</button
            >
          </span>
        </label>
        {#if showLog}
          <pre class="logpeek">{capturedLog.map(logLine).join("\n")}</pre>
          <p class="muted small">
            Frames this browser sent and received, plus any JavaScript errors. No hidden game state
            — nothing here that wasn't already on your screen.
          </p>
        {/if}
      {/if}

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
          {submitting ? "Filing…" : spec.cta}
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
  .kinds {
    border: none;
    margin: 10px 0 8px;
    padding: 0;
    min-inline-size: 0;
  }
  .kinds legend {
    padding: 0;
    margin-bottom: 6px;
  }
  .kind-row {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .kind {
    display: inline-flex;
    align-items: center;
    padding: 6px 14px;
    border-radius: 999px;
    border: 1px solid rgba(122, 167, 255, 0.28);
    background: rgba(122, 167, 255, 0.08);
    color: var(--fg-muted);
    font-size: 12px;
    cursor: pointer;
    white-space: nowrap;
    transition:
      border-color 120ms var(--ease),
      color 120ms var(--ease);
  }
  /* The radio itself is the a11y surface and the keyboard target; the
     chip is the paint. Hidden with a clip rather than display:none so
     it keeps taking focus and arrow keys still move between kinds. */
  .kind input {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
    pointer-events: none;
  }
  .kind:hover {
    border-color: var(--accent);
    color: var(--fg);
  }
  .kind.picked {
    border-color: var(--gold);
    color: var(--fg);
    background: rgba(255, 213, 128, 0.14);
    font-weight: 700;
  }
  .kind:focus-within {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .kinds:disabled .kind {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .attaches {
    font-size: 11.5px;
    line-height: 1.45;
    color: var(--fg-muted);
    margin: 0 0 12px;
  }
  .attaches code,
  .hint code {
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 11px;
    padding: 1px 5px;
    border-radius: 4px;
    background: rgba(122, 167, 255, 0.14);
    color: var(--fg);
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

  .shots {
    list-style: none;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin: 0 0 8px;
    padding: 0;
  }
  .shots li {
    position: relative;
    width: 96px;
  }
  .shots img {
    width: 96px;
    height: 64px;
    object-fit: cover;
    display: block;
    border-radius: var(--radius);
    border: 1px solid rgba(122, 167, 255, 0.28);
    background: rgba(0, 0, 0, 0.4);
  }
  .shot-size {
    display: block;
    font-size: 10px;
    color: var(--fg-muted);
    text-align: center;
    margin-top: 2px;
  }
  .remove {
    position: absolute;
    top: -6px;
    right: -6px;
    width: 20px;
    height: 20px;
    line-height: 1;
    border-radius: 999px;
    background: rgba(12, 16, 26, 0.95);
    color: var(--fg);
    border: 1px solid rgba(255, 255, 255, 0.28);
    cursor: pointer;
    font-size: 13px;
    padding: 0;
  }
  .remove:hover:not(:disabled) {
    color: #ff8a8a;
    border-color: #ff8a8a;
  }
  .pick-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .pick {
    padding: 5px 14px;
    border-radius: 999px;
    background: rgba(122, 167, 255, 0.12);
    color: var(--fg);
    border: 1px solid rgba(122, 167, 255, 0.32);
    font-size: 12px;
    cursor: pointer;
  }
  .pick:hover:not(:disabled) {
    border-color: var(--accent);
  }
  .pick:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .peek {
    background: none;
    border: none;
    padding: 0 0 0 6px;
    color: var(--accent);
    font-size: 11px;
    cursor: pointer;
    text-decoration: underline;
  }
  .logpeek {
    max-height: 200px;
    overflow: auto;
    margin: 6px 0 0;
    padding: 8px 10px;
    background: rgba(0, 0, 0, 0.45);
    border: 1px solid rgba(122, 167, 255, 0.18);
    border-radius: var(--radius);
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 10.5px;
    line-height: 1.45;
    color: var(--fg-muted);
    white-space: pre;
  }
  .small {
    font-size: 11px;
    line-height: 1.4;
  }
</style>
