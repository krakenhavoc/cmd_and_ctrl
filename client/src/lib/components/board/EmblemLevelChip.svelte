<script lang="ts">
  // EmblemLevelChip — the Ring on the player panel (ADR 0114 owner
  // decision 1, #2076). It is the emblem chip PlayerIdentity draws for
  // every emblem (CR 114, #623), with the level as a pip ("The Ring ·
  // 3"), and a card that lists all of the emblem's lines: the ones it
  // has gained in full, the rest dimmed AND marked "after the Nth
  // temptation", so the difference never rests on colour alone.
  //
  // How it opens:
  //   - a mouse resting on it, or keyboard focus, shows the card while
  //     it lasts;
  //   - a click, a tap or a long-press (the context menu a touch screen
  //     raises) pins it open, and a second press, Escape or a press
  //     anywhere else closes it.
  // A screen reader does not need it opened at all: the chip's name says
  // the level ("The Ring, tempted 3 times") and its description says
  // every line and whether it is gained (ringEmblem.ts). The card is a
  // visual copy of that description, so it is aria-hidden.
  //
  // The card is portalled to <body> and placed with position: fixed:
  // the panel rail scrolls and clips (PlayerPanel's .rail), and a
  // filtered ancestor (the disabled board's greyscale) would make
  // "fixed" mean "fixed to that ancestor". TokenGroupModal and the
  // hand's drag ghost do the same.

  import type { EmblemView } from "../../protocol";
  import {
    emblemChipName,
    emblemDescription,
    emblemLines,
    pendingNote,
    temptedPhrase,
  } from "../../ringEmblem";
  import Icon from "../Icon.svelte";

  interface Props {
    emblem: EmblemView;
  }
  const { emblem }: Props = $props();

  const uid = $props.id();
  const descID = `${uid}-desc`;

  const level = $derived(emblem.level ?? 0);
  const lines = $derived(emblemLines(emblem));
  const name = $derived(emblemChipName(emblem));
  const description = $derived(emblemDescription(emblem));

  // Open while hovered or focused; pinned by a press.
  let hovered = $state(false);
  let focused = $state(false);
  let pinned = $state(false);
  const open = $derived(hovered || focused || pinned);

  let chip: HTMLButtonElement | undefined = $state();
  // Where the card goes: under the chip when the chip is in the top half
  // of the screen (an opponent's panel), over it in the bottom half (the
  // viewer's own), centred on it and kept 8px inside the viewport.
  let place = $state<{ left: number; top?: number; bottom?: number }>({ left: 8 });
  const CARD_MAX = 300;

  function measure(): void {
    if (!chip || typeof window === "undefined") return;
    const r = chip.getBoundingClientRect();
    const vw = window.innerWidth || document.documentElement.clientWidth || 0;
    const vh = window.innerHeight || document.documentElement.clientHeight || 0;
    const width = Math.min(CARD_MAX, Math.max(0, vw - 16));
    const left = Math.max(8, Math.min(r.left + r.width / 2 - width / 2, vw - width - 8));
    place = r.top > vh / 2 ? { left, bottom: vh - r.top + 6 } : { left, top: r.bottom + 6 };
  }

  $effect(() => {
    if (!open) return;
    measure();
    const onMove = () => measure();
    window.addEventListener("resize", onMove);
    window.addEventListener("scroll", onMove, true);
    return () => {
      window.removeEventListener("resize", onMove);
      window.removeEventListener("scroll", onMove, true);
    };
  });

  function portal(node: HTMLElement): { destroy(): void } {
    document.body.appendChild(node);
    return {
      destroy() {
        node.remove();
      },
    };
  }

  function onPointerEnter(e: PointerEvent): void {
    // A touch "hover" is the start of a tap, which pins it instead.
    if (e.pointerType === "mouse" || e.pointerType === "pen") hovered = true;
  }
  // Keyboard focus opens it; the focus a mouse click leaves behind
  // does not, or the click that closes a pinned card would leave it
  // open.
  function onFocus(): void {
    let keyboard = true;
    try {
      keyboard = chip?.matches(":focus-visible") ?? true;
    } catch {
      // A selector engine without :focus-visible: treat as keyboard.
    }
    focused = keyboard;
  }
  function onClick(e: MouseEvent): void {
    e.stopPropagation();
    pinned = !pinned;
  }
  function onContextMenu(e: MouseEvent): void {
    // A long-press on a touch screen raises the context menu: open the
    // card instead of the browser's menu.
    e.preventDefault();
    e.stopPropagation();
    pinned = true;
  }
  function onKeydown(e: KeyboardEvent): void {
    if (e.key !== "Escape" || !open) return;
    e.stopPropagation();
    pinned = false;
    hovered = false;
    focused = false;
  }
  function onWindowPointerDown(e: PointerEvent): void {
    if (pinned && chip && !chip.contains(e.target as Node)) pinned = false;
  }
</script>

<svelte:window onpointerdown={onWindowPointerDown} />

<button
  type="button"
  class="emblem levelled"
  bind:this={chip}
  aria-label={name}
  aria-describedby={descID}
  aria-expanded={open}
  data-level={level}
  onpointerenter={onPointerEnter}
  onpointerleave={() => (hovered = false)}
  onfocus={onFocus}
  onblur={() => {
    focused = false;
  }}
  onclick={onClick}
  oncontextmenu={onContextMenu}
  onkeydown={onKeydown}
>
  <Icon name="ring" size={12} />
  <span class="label">{emblem.label}</span>
  <span class="sep" aria-hidden="true">·</span>
  <span class="level" aria-hidden="true">{level}</span>
</button>
<span id={descID} class="sr-only">{description}</span>

{#if open}
  <div
    class="emblem-card"
    aria-hidden="true"
    use:portal
    style:left="{place.left}px"
    style:top={place.top !== undefined ? `${place.top}px` : null}
    style:bottom={place.bottom !== undefined ? `${place.bottom}px` : null}
    style:width="min({CARD_MAX}px, calc(100vw - 16px))"
  >
    <div class="head">
      <Icon name="ring" size={14} />
      <span class="title">{emblem.label}</span>
      <span class="count">{temptedPhrase(level)}</span>
    </div>
    <ol class="lines">
      {#each lines as line (line.at)}
        <li class="line" class:gained={line.gained} class:pending={!line.gained}>
          <span class="mark">
            {#if line.gained}<Icon name="check" size={11} strokeWidth={2.25} />{:else}{line.at}{/if}
          </span>
          <span class="body">
            {#if !line.gained}<span class="when">{pendingNote(line.at)}</span>{/if}
            <span class="text">{line.text}</span>
          </span>
        </li>
      {/each}
    </ol>
  </div>
{/if}

<style>
  /* The chip keeps PlayerIdentity's emblem look (#623) — gold on a gold
     wash, a pill — and adds the level pip. It is a button now, so it
     resets the browser's button chrome and draws a focus ring. */
  .emblem {
    font: inherit;
    font-size: 11px;
    line-height: 1;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 3px 2px 7px;
    border-radius: 999px;
    background: var(--accent-soft);
    border: 1px solid rgba(255, 208, 122, 0.45);
    color: var(--accent);
    max-width: 100%;
    min-width: 0;
    white-space: nowrap;
    cursor: pointer;
    box-shadow: none;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
  }
  .emblem:hover,
  .emblem[aria-expanded="true"] {
    background: color-mix(in srgb, var(--accent) 24%, transparent);
  }
  .emblem:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }
  .label {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .sep {
    opacity: 0.7;
  }
  /* The level pip: the number on a solid gold disc, so it reads as a
     count at a glance and stays legible at 11px. */
  .level {
    display: inline-grid;
    place-items: center;
    min-width: 15px;
    height: 15px;
    padding: 0 3px;
    box-sizing: border-box;
    border-radius: 999px;
    background: var(--accent);
    color: var(--accent-fg, #1c1503);
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
  }
  :global(:root[data-theme="light"]) .emblem {
    color: var(--accent-strong);
    border-color: color-mix(in srgb, var(--accent) 55%, transparent);
    background: var(--accent-soft);
  }
  :global(:root[data-theme="light"]) .level {
    background: var(--accent-strong);
    color: #fff;
  }

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  /* The card. Portalled to <body>, so it is placed on the viewport. */
  .emblem-card {
    position: fixed;
    z-index: 1000;
    box-sizing: border-box;
    padding: 8px 10px 9px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface-raised);
    color: var(--fg);
    box-shadow: var(--shadow);
    font-family: var(--font-ui);
    font-size: 12px;
    line-height: 1.35;
    pointer-events: none;
  }
  .head {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--accent);
    margin-bottom: 6px;
  }
  :global(:root[data-theme="light"]) .head {
    color: var(--accent-strong);
  }
  .title {
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 13px;
  }
  .count {
    margin-left: auto;
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--fg-muted);
  }
  .lines {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .line {
    display: flex;
    align-items: flex-start;
    gap: 7px;
  }
  /* Gained: a filled check disc. Not yet: a hollow disc with the count
     it is gained at — a different shape as well as a dimmer line. */
  .mark {
    flex: 0 0 auto;
    display: inline-grid;
    place-items: center;
    width: 16px;
    height: 16px;
    margin-top: 1px;
    box-sizing: border-box;
    border-radius: 50%;
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 800;
  }
  .gained .mark {
    background: var(--accent);
    color: var(--accent-fg, #1c1503);
  }
  :global(:root[data-theme="light"]) .gained .mark {
    background: var(--accent-strong);
    color: #fff;
  }
  .pending .mark {
    border: 1px dashed var(--fg-dim);
    color: var(--fg-muted);
  }
  .body {
    min-width: 0;
  }
  .pending .text {
    color: var(--fg-muted);
    opacity: 0.75;
  }
  .when {
    display: block;
    font-family: var(--font-mono);
    font-size: 9.5px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--fg-muted);
  }
  :global(:root[data-theme="high-contrast"]) .pending .text {
    opacity: 1;
  }
</style>
