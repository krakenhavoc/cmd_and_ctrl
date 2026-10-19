<script lang="ts">
  // DeckCard — one saved deck in the Decks grid (ADR 0158 §2).
  //
  // Led by the commander's art crop (partners side by side), framed by
  // an aura in the commander's colour identity, then the name, the
  // commander and card count, a coverage ring and the notes line. The
  // actions are Coverage report, Request missing when the deck has cards
  // to request, and a ⋯ menu with Rename and Delete; Delete still asks.
  //
  // The page owns the requests and the report (it holds one open report
  // at a time); the card owns its menu, its rename field and its
  // confirm, and hands the page the new name or the delete.

  import { L } from "../../labels";
  import PrintingPicker from "./PrintingPicker.svelte";
  import { tick, type Snippet } from "svelte";
  import Icon from "../Icon.svelte";
  import CommanderArt from "./CommanderArt.svelte";
  import CoverageRing from "./CoverageRing.svelte";
  import { auraGradient, deckArt, deckIdentity, deckNotes, deckRing } from "../../deckGrid";
  import { deckSubtitle, sourceHost, type MyDeckInfo } from "../../myDecks";

  interface Props {
    deck: MyDeckInfo;
    reportOpen: boolean;
    canRequest: boolean;
    requesting: boolean;
    /** Another deck's request is in flight. */
    requestLocked: boolean;
    /** A rename or delete is in flight somewhere in the library. */
    busy: boolean;
    onReport: () => void;
    onPrintingSaved: (deck: MyDeckInfo) => void;
    onRequest: () => void;
    /** Resolves to an error message, or "" when the rename landed. */
    onRename: (name: string) => Promise<string>;
    onDelete: () => Promise<void>;
    /** The request outcome and the open report, under the actions. */
    below?: Snippet;
  }

  const {
    deck,
    reportOpen,
    canRequest,
    requesting,
    requestLocked,
    busy,
    onReport,
    onPrintingSaved,
    onRequest,
    onRename,
    onDelete,
    below,
  }: Props = $props();

  const ring = $derived(deckRing(deck));
  const art = $derived(deckArt(deck));
  const aura = $derived(auraGradient(deckIdentity(deck)));
  const notes = $derived(deckNotes(deck));

  // Crops that failed to load; when every one has, the band is plain.
  let failed = $state<string[]>([]);
  const shown = $derived(art.filter((a) => !failed.includes(a.scryfallId)));

  let menuOpen = $state(false);
  let printingsOpen = $state(false);
  let printingButton = $state<HTMLButtonElement | null>(null);

  function closePrintings(): void {
    printingsOpen = false;
    void tick().then(() => printingButton?.focus());
  }
  let renaming = $state(false);
  let renameValue = $state("");
  let renameError = $state("");
  let confirming = $state(false);
  let menuButton = $state<HTMLButtonElement | null>(null);
  let menuEl = $state<HTMLElement | null>(null);

  function closeMenu(refocus = false): void {
    menuOpen = false;
    if (refocus) menuButton?.focus();
  }

  function startRename(): void {
    closeMenu();
    confirming = false;
    renaming = true;
    renameValue = deck.name;
    renameError = "";
  }

  async function saveRename(e: SubmitEvent): Promise<void> {
    e.preventDefault();
    const name = renameValue.trim();
    if (!name || busy) return;
    if (name === deck.name) {
      renaming = false;
      return;
    }
    renameError = await onRename(name);
    if (!renameError) renaming = false;
  }

  function onWindowClick(e: MouseEvent): void {
    if (!menuOpen) return;
    const t = e.target as Node;
    if (menuEl?.contains(t) || menuButton?.contains(t)) return;
    closeMenu();
  }

  function onMenuKey(e: KeyboardEvent): void {
    if (e.key === "Escape") {
      e.stopPropagation();
      closeMenu(true);
    }
  }

  $effect(() => {
    if (menuOpen) menuEl?.querySelector<HTMLButtonElement>("button")?.focus();
  });
</script>

<svelte:window onclick={onWindowClick} />

<li class="cell" class:has-aura={aura !== ""} class:raised={menuOpen} style:--aura={aura || null}>
  <article class="deck-card" aria-label={deck.name}>
    <div class="band" class:plain={shown.length === 0}>
      {#each shown as a (a.scryfallId)}
        <CommanderArt
          scryfallId={a.scryfallId}
          name={a.name}
          onfail={() => (failed = [...failed, a.scryfallId])}
        />
      {/each}
    </div>

    <div class="body">
      <div class="head">
        <div class="ident">
          {#if renaming}
            <form class="rename" onsubmit={saveRename}>
              <input
                type="text"
                bind:value={renameValue}
                maxlength="100"
                aria-label={`new name for ${deck.name}`}
              />
              <button type="submit" class="btn sm primary" disabled={busy || !renameValue.trim()}>
                Save
              </button>
              <button type="button" class="btn sm ghost" onclick={() => (renaming = false)}>
                Cancel
              </button>
            </form>
          {:else}
            <h3 class="name">{deck.name}</h3>
          {/if}
          <p class="sub">{deckSubtitle(deck)}</p>
        </div>
        {#if ring}<CoverageRing {ring} />{/if}
      </div>
      {#if renaming && renameError}
        <p class="err" role="alert">{renameError}</p>
      {/if}

      {#if notes || deck.source_url}
        <p class="notes">
          {notes}{#if notes && deck.source_url}{" · "}{/if}{#if deck.source_url}from <a
              href={deck.source_url}
              target="_blank"
              rel="noopener noreferrer">{sourceHost(deck.source_url)}</a
            >{/if}
        </p>
      {/if}

      {#if confirming}
        <div class="confirm" role="alert">
          <span>Delete {deck.name}?</span>
          <button
            type="button"
            class="btn sm danger"
            disabled={busy}
            onclick={async () => {
              await onDelete();
              confirming = false;
            }}
          >
            Delete
          </button>
          <button type="button" class="btn sm ghost" onclick={() => (confirming = false)}>
            Keep
          </button>
        </div>
      {:else}
        <div class="actions">
          <button
            type="button"
            class="btn sm"
            onclick={() => (printingsOpen = true)}
            bind:this={printingButton}
            aria-expanded={printingsOpen}
            disabled={busy || printingsOpen}>{L.choosePrintings}</button
          >
          {#if deck.coverage}
            <button type="button" class="btn sm" onclick={onReport} aria-expanded={reportOpen}>
              {reportOpen ? "Hide report" : "Coverage report"}
            </button>
          {/if}
          {#if canRequest}
            <button
              type="button"
              class="btn sm"
              onclick={onRequest}
              disabled={requestLocked}
              aria-label={`request missing cards for ${deck.name}`}
            >
              {requesting ? "Requesting…" : "Request missing"}
            </button>
          {/if}
          <span class="spacer"></span>
          <div class="more">
            <button
              type="button"
              class="btn sm ghost icon"
              bind:this={menuButton}
              aria-label={`more for ${deck.name}`}
              aria-haspopup="menu"
              aria-expanded={menuOpen}
              disabled={busy}
              onclick={() => (menuOpen = !menuOpen)}
            >
              <Icon name="more" size={14} />
            </button>
            {#if menuOpen}
              <div
                class="menu"
                role="menu"
                tabindex="-1"
                aria-label={`${deck.name} actions`}
                bind:this={menuEl}
                onkeydown={onMenuKey}
              >
                <button type="button" role="menuitem" onclick={startRename}>Rename</button>
                <button
                  type="button"
                  role="menuitem"
                  class="danger"
                  aria-label={`delete ${deck.name}`}
                  onclick={() => {
                    closeMenu();
                    renaming = false;
                    confirming = true;
                  }}
                >
                  Delete
                </button>
              </div>
            {/if}
          </div>
        </div>
      {/if}

      {#if printingsOpen}
        <PrintingPicker
          {deck}
          onclose={closePrintings}
          onsaved={(updated) => {
            onPrintingSaved(updated);
            closePrintings();
          }}
        />
      {/if}
      {@render below?.()}
    </div>
  </article>
</li>

<style>
  /* The grid cell carries the glow, so it can sit under the card: a
     pseudo-element inside the card would paint over its background. */
  .cell {
    position: relative;
    isolation: isolate;
    min-width: 0;
    list-style: none;
  }
  .cell.raised {
    z-index: 5;
  }
  .cell.has-aura::before {
    content: "";
    position: absolute;
    inset: 2px;
    z-index: -1;
    border-radius: var(--radius-lg);
    background: var(--aura);
    filter: blur(16px);
    opacity: 0.3;
    pointer-events: none;
  }

  .deck-card {
    position: relative;
    height: 100%;
    display: flex;
    flex-direction: column;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-strong);
    background: var(--surface);
    box-sizing: border-box;
  }
  /* The aura's border: the surface on the padding box, the identity
     gradient showing through a transparent 1px border. */
  .has-aura .deck-card {
    border-color: transparent;
    background:
      linear-gradient(var(--surface), var(--surface)) padding-box,
      var(--aura) border-box;
  }

  /* The art, faded into the card at the foot. */
  .band {
    position: relative;
    display: flex;
    gap: 2px;
    height: 140px;
    overflow: hidden;
    border-radius: calc(var(--radius-lg) - 1px) calc(var(--radius-lg) - 1px) 0 0;
    background: var(--surface-sunken);
  }
  .band::after {
    content: "";
    position: absolute;
    inset: 0;
    background: linear-gradient(
      to bottom,
      transparent 45%,
      color-mix(in srgb, var(--surface) 70%, transparent) 78%,
      var(--surface)
    );
    pointer-events: none;
  }
  /* No art: the identity as a wash, never a broken image. */
  .band.plain {
    height: 72px;
    background: linear-gradient(120deg, var(--surface-raised), var(--surface-sunken));
  }
  .has-aura .band.plain {
    background: linear-gradient(to bottom, transparent, var(--surface)), var(--aura);
    background-blend-mode: normal;
  }
  .has-aura .band.plain::before {
    content: "";
    position: absolute;
    inset: 0;
    background: color-mix(in srgb, var(--surface) 45%, transparent);
  }

  .body {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 0 16px 16px;
    margin-top: -26px;
    flex: 1;
    min-width: 0;
  }
  .head {
    display: flex;
    align-items: flex-end;
    gap: 12px;
  }
  .ident {
    flex: 1;
    min-width: 0;
  }
  .name {
    margin: 0;
    font-family: var(--font-display);
    font-size: 17px;
    font-weight: 800;
    letter-spacing: -0.01em;
    color: var(--fg);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .sub {
    margin: 2px 0 0;
    font-size: 12px;
    color: var(--fg-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .notes {
    margin: 0;
    font-size: 12px;
    color: var(--fg-dim);
    overflow-wrap: anywhere;
  }
  .notes a {
    color: var(--fg-muted);
  }
  .notes a:hover {
    color: var(--accent-strong);
  }
  .err {
    margin: 0;
    font-size: 12px;
    color: var(--danger);
  }

  .actions,
  .confirm {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    margin-top: auto;
    padding-top: 4px;
  }
  .confirm {
    font-size: 12.5px;
    color: var(--danger);
  }
  .spacer {
    flex: 1;
  }
  .btn.icon {
    width: 32px;
    padding: 0;
  }
  .btn.danger {
    color: var(--danger);
    border-color: var(--danger-line);
    background: var(--danger-soft);
  }

  .more {
    position: relative;
  }
  .menu {
    position: absolute;
    right: 0;
    top: calc(100% + 6px);
    min-width: 150px;
    display: flex;
    flex-direction: column;
    padding: 4px;
    border-radius: var(--radius);
    border: 1px solid var(--border-strong);
    background: var(--surface-raised);
    box-shadow: var(--shadow);
  }
  .menu button {
    all: unset;
    box-sizing: border-box;
    padding: 8px 10px;
    border-radius: var(--radius-sm);
    font-size: 13px;
    color: var(--fg);
    cursor: pointer;
  }
  .menu button:hover,
  .menu button:focus-visible {
    background: var(--surface-hover);
  }
  .menu button.danger {
    color: var(--danger);
  }

  .rename {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .rename input {
    flex: 1 1 140px;
    min-width: 0;
  }
</style>
