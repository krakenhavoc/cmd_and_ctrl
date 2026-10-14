<script lang="ts">
  // FieldMotion — permanents arriving and leaving where you can see it.
  //
  // Each frame it remembers where every permanent was drawn
  // (`[data-instance-id]` of a battlefield card). When one leaves, a
  // copy is drawn where it stood, captioned with where it went ("to
  // graveyard", "exiled", "to hand"), and after a short hold flies to
  // its owner's pile, its card in the exile strip or hand, or the
  // owner's seat. A token, or a card that went nowhere this viewer can
  // find, fades in place. A permanent new on someone else's board is
  // ringed for a few seconds.
  //
  // Every rule lives in lib/fieldMotion.ts and is unit-tested there;
  // this file only measures, draws and flies. Display only, as
  // StackLinger is: the layer is aria-hidden and takes no pointer
  // events, and the public log already records every one of these
  // moves for a screen reader.
  //
  // Gating, as for the stack linger's flight: the flight is motion,
  // gated by animations.cardPlay under the master switch and scaled by
  // animations.speed; under reduced motion the copy fades where it
  // stood instead. The ring is information, so it shows with motion
  // off, without its pulse. The hold and the ring's life are reading
  // time and are never scaled.

  import { onDestroy, untrack } from "svelte";
  import { get } from "svelte/store";
  import type { CardView, GameView } from "../../protocol";
  import { settings } from "../../settings";
  import { BOARD_EXPANDED_ATTR, findAnchor, hasSize } from "../../boardAnchor";
  import { cardImageURL } from "../../cardImage";
  import { flyTo } from "../../animations";
  import { cancelAnimations, settled } from "../../motion";
  import { flightAllowed, flightTransform, type Box } from "../../stackLinger";
  import {
    ArrivalClock,
    DEPART_FADE_MS,
    DEPART_FLIGHT_MAX,
    DEPART_FLIGHT_MS,
    DEPART_HOLD_MS,
    departureCaptions,
    emptyFieldTracker,
    fieldLandingSelectors,
    trackField,
    uprightBox,
    type FieldCard,
    type FieldDeparture,
  } from "../../fieldMotion";

  interface Props {
    view: GameView;
    viewerID: string | null;
    boardEl: HTMLElement | null;
    // Changing this primes the tracker on the next frame (Game.svelte's
    // beatsPrimeKey: a reconnect, a replay toggle).
    beatsPrimeKey?: string;
  }

  const { view, viewerID, boardEl, beatsPrimeKey = "" }: Props = $props();

  interface Ghost {
    key: string;
    dep: FieldDeparture;
    box: Box;
    /** Null: another copy of this frame carries the shared caption. */
    caption: string | null;
    fading: boolean;
  }

  interface Ring {
    id: string;
    box: Box;
  }

  const LAYER_ATTR = "data-field-motion-layer";

  let ghosts = $state<Ghost[]>([]);
  let rings = $state<Ring[]>([]);
  const still = $derived($settings.accessibility.reduceMotion || !$settings.animations.enabled);

  let tracker = emptyFieldTracker();
  // What the board drew for each permanent as of the last measure, read
  // when the permanent has gone.
  let drawn = new Map<string, FieldCard>();
  const clock = new ArrivalClock();
  let reprimeNext = false;
  let serial = 0;
  let destroyed = false;
  let raf = 0;
  let ringTimer: ReturnType<typeof setTimeout> | null = null;
  const els = new Map<string, HTMLElement>();
  const timers = new Set<ReturnType<typeof setTimeout>>();

  function later(fn: () => void, ms: number): void {
    const t = setTimeout(() => {
      timers.delete(t);
      if (!destroyed) fn();
    }, ms);
    timers.add(t);
  }

  // ---- measuring ---------------------------------------------------------

  function toBox(r: DOMRect, b: DOMRect): Box {
    return { left: r.left - b.left, top: r.top - b.top, width: r.width, height: r.height };
  }

  // Every drawn tile of a permanent in `cards`, one pass over the board:
  // a copy in an expanded board overlay wins over the table's (ADR 0120
  // §3), as findAnchor would pick it.
  function tilesFor(cards: ReadonlyMap<string, CardView>): Map<string, HTMLElement> {
    const out = new Map<string, HTMLElement>();
    if (!boardEl) return out;
    const inOverlay = new Set<string>();
    for (const el of boardEl.querySelectorAll<HTMLElement>("[data-instance-id]")) {
      const id = el.dataset.instanceId;
      if (!id || !cards.has(id) || el.closest(`[${LAYER_ATTR}]`)) continue;
      const overlay = el.closest(`[${BOARD_EXPANDED_ATTR}]`) !== null;
      if (out.has(id) && (inOverlay.has(id) || !overlay)) continue;
      if (!hasSize(el)) continue;
      out.set(id, el);
      if (overlay) inOverlay.add(id);
    }
    return out;
  }

  function fieldCards(v: GameView): Map<string, CardView> {
    const out = new Map<string, CardView>();
    for (const c of v.battlefield?.cards ?? []) out.set(c.instance_id, c);
    for (const c of v.phased_out?.cards ?? []) out.set(c.instance_id, c);
    return out;
  }

  function imageOf(el: HTMLElement, c: CardView): string | null {
    const img = el.querySelector("img");
    const src = img?.currentSrc || img?.getAttribute("src") || null;
    return src ?? cardImageURL(c, "normal") ?? (c.face_down ? "/card-back.jpg" : null);
  }

  function measure(v: GameView): void {
    const b = boardEl?.getBoundingClientRect();
    if (!b) return;
    const cards = fieldCards(v);
    const next = new Map<string, FieldCard>();
    for (const [id, el] of tilesFor(cards)) {
      const c = cards.get(id)!;
      next.set(id, {
        id,
        owner: c.owner,
        controller: c.controller,
        isToken: c.is_token === true,
        tapped: c.tapped === true,
        image: imageOf(el, c),
        rect: toBox(el.getBoundingClientRect(), b),
      });
    }
    drawn = next;
    paintRings();
  }

  // An arriving card is still in its entrance pop (etbPulse scales it
  // down from 0.55), and a rect read now carries that transform. Measure
  // again once the pop has settled so the ring frames the whole card.
  function remeasureArrivals(v: GameView, ids: readonly string[]): void {
    if (ids.length === 0) return;
    const tiles = tilesFor(fieldCards(v));
    const waits = ids.flatMap((id) => {
      const el = tiles.get(id);
      return el ? [settled(el)] : [];
    });
    void Promise.all(waits).then(() => {
      if (!destroyed) measure(view);
    });
  }

  function measureSoon(): void {
    if (typeof requestAnimationFrame === "undefined") return;
    cancelAnimationFrame(raf);
    raf = requestAnimationFrame(() => {
      raf = 0;
      if (!destroyed) measure(view);
    });
  }

  // ---- arrivals ----------------------------------------------------------

  function paintRings(): void {
    const ids = clock.live(Date.now());
    rings = ids.flatMap((id) => {
      const d = drawn.get(id);
      return d ? [{ id, box: d.rect }] : [];
    });
    if (ringTimer) clearTimeout(ringTimer);
    ringTimer = null;
    const due = clock.next();
    if (due !== null) {
      ringTimer = setTimeout(
        () => {
          ringTimer = null;
          if (!destroyed) paintRings();
        },
        Math.max(16, due - Date.now()),
      );
    }
  }

  // ---- departures --------------------------------------------------------

  function removeGhost(key: string): void {
    const el = els.get(key);
    if (el) cancelAnimations(el);
    ghosts = ghosts.filter((g) => g.key !== key);
  }

  function landingFor(dep: FieldDeparture): HTMLElement | null {
    if (!boardEl || !dep.dest) return null;
    const accept = (el: HTMLElement) => hasSize(el) && !el.closest(`[${LAYER_ATTR}]`);
    for (const sel of fieldLandingSelectors(dep.dest, dep.card.id)) {
      const el = findAnchor(boardEl, sel, { accept });
      if (el) return el;
    }
    return null;
  }

  function fadeGhost(key: string): void {
    ghosts = ghosts.map((g) => (g.key === key ? { ...g, fading: true } : g));
    later(() => removeGhost(key), DEPART_FADE_MS);
  }

  function depart(deps: readonly FieldDeparture[]): void {
    const s = get(settings);
    if (!s.animations.enabled || deps.length === 0) return;
    const fly = flightAllowed(s);
    const captions = departureCaptions(deps);
    const added: Ghost[] = deps.map((dep, i) => ({
      key: `field-${++serial}`,
      dep,
      box: uprightBox(dep.card),
      caption: captions[i],
      fading: false,
    }));
    ghosts = [...ghosts, ...added];
    added.forEach((g, i) => {
      later(() => {
        const el = els.get(g.key);
        const landing = fly && i < DEPART_FLIGHT_MAX ? landingFor(g.dep) : null;
        const b = boardEl?.getBoundingClientRect();
        if (!el || !landing || !b) {
          fadeGhost(g.key);
          return;
        }
        const to = toBox(landing.getBoundingClientRect(), b);
        void flyTo(el, flightTransform(g.box, to), DEPART_FLIGHT_MS).then(() => {
          if (!destroyed) removeGhost(g.key);
        });
      }, DEPART_HOLD_MS);
    });
  }

  function register(node: HTMLElement, key: string) {
    els.set(key, node);
    return {
      destroy() {
        cancelAnimations(node);
        els.delete(key);
      },
    };
  }

  // ---- frames ------------------------------------------------------------

  function frame(v: GameView): void {
    const res = trackField(tracker, v, viewerID, drawn, { reprime: reprimeNext });
    reprimeNext = false;
    tracker = res.tracker;
    if (res.quiet) {
      for (const g of ghosts) removeGhost(g.key);
      clock.clear();
    } else {
      depart(res.departures);
      clock.ring(res.arrivals, Date.now());
      remeasureArrivals(v, res.arrivals);
    }
    measure(v);
    measureSoon();
  }

  $effect(() => {
    void beatsPrimeKey;
    untrack(() => {
      reprimeNext = true;
    });
  });

  // Effects run after the DOM update, so `drawn` still holds where the
  // previous frame drew each permanent when this frame's departures are
  // read.
  $effect(() => {
    const v = view;
    untrack(() => frame(v));
  });

  $effect(() => {
    if (!boardEl) return;
    untrack(() => measure(view));
    if (typeof ResizeObserver === "undefined") return;
    const ro = new ResizeObserver(() => measure(view));
    ro.observe(boardEl);
    return () => ro.disconnect();
  });

  onDestroy(() => {
    destroyed = true;
    for (const t of timers) clearTimeout(t);
    timers.clear();
    if (ringTimer) clearTimeout(ringTimer);
    if (raf && typeof cancelAnimationFrame !== "undefined") cancelAnimationFrame(raf);
  });
</script>

<div class="field-layer" class:still aria-hidden="true" data-field-motion-layer>
  {#each rings as r (r.id)}
    <div
      class="ring"
      data-field-arrival={r.id}
      style:left={`${r.box.left}px`}
      style:top={`${r.box.top}px`}
      style:width={`${r.box.width}px`}
      style:height={`${r.box.height}px`}
    >
      <span class="tag">new</span>
    </div>
  {/each}
  {#each ghosts as g (g.key)}
    <div
      class="ghost"
      class:fading={g.fading}
      data-field-departure={g.dep.card.id}
      data-field-dest={g.dep.dest?.zone ?? "none"}
      style:left={`${g.box.left}px`}
      style:top={`${g.box.top}px`}
      style:width={`${g.box.width}px`}
      style:height={`${g.box.height}px`}
      use:register={g.key}
    >
      <div class="face" class:tapped={g.dep.card.tapped}>
        {#if g.dep.card.image}
          <img src={g.dep.card.image} alt="" decoding="async" />
        {/if}
      </div>
      {#if g.caption}
        <span class="caption" data-dest={g.dep.dest?.zone ?? "none"}>{g.caption}</span>
      {/if}
    </div>
  {/each}
</div>

<style>
  /* Under the stack linger (41), over the board. No pointer events, so
     the board under it stays live. */
  .field-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 40;
    overflow: hidden;
  }

  .ring {
    position: absolute;
    box-sizing: border-box;
    border-radius: 6px;
    box-shadow:
      0 0 0 2px var(--accent),
      0 0 14px 2px color-mix(in srgb, var(--accent) 60%, transparent);
    animation:
      ring-pulse 900ms ease-in-out 2,
      ring-out 2600ms linear forwards;
  }
  .still .ring {
    animation: none;
  }
  @keyframes ring-pulse {
    50% {
      box-shadow:
        0 0 0 3px var(--accent),
        0 0 22px 6px color-mix(in srgb, var(--accent) 70%, transparent);
    }
  }
  @keyframes ring-out {
    0%,
    80% {
      opacity: 1;
    }
    100% {
      opacity: 0;
    }
  }
  .tag {
    position: absolute;
    left: 50%;
    top: -9px;
    transform: translateX(-50%);
    padding: 1px 6px;
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: #1a1204;
    background: var(--accent);
    white-space: nowrap;
  }

  .ghost {
    position: absolute;
    box-sizing: border-box;
  }
  /* The lift is on the face, not the ghost: flyTo owns the ghost's
     translate and scale (about its top-left corner). */
  .ghost .face {
    animation: lift 260ms ease-out both;
  }
  .still .ghost .face {
    animation: none;
  }
  .ghost.fading {
    animation: depart-fade 320ms ease-in forwards;
  }
  @keyframes lift {
    from {
      scale: 1;
      filter: none;
    }
    to {
      scale: 1.08;
      filter: drop-shadow(0 10px 18px rgba(0, 0, 0, 0.6));
    }
  }
  @keyframes depart-fade {
    to {
      opacity: 0;
      scale: 0.85;
    }
  }
  .face {
    position: absolute;
    inset: 0;
    border-radius: 4.75% / 3.4%;
    overflow: hidden;
    background: #111;
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.6),
      var(--shadow-lg);
  }
  .face.tapped {
    rotate: 90deg;
  }
  .face img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .caption {
    position: absolute;
    left: 50%;
    bottom: -4px;
    transform: translate(-50%, 100%);
    padding: 2px 8px;
    border-radius: 999px;
    font-family: var(--font-mono);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    white-space: nowrap;
    color: #fff;
    background: #3c3a36;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  }
  .caption[data-dest="graveyard"] {
    background: color-mix(in srgb, var(--danger) 80%, #2a0505);
  }
  .caption[data-dest="exile"] {
    background: #5b4a8a;
  }
  .caption[data-dest="hand"],
  .caption[data-dest="library"] {
    background: #2f5d7a;
  }
</style>
