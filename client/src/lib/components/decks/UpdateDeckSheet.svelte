<script lang="ts">
  // UpdateDeckSheet — update a saved deck from its own link, a new link or a
  // pasted list (#2036, ADR 0095 amendment 2026-10-19).
  //
  // Two steps, because an update replaces the list: choose where the new
  // list comes from and preview it (nothing is written), then read the diff
  // and the fresh coverage and Save or Cancel. The save commits the list that
  // was previewed, not a second fetch. A saved deck keeps its id, name and
  // the printing choices of cards that carry over; a table already seated
  // with it keeps the list it was seated with.

  import { L } from "../../labels";
  import { confirmDeckUpdate, previewDeckUpdate } from "../../api";
  import { LobbyApiError } from "../../session";
  import { PASTE_LIST_HINT } from "../../deckcheck";
  import type { MyDeckInfo } from "../../myDecks";
  import {
    commanderLine,
    coverageLine,
    defaultChoice,
    diffCardLabel,
    diffChangeLabel,
    diffSummary,
    missingCount,
    requestOptionText,
    updateChoices,
    updateInput,
    type DeckUpdateChoice,
    type DeckUpdateConfirm,
    type DeckUpdatePreview,
  } from "../../deckUpdate";

  const {
    deck,
    onclose,
    onupdated,
  }: { deck: MyDeckInfo; onclose: () => void; onupdated: (res: DeckUpdateConfirm) => void } =
    $props();

  const choices = $derived(updateChoices(deck));
  // The sheet opens on the choice the deck's own state suggests, and the
  // deck is not swapped under an open sheet, so reading it once is right.
  // svelte-ignore state_referenced_locally
  let choice = $state<DeckUpdateChoice>(defaultChoice(deck));
  let url = $state("");
  let list = $state("");
  let preview = $state<DeckUpdatePreview | null>(null);
  let request = $state(false);
  let busy = $state(false);
  let error = $state("");
  let pasteHint = $state(false);

  const input = $derived(updateInput(choice, url, list));

  function pick(next: DeckUpdateChoice): void {
    if (busy || next === choice) return;
    choice = next;
    preview = null;
    error = "";
    pasteHint = false;
  }

  function fail(err: unknown, fallback: string): void {
    error = err instanceof Error && err.message ? err.message : fallback;
    pasteHint =
      err instanceof LobbyApiError &&
      (err.body as { hint?: string } | undefined)?.hint === PASTE_LIST_HINT;
  }

  async function runPreview(): Promise<void> {
    if (!input || busy) return;
    busy = true;
    error = "";
    pasteHint = false;
    try {
      const p = await previewDeckUpdate(deck.id, input);
      preview = p;
      // Ask for the new cards by default: that is why a player updates
      // a deck whose missing cards they already follow.
      request = p.request.available && p.request.new_cards.length > 0;
    } catch (err) {
      preview = null;
      fail(err, "Could not read that list. Try again.");
    } finally {
      busy = false;
    }
  }

  async function save(): Promise<void> {
    if (!preview || busy) return;
    busy = true;
    error = "";
    try {
      const done = await confirmDeckUpdate(deck.id, preview.preview_id, request);
      onupdated(done);
    } catch (err) {
      fail(err, "Could not save the update. Preview it again.");
      // A used or expired preview cannot be confirmed: send them back to it.
      if (err instanceof LobbyApiError && (err.status === 410 || err.status === 409))
        preview = null;
    } finally {
      busy = false;
    }
  }

  function backToInput(): void {
    preview = null;
    error = "";
  }

  const before = $derived(preview ? coverageLine(preview.coverage_before) : "");
  const after = $derived(preview ? coverageLine(preview.coverage_after) : "");
  const issues = $derived(preview ? preview.report.violations : []);
</script>

<section class="sheet" aria-label={L.deckUpdateSheet}>
  <h4>{L.updateDeck}</h4>
  <p class="note">
    Replace this deck's list. It keeps its name, and the art you chose for cards that stay. Tables
    already seated with it keep the list they sat down with.
  </p>

  <div class="tabs" role="group" aria-label={L.deckUpdateChoices}>
    {#each choices as c (c.key)}
      <button
        type="button"
        class="tab"
        class:on={choice === c.key}
        aria-pressed={choice === c.key}
        disabled={busy}
        onclick={() => pick(c.key)}>{c.label}</button
      >
    {/each}
  </div>

  {#if !preview}
    {#if choice === "relink"}
      <p class="note">Read the deck again from {deck.source_url} and show what changed.</p>
    {:else if choice === "url"}
      <label class="field">
        <span>Deck link</span>
        <input
          type="url"
          name="update-deck-link"
          bind:value={url}
          placeholder="https://archidekt.com/decks/…"
          autocomplete="off"
          disabled={busy}
        />
      </label>
    {:else}
      <label class="field">
        <span>Decklist</span>
        <textarea
          name="update-deck-list"
          bind:value={list}
          rows="8"
          placeholder={"1 Sol Ring\n1 Arcane Signet\n…"}
          disabled={busy}
        ></textarea>
      </label>
      {#if deck.source_url}
        <p class="note">A pasted list replaces this deck's link: it will no longer re-pull.</p>
      {/if}
    {/if}
    <div class="actions">
      <button type="button" class="btn sm primary" onclick={runPreview} disabled={busy || !input}
        >{busy ? "Reading…" : L.previewUpdate}</button
      >
      <button type="button" class="btn sm ghost" onclick={onclose} disabled={busy}
        >{L.cancelUpdate}</button
      >
    </div>
  {:else}
    <div class="preview" aria-busy={busy}>
      <p class="summary" role="status">
        {diffSummary(preview.diff, preview.no_change)}{#if preview.diff.unchanged > 0}
          · {preview.diff.unchanged} cards unchanged{/if}
      </p>

      {#if preview.diff.commander_changed}
        <p class="cmdr">Commander: {commanderLine(preview.diff)}</p>
      {/if}

      {#if preview.diff.added.length > 0}
        <details open>
          <summary>Added ({preview.diff.added.length})</summary>
          <ul>
            {#each preview.diff.added as c (c.name + (c.sideboard ? "·sb" : ""))}
              <li>{diffCardLabel(c)}</li>
            {/each}
          </ul>
        </details>
      {/if}
      {#if preview.diff.removed.length > 0}
        <details open>
          <summary>Removed ({preview.diff.removed.length})</summary>
          <ul>
            {#each preview.diff.removed as c (c.name + (c.sideboard ? "·sb" : ""))}
              <li>{diffCardLabel(c)}</li>
            {/each}
          </ul>
        </details>
      {/if}
      {#if preview.diff.changed.length > 0}
        <details>
          <summary>Changed copies ({preview.diff.changed.length})</summary>
          <ul>
            {#each preview.diff.changed as c (c.name + (c.sideboard ? "·sb" : ""))}
              <li>{diffChangeLabel(c)}</li>
            {/each}
          </ul>
        </details>
      {/if}

      <div class="coverage">
        {#if before}<p>Now: {before}</p>{/if}
        {#if after}<p>After: {after}</p>{/if}
        {#if preview.coverage_after}
          <p class="muted">
            Cards the engine does not fully run yet: {missingCount(
              preview.coverage_after,
            )}{#if preview.coverage_before && missingCount(preview.coverage_before) !== missingCount(preview.coverage_after)}
              (was {missingCount(preview.coverage_before)}){/if}
          </p>
        {/if}
        {#if preview.report.unknown.length > 0}
          <p class="warn">
            Not found in the card index: {preview.report.unknown.slice(0, 8).join(", ")}{preview
              .report.unknown.length > 8
              ? ` and ${preview.report.unknown.length - 8} more`
              : ""}
          </p>
        {/if}
        {#if issues.length > 0}
          <details>
            <summary>Commander rules: {issues.length} to look at</summary>
            <ul>
              {#each issues.slice(0, 8) as v, i (v.code + i)}
                <li>{v.message}</li>
              {/each}
            </ul>
          </details>
        {/if}
      </div>

      {#if preview.printings_kept > 0}
        <p class="note">
          {preview.printings_kept} card row{preview.printings_kept === 1 ? "" : "s"} keep the art you
          chose.
        </p>
      {/if}
      {#if preview.link_cleared}
        <p class="note">This deck's link will be removed.</p>
      {:else if preview.source_url && preview.source_url !== deck.source_url}
        <p class="note">This deck will now come from {preview.source_url}.</p>
      {/if}
      {#if preview.live_tables > 0}
        <p class="note">
          {preview.live_tables === 1 ? "A table is" : `${preview.live_tables} tables are`} seated with
          this deck. They keep their list; the update applies from the next game.
        </p>
      {/if}

      {#if preview.request.new_cards.length > 0}
        {#if preview.request.available}
          <label class="check">
            <input type="checkbox" bind:checked={request} disabled={busy} />
            <span>{requestOptionText(preview.request)}</span>
          </label>
          <p class="muted">
            {preview.request.new_cards.slice(0, 6).join(", ")}{preview.request.new_cards.length > 6
              ? "…"
              : ""}
          </p>
        {:else if preview.request.reason}
          <p class="muted">{preview.request.reason}</p>
        {/if}
      {/if}
    </div>

    <div class="actions">
      <button type="button" class="btn sm primary" onclick={save} disabled={busy}
        >{busy ? "Saving…" : L.saveUpdate}</button
      >
      <button type="button" class="btn sm" onclick={backToInput} disabled={busy}
        >Change source</button
      >
      <button type="button" class="btn sm ghost" onclick={onclose} disabled={busy}
        >{L.cancelUpdate}</button
      >
    </div>
  {/if}

  {#if error}
    <p class="error" role="alert">{error}</p>
    {#if pasteHint}
      <button type="button" class="btn sm" onclick={() => pick("list")}
        >Paste the list instead</button
      >
    {/if}
  {/if}
</section>

<style>
  .sheet {
    border-top: 1px solid var(--border-strong);
    padding-top: 14px;
    min-width: 0;
  }
  h4 {
    margin: 0 0 6px;
    color: var(--fg);
    font-size: 15px;
  }
  .note,
  .muted,
  .cmdr,
  .coverage p {
    color: var(--fg-muted);
    font-size: 12px;
    line-height: 1.5;
    margin: 6px 0;
    overflow-wrap: anywhere;
  }
  .summary {
    margin: 8px 0 4px;
    font-size: 13px;
    font-weight: 700;
    color: var(--fg);
  }
  .warn {
    color: var(--danger);
  }
  .tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin: 10px 0;
  }
  .tab {
    all: unset;
    box-sizing: border-box;
    padding: 6px 10px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-strong);
    background: var(--surface-sunken);
    color: var(--fg-muted);
    font-size: 12.5px;
    cursor: pointer;
    max-width: 100%;
    overflow-wrap: anywhere;
  }
  .tab:hover {
    background: var(--surface-hover);
  }
  .tab.on {
    border-color: var(--accent);
    background: var(--accent-soft);
    color: var(--fg);
  }
  .tab:focus-visible,
  input:focus-visible,
  textarea:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }
  .field {
    display: grid;
    gap: 5px;
    margin: 10px 0;
    color: var(--fg-muted);
    font-size: 12px;
  }
  input[type="url"],
  textarea {
    width: 100%;
    min-width: 0;
    box-sizing: border-box;
    background: var(--surface-sunken);
    color: var(--fg);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-sm);
    padding: 8px;
    font: inherit;
  }
  textarea {
    resize: vertical;
    min-height: 120px;
    font-family: var(--font-mono, monospace);
    font-size: 12px;
  }
  details {
    margin: 6px 0;
    font-size: 12.5px;
    color: var(--fg);
  }
  summary {
    cursor: pointer;
    color: var(--fg-muted);
  }
  ul {
    margin: 4px 0 0;
    padding-left: 18px;
    max-height: 200px;
    overflow-y: auto;
    overscroll-behavior: contain;
  }
  .check {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin: 10px 0 2px;
    font-size: 12.5px;
    color: var(--fg);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 12px;
  }
  .error {
    color: var(--danger);
    font-size: 12px;
    overflow-wrap: anywhere;
  }
</style>
