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
  //   - the "+N more" chip, on the cascade's outer top corner, when
  //     more items sit under the top one than the cascade shows. It
  //     opens the whole stack as a column of compact rows over the
  //     pile, top first; Escape or a second click closes it;
  //   - the cascade (ADR 0119 amendment 2026-10-16 §A): up to four
  //     lower items as CARDS behind the top one, the same face the top
  //     slot would draw, each a step up and toward the board's centre,
  //     a little smaller and dimmed under a --bg scrim (lib/stackPile.ts
  //     cascadeLayout). The strip of each that shows carries the
  //     caster's colour band, the title as text, and Counter; hovering
  //     one lifts the scrim and opens the hover zoom. The host says how
  //     many show (fewer on a short band) and which way is inward;
  //   - the top card, the Scryfall `normal` image at the host's width,
  //     the caster's colour as a band on its left edge. An ability has
  //     no card of its own (CR 405.1), so it is drawn as a C2 ability
  //     card (StackAbilityCard; ADR 0119 amendment 2026-10-16 §B, §E):
  //     its source's art crop and name, its cost badge, its kind, and
  //     its printed text (#3830), or its label when it has none.
  //     Every line too narrow for the text (the lower items, the
  //     "+N more" rows, the pending rows) truncates it and carries all
  //     of it in its title. A face-down spell or a source the viewer
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
  // runs from the top card (gold) and from each lower card's visible
  // strip (the caster's colour; `data-stack-arrow-from`) to each
  // permanent or player it targets. The
  // `data-stack-lane-target` ring on each target is StackTargetRings'
  // job, for every style (ADR 0119 §4). A target that is
  // another stack item gets no arc here; the target line and the ring
  // on that item's line say it.

  import { onDestroy, untrack } from "svelte";
  import type { StackLaneItem, StackLaneStyle, StackLaneStyleProps } from "../../stackLane";
  import {
    TOP_ARROW_COLOR,
    curvePath,
    measureStackArrows,
    type StackArrow,
  } from "../../stackArrows";
  import {
    PILE_CARD_MAX_W,
    PILE_PEEK_MAX,
    PILE_SHRUNK_W,
    cascadeLayout,
    pileDepth,
  } from "../../stackPile";
  import { cardImageURL } from "../../cardImage";
  import { identityEdgeBackground } from "../../stackIdentityEdge";
  import { cardArt } from "../../cardArt";
  import { boardExpandLayout } from "../../boardExpand";
  import Icon from "../Icon.svelte";
  import StackAbilityButtons from "./StackAbilityButtons.svelte";
  import StackAbilityCard from "./StackAbilityCard.svelte";

  interface Props extends StackLaneStyleProps {
    styleName: StackLaneStyle;
  }

  const { model, controls, styleName, pile }: Props = $props();

  const shrunk = $derived(pile?.shrunk === true);
  const cardW = $derived(shrunk ? PILE_SHRUNK_W : (pile?.cardWidth ?? PILE_CARD_MAX_W));

  const top = $derived(model.stackItems[0] ?? null);
  const depth = $derived(pileDepth(model.stackItems.length));
  // The cascade's depth: the host's fit (fewer on a short band), never
  // more than pileDepth allows.
  const shown = $derived(Math.max(0, Math.min(depth.peeking, pile?.lowerShown ?? depth.peeking)));
  // Nearest the top card first; drawn deepest first, so reversed below.
  const peeking = $derived(shrunk ? [] : model.stackItems.slice(1, 1 + shown));
  const more = $derived(shrunk ? 0 : depth.peeking + depth.more - shown);
  const inward = $derived(pile?.inward ?? "right");
  const cascade = $derived(cascadeLayout(cardW, peeking.length, inward));
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
  /** An ability on the stack: drawn as an ability card, not a card image. */
  function isAbility(item: StackLaneItem): boolean {
    return item.kind === "triggered" || item.kind === "activated";
  }

  // The model's chips, less the kind chip the ability card already says.
  function chipsOf(item: StackLaneItem) {
    return item.chips.filter((c) => !(item.raw.kind !== "spell" && c.label === item.raw.kind));
  }

  /**
   * What a lower ability card's strip shows instead of its text: its
   * source's name. Null for a spell, and for a source the viewer may not
   * see, whose strip keeps the title as it is.
   */
  function stripNameOf(item: StackLaneItem): string | null {
    if (!isAbility(item) || item.previewCard === null) return null;
    const name = item.previewCard.name || item.sourceName;
    return name && name !== item.text ? name : null;
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
  const boardColors = $derived([...new Set(boardArrows.map((a) => a.color))]);

  function remeasure(): void {
    const board = boardEl;
    const track = cardsEl;
    const lane = rootEl?.closest(".stack-lane") ?? rootEl;
    if (!board || !track || !lane) {
      boardArrows = [];
      return;
    }
    const m = measureStackArrows({ board, lane, track, items: model.stackItems });
    boardArrows = m.board;

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
    void cascade;
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
  });
</script>

<svelte:window onkeydown={onWindowKey} />

{#snippet counterButton(item: StackLaneItem, small: boolean)}
  <!-- #1865: a spell's own stack ability (Lightning Storm) sits beside
       Counter on the top card. A peek's name line has room for Counter
       alone; a storm under its own ability is back on top, button and
       all, before it can resolve. -->
  {#if !small}
    <StackAbilityButtons
      {item}
      enabled={model.priority.viewerHolds}
      onActivate={controls.activate}
    />
  {/if}
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
  class:inward-left={inward === "left"}
  class:narrow={cardW < 260}
  data-stack-body={styleName}
  style:--card-w={`${cardW}px`}
  style:--pile-w={`${cascade.width}px`}
  style:--top-left={`${cascade.topLeft}px`}
  bind:this={rootEl}
>
  {#if top}
    {@const topAbility = isAbility(top)}
    {@const topSrc = topAbility ? null : imageOf(top, "normal")}
    {@const topRing = targetedRing(top)}
    {@const topEdge = identityEdgeBackground(top.previewCard, topAbility ? 63 / 88 : 488 / 680)}
    {@const chips = chipsOf(top)}
    <div class="cards" bind:this={cardsEl} style:--rise={`${cascade.rise}px`}>
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
      <!-- The cascade's lower cards, deepest first, so each is drawn
           behind the one nearer the top. -->
      {#each [...peeking].reverse() as item, i (item.id)}
        {@const ability = isAbility(item)}
        {@const src = ability ? null : imageOf(item, "normal")}
        {@const ring = targetedRing(item)}
        {@const g = cascade.cards[peeking.length - 1 - i]}
        {@const short = stripNameOf(item)}
        <!-- Pointer and focus only drive the hover preview; the controls
             inside are the buttons. -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          class="peek"
          class:ability
          class:targeted={ring !== null}
          class:cast-targetable={controls.targetable(item)}
          class:previewable={item.previewCard !== null}
          style:--seat-color={item.casterColor}
          style:--targeted-ring={ring}
          style:--dim={g.dim}
          style:--strip-h={`${cascade.step}px`}
          style:left={`${g.left}px`}
          style:top={`${g.top}px`}
          style:width={`${g.width}px`}
          style:z-index={PILE_PEEK_MAX + 1 - g.depth}
          data-depth={g.depth}
          data-stack-item-id={item.id}
          title={item.previewCard ? `${item.text} — hover to preview` : item.text}
          onpointerenter={() => controls.hoverEnter(item)}
          onpointerleave={() => controls.hoverLeave(item)}
          onfocusin={() => controls.hoverEnter(item)}
          onfocusout={() => controls.hoverLeave(item)}
        >
          <!-- The face is what the eye reads; the strip below says it in
               text, so the face is hidden from a screen reader. -->
          <div class="peek-face" aria-hidden="true">
            {#if ability}
              <StackAbilityCard {item} />
            {:else if src}
              <img class="peek-img" {src} alt="" loading="lazy" decoding="async" />
            {:else}
              <span class="no-art-glyph">
                <Icon name="spark" size={22} />
              </span>
            {/if}
          </div>
          <span class="peek-scrim" aria-hidden="true"></span>
          <span class="seat-band" aria-hidden="true"></span>
          {@render targetButton(item)}
          <!-- `data-stack-arrow-from` (ARROW_FROM_ATTR): its arrows leave
               from this strip, the part of the card that shows. -->
          <span class="peek-line" data-stack-arrow-from>
            <span class="peek-label">
              {#if short !== null}
                <!-- An ability's strip names its source: the start of its
                     text ("{3}{W}, {T}: Put a +…") does not read at a
                     glance. The full text stays as the title, for a
                     screen reader and the specs, and in the tooltip. -->
                <span class="peek-kind" aria-hidden="true"><Icon name="bolt" size={11} /></span>
                <span class="peek-source" aria-hidden="true">{short}</span>
                <span class="peek-title visually-hidden">{item.text}</span>
              {:else}
                <span class="peek-title">{item.text}</span>
              {/if}
              <span class="peek-caster">{whoOf(item)}</span>
            </span>
            {@render counterButton(item, true)}
          </span>
        </div>
      {/each}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        class="top-card"
        class:targeted={topRing !== null}
        class:cast-targetable={controls.targetable(top)}
        class:previewable={top.previewCard !== null}
        class:ability={topAbility}
        class:no-art={!topAbility && topSrc === null}
        class:edged={topEdge !== null}
        style:--seat-color={top.casterColor}
        style:--targeted-ring={topRing}
        style:--identity-edge={topEdge}
        data-stack-item-id={top.id}
        title={top.previewCard ? `${top.text} — hover to preview` : top.text}
        onpointerenter={() => controls.hoverEnter(top)}
        onpointerleave={() => controls.hoverLeave(top)}
        onfocusin={() => controls.hoverEnter(top)}
        onfocusout={() => controls.hoverLeave(top)}
      >
        {#if topEdge !== null}
          <!-- ADR 0119 amendment 2026-10-16 §C: the colour-identity edge,
               hard W U B R G bands round the card under a low glow of the
               same colours. Behind the face, which covers all but its
               rim; hidden with the card while a cast flight lands. -->
          <span class="identity-glow" aria-hidden="true"></span>
          <span class="identity-edge" aria-hidden="true"></span>
        {/if}
        <div class="top-face">
          {#if topAbility}
            <StackAbilityCard item={top} />
          {:else if topSrc}
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
          {@render targetButton(top)}
        </div>
      </div>
    </div>
    <div class="caption" style:--seat-color={top.casterColor}>
      <div class="who-row">
        <span class="seat-dot"></span>
        <span class="caster">{whoOf(top)}</span>
        <span class="grow"></span>
        {@render counterButton(top, false)}
      </div>
      {#if !topAbility}
        <div class="title top-title">{top.text}</div>
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
            <span class="row-title" title={item.text}>{item.text}</span>
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
            <span class="waiting-name" title={t.text}>{t.text}</span>
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
    width: var(--pile-w, var(--card-w));
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 0;
  }
  /* The cascade box: the lower cards are placed in it from
     lib/stackPile.ts cascadeLayout, and the top card sits `--rise`
     below its top edge, so the box is as tall as what it draws. */
  .cards {
    position: relative;
    display: flex;
    flex-direction: column;
    padding-top: var(--rise, 0px);
    /* The cards' z-indexes order them against each other only. */
    isolation: isolate;
  }
  /* Room for the "+N more" chip on the outer top corner. */
  .cards:has(.more-chip) {
    padding-top: calc(var(--rise, 0px) + 14px);
  }
  .cards:has(.more-chip) > .peek {
    margin-top: 14px;
  }

  /* ---- the cascade's lower cards (ADR 0119 amendment §A) -------------- */
  .peek {
    position: absolute;
    box-sizing: border-box;
    aspect-ratio: 488 / 680;
    border-radius: calc(var(--card-w) * 0.0475);
    overflow: hidden;
    background: var(--surface-sunken);
    box-shadow:
      0 0 0 1px var(--border-strong),
      0 -3px 10px rgba(0, 0, 0, 0.45);
  }
  .peek.ability {
    aspect-ratio: 63 / 88;
    background: var(--surface);
  }
  .peek.previewable {
    cursor: zoom-in;
  }
  .peek.targeted {
    box-shadow:
      inset 0 0 0 2px var(--targeted-ring),
      0 -3px 10px rgba(0, 0, 0, 0.45);
  }
  .peek.cast-targetable {
    box-shadow:
      inset 0 0 0 2px var(--accent),
      0 0 12px var(--accent-soft);
  }
  .peek-face {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }
  .peek-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  /* The step's dimming: a --bg scrim, 28% a step up to 60%. Hovering or
     focusing the card lifts it, so the card brightens to full while the
     hover zoom opens. */
  .peek-scrim {
    position: absolute;
    inset: 0;
    z-index: 1;
    pointer-events: none;
    background: var(--bg);
    opacity: var(--dim, 0.28);
    transition: opacity 120ms var(--ease);
  }
  .peek:hover .peek-scrim,
  .peek:focus-within .peek-scrim {
    opacity: 0;
  }
  /* The strip that shows above the card in front: the title over the
     printed name bar (or the ability card's art), Counter at its end. It
     takes no pointer itself, so a click on it reaches the Target
     button under it; Counter takes its own. */
  .peek-line {
    position: absolute;
    left: 5px;
    right: 0;
    top: 0;
    height: var(--strip-h, 34px);
    z-index: 3;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 0 6px 0 7px;
    pointer-events: none;
  }
  .peek-line > .counter-btn {
    margin-left: auto;
    pointer-events: auto;
  }
  /* The title and caster on a tab of the panel's surface, as narrow as
     they are: the card's own top, dimmed by its step, shows around it,
     so the strip reads as the edge of a card and not as a list row. */
  .peek-label {
    flex: 0 1 auto;
    min-width: 0;
    display: flex;
    align-items: baseline;
    gap: 6px;
    padding: 3px 8px;
    border-radius: var(--radius-sm);
    background: var(--surface);
    box-shadow: 0 0 0 1px var(--border-strong);
  }
  /* Counter on a strip sits over the card's mana cost: without priority
     it is dimmed by colour, not opacity, so the cost does not show
     through it. */
  .peek-line > .counter-btn:disabled {
    opacity: 1;
    color: var(--fg-muted);
  }
  /* The ability card's cost badge sits where the label does; the label
     starts with the same cost. */
  .peek-face :global(.cost-badge) {
    display: none;
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
  .peek-source {
    min-width: 0;
    font-family: var(--font-display);
    font-weight: 700;
    font-size: 12.5px;
    color: var(--fg);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  /* The ability mark, the glyph the pile's no-art abilities use. */
  .peek-kind {
    flex: 0 0 auto;
    align-self: center;
    display: inline-flex;
    color: var(--fg-muted);
  }
  .peek-title.visually-hidden {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
  /* Below a 260px top card the strips give their width to the title:
     the seat band already says whose card it is, and the caster's name
     stays for a screen reader. */
  .pile.narrow .peek-caster {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }
  .seat-band {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 5px;
    background: var(--seat-color);
    /* Over a lower card's scrim: the seat colour is never dimmed. */
    z-index: 2;
  }

  /* ---- the top card ------------------------------------------------- */
  /* The top card is a stacking context (z-index) of three layers: its
     own box-shadow (the ring and the drop shadow), then the identity
     glow and edge, then the face, which clips the art to the card's
     corners and covers the edge but for its rim. */
  .top-card {
    position: relative;
    z-index: 5;
    margin-left: var(--top-left, 0px);
    width: var(--card-w);
    aspect-ratio: 488 / 680;
    border-radius: calc(var(--card-w) * 0.0475);
    box-shadow:
      0 0 0 2px var(--accent),
      0 8px 24px rgba(0, 0, 0, 0.55);
    --art-error-top: 8px;
    --art-error-right: 8px;
  }
  .top-face {
    position: absolute;
    inset: 0;
    border-radius: inherit;
    overflow: hidden;
    background: var(--surface-sunken);
  }
  .top-card.previewable {
    cursor: zoom-in;
  }
  /* An ability card (StackAbilityCard): a card's 63:88, not a scan's. */
  .top-card.ability {
    aspect-ratio: 63 / 88;
  }
  .top-card.ability .top-face {
    background: var(--surface);
  }
  .top-card.targeted {
    box-shadow:
      0 0 0 2px var(--accent),
      0 0 0 5px var(--targeted-ring),
      0 8px 24px rgba(0, 0, 0, 0.55);
  }
  .top-card.cast-targetable {
    box-shadow:
      var(--ring),
      0 0 18px var(--accent-soft);
  }

  /* ---- the colour-identity edge (ADR 0119 amendment 2026-10-16 §C) ----
     It replaces the accent ring as the top card's edge, so there is one
     edge and not two: 3px of hard bands from the fixed --identity-*
     tokens (lib/stackIdentityEdge.ts), still, the site mark's language.
     A ring that means something more sits outside it, over a 1px gap of
     the ground: the targeted ring, or the accent while the card is a
     legal target. The seat band stays inside the face, on its left. */
  .top-card.edged {
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.55);
  }
  .top-card.edged.targeted {
    box-shadow:
      0 0 0 4px var(--bg),
      0 0 0 7px var(--targeted-ring),
      0 8px 24px rgba(0, 0, 0, 0.55);
  }
  .top-card.edged.cast-targetable {
    box-shadow:
      0 0 0 4px var(--bg),
      0 0 0 6px var(--accent),
      0 0 18px var(--accent-soft);
  }
  .identity-edge,
  .identity-glow {
    position: absolute;
    inset: -3px;
    border-radius: calc(var(--card-w) * 0.0475 + 3px);
    background: var(--identity-edge);
    pointer-events: none;
  }
  /* Low and soft: the colours, not a halo. */
  .identity-glow {
    inset: -2px;
    filter: blur(9px);
    opacity: 0.32;
  }
  /* High contrast has no glows (its --ready-glow is transparent): the
     bands alone, which read on its black. */
  :global(:root[data-theme="high-contrast"]) .identity-glow {
    display: none;
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
    outline: 2px solid var(--accent);
    outline-offset: -3px;
  }

  /* ---- the caption ---------------------------------------------------- */
  .caption {
    display: flex;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
    /* Under the top card, not under the cascade's overhang. */
    margin-left: var(--top-left, 0px);
    width: var(--card-w);
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
    white-space: pre-line;
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
    color: var(--accent-strong);
    border-color: color-mix(in srgb, var(--accent) 45%, transparent);
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
  /* On the outer top corner, which the cascade leaves open: it steps
     toward the board's centre. */
  .more-chip {
    position: absolute;
    left: 0;
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
    border: 1px solid color-mix(in srgb, var(--accent) 55%, transparent);
    color: var(--accent);
  }
  .pile.inward-left .more-chip {
    left: auto;
    right: 0;
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
    border-color: color-mix(in srgb, var(--accent) 55%, transparent);
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
    /* --stack-arrival-hold: while a new item's arrival trace runs
       (StackArrival.svelte sets it on the board), an arrow that mounts
       waits for the trace to land. */
    animation: arrow-in 180ms var(--ease) var(--stack-arrival-hold, 0ms) both;
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
</style>
