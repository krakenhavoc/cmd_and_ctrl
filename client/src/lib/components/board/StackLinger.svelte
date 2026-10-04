<script lang="ts">
  // StackLinger — the linger on resolution (ADR 0119 §3).
  //
  // One overlay for every stack style. Each frame it remembers where
  // each stack item was drawn (`[data-stack-item-id]`: the pile's top
  // card and peeking lines, a compact row, a lane tile). When an item
  // leaves and the log says why, a copy of it is drawn there for about
  // 1.5 s with a badge — Resolved, Countered (by what), Fizzled — and
  // then flies to where the card went: its new slot on the battlefield,
  // its owner's graveyard or exile pile, the command zone. An ability,
  // a copy, or a card that went somewhere this viewer cannot see fades
  // in place instead. An item that left with no log entry fades in
  // 200 ms with no badge.
  //
  // Every rule lives in lib/stackLinger.ts and is unit-tested there;
  // this file only measures, draws and flies. It is display only: the
  // layer is aria-hidden, outside every labelled region, and takes no
  // pointer events, so the board under it stays live and the stack's
  // `stack: N on the stack` label still disappears with the stack. The
  // news is read out by its own polite live region instead.
  //
  // The badge is information, so it shows with animations off. Only
  // the flight is motion: gated by animations.cardPlay under the master
  // switch, off under reduced motion, and scaled by animations.speed.
  // The 1.5 s is reading time and is never scaled.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import { gsap } from "gsap";
  import type { GameView } from "../../protocol";
  import { buildStackLane } from "../../stackLane";
  import { settings } from "../../settings";
  import { findAnchor, hasSize } from "../../boardAnchor";
  import { etbPulse, flyTo } from "../../animations";
  import {
    LINGER_FADE_MS,
    LINGER_FLIGHT_MS,
    LingerQueue,
    OUTCOME_BADGE,
    destinationSelectors,
    emptyLingerTracker,
    flightAllowed,
    flightTransform,
    lingerAnnouncement,
    lingerDestination,
    lingerItemsFrom,
    lingerShape,
    placeLinger,
    trackLinger,
    type Box,
    type Departure,
    type LingerItem,
    type LingerOutcome,
    type QueuedLinger,
  } from "../../stackLinger";
  import Icon from "../Icon.svelte";

  interface Props {
    view: GameView;
    viewerID: string | null;
    boardEl: HTMLElement | null;
    // Changing this primes the tracker on the next frame, as on a first
    // frame (Game.svelte's beatsPrimeKey: a reconnect, a replay toggle).
    beatsPrimeKey?: string;
  }

  const { view, viewerID, boardEl, beatsPrimeKey = "" }: Props = $props();

  interface Ghost {
    key: string;
    item: LingerItem;
    outcome: LingerOutcome | null;
    by: string | null;
    /** Where the item was drawn. */
    origin: Box;
    /** Where the copy is drawn now: the origin, or moved out of the live stack's way. */
    box: Box;
    shape: "card" | "row";
    /** Fading out (no flight). */
    fading: boolean;
    /** Flying to its landing; no longer re-placed. */
    flying: boolean;
  }

  const STACK_ITEM = "[data-stack-item-id]";
  // What else counts as "the live stack" a lingering copy must not
  // cover: the compact card and a floating lane's panel.
  const STACK_BOX = "[data-stack-overlay], [data-stack-style] > .stack-lane";
  const LAYER_ATTR = "data-stack-linger-layer";

  let ghosts = $state<Ghost[]>([]);
  let announcement = $state("");
  const still = $derived($settings.accessibility.reduceMotion || !$settings.animations.enabled);

  let tracker = emptyLingerTracker();
  // Where each live stack item was drawn, in board pixels, as of the
  // last measure: read when the item has gone.
  let rects = new Map<string, Box>();
  let reprimeNext = false;
  let serial = 0;
  let destroyed = false;
  let raf = 0;
  const els = new Map<string, HTMLElement>();
  const keyOf = new Map<QueuedLinger, string>();
  const timers = new Set<ReturnType<typeof setTimeout>>();

  const queue = new LingerQueue(
    (l) => {
      const key = addGhost(l, false);
      if (key) keyOf.set(l, key);
    },
    (l) => {
      const key = keyOf.get(l);
      keyOf.delete(l);
      if (key) leave(key);
    },
  );

  function later(fn: () => void, ms: number): void {
    const t = setTimeout(() => {
      timers.delete(t);
      if (!destroyed) fn();
    }, ms);
    timers.add(t);
  }

  // ---- measuring ---------------------------------------------------------

  function boardRect(): DOMRect | null {
    return boardEl ? boardEl.getBoundingClientRect() : null;
  }

  function toBox(r: DOMRect, b: DOMRect): Box {
    return { left: r.left - b.left, top: r.top - b.top, width: r.width, height: r.height };
  }

  function measureItems(b: DOMRect): Map<string, Box> {
    const out = new Map<string, Box>();
    if (!boardEl) return out;
    for (const el of boardEl.querySelectorAll<HTMLElement>(STACK_ITEM)) {
      const id = el.dataset.stackItemId;
      if (!id || out.has(id) || !hasSize(el)) continue;
      out.set(id, toBox(el.getBoundingClientRect(), b));
    }
    return out;
  }

  function liveBoxes(b: DOMRect): Box[] {
    const out = [...measureItems(b).values()];
    if (!boardEl) return out;
    for (const el of boardEl.querySelectorAll<HTMLElement>(STACK_BOX)) {
      if (hasSize(el)) out.push(toBox(el.getBoundingClientRect(), b));
    }
    return out;
  }

  // Re-read where the stack is drawn, and move any lingering copy out
  // of the way of what is live on it now.
  function remeasure(): void {
    const b = boardRect();
    if (!b) return;
    rects = measureItems(b);
    if (ghosts.length === 0) return;
    const live = liveBoxes(b);
    const size = { width: b.width, height: b.height };
    let moved = false;
    const next = ghosts.map((g) => {
      if (g.flying || g.fading) return g;
      const box = placeLinger(g.origin, live, size);
      if (box.left === g.box.left && box.top === g.box.top) return g;
      moved = true;
      return { ...g, box };
    });
    if (moved) ghosts = next;
  }

  function remeasureSoon(): void {
    if (typeof requestAnimationFrame === "undefined") return;
    cancelAnimationFrame(raf);
    raf = requestAnimationFrame(() => {
      raf = 0;
      if (!destroyed) remeasure();
    });
  }

  // ---- ghosts ------------------------------------------------------------

  function addGhost(d: Departure, fade: boolean): string | null {
    const b = boardRect();
    if (!d.rect || !b) return null;
    const key = `linger-${++serial}`;
    const box = placeLinger(d.rect, liveBoxes(b), { width: b.width, height: b.height });
    ghosts = [
      ...ghosts,
      {
        key,
        item: d.item,
        outcome: fade ? null : d.outcome,
        by: fade ? null : d.by,
        origin: d.rect,
        box,
        shape: lingerShape(d.rect),
        fading: fade,
        flying: false,
      },
    ];
    if (fade) later(() => removeGhost(key), LINGER_FADE_MS);
    return key;
  }

  function removeGhost(key: string): void {
    const el = els.get(key);
    if (el) gsap.killTweensOf(el);
    ghosts = ghosts.filter((g) => g.key !== key);
  }

  function patchGhost(key: string, patch: Partial<Ghost>): void {
    ghosts = ghosts.map((g) => (g.key === key ? { ...g, ...patch } : g));
  }

  function landingFor(item: LingerItem): HTMLElement | null {
    if (!boardEl) return null;
    const dest = lingerDestination(view, item);
    if (!dest) return null;
    const accept = (el: HTMLElement) => hasSize(el) && !el.closest(`[${LAYER_ATTR}]`);
    for (const sel of destinationSelectors(dest)) {
      const el = findAnchor(boardEl, sel, { accept });
      if (el) return el;
    }
    return null;
  }

  // The dwell is over: fly to where the card went, or fade in place.
  function leave(key: string): void {
    const g = ghosts.find((x) => x.key === key);
    if (!g) return;
    const s = get(settings);
    const el = els.get(key);
    const landing = flightAllowed(s) ? landingFor(g.item) : null;
    const b = boardRect();
    if (el && landing && b) {
      patchGhost(key, { flying: true });
      const to = toBox(landing.getBoundingClientRect(), b);
      const onField = lingerDestination(view, g.item)?.zone === "battlefield";
      void flyTo(el, flightTransform(g.box, to), LINGER_FLIGHT_MS).then(() => {
        if (destroyed) return;
        removeGhost(key);
        // The permanent's slot pulses as the card lands in it.
        const slot = onField ? landing.closest<HTMLElement>('[role="listitem"]') : null;
        if (slot) etbPulse(slot);
      });
      return;
    }
    if (!s.animations.enabled) {
      removeGhost(key);
      return;
    }
    patchGhost(key, { fading: true });
    later(() => removeGhost(key), LINGER_FADE_MS);
  }

  function register(node: HTMLElement, key: string) {
    els.set(key, node);
    return {
      destroy() {
        gsap.killTweensOf(node);
        els.delete(key);
      },
    };
  }

  // ---- frames ------------------------------------------------------------

  function frame(v: GameView): void {
    const model = buildStackLane({
      stack: v.stack,
      stackItems: v.stack_items,
      pendingTriggers: v.pending_triggers,
      seats: v.seats,
      battlefield: v.battlefield,
      exile: v.exile,
      viewerID,
      priorityHolder: null,
      splitSecondActive: false,
    });
    const res = trackLinger(tracker, v.log, lingerItemsFrom(model), {
      reprime: reprimeNext,
      rects,
    });
    reprimeNext = false;
    tracker = res.tracker;
    if (res.primed) {
      queue.clear();
      keyOf.clear();
      ghosts = [];
    } else {
      if (get(settings).animations.enabled) {
        for (const f of res.fades) addGhost(f, true);
      }
      queue.push(res.lingers.filter((d) => d.rect !== null));
      if (res.lingers.length > 0) {
        announcement = res.lingers.map(lingerAnnouncement).join(". ");
      }
    }
    remeasure();
    remeasureSoon();
  }

  // A prime-key change primes the next frame. Declared before the frame
  // effect so a key and a view that change together prime that frame.
  $effect(() => {
    void beatsPrimeKey;
    untrack(() => {
      reprimeNext = true;
    });
  });

  // Effects run after the DOM update, so `rects` still holds where the
  // previous frame drew each item when this frame's departures are read.
  $effect(() => {
    const v = view;
    untrack(() => frame(v));
  });

  // The board element arrives after the first frame (bind:this), so
  // measure as soon as it does, and again whenever it resizes.
  $effect(() => {
    if (!boardEl) return;
    untrack(() => remeasure());
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(() => remeasure());
    ro.observe(boardEl);
    return () => ro.disconnect();
  });

  onDestroy(() => {
    destroyed = true;
    queue.clear();
    for (const t of timers) clearTimeout(t);
    timers.clear();
    if (raf && typeof cancelAnimationFrame !== "undefined") cancelAnimationFrame(raf);
  });

  const ICON: Record<LingerOutcome, "check" | "x" | "exile"> = {
    resolved: "check",
    countered: "x",
    fizzled: "exile",
  };
</script>

<div class="linger-layer" class:still aria-hidden="true" data-stack-linger-layer>
  {#each ghosts as g (g.key)}
    <div
      class="ghost"
      class:as-row={g.shape === "row"}
      class:as-card={g.shape === "card"}
      class:fading={g.fading}
      data-stack-linger={g.item.id}
      data-linger-outcome={g.outcome ?? "none"}
      style:left={`${g.box.left}px`}
      style:top={`${g.box.top}px`}
      style:width={`${g.box.width}px`}
      style:height={`${g.box.height}px`}
      style:--seat-color={g.item.casterColor}
      use:register={g.key}
    >
      {#if g.shape === "card"}
        {#if g.item.image}
          <img class="art" src={g.item.image} alt="" decoding="async" />
        {:else}
          <span class="no-art"
            ><Icon name={g.item.kind === "spell" ? "spark" : "bolt"} size={28} /></span
          >
        {/if}
        {#if g.item.kind !== "spell"}
          <span class="ability">{g.item.name}</span>
        {/if}
      {:else}
        <span class="thumb">
          {#if g.item.thumb}<img src={g.item.thumb} alt="" decoding="async" />{/if}
        </span>
        <span class="title">{g.item.name}</span>
      {/if}
      <span class="band"></span>
      {#if g.outcome}
        <span class="badge" data-outcome={g.outcome} title={OUTCOME_BADGE[g.outcome].title}>
          <span class="badge-main">
            <Icon name={ICON[g.outcome]} size={13} />
            {OUTCOME_BADGE[g.outcome].label}
          </span>
          {#if g.outcome === "countered" && g.by}
            <span class="badge-sub">by {g.by}</span>
          {:else if g.outcome === "fizzled"}
            <span class="badge-sub">no legal targets</span>
          {/if}
        </span>
      {/if}
    </div>
  {/each}
</div>
<!-- Mounted for the life of the board: a live region has to exist
     before its text changes for a screen reader to announce it. -->
<p class="sr-only" aria-live="polite" data-stack-linger-announcer>{announcement}</p>

<style>
  /* Over the stack lane (38) and the attention strip (40), under every
     modal: a lingering compact row is drawn where the strip drew it.
     It takes no pointer events, so the board under it stays live. */
  .linger-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 41;
    overflow: hidden;
  }
  .ghost {
    position: absolute;
    box-sizing: border-box;
    transition:
      left 180ms var(--ease),
      top 180ms var(--ease);
  }
  .still .ghost {
    transition: none;
  }
  .ghost.fading {
    animation: linger-fade 200ms linear forwards;
  }
  @keyframes linger-fade {
    to {
      opacity: 0;
    }
  }
  .as-card {
    border-radius: 4.75% / 3.4%;
    overflow: hidden;
    background: #111;
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.6),
      var(--shadow-lg);
  }
  .as-card .art {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .no-art {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--fg-dim);
    background: var(--surface-raised);
  }
  .ability {
    position: absolute;
    left: 0;
    right: 0;
    top: 58%;
    padding: 6px 8px 6px 10px;
    background: color-mix(in srgb, var(--surface) 92%, transparent);
    font-size: 12px;
    line-height: 1.3;
    color: var(--fg);
  }
  .band {
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 4px;
    background: var(--seat-color, var(--gold));
  }
  .as-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0 8px 0 10px;
    border-radius: var(--radius);
    background: color-mix(in srgb, var(--surface) 96%, transparent);
    border: 1px solid var(--border-strong);
    box-shadow: var(--shadow);
    overflow: hidden;
    color: var(--fg);
    font-size: 12px;
  }
  .thumb {
    flex: 0 0 auto;
    height: calc(100% - 8px);
    max-height: 56px;
    aspect-ratio: 63 / 88;
    border-radius: 3px;
    overflow: hidden;
    background: var(--surface-raised);
  }
  .thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .title {
    flex: 1 1 auto;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 600;
    opacity: 0.7;
  }
  .badge {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    gap: 1px;
    padding: 4px 10px;
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    white-space: nowrap;
    color: #fff;
    background: var(--badge-bg);
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.45),
      0 6px 18px rgba(0, 0, 0, 0.55);
    animation: badge-in 160ms ease-out both;
  }
  .as-card .badge {
    position: absolute;
    left: 50%;
    top: 38%;
    transform: translate(-50%, -50%);
    border-radius: 12px;
    padding: 6px 14px;
    font-size: 13px;
  }
  .as-row .badge {
    flex: 0 0 auto;
    flex-direction: row;
    gap: 6px;
    padding: 3px 8px;
  }
  .badge-main {
    display: inline-flex;
    align-items: center;
    gap: 5px;
  }
  .badge-sub {
    font-family: var(--font-ui);
    font-size: 10.5px;
    font-weight: 600;
    letter-spacing: 0;
    text-transform: none;
    opacity: 0.92;
  }
  .badge[data-outcome="resolved"] {
    --badge-bg: color-mix(in srgb, var(--mint) 82%, #0b3b28);
    color: #04140d;
  }
  .badge[data-outcome="countered"] {
    --badge-bg: color-mix(in srgb, var(--danger) 88%, #3a0606);
  }
  .badge[data-outcome="fizzled"] {
    --badge-bg: #6b6760;
  }
  /* The copy dims a little under its badge, so the badge reads as the
     news and the card as what it happened to. */
  .as-card[data-linger-outcome="countered"]::after,
  .as-card[data-linger-outcome="fizzled"]::after {
    content: "";
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.32);
  }
  .as-card .badge {
    z-index: 1;
  }
  @keyframes badge-in {
    from {
      opacity: 0;
    }
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
</style>
