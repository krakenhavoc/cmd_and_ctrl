// effectCueMotion — how each generic effect cue moves (ADR 0209 §3, #4416).
//
// TableEffects.svelte draws a cue's parts (the art copy, its shards, the
// void, the knife, the palm, a wipe's cuts or vortex) and calls cueMotion
// once they are mounted. Every part runs on the Web Animations API, from
// the cue's own clock: `delay` after the frame, for `ms`, both already
// capped, scaled by animations.speed and compressed on a busy table
// (lib/tableEffects.ts). A part's first keyframe is its rest state and
// every animation fills both ways, so a wipe's card holds still until its
// turn and a finished part stays gone until the layer removes the cue.
//
// Nothing here touches game state, and nothing waits on it.

import { EASE, run, type Run } from "./motion";
import {
  CRACK_SNAP,
  PILE_LAND,
  PILE_STAGGER,
  SHATTER_BREAK,
  SHATTER_BREAK_LIGHT,
  TRANSFORM_TURN,
  sacrificeBeats,
} from "./tableEffects";
import type { CounterTone, EffectKind, WipeTone } from "./tableEffects";

export interface CueMotion {
  kind: EffectKind | "wipe" | "signature";
  /** When the cue starts after its frame, and how long it runs, in ms. */
  delay: number;
  ms: number;
  /** Inside a board wipe: lighter. */
  light: boolean;
  /** The card's size at rest (unturned), px. */
  width: number;
  height: number;
  /** Where a bounced card flies to, from its own centre. */
  fly: { x: number; y: number };
  /** An exile drawn into a board wipe's vortex: its path, from its own centre. */
  pull?: readonly { x: number; y: number }[] | null;
  /** A destruction replaced by exile: the exile, with a slight crumble. */
  fromDestroy?: boolean;
  board: { w: number; h: number };
  tone: WipeTone;
  /** A counter cue's counters (#4442): which way its tick moves. */
  counterTone?: CounterTone;
  /**
   * A commander going home (#4443): the copy's tap turn at the start,
   * which it straightens out of, and its scale on arrival (the command
   * zone's card is smaller than a permanent).
   */
  rot?: number;
  commandScale?: number;
  /**
   * An attachment sliding onto its host (#4444): `fly` is where its tile
   * starts, from where it rests, in the tile's own px; `seam` is where it
   * meets the host, from the cue's centre (the host's), in board px.
   */
  seam?: { x: number; y: number };
  /**
   * Cards milled or discarded (#4445): `fly` is the graveyard pile, from
   * where the cards start; `sweep` moves them as one handful.
   */
  sweep?: boolean;
  /**
   * Cards returned from a graveyard (#4446), each on its own way: from
   * the cue's origin, where it starts (x, y: its graveyard pile, where
   * the layer draws it), how far it travels from there (dx, dy) and its
   * scale where it lands (s). Absent: every card goes `fly`, at scale 1.
   */
  paths?: readonly PilePath[];
  /** The game became day or night (#4448): which sky washes the table. */
  sky?: "day" | "night";
}

export interface PilePath {
  x: number;
  y: number;
  dx: number;
  dy: number;
  s: number;
}

type Timing = KeyframeAnimationOptions;

/** The part of a cue from share `a` to share `b`, as one animation's timing. */
function span(f: CueMotion, a: number, b: number, easing: string = EASE.linear): Timing {
  const from = clamp01(a);
  const to = Math.max(from, clamp01(b));
  return {
    delay: f.delay + from * f.ms,
    duration: Math.max(1, (to - from) * f.ms),
    easing,
    fill: "both",
  };
}

function clamp01(v: number): number {
  return Math.min(1, Math.max(0, v));
}

function all<T extends Element>(node: ParentNode, sel: string): T[] {
  return [...node.querySelectorAll<T>(sel)];
}

function num(el: HTMLElement, key: string): number {
  const v = Number(el.dataset[key]);
  return Number.isFinite(v) ? v : 0;
}

/**
 * cueMotion starts a mounted cue's motion. The caller cancels it (every
 * animation under the node) when the cue is removed or skipped.
 */
export function cueMotion(node: HTMLElement, f: CueMotion): void {
  switch (f.kind) {
    case "destroy":
      shudder(node, f, 0, 0.14);
      shatter(node, f, 0, 1);
      return;
    case "sacrifice":
      sacrifice(node, f);
      return;
    case "exile":
      exile(node, f);
      return;
    case "bounce":
      bounce(node, f);
      return;
    case "counter":
      counter(node, f);
      return;
    case "fizzle":
      fizzle(node, f);
      return;
    case "drain_table":
      drainTable(node, f);
      return;
    case "monarch":
    case "initiative":
      crownGlint(node, f);
      return;
    case "pw_leave":
      pwLeave(node, f);
      return;
    case "pw_enter":
      pwEnter(node, f);
      return;
    case "token_create":
      tokenGlow(node, f);
      return;
    case "counters":
      counterPulse(node, f);
      return;
    case "commander_cast":
      commanderRise(node, f);
      return;
    case "commander_return":
      commanderDrift(node, f);
      return;
    case "attach_equip":
      attachClasp(node, f);
      return;
    case "attach_aura":
      attachShimmer(node, f);
      return;
    case "mill":
    case "discard":
    case "reanimate":
      pileSlide(node, f);
      return;
    case "library_search":
      libraryRiffle(node, f);
      return;
    case "transform":
      cardFlip(node, f);
      return;
    case "day_night":
      skyWash(node, f);
      return;
    case "crack":
      crackSplit(node, f);
      return;
    case "wipe":
      if (f.tone === "cold") vortex(node, f);
      else if (f.tone === "dark") windCuts(node, f);
      else sweep(node, f);
      return;
    default:
      return;
  }
}

// ---- The shatter (D1, and the end of a sacrifice) ----

/** The darkening the card takes as it cracks, and as its pieces fall. */
const DIM = "brightness(0.62) saturate(0.7)";
const DIMMER = "brightness(0.22) saturate(0.35)";

/** A sharp knock on the card as it is struck. */
function shudder(node: HTMLElement, f: CueMotion, a: number, b: number): void {
  const copy = node.querySelector<HTMLElement>(".copy:not(.ghost)");
  if (!copy) return;
  run(
    copy,
    [
      { offset: 0, translate: "0px 0px", scale: "1" },
      { offset: 0.12, translate: "2px -1px", scale: "1.035" },
      { offset: 0.3, translate: "-2px 1px", scale: "1.02" },
      { offset: 0.5, translate: "1.5px 0px", scale: "1.01" },
      { offset: 0.75, translate: "-1px 0px", scale: "1" },
      { offset: 1, translate: "0px 0px", scale: "1" },
    ],
    span(f, a, b),
  );
}

/**
 * shatter runs a card's break between shares `from` and `to` of the cue:
 * a flash at the impact, cracks racing from it to the edges, the card
 * darkening, then its pieces letting go from the centre outward, flying
 * away from the impact and falling, with dust thrown out.
 */
function shatter(node: HTMLElement, f: CueMotion, from: number, to: number): void {
  const box = node.querySelector<HTMLElement>(".shatter");
  if (!box) return;
  const at = (x: number) => from + (to - from) * x;
  const local = (a: number, b: number, easing?: string) => span(f, at(a), at(b), easing);
  const shards = all<HTMLElement>(box, ".shard");
  const first = shards.reduce((m, s) => Math.min(m, num(s, "at")), 1);
  const open = shards.length > 0 ? first : f.light ? SHATTER_BREAK_LIGHT : SHATTER_BREAK;

  const face = node.querySelector<HTMLElement>(".face");
  if (face) {
    run(
      face,
      [
        { offset: 0, filter: "brightness(1)", opacity: 1 },
        { offset: 0.06, filter: "brightness(1.7)", opacity: 1 },
        { offset: Math.max(0.07, open - 0.02), filter: DIM, opacity: 1 },
        { offset: open, filter: DIM, opacity: 0 },
        { offset: 1, filter: DIM, opacity: 0 },
      ],
      local(0, 1),
    );
  }

  const flare = box.querySelector<HTMLElement>(".flare");
  if (flare) {
    run(
      flare,
      [
        { offset: 0, scale: "0.2", opacity: 0 },
        { offset: 0.08, scale: "1", opacity: 1 },
        { offset: 0.5, scale: "0.7", opacity: 0.5 },
        { offset: 0.85, scale: "1.5", opacity: 0.6 },
        { offset: 1, scale: "2.2", opacity: 0 },
      ],
      local(0, Math.min(1, open + 0.12), EASE.power2Out),
    );
  }

  const cracks = box.querySelector<HTMLElement>(".crk");
  if (cracks) {
    run(
      cracks,
      [
        { offset: 0, opacity: 1 },
        { offset: open, opacity: 1 },
        { offset: Math.min(1, open + 0.16), opacity: 0 },
        { offset: 1, opacity: 0 },
      ],
      local(0, 1),
    );
    all<SVGPathElement>(cracks, ".spoke").forEach((p, i) => {
      run(
        p,
        [{ strokeDashoffset: 1 }, { strokeDashoffset: 0 }],
        local(0.02 + (i % 4) * 0.012, open * 0.7, EASE.power2Out),
      );
    });
    for (const p of all<SVGPathElement>(cracks, ".rim")) {
      run(
        p,
        [
          { strokeDashoffset: 1, opacity: 0 },
          { strokeDashoffset: 0, opacity: 1 },
        ],
        local(open * 0.35, open * 0.95, EASE.power1Out),
      );
    }
  }

  for (const s of shards) {
    const a = num(s, "at");
    const dx = num(s, "dx");
    const dy = num(s, "dy");
    const rot = num(s, "rot");
    const mid = a + (1 - a) * 0.42;
    run(
      s,
      [
        { offset: 0, opacity: 0, translate: "0px 0px", rotate: "0deg", filter: DIM },
        { offset: Math.max(0, open - 0.001), opacity: 0, translate: "0px 0px", rotate: "0deg" },
        { offset: open, opacity: 1, translate: "0px 0px", rotate: "0deg", filter: DIM },
        {
          offset: a,
          opacity: 1,
          translate: "0px 0px",
          rotate: "0deg",
          filter: DIM,
          easing: EASE.power2Out,
        },
        {
          offset: mid,
          opacity: 0.95,
          translate: `${dx * 0.55}px ${dy * 0.32}px`,
          rotate: `${rot * 0.5}deg`,
          scale: "0.9",
          filter: "brightness(0.42) saturate(0.5)",
          easing: EASE.power1In,
        },
        {
          offset: 1,
          opacity: 0,
          translate: `${dx}px ${dy}px`,
          rotate: `${rot}deg`,
          scale: "0.5",
          filter: DIMMER,
        },
      ],
      local(0, 1),
    );
  }

  for (const m of all<HTMLElement>(box, ".mote")) {
    const a = num(m, "at");
    run(
      m,
      [
        { translate: "0px 0px", opacity: 0, easing: EASE.power3Out },
        { translate: "0px 0px", opacity: 1, offset: 0.02 },
        { translate: `${num(m, "dx")}px ${num(m, "dy")}px`, opacity: 0 },
      ],
      local(a, Math.min(1, a + 0.5)),
    );
  }
}

// ---- Sacrifice (D5): an altar, a knife, then the shatter ----

function sacrifice(node: HTMLElement, f: CueMotion): void {
  const { stab, shatter: breakAt } = sacrificeBeats(f.light);
  const altar = node.querySelector<HTMLElement>(".altar");
  if (altar) {
    run(
      altar,
      [
        { offset: 0, opacity: 0, translate: "0px 14px", scale: "0.92" },
        { offset: 0.12, opacity: 1, translate: "0px 0px", scale: "1" },
        // Held until the knife has gone, then sinks away (Ian, round 3).
        { offset: breakAt + 0.4, opacity: 1, translate: "0px 0px", scale: "1" },
        { offset: breakAt + 0.6, opacity: 0, translate: "0px 6px", scale: "1" },
        { offset: 1, opacity: 0, translate: "0px 6px", scale: "1" },
      ],
      span(f, 0, 1, EASE.power2Out),
    );
  }
  const copy = node.querySelector<HTMLElement>(".copy:not(.ghost)");
  if (copy) {
    // Laid on the altar, then struck.
    run(
      copy,
      [
        { offset: 0, translate: "0px 0px", scale: "1" },
        { offset: f.light ? 0.02 : 0.12, translate: "0px 3px", scale: "0.96" },
        { offset: stab, translate: "0px 3px", scale: "0.96", easing: EASE.power2Out },
        { offset: stab + 0.025, translate: "0px 6px", scale: "0.94" },
        { offset: stab + 0.07, translate: "0px 3px", scale: "0.96" },
        { offset: 1, translate: "0px 3px", scale: "0.96" },
      ],
      span(f, 0, 1),
    );
  }
  const knife = node.querySelector<HTMLElement>(".knife");
  if (knife) {
    const rise = -Math.round(f.height * 0.62);
    const high = -Math.round(f.height * 0.78);
    const sink = Math.round(f.height * 0.06);
    const show = f.light ? 0.03 : 0.12;
    run(
      knife,
      [
        { offset: 0, opacity: 0, translate: `0px ${high}px`, rotate: "-10deg" },
        { offset: show, opacity: 1, translate: `0px ${rise}px`, rotate: "-4deg" },
        {
          offset: Math.max(show, stab - 0.07),
          opacity: 1,
          translate: `0px ${high}px`,
          rotate: "0deg",
          easing: EASE.power2In,
        },
        { offset: stab, opacity: 1, translate: `0px ${sink}px`, rotate: "0deg" },
        { offset: stab + 0.02, opacity: 1, translate: `1px ${sink + 2}px`, rotate: "0.6deg" },
        { offset: stab + 0.05, opacity: 1, translate: `-1px ${sink}px`, rotate: "-0.4deg" },
        { offset: breakAt + 0.12, opacity: 1, translate: `0px ${sink}px`, rotate: "0deg" },
        { offset: breakAt + 0.26, opacity: 0, translate: `0px ${sink - 10}px`, rotate: "0deg" },
        { offset: 1, opacity: 0, translate: `0px ${sink - 10}px`, rotate: "0deg" },
      ],
      span(f, 0, 1),
    );
    const glint = knife.querySelector<HTMLElement>(".glint");
    if (glint) {
      run(
        glint,
        [
          { offset: 0, opacity: 0, translate: "0px 40%" },
          { offset: 0.5, opacity: 1 },
          { offset: 1, opacity: 0, translate: "0px -40%" },
        ],
        span(f, show, Math.max(show + 0.02, stab - 0.08)),
      );
    }
  }
  shatter(node, f, breakAt, 1);
}

// ---- Exile (D3, D4): into a white void, spinning ----

function exile(node: HTMLElement, f: CueMotion): void {
  const pulled = !!f.pull && f.pull.length > 1;
  const copy = node.querySelector<HTMLElement>(".copy:not(.ghost)");
  if (copy) {
    if (pulled) {
      // Into a board wipe's vortex: round its centre, closing on it.
      const path = f.pull!;
      const n = path.length - 1;
      const frames: Keyframe[] = path.map((p, i) => {
        const t = i / n;
        return {
          offset: 0.04 + t * 0.84,
          translate: `${p.x}px ${p.y}px`,
          rotate: `${Math.round(t * t * 560)}deg`,
          scale: String(Math.max(0.03, 1 - Math.pow(t, 0.85) * 0.97)),
          opacity: t < 0.8 ? 1 : Math.max(0, 1 - (t - 0.8) * 5),
          easing: EASE.power1In,
        };
      });
      // Explicit ends: a keyframe list that stops short of 0 or 1 would
      // ease back to the card's rest at the end.
      frames.unshift({ ...frames[0], offset: 0 });
      frames.push({ ...frames[frames.length - 1], offset: 1 });
      run(copy, frames, span(f, 0, 1));
    } else {
      run(
        copy,
        [
          { offset: 0, scale: "1", rotate: "0deg", opacity: 1 },
          { offset: 0.16, scale: "1.06", rotate: "-16deg", opacity: 1, easing: EASE.power2In },
          { offset: 0.66, scale: "0.32", rotate: "400deg", opacity: 1, easing: EASE.power1In },
          { offset: 0.82, scale: "0.04", rotate: "700deg", opacity: 0.9 },
          { offset: 0.86, scale: "0", rotate: "760deg", opacity: 0 },
          { offset: 1, scale: "0", rotate: "760deg", opacity: 0 },
        ],
        span(f, 0, 1),
      );
    }
  }
  // The card turns to white light as it goes.
  for (const el of all<HTMLElement>(node, ".face, .shatter")) {
    run(
      el,
      [
        { offset: 0, filter: "brightness(1) saturate(1)" },
        { offset: 0.2, filter: "brightness(1.15) saturate(0.85)" },
        { offset: 0.5, filter: "brightness(1.9) saturate(0.3)" },
        { offset: 0.75, filter: "brightness(3.4) saturate(0)" },
        { offset: 1, filter: "brightness(3.4) saturate(0)" },
      ],
      span(f, 0, 1),
    );
  }
  if (f.fromDestroy) crumbleAway(node, f);

  const voidEl = node.querySelector<HTMLElement>(".void");
  if (voidEl) {
    run(
      voidEl,
      [
        { offset: 0, scale: "0.1", opacity: 0, easing: EASE.power3Out },
        { offset: 0.2, scale: "1", opacity: 0.95 },
        { offset: 0.66, scale: "1.1", opacity: 1, easing: EASE.power2In },
        { offset: 0.88, scale: "0.16", opacity: 1 },
        { offset: 1, scale: "0", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  const ring = node.querySelector<HTMLElement>(".vring");
  if (ring) {
    run(
      ring,
      [
        { offset: 0, scale: "1.5", opacity: 0 },
        { offset: 0.2, scale: "1.2", opacity: 0.9 },
        { offset: 0.84, scale: "0.08", opacity: 0.7 },
        { offset: 0.88, scale: "0.04", opacity: 0 },
        { offset: 1, scale: "0.04", opacity: 0 },
      ],
      span(f, 0, 1, EASE.power1In),
    );
  }
  const pinch = node.querySelector<HTMLElement>(".pinch");
  if (pinch) {
    run(
      pinch,
      [
        { offset: 0, scale: "0", opacity: 0 },
        { offset: 0.35, scale: "1", opacity: 1 },
        { offset: 1, scale: "0.2", opacity: 0 },
      ],
      span(f, 0.84, 1, EASE.power2Out),
    );
  }
  for (const orbit of all<HTMLElement>(node, ".orbit")) {
    const a = num(orbit, "at");
    const angle = num(orbit, "angle");
    const speck = orbit.querySelector<HTMLElement>(".speck");
    const end = Math.min(0.86, a + 0.55);
    run(
      orbit,
      [{ rotate: `${angle}deg` }, { rotate: `${angle + num(orbit, "spin")}deg` }],
      span(f, a, end, EASE.power1In),
    );
    if (speck) {
      run(
        speck,
        [
          { translate: `${num(orbit, "r")}px 0px`, opacity: 0 },
          { translate: `${num(orbit, "r") * 0.8}px 0px`, opacity: 1, offset: 0.2 },
          { translate: "0px 0px", opacity: 0 },
        ],
        span(f, a, end, EASE.power2In),
      );
    }
  }
}

/**
 * crumbleAway is D4's slight crumble: the card cracks at once, and its
 * pieces drift a little apart as the void takes it, a few flakes falling.
 */
function crumbleAway(node: HTMLElement, f: CueMotion): void {
  const box = node.querySelector<HTMLElement>(".shatter");
  if (!box) return;
  const cracks = box.querySelector<HTMLElement>(".crk");
  if (cracks) {
    run(
      cracks,
      [{ opacity: 1 }, { opacity: 1, offset: 0.4 }, { opacity: 0, offset: 0.7 }, { opacity: 0 }],
      span(f, 0, 1),
    );
    for (const p of all<SVGPathElement>(cracks, "path")) {
      run(p, [{ strokeDashoffset: 1 }, { strokeDashoffset: 0 }], span(f, 0, 0.14, EASE.power2Out));
    }
  }
  for (const s of all<HTMLElement>(box, ".shard")) {
    const dx = num(s, "dx") * 0.12;
    const dy = num(s, "dy") * 0.12;
    const rot = num(s, "rot") * 0.1;
    run(
      s,
      [
        { offset: 0, translate: "0px 0px", rotate: "0deg" },
        { offset: 0.14, translate: "0px 0px", rotate: "0deg", easing: EASE.power2Out },
        { offset: 0.7, translate: `${dx}px ${dy}px`, rotate: `${rot}deg` },
        { offset: 1, translate: `${dx}px ${dy}px`, rotate: `${rot}deg` },
      ],
      span(f, 0, 1),
    );
  }
  for (const m of all<HTMLElement>(box, ".mote")) {
    const a = 0.12 + num(m, "at") * 0.5;
    run(
      m,
      [
        { translate: "0px 0px", opacity: 0 },
        { translate: "0px 0px", opacity: 1, offset: 0.05 },
        { translate: `${num(m, "dx") * 0.4}px ${Math.abs(num(m, "dy")) * 0.6}px`, opacity: 0 },
      ],
      span(f, a, Math.min(0.8, a + 0.4), EASE.power1In),
    );
  }
}

// ---- Bounce (D6): lift, a flip, shown to the table, then home ----

function bounce(node: HTMLElement, f: CueMotion): void {
  // Up, turned over and held up to the table (its identity is public
  // now), then home to its owner's hand, still face up.
  const path: Keyframe[] = [
    { offset: 0, translate: "0px 0px", scale: "1", easing: EASE.power2Out },
    { offset: 0.22, translate: "0px -16px", scale: "1.12" },
    { offset: 0.48, translate: "0px -18px", scale: "1.12", easing: EASE.power2InOut },
    { offset: 0.94, translate: `${f.fly.x}px ${f.fly.y}px`, scale: "0.42" },
    { offset: 1, translate: `${f.fly.x}px ${f.fly.y}px`, scale: "0.4" },
  ];
  const copy = node.querySelector<HTMLElement>(".copy:not(.ghost)");
  if (copy) run(copy, path, span(f, 0, 1));
  const flip = node.querySelector<HTMLElement>(".flip");
  if (flip) {
    run(
      flip,
      [
        // End over end, along the card's long axis.
        { transform: "perspective(700px) rotateX(0deg)" },
        { transform: "perspective(700px) rotateX(360deg)" },
      ],
      span(f, 0.03, 0.4, EASE.power2InOut),
    );
  }
  // The fade is on the copy, not the flip: opacity on the flip itself
  // would flatten it, and the card's face would show through its back.
  if (copy) {
    run(copy, [{ opacity: 1 }, { opacity: 1, offset: 0.88 }, { opacity: 0 }], span(f, 0, 1));
  }
  const halo = node.querySelector<HTMLElement>(".halo");
  if (halo) {
    run(
      halo,
      [
        { offset: 0, opacity: 0 },
        { offset: 0.3, opacity: 0 },
        { offset: 0.42, opacity: 1 },
        { offset: 0.62, opacity: 0.85 },
        { offset: 0.85, opacity: 0 },
        { offset: 1, opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  // Two fading after-images trail it home.
  all<HTMLElement>(node, ".ghost").forEach((g, i) => {
    const peak = i === 0 ? 0.32 : 0.16;
    const lag = (i + 1) * 0.035;
    run(g, path, span(f, lag, 1));
    const face = g.querySelector<HTMLElement>(".face");
    if (face) {
      run(
        face,
        [
          { opacity: 0 },
          { opacity: 0, offset: 0.5 },
          { opacity: peak, offset: 0.56 },
          { opacity: peak, offset: 0.84 },
          { opacity: 0 },
        ],
        span(f, lag, 1),
      );
    }
  });
}

// ---- Counterspell (D7): a palm stops it, and it cracks ----

function counter(node: HTMLElement, f: CueMotion): void {
  // The palm lands at `hit`; the stack linger already dims the spell
  // under its "Countered" badge, so nothing here covers it.
  const hit = 0.13;
  const palm = node.querySelector<HTMLElement>(".palm");
  if (palm) {
    run(
      palm,
      [
        { offset: 0, scale: "1.75", opacity: 0, easing: EASE.power2In },
        { offset: hit, scale: "0.93", opacity: 0.95, easing: EASE.power2Out },
        { offset: hit + 0.05, scale: "1", opacity: 0.92 },
        { offset: 0.46, scale: "1", opacity: 0.88, easing: EASE.power1In },
        { offset: 0.64, scale: "1.12", opacity: 0 },
        { offset: 1, scale: "1.12", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  const shock = node.querySelector<HTMLElement>(".shock");
  if (shock) {
    run(
      shock,
      [
        { offset: 0, scale: "0.3", opacity: 0 },
        { offset: hit, scale: "0.32", opacity: 0 },
        { offset: hit + 0.01, scale: "0.35", opacity: 0.95, easing: EASE.power2Out },
        { offset: 0.44, scale: "1.8", opacity: 0 },
        { offset: 1, scale: "1.8", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  const flash = node.querySelector<HTMLElement>(".flash");
  if (flash) {
    run(
      flash,
      [
        { offset: 0, scale: "0.4", opacity: 0 },
        { offset: hit, scale: "0.4", opacity: 0 },
        { offset: hit + 0.03, scale: "1", opacity: 0.9, easing: EASE.power2Out },
        { offset: 0.42, scale: "1.4", opacity: 0 },
        { offset: 1, scale: "1.4", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  const cracks = node.querySelector<HTMLElement>(".cracks");
  if (cracks) {
    // They stay on the spell while it leaves the stack (the stack linger
    // shows it for 1.5 s), and go with it.
    run(
      cracks,
      [
        { offset: 0, opacity: 1 },
        { offset: 0.86, opacity: 1 },
        { offset: 1, opacity: 0 },
      ],
      span(f, 0, 1),
    );
    for (const p of all<SVGPathElement>(cracks, "path")) {
      run(
        p,
        [{ strokeDashoffset: 1 }, { strokeDashoffset: 0 }],
        span(f, hit + 0.01, hit + 0.2, EASE.power2Out),
      );
    }
  }
  for (const el of all<HTMLElement>(node, ".glass")) {
    run(
      el,
      [
        { offset: 0, translate: "0px 0px", rotate: "0deg", opacity: 0 },
        { offset: 0.04, translate: "0px 0px", rotate: "0deg", opacity: 1, easing: EASE.power2Out },
        {
          offset: 1,
          translate: `${num(el, "dx")}px ${num(el, "dy")}px`,
          rotate: `${num(el, "rot")}deg`,
          opacity: 0,
        },
      ],
      span(f, hit, 0.62),
    );
  }
}

// ---- A spell fizzles (#4450): a soft dissolve, no palm and no glass ----

/**
 * fizzle thins the stack item away: a cool grey haze settles over it and
 * lifts, its edge glows faintly and fades, and specks come off it from
 * the top down and drift up as they go. The stack linger dims the spell
 * under its "Fizzled" badge, so nothing here covers it for long.
 */
function fizzle(node: HTMLElement, f: CueMotion): void {
  const mist = node.querySelector<HTMLElement>(".fmist");
  if (mist) {
    run(
      mist,
      [
        { offset: 0, translate: "0px 0px", opacity: 0 },
        { offset: 0.25, translate: "0px 0px", opacity: 0.7, easing: EASE.power1In },
        { offset: 1, translate: `0px ${-Math.round(f.height * 0.12)}px`, opacity: 0 },
      ],
      span(f, 0, 1, EASE.power1Out),
    );
  }
  const edge = node.querySelector<HTMLElement>(".fedge");
  if (edge) {
    run(
      edge,
      [
        { offset: 0, scale: "1", opacity: 0 },
        { offset: 0.15, scale: "1", opacity: 0.85, easing: EASE.power1In },
        { offset: 0.7, scale: "1.04", opacity: 0 },
        { offset: 1, scale: "1.04", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  for (const el of all<HTMLElement>(node, ".fmote")) {
    const at = num(el, "at");
    run(
      el,
      [
        { offset: 0, translate: "0px 0px", opacity: 0 },
        { offset: 0.12, translate: "0px 0px", opacity: 0.95, easing: EASE.power2Out },
        { offset: 1, translate: `${num(el, "dx")}px ${num(el, "dy")}px`, opacity: 0 },
      ],
      span(f, at, at + 0.5, EASE.power1Out),
    );
  }
}

// ---- A table-wide drain (#4450): a thin dark line to each player ----

/**
 * drainTable draws each line from the source out to its player (the
 * stroke runs along it, over the first half), holds it a beat and lets
 * it go, so the whole thing ends inside the cue's 0.8 s. The lines start
 * one after another a little apart. Display only.
 */
function drainTable(node: HTMLElement, f: CueMotion): void {
  const lines = all<SVGElement>(node, ".drain-line");
  lines.forEach((el, i) => {
    const a = Math.min(0.2, i * 0.06);
    run(
      el as unknown as HTMLElement,
      [
        { offset: 0, strokeDashoffset: "1", opacity: 0 },
        { offset: 0.12, strokeDashoffset: "0.85", opacity: 0.9, easing: EASE.power1Out },
        { offset: 0.5, strokeDashoffset: "0", opacity: 0.9 },
        { offset: 1, strokeDashoffset: "0", opacity: 0 },
      ],
      span(f, a, 1),
    );
  });
  const dot = node.querySelector<HTMLElement>(".drain-source");
  if (dot) {
    run(
      dot,
      [
        { offset: 0, scale: "0.4", opacity: 0 },
        { offset: 0.2, scale: "1", opacity: 0.9, easing: EASE.power1Out },
        { offset: 1, scale: "1.6", opacity: 0 },
      ],
      span(f, 0, 0.7),
    );
  }
}

// ---- The crown or the initiative changes hands (#4450): a short glint ----

/**
 * crownGlint rings the new holder's panel with a soft light that opens
 * and fades, and flicks four small sparks out from its centre, so the
 * marker that has just moved there is noticed. Gold for the crown,
 * violet for the initiative (styled by the cue's `data-effect`).
 */
function crownGlint(node: HTMLElement, f: CueMotion): void {
  const ring = node.querySelector<HTMLElement>(".glint-ring");
  if (ring) {
    run(
      ring,
      [
        { offset: 0, scale: "0.7", opacity: 0 },
        { offset: 0.2, scale: "0.9", opacity: 0.95, easing: EASE.power2Out },
        { offset: 1, scale: "1.15", opacity: 0 },
      ],
      span(f, 0, 1, EASE.sineInOut),
    );
  }
  all<HTMLElement>(node, ".glint-spark").forEach((el, i) => {
    run(
      el,
      [
        { offset: 0, translate: "0px 0px", scale: "0.3", opacity: 0 },
        { offset: 0.3, scale: "1", opacity: 1, easing: EASE.power2Out },
        { offset: 1, translate: `${num(el, "dx")}px ${num(el, "dy")}px`, scale: "0.2", opacity: 0 },
      ],
      span(f, 0.05 + i * 0.05, 0.85, EASE.power1Out),
    );
  });
}

// ---- A planeswalker that dies (D9): cracks, then leaves the plane ----

function pwLeave(node: HTMLElement, f: CueMotion): void {
  const copy = node.querySelector<HTMLElement>(".copy:not(.ghost)");
  if (copy) {
    run(
      copy,
      [
        { offset: 0, translate: "0px 0px", scale: "1", opacity: 1 },
        { offset: 0.1, translate: "-3px 0px", scale: "1", opacity: 1 },
        { offset: 0.16, translate: "3px 0px", scale: "1", opacity: 1 },
        { offset: 0.22, translate: "-2px 0px", scale: "1", opacity: 1 },
        { offset: 0.3, translate: "0px 0px", scale: "1", opacity: 1, easing: EASE.power2Out },
        { offset: 0.44, translate: "0px -6px", scale: "1.05", opacity: 1, easing: EASE.power2In },
        { offset: 0.8, translate: "0px -30px", scale: "0.2", opacity: 0.85 },
        { offset: 0.9, translate: "0px -34px", scale: "0.06", opacity: 0 },
        { offset: 1, translate: "0px -34px", scale: "0.06", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  const face = node.querySelector<HTMLElement>(".face");
  if (face) {
    run(
      face,
      [
        { offset: 0, filter: "brightness(1) saturate(1)" },
        { offset: 0.12, filter: "brightness(2) saturate(1)" },
        { offset: 0.28, filter: "brightness(1.05) saturate(0.9)" },
        { offset: 0.5, filter: "brightness(1.4) saturate(0.7)" },
        { offset: 0.85, filter: "brightness(2.8) saturate(0.2)" },
        { offset: 1, filter: "brightness(2.8) saturate(0.2)" },
      ],
      span(f, 0, 1),
    );
  }
  const crack = node.querySelector<HTMLElement>(".crack");
  if (crack) {
    run(
      crack,
      [{ opacity: 1 }, { opacity: 1, offset: 0.42 }, { opacity: 0, offset: 0.6 }, { opacity: 0 }],
      span(f, 0, 1),
    );
    for (const p of all<SVGPathElement>(crack, "path")) {
      run(
        p,
        [{ strokeDashoffset: 1 }, { strokeDashoffset: 0 }],
        span(f, 0.02, 0.16, EASE.power2Out),
      );
    }
  }
  const rift = node.querySelector<HTMLElement>(".rift");
  if (rift) {
    run(
      rift,
      [
        { offset: 0, scale: "0.02 0.2", opacity: 0 },
        { offset: 0.34, scale: "0.02 0.2", opacity: 0, easing: EASE.power3Out },
        { offset: 0.5, scale: "1 1", opacity: 0.95 },
        { offset: 0.8, scale: "0.9 0.9", opacity: 0.9, easing: EASE.power2In },
        { offset: 0.94, scale: "0.02 0.5", opacity: 0.6 },
        { offset: 1, scale: "0.02 0.4", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
}

// ---- A planeswalker that arrives (D8): unchanged, a ring and a glow ----

function pwEnter(node: HTMLElement, f: CueMotion): void {
  const ring = node.querySelector<HTMLElement>(".ring");
  if (ring) {
    run(
      ring,
      [
        { offset: 0, scale: "0.85", opacity: 0 },
        { offset: 0.3, scale: "1", opacity: 1 },
        { offset: 1, scale: "1.18", opacity: 0 },
      ],
      span(f, 0, 1, EASE.power2Out),
    );
  }
  const glow = node.querySelector<HTMLElement>(".glow");
  if (glow) {
    run(
      glow,
      [
        { offset: 0, scale: "0.6", opacity: 0 },
        { offset: 0.35, scale: "1", opacity: 0.9 },
        { offset: 1, scale: "1.4", opacity: 0 },
      ],
      span(f, 0, 1, EASE.power2Out),
    );
  }
}

// ---- A token is created (#4441): a small pop-in with a soft glow ----

/** The token's own scale as its pop-in starts: about 80%. */
export const TOKEN_POP_FROM = 0.8;

/**
 * tokenPop pops the new token's own tile in: from about 80% and hidden,
 * up a touch past full size and bright, then settling at rest. It is the
 * one cue that moves a live tile, because a batch's tokens are on the
 * board at once and only the tile can wait for its turn: `fill:
 * "backwards"` holds a later token hidden through its stagger, and once
 * the pop is over the tile is its own style again (no end state is kept).
 * The caller cancels the returned run when the cue is removed or skipped.
 */
export function tokenPop(tile: HTMLElement, f: CueMotion): Run {
  return run(
    tile,
    [
      { offset: 0, scale: String(TOKEN_POP_FROM), opacity: 0, filter: "brightness(1.6)" },
      { offset: 0.55, scale: "1.04", opacity: 1, filter: "brightness(1.3)" },
      { offset: 1, scale: "1", opacity: 1, filter: "brightness(1)" },
    ],
    { delay: f.delay, duration: Math.max(1, f.ms), easing: EASE.power2Out, fill: "backwards" },
  );
}

/** The soft glow that blooms over the token as it pops in, and fades. */
function tokenGlow(node: HTMLElement, f: CueMotion): void {
  const glow = node.querySelector<HTMLElement>(".tglow");
  if (!glow) return;
  run(
    glow,
    [
      { offset: 0, scale: "0.7", opacity: 0 },
      { offset: 0.4, scale: "1", opacity: 0.85 },
      { offset: 1, scale: "1.2", opacity: 0 },
    ],
    span(f, 0, 1, EASE.power2Out),
  );
}

// ---- Counters are put on a permanent (#4442): a pulse and a tick ----

/** How far a counter cue's tick travels, px. */
export const COUNTER_TICK_PX = 14;

/**
 * counterPulse rings the permanent's power/toughness badge (or its
 * counter's pip) once, brightening and fading, while a small tick above
 * it rises for +1/+1 counters and falls for -1/-1 ones. Other counters
 * have the ring alone.
 */
function counterPulse(node: HTMLElement, f: CueMotion): void {
  const ring = node.querySelector<HTMLElement>(".cpulse");
  if (ring) {
    run(
      ring,
      [
        { offset: 0, scale: "0.85", opacity: 0 },
        { offset: 0.3, scale: "1.12", opacity: 1 },
        { offset: 1, scale: "1.3", opacity: 0 },
      ],
      span(f, 0, 1, EASE.power2Out),
    );
  }
  const tick = node.querySelector<HTMLElement>(".ctick");
  if (!tick) return;
  const up = f.counterTone !== "minus";
  const from = up ? COUNTER_TICK_PX / 2 : -COUNTER_TICK_PX / 2;
  const to = up ? -COUNTER_TICK_PX / 2 : COUNTER_TICK_PX / 2;
  run(
    tick,
    [
      { offset: 0, translate: `0px ${from}px`, opacity: 0 },
      { offset: 0.25, opacity: 1 },
      { offset: 1, translate: `0px ${to}px`, opacity: 0 },
    ],
    span(f, 0, 1, EASE.power2Out),
  );
}

// ---- A commander is cast from the command zone, or goes home (#4443) ----

/** The share of a commander cast at which its light reaches the stack, and the crown blooms. */
export const COMMANDER_ARRIVE = 0.45;
/** The share of a commander's drift home at which it lands on its command-zone card. */
export const COMMANDER_LAND = 0.9;

/**
 * commanderRise is a commander cast (CR 903.8): a golden bloom where it
 * sat in the command zone, a beam of light that rises from there to the
 * spell on the stack with motes riding it, and a crown of light that
 * blooms on the spell as the beam arrives. CastFlight still flies the
 * card itself; this is the light around it.
 */
function commanderRise(node: HTMLElement, f: CueMotion): void {
  const bloom = node.querySelector<HTMLElement>(".cmd-bloom");
  if (bloom) {
    run(
      bloom,
      [
        { offset: 0, scale: "0.5", opacity: 0 },
        { offset: 0.25, scale: "1", opacity: 0.9 },
        { offset: 1, scale: "1.25", opacity: 0 },
      ],
      span(f, 0, 0.6, EASE.power2Out),
    );
  }
  const beam = node.querySelector<HTMLElement>(".cmd-beam");
  if (beam) {
    run(
      beam,
      [
        { offset: 0, scale: "1 0", opacity: 0, easing: EASE.power1Out },
        // Full length as the cue reaches COMMANDER_ARRIVE (the beam starts at 0.05).
        {
          offset: (COMMANDER_ARRIVE - 0.05) / 0.95,
          scale: "1 1",
          opacity: 0.95,
          easing: EASE.power1In,
        },
        { offset: 1, scale: "1 1", opacity: 0 },
      ],
      span(f, 0.05, 1),
    );
  }
  // The motes ride up the beam, each a little off it, a beat apart.
  const len = Math.hypot(f.fly.x, f.fly.y) || 1;
  const px = -f.fly.y / len;
  const py = f.fly.x / len;
  const motes = all<HTMLElement>(node, ".cmd-mote");
  motes.forEach((m, i) => {
    const reach = 0.55 + (0.45 * ((i * 3) % motes.length)) / Math.max(1, motes.length);
    const side = (i % 2 === 0 ? -1 : 1) * (5 + ((i * 7) % 9));
    const x = f.fly.x * reach + px * side;
    const y = f.fly.y * reach + py * side;
    const a = 0.05 + i * 0.05;
    run(
      m,
      [
        { offset: 0, translate: "0px 0px", scale: "0.6", opacity: 0 },
        { offset: 0.2, opacity: 1 },
        { offset: 1, translate: `${x}px ${y}px`, scale: "1", opacity: 0 },
      ],
      span(f, a, a + 0.55, EASE.power2Out),
    );
  });
  const crown = node.querySelector<HTMLElement>(".cmd-crown");
  if (crown) {
    run(
      crown,
      [
        { offset: 0, scale: "0.7", opacity: 0 },
        { offset: 0.35, scale: "1.05", opacity: 1 },
        { offset: 1, scale: "1.3", opacity: 0 },
      ],
      span(f, COMMANDER_ARRIVE, 1, EASE.power2Out),
    );
  }
}

/**
 * The path of a commander drifting home: a small lift, then a gentle arc
 * to its card in the command zone, straightening out of a tap and
 * shrinking to that card's size on the way.
 */
function driftPath(f: CueMotion): Keyframe[] {
  const s = f.commandScale ?? 1;
  const r = -(f.rot ?? 0);
  const arc = Math.min(60, 20 + Math.hypot(f.fly.x, f.fly.y) * 0.12);
  const end = `${f.fly.x}px ${f.fly.y}px`;
  return [
    { offset: 0, translate: "0px 0px", scale: "1", rotate: "0deg", easing: EASE.power1Out },
    { offset: 0.14, translate: "0px -8px", scale: "1.05", rotate: "0deg", easing: EASE.sineInOut },
    {
      offset: 0.55,
      translate: `${f.fly.x / 2}px ${f.fly.y / 2 - arc}px`,
      scale: String((1.05 + s) / 2),
      rotate: `${r / 2}deg`,
      easing: EASE.sineInOut,
    },
    { offset: COMMANDER_LAND, translate: end, scale: String(s), rotate: `${r}deg` },
    { offset: 1, translate: end, scale: String(s), rotate: `${r}deg` },
  ];
}

/**
 * commanderDrift is a commander going home (CR 903.9a, 903.9b): its copy
 * drifts back to its card in the command zone with a soft golden trail,
 * two after-images and a faint streak behind it, and a soft halo blooms
 * on the card as it lands.
 */
function commanderDrift(node: HTMLElement, f: CueMotion): void {
  const path = driftPath(f);
  const copy = node.querySelector<HTMLElement>(".cmd-copy");
  if (copy) {
    run(copy, path, span(f, 0, 1));
    run(
      copy,
      [
        { offset: 0, opacity: 1 },
        { offset: COMMANDER_LAND - 0.04, opacity: 1 },
        { offset: 1, opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  all<HTMLElement>(node, ".cmd-ghost").forEach((g, i) => {
    const lag = (i + 1) * 0.09;
    const peak = i === 0 ? 0.4 : 0.22;
    run(g, path, span(f, lag, 1));
    run(
      g,
      [
        { offset: 0, opacity: 0 },
        { offset: 0.18, opacity: peak },
        { offset: 0.8, opacity: peak },
        { offset: 1, opacity: 0 },
      ],
      span(f, lag, 1),
    );
  });
  const streak = node.querySelector<HTMLElement>(".cmd-streak");
  if (streak) {
    run(
      streak,
      [
        { offset: 0, scale: "1 0", opacity: 0 },
        { offset: 0.2, opacity: 0.55 },
        { offset: 0.8, scale: "1 1", opacity: 0.4 },
        { offset: 1, scale: "1 1", opacity: 0 },
      ],
      span(f, 0.1, 1, EASE.sineInOut),
    );
  }
  const home = node.querySelector<HTMLElement>(".cmd-home");
  if (home) {
    run(
      home,
      [
        { offset: 0, scale: "0.8", opacity: 0 },
        { offset: 0.4, scale: "1.08", opacity: 0.9 },
        { offset: 1, scale: "1.25", opacity: 0 },
      ],
      span(f, COMMANDER_LAND - 0.1, 1, EASE.power2Out),
    );
  }
}

/**
 * commandArrive holds the commander's own card in the command zone
 * hidden until its copy lands there, then shows it: so the copy is the
 * one card on screen on its way home. `fill: "backwards"` hides it from
 * the frame on, and once the run is over the card is its own style again.
 * The caller cancels the returned run when the cue is removed or skipped,
 * which shows the card at once.
 */
export function commandArrive(card: HTMLElement, f: CueMotion): Run {
  return run(
    card,
    [
      { offset: 0, opacity: 0 },
      { offset: COMMANDER_LAND, opacity: 0 },
      { offset: 1, opacity: 1 },
    ],
    { delay: f.delay, duration: Math.max(1, f.ms), easing: EASE.linear, fill: "backwards" },
  );
}

// ---- An Equipment is attached, or an Aura enchants a permanent (#4444) ----

/** The share of an attach cue at which the attachment reaches its host, and clasps. */
export const ATTACH_LAND = 0.6;
/** The most an attachment slides past its rest before it settles back, px. */
export const ATTACH_OVERSHOOT_PX = 4;

/**
 * attachSlide moves the attachment's own tile onto its host: from where it
 * was (its last box on the battlefield, or just above and to the side of
 * its rest when it is new there), sliding in with a slight tilt that
 * straightens, a touch past its rest at ATTACH_LAND, then a small settle.
 * The board already draws it tucked under the host, so the host stays on
 * top the whole way. Like tokenPop it moves a live tile, presentation
 * only: `fill: "backwards"` holds it at the start from the frame on, and
 * once the slide is over the tile is its own style again. Only translate
 * and rotate move, so a token's pop (scale, opacity) on the same tile
 * plays with it. The caller cancels the run when the cue is removed or
 * skipped, which puts the tile at rest at once.
 */
export function attachSlide(tile: HTMLElement, f: CueMotion): Run {
  const len = Math.hypot(f.fly.x, f.fly.y) || 1;
  const over = Math.min(ATTACH_OVERSHOOT_PX, len * 0.06);
  const ox = (-f.fly.x / len) * over;
  const oy = (-f.fly.y / len) * over;
  // An Equipment swings in with more tilt, and drops a little harder.
  const metal = f.kind === "attach_equip";
  const tilt = (f.fly.x <= 0 ? -1 : 1) * (metal ? 5 : 3);
  const drop = metal ? 1.5 : 0.5;
  return run(
    tile,
    [
      {
        offset: 0,
        translate: `${f.fly.x}px ${f.fly.y}px`,
        rotate: `${tilt}deg`,
        easing: EASE.power2Out,
      },
      {
        offset: ATTACH_LAND,
        translate: `${ox}px ${oy}px`,
        rotate: `${-tilt * 0.2}deg`,
        easing: EASE.sineInOut,
      },
      {
        offset: 0.8,
        translate: `${-ox * 0.35}px ${-oy * 0.35 + drop}px`,
        rotate: "0deg",
        easing: EASE.sineInOut,
      },
      { offset: 1, translate: "0px 0px", rotate: "0deg" },
    ],
    { delay: f.delay, duration: Math.max(1, f.ms), easing: EASE.linear, fill: "backwards" },
  );
}

/**
 * attachClasp is an Equipment's clang as it reaches its host: a hard
 * white-silver flash where the two meet, a fan of sparks thrown out from
 * there, and a quick glint along the host's edge.
 */
function attachClasp(node: HTMLElement, f: CueMotion): void {
  const flash = node.querySelector<HTMLElement>(".att-flash");
  if (flash) {
    run(
      flash,
      [
        { offset: 0, scale: "0.3", opacity: 0 },
        { offset: 0.25, scale: "1.1", opacity: 1 },
        { offset: 1, scale: "1.4", opacity: 0 },
      ],
      span(f, ATTACH_LAND - 0.04, 1, EASE.power2Out),
    );
  }
  const sparks = all<HTMLElement>(node, ".att-spark");
  sparks.forEach((s, i) => {
    // Out from the seam, away from the host: a fan to the left.
    const a = ((180 + (i - (sparks.length - 1) / 2) * 28) * Math.PI) / 180;
    const d = 20 + ((i * 7) % 11);
    run(
      s,
      [
        { offset: 0, translate: "0px 0px", rotate: `${(a * 180) / Math.PI}deg`, opacity: 0 },
        { offset: 0.15, opacity: 1 },
        {
          offset: 1,
          translate: `${Math.cos(a) * d}px ${Math.sin(a) * d}px`,
          rotate: `${(a * 180) / Math.PI}deg`,
          opacity: 0,
        },
      ],
      span(f, ATTACH_LAND - 0.02, 1, EASE.power2Out),
    );
  });
  const rim = node.querySelector<HTMLElement>(".att-rim");
  if (rim) {
    run(
      rim,
      [
        { offset: 0, opacity: 0 },
        { offset: 0.3, opacity: 0.8 },
        { offset: 1, opacity: 0 },
      ],
      span(f, ATTACH_LAND, 1, EASE.power1Out),
    );
  }
}

/**
 * attachShimmer is an Aura settling on its host: a soft light rings the
 * host as the Aura slides in, and a few glints twinkle around its edge
 * as it lands, then all of it fades.
 */
function attachShimmer(node: HTMLElement, f: CueMotion): void {
  const halo = node.querySelector<HTMLElement>(".att-halo");
  if (halo) {
    run(
      halo,
      [
        { offset: 0, scale: "0.97", opacity: 0 },
        { offset: 0.45, scale: "1.02", opacity: 0.85 },
        { offset: 1, scale: "1.06", opacity: 0 },
      ],
      span(f, 0.2, 1, EASE.sineInOut),
    );
  }
  all<HTMLElement>(node, ".att-glint").forEach((g, i) => {
    const a = ATTACH_LAND - 0.25 + i * 0.06;
    run(
      g,
      [
        { offset: 0, translate: "0px 0px", scale: "0", opacity: 0 },
        { offset: 0.4, scale: "1", opacity: 1 },
        { offset: 1, translate: "0px -6px", scale: "0.2", opacity: 0 },
      ],
      span(f, a, Math.min(1, a + 0.45), EASE.power1Out),
    );
  });
}

// ---- Cards milled or discarded into a graveyard (#4445), or returned
// from one (#4446) ----

/**
 * pileWindows is when each of `n` cards slides, as shares of the cue: each
 * slide is the same length, and the next sets off PILE_STAGGER into the
 * one before, so the last lands as the cue ends. A sweep's cards all go
 * at once.
 */
export function pileWindows(n: number, sweep = false): { a: number; b: number }[] {
  const k = Math.max(1, n);
  if (sweep) return Array.from({ length: k }, () => ({ a: 0, b: 1 }));
  const len = 1 / (1 + PILE_STAGGER * (k - 1));
  return Array.from({ length: k }, (_, i) => {
    const a = i * PILE_STAGGER * len;
    return { a, b: a + len };
  });
}

/**
 * pileLandAt is the share of a cue of `n` cards at which card `i` lands
 * (PILE_LAND into its own slide): when a returned card's tile, held
 * hidden until then, takes over from its copy (#4446).
 */
export function pileLandAt(n: number, i: number, sweep = false): number {
  const w = pileWindows(n, sweep)[i] ?? { a: 0, b: 1 };
  return w.a + (w.b - w.a) * PILE_LAND;
}

/** How far a returned card rises straight up out of the graveyard pile before it travels, px. */
export const REANIMATE_RISE_PX = 26;

/**
 * pileSlide moves cards into their owner's graveyard pile, or out of it.
 * A mill slides each card off the library on a low arc, turning it face
 * up on the way; a discard tosses it from the hand on a higher arc with
 * a little spin. Each lands a touch past the pile and settles back, and
 * fades as it lands, since the pile's own top card is already the last
 * of them. A puff of dust rises from the pile as the last card lands. A
 * sweep (five cards or more) moves a fanned handful as one, which closes
 * up on the way, with a faint streak behind it.
 *
 * A reanimate (#4446) is the mirror image: each card rises straight up
 * out of its graveyard pile in a dark-green glow, which blooms from the
 * pile as it leaves, then travels on its own path (`paths`) to where it
 * went (its tile on the battlefield, growing to the tile's size, or a
 * hand, turning face down on the way into a hand the viewer cannot see),
 * lands a touch past and settles, and fades as its tile takes over.
 */
function pileSlide(node: HTMLElement, f: CueMotion): void {
  const cards = all<HTMLElement>(node, ".pile-card");
  const toss = f.kind === "discard";
  const rise = f.kind === "reanimate";
  const sweep = f.sweep === true;
  const wins = pileWindows(cards.length, sweep);
  cards.forEach((card, i) => {
    const w = wins[i];
    const p = f.paths?.[i];
    const dx = p?.dx ?? f.fly.x;
    const dy = p?.dy ?? f.fly.y;
    const s = p?.s ?? 1;
    const len = Math.hypot(dx, dy) || 1;
    // A hop even between two piles side by side, so the slide reads.
    const lift = rise
      ? Math.min(60, 20 + len * 0.12)
      : toss
        ? Math.min(70, 24 + len * 0.18)
        : Math.min(40, 18 + len * 0.08);
    // Past where it lands, along the way in, then back.
    const ox = (dx / len) * 3;
    const oy = (dy / len) * 3;
    // A sweep's handful is fanned at the start and closes on the way.
    const fan = sweep ? i - (cards.length - 1) / 2 : 0;
    const sx = fan * 7;
    const sy = -Math.abs(fan) * 3;
    const spin = toss ? (i % 2 === 0 ? 10 : -8) : 2;
    const tilt = sweep ? fan * 7 : toss ? -8 : rise ? 0 : -3;
    const start: Keyframe = {
      offset: 0,
      translate: `${sx}px ${sy}px`,
      rotate: `${tilt}deg`,
      scale: "1",
      easing: rise ? EASE.power2Out : EASE.power1Out,
    };
    const land: Keyframe = {
      offset: PILE_LAND,
      translate: `${dx + ox}px ${dy + oy}px`,
      rotate: "0deg",
      scale: String(s),
      easing: EASE.sineInOut,
    };
    const end: Keyframe = {
      offset: 1,
      translate: `${dx}px ${dy}px`,
      rotate: "0deg",
      scale: String(s),
    };
    const way: Keyframe[] = rise
      ? [
          // Up out of the pile first, a little larger, then on its way.
          {
            offset: 0.26,
            translate: `${sx}px ${sy - REANIMATE_RISE_PX}px`,
            rotate: "0deg",
            scale: "1.08",
            easing: EASE.sineInOut,
          },
          {
            offset: 0.56,
            translate: `${dx * 0.5 + sx * 0.5}px ${dy * 0.5 - lift + sy * 0.5}px`,
            rotate: `${spin}deg`,
            scale: String((1.08 + s) / 2),
            easing: EASE.power2In,
          },
        ]
      : [
          {
            offset: 0.48,
            translate: `${dx / 2 + sx * 0.5}px ${dy / 2 - lift + sy * 0.5}px`,
            rotate: `${spin}deg`,
            scale: toss ? "1.06" : "1.02",
            easing: EASE.power2In,
          },
        ];
    run(card, [start, ...way, land, end], span(f, w.a, w.b));
    run(
      card,
      [
        { offset: 0, opacity: 0 },
        { offset: 0.06, opacity: 1 },
        { offset: PILE_LAND + 0.04, opacity: 1 },
        { offset: 1, opacity: 0 },
      ],
      span(f, w.a, w.b),
    );
    // A milled card leaves the library face down and lands face up; a
    // card returned to a hand the viewer cannot see leaves the graveyard
    // face up and goes into the hand face down (#4446).
    const flip = card.querySelector<HTMLElement>(".flip");
    if (flip) {
      const down = flip.dataset.flip === "down";
      run(
        flip,
        [
          { transform: `perspective(600px) rotateX(${down ? 0 : 180}deg)` },
          { transform: `perspective(600px) rotateX(${down ? 180 : 360}deg)` },
        ],
        down
          ? span(f, w.a + (w.b - w.a) * 0.4, w.a + (w.b - w.a) * 0.8, EASE.power2InOut)
          : span(f, w.a + (w.b - w.a) * 0.12, w.a + (w.b - w.a) * 0.62, EASE.power2InOut),
      );
    }
    // A returned card's glow, about it as it rises, fading as it lands.
    const aura = card.querySelector<HTMLElement>(".rise-aura");
    if (aura) {
      run(
        aura,
        [
          { offset: 0, opacity: 0 },
          { offset: 0.22, opacity: 0.95 },
          { offset: 0.6, opacity: 0.7 },
          { offset: PILE_LAND, opacity: 0.35 },
          { offset: 1, opacity: 0 },
        ],
        span(f, w.a, w.b, EASE.sineInOut),
      );
    }
  });
  // And the glow a returned card leaves behind: a bloom from its pile as
  // it rises out of it.
  all<HTMLElement>(node, ".rise-glow").forEach((glow, i) => {
    const w = wins[i] ?? { a: 0, b: 1 };
    run(
      glow,
      [
        { offset: 0, scale: "0.5", opacity: 0 },
        { offset: 0.35, scale: "1", opacity: 0.9 },
        { offset: 1, scale: "1.35", opacity: 0 },
      ],
      span(f, w.a, w.a + (w.b - w.a) * 0.6, EASE.power2Out),
    );
  });
  const streak = node.querySelector<HTMLElement>(".pile-streak");
  if (streak) {
    run(
      streak,
      [
        { offset: 0, scale: "1 0", opacity: 0 },
        { offset: 0.25, opacity: 0.6 },
        { offset: 0.75, scale: "1 1", opacity: 0.4 },
        { offset: 1, scale: "1 1", opacity: 0 },
      ],
      span(f, 0.05, 0.95, EASE.sineInOut),
    );
  }
  const dust = node.querySelector<HTMLElement>(".pile-dust");
  if (dust) {
    const last = wins[wins.length - 1] ?? { a: 0, b: 1 };
    const at = last.a + (last.b - last.a) * PILE_LAND;
    run(
      dust,
      [
        { offset: 0, scale: "0.5", opacity: 0 },
        { offset: 0.3, scale: "1", opacity: 0.85 },
        { offset: 1, scale: "1.25", opacity: 0 },
      ],
      span(f, at - 0.04, 1, EASE.power2Out),
    );
  }
}

// ---- A library searched (#4447) ----

/** How high a riffled card lifts off its library pile, as a share of the card's height. */
export const RIFFLE_LIFT = 0.36;

/**
 * libraryRiffle thumbs through the top of a library pile while it is
 * searched: a few card backs, drawn on the pile, flick up off it one
 * after another, each tipping back a little by its bottom edge and
 * fanned to alternate sides, and drop back onto it, the next setting off
 * PILE_STAGGER into the one before (pileWindows), so the last settles as
 * the cue ends. A cool light comes up under the pile while it is looked
 * through and fades as the last card settles. Every card is a back: the
 * cue shows nothing of what the library holds, or of what the search
 * found (CR 401.2, 701.23b).
 */
function libraryRiffle(node: HTMLElement, f: CueMotion): void {
  const cards = all<HTMLElement>(node, ".riffle-card");
  const wins = pileWindows(cards.length);
  const lift = Math.max(6, f.height * RIFFLE_LIFT);
  cards.forEach((card, i) => {
    const w = wins[i];
    const side = i % 2 === 0 ? 1 : -1;
    run(
      card,
      [
        {
          offset: 0,
          transform: "perspective(420px) translate(0px, 0px) rotateX(0deg) rotate(0deg)",
          easing: EASE.power2Out,
        },
        {
          offset: 0.42,
          transform: `perspective(420px) translate(${side * 4}px, ${-lift}px) rotateX(14deg) rotate(${side * 8}deg)`,
          easing: EASE.power2In,
        },
        {
          offset: 0.86,
          transform: `perspective(420px) translate(0px, 1px) rotateX(-3deg) rotate(${side * -1}deg)`,
          easing: EASE.sineInOut,
        },
        {
          offset: 1,
          transform: "perspective(420px) translate(0px, 0px) rotateX(0deg) rotate(0deg)",
        },
      ],
      span(f, w.a, w.b),
    );
    // On the pile, a back over a back: seen only while it moves, and
    // gone once it settles, so the pile's own top card is what stays.
    run(
      card,
      [
        { offset: 0, opacity: 0 },
        { offset: 0.08, opacity: 1 },
        { offset: 0.9, opacity: 1 },
        { offset: 1, opacity: 0 },
      ],
      span(f, w.a, w.b),
    );
  });
  const glow = node.querySelector<HTMLElement>(".riffle-glow");
  if (glow) {
    run(
      glow,
      [
        { offset: 0, scale: "0.7", opacity: 0 },
        { offset: 0.25, scale: "1", opacity: 1 },
        { offset: 0.75, scale: "1.05", opacity: 0.7 },
        { offset: 1, scale: "1.15", opacity: 0 },
      ],
      span(f, 0, 1, EASE.sineInOut),
    );
  }
}

/**
 * riseArrive holds a returned card's own tile on the battlefield hidden
 * until its copy lands there (share `at` of the cue), then shows it, so
 * the copy is the one card on screen on its way out of the graveyard
 * (#4446), as commandArrive does for a commander going home. `fill:
 * "backwards"` hides it from the frame on, and once the run is over the
 * tile is its own style again. The caller cancels the returned run when
 * the cue is removed or skipped, which shows the tile at once.
 */
export function riseArrive(tile: HTMLElement, f: CueMotion, at: number): Run {
  const land = clamp01(at);
  return run(
    tile,
    [
      { offset: 0, opacity: 0 },
      { offset: land, opacity: 0 },
      { offset: Math.min(1, land + 0.08), opacity: 1 },
      { offset: 1, opacity: 1 },
    ],
    { delay: f.delay, duration: Math.max(1, f.ms), easing: EASE.linear, fill: "backwards" },
  );
}

// ---- A double-faced card transforms, or day and night (#4448) ----

/** How far a flipping card lifts toward the viewer at its edge-on moment. */
export const FLIP_LIFT = 1.08;

/**
 * cardFlip turns a double-faced permanent over to its other face, on its
 * own vertical axis (the copy is turned with a tapped card, so the axis
 * is the card's): the face it was eases away, is edge-on at
 * TRANSFORM_TURN lifted a touch toward the table, and the face it is now
 * comes round, a few degrees past, and settles. A pale light rims the
 * card as it turns edge-on, the moment it changes, and is gone as it
 * settles. Both faces are copies drawn over the tile, which waits hidden
 * under them (flipHold).
 */
function cardFlip(node: HTMLElement, f: CueMotion): void {
  const card = node.querySelector<HTMLElement>(".turn-card");
  if (card) {
    const turn = TRANSFORM_TURN;
    run(
      card,
      [
        {
          offset: 0,
          transform: "perspective(900px) rotateY(0deg) scale(1)",
          easing: EASE.power2In,
        },
        {
          offset: turn,
          transform: `perspective(900px) rotateY(90deg) scale(${FLIP_LIFT})`,
          easing: EASE.power2Out,
        },
        {
          offset: turn + (1 - turn) * 0.7,
          transform: "perspective(900px) rotateY(188deg) scale(1.02)",
          easing: EASE.sineInOut,
        },
        { offset: 1, transform: "perspective(900px) rotateY(180deg) scale(1)" },
      ],
      span(f, 0, 1),
    );
  }
  const rim = node.querySelector<HTMLElement>(".turn-rim");
  if (rim) {
    run(
      rim,
      [
        { offset: 0, opacity: 0, scale: "0.96" },
        { offset: TRANSFORM_TURN * 0.6, opacity: 0.35, scale: "1" },
        { offset: TRANSFORM_TURN, opacity: 1, scale: "1.04" },
        { offset: 0.85, opacity: 0.25, scale: "1.06" },
        { offset: 1, opacity: 0, scale: "1.08" },
      ],
      span(f, 0, 1, EASE.sineInOut),
    );
  }
}

// ---- A permanent sacrificed for its own ability (#4427) ----

/**
 * crackSplit is a fetch land cracked, or a Treasure, Clue or Food spent: no
 * altar and no knife, just the card snapping like a dry branch. A jagged
 * line runs down its length (to CRACK_SNAP) while the card gives a small
 * knock, a puff of dust comes off the crack as it snaps, and the two
 * halves fall away from each other and down the screen, turning outward
 * and fading, gravity taking them faster as they go. About 0.6 s: it
 * plays several times a game.
 */
function crackSplit(node: HTMLElement, f: CueMotion): void {
  const snap = CRACK_SNAP;
  for (const line of all<SVGPathElement>(node, ".crack-line path")) {
    run(
      line,
      [
        { offset: 0, strokeDashoffset: 1, opacity: 1 },
        { offset: snap, strokeDashoffset: 0, opacity: 1 },
        { offset: Math.min(1, snap + 0.25), strokeDashoffset: 0, opacity: 0 },
        { offset: 1, strokeDashoffset: 0, opacity: 0 },
      ],
      span(f, 0, 1, EASE.power2Out),
    );
  }
  for (const half of all<HTMLElement>(node, ".crack-half")) {
    const dx = num(half, "dx");
    const dy = num(half, "dy");
    const rot = num(half, "rot");
    const side = Math.sign(rot) || 1;
    run(
      half,
      [
        { offset: 0, translate: "0px 0px", rotate: "0deg", opacity: 1 },
        { offset: snap * 0.4, translate: `${side * 1.5}px -1px`, rotate: "0deg", opacity: 1 },
        { offset: snap, translate: "0px 0px", rotate: `${rot * 0.08}deg`, opacity: 1 },
        {
          offset: snap + (1 - snap) * 0.35,
          translate: `${dx * 0.4}px ${dy * 0.12}px`,
          rotate: `${rot * 0.45}deg`,
          opacity: 1,
          easing: EASE.power1In,
        },
        { offset: 1, translate: `${dx}px ${dy}px`, rotate: `${rot}deg`, opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  const puff = node.querySelector<HTMLElement>(".crack-puff");
  if (puff) {
    run(
      puff,
      [
        { offset: 0, opacity: 0, scale: "0.4" },
        { offset: snap * 0.8, opacity: 0, scale: "0.5" },
        { offset: snap + 0.08, opacity: 0.85, scale: "1" },
        { offset: 1, opacity: 0, scale: "1.6" },
      ],
      span(f, 0, 1, EASE.power2Out),
    );
  }
}

/**
 * flipHold holds a flipping permanent's own tile hidden while its copy
 * turns over (#4448), then shows it as the copy settles on the same face:
 * the copy is the one card on screen. `fill: "backwards"` hides it from
 * the frame on, through a flip's delay (the copy shows the face it was
 * meanwhile), and once the run is over the tile is its own style again.
 * The caller cancels the returned run when the cue is removed or skipped,
 * which shows the tile at once.
 */
export function flipHold(tile: HTMLElement, f: CueMotion): Run {
  return run(
    tile,
    [
      { offset: 0, opacity: 0 },
      { offset: 0.96, opacity: 0 },
      { offset: 1, opacity: 1 },
    ],
    { delay: f.delay, duration: Math.max(1, f.ms), easing: EASE.linear, fill: "backwards" },
  );
}

/**
 * skyWash is the game becoming day or night: the table takes the sky's
 * tint (a deep, cool blue for night, a warm gold for day) while a soft
 * band of that light crosses it, the way dusk or dawn moves (night from
 * the right, day from the left), and the moon or the sun rises a little
 * at the top and fades. It is over in about a second, and the permanents
 * it turns over flip inside it.
 */
function skyWash(node: HTMLElement, f: CueMotion): void {
  const night = f.sky !== "day";
  const tint = node.querySelector<HTMLElement>(".sky-tint");
  if (tint) {
    run(
      tint,
      [
        { offset: 0, opacity: 0 },
        { offset: 0.3, opacity: 1 },
        { offset: 0.65, opacity: 0.85 },
        { offset: 1, opacity: 0 },
      ],
      span(f, 0, 1, EASE.sineInOut),
    );
  }
  const band = node.querySelector<HTMLElement>(".sky-band");
  if (band) {
    const w = Math.max(160, f.board.w * 0.45);
    const from = night ? f.board.w : -w;
    const to = night ? -w : f.board.w;
    run(
      band,
      [
        { offset: 0, translate: `${from}px 0px`, opacity: 0 },
        { offset: 0.15, translate: `${from + (to - from) * 0.15}px 0px`, opacity: 1 },
        { offset: 0.85, translate: `${from + (to - from) * 0.85}px 0px`, opacity: 1 },
        { offset: 1, translate: `${to}px 0px`, opacity: 0 },
      ],
      span(f, 0, 1, EASE.power1Out),
    );
  }
  const orb = node.querySelector<HTMLElement>(".sky-orb");
  if (orb) {
    run(
      orb,
      [
        { offset: 0, translate: "0px 18px", opacity: 0 },
        { offset: 0.35, translate: "0px 0px", opacity: 1 },
        { offset: 0.7, translate: "0px -4px", opacity: 0.9 },
        { offset: 1, translate: "0px -8px", opacity: 0 },
      ],
      span(f, 0, 1, EASE.power2Out),
    );
  }
}

// ---- Board wipes ----

function tint(node: HTMLElement, f: CueMotion, hold: number): void {
  const el = node.querySelector<HTMLElement>(".tint");
  if (!el) return;
  run(
    el,
    [
      { offset: 0, opacity: 0 },
      { offset: 0.12, opacity: 1 },
      { offset: hold, opacity: 1 },
      { offset: 1, opacity: 0 },
    ],
    span(f, 0, 1),
  );
}

/** D10: quick cuts slicing across the battlefield, like horizontal winds. */
function windCuts(node: HTMLElement, f: CueMotion): void {
  tint(node, f, 0.75);
  const W = f.board.w;
  for (const cut of all<HTMLElement>(node, ".cut")) {
    const a = num(cut, "at");
    const len = num(cut, "len");
    const dir = num(cut, "dir") < 0 ? -1 : 1;
    const blade = cut.querySelector<HTMLElement>(".blade");
    const bw = blade?.offsetWidth || Math.max(160, W * 0.42);
    const from = dir > 0 ? -bw : W;
    const to = dir > 0 ? W : -bw;
    if (blade) {
      run(
        blade,
        [
          { translate: `${from}px 0px`, opacity: 0 },
          { translate: `${from + (to - from) * 0.12}px 0px`, opacity: 1, offset: 0.1 },
          { translate: `${from + (to - from) * 0.88}px 0px`, opacity: 1, offset: 0.9 },
          { translate: `${to}px 0px`, opacity: 0 },
        ],
        span(f, a, a + len, EASE.sineInOut),
      );
    }
    const gust = cut.querySelector<HTMLElement>(".gust");
    if (gust) {
      run(
        gust,
        [
          { translate: `${from}px 0px`, opacity: 0 },
          { translate: `${from + (to - from) * 0.3}px 0px`, opacity: 0.7, offset: 0.25 },
          { translate: `${to}px 0px`, opacity: 0 },
        ],
        span(f, a + 0.02, a + len * 1.35, EASE.sineInOut),
      );
    }
    // The slice it leaves on the battlefield, drawn behind the blade.
    const scar = cut.querySelector<HTMLElement>(".scar");
    if (scar) {
      run(
        scar,
        [
          { scale: "0 1", opacity: 0.9 },
          { scale: "1 1", opacity: 0.8, offset: 0.55 },
          { scale: "1 1", opacity: 0 },
        ],
        span(f, a + len * 0.08, a + len * 1.9, EASE.power1Out),
      );
    }
  }
}

/** D11: the board swirls into a white void with a small darkness at its heart. */
function vortex(node: HTMLElement, f: CueMotion): void {
  tint(node, f, 0.8);
  const swirl = node.querySelector<HTMLElement>(".swirl");
  if (swirl) {
    run(
      swirl,
      [
        { offset: 0, rotate: "0deg", scale: "1.3", opacity: 0 },
        { offset: 0.16, rotate: "70deg", scale: "1.2", opacity: 0.6 },
        { offset: 0.8, rotate: "460deg", scale: "0.55", opacity: 0.6 },
        { offset: 1, rotate: "640deg", scale: "0.05", opacity: 0 },
      ],
      span(f, 0, 1, EASE.power1In),
    );
  }
  const haze = node.querySelector<HTMLElement>(".haze");
  if (haze) {
    run(
      haze,
      [
        { offset: 0, opacity: 0 },
        { offset: 0.2, opacity: 1 },
        { offset: 0.85, opacity: 1 },
        { offset: 1, opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  const white = node.querySelector<HTMLElement>(".whitevoid");
  if (white) {
    run(
      white,
      [
        { offset: 0, scale: "0", opacity: 0, easing: EASE.power2Out },
        { offset: 0.28, scale: "0.7", opacity: 0.95 },
        { offset: 0.78, scale: "1", opacity: 1, easing: EASE.power2In },
        { offset: 0.94, scale: "0.12", opacity: 1 },
        { offset: 1, scale: "0", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
  const core = node.querySelector<HTMLElement>(".core");
  if (core) {
    run(
      core,
      [
        { offset: 0, scale: "0", opacity: 0 },
        { offset: 0.16, scale: "0", opacity: 0, easing: EASE.power3Out },
        { offset: 0.36, scale: "1", opacity: 1 },
        { offset: 0.88, scale: "1.1", opacity: 1, easing: EASE.power2In },
        { offset: 1, scale: "0", opacity: 0 },
      ],
      span(f, 0, 1),
    );
  }
}

/** A bounce wipe keeps the first build's single sweep, cool. */
function sweep(node: HTMLElement, f: CueMotion): void {
  const band = node.querySelector<HTMLElement>(".sweep");
  if (band) {
    const bandW = Math.max(120, f.board.w * 0.34);
    band.style.width = `${bandW}px`;
    run(
      band,
      [
        { offset: 0, translate: `${-bandW}px 0px`, opacity: 0 },
        { offset: 0.1, translate: `${-bandW * 0.7}px 0px`, opacity: 1 },
        { offset: 0.9, translate: `${f.board.w - bandW * 0.3}px 0px`, opacity: 1 },
        { offset: 1, translate: `${f.board.w}px 0px`, opacity: 0 },
      ],
      span(f, 0, 1, EASE.power1Out),
    );
  }
  const el = node.querySelector<HTMLElement>(".tint");
  if (el) {
    run(el, [{ opacity: 0 }, { opacity: 1, offset: 0.3 }, { opacity: 0 }], span(f, 0, 1));
  }
}
