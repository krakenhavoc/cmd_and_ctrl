// Signature effects for named cards (ADR 0209 §4, #4397).
//
// A signature is a short scene, plus its sound, that a named card plays in
// place of the generic cue (lib/tableEffects.ts) for the batch it names:
//
//   resolve  the card's spell resolved. Its batch is every permanent that
//            left the battlefield with the spell as their `source` (a
//            board wipe's sweep is replaced; each card's own cue still
//            plays, lighter, once the scene hands over).
//   arrive   the card was cast, or entered the battlefield without having
//            been cast (reanimated, flickered, put onto it). One scene per
//            arrival: a card whose cast played it does not play it again
//            when it resolves. It replaces the planeswalker's arrival cue.
//
// This file is the eager half: which oracle_id has which scene, how long
// each scene is, and every rule (identification, caps, Short, still,
// lazy loading). The scenes themselves, built from the scene kit, are in
// fx/scenes/ and are imported on demand, the first time a card with a
// signature appears in a game (SignatureLoader below). A signature never
// changes rules, the snapshot or the wire, and nothing waits on it.
//
// The wire carries no oracle_id, so a card on the table is matched by its
// printed name. SIGNATURES is keyed by oracle_id, the catalog's identity,
// and signatures.test.ts checks every key and name against the server's
// catalog (server/internal/cards/effects and the oracle fixtures), so a
// name here can never drift from the card it stands for.
//
// Adding a signature: docs/adding-a-signature.md.

import type { LogEvent } from "../protocol";
import { preloadSound, type SoundName } from "../sounds";
import {
  EFFECT_BASE_MS,
  EFFECT_SHORT_CAP_MS,
  WIPE_CARD_MS,
  WIPE_SPREAD_MS,
  WIPE_STAGGER_MAX_MS,
  speedFactor,
  type EffectCue,
  type EffectMotion,
  type EffectPlan,
} from "../tableEffects";
import type { SceneModule } from "./scenes/kit";

// ---- The catalog ----

/** A scene, which one or more cards share. */
export type SceneID = "blasphemous_act" | "toxic_deluge" | "bitter_triumph" | "nicol_bolas";

export interface SignatureDef {
  /** The card's printed name, each face's when it has more than one. */
  names: readonly string[];
  scene: SceneID;
}

/**
 * oracle_id -> signature. Only cards in the catalog (each key is a
 * `server/internal/cards/effects/*.go` Spec); signatures.test.ts holds
 * the name to the catalog's.
 */
export const SIGNATURES: Readonly<Record<string, SignatureDef>> = {
  // Blasphemous Act (blasphemous_act.go)
  "7a2484a9-04fd-41a0-8224-610c1c07ed10": { names: ["Blasphemous Act"], scene: "blasphemous_act" },
  // Toxic Deluge (toxic_deluge.go)
  "afaef788-34d1-460b-b884-9d7ae6ddeb18": { names: ["Toxic Deluge"], scene: "toxic_deluge" },
  // Bitter Triumph (bitter_triumph.go)
  "776341cb-d2ec-423f-9250-92dc8bd8d503": { names: ["Bitter Triumph"], scene: "bitter_triumph" },
  // Every catalog printing named Nicol Bolas shares one scene. Today that
  // is Nicol Bolas, God-Pharaoh (nicol_bolas_god_pharaoh.go) alone.
  "cb33e07b-9599-4979-991f-df0c43ddac31": {
    names: ["Nicol Bolas, God-Pharaoh"],
    scene: "nicol_bolas",
  },
};

export type SignatureTrigger = "resolve" | "arrive";

export interface SceneSpec {
  trigger: SignatureTrigger;
  /** The sound it plays once for its batch, in the effects category. */
  sound: SoundName;
  /** Its length at speed 1. The cap (SIGNATURE_CAP_MS) holds whatever this says. */
  ms: number;
  /** Its first beat at speed 1: all that Short plays (capped at 0.8 s). */
  firstBeatMs: number;
  /** When the cards its batch took begin their own (lighter) cues. */
  cardsAtMs: number;
  /** A resolve scene with nothing to draw on when its batch took no card (a spear needs a target). */
  needsCards: boolean;
  /** The scene's code, imported on demand. */
  load: () => Promise<SceneModule>;
}

export const SCENES: Readonly<Record<SceneID, SceneSpec>> = {
  blasphemous_act: {
    trigger: "resolve",
    sound: "sig_blasphemous_act",
    ms: 2400,
    firstBeatMs: 800,
    cardsAtMs: 650,
    needsCards: false,
    load: () => import("./scenes/blasphemous_act"),
  },
  toxic_deluge: {
    trigger: "resolve",
    sound: "sig_toxic_deluge",
    ms: 2400,
    firstBeatMs: 800,
    cardsAtMs: 900,
    needsCards: false,
    load: () => import("./scenes/toxic_deluge"),
  },
  bitter_triumph: {
    trigger: "resolve",
    sound: "sig_bitter_triumph",
    ms: 1500,
    firstBeatMs: 650,
    cardsAtMs: 550,
    needsCards: true,
    load: () => import("./scenes/bitter_triumph"),
  },
  nicol_bolas: {
    trigger: "arrive",
    sound: "sig_nicol_bolas",
    ms: 2450,
    firstBeatMs: 800,
    cardsAtMs: 0,
    needsCards: false,
    load: () => import("./scenes/nicol_bolas"),
  },
};

const BY_NAME: ReadonlyMap<string, SceneID> = (() => {
  const m = new Map<string, SceneID>();
  for (const def of Object.values(SIGNATURES)) {
    for (const n of def.names) m.set(n.toLowerCase(), def.scene);
  }
  return m;
})();

/** signatureFor is the scene a card's oracle_id has, or null. */
export function signatureFor(oracleID: string): SceneID | null {
  return SIGNATURES[oracleID]?.scene ?? null;
}

/**
 * sceneForCard is the scene a card on the table has, from its printed
 * name, or null. A token has none unless it is a copy of a real card (a
 * copy of Nicol Bolas is Nicol Bolas); a face-down card has no name.
 */
export function sceneForCard(card: {
  name?: string;
  is_token?: boolean;
  token_copy?: boolean;
}): SceneID | null {
  if (!card.name) return null;
  if (card.is_token && !card.token_copy) return null;
  return BY_NAME.get(card.name.toLowerCase()) ?? null;
}

// ---- Caps ----

/** The longest a signature plays, before animations.speed (ADR 0209 §1). */
export const SIGNATURE_CAP_MS = 2500;
/** The longest Short plays: the first beat only. */
export const SIGNATURE_SHORT_CAP_MS = EFFECT_SHORT_CAP_MS;

/**
 * signatureTiming is how long a scene runs, in real ms, and `k`, the
 * factor that turns the scene's own ms (at speed 1) into real ms. Full
 * plays the whole scene, cut to the cap; Short plays the first beat,
 * squeezed to fit 0.8 s. Both then scale with animations.speed. A still
 * motion plays no scene (the generic still cues stand).
 */
export function signatureTiming(
  spec: Pick<SceneSpec, "ms" | "firstBeatMs">,
  motion: EffectMotion,
  speed: number,
): { ms: number; k: number } | null {
  if (motion === "still") return null;
  const s = speedFactor(speed);
  if (motion === "short") {
    const beat = Math.max(1, spec.firstBeatMs);
    const k = Math.min(1, SIGNATURE_SHORT_CAP_MS / beat);
    return { ms: beat * k * s, k: k * s };
  }
  const len = Math.max(1, spec.ms);
  const k = Math.min(1, SIGNATURE_CAP_MS / len);
  return { ms: len * k * s, k: k * s };
}

/**
 * signatureCueTiming is when a card inside a signature's batch begins its
 * own cue, and how long that runs. The scene hands over at cardsAtMs and
 * the cards follow a stagger apart (cue.delayMs, set by withSignatures);
 * the whole run, the last card's cue included, fits the cap (0.8 s in
 * Short), then scales with animations.speed.
 */
export function signatureCueTiming(
  cue: Pick<EffectCue, "kind" | "delayMs">,
  motion: EffectMotion,
  speed: number,
): { delayMs: number; ms: number } | null {
  if (motion === "still") return null;
  const s = speedFactor(speed);
  const len = Math.min(EFFECT_BASE_MS[cue.kind], WIPE_CARD_MS);
  const cap = motion === "short" ? SIGNATURE_SHORT_CAP_MS : SIGNATURE_CAP_MS;
  const k = Math.min(1, cap / Math.max(1, cue.delayMs + len));
  return { delayMs: cue.delayMs * k * s, ms: Math.max(1, len * k * s) };
}

// ---- Planning ----

export interface SignatureCue {
  /** Unique within a frame: the triggering entry's seq. */
  id: string;
  seq: number;
  scene: SceneID;
  /** The card behind it: the spell that resolved, or the card that arrived. */
  cardID: string;
  /** The batch its sound is capped on. */
  batch: number | null;
  /** Where an arrival lands: on the stack (a cast) or on the battlefield. */
  lands: "stack" | "battlefield" | null;
  /** The permanents its batch took, in log order (resolve). */
  took: string[];
  /** The planeswalker arrival cue it replaces (arrive), by cue id. */
  replacesCue: string | null;
}

export interface SignatureContext {
  /** The scene the card with this instance ID has, or null. */
  sceneOf(cardID: string): SceneID | null;
  /**
   * Whether this card's arrival scene played on its cast, and it has not
   * entered the battlefield since (the caller forgets it once it does).
   */
  arrived(cardID: string): boolean;
}

function isExit(e: LogEvent): boolean {
  return (e.kind === "zone" && e.old_zone === "battlefield") || e.kind === "sacrifice";
}

/**
 * planSignatures finds the signatures in a frame's new entries. It names
 * what each one would replace; withSignatures decides, from what has
 * loaded and the motion, what actually does.
 */
export function planSignatures(fresh: readonly LogEvent[], ctx: SignatureContext): SignatureCue[] {
  const out: SignatureCue[] = [];
  const entering = new Set<string>();
  for (const e of fresh) {
    if (e.kind === "zone" && e.new_zone === "battlefield" && e.old_zone !== "battlefield") {
      if (e.card_id) entering.add(e.card_id);
    }
  }
  for (const e of fresh) {
    const id = e.card_id;
    if (!id) continue;
    const scene = ctx.sceneOf(id);
    if (!scene) continue;
    const spec = SCENES[scene];
    if (spec.trigger === "resolve") {
      // A spell's resolve carries no label; an ability's carries its own.
      if (e.kind !== "resolve" || e.label) continue;
      const took: string[] = [];
      for (const x of fresh) {
        if (!isExit(x) || x.source !== id || !x.card_id) continue;
        if (!took.includes(x.card_id)) took.push(x.card_id);
      }
      if (spec.needsCards && took.length === 0) continue;
      out.push({
        id: `sig-${e.seq}`,
        seq: e.seq,
        scene,
        cardID: id,
        batch: e.batch ?? null,
        lands: null,
        took,
        replacesCue: null,
      });
      continue;
    }
    // arrive
    if (e.kind === "cast") {
      // Cast and resolved in one frame: it lands once, on the battlefield.
      if (entering.has(id)) continue;
      out.push(arrival(e, scene, "stack"));
      continue;
    }
    if (e.kind === "zone" && e.new_zone === "battlefield" && e.old_zone !== "battlefield") {
      // A permanent spell whose cast already played the scene.
      if (e.old_zone === "stack" && ctx.arrived(id)) continue;
      out.push(arrival(e, scene, "battlefield"));
    }
  }
  return out;
}

function arrival(e: LogEvent, scene: SceneID, lands: "stack" | "battlefield"): SignatureCue {
  return {
    id: `sig-${e.seq}`,
    seq: e.seq,
    scene,
    cardID: e.card_id ?? "",
    batch: e.batch ?? null,
    lands,
    took: [],
    replacesCue: lands === "battlefield" ? String(e.seq) : null,
  };
}

export interface SignaturePlan {
  /** The signatures that play. */
  sigs: SignatureCue[];
  /** The generic plan, after the signatures took their batches. */
  plan: EffectPlan;
  /** Generic cues that keep their motion but run on a signature's clock. */
  inScene: Set<string>;
  /**
   * Generic cue ids and wipe batches whose sound a signature's replaces:
   * the arrival cue, the board wipe, and the one card of a one-card batch.
   */
  silencedCues: Set<string>;
  silencedWipes: Set<number>;
}

/**
 * withSignatures folds the signatures whose scene has loaded into a
 * frame's generic plan. Each one's sound replaces its batch's sound (the
 * board wipe's, the planeswalker's arrival), in every motion mode. With
 * motion, it also replaces the batch's generic picture (the sweep, the
 * arrival ring), and the cards its batch took keep their own cue, lighter,
 * staggered from the moment the scene hands over. Still (Off, reduced
 * motion) leaves the generic still cues as they are. A signature whose
 * scene has not loaded yet changes nothing: the generic cues play.
 */
export function withSignatures(
  plan: EffectPlan,
  found: readonly SignatureCue[],
  ready: (scene: SceneID) => boolean,
  motion: EffectMotion,
): SignaturePlan {
  const sigs = found.filter((s) => ready(s.scene));
  const inScene = new Set<string>();
  const silencedCues = new Set<string>();
  const silencedWipes = new Set<number>();
  if (sigs.length === 0) return { sigs, plan, inScene, silencedCues, silencedWipes };

  const replacedWipe = new Set<string>();
  const retimed = new Map<string, { delayMs: number }>();
  const dropped = new Set<string>();
  for (const sig of sigs) {
    const spec = SCENES[sig.scene];
    if (sig.replacesCue) {
      silencedCues.add(sig.replacesCue);
      if (motion !== "still") dropped.add(sig.replacesCue);
    }
    if (spec.trigger !== "resolve") continue;
    for (const w of plan.wipes) {
      if (w.source === sig.cardID) {
        silencedWipes.add(w.batch);
        replacedWipe.add(`${w.batch}|${w.source}`);
      }
    }
    const n = sig.took.length;
    const stagger = Math.min(WIPE_STAGGER_MAX_MS, WIPE_SPREAD_MS / Math.max(1, n - 1));
    sig.took.forEach((cardID, i) => {
      for (const c of plan.cues) {
        if (c.cardID !== cardID || c.kind === "counter" || c.kind === "pw_enter") continue;
        // One card is one sound: the signature's own speaks for it.
        if (n === 1) silencedCues.add(c.id);
        if (motion !== "still") retimed.set(c.id, { delayMs: spec.cardsAtMs + i * stagger });
      }
    });
  }

  if (motion === "still") return { sigs, plan, inScene, silencedCues, silencedWipes };

  const cues: EffectCue[] = [];
  for (const c of plan.cues) {
    if (dropped.has(c.id)) continue;
    const r = retimed.get(c.id);
    if (r) {
      inScene.add(c.id);
      cues.push({ ...c, weight: "light", delayMs: r.delayMs });
    } else {
      cues.push(c);
    }
  }
  const wipes = plan.wipes.filter((w) => !replacedWipe.has(`${w.batch}|${w.source}`));
  return { sigs, plan: { cues, wipes }, inScene, silencedCues, silencedWipes };
}

// ---- Lazy loading ----

export type LoadState = "idle" | "loading" | "ready" | "failed";

export interface LoaderDeps {
  load(scene: SceneID): Promise<SceneModule>;
  /** Fetches a sound's files ahead of its first play. */
  preloadSound?(name: SoundName): void;
}

/** A load that failed is tried again on a later sighting, this many times in all. */
export const SCENE_LOAD_ATTEMPTS = 3;

/**
 * SignatureLoader imports each scene the first time a card that has it
 * appears in a game (in any zone the viewer can see), and preloads its
 * sound. Until the import resolves, ready() is null and the generic cues
 * play; nothing ever waits on it. One loader serves the whole app, so a
 * scene loads once per session.
 */
export class SignatureLoader {
  private state = new Map<SceneID, LoadState>();
  private mods = new Map<SceneID, SceneModule>();
  private tries = new Map<SceneID, number>();
  private readonly deps: LoaderDeps;

  constructor(deps: LoaderDeps) {
    this.deps = deps;
  }

  /** notice takes the cards a frame shows and starts loading any scene among them. */
  notice(cards: Iterable<{ name?: string; is_token?: boolean; token_copy?: boolean }>): void {
    for (const c of cards) {
      const scene = sceneForCard(c);
      if (scene) this.want(scene);
    }
  }

  /** want starts loading one scene, unless it is loading, loaded or out of tries. */
  want(scene: SceneID): void {
    const st = this.state.get(scene) ?? "idle";
    if (st === "loading" || st === "ready") return;
    const n = this.tries.get(scene) ?? 0;
    if (n >= SCENE_LOAD_ATTEMPTS) return;
    this.tries.set(scene, n + 1);
    this.state.set(scene, "loading");
    this.deps.preloadSound?.(SCENES[scene].sound);
    this.deps.load(scene).then(
      (mod) => {
        this.mods.set(scene, mod);
        this.state.set(scene, "ready");
      },
      () => {
        this.state.set(scene, "failed");
      },
    );
  }

  /** The loaded scene, or null while it is not. */
  ready(scene: SceneID): SceneModule | null {
    return this.mods.get(scene) ?? null;
  }

  status(scene: SceneID): LoadState {
    return this.state.get(scene) ?? "idle";
  }
}

/** The app's loader: scenes from fx/scenes/, sounds through sounds.ts. */
export const signatureScenes = new SignatureLoader({
  load: (scene) => SCENES[scene].load(),
  preloadSound: (name) => preloadSound(name),
});
