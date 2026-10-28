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

import { EASE, run } from "./motion";
import { SHATTER_BREAK, SHATTER_BREAK_LIGHT, sacrificeBeats } from "./tableEffects";
import type { EffectKind, WipeTone } from "./tableEffects";

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
    case "pw_leave":
      pwLeave(node, f);
      return;
    case "pw_enter":
      pwEnter(node, f);
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
        { offset: breakAt + 0.18, opacity: 1, translate: "0px 0px", scale: "1" },
        { offset: breakAt + 0.42, opacity: 0, translate: "0px 6px", scale: "1" },
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
        { transform: "perspective(700px) rotateY(0deg)" },
        { transform: "perspective(700px) rotateY(360deg)" },
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
