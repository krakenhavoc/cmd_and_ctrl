<script lang="ts">
  // AbilityKindChips — #2219: an art tile's ability chips. The art crop
  // shows no rules text, so the tile says what the card can do, by
  // kind: ⚡ triggered, ◆ static, ↻ activated, each with how many. The
  // list is the server's (`ability_rows`, game.AbilityRowsOf): keyword
  // abilities are on the keyword chips beside these and are not in it,
  // and an uncatalogued card has none and keeps its `manual` mark. The
  // client never reads oracle text for them.
  //
  // How the ↻ chip relates to ADR 0105's bolt pip: the chip is the
  // inventory — "this permanent has N activated abilities", on every
  // viewer's copy, all the time — and the pip is the state — "you can
  // activate N of them now", for the controller, while they owe a
  // decision, as a button that opens the popover. They answer different
  // questions and both can show.
  //
  // How a chip opens its list, as EmblemLevelChip's card does:
  //   - a mouse resting on it, or keyboard focus, shows it while it lasts;
  //   - a click, a tap or a long-press pins it open; a second press,
  //     Escape or a press anywhere else closes it.
  // The chip's name says the count and the kind ("2 triggered
  // abilities") and its description is every label, so a screen reader
  // needs nothing opened. The list is portalled to <body> with
  // position: fixed, because the tile clips (overflow: hidden) and is
  // its own stacking context.

  import { ABILITY_CHIP_TITLE, abilityChipDescription, abilityChips } from "../../abilityChips";
  import type { IconName } from "../../icons";
  import { L } from "../../labels";
  import type { AbilityRowKind, AbilityRowView } from "../../protocol";
  import Icon from "../Icon.svelte";

  interface Props {
    rows?: AbilityRowView[] | null;
  }
  const { rows = null }: Props = $props();

  const chips = $derived(abilityChips(rows));
  const uid = $props.id();

  const ICON: Record<AbilityRowKind, IconName> = {
    triggered: "bolt",
    static: "diamond",
    activated: "cycle",
  };

  let hovered = $state<AbilityRowKind | null>(null);
  let focused = $state<AbilityRowKind | null>(null);
  let pinned = $state<AbilityRowKind | null>(null);
  const open = $derived(pinned ?? focused ?? hovered);
  const openChip = $derived(chips.find((c) => c.kind === open) ?? null);

  const els: Partial<Record<AbilityRowKind, HTMLButtonElement>> = $state({});
  let place = $state<{ left: number; top?: number; bottom?: number }>({ left: 8 });
  const LIST_MAX = 280;

  function measure(): void {
    const el = open ? els[open] : undefined;
    if (!el || typeof window === "undefined") return;
    const r = el.getBoundingClientRect();
    const vw = window.innerWidth || document.documentElement.clientWidth || 0;
    const vh = window.innerHeight || document.documentElement.clientHeight || 0;
    const width = Math.min(LIST_MAX, Math.max(0, vw - 16));
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

  function onPointerEnter(e: PointerEvent, kind: AbilityRowKind): void {
    // A touch "hover" is the start of a tap, which pins it instead.
    if (e.pointerType === "mouse" || e.pointerType === "pen") hovered = kind;
  }
  function onPointerLeave(kind: AbilityRowKind): void {
    if (hovered === kind) hovered = null;
  }
  // Keyboard focus opens it; the focus a mouse click leaves behind does
  // not, or the click that closes a pinned list would leave it open.
  function onFocus(e: FocusEvent, kind: AbilityRowKind): void {
    let keyboard = true;
    try {
      keyboard = (e.currentTarget as HTMLElement | null)?.matches(":focus-visible") ?? true;
    } catch {
      // A selector engine without :focus-visible: treat as keyboard.
    }
    focused = keyboard ? kind : null;
  }
  function onBlur(kind: AbilityRowKind): void {
    if (focused === kind) focused = null;
  }
  // The chip sits inside the card, whose own click taps, selects or
  // targets it, and inside a hand slot that starts a drag on
  // pointerdown: none of that is what pressing the chip means.
  function onPointerDown(e: PointerEvent): void {
    e.stopPropagation();
  }
  function onClick(e: MouseEvent, kind: AbilityRowKind): void {
    e.preventDefault();
    e.stopPropagation();
    pinned = pinned === kind ? null : kind;
  }
  function onContextMenu(e: MouseEvent, kind: AbilityRowKind): void {
    // A long-press on a touch screen raises the context menu: open the
    // list instead of the card's menu or the browser's.
    e.preventDefault();
    e.stopPropagation();
    pinned = kind;
  }
  function onKeydown(e: KeyboardEvent): void {
    if (e.key === "Escape" && open) {
      e.stopPropagation();
      pinned = null;
      hovered = null;
      focused = null;
      return;
    }
    // Enter and Space press the button (its click pins the list); the
    // card's own keydown must not take them as a press on the card.
    if (e.key === "Enter" || e.key === " ") e.stopPropagation();
  }
  function onWindowPointerDown(e: PointerEvent): void {
    if (!pinned) return;
    const el = els[pinned];
    if (el && !el.contains(e.target as Node)) pinned = null;
  }
</script>

<svelte:window onpointerdown={onWindowPointerDown} />

{#each chips as chip (chip.kind)}
  {@const descID = `${uid}-${chip.kind}`}
  <button
    type="button"
    class="ability-chip ability-chip-{chip.kind}"
    data-ability-kind={chip.kind}
    bind:this={els[chip.kind]}
    aria-label={L.abilityChip(chip.count, chip.kind)}
    aria-describedby={descID}
    aria-expanded={open === chip.kind}
    onpointerenter={(e) => onPointerEnter(e, chip.kind)}
    onpointerleave={() => onPointerLeave(chip.kind)}
    onpointerdown={onPointerDown}
    onfocus={(e) => onFocus(e, chip.kind)}
    onblur={() => onBlur(chip.kind)}
    onclick={(e) => onClick(e, chip.kind)}
    oncontextmenu={(e) => onContextMenu(e, chip.kind)}
    onkeydown={onKeydown}
  >
    <Icon name={ICON[chip.kind]} size={10} strokeWidth={2.25} />
    <span class="count" aria-hidden="true">{chip.count}</span>
  </button>
  <span id={descID} class="sr-only">{abilityChipDescription(chip)}</span>
{/each}

{#if openChip}
  <div
    class="ability-list"
    data-ability-list={openChip.kind}
    aria-hidden="true"
    use:portal
    style:left="{place.left}px"
    style:top={place.top !== undefined ? `${place.top}px` : null}
    style:bottom={place.bottom !== undefined ? `${place.bottom}px` : null}
    style:max-width="min({LIST_MAX}px, calc(100vw - 16px))"
  >
    <div class="head">
      <Icon name={ICON[openChip.kind]} size={12} strokeWidth={2.25} />
      <span class="title">{ABILITY_CHIP_TITLE[openChip.kind]}</span>
    </div>
    <ul class="labels">
      {#each openChip.labels as text, i (i)}
        <li>{text}</li>
      {/each}
    </ul>
  </div>
{/if}

<style>
  /* The chip matches the keyword badges it follows (KeywordBadgeRow's
     .kw-badge): a dark pill on the art, white glyph. It is a button, so
     it takes pointer events the keyword row itself does not, and draws
     a focus ring. */
  .ability-chip {
    font: inherit;
    display: inline-flex;
    align-items: center;
    gap: 1px;
    min-width: 14px;
    height: 14px;
    padding: 0 3px 0 2px;
    margin: 0;
    box-sizing: border-box;
    color: #fff;
    background: rgba(0, 0, 0, 0.72);
    border: 1px solid rgba(255, 255, 255, 0.45);
    border-radius: 3px;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    line-height: 1;
    cursor: help;
    pointer-events: auto;
    box-shadow: none;
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
  }
  .ability-chip:hover,
  .ability-chip[aria-expanded="true"] {
    background: rgba(0, 0, 0, 0.88);
    border-color: rgba(255, 255, 255, 0.8);
  }
  .ability-chip:focus-visible {
    outline: 2px solid var(--accent, #ffd07a);
    outline-offset: 1px;
  }
  .count {
    font-family: var(--font-mono, monospace);
    font-size: 9px;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
    text-shadow: 0 1px 0 rgba(0, 0, 0, 0.6);
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

  /* The list. Portalled to <body>, so it is placed on the viewport. */
  .ability-list {
    position: fixed;
    z-index: 1000;
    box-sizing: border-box;
    min-width: 160px;
    padding: 7px 10px 8px;
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
    margin-bottom: 5px;
    color: var(--fg-muted);
  }
  .title {
    font-weight: 700;
    font-size: 11px;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .labels {
    margin: 0;
    padding: 0 0 0 14px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
</style>
