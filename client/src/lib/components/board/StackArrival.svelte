<script lang="ts">
  // StackArrival — the trace an item leaves as it goes on the stack (ADR
  // 0119 amendment 2026-10-16 §D). Every rule lives in
  // lib/stackArrival.ts and is unit-tested there; this file only finds
  // the two ends, draws and clears.
  //
  // For each new item the frame brings (trackArrivals: never on the
  // first frame, a reconnect, a replay toggle or an undo), it finds the
  // start, the ability's source permanent or else the caster's avatar,
  // and the end, the item where the stack draws it (a lower card of the
  // pile's cascade by its visible strip), and runs a trace along the
  // stack arrows' curve between them in the caster's seat colour. An end
  // that is not drawn (behind "+N more", a collapsed panel) traces
  // nothing, and so does a start that does not show: off the board, or
  // under the stack panel (the opponent's avatar on a two-seat table,
  // where the pile docks over it).
  //
  // While a trace runs, the board carries `--stack-arrival-hold`, and the
  // pile's target arrows that mount in that window wait it out, so they
  // draw once the trace has landed (StackLanePile.svelte's `.arrow`).
  //
  // Display only: aria-hidden, no pointer events. Gated by
  // arrivalAllowed (reduced motion, the master switch, its own toggle),
  // scaled by animations.speed, and hidden outright under reduced motion.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { GameView } from "../../protocol";
  import { settings } from "../../settings";
  import { seatColor } from "../../colors";
  import { cardSelector, findAnchor, hasSize, seatSelector } from "../../boardAnchor";
  import { ARROW_FROM_ATTR, curvePath, type Box } from "../../stackArrows";
  import { stackItemSelector } from "../../castFlight";
  import {
    arrivalAllowed,
    arrivalCurve,
    arrivalDuration,
    curveLength,
    emptyArrivalTracker,
    startShows,
    tracePlan,
    trackArrivals,
    type Arrival,
    type TraceSegment,
  } from "../../stackArrival";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
    // Changing this primes the tracker on the next frame (a reconnect,
    // a replay toggle), as for the cast flight and the linger.
    beatsPrimeKey?: string;
  }

  const { view, boardEl, beatsPrimeKey = "" }: Props = $props();

  interface Trace {
    key: string;
    d: string;
    color: string;
    segments: TraceSegment[];
    travel: number;
    dur: number;
    delay: number;
  }

  const LAYER_ATTR = "data-stack-arrival-layer";
  // Copies drawn by the other motion layers: never a trace's end.
  const OVERLAYS = `[${LAYER_ATTR}], [data-cast-flight-layer], [data-stack-linger-layer], [data-field-motion-layer]`;
  // The stack's own surfaces: never a trace's start.
  const STACK_SURFACES = ".stack-lane, [data-stack-body], [data-stack-item-id]";
  const HOLD_VAR = "--stack-arrival-hold";

  let traces = $state<Trace[]>([]);
  let tracker = emptyArrivalTracker();
  let reprimeNext = false;
  let serial = 0;
  let holdUntil = 0;
  let holdTimer: ReturnType<typeof setTimeout> | null = null;
  const timers = new Set<ReturnType<typeof setTimeout>>();

  function toBox(r: DOMRect, b: DOMRect): Box {
    return { left: r.left - b.left, top: r.top - b.top, width: r.width, height: r.height };
  }

  /**
   * The start: the ability's source permanent, else the caster's avatar,
   * whichever shows first (startShows: on the board, not under the stack
   * panel `cover`).
   */
  function findStart(
    board: HTMLElement,
    a: Arrival,
    b: DOMRect,
    cover: Box | null,
  ): HTMLElement | null {
    const accept = (el: HTMLElement) =>
      el.closest(STACK_SURFACES) === null &&
      el.closest(OVERLAYS) === null &&
      hasSize(el) &&
      startShows(toBox(el.getBoundingClientRect(), b), b, cover);
    if (a.sourceCardID) {
      const src = findAnchor(board, cardSelector(a.sourceCardID), { accept });
      if (src) return src;
    }
    return findAnchor(board, seatSelector(a.casterSeat), { accept });
  }

  function findEnd(board: HTMLElement, a: Arrival): HTMLElement | null {
    const tile = findAnchor(board, stackItemSelector(a.itemID), {
      accept: (el) => hasSize(el) && el.closest(OVERLAYS) === null,
    });
    if (!tile) return null;
    // A lower card of the cascade lands by the strip of it that shows.
    const strip = tile.querySelector<HTMLElement>(`[${ARROW_FROM_ATTR}]`);
    return strip && hasSize(strip) ? strip : tile;
  }

  function hold(ms: number): void {
    const board = boardEl;
    if (!board || ms <= 0) return;
    const until = Date.now() + ms;
    if (until <= holdUntil) return;
    holdUntil = until;
    board.style.setProperty(HOLD_VAR, `${Math.round(ms)}ms`);
    if (holdTimer) clearTimeout(holdTimer);
    holdTimer = setTimeout(() => {
      holdTimer = null;
      holdUntil = 0;
      board.style.removeProperty(HOLD_VAR);
    }, ms);
  }

  function launch(arrivals: readonly Arrival[], speed: number): void {
    const board = boardEl;
    if (!board) return;
    const b = board.getBoundingClientRect();
    const dur = arrivalDuration(speed);
    const scale = dur / arrivalDuration(1);
    const seats = new Map((view.seats ?? []).map((s) => [s.id, s.seat]));
    const added: Trace[] = [];
    let longest = 0;
    for (const a of arrivals) {
      const end = findEnd(board, a);
      if (!end) continue;
      const lane = end.closest<HTMLElement>(".stack-lane");
      const cover = lane ? toBox(lane.getBoundingClientRect(), b) : null;
      const start = findStart(board, a, b, cover);
      if (!start) continue;
      const curve = arrivalCurve(
        toBox(start.getBoundingClientRect(), b),
        toBox(end.getBoundingClientRect(), b),
        b,
        cover,
      );
      const plan = tracePlan(curveLength(curve));
      const delay = a.delayMs * scale;
      const key = `arrival-${++serial}`;
      added.push({
        key,
        d: curvePath(curve),
        color: seatColor(seats.get(a.casterSeat) ?? -1),
        segments: plan.segments,
        travel: plan.travel,
        dur,
        delay,
      });
      longest = Math.max(longest, delay + dur);
      const t = setTimeout(
        () => {
          timers.delete(t);
          traces = traces.filter((x) => x.key !== key);
        },
        delay + dur + 40,
      );
      timers.add(t);
    }
    if (added.length === 0) return;
    hold(longest);
    traces = [...traces, ...added];
  }

  function frame(v: GameView): void {
    const res = trackArrivals(
      tracker,
      { stackItems: v.stack_items, log: v.log },
      { reprime: reprimeNext },
    );
    reprimeNext = false;
    tracker = res.tracker;
    const s = get(settings);
    if (res.arrivals.length === 0 || !arrivalAllowed(s)) return;
    // Effects run once this frame's DOM is in place, so the new stack
    // card and its source are both there to measure now.
    launch(res.arrivals, s.animations.speed);
  }

  $effect(() => {
    void beatsPrimeKey;
    untrack(() => {
      reprimeNext = true;
    });
  });

  $effect(() => {
    const v = view;
    untrack(() => frame(v));
  });

  onDestroy(() => {
    for (const t of timers) clearTimeout(t);
    timers.clear();
    if (holdTimer) clearTimeout(holdTimer);
    boardEl?.style.removeProperty(HOLD_VAR);
  });

  /** Where a segment's dash starts and ends its run, as stroke-dashoffset. */
  function offsets(s: TraceSegment, travel: number): { from: string; to: string } {
    // A round cap reaches half the stroke past the dash: start it a
    // little further back, so nothing shows at the source before the run.
    const from = s.len + s.back + 4;
    return { from: `${from}px`, to: `${from - travel - 4}px` };
  }
</script>

<!-- On the z ladder (DESIGN.md §4) at 39: over the stack panel (38), so
     the head meets the card's edge, and under the attention strip and
     the field motion (40), the linger and the dice (41) and the cast
     flight (42), so a spell's flying copy lands over its own trace. -->
<svg class="arrival-layer" aria-hidden="true" data-stack-arrival-layer>
  {#each traces as t (t.key)}
    <g
      class="trace"
      style:--seat={t.color}
      style:--dur={`${Math.round(t.dur)}ms`}
      style:--delay={`${Math.round(t.delay)}ms`}
    >
      {#each t.segments as s, i (i)}
        {@const o = offsets(s, t.travel)}
        <path
          class="seg"
          class:head={s.head}
          d={t.d}
          stroke-dasharray={`${s.len} ${Math.ceil(t.travel * 2 + 100)}`}
          style:--from={o.from}
          style:--to={o.to}
          style:opacity={s.opacity}
        />
      {/each}
    </g>
  {/each}
</svg>

<style>
  .arrival-layer {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    overflow: visible;
    pointer-events: none;
    z-index: 39;
  }
  /* The same dark lift the target arrows have, so the trace reads on
     any art and any skin. */
  .trace {
    filter: drop-shadow(0 0 3px rgba(0, 0, 0, 0.6));
  }
  .seg {
    fill: none;
    stroke: var(--seat);
    stroke-width: 2.5;
    stroke-linecap: round;
    stroke-dashoffset: var(--from);
    /* Not --ease, which spends most of its travel in the first fifth:
       the head eases out of the source and settles into the card, so
       the run is seen and not a blink. */
    animation: arrival-run var(--dur) cubic-bezier(0.35, 0, 0.25, 1) var(--delay) both;
  }
  /* The head: the seat colour lifted toward the ink, a touch wider. */
  .seg.head {
    stroke: color-mix(in srgb, var(--seat) 55%, var(--fg));
    stroke-width: 3.5;
  }
  @keyframes arrival-run {
    from {
      stroke-dashoffset: var(--from);
    }
    to {
      stroke-dashoffset: var(--to);
    }
  }
  /* Reduced motion: no trace at all, not a still one. */
  @media (prefers-reduced-motion: reduce) {
    .arrival-layer {
      display: none;
    }
    .seg {
      animation: none;
    }
  }
  :global(:root[data-reduce-motion="1"]) .arrival-layer {
    display: none;
  }
</style>
