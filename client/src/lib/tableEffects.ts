// tableEffects — the generic effect cues (ADR 0209 §1 and §3, #4396).
//
// A permanent that is destroyed, exiled, sacrificed or bounced, a spell
// that is countered, a planeswalker that arrives or goes, and a board
// wipe each get one cue over the table, so a player sees what happened
// before they read the log. TableEffects.svelte measures, draws and
// plays; every rule is here, pure and unit-tested:
//
//   - which log entries are new (the combat beats' watermark, reused:
//     the first frame primes, a rewind forgets, a re-prime cancels);
//   - which cue a battlefield exit is, read from `new_zone` FIRST and
//     `cause` second (a destruction replaced by exile has cause
//     `destroy` and new_zone `exile`, and must look like an exile);
//   - which exits are a board wipe: three or more permanents leaving in
//     one `batch` from one `source`;
//   - the sound, the length (capped, then scaled by animations.speed,
//     and compressed when the table is busy) and the three motion gates;
//   - the seeded shapes the cues draw: the shatter, the void's motes,
//     a vortex's spiral and a board wipe's wind cuts.
//
// It never touches game state, and nothing waits on it (ADR 0209 §1).

import type { LogEvent } from "./protocol";
import { seeded } from "./combatStrikes";
import type { SoundName } from "./sounds";

// ---- The setting and the gates ----

/** animations.effects (ADR 0209 §1): how much of each cue plays. */
export type EffectsMode = "full" | "short" | "off";

export const DEFAULT_EFFECTS_MODE: EffectsMode = "full";

export function isEffectsMode(v: unknown): v is EffectsMode {
  return v === "full" || v === "short" || v === "off";
}

/**
 * How a cue is shown: "full" plays it, "short" plays it cut to
 * EFFECT_SHORT_CAP_MS, "still" shows the outcome badge and no motion.
 */
export type EffectMotion = "full" | "short" | "still";

/**
 * effectMotion folds the signals. Any of them makes a cue still: the
 * animations master switch off, `animations.effects` off, the setting
 * `accessibility.reduceMotion`, or the OS's `prefers-reduced-motion`
 * (DESIGN.md §4, all three). accessibility.reduceMotion is never pushed
 * into animations.ts, so the caller passes every signal.
 */
export function effectMotion(s: {
  enabled: boolean;
  effects: EffectsMode;
  reduceMotion: boolean;
  osReduceMotion?: boolean;
}): EffectMotion {
  if (!s.enabled || s.effects === "off" || s.reduceMotion || s.osReduceMotion === true) {
    return "still";
  }
  return s.effects === "short" ? "short" : "full";
}

/**
 * The longest a generic cue plays, before animations.speed (ADR 0209 §1,
 * amended 2026-10-28 for #4416: every cue about a third slower, so the
 * cap rose from 1.5 s).
 */
export const EFFECT_CAP_MS = 2000;
/** The longest a cue plays in the Short setting, before animations.speed (was 0.8 s). */
export const EFFECT_SHORT_CAP_MS = 1000;
/**
 * How long a still badge stays up. It is reading time, not motion, so
 * animations.speed does not scale it (StackLinger's 1.5 s is the same).
 */
export const EFFECT_STILL_HOLD_MS = 1200;
/** A cue that starts later than this after it was due is dropped. */
export const EFFECT_LATE_MS = 1000;
/** Most cues drawn from one frame; a bigger wipe draws its sweep and the first few. */
export const EFFECT_CUE_CAP = 12;

/** animations.speed as a safe multiplier (2 = twice as slow). */
export function speedFactor(speed: number): number {
  return Number.isFinite(speed) && speed > 0 ? speed : 1;
}

/**
 * cueMs is a cue's length: its base length, capped (2 s, or 1 s in
 * Short), then scaled by animations.speed. A still cue has no motion:
 * it is the badge's hold, which speed does not scale.
 */
export function cueMs(baseMs: number, motion: EffectMotion, speed: number): number {
  if (motion === "still") return EFFECT_STILL_HOLD_MS;
  const cap = motion === "short" ? EFFECT_SHORT_CAP_MS : EFFECT_CAP_MS;
  return Math.max(1, Math.min(baseMs, cap) * speedFactor(speed));
}

// ---- The cues ----

export type EffectKind =
  | "destroy"
  | "exile"
  | "dies"
  | "sacrifice"
  | "bounce"
  | "counter"
  | "pw_enter"
  | "pw_leave";

/**
 * Base length of each cue at speed 1, before the cap. About 1.35 times
 * the first build's (#4416: "so you can see it"). A sacrifice is longer
 * again, because it is now an altar, a knife and then the destroy; a
 * counterspell holds its cracks while the stack linger shows the spell;
 * a planeswalker that dies now also withdraws from the plane.
 */
export const EFFECT_BASE_MS: Record<EffectKind, number> = {
  destroy: 1250,
  exile: 1500,
  // The existing crumble (STRIKE_CRUMBLE_MS in animations.ts, 420 ms in combat).
  dies: 560,
  sacrifice: 1750,
  bounce: 1300,
  counter: 1600,
  pw_enter: 1200,
  pw_leave: 1500,
};

/**
 * A sacrifice (#4416): a quick altar, a knife that stabs the card at
 * `stab` of the cue, then the destroy's shatter from `shatter` on.
 * Inside a board wipe there is no altar, and the knife comes at once.
 */
export function sacrificeBeats(light: boolean): { stab: number; shatter: number } {
  return light ? { stab: 0.16, shatter: 0.2 } : { stab: 0.3, shatter: 0.34 };
}

/** The sound each cue plays. A death by state-based action has none new. */
export const EFFECT_SOUND: Record<EffectKind, SoundName | null> = {
  destroy: "destroy",
  exile: "exile",
  dies: null,
  sacrifice: "sacrifice",
  bounce: "bounce",
  counter: "counter",
  pw_enter: "pw_enter",
  pw_leave: "pw_leave",
};

/**
 * A sacrifice's second sound, as its knife lands (#4416). Silent until
 * its file exists (sounds.ts), like every generic cue's.
 */
export const SACRIFICE_KNIFE_SOUND: SoundName = "sacrifice_knife";

/** Three or more permanents leaving in one batch from one source. */
export const WIPE_MIN = 3;
/**
 * The sweep's length (a destroy wipe's wind cuts), and a per-card cue's
 * length inside a wipe. An exile wipe's vortex runs longer, WIPE_VORTEX_MS:
 * it stays open until the last card has been drawn into it.
 */
export const WIPE_SWEEP_MS = 950;
export const WIPE_CARD_MS = 800;
/** The first per-card cue starts this long into the sweep. */
export const WIPE_FIRST_CARD_MS = 300;
/** Per-card cues are spread over this long at most, one stagger apart. */
export const WIPE_SPREAD_MS = 600;
export const WIPE_STAGGER_MAX_MS = 80;
/** The volume a per-card sound plays at inside a wipe. */
export const WIPE_CARD_VOLUME = 0.45;

/** The whole wipe, sweep and the last per-card cue, fits one cue's cap. */
export const WIPE_TOTAL_MS = WIPE_FIRST_CARD_MS + WIPE_SPREAD_MS + WIPE_CARD_MS;
/** An exile wipe's vortex: the whole wipe, and a beat for the void to close. */
export const WIPE_VORTEX_MS = WIPE_TOTAL_MS + 150;

export type WipeTone = "dark" | "cold" | "cool";

export interface EffectCue {
  /** Unique within the plan: the entry's seq, with a suffix for a second cue of one entry. */
  id: string;
  seq: number;
  kind: EffectKind;
  /** The permanent (or, for a counter, the spell's card), when the entry names one. */
  cardID: string | null;
  /** For a counter: the stack item that was countered. */
  stackItemID: string | null;
  batch: number | null;
  /** "light" inside a board wipe: shorter, no flourish, a quieter sound. */
  weight: "full" | "light";
  /** Start offset at speed 1 (a wipe staggers its per-card cues). */
  delayMs: number;
  /**
   * An exile that was a destruction replaced by exile (Rest in Peace):
   * it looks like an exile, with a slight crumble as it goes (#4416).
   */
  fromDestroy?: boolean;
}

export interface BoardWipe {
  batch: number;
  source: string;
  /** How many permanents left. */
  count: number;
  tone: WipeTone;
  /** The first of its entries, to order it among the frame's cues. */
  seq: number;
}

export interface EffectPlan {
  cues: EffectCue[];
  wipes: BoardWipe[];
}

export interface PlanContext {
  /** Whether this card is (or just was) a planeswalker, from the views the client holds. */
  isPlaneswalker(cardID: string): boolean;
  /**
   * The combat strike layer is playing this frame (ADR 0134): it crumbles
   * the creatures combat damage kills, so a state-based death in a combat
   * damage step is not drawn twice.
   */
  combatMotion: boolean;
}

const DAMAGE_STEPS = new Set(["first_strike_damage", "combat_damage"]);

interface Exit {
  e: LogEvent;
  step: string;
  /** The zone entry for this card in this batch, if there is one. */
  zone: LogEvent | null;
  sacrificed: boolean;
}

/**
 * exitKind is the cue a permanent leaving the battlefield gets, or null
 * for none. `new_zone` is read first: an exile is an exile whatever the
 * instruction said, and a card that went to a hand is a bounce.
 */
export function exitKind(
  x: Pick<Exit, "zone" | "sacrificed" | "step">,
  ctx: PlanContext,
): EffectKind | null {
  const z = x.zone;
  // A destruction replaced by exile (Rest in Peace) is `destroy` with
  // new_zone exile: it must show the exile look, never a death.
  if (z?.new_zone === "exile") return "exile";
  if (z?.new_zone === "hand" || (z && z.cause === "bounce")) return "bounce";
  if (x.sacrificed) return "sacrifice";
  if (!z) return null;
  // Into a library, the command zone or anywhere unseen: no cue.
  if (z.new_zone !== "graveyard") return null;
  const cardID = z.card_id ?? "";
  switch (z.cause) {
    case "destroy":
      return cardID && ctx.isPlaneswalker(cardID) ? "pw_leave" : "destroy";
    case "sba":
    case "other":
    case undefined:
    case "": {
      // Combat damage's deaths are the strike layer's crumble.
      if (z.cause === "sba" && ctx.combatMotion && DAMAGE_STEPS.has(x.step)) return null;
      return cardID && ctx.isPlaneswalker(cardID) ? "pw_leave" : "dies";
    }
    default:
      return "dies";
  }
}

function isExit(e: LogEvent): boolean {
  return (e.kind === "zone" && e.old_zone === "battlefield") || e.kind === "sacrifice";
}

/**
 * planEffects turns a frame's new log entries into cues. `fresh` is the
 * entries this client has not handled, `log` the frame's whole log (an
 * entry's step is on an older entry).
 */
export function planEffects(
  fresh: readonly LogEvent[],
  log: readonly LogEvent[] | undefined,
  ctx: PlanContext,
): EffectPlan {
  if (fresh.length === 0) return { cues: [], wipes: [] };
  const freshSeqs = new Set(fresh.map((e) => e.seq));

  // One pass over the log, in order, so each fresh entry knows its step.
  let step = "";
  const exits = new Map<string, Exit>();
  const order: Exit[] = [];
  const cues: EffectCue[] = [];
  for (const e of log ?? []) {
    if (e.kind === "step") {
      step = e.step ?? "";
      continue;
    }
    if (!freshSeqs.has(e.seq)) continue;

    if (e.kind === "counter") {
      const item = e.stack_item_id ?? e.target ?? null;
      cues.push({
        id: String(e.seq),
        seq: e.seq,
        kind: "counter",
        cardID: e.target ?? null,
        stackItemID: item,
        batch: e.batch ?? null,
        weight: "full",
        delayMs: 0,
      });
      continue;
    }

    if (e.kind === "zone" && e.new_zone === "battlefield" && e.old_zone !== "battlefield") {
      if (e.card_id && ctx.isPlaneswalker(e.card_id)) {
        cues.push({
          id: String(e.seq),
          seq: e.seq,
          kind: "pw_enter",
          cardID: e.card_id,
          stackItemID: null,
          batch: e.batch ?? null,
          weight: "full",
          delayMs: 0,
        });
      }
      continue;
    }

    if (!isExit(e)) continue;
    // A card's exits in one batch are one: a sacrifice entry and the zone
    // move that follows it (when the route was not the graveyard) are one cue.
    const key = `${e.batch ?? "-"}:${e.card_id ?? `q${e.seq}`}`;
    let x = exits.get(key);
    if (!x) {
      x = { e, step, zone: null, sacrificed: false };
      exits.set(key, x);
      order.push(x);
    }
    if (e.kind === "sacrifice") x.sacrificed = true;
    else x.zone = e;
  }

  // Board wipes: three or more exits in one batch from one source.
  const groups = new Map<string, Exit[]>();
  for (const x of order) {
    const batch = x.e.batch;
    const source = x.e.source;
    if (batch === undefined || batch === null || !source) continue;
    const k = `${batch}|${source}`;
    const g = groups.get(k);
    if (g) g.push(x);
    else groups.set(k, [x]);
  }
  const wipes: BoardWipe[] = [];
  const inWipe = new Map<Exit, { index: number; of: number }>();
  for (const [, g] of groups) {
    if (g.length < WIPE_MIN) continue;
    g.forEach((x, index) => inWipe.set(x, { index, of: g.length }));
    wipes.push({
      batch: g[0].e.batch as number,
      source: g[0].e.source as string,
      count: g.length,
      tone: toneOf(g, ctx),
      seq: g[0].e.seq,
    });
  }

  for (const x of order) {
    const kind = exitKind(x, ctx);
    if (!kind) continue;
    const w = inWipe.get(x);
    const stagger = w ? Math.min(WIPE_STAGGER_MAX_MS, WIPE_SPREAD_MS / Math.max(1, w.of - 1)) : 0;
    const cue: EffectCue = {
      id: String(x.e.seq),
      seq: x.e.seq,
      kind,
      cardID: x.e.card_id ?? null,
      stackItemID: null,
      batch: x.e.batch ?? null,
      weight: w ? "light" : "full",
      delayMs: w ? WIPE_FIRST_CARD_MS + w.index * stagger : 0,
    };
    if (kind === "exile" && x.zone?.cause === "destroy") cue.fromDestroy = true;
    cues.push(cue);
  }
  cues.sort((a, b) => a.seq - b.seq);
  return { cues, wipes };
}

function toneOf(g: readonly Exit[], ctx: PlanContext): WipeTone {
  const n = { dark: 0, cold: 0, cool: 0 };
  for (const x of g) {
    const k = exitKind(x, ctx);
    if (k === "exile") n.cold++;
    else if (k === "bounce") n.cool++;
    else n.dark++;
  }
  if (n.cold > n.dark && n.cold >= n.cool) return "cold";
  if (n.cool > n.dark && n.cool > n.cold) return "cool";
  return "dark";
}

/**
 * effectLate: a cue starting more than EFFECT_LATE_MS after it was due
 * (its frame's arrival plus its stagger) is dropped, the way a late
 * combat strike is (ADR 0134 §4): the board has moved on.
 */
export function effectLate(frameAt: number, atMs: number, now: number): boolean {
  return now - (frameAt + atMs) > EFFECT_LATE_MS;
}

/**
 * effectsScheduled: a frame that arrives while the page is hidden draws
 * no motion (Web Animations follow the frame clock, which a hidden tab
 * does not run), though its sounds still play.
 */
export function effectsScheduled(visibility: string | undefined): boolean {
  return visibility === undefined || visibility === "visible";
}

/** The key one logical sound is capped on: once per batch, per game. */
export function soundBatch(gameID: string, batch: number | null): string | undefined {
  return batch === null ? undefined : `${gameID}:${batch}`;
}

/** The sound a cue plays, and its volume hint (a wipe's per-card sounds are quieter). */
export function cueSound(cue: Pick<EffectCue, "kind" | "weight">): {
  name: SoundName;
  volume: number;
} | null {
  const name = EFFECT_SOUND[cue.kind];
  if (!name) return null;
  return { name, volume: cue.weight === "light" ? WIPE_CARD_VOLUME : 1 };
}

// ---- Shapes ----

interface Pt {
  x: number;
  y: number;
}

/** One piece of a shattered card. */
export interface ShatterShard {
  /** Its outline: a clip-path polygon in percent of the card's box. */
  clip: string;
  /** 0 at the impact, rising outward: the inner pieces let go first. */
  ring: number;
  /** Its centre, px in the card's box: what it spins about. */
  ox: number;
  oy: number;
  /**
   * Its flight, in px in the card's own frame (gravity is already turned
   * into that frame, so a tapped card's pieces still fall down the
   * screen), and its spin in degrees.
   */
  dx: number;
  dy: number;
  rot: number;
  /** When it lets go, as a share of the shatter (0..1). */
  at: number;
}

/** A speck of dust thrown out of the impact. */
export interface ShatterMote {
  dx: number;
  dy: number;
  size: number;
  at: number;
}

export interface Shatter {
  /** The impact, in px in the card's box (near its centre). */
  ix: number;
  iy: number;
  /** The cracks: spokes from the impact to the edge, then the rings across them (SVG path data, px). */
  spokes: string[];
  rings: string[];
  shards: ShatterShard[];
  motes: ShatterMote[];
}

/** The share of a shatter before its first piece lets go: the cracks run first. */
export const SHATTER_BREAK = 0.3;
/** A wipe's lighter shatter breaks sooner. */
export const SHATTER_BREAK_LIGHT = 0.18;

/**
 * shatterPlan breaks a card `w` x `h` (px, unturned) from an impact near
 * its centre outward (#4416, D1): spokes run from the impact to the edge,
 * rings cross them, and the pieces between let go ring by ring, inner
 * first, flying away from the impact and falling. `rot` is the card's
 * turn on screen (90 for a tapped card). Seeded, so a card always breaks
 * the same way. A light plan (inside a board wipe) has fewer pieces.
 */
export function shatterPlan(
  seed: string,
  w: number,
  h: number,
  opts: { rot?: number; light?: boolean } = {},
): Shatter {
  const rnd = seeded(`shatter:${seed}`);
  const light = opts.light === true;
  const spokeCount = light ? 8 : 11;
  const ringFracs = light ? [0.48] : [0.3, 0.64];
  const breakAt = light ? SHATTER_BREAK_LIGHT : SHATTER_BREAK;
  const W = Math.max(1, w);
  const H = Math.max(1, h);
  const I = { x: W * (0.5 + (rnd() - 0.5) * 0.12), y: H * (0.5 + (rnd() - 0.5) * 0.12) };
  const TAU = Math.PI * 2;
  const angles: number[] = [];
  for (let i = 0; i < spokeCount; i++) angles.push(((i + 0.2 + rnd() * 0.6) / spokeCount) * TAU);
  const edge = (a: number): Pt => {
    const c = Math.cos(a);
    const s = Math.sin(a);
    let t = Infinity;
    if (c > 1e-9) t = Math.min(t, (W - I.x) / c);
    if (c < -1e-9) t = Math.min(t, -I.x / c);
    if (s > 1e-9) t = Math.min(t, (H - I.y) / s);
    if (s < -1e-9) t = Math.min(t, -I.y / s);
    return { x: I.x + c * t, y: I.y + s * t };
  };
  // P[i][k]: the impact (k 0), the rings, then the edge (k R).
  const P: Pt[][] = angles.map((a) => {
    const e = edge(a);
    const pts: Pt[] = [I];
    for (const f of ringFracs) {
      const ff = clamp(f + (rnd() - 0.5) * 0.14, 0.15, 0.85);
      pts.push({ x: I.x + (e.x - I.x) * ff, y: I.y + (e.y - I.y) * ff });
    }
    pts.push(e);
    return pts;
  });
  const R = ringFracs.length + 1;
  const angleOf = (p: Pt) => {
    const a = Math.atan2(p.y - I.y, p.x - I.x);
    return a < 0 ? a + TAU : a;
  };
  const corners: Pt[] = [
    { x: 0, y: 0 },
    { x: W, y: 0 },
    { x: W, y: H },
    { x: 0, y: H },
  ];
  const rotRad = ((opts.rot ?? 0) * Math.PI) / 180;
  const reach = Math.max(W, H);
  const pct = (p: Pt) => `${round1((p.x / W) * 100)}% ${round1((p.y / H) * 100)}%`;

  const shards: ShatterShard[] = [];
  for (let i = 0; i < spokeCount; i++) {
    const j = (i + 1) % spokeCount;
    const a0 = angles[i];
    const a1 = j === 0 ? angles[0] + TAU : angles[j];
    for (let k = 0; k < R; k++) {
      let pts: Pt[];
      if (k === 0) pts = [I, P[i][1], P[j][1]];
      else if (k < R - 1) pts = [P[i][k], P[i][k + 1], P[j][k + 1], P[j][k]];
      else {
        const between = corners
          .map((c) => {
            let a = angleOf(c);
            if (a < a0) a += TAU;
            return { c, a };
          })
          .filter((x) => x.a > a0 && x.a < a1)
          .sort((p, q) => p.a - q.a)
          .map((x) => x.c);
        pts = [P[i][k], P[i][R], ...between, P[j][R], P[j][k]];
      }
      const cx = pts.reduce((t, p) => t + p.x, 0) / pts.length;
      const cy = pts.reduce((t, p) => t + p.y, 0) / pts.length;
      const len = Math.hypot(cx - I.x, cy - I.y) || 1;
      const dist = reach * (0.14 + rnd() * 0.2) * (0.8 + 0.3 * k);
      // Down the screen, turned into the card's frame.
      const g = H * (0.25 + rnd() * 0.25);
      shards.push({
        clip: `polygon(${pts.map(pct).join(", ")})`,
        ring: k,
        ox: round1(cx),
        oy: round1(cy),
        dx: round1(((cx - I.x) / len) * dist + g * Math.sin(rotRad)),
        dy: round1(((cy - I.y) / len) * dist + g * Math.cos(rotRad)),
        rot: round1((rnd() * 2 - 1) * (60 + 50 * k)),
        at: round2(breakAt + k * 0.08 + rnd() * 0.05),
      });
    }
  }

  const path = (pts: Pt[], close = false) =>
    `M${pts.map((p) => `${round1(p.x)} ${round1(p.y)}`).join(" L")}${close ? " Z" : ""}`;
  const spokes = P.map((pts) => {
    // A little jag between the joints, so a crack never reads as a ruled line.
    const out: Pt[] = [pts[0]];
    for (let k = 1; k < pts.length; k++) {
      const a = pts[k - 1];
      const b = pts[k];
      const nx = -(b.y - a.y);
      const ny = b.x - a.x;
      const n = Math.hypot(nx, ny) || 1;
      const j = (rnd() - 0.5) * 0.18 * n;
      out.push({ x: (a.x + b.x) / 2 + (nx / n) * j, y: (a.y + b.y) / 2 + (ny / n) * j }, b);
    }
    return path(out);
  });
  const rings: string[] = [];
  for (let k = 1; k < R; k++)
    rings.push(
      path(
        P.map((pts) => pts[k]),
        true,
      ),
    );

  const motes: ShatterMote[] = [];
  const moteCount = light ? 5 : 12;
  for (let m = 0; m < moteCount; m++) {
    const a = rnd() * TAU;
    const d = reach * (0.35 + rnd() * 0.55);
    const g = H * (0.15 + rnd() * 0.2);
    motes.push({
      dx: round1(Math.cos(a) * d + g * Math.sin(rotRad)),
      dy: round1(Math.sin(a) * d + g * Math.cos(rotRad)),
      size: round1(1.5 + rnd() * 2.5),
      at: round2(breakAt + rnd() * 0.06),
    });
  }
  return { ix: round1(I.x), iy: round1(I.y), spokes, rings, shards, motes };
}

/** A speck of light drawn into an exile's void, on a spiral. */
export interface VoidMote {
  /** Where it starts: its angle (degrees) and distance (px) from the card's centre. */
  angle: number;
  r: number;
  /** How far round it turns on the way in, in degrees. */
  spin: number;
  size: number;
  /** When it sets off, as a share of the cue. */
  at: number;
}

/**
 * voidMotes plans the specks an exile's white void pulls in (#4416, D3):
 * each starts out past the card's edge and spirals into the centre.
 */
export function voidMotes(seed: string, w: number, h: number, light = false): VoidMote[] {
  const rnd = seeded(`void:${seed}`);
  const reach = Math.hypot(w, h) * 0.62;
  const out: VoidMote[] = [];
  const n = light ? 6 : 14;
  for (let i = 0; i < n; i++) {
    out.push({
      angle: round1(((i + rnd() * 0.8) / n) * 360),
      r: round1(reach * (0.8 + rnd() * 0.45)),
      spin: round1(200 + rnd() * 160),
      size: round1(1.5 + rnd() * 2.5),
      at: round2(0.05 + rnd() * 0.3),
    });
  }
  return out;
}

/**
 * spiralPath is the way a card is drawn into a board-wide vortex (#4416,
 * D11): `steps` + 1 offsets from where it stood, turning `turns` times
 * round the vortex's centre `to` as it closes on it. The first is 0, 0
 * and the last is the centre.
 */
export function spiralPath(from: Pt, to: Pt, turns = 0.6, steps = 5): Pt[] {
  const r0 = Math.hypot(from.x - to.x, from.y - to.y);
  const a0 = Math.atan2(from.y - to.y, from.x - to.x);
  const out: Pt[] = [];
  for (let k = 0; k <= steps; k++) {
    const t = k / steps;
    const r = r0 * Math.pow(1 - t, 1.15);
    const a = a0 + turns * Math.PI * 2 * t;
    out.push({
      x: round1(to.x + Math.cos(a) * r - from.x),
      y: round1(to.y + Math.sin(a) * r - from.y),
    });
  }
  return out;
}

/**
 * swirlArms draws an exile wipe's vortex (#4416, D11): `n` arms, each a
 * tapered crescent winding `turns` times from the rim into the centre.
 * Path data in a box from -100 to 100 on both axes, centred on 0, 0.
 */
export function swirlArms(n = 5, turns = 1.1, steps = 28): string[] {
  const arms: string[] = [];
  for (let k = 0; k < n; k++) {
    const base = (k / n) * Math.PI * 2;
    const outer: Pt[] = [];
    const inner: Pt[] = [];
    for (let i = 0; i <= steps; i++) {
      const t = i / steps;
      const r = 100 * Math.pow(1 - t, 1.25);
      const a = base + turns * Math.PI * 2 * t;
      // The arm is widest at the rim and closes to a point at the centre.
      const w = 0.42 * (1 - t);
      outer.push({ x: Math.cos(a) * r, y: Math.sin(a) * r });
      inner.push({ x: Math.cos(a + w) * r, y: Math.sin(a + w) * r });
    }
    const pts = [...outer, ...inner.reverse()];
    arms.push(`M${pts.map((p) => `${round1(p.x)} ${round1(p.y)}`).join(" L")} Z`);
  }
  return arms;
}

/** One gust in a destroy wipe: a cut that races across the whole board. */
export interface WindCut {
  /** Its height on the board, in percent, and its tilt in degrees. */
  y: number;
  tilt: number;
  /** 1 left to right, -1 right to left. */
  dir: 1 | -1;
  /** When it sets off and how long it crosses for, as shares of the sweep. */
  at: number;
  len: number;
  /** Its thickness, px. */
  thick: number;
}

/**
 * windCuts plans a destroy wipe's sweep (#4416, D10): quick cuts slicing
 * across the battlefield like horizontal winds, at spread-out heights,
 * one after another, most of them left to right.
 */
export function windCuts(seed: string, n = 7): WindCut[] {
  const rnd = seeded(`wind:${seed}`);
  // Spread-out heights, shuffled: the gusts land all over the board, not
  // top to bottom in order.
  const heights: number[] = [];
  for (let i = 0; i < n; i++) heights.push(round1(((i + 0.25 + rnd() * 0.5) / n) * 84 + 8));
  for (let i = n - 1; i > 0; i--) {
    const j = Math.floor(rnd() * (i + 1));
    [heights[i], heights[j]] = [heights[j], heights[i]];
  }
  const out: WindCut[] = [];
  for (let i = 0; i < n; i++) {
    out.push({
      y: heights[i],
      tilt: round1((rnd() * 2 - 1) * 4),
      dir: rnd() < 0.72 ? 1 : -1,
      at: round2(Math.min(0.62, (i / Math.max(1, n)) * 0.58 + rnd() * 0.05)),
      len: round2(0.24 + rnd() * 0.1),
      thick: round1(2.2 + rnd() * 2.2),
    });
  }
  return out;
}

export interface BurstShard {
  /** Where it starts, in percent of the spot's box. */
  x: number;
  y: number;
  /** How far it flies, in px at the box's own scale, its spin in degrees and its size in px. */
  dx: number;
  dy: number;
  rot: number;
  size: number;
}

export const BURST_CRACKS = 7;
export const BURST_SHARDS = 9;

/**
 * counterBurst plans a counterspell's shatter over a stack item `w` x `h`:
 * jagged cracks run out from the middle (path data over a 0..100 box,
 * each drawn with pathLength 1), and glass shards fly out of it. Seeded by
 * the entry, so a cue always looks the same.
 */
export function counterBurst(
  seed: string,
  w: number,
  h: number,
): { cracks: string[]; shards: BurstShard[] } {
  const rnd = seeded(seed);
  const cracks: string[] = [];
  for (let i = 0; i < BURST_CRACKS; i++) {
    const base = ((i + rnd() * 0.6) / BURST_CRACKS) * Math.PI * 2;
    let x = 50;
    let y = 50;
    let d = `M50 50`;
    const steps = 3 + Math.floor(rnd() * 2);
    for (let s = 1; s <= steps; s++) {
      const a = base + (rnd() - 0.5) * 0.7;
      const r = (s / steps) * (46 + rnd() * 10);
      x = clamp(50 + Math.cos(a) * r, 0, 100);
      y = clamp(50 + Math.sin(a) * r, 0, 100);
      d += ` L${round1(x)} ${round1(y)}`;
    }
    cracks.push(d);
  }
  const reach = Math.max(w, h);
  const shards: BurstShard[] = [];
  for (let i = 0; i < BURST_SHARDS; i++) {
    const a = ((i + rnd() * 0.8) / BURST_SHARDS) * Math.PI * 2;
    const dist = reach * (0.35 + rnd() * 0.45);
    shards.push({
      x: round1(50 + Math.cos(a) * 12),
      y: round1(50 + Math.sin(a) * 12),
      dx: round1(Math.cos(a) * dist),
      dy: round1(Math.sin(a) * dist + reach * 0.15),
      rot: round1((rnd() * 2 - 1) * 200),
      size: round1(8 + rnd() * 10),
    });
  }
  return { cracks, shards };
}

function clamp(v: number, lo: number, hi: number): number {
  return Math.min(hi, Math.max(lo, v));
}

function round1(n: number): number {
  return Math.round(n * 10) / 10;
}

function round2(n: number): number {
  return Math.round(n * 100) / 100;
}

// ---- Timing ----

/**
 * How much a board wipe is compressed in Short: the whole wipe (the sweep
 * and the last per-card cue) fits EFFECT_SHORT_CAP_MS instead of
 * WIPE_TOTAL_MS.
 */
function wipeScale(motion: EffectMotion): number {
  return motion === "short" ? Math.min(1, EFFECT_SHORT_CAP_MS / WIPE_TOTAL_MS) : 1;
}

/**
 * cueTiming is when a cue's motion starts (an offset from its frame) and
 * how long it runs, both in real milliseconds, scaled by animations.speed.
 * A cue inside a wipe waits for the sweep and is lighter; the cap holds
 * for the whole wipe, so delay plus length never passes the cap (before
 * speed). A still cue starts at once and is the badge's hold.
 */
export function cueTiming(
  cue: Pick<EffectCue, "kind" | "weight" | "delayMs">,
  motion: EffectMotion,
  speed: number,
): { delayMs: number; ms: number } {
  if (motion === "still") return { delayMs: 0, ms: EFFECT_STILL_HOLD_MS };
  const s = speedFactor(speed);
  if (cue.weight === "light") {
    const k = wipeScale(motion);
    return {
      delayMs: cue.delayMs * k * s,
      ms: Math.max(1, Math.min(EFFECT_BASE_MS[cue.kind], WIPE_CARD_MS) * k * s),
    };
  }
  return { delayMs: 0, ms: cueMs(EFFECT_BASE_MS[cue.kind], motion, speed) };
}

/**
 * The sweep's length for a board wipe, in real milliseconds. A cold
 * (exile) wipe's vortex stays open for the whole wipe.
 */
export function wipeTiming(motion: EffectMotion, speed: number, tone?: WipeTone): { ms: number } {
  if (motion === "still") return { ms: EFFECT_STILL_HOLD_MS };
  const base = tone === "cold" ? WIPE_VORTEX_MS : WIPE_SWEEP_MS;
  const cap = motion === "short" ? EFFECT_SHORT_CAP_MS : EFFECT_CAP_MS;
  return { ms: Math.max(1, Math.min(base * wipeScale(motion), cap) * speedFactor(speed)) };
}

// ---- A busy table ----

/** How far back a cue still counts toward a busy table. */
export const BUSY_WINDOW_MS = 2000;
/** Each further cue in the window takes this much off (1 / (1 + step * (n - 1))). */
export const BUSY_STEP = 0.18;
/** The most a busy table compresses a cue: never under this share of its length. */
export const BUSY_MIN = 0.55;

/**
 * busyScale is how much a cue is compressed when `n` cues (this one
 * included) have started within BUSY_WINDOW_MS (#4416: "speed effects
 * back up when a lot is happening at once"). One cue plays at full
 * length; two at 0.85; four at about 0.65 (the first build's lengths);
 * six or more at BUSY_MIN. It multiplies the cue's length like a faster
 * animations.speed, so a cap scaled by speed still holds.
 */
export function busyScale(n: number): number {
  if (!Number.isFinite(n) || n <= 1) return 1;
  return Math.max(BUSY_MIN, 1 / (1 + BUSY_STEP * (n - 1)));
}

/**
 * headlines is how many cues a plan adds to a busy table: each cue of
 * its own, each board wipe and each signature. A wipe's per-card cues
 * are not counted: the wipe already fits them inside its cap.
 */
export function headlines(plan: EffectPlan, signatures = 0): number {
  return plan.cues.filter((c) => c.weight === "full").length + plan.wipes.length + signatures;
}

/**
 * BusyClock remembers when recent cues started, so a frame knows how
 * busy the table is. `now` is a clock in ms (Date.now in the table).
 */
export class BusyClock {
  private starts: number[] = [];

  /** The scale for a frame adding `n` cues at `now`; it records them. */
  frame(n: number, now: number): number {
    this.starts = this.starts.filter((t) => now - t < BUSY_WINDOW_MS);
    if (n <= 0) return 1;
    const scale = busyScale(this.starts.length + n);
    for (let i = 0; i < n; i++) this.starts.push(now);
    return scale;
  }

  /** A skip or a re-prime: the table is quiet again. */
  reset(): void {
    this.starts = [];
  }
}
