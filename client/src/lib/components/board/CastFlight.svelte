<script lang="ts">
  // CastFlight — a spell travelling to the stack from where it was cast
  // (#3381). Every rule lives in lib/castFlight.ts and is unit-tested
  // there; this file only measures, draws and flies.
  //
  // After a frame that brings a `cast` entry has rendered, a copy of the
  // spell is drawn over its place on the stack and flown in from where it
  // came from: the caster's hand, their graveyard / exile / library pile,
  // or their seat. The real stack card is hidden for the length of the
  // flight so there is one card on screen, not two, and shown again when
  // the copy lands (or this layer goes away).
  //
  // Display only: aria-hidden, no pointer events, and the log and the
  // stack's own label already say what was cast. Gated like the linger's
  // flight: animations.cardPlay under the master switch, scaled by
  // animations.speed, and nothing at all under reduced motion.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { GameView } from "../../protocol";
  import { settings } from "../../settings";
  import { findAnchor, hasSize } from "../../boardAnchor";
  import { flyFrom } from "../../animations";
  import { cancelAnimations } from "../../motion";
  import { flightAllowed, flightTransform, type Box } from "../../stackLinger";
  import {
    CAST_FLIGHT_MS,
    castOriginSelectors,
    emptyCastTracker,
    stackItemSelector,
    trackCasts,
    type CastFlight,
  } from "../../castFlight";

  interface Props {
    view: GameView;
    boardEl: HTMLElement | null;
    // Changing this primes the tracker on the next frame (a reconnect,
    // a replay toggle), as for the linger.
    beatsPrimeKey?: string;
  }

  const { view, boardEl, beatsPrimeKey = "" }: Props = $props();

  interface Ghost {
    key: string;
    image: string;
    box: Box;
    from: { x: number; y: number; scale: number };
  }

  const LAYER_ATTR = "data-cast-flight-layer";
  // Layers whose copies are never a flight's start or landing.
  const OVERLAYS = `[${LAYER_ATTR}], [data-stack-linger-layer], [data-field-motion-layer]`;

  let ghosts = $state<Ghost[]>([]);
  let tracker = emptyCastTracker();
  let reprimeNext = false;
  let serial = 0;
  let destroyed = false;
  let raf = 0;
  // Stack cards hidden while their copy flies, with what to restore.
  const hidden = new Map<string, { el: HTMLElement; was: string }>();
  const els = new Map<string, HTMLElement>();

  function toBox(r: DOMRect, b: DOMRect): Box {
    return { left: r.left - b.left, top: r.top - b.top, width: r.width, height: r.height };
  }

  const accept = (el: HTMLElement) => hasSize(el) && !el.closest(OVERLAYS);

  function find(selectors: readonly string[]): HTMLElement | null {
    if (!boardEl) return null;
    for (const sel of selectors) {
      const el = findAnchor(boardEl, sel, { accept });
      if (el) return el;
    }
    return null;
  }

  function reveal(key: string): void {
    const h = hidden.get(key);
    if (!h) return;
    h.el.style.visibility = h.was;
    hidden.delete(key);
  }

  function land(key: string): void {
    reveal(key);
    const el = els.get(key);
    if (el) cancelAnimations(el);
    ghosts = ghosts.filter((g) => g.key !== key);
  }

  function launch(flights: readonly CastFlight[]): void {
    const b = boardEl?.getBoundingClientRect();
    if (!b) return;
    const added: Ghost[] = [];
    for (const f of flights) {
      const dest = find([stackItemSelector(f.itemID)]);
      const origin = find(castOriginSelectors(f));
      if (!dest || !origin) continue;
      const box = toBox(dest.getBoundingClientRect(), b);
      const from = flightTransform(box, toBox(origin.getBoundingClientRect(), b));
      const key = `cast-${++serial}`;
      hidden.set(key, { el: dest, was: dest.style.visibility });
      dest.style.visibility = "hidden";
      added.push({ key, image: f.image, box, from });
    }
    if (added.length > 0) ghosts = [...ghosts, ...added];
  }

  function register(node: HTMLElement, g: Ghost) {
    els.set(g.key, node);
    void flyFrom(node, g.from, CAST_FLIGHT_MS).then(() => {
      if (!destroyed) land(g.key);
    });
    return {
      destroy() {
        cancelAnimations(node);
        els.delete(g.key);
        reveal(g.key);
      },
    };
  }

  function frame(v: GameView): void {
    const res = trackCasts(tracker, v, { reprime: reprimeNext });
    reprimeNext = false;
    tracker = res.tracker;
    if (res.flights.length === 0 || !flightAllowed(get(settings))) return;
    // The stack draws the new item in this same update; measure once it
    // has laid out.
    const flights = res.flights;
    if (typeof requestAnimationFrame === "undefined") return;
    cancelAnimationFrame(raf);
    raf = requestAnimationFrame(() => {
      raf = 0;
      if (!destroyed) launch(flights);
    });
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
    destroyed = true;
    if (raf && typeof cancelAnimationFrame !== "undefined") cancelAnimationFrame(raf);
    for (const key of [...hidden.keys()]) reveal(key);
  });
</script>

<div class="cast-layer" aria-hidden="true" data-cast-flight-layer>
  {#each ghosts as g (g.key)}
    <div
      class="ghost"
      style:left={`${g.box.left}px`}
      style:top={`${g.box.top}px`}
      style:width={`${g.box.width}px`}
      style:height={`${g.box.height}px`}
      use:register={g}
    >
      <img src={g.image} alt="" decoding="async" />
    </div>
  {/each}
</div>

<style>
  /* Over the stack linger (41): the arriving card lands on top. No
     pointer events, so the board under it stays live. */
  .cast-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 42;
    overflow: hidden;
  }
  .ghost {
    position: absolute;
    box-sizing: border-box;
    border-radius: 4.75% / 3.4%;
    overflow: hidden;
    background: #111;
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.6),
      var(--shadow-lg);
  }
  .ghost img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
</style>
