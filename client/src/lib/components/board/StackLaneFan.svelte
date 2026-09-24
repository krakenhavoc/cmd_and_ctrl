<script lang="ts">
  // StackLaneFan — design A of the floating stack lane (#1467): the
  // stack as a fan of card-sized tiles, left to right, top of the stack
  // at the LEFT and largest, with a curved dashed arrow from each item
  // to each thing it targets.
  //
  // Presentation only. The model (lib/stackLane.ts) decides names,
  // order, targets and colours; lib/stackArrows.ts decides which arrows
  // exist and where they go; StackLaneHost owns the header, the summary
  // line, the aria-live announcer and the controls.
  //
  // Layout
  //   - The top item is ~180×252 and ringed in gold with a "Next"
  //     badge, the second ~150×210, the rest ~130×182, each capped by
  //     the viewport height so the lane still fits a short window.
  //     Lower items are slightly dimmed.
  //   - Under each tile: the caster's dot and name in their seat colour,
  //     then "→ target phrase", then Counter.
  //   - A deep stack SCROLLS horizontally rather than collapsing into a
  //     "+N more" stub: every item stays in the DOM, so every item can
  //     still be targeted (a counterspell aimed at the bottom spell) and
  //     can still be the end of an arrow. The top item is at the left,
  //     which is where the scroll starts, so the item that matters next
  //     is always in view.
  //   - Pending triggers go last, after a divider, marked "Waiting".
  //
  // Arrows
  //   Two SVG layers, both pointer-events: none:
  //   - the BOARD layer, board-sized and placed over the board, for
  //     arrows to permanents and players. It is a negative-z child of
  //     the lane, so it paints above the table (the host is z 38, over
  //     CombatArrows' 35/36) but under the lane's own content. Each
  //     arrow starts on the lane's border straight above or below its
  //     tile, so it never crosses the header, the summary or another
  //     tile. A tile scrolled out of the track draws no board arrow,
  //     and nor does a target hidden under the lane (it is still lit);
  //   - the LANE layer, inside the scrolling track, for an arc from one
  //     tile to another (Counterspell → Lightning Bolt). It scrolls and
  //     clips with the tiles.
  //   Re-measured on every snapshot (the model is rebuilt per view), on
  //   board / track resize, and on track scroll. No draw-on animation:
  //   the arrows fade in, and not at all under reduced motion.
  //
  // Target highlight
  //   Each board element an arrow reaches gets `data-stack-lane-target`
  //   and a `--stack-lane-ring` colour (syncTargetMarks). The ring rule
  //   is the :global one at the bottom of this file, so it exists only
  //   while this style is mounted, and the marks are cleared on unmount
  //   — the board itself carries no fan-specific code.

  import { onDestroy, untrack } from "svelte";
  import type { StackLaneItem, StackLaneStyle, StackLaneStyleProps } from "../../stackLane";
  import {
    TOP_ARROW_COLOR,
    curvePath,
    measureStackArrows,
    syncTargetMarks,
    type StackArrow,
  } from "../../stackArrows";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import Icon from "../Icon.svelte";

  interface Props extends StackLaneStyleProps {
    styleName: StackLaneStyle;
  }

  const { model, controls, styleName }: Props = $props();

  type Size = "top" | "second" | "rest";
  function sizeOf(item: StackLaneItem): Size {
    if (item.kind === "pending") return "rest";
    if (item.isTop) return "top";
    return item.position === 2 ? "second" : "rest";
  }

  function tagOf(item: StackLaneItem): string | null {
    switch (item.kind) {
      case "triggered":
        return "Trigger";
      case "activated":
        return "Ability";
      case "pending":
        return "Waiting";
      default:
        return null;
    }
  }

  // The ring on a tile that another item targets: the colour of the
  // first item aiming at it, gold when that item is the top one.
  const byID = $derived(new Map(model.items.map((it) => [it.id, it])));
  function targetedRing(item: StackLaneItem): string | null {
    const by = item.targetedBy.map((id) => byID.get(id)).find((it) => it !== undefined);
    if (!by) return null;
    return by.isTop ? TOP_ARROW_COLOR : by.casterColor;
  }

  // The model's chips, less the kind chip ("triggered", "activated")
  // the tile's tag already says.
  function chipsOf(item: StackLaneItem) {
    return item.chips.filter((c) => !(item.raw.kind !== "spell" && c.label === item.raw.kind));
  }

  function captionTitle(item: StackLaneItem): string {
    const who = item.casterIsViewer ? "You" : item.casterName;
    return item.targets.length > 0
      ? `${who} → ${item.targets.map((t) => t.phrase).join(" / ")}`
      : who;
  }

  // ---- arrows ------------------------------------------------------
  // Each SVG gets its own marker ids: an id is document-wide, and a
  // second instance of this component (a remount mid-transition) must
  // not steal the first one's arrowheads.
  const uid = `stack-fan-${Math.random().toString(36).slice(2, 9)}`;

  let rootEl = $state<HTMLElement | null>(null);
  let trackEl: HTMLElement | null = $state(null);
  let boardLayerEl: SVGSVGElement | null = $state(null);
  // The board the lane floats over — the arrows' coordinate space.
  const boardEl = $derived(rootEl?.closest<HTMLElement>(".board") ?? null);

  let boardArrows = $state<StackArrow[]>([]);
  let laneArrows = $state<StackArrow[]>([]);
  // Where the board layer sits so its 0,0 is the board's top-left.
  let layer = $state({ left: 0, top: 0, width: 0, height: 0 });

  let marked = new Set<HTMLElement>();

  function markerColors(arrows: StackArrow[]): string[] {
    return [...new Set(arrows.map((a) => a.color))];
  }
  const boardColors = $derived(markerColors(boardArrows));
  const laneColors = $derived(markerColors(laneArrows));

  function remeasure(): void {
    const board = boardEl;
    const track = trackEl;
    const lane = rootEl?.closest(".stack-lane") ?? rootEl;
    if (!board || !track || !lane) {
      boardArrows = [];
      laneArrows = [];
      marked = syncTargetMarks(marked, []);
      return;
    }
    const m = measureStackArrows({ board, lane, track, items: model.items });
    boardArrows = m.board;
    laneArrows = m.lane;
    marked = syncTargetMarks(marked, m.targets);

    // The board layer's containing block is whatever the host makes it
    // (today the lane, whose backdrop-filter establishes one). Rather
    // than assume that, read where the layer's 0,0 actually landed and
    // move it onto the board's origin.
    const b = board.getBoundingClientRect();
    if (boardLayerEl) {
      const r = boardLayerEl.getBoundingClientRect();
      const originX = r.left - layer.left;
      const originY = r.top - layer.top;
      const next = {
        left: b.left - originX,
        top: b.top - originY,
        width: b.width,
        height: b.height,
      };
      if (
        next.left !== layer.left ||
        next.top !== layer.top ||
        next.width !== layer.width ||
        next.height !== layer.height
      ) {
        layer = next;
      }
    }
  }

  const raf: (cb: () => void) => number =
    typeof requestAnimationFrame === "function"
      ? (cb) => requestAnimationFrame(cb)
      : (cb) => setTimeout(cb, 16) as unknown as number;
  const cancelRaf: (h: number) => void =
    typeof cancelAnimationFrame === "function"
      ? (h) => cancelAnimationFrame(h)
      : (h) => clearTimeout(h);

  let pending = 0;
  function schedule(): void {
    cancelRaf(pending);
    pending = raf(() => untrack(remeasure));
  }

  $effect(() => {
    // Every snapshot rebuilds the model, and a snapshot is when cards
    // move: re-measure now (the DOM is updated when effects run), and
    // again next frame for anything still laying out (art, fonts).
    void model;
    const board = boardEl;
    const track = trackEl;
    if (!board || !track) return;
    untrack(remeasure);
    schedule();
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(schedule);
    ro.observe(board);
    ro.observe(track);
    return () => ro.disconnect();
  });

  onDestroy(() => {
    cancelRaf(pending);
    marked = syncTargetMarks(marked, []);
  });
</script>

{#snippet tile(item: StackLaneItem)}
  {@const src = cardImageURL(item.artCard, "art_crop")}
  {@const tag = tagOf(item)}
  {@const targetable = item.kind !== "pending" && controls.targetable(item)}
  {@const ring = targetedRing(item)}
  {@const chips = chipsOf(item)}
  <li
    class="slot"
    data-size={sizeOf(item)}
    class:top={item.isTop}
    class:pending={item.kind === "pending"}
    style:--seat-color={item.casterColor}
    onpointerenter={() => controls.hoverEnter(item)}
    onpointerleave={() => controls.hoverLeave(item)}
    onfocusin={() => controls.hoverEnter(item)}
    onfocusout={() => controls.hoverLeave(item)}
  >
    <div class="stage">
      <div
        class="tile"
        class:previewable={item.previewCard !== null}
        class:cast-targetable={targetable}
        class:targeted={ring !== null}
        style:--targeted-ring={ring}
        data-stack-item-id={item.id}
        title={item.previewCard ? `${item.name} — hover to preview` : item.name}
      >
        {#if item.isTop}<span class="next-badge">Next</span>{/if}
        {#if tag}<span class="tag" data-tag={item.kind}>{tag}</span>{/if}
        <span class="name">{item.name}</span>
        <div class="art">
          {#if src}
            <img {src} alt="" loading="lazy" decoding="async" use:cardArt={src} />
          {:else}
            <Icon name={item.kind === "spell" ? "spark" : "bolt"} size={22} />
          {/if}
        </div>
        {#if item.effect || chips.length > 0}
          <div class="tile-foot">
            {#if item.effect}<span class="effect">{item.effect.text}</span>{/if}
            {#if chips.length > 0}
              <span class="chips">
                {#each chips as chip, ci (ci)}
                  <span
                    class="chip"
                    class:flag={chip.tone === "flag"}
                    class:manual={chip.tone === "manual"}
                    title={chip.title}>{chip.label}</span
                  >
                {/each}
              </span>
            {/if}
          </div>
        {/if}
      </div>
      {#if targetable}
        <!-- A real button over the tile while a targeting prompt may
             point at this item. -->
        <button
          type="button"
          class="target-btn"
          aria-label={`Target ${item.name}`}
          onclick={() => controls.target(item)}
        ></button>
      {/if}
    </div>
    <div class="caption" title={captionTitle(item)}>
      <span class="seat-dot"></span>
      <span class="caster">{item.casterIsViewer ? "You" : item.casterName}</span>
      {#if item.targets.length > 0}
        <span class="aim">→ {item.targets.map((t) => t.phrase).join(" / ")}</span>
      {/if}
    </div>
    {#if item.kind !== "pending"}
      <button
        type="button"
        class="act counter-btn"
        disabled={!model.priority.viewerHolds}
        aria-label={`Counter ${item.name}`}
        onclick={() => controls.counter(item)}
        title={model.priority.viewerHolds ? "counter this item" : "you don't hold priority"}
      >
        Counter
      </button>
    {/if}
  </li>
{/snippet}

<div class="fan" data-stack-body={styleName} bind:this={rootEl}>
  <div
    class="track"
    class:has-stack={model.stackItems.length > 0}
    bind:this={trackEl}
    onscroll={schedule}
  >
    <svg class="lane-arcs" aria-hidden="true">
      <defs>
        {#each laneColors as color, i (color)}
          <marker
            id={`${uid}-lane-${i}`}
            viewBox="0 0 10 10"
            refX="8"
            refY="5"
            markerWidth="6"
            markerHeight="6"
            orient="auto-start-reverse"
          >
            <path d="M 0 0 L 10 5 L 0 10 z" style:fill={color} />
          </marker>
        {/each}
      </defs>
      {#each laneArrows as a (a.id)}
        <path
          class="arrow"
          class:from-top={a.fromTop}
          d={curvePath(a)}
          style:stroke={a.color}
          marker-end={`url(#${uid}-lane-${laneColors.indexOf(a.color)})`}
          data-arrow-target={a.targetID}
          data-arrow-kind={a.targetKind}
        />
      {/each}
    </svg>
    {#if model.stackItems.length > 0}
      <ol class="items" aria-label="the stack, top first">
        {#each model.stackItems as item (item.id)}
          {@render tile(item)}
        {/each}
      </ol>
    {/if}
    {#if model.pendingTriggers.length > 0}
      <div class="waiting">
        <span class="waiting-label">waiting to go on the stack</span>
        <ol class="items" aria-label="waiting to go on the stack">
          {#each model.pendingTriggers as item (item.id)}
            {@render tile(item)}
          {/each}
        </ol>
      </div>
    {/if}
  </div>
  <!-- The board layer. Last child, so no sibling is laid out after it;
       placed onto the board's origin by remeasure(). -->
  <svg
    class="board-arrows"
    bind:this={boardLayerEl}
    aria-hidden="true"
    style:left={`${layer.left}px`}
    style:top={`${layer.top}px`}
    style:width={`${layer.width}px`}
    style:height={`${layer.height}px`}
  >
    <defs>
      {#each boardColors as color, i (color)}
        <marker
          id={`${uid}-board-${i}`}
          viewBox="0 0 10 10"
          refX="8"
          refY="5"
          markerWidth="6"
          markerHeight="6"
          orient="auto-start-reverse"
        >
          <path d="M 0 0 L 10 5 L 0 10 z" style:fill={color} />
        </marker>
      {/each}
    </defs>
    {#each boardArrows as a (a.id)}
      <path
        class="arrow"
        class:from-top={a.fromTop}
        d={curvePath(a)}
        style:stroke={a.color}
        marker-end={`url(#${uid}-board-${boardColors.indexOf(a.color)})`}
        data-arrow-target={a.targetID}
        data-arrow-kind={a.targetKind}
      />
    {/each}
  </svg>
</div>

<style>
  .fan {
    /* Tile heights; widths follow at the card ratio. The top tile is
       252px when the lane has room for it: the host caps the lane at
       min(420px, 48vh), and the header, summary, headroom, caption and
       Counter take ~172px of that, so a short window scales every tile
       down together rather than clipping the Counter row. */
    --h-top: max(120px, min(252px, calc(48vh - 172px)));
    --h-second: calc(var(--h-top) * 210 / 252);
    --h-rest: calc(var(--h-top) * 182 / 252);
    flex: 1 1 auto;
    display: flex;
    flex-direction: column;
    min-height: 0;
  }
  .track {
    position: relative;
    display: flex;
    align-items: flex-start;
    gap: 22px;
    flex: 1 1 auto;
    min-height: 0;
    overflow-x: auto;
    /* Only ever needed when the summary wraps on a very short window. */
    overflow-y: auto;
    /* Headroom for the Next badge and for the arcs between tiles. */
    padding: 28px 4px 4px;
    scrollbar-width: thin;
    --stage-h: var(--h-rest);
  }
  .track.has-stack {
    --stage-h: var(--h-top);
  }
  .lane-arcs {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    overflow: visible;
    pointer-events: none;
    /* Under the tiles and buttons, above the lane's panel. */
    z-index: -1;
  }
  .board-arrows {
    position: absolute;
    overflow: visible;
    pointer-events: none;
    z-index: -1;
  }
  .arrow {
    fill: none;
    stroke-width: 2.5;
    stroke-dasharray: 6 5;
    stroke-linecap: round;
    filter: drop-shadow(0 0 3px rgba(0, 0, 0, 0.6));
    animation: arrow-in 180ms var(--ease) both;
  }
  .arrow.from-top {
    stroke-width: 3;
  }
  @keyframes arrow-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .arrow {
      animation: none;
    }
  }
  :global(:root[data-reduce-motion="1"]) .arrow {
    animation: none;
  }
  .items {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    align-items: flex-start;
    gap: 22px;
    flex: 0 0 auto;
  }
  .slot {
    flex: 0 0 auto;
    display: flex;
    flex-direction: column;
    gap: 4px;
    --h: var(--h-rest);
    width: calc(var(--h) * 5 / 7);
  }
  .slot[data-size="top"] {
    --h: var(--h-top);
  }
  .slot[data-size="second"] {
    --h: var(--h-second);
  }
  /* Every tile is centred on one line, so captions share a baseline. */
  .stage {
    position: relative;
    height: var(--stage-h);
    display: flex;
    align-items: center;
  }
  .tile {
    position: relative;
    box-sizing: border-box;
    width: 100%;
    height: var(--h);
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    border-radius: 10px;
    background: color-mix(in srgb, var(--seat-color) 12%, var(--surface-raised));
    border: 1px solid var(--border-strong);
    box-shadow: var(--shadow);
    opacity: 0.85;
    transition:
      opacity 120ms var(--ease),
      box-shadow 120ms var(--ease),
      border-color 120ms var(--ease);
  }
  .slot[data-size="second"] .tile {
    opacity: 0.92;
  }
  .slot.top .tile {
    opacity: 1;
    padding: 10px;
    border-color: var(--gold);
    box-shadow:
      0 0 0 3px var(--gold),
      var(--shadow-glow);
  }
  .slot:hover .tile,
  .slot:focus-within .tile {
    opacity: 1;
  }
  .tile.previewable {
    cursor: zoom-in;
  }
  .tile.targeted {
    box-shadow:
      0 0 0 2px var(--targeted-ring),
      var(--shadow);
  }
  .slot.top .tile.targeted {
    box-shadow:
      0 0 0 3px var(--gold),
      0 0 0 6px var(--targeted-ring),
      var(--shadow-glow);
  }
  .tile.cast-targetable {
    border-color: var(--gold);
    box-shadow:
      var(--ring-gold),
      0 0 16px var(--gold-soft);
    opacity: 1;
  }
  .slot.pending .tile {
    border-style: dashed;
    opacity: 0.7;
    box-shadow: none;
  }
  .next-badge {
    position: absolute;
    left: 0;
    top: -22px;
    padding: 2px 8px;
    border-radius: 999px;
    background: var(--gold);
    color: var(--accent-fg);
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }
  .name {
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 13px;
    line-height: 1.2;
    color: var(--fg);
    min-width: 0;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow-wrap: anywhere;
  }
  .slot.top .name {
    font-size: 16px;
  }
  /* A notch on the tile's top edge, right-hand end (the Next badge is
     on the left), so the name keeps the full width — an ability's name
     is its whole label — and nothing overlaps the art. */
  .tag {
    position: absolute;
    right: 8px;
    top: -9px;
    z-index: 1;
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    padding: 1px 5px;
    border-radius: var(--radius-sm);
    background: color-mix(in srgb, var(--magenta) 28%, var(--surface));
    color: var(--magenta);
  }
  .tag[data-tag="activated"] {
    background: color-mix(in srgb, var(--mint) 24%, var(--surface));
    color: var(--mint);
  }
  .tag[data-tag="pending"] {
    background: var(--surface);
    border: 1px dashed var(--border-strong);
    color: var(--fg-muted);
  }
  .art {
    position: relative;
    --art-error-top: 4px;
    --art-error-right: 4px;
    flex: 1 1 auto;
    min-height: 0;
    border-radius: 6px;
    overflow: hidden;
    background: color-mix(in srgb, black 28%, transparent);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--seat-color);
  }
  .art img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .tile-foot {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }
  .effect {
    font-size: 11px;
    line-height: 1.35;
    color: var(--fg);
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }
  .slot.top .effect {
    font-size: 12px;
    -webkit-line-clamp: 3;
    line-clamp: 3;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 3px;
  }
  .chip {
    font-family: var(--font-mono);
    font-size: 8.5px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    padding: 0 5px;
  }
  .chip.flag {
    color: var(--gold-strong);
    border-color: color-mix(in srgb, var(--gold) 45%, transparent);
  }
  .chip.manual {
    border-style: dashed;
  }
  .target-btn {
    position: absolute;
    inset: 0;
    padding: 0;
    border: 0;
    border-radius: 10px;
    background: transparent;
    cursor: pointer;
  }
  .target-btn:focus-visible {
    outline: 2px solid var(--gold);
    outline-offset: 3px;
  }
  .caption {
    display: flex;
    align-items: center;
    gap: 5px;
    min-width: 0;
    font-size: 12px;
    white-space: nowrap;
  }
  .seat-dot {
    flex: 0 0 auto;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--seat-color);
  }
  .caster {
    flex: 0 0 auto;
    font-weight: 600;
    color: var(--seat-color);
  }
  .aim {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--fg-muted);
  }
  .act {
    align-self: flex-start;
    height: 24px;
    padding: 0 9px;
    font-size: 11px;
    border-radius: 6px;
  }
  .act:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
  .counter-btn:hover:not(:disabled) {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 50%, transparent);
  }
  .waiting {
    flex: 0 0 auto;
    position: relative;
    padding-left: 22px;
    border-left: 1px dashed var(--border-strong);
  }
  .waiting-label {
    position: absolute;
    left: 22px;
    top: -26px;
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
    white-space: nowrap;
  }

  /* The highlight ring on a board element the lane points at
     (stackArrows.syncTargetMarks). Outline and a drop-shadow rather
     than box-shadow, so it composes with a card's own selection ring
     instead of replacing it. */
  :global([data-stack-lane-target]) {
    outline: 2px solid var(--stack-lane-ring, var(--gold));
    outline-offset: 3px;
    filter: drop-shadow(
      0 0 8px color-mix(in srgb, var(--stack-lane-ring, var(--gold)) 55%, transparent)
    );
  }
</style>
