<script lang="ts">
  // StackOverlay is the stack card of the board's attention strip
  // (Board.svelte's .strip column). Each item is one compact row —
  // 40×56 thumb, title, caster → targets — so a deep stack still
  // reads top-down without covering the table. The top item (next
  // to resolve) carries a gold ring. Every item carries Counter; Pass
  // and hold are the action dock's (ADR 0111 §7), so nothing here
  // duplicates its `next`.
  //
  // Items stack top-to-bottom with the top of the visual stack
  // (most-recently-cast, resolves first) at the top of the overlay.
  // The container scrolls if a deep stack overflows the strip.
  //
  // #322 — hover preview. The rows are compact by design, which made
  // the stack the one zone you could not actually read: it renders
  // its own <img> thumbs rather than mounting Card.svelte, and
  // Card.svelte was the only writer of the shared `hoveredCard`
  // store, so HoverZoomOverlay never heard about a stack row. Fixed
  // by writing the store from here on pointerenter, honouring the
  // same settings.display.hoverDelayMs dwell as the table. The
  // overlay pins top-right and the strip pins top-left, so the big
  // preview lands beside the stack rather than over it — which is
  // why this is the fix instead of permanently growing the rows and
  // paying board space every turn for a read that only happens while
  // the stack is live.
  //
  // #3429, ADR 0176 — movable. The header is a drag handle (pointer
  // events, so touch works). A drag takes the card out of the strip's
  // flow and pins it where it was dropped, as a fraction of the board
  // (lib/stackPosition.ts), clamped to stay on screen. The saved spot
  // comes in as `position` (the server's, per seat) and a drop goes out
  // through `onSave`; a spectator passes no `onSave`, so the card moves
  // for them and nothing is kept. Double-clicking the header, or the
  // reset button that shows once a spot exists, puts it back.
  //
  // The drag is captured on the header, so the release can never click a
  // stack item under the pointer.
  //
  // For an ability item the previewed card is the source permanent
  // (abilities have no card of their own on the stack) — the same
  // card artCardFor() already draws the thumb from.

  // #1467: every derivation this card used to make for itself — the
  // title, the art card, the caster, the "→ target" line, the chips,
  // the order — now comes from lib/stackLane.ts, the model the
  // floating lane styles read too, so the docked card and the lane
  // cannot say different things about the same stack. The #322 hover
  // bookkeeping is lib/stackHover.ts for the same reason.

  import type { PlayerView, StackItemView, StackPositionView, ZoneView } from "../../protocol";
  import { stackPixels, type Size } from "../../stackPosition";
  import { createStackPlacement } from "../../stackPlacement.svelte";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import Icon from "../Icon.svelte";
  import { targeting, isLegalCardTarget } from "../../targeting";
  import { hoveredCard } from "../../cardTypes";
  import { settings } from "../../settings";
  import { buildStackLane, type StackLaneItem } from "../../stackLane";
  import { createStackHover } from "../../stackHover";
  import { L } from "../../labels";

  interface Props {
    stack: ZoneView;
    // S19 sub-PR 8: the zones an ability item's SOURCE card (and a
    // trigger's target) can live in. Abilities have no card in the
    // stack zone, so their art + target names resolve against
    // these. Optional — the overlay degrades to the label glyph.
    battlefield?: ZoneView;
    exile?: ZoneView;
    stackItems: StackItemView[];
    pendingTriggers: StackItemView[];
    seats: PlayerView[];
    viewerHasPriority: boolean;
    splitSecondActive: boolean;
    onCounter: (item: StackItemView) => void;
    onTargetStackItem?: (item: StackItemView) => void;
    // Whose window it is when the viewer doesn't hold priority — the
    // header names them instead of showing actions.
    priorityHolderName?: string | null;
    // #1307: true when the named priority holder is the seat Board
    // has decided is "considering a response" — a public-timing
    // read, never a hint about whether they actually have one.
    priorityHolderConsidering?: boolean;
    // #3429: the board the card is placed on, the seat's saved spot, and
    // what a drop and a reset do about it. Without a board the card stays
    // docked.
    boardEl?: HTMLElement | null;
    position?: StackPositionView | null;
    onSave?: (pos: StackPositionView) => void | Promise<void>;
    onReset?: () => void | Promise<void>;
  }

  const {
    stack,
    battlefield,
    exile,
    stackItems,
    pendingTriggers,
    seats,
    viewerHasPriority,
    splitSecondActive,
    onCounter,
    onTargetStackItem,
    priorityHolderName = null,
    priorityHolderConsidering = false,
    boardEl = null,
    position = null,
    onSave,
    onReset,
  }: Props = $props();

  // The docked card never says "your", so it needs no viewer; the
  // priority header keeps reading its own props, which Board computes
  // from the same snapshot the lane's priority block is built from.
  const model = $derived(
    buildStackLane({
      stack,
      stackItems,
      pendingTriggers,
      seats,
      battlefield,
      exile,
      viewerID: null,
      priorityHolder: null,
      splitSecondActive,
    }),
  );

  // S20: per-item legality — with a server legal set only the
  // matching spells light up (Negate can't point at a creature
  // spell); free-form prompts fall back to "any stack item".
  function itemTargetable(item: StackLaneItem): boolean {
    const t = $targeting;
    return t !== null && isLegalCardTarget(t, item.id);
  }

  const displayItems = $derived(model.stackItems);
  const triggers = $derived(model.pendingTriggers);
  const visible = $derived(displayItems.length > 0 || triggers.length > 0 || stack.count > 0);

  // ---- #3429 movable --------------------------------------------
  // The drag, save and reset are lib/stackPlacement.svelte.ts, shared with
  // the floating lane.
  let overlayEl = $state<HTMLDivElement | undefined>();
  let geom = $state<{ board: Size; panel: Size; ox: number; oy: number } | null>(null);

  const move = createStackPlacement({
    position: () => position,
    onSave: () => onSave,
    onReset: () => onReset,
    panel: () => overlayEl,
    board: () => boardEl,
    before: () => measure(),
  });
  const placed = $derived(move.placed);
  const placedStyle = $derived.by(() => {
    if (!placed || !geom) return null;
    const px = stackPixels(placed, geom.board, geom.panel);
    return `left:${geom.ox + px.left}px;top:${geom.oy + px.top}px`;
  });

  // The card's offset parent is the strip; the saved spot is relative to
  // the board, so both rects go into one offset.
  function measure(): void {
    const el = overlayEl;
    const parent = el?.offsetParent;
    if (!el || !boardEl || !(parent instanceof HTMLElement)) return;
    const b = boardEl.getBoundingClientRect();
    const p = parent.getBoundingClientRect();
    const next = {
      board: { width: b.width, height: b.height },
      panel: { width: el.offsetWidth, height: el.offsetHeight },
      ox: b.left - p.left,
      oy: b.top - p.top,
    };
    const g = geom;
    if (
      g &&
      g.board.width === next.board.width &&
      g.board.height === next.board.height &&
      g.panel.width === next.panel.width &&
      g.panel.height === next.panel.height &&
      g.ox === next.ox &&
      g.oy === next.oy
    ) {
      return;
    }
    geom = next;
  }

  // Placed or not, keep the geometry current: the board resizes with the
  // window and the card with its stack.
  $effect(() => {
    const el = overlayEl;
    if (!el || !boardEl) return;
    measure();
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(measure);
    ro.observe(boardEl);
    ro.observe(el);
    return () => ro.disconnect();
  });

  // ---- #322 hover preview -------------------------------------
  // The overlay pins top-right and the strip pins top-left, so the
  // big preview lands beside the stack rather than over it.
  const hover = createStackHover(hoveredCard);

  function handleItemEnter(item: StackLaneItem): void {
    hover.enter(item.id, item.previewCard, $settings.display.hoverDelayMs);
  }

  function handleItemLeave(item: StackLaneItem): void {
    hover.leave(item.id);
  }

  // A stack item resolves out from under the cursor without ever
  // firing pointerleave — the row is simply removed.
  $effect(() => {
    hover.sync(displayItems.map((it) => it.id));
  });

  // Unmount (last item resolved, game over, seat switch) is the other
  // way the row can vanish mid-hover.
  $effect(() => {
    return () => hover.destroy();
  });
</script>

{#if visible}
  <!-- data-stack-overlay: lib/stackPile.ts leaves this card out when it
       measures the strip's content (ADR 0119 §1). -->
  <div
    class="overlay"
    class:placed={placed !== null}
    class:dragging={move.dragging}
    data-stack-overlay
    aria-label={L.stackPile(displayItems.length)}
    style={placedStyle}
    bind:this={overlayEl}
  >
    <!-- #3429: the header is the drag handle; double-click resets. The
         keyboard route is the reset button below, so the pointer handlers
         need no role. -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <header
      class="stack-head"
      class:movable={boardEl !== null}
      data-stack-handle
      title={boardEl ? "drag to move the stack · double-click to reset" : undefined}
      onpointerdown={move.onPointerDown}
      onpointermove={move.onPointerMove}
      onpointerup={move.onPointerUp}
      onpointercancel={move.onPointerCancel}
      ondblclick={boardEl ? move.reset : undefined}
    >
      <span class="label">stack <b class="count">{displayItems.length}</b></span>
      {#if splitSecondActive}
        <span class="hint split-second" title="no responses allowed (split second is active)">
          <Icon name="bolt" size={11} /> split second — no responses
        </span>
      {:else if viewerHasPriority}
        <span class="hint">top resolves when everyone passes · you hold priority</span>
      {:else if priorityHolderName}
        <span class="hint">
          {#if priorityHolderConsidering}
            {priorityHolderName} is considering a response…
          {:else}
            {priorityHolderName} holds priority
          {/if}
        </span>
      {/if}
      {#if placed}
        <button
          type="button"
          class="reset-pos"
          title="Reset position"
          aria-label="Reset stack position"
          onclick={move.reset}
        >
          <Icon name="undo" size={12} />
        </button>
      {/if}
      <!-- ADR 0111 §7: no hold or Pass here. Both live in the action
           dock (bottom right), and its `next` is the one pass button. -->
    </header>
    <div class="items">
      {#each displayItems as item, i (item.id)}
        {@const src = cardImageURL(item.artCard, "small")}
        {@const stackTargetable = itemTargetable(item)}
        {@const previewable = item.previewCard !== null}
        <div class="line">
          <!-- svelte-ignore a11y_no_noninteractive_tabindex -->
          <div
            class="item"
            class:top={i === 0}
            class:cast-targetable={stackTargetable}
            class:previewable
            style:--seat-color={item.casterColor}
            data-stack-item-id={item.id}
            title={previewable ? `${item.text} — hover to preview` : item.text}
            onpointerenter={() => handleItemEnter(item)}
            onpointerleave={() => handleItemLeave(item)}
            onfocusin={() => handleItemEnter(item)}
            onfocusout={() => handleItemLeave(item)}
            onclick={stackTargetable ? () => onTargetStackItem?.(item.raw) : undefined}
            onkeydown={(e) => {
              if (stackTargetable && (e.key === "Enter" || e.key === " ")) {
                e.preventDefault();
                onTargetStackItem?.(item.raw);
              }
            }}
            role={stackTargetable ? "button" : undefined}
            tabindex={stackTargetable ? 0 : undefined}
          >
            <div class="thumb">
              {#if src}
                <img {src} alt="" loading="lazy" decoding="async" use:cardArt={src} />
              {:else}
                <span class="glyph">
                  {#if item.kind === "triggered"}<Icon name="bolt" size={16} />{:else}<Icon
                      name="spark"
                      size={16}
                    />{/if}
                </span>
              {/if}
            </div>
            <div class="info">
              <div class="title">{item.text}</div>
              <div class="sub">
                <span class="seat-dot"></span>
                <span class="caster-name">{item.casterName}</span>
                {#if item.targetText}
                  <span class="target-text">{item.targetText}</span>
                {/if}
                <!-- The chips — kind, doubled trigger, auto / manual,
                     X, alt cost, gift, chosen modes, split second, held
                     priority — and why each exists are in
                     lib/stackLane.ts (chipsFor). -->
                {#each item.chips as chip, ci (ci)}
                  <span
                    class="chip"
                    class:flag={chip.tone === "flag"}
                    class:manual={chip.tone === "manual"}
                    title={chip.title}
                  >
                    {chip.label}
                  </span>
                {/each}
              </div>
            </div>
          </div>
          <button
            type="button"
            class="act counter-btn"
            disabled={!viewerHasPriority}
            onclick={(e) => {
              e.stopPropagation();
              onCounter(item.raw);
            }}
            title={viewerHasPriority ? "counter this item" : "you don't hold priority"}
          >
            Counter
          </button>
        </div>
      {/each}
    </div>
    {#if triggers.length > 0}
      <div class="pending-section">
        <span class="label">pending triggers <b class="count">{triggers.length}</b></span>
        <ul class="trigger-list">
          {#each triggers as t (t.id)}
            <li
              class="trigger-row"
              style:--seat-color={t.casterColor}
              title={`from ${t.casterName}`}
            >
              <span class="seat-dot"></span>
              <span class="trigger-label" title={t.text}>{t.text}</span>
            </li>
          {/each}
        </ul>
      </div>
    {/if}
  </div>
{/if}

<style>
  .overlay {
    width: 100%;
    max-height: 60vh;
    background: color-mix(in srgb, var(--surface) 94%, transparent);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-lg);
    display: flex;
    flex-direction: column;
    color: var(--fg);
    font-size: 12px;
    overflow: hidden;
    box-sizing: border-box;
  }
  /* #3429: a placed card leaves the strip's flow; the script sets its
     left and top from the saved spot, and the width stays the strip's. */
  .overlay.placed {
    position: absolute;
    top: 0;
    left: 0;
    z-index: 1;
  }
  .overlay.dragging {
    box-shadow: var(--shadow-lg), var(--ring);
    user-select: none;
  }
  .stack-head.movable {
    cursor: grab;
    touch-action: none;
  }
  .overlay.dragging .stack-head.movable {
    cursor: grabbing;
  }
  .reset-pos {
    flex: 0 0 auto;
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    padding: 0;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--surface-raised);
    color: var(--fg-dim);
    cursor: pointer;
  }
  .reset-pos:hover,
  .reset-pos:focus-visible {
    color: var(--fg);
    border-color: var(--border-strong);
  }
  .stack-head {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px 6px 14px;
    min-width: 0;
  }
  .label {
    font-family: var(--font-mono);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.14em;
    color: var(--fg-dim);
    font-weight: 700;
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .count {
    color: var(--fg);
    font-weight: 700;
    font-variant-numeric: tabular-nums;
  }
  .hint {
    font-size: 11px;
    color: var(--fg-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
    flex: 1 1 auto;
  }
  .hint :global(svg) {
    vertical-align: -2px;
    margin-right: 4px;
  }
  .hint.split-second {
    color: var(--danger);
    font-family: var(--font-mono);
    font-size: 10px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    font-weight: 700;
  }
  .items {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 0 10px 10px 12px;
    overflow-y: auto;
    min-height: 0;
  }
  .line {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }
  .item {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 4px 10px 4px 4px;
    border-radius: 9px;
    background: var(--surface-raised);
    border: 1px solid var(--border);
    transition:
      border-color 120ms var(--ease),
      box-shadow 120ms var(--ease);
  }
  .item.top {
    border-color: var(--accent-line);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent) 15%, transparent);
  }
  .item.cast-targetable {
    cursor: pointer;
    box-shadow: var(--ring);
    border-color: var(--accent);
  }
  .item.cast-targetable:hover {
    background: var(--surface-hover);
  }
  /* #322: the row is the hover-preview trigger, so it needs to look
     like one. zoom-in reads as "there is more of this to see"; the
     border lift is the same 120ms transition the row already
     declares. Targeting wins the cursor when it's live — a click
     there means "point my spell at this", not "show me the art". */
  .item.previewable {
    cursor: zoom-in;
  }
  .item.previewable:hover {
    border-color: var(--border-strong);
    background: var(--surface-hover);
  }
  .item.cast-targetable.previewable {
    cursor: pointer;
  }
  .thumb {
    /* positioned for the failed-art pip (#33) */
    position: relative;
    --art-error-top: 2px;
    --art-error-right: 2px;
    width: 40px;
    height: 56px;
    border-radius: 3px;
    overflow: hidden;
    flex: 0 0 auto;
    background: var(--surface-sunken);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--seat-color);
  }
  .thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .info {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  /* #3830: an ability's printed text gets two lines; the item's title
     attribute has the rest. */
  .title {
    font-weight: 700;
    font-size: 12.5px;
    line-height: 1.25;
    color: var(--fg);
    overflow: hidden;
    overflow-wrap: anywhere;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }
  .sub {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 11px;
    color: var(--fg-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }
  .seat-dot {
    display: inline-block;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--seat-color);
    flex: 0 0 auto;
  }
  .caster-name {
    color: var(--fg);
    font-weight: 600;
  }
  .target-text {
    color: var(--fg-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }
  .chip {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--fg-dim);
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 1px 6px;
    flex: 0 0 auto;
  }
  .chip.flag {
    color: var(--accent-strong);
    border-color: color-mix(in srgb, var(--accent) 45%, transparent);
  }
  /* Dashed rather than coloured. The gold `flag` chips mark things
     the engine is doing; this one marks the absence of one, and an
     outline with a gap in it says that without competing with them
     for attention. */
  .chip.manual {
    border-style: dashed;
    border-color: var(--border-strong);
  }
  .act {
    flex: 0 0 auto;
    height: 28px;
    padding: 0 10px;
    font-size: 11.5px;
    border-radius: 7px;
    display: inline-flex;
    align-items: center;
    gap: 3px;
  }
  .act:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .counter-btn:hover:not(:disabled) {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 50%, transparent);
  }
  .pending-section {
    padding: 8px 12px 10px 14px;
    border-top: 1px solid var(--border);
    background: var(--surface-sunken);
  }
  .trigger-list {
    list-style: none;
    margin: 6px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .trigger-row {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: var(--fg-muted);
  }
  .trigger-label {
    color: var(--fg-muted);
  }
</style>
