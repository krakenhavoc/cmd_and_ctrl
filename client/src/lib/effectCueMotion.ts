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
import { SHATTER_BREAK, SHATTER_BREAK_LIGHT, sacrificeBeats } from "./tableEffects";
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
