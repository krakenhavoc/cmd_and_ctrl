<script lang="ts">
  // StackLanePile — the stack as a pile of large, readable cards on the
  // left of the table (ADR 0119 §1), the default style.
  //
  // Presentation only, like the other lane designs: the model
  // (lib/stackLane.ts) decides names, order, targets, colours and
  // chips; lib/stackPile.ts decides sizes and placement; StackLaneHost
  // owns the header, the summary line, the collapse tab, the announcer
  // and the controls, and positions the pile.
  //
  // Layout, top to bottom:
  //   - the "+N more" chip, when more than four items sit under the top
  //     one. It opens the whole stack as a column of compact rows over
  //     the pile, top first; Escape or a second click closes it;
  //   - up to four lower items, deepest first, each a 34px name line
  //     cut from its card's top edge (a `normal` image's name bar):
  //     the caster's colour, the title as text, and Counter;
  //   - the top card, the Scryfall `normal` image at the host's width,
  //     the caster's colour as a band on its left edge. An ability has
  //     no card of its own (CR 405.1), so it is drawn as its source's
  //     card with a band across the text box carrying its verbatim
  //     label and its kind. A face-down spell or a source the viewer
  //     may not see is a card back (#697);
  //   - the caption: caster, title, the target line, the chips
  //     (`manual` and the taken-from chip included) and Counter;
  //   - the pending triggers, as compact rows in a group headed
  //     "waiting to go on the stack".
  //
  // While the viewer chooses a target the host says `shrunk`: the top
  // card drops to 160px and the lower items and pending rows go, so
  // the board under the pile can be reached. Unless a stack item is
  // itself a legal target — then the host does not shrink it, because
  // the lower items are what the viewer has to click.
  //
  // Arrows: the fan's machinery (lib/stackArrows.ts). A board arrow
  // runs from the top card (gold) and from each peeking line (the
  // caster's colour) to each permanent or player it targets, and each
  // target gets the `data-stack-lane-target` ring. A target that is
  // another stack item gets no arc here; the target line and the ring
  // on that item's line say it.

  import { onDestroy, untrack } from "svelte";
  import type { StackLaneItem, StackLaneStyle, StackLaneStyleProps } from "../../stackLane";
  import {
    TOP_ARROW_COLOR,
    curvePath,
    measureStackArrows,
    syncTargetMarks,
    type StackArrow,
  } from "../../stackArrows";
  import { PILE_CARD_MAX_W, PILE_SHRUNK_W, pileDepth } from "../../stackPile";
  import { cardImageURL } from "../../cardImage";
  import { cardArt } from "../../cardArt";
  import { boardExpandLayout } from "../../boardExpand";
  import Icon from "../Icon.svelte";

  interface Props extends StackLaneStyleProps {
    styleName: StackLaneStyle;
  }

  const { model, controls, styleName, pile }: Props = $props();

  const shrunk = $derived(pile?.shrunk === true);
  const cardW = $derived(shrunk ? PILE_SHRUNK_W : (pile?.cardWidth ?? PILE_CARD_MAX_W));

  const top = $derived(model.stackItems[0] ?? null);
  const depth = $derived(pileDepth(model.stackItems.length));
  // Nearest the top card first; drawn deepest first, so reversed below.
  const peeking = $derived(shrunk ? [] : model.stackItems.slice(1, 1 + depth.peeking));
  const more = $derived(shrunk ? 0 : depth.more);
  const showPending = $derived(model.pendingTriggers.length > 0 && (!shrunk || top === null));

  // ---- "+N more" ------------------------------------------------------
  let moreOpen = $state(false);
  $effect(() => {
    if (more === 0 && moreOpen) moreOpen = false;
  });
  function onWindowKey(e: KeyboardEvent): void {
    if (moreOpen && e.key === "Escape") {
      moreOpen = false;
      // Taken: the expanded board's Escape (ADR 0120 §4) reads this
      // when it runs after this handler, `data-escape-owner` below when
      // it runs before.
      e.preventDefault();
      e.stopPropagation();
    }
  }

  // ---- per-item helpers -------------------------------------------------
  function kindTag(item: StackLaneItem): string | null {
    if (item.kind === "triggered") return "triggered ability";
    if (item.kind === "activated") return "activated ability";
    return null;
  }

  // The model's chips, less the kind chip the ability band already says.
  function chipsOf(item: StackLaneItem) {
    return item.chips.filter((c) => !(item.raw.kind !== "spell" && c.label === item.raw.kind));
  }

  function whoOf(item: StackLaneItem): string {
    return item.casterIsViewer ? "You" : item.casterName;
  }

  function aimOf(item: StackLaneItem): string {
    return item.targets.length > 0 ? `→ ${item.targets.map((t) => t.phrase).join(" / ")}` : "";
  }

  /** What a screen reader reads for the card image: what the eye reads. */
  function altOf(item: StackLaneItem): string {
    if (!item.previewCard) return "card back";
    const name = item.previewCard.name || item.name;
    return item.cardText ? `${name}. ${item.cardText}` : name;
  }

  /** The image for an item: its card's, a card back, or none (no art to serve). */
  function imageOf(item: StackLaneItem, size: "normal" | "small"): string | null {
    if (!item.previewCard) return size === "normal" ? "/card-back.jpg" : "/card-back-small.jpg";
    return cardImageURL(item.previewCard, size);
  }

  // The ring on an item another item targets: the first one aiming at
  // it, gold when that one is the top.
  const byID = $derived(new Map(model.items.map((it) => [it.id, it])));
  function targetedRing(item: StackLaneItem): string | null {
    const by = item.targetedBy.map((id) => byID.get(id)).find((it) => it !== undefined);
    if (!by) return null;
    return by.isTop ? TOP_ARROW_COLOR : by.casterColor;
  }

  function counterTitle(): string {
    return model.priority.viewerHolds ? "counter this item" : "you don't hold priority";
  }

  // ---- arrows --------------------------------------------------------
  const uid = `stack-pile-${Math.random().toString(36).slice(2, 9)}`;
  let rootEl = $state<HTMLElement | null>(null);
  let cardsEl = $state<HTMLElement | null>(null);
  let boardLayerEl: SVGSVGElement | null = $state(null);
  const boardEl = $derived(rootEl?.closest<HTMLElement>(".board") ?? null);

  let boardArrows = $state<StackArrow[]>([]);
  let layer = $state({ left: 0, top: 0, width: 0, height: 0 });
  let marked = new Set<HTMLElement>();
  const boardColors = $derived([...new Set(boardArrows.map((a) => a.color))]);

  function remeasure(): void {
    const board = boardEl;
    const track = cardsEl;
    const lane = rootEl?.closest(".stack-lane") ?? rootEl;
    if (!board || !track || !lane) {
      boardArrows = [];
      marked = syncTargetMarks(marked, []);
      return;
    }
    const m = measureStackArrows({ board, lane, track, items: model.stackItems });
    boardArrows = m.board;
    marked = syncTargetMarks(marked, m.targets);

    // Put the board layer's 0,0 on the board's, whatever its
    // containing block turned out to be (as the fan does).
    const b = board.getBoundingClientRect();
    if (boardLayerEl) {
      const r = boardLayerEl.getBoundingClientRect();
      const next = {
        left: b.left - (r.left - layer.left),
        top: b.top - (r.top - layer.top),
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
    void model;
    void cardW;
    // ADR 0120 §3: the expanded board opening, closing, changing seat
    // or resizing moves the anchors too, with no snapshot.
    void $boardExpandLayout;
    const board = boardEl;
    const track = cardsEl;
    if (!board || !track) {
      untrack(remeasure);
      return;
    }
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

<svelte:window onkeydown={onWindowKey} />

{#snippet counterButton(item: StackLaneItem, small: boolean)}
  <button
    type="button"
    class="act counter-btn"
    class:small
    disabled={!model.priority.viewerHolds}
    aria-label={`Counter ${item.name}`}
    title={counterTitle()}
    onclick={(e) => {
      e.stopPropagation();
      controls.counter(item);
    }}
  >
    Counter
  </button>
{/snippet}

{#snippet targetButton(item: StackLaneItem)}
  {#if controls.targetable(item)}
    <button
      type="button"
      class="target-btn"
      aria-label={`Target ${item.name}`}
      onclick={() => controls.target(item)}
    ></button>
  {/if}
{/snippet}

<div
  class="pile"
  class:shrunk
  data-stack-body={styleName}
  style:--card-w={`${cardW}px`}
  bind:this={rootEl}
>
  {#if top}
    {@const topSrc = imageOf(top, "normal")}
    {@const topRing = targetedRing(top)}
    {@const tag = kindTag(top)}
    {@const chips = chipsOf(top)}
    <div class="cards" bind:this={cardsEl}>
      {#if more > 0}
        <button
          type="button"
          class="more-chip"
          aria-label={`show ${more} more on the stack`}
          aria-expanded={moreOpen}
          onclick={() => (moreOpen = !moreOpen)}
        >
          +{more} more
        </button>
      {/if}
      <!-- Lower items, deepest first. -->
      {#each [...peeking].reverse() as item (item.id)}
        {@const src = imageOf(item, "normal")}
        {@const ring = targetedRing(item)}
        {@const d = (item.position ?? 2) - 1}
        <!-- Pointer and focus only drive the hover preview; the controls
             inside are the buttons. -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          class="peek"
          class:targeted={ring !== null}
          class:cast-targetable={controls.targetable(item)}
          class:previewable={item.previewCard !== null}
          style:--seat-color={item.casterColor}
          style:--targeted-ring={ring}
          style:--depth={d}
          data-stack-item-id={item.id}
          title={item.previewCard ? `${item.name} — hover to preview` : item.name}
          onpointerenter={() => controls.hoverEnter(item)}
          onpointerleave={() => controls.hoverLeave(item)}
          onfocusin={() => controls.hoverEnter(item)}
          onfocusout={() => controls.hoverLeave(item)}
        >
          {#if src}
            <img class="peek-img" {src} alt="" loading="lazy" decoding="async" />
          {/if}
          <span class="seat-band" aria-hidden="true"></span>
          <span class="peek-line">
            <span class="peek-title">{item.name}</span>
            <span class="peek-caster">{whoOf(item)}</span>
          </span>
          {@render targetButton(item)}
          {@render counterButton(item, true)}
        </div>
      {/each}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        class="top-card"
        class:targeted={topRing !== null}
        class:cast-targetable={controls.targetable(top)}
        class:previewable={top.previewCard !== null}
        class:no-art={topSrc === null}
        style:--seat-color={top.casterColor}
        style:--targeted-ring={topRing}
        data-stack-item-id={top.id}
        title={top.previewCard ? `${top.name} — hover to preview` : top.name}
        onpointerenter={() => controls.hoverEnter(top)}
        onpointerleave={() => controls.hoverLeave(top)}
        onfocusin={() => controls.hoverEnter(top)}
        onfocusout={() => controls.hoverLeave(top)}
      >
        {#if topSrc}
          <img
            class="card-img"
            src={topSrc}
            alt={altOf(top)}
            decoding="async"
            use:cardArt={topSrc}
          />
        {:else}
          <span class="no-art-glyph" aria-hidden="true">
            <Icon name={top.kind === "spell" ? "spark" : "bolt"} size={28} />
          </span>
        {/if}
        <span class="seat-band" aria-hidden="true"></span>
        {#if tag}
          <div class="ability-band">
            <span class="kind-tag" data-tag={top.kind}>{tag}</span>
            <span class="title band-title">{top.name}</span>
          </div>
        {/if}
        {@render targetButton(top)}
      </div>
    </div>
    <div class="caption" style:--seat-color={top.casterColor}>
      <div class="who-row">
        <span class="seat-dot"></span>
        <span class="caster">{whoOf(top)}</span>
        <span class="grow"></span>
        {@render counterButton(top, false)}
      </div>
      {#if !tag}
        <div class="title top-title">{top.name}</div>
      {/if}
      {#if top.targets.length > 0}
        <div class="aim" title={aimOf(top)}>{aimOf(top)}</div>
      {/if}
      {#if chips.length > 0}
        <div class="chips">
          {#each chips as chip, ci (ci)}
            <span
              class="chip"
              class:flag={chip.tone === "flag"}
              class:manual={chip.tone === "manual"}
              title={chip.title}>{chip.label}</span
            >
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  {#if moreOpen}
    <!-- The whole stack, top first, as compact rows over the pile.
         `data-escape-owner`: its Escape closes this list, so the
         expanded board's Escape stands down while it is open (ADR 0120
         §4, Board.svelte's onExpandKey). -->
    <ol class="all-items" aria-label="the stack, top first" data-escape-owner>
      {#each model.stackItems as item (item.id)}
        {@const src = imageOf(item, "small")}
        <li
          class="row"
          class:top={item.isTop}
          style:--seat-color={item.casterColor}
          onpointerenter={() => controls.hoverEnter(item)}
          onpointerleave={() => controls.hoverLeave(item)}
          onfocusin={() => controls.hoverEnter(item)}
          onfocusout={() => controls.hoverLeave(item)}
        >
          <span class="row-pos" aria-hidden="true">{item.position}</span>
          <span class="row-thumb">
            {#if src}<img {src} alt="" loading="lazy" decoding="async" />{/if}
          </span>
          <span class="row-info">
            <span class="row-title">{item.name}</span>
            <span class="row-sub">
              <span class="seat-dot"></span>
              <span class="caster">{whoOf(item)}</span>
              {#if item.targets.length > 0}<span class="aim">{aimOf(item)}</span>{/if}
            </span>
          </span>
          {#if controls.targetable(item)}
            <button
              type="button"
              class="act"
              aria-label={`Target ${item.name}`}
              onclick={() => controls.target(item)}>Target</button
            >
          {/if}
          {@render counterButton(item, true)}
        </li>
      {/each}
    </ol>
  {/if}

  {#if showPending}
    <div class="waiting" role="group" aria-label="waiting to go on the stack">
      <span class="waiting-label" aria-hidden="true">waiting to go on the stack</span>
      <ul>
        {#each model.pendingTriggers as t (t.id)}
          <li
            class="waiting-row"
            style:--seat-color={t.casterColor}
            title={`from ${t.casterName}`}
            onpointerenter={() => controls.hoverEnter(t)}
            onpointerleave={() => controls.hoverLeave(t)}
          >
            <span class="seat-dot"></span>
            <span class="waiting-name">{t.name}</span>
          </li>
        {/each}
      </ul>
    </div>
  {/if}

  <!-- The board layer, placed onto the board's origin by remeasure(). -->
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
  .pile {
    position: relative;
    width: var(--card-w);
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 0;
  }
  .cards {
    position: relative;
    display: flex;
    flex-direction: column;
    /* Room for the "+N more" chip on the top edge. */
    padding-top: 0;
  }
  .cards:has(.more-chip) {
    padding-top: 14px;
  }

  /* ---- the lower items: a 34px name line cut from each card's top ---- */
  .peek {
    position: relative;
    box-sizing: border-box;
    height: 34px;
    /* Each deeper card is a little narrower, so the pile reads as cards
       behind cards rather than as a list. */
    margin: 0 calc(var(--depth, 1) * 5px);
    border-radius: 10px 10px 0 0;
    overflow: hidden;
    background: color-mix(in srgb, var(--seat-color) 16%, var(--surface-raised));
    border: 1px solid var(--border-strong);
    border-bottom: none;
    box-shadow: 0 -2px 6px rgba(0, 0, 0, 0.35);
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 6px 0 12px;
  }
  .peek.previewable {
    cursor: zoom-in;
  }
  .peek.targeted {
    box-shadow:
      inset 0 0 0 2px var(--targeted-ring),
      0 -2px 6px rgba(0, 0, 0, 0.35);
  }
  .peek.cast-targetable {
    box-shadow:
      inset 0 0 0 2px var(--gold),
      0 0 12px var(--gold-soft);
  }
  .peek-img {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    height: auto;
    display: block;
    pointer-events: none;
  }
  /* The name line covers the printed name bar, the card's own frame
     showing through toward the right. Counter sits over its end. */
  .peek-line {
    position: absolute;
    inset: 0 0 0 5px;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 76px 0 9px;
    background: linear-gradient(
      90deg,
      color-mix(in srgb, var(--surface) 95%, transparent) 0%,
      color-mix(in srgb, var(--surface) 88%, transparent) 65%,
      color-mix(in srgb, var(--surface) 60%, transparent) 100%
    );
  }
  .peek > .counter-btn {
    margin-left: auto;
  }
  .peek-title {
    min-width: 0;
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 12.5px;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .peek-caster {
    flex: 0 0 auto;
    font-size: 11px;
    font-weight: 600;
    color: var(--seat-color);
  }
  .seat-band {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 5px;
    background: var(--seat-color);
    z-index: 1;
  }

  /* ---- the top card ------------------------------------------------- */
  .top-card {
    position: relative;
    width: var(--card-w);
    aspect-ratio: 488 / 680;
    border-radius: calc(var(--card-w) * 0.0475);
    overflow: hidden;
    background: var(--surface-sunken);
    box-shadow:
      0 0 0 2px var(--gold),
      0 8px 24px rgba(0, 0, 0, 0.55);
    --art-error-top: 8px;
    --art-error-right: 8px;
  }
  .top-card.previewable {
    cursor: zoom-in;
  }
  .top-card.targeted {
    box-shadow:
      0 0 0 2px var(--gold),
      0 0 0 5px var(--targeted-ring),
      0 8px 24px rgba(0, 0, 0, 0.55);
  }
  .top-card.cast-targetable {
    box-shadow:
      var(--ring-gold),
      0 0 18px var(--gold-soft);
  }
  .card-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .no-art-glyph {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--seat-color);
    background: color-mix(in srgb, var(--seat-color) 14%, var(--surface-raised));
  }
  /* An ability: its source's card, with a band across the text box. */
  .ability-band {
    position: absolute;
    left: 5px;
    right: 0;
    top: 60%;
    bottom: 7%;
    padding: 8px 10px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 4px;
    background: color-mix(in srgb, var(--surface) 92%, transparent);
    border-top: 1px solid color-mix(in srgb, var(--gold) 45%, transparent);
    border-bottom: 1px solid color-mix(in srgb, var(--gold) 45%, transparent);
    overflow: hidden;
  }
  .kind-tag {
    align-self: flex-start;
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    padding: 1px 6px;
    border-radius: var(--radius-sm);
    background: color-mix(in srgb, var(--magenta) 26%, var(--surface));
    color: var(--magenta);
  }
  .kind-tag[data-tag="activated"] {
    background: color-mix(in srgb, var(--mint) 22%, var(--surface));
    color: var(--mint);
  }
  .band-title {
    font-size: 13px;
    line-height: 1.3;
    font-weight: 600;
    color: var(--fg);
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    line-clamp: 4;
    -webkit-box-orient: vertical;
  }
  .target-btn {
    position: absolute;
    inset: 0;
    z-index: 2;
    padding: 0;
    border: 0;
    border-radius: inherit;
    background: transparent;
    cursor: pointer;
  }
  .target-btn:focus-visible {
    outline: 2px solid var(--gold);
    outline-offset: -3px;
  }

  /* ---- the caption ---------------------------------------------------- */
  .caption {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }
  .who-row {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }
  .grow {
    flex: 1 1 auto;
  }
  .seat-dot {
    flex: 0 0 auto;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--seat-color);
  }
  .caster {
    font-size: 12px;
    font-weight: 600;
    color: var(--seat-color);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .top-title {
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 16px;
    line-height: 1.2;
    color: var(--fg);
    overflow-wrap: anywhere;
  }
  .aim {
    font-size: 12px;
    color: var(--fg-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 3px;
  }
  .chip {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--fg-muted);
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    padding: 0 6px;
  }
  .chip.flag {
    color: var(--gold-strong);
    border-color: color-mix(in srgb, var(--gold) 45%, transparent);
  }
  .chip.manual {
    border-style: dashed;
  }
  .act {
    flex: 0 0 auto;
    height: 24px;
    padding: 0 9px;
    font-size: 11px;
    border-radius: 6px;
  }
  .act.small {
    position: relative;
    z-index: 3;
    height: 22px;
    padding: 0 7px;
    font-size: 10.5px;
  }
  .act:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .counter-btn:hover:not(:disabled) {
    color: var(--danger);
    border-color: color-mix(in srgb, var(--danger) 50%, transparent);
  }

  /* ---- "+N more" -------------------------------------------------------- */
  .more-chip {
    position: absolute;
    right: 6px;
    top: 0;
    z-index: 6;
    height: 22px;
    padding: 0 9px;
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.06em;
    background: var(--surface);
    border: 1px solid color-mix(in srgb, var(--gold) 55%, transparent);
    color: var(--gold);
  }
  .all-items {
    position: absolute;
    inset: 0;
    z-index: 5;
    margin: 0;
    padding: 30px 4px 4px;
    list-style: none;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 5px;
    background: color-mix(in srgb, var(--surface) 97%, transparent);
    border-radius: var(--radius-lg);
  }
  .row {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 3px 4px;
    border-radius: 8px;
    background: var(--surface-raised);
    border: 1px solid var(--border);
    min-width: 0;
  }
  .row.top {
    border-color: color-mix(in srgb, var(--gold) 55%, transparent);
  }
  .row-pos {
    flex: 0 0 auto;
    width: 14px;
    text-align: center;
    font-family: var(--font-mono);
    font-size: 10px;
    color: var(--fg-dim);
  }
  .row-thumb {
    flex: 0 0 auto;
    width: 30px;
    height: 42px;
    border-radius: 3px;
    overflow: hidden;
    background: var(--surface-sunken);
  }
  .row-thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .row-info {
    flex: 1 1 auto;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }
  .row-title {
    font-weight: 700;
    font-size: 12px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .row-sub {
    display: flex;
    align-items: center;
    gap: 5px;
    min-width: 0;
    font-size: 11px;
  }

  /* ---- pending triggers ------------------------------------------------ */
  .waiting {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding-top: 6px;
    border-top: 1px dashed var(--border-strong);
  }
  .waiting-label {
    font-family: var(--font-mono);
    font-size: 9px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: var(--fg-dim);
  }
  .waiting ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .waiting-row {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11.5px;
    color: var(--fg-muted);
    min-width: 0;
  }
  .waiting-name {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }

  /* ---- arrows ------------------------------------------------------------ */
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

  /* The ring on a board element the pile points at
     (stackArrows.syncTargetMarks), as the fan draws it. */
  :global([data-stack-lane-target]) {
    outline: 2px solid var(--stack-lane-ring, var(--gold));
    outline-offset: 3px;
    filter: drop-shadow(
      0 0 8px color-mix(in srgb, var(--stack-lane-ring, var(--gold)) 55%, transparent)
    );
  }
</style>
