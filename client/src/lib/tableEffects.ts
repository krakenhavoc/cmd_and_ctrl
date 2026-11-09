// tableEffects — the generic effect cues (ADR 0209 §1 and §3, #4396).
//
// A permanent that is destroyed, exiled, sacrificed or bounced, a spell
// that is countered, a planeswalker that arrives or goes, a token that is
// created (#4441), counters put on a permanent (#4442), a commander cast
// from the command zone or put back into it (#4443), an Equipment
// attached or an Aura enchanting a permanent (#4444), cards milled or
// discarded into a graveyard (#4445), cards returned from a graveyard to
// the battlefield or a hand (#4446), a library searched (#4447), a
// double-faced permanent transforming or the game becoming day or night
// (#4448), a permanent sacrificed for its own ability (a fetch land, a
// Treasure: #4427) and a board wipe
// each get one cue over the table, so a player sees what happened before
// they read the log. TableEffects.svelte measures, draws and plays; every
// rule is here, pure and unit-tested:
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
import { COUNTER_MINUS_ONE, COUNTER_PLUS_ONE } from "./counterTypes";
import { entryItemID, outcomeOf } from "./stackLinger";
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
  // #4450: a spell or ability that left the stack with no legal target
  // (CR 608.2b), the counterspell's quiet cousin.
  | "fizzle"
  // #4450: one source made every opponent lose life at once (a drain), and
  // a player became the monarch or took the initiative.
  | "drain_table"
  | "monarch"
  | "initiative"
  | "pw_enter"
  | "pw_leave"
  // S80 (#4441 on): the common moments.
  | "token_create"
  // Counters put on a permanent (#4442). Not `counter`, the counterspell.
  | "counters"
  // A commander cast from the command zone (CR 903.8), or put back into
  // it (CR 903.9a, 903.9b) (#4443).
  | "commander_cast"
  | "commander_return"
  // An Equipment (or a Fortification) attached to a permanent, or an
  // Aura enchanting one (CR 701.3a, 301.5, 303.4) (#4444).
  | "attach_equip"
  | "attach_aura"
  // Cards put into a graveyard without dying (#4445): milled from a
  // library (CR 701.17a) or discarded from a hand (CR 701.9a).
  | "mill"
  | "discard"
  // A card returned from a graveyard to the battlefield (Reanimate,
  // unearth, CR 702.84a, undying, a land put back) or to a hand
  // (Regrowth) (#4446).
  | "reanimate"
  // A player searched a library (CR 701.23a): a tutor, a fetch land's
  // effect (#4447).
  | "library_search"
  // A double-faced permanent turned over to its other face (CR 701.27a,
  // a convert, CR 701.28a, or two cards melded into one, CR 701.42a),
  // and the game becoming day or night (CR 731.1) (#4448).
  | "transform"
  | "day_night"
  // A permanent sacrificed as a cost of its own ability (#4427): a fetch
  // land cracked, a Treasure, Clue or Food token spent. Quick, with no
  // altar, in place of the `sacrifice` cue.
  | "crack";

/**
 * Base length of each cue at speed 1, before the cap. About 1.35 times
 * the first build's (#4416: "so you can see it"). A sacrifice is longer
 * again, because it is now an altar, a knife and then the destroy; a
 * counterspell holds its cracks while the stack linger shows the spell;
 * a planeswalker that dies now also withdraws from the plane. The S80
 * cues are their issues' own lengths: a token's pop-in is about 0.4 s,
 * a counter's pulse a little longer, so its tick can be read, a
 * commander's rise to the stack or drift home about 0.8 s, an
 * attachment's slide onto its host about 0.4 s (an Aura's shimmer a
 * beat longer, so it can fade), one card's slide (mill) or toss
 * (discard) into a graveyard about half a second, and one card's rise out
 * of a graveyard (reanimate) about 0.8 s (pileBaseMs for more), and a
 * riffle through a library pile while it is searched about half a second.
 * A double-faced card's flip to its other face is about 0.6 s, and the
 * sky's wash across the table as it becomes day or night about a second.
 */
export const EFFECT_BASE_MS: Record<EffectKind, number> = {
  destroy: 1250,
  exile: 1500,
  // The existing crumble (STRIKE_CRUMBLE_MS in animations.ts, 420 ms in combat).
  dies: 560,
  sacrifice: 1750,
  bounce: 1300,
  counter: 1600,
  // #4450: a soft dissolve on the stack item, inside its sound's 0.8 s.
  fizzle: 800,
  // #4450: a thin dark line from the source to each player, 0.8 s; a short
  // glint on the new holder's panel, inside its sound's 1 s.
  drain_table: 800,
  monarch: 700,
  initiative: 700,
  pw_enter: 1200,
  pw_leave: 1500,
  token_create: 400,
  counters: 450,
  commander_cast: 800,
  commander_return: 800,
  attach_equip: 400,
  attach_aura: 450,
  // One card's slide or toss into the graveyard (#4445); a run of them
  // is longer, pileBaseMs.
  mill: 520,
  discard: 420,
  // One card's rise out of the graveyard (#4446), pileBaseMs for more.
  reanimate: 800,
  // A riffle through the library pile (#4447).
  library_search: 500,
  // A card's flip to its other face, and the sky's wash (#4448).
  transform: 600,
  day_night: 1000,
  // A self-sacrifice's snap (#4427): it plays several times a game.
  crack: 600,
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
  fizzle: "fizzle",
  drain_table: "drain_table",
  monarch: "monarch",
  initiative: "initiative",
  pw_enter: "pw_enter",
  pw_leave: "pw_leave",
  token_create: "token_create",
  // Which of two sounds a counter cue plays is its tone's (COUNTER_SOUND).
  counters: null,
  commander_cast: "commander_cast",
  commander_return: "commander_return",
  attach_equip: "attach_equip",
  attach_aura: "attach_aura",
  mill: "mill",
  discard: "discard",
  reanimate: "reanimate",
  library_search: "library_search",
  transform: "transform",
  day_night: "day_night",
  // Which sound a crack plays is its permanent's (CRACK_SOUND).
  crack: null,
};

/**
 * What a permanent sacrificed for its own ability was (#4427), for its
 * crack's sound and dust: a land (a fetch land, Evolving Wilds), a
 * Treasure, a Clue or a Food token, or anything else (Blood, Map,
 * Powerstone, Lotus Petal, Mind Stone, a Seal).
 */
export type CrackKind = "land" | "treasure" | "clue" | "food" | "other";

/**
 * Each crack's sound. Any other self-sacrificing permanent plays the
 * Treasure's, the issue's generic choice for it.
 */
export const CRACK_SOUND: Record<CrackKind, SoundName> = {
  land: "fetch_crack",
  treasure: "treasure_crack",
  clue: "clue_crack",
  food: "food_crack",
  other: "treasure_crack",
};

/**
 * A crack plays a little quieter than the other effect cues (#4427): it
 * comes up several times a game, often on turn one.
 */
export const CRACK_VOLUME = 0.75;

/**
 * isSelfCost: a `sacrifice` entry paid as a cost of the sacrificed
 * permanent's own ability (#4427, CR 602.2b, 605.3b): `cause` "cost" and
 * `source` its own `card_id`. A cost the permanent paid for another's
 * ability (a creature fed to Viscera Seer), a spell's additional cost
 * and an effect's sacrifice are not.
 */
export function isSelfCost(e: LogEvent): boolean {
  return e.kind === "sacrifice" && e.cause === "cost" && !!e.card_id && e.source === e.card_id;
}

/**
 * crackKind is the crack a permanent sacrificed for its own ability gets,
 * read off its type line as the client last saw it (#4427), or null when
 * it keeps the sacrifice's altar and knife: a creature (Sakura-Tribe
 * Elder, an Eldrazi Spawn), which is what the altar is for. A land is a
 * land first, whatever else it is; a Treasure, Clue or Food is its own
 * (CR 111.10a, 111.10b, 111.10f), and anything else, or a card the client
 * never saw, is "other".
 */
export function crackKind(typeLine: string | undefined): CrackKind | null {
  const t = typeLine ?? "";
  if (/\bcreature\b/i.test(t)) return null;
  if (/\bland\b/i.test(t)) return "land";
  if (/\btreasure\b/i.test(t)) return "treasure";
  if (/\bclue\b/i.test(t)) return "clue";
  if (/\bfood\b/i.test(t)) return "food";
  return "other";
}

/**
 * A card returned to a hand (Regrowth) plays the reanimate sound this
 * much quieter than one returned to the battlefield (#4446).
 */
export const REANIMATE_HAND_VOLUME = 0.6;

/**
 * What a counter cue is about (#4442): +1/+1 counters (green, a rising
 * tick), -1/-1 counters (purple, a falling one) or any other kind
 * (charge, oil, stun, a keyword counter: a neutral pulse and no sound).
 */
export type CounterTone = "plus" | "minus" | "other";

export const COUNTER_SOUND: Record<CounterTone, SoundName | null> = {
  plus: "counter_add",
  minus: "counter_remove",
  other: null,
};

export function counterTone(label: string): CounterTone {
  if (label === COUNTER_PLUS_ONE) return "plus";
  if (label === COUNTER_MINUS_ONE) return "minus";
  return "other";
}

/**
 * Tokens created together (#4441): one sound for the batch, and the
 * pop-ins a stagger apart, spread over TOKEN_SPREAD_MS at most, so the
 * whole batch fits even the Short cap (EFFECT_SHORT_CAP_MS) before speed.
 */
export const TOKEN_STAGGER_MS = 90;
export const TOKEN_SPREAD_MS = 600;

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
  /** For a counter or a fizzle (#4450): the stack item that left. */
  stackItemID: string | null;
  batch: number | null;
  /** "light" inside a board wipe: shorter, no flourish, a quieter sound. */
  weight: "full" | "light";
  /** Start offset at speed 1 (a wipe staggers its per-card cues, a token batch its pop-ins). */
  delayMs: number;
  /**
   * A later cue of a batch an earlier cue already announces (the second
   * and later tokens created together, #4441): it plays no sound of its
   * own and does not count again toward a busy table.
   */
  follows?: boolean;
  /** For a `counters` cue: which counters were put on, and their kind as the log names it. */
  counterTone?: CounterTone;
  counterLabel?: string;
  /**
   * For a `commander_return` cue: the zone the commander left (the log's
   * `old_zone`), which is where its drift home starts (#4443).
   */
  commanderFrom?: string;
  /** For an attach cue: the permanent the attachment is now on (#4444). */
  hostID?: string;
  /**
   * For a mill or discard cue (#4445): the cards it slides, in order (as
   * many as the log names), how many went (a mill's `amount`; a big one
   * names only the first few), and the seat whose library or hand they
   * left, as the log's seat index.
   */
  cardIDs?: string[];
  count?: number;
  seat?: number;
  /**
   * For a reanimate cue (#4446): where its cards went. Its `cardIDs` are
   * the cards that rose out of a graveyard in one batch, in log order,
   * `count` how many, and `seat` the first entry's seat (who returned
   * them; not always whose graveyard it was).
   */
  reanimateTo?: "battlefield" | "hand";
  /**
   * For a library_search cue (#4447): whose library was searched, as the
   * log's seat index (not always the searcher's: Bribery searches an
   * opponent's). `seat` is the searcher.
   */
  librarySeat?: number;
  /**
   * For a drain_table cue (#4450): the log seat indexes that lost life,
   * ascending. `cardID` is the source. For a monarch or initiative cue
   * `seat` is the new holder.
   */
  drainSeats?: number[];
  /**
   * For a transform cue (#4448): the permanent was melded out of two
   * cards (CR 701.42a), so the face it turns from is its first card's,
   * not its own other face. For a day_night cue: what the game became.
   */
  melded?: boolean;
  sky?: "day" | "night";
  /** For a crack cue (#4427): what was sacrificed, for its sound and its dust. */
  crackOf?: CrackKind;
  /**
   * The key the cue's sound is capped on in place of its batch. A
   * crack's is its payment's (#4427): a player's separate activations in
   * one priority window share an event batch, and each is a moment.
   */
  soundKey?: string;
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
  /**
   * Whether this card is a commander in a command zone now (#4443). A
   * card put into the command zone that is not one (an Attraction going
   * to its junkyard, CR 717.6a) gets no commander_return cue. Absent:
   * no card is, so no commander_return cue at all.
   */
  isCommander?(cardID: string): boolean;
  /**
   * The type line the client last saw for this card (#4427): a
   * permanent sacrificed for its own ability has just left, so this is
   * the view before it went. Absent or undefined: not known.
   */
  typeLineOf?(cardID: string): string | undefined;
}

const DAMAGE_STEPS = new Set(["first_strike_damage", "combat_damage"]);

interface Exit {
  e: LogEvent;
  step: string;
  /** The zone entry for this card in this batch, if there is one. */
  zone: LogEvent | null;
  sacrificed: boolean;
  /** Sacrificed for its own ability, and not a creature: its crack (#4427). */
  crack?: CrackKind | null;
}

/**
 * exitKind is the cue a permanent leaving the battlefield gets, or null
 * for none. `new_zone` is read first: an exile is an exile whatever the
 * instruction said, and a card that went to a hand is a bounce.
 */
export function exitKind(
  x: Pick<Exit, "zone" | "sacrificed" | "step" | "crack">,
  ctx: PlanContext,
): EffectKind | null {
  const z = x.zone;
  // A destruction replaced by exile (Rest in Peace) is `destroy` with
  // new_zone exile: it must show the exile look, never a death.
  if (z?.new_zone === "exile") return "exile";
  if (z?.new_zone === "hand" || (z && z.cause === "bounce")) return "bounce";
  // #4427: one sacrifice is one cue. Sacrificed for its own ability, it
  // cracks; otherwise it is laid on the altar.
  if (x.sacrificed) return x.crack ? "crack" : "sacrifice";
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
  const tokens: LogEvent[] = [];
  const counters: LogEvent[] = [];
  const commanders: LogEvent[] = [];
  const attaches: LogEvent[] = [];
  const piles: { kind: "mill" | "discard"; e: LogEvent }[] = [];
  const rises: LogEvent[] = [];
  const searches: LogEvent[] = [];
  const turns: LogEvent[] = [];
  const losses: LogEvent[] = [];
  const crowns: LogEvent[] = [];
  const damaged = new Set<string>();
  for (const e of log ?? []) {
    if (e.kind === "step") {
      step = e.step ?? "";
      continue;
    }
    if (!freshSeqs.has(e.seq)) continue;

    // #4450: a drain's life losses (grouped below), and the crown or the
    // initiative changing hands.
    if (e.kind === "life") {
      losses.push(e);
      continue;
    }
    if (e.kind === "damage") {
      if (e.batch != null && e.card_id) damaged.add(`${e.batch}|${e.card_id}`);
    }
    if (e.kind === "monarch" || e.kind === "initiative") {
      crowns.push(e);
      continue;
    }

    // A token arrives from no zone: the engine logs its creation as a
    // `token` entry, never a `zone` one as well (#762).
    if (e.kind === "token") {
      tokens.push(e);
      continue;
    }
    if (e.kind === "counters") {
      counters.push(e);
      continue;
    }
    // #4448: a permanent turned over, two cards melded, or the game
    // becoming day or night. None of them is a zone move.
    if (e.kind === "transform" || e.kind === "meld" || e.kind === "day_night") {
      turns.push(e);
      continue;
    }
    // #4447: a library searched, told by its `search` line or by the
    // reveal that took that line's place. Neither is anything else here.
    if (searchedLibrary(e) !== null) {
      searches.push(e);
      continue;
    }
    if (e.kind === "attach") {
      if (attachKind(e)) attaches.push(e);
      continue;
    }
    // #4443: before the exits, so a commander that goes home from the
    // battlefield (CR 903.9b: a bounce or a tuck its owner sent to the
    // command zone instead) is this cue, never a bounce.
    if (isCommanderCast(e) || isCommanderReturn(e, ctx)) {
      commanders.push(e);
      continue;
    }
    // #4445: into a graveyard from a library or a hand, never from the
    // battlefield, so never a death.
    const pk = pileKind(e);
    if (pk) {
      piles.push({ kind: pk, e });
      continue;
    }
    // #4446: out of a graveyard onto the battlefield or into a hand. An
    // arrival on the battlefield is still one below too (a planeswalker's).
    const to = reanimateTo(e);
    if (to) rises.push(e);
    if (to === "hand") continue;

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

    // #4450: a fizzle, read as the stack linger reads it.
    if (e.kind === "fizzle") {
      const cue = fizzleCue(e);
      if (cue) cues.push(cue);
      continue;
    }

    if (e.kind === "zone" && e.new_zone === "battlefield" && e.old_zone !== "battlefield") {
      if (e.card_id && ctx.isPlaneswalker(e.card_id)) {
        const cue: EffectCue = {
          id: String(e.seq),
          seq: e.seq,
          kind: "pw_enter",
          cardID: e.card_id,
          stackItemID: null,
          batch: e.batch ?? null,
          weight: "full",
          delayMs: 0,
        };
        // A planeswalker returned from a graveyard (#4446) is one arrival:
        // its rise plays the sound, and its arrival here only draws.
        if (to) cue.follows = true;
        cues.push(cue);
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
    if (e.kind === "sacrifice") {
      x.sacrificed = true;
      if (isSelfCost(e)) x.crack = crackKind(ctx.typeLineOf?.(e.card_id ?? ""));
    } else x.zone = e;
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
    if (kind === "crack") cue.crackOf = x.crack ?? "other";
    cues.push(cue);
  }
  crackIsOneSound(cues);
  // A token created attached (a Role, a token copy of an Aura spell) is
  // one arrival with two cues: its attach cue plays the sound, and its
  // token cue only pops it in (#4444).
  const attached = new Set(attaches.map(arrivalKey));
  cues.push(...tokenCues(tokens, attached));
  cues.push(...counterCues(counters));
  cues.push(...commanderCues(commanders));
  // An Aura returned from a graveyard attached (#4446) is one arrival
  // too: its rise plays the sound, and its attach cue only draws.
  cues.push(...attachCues(attaches, new Set(rises.map(arrivalKey))));
  cues.push(...pileCues(piles));
  cues.push(...reanimateCues(rises));
  cues.push(...searchCues(searches));
  cues.push(...drainCues(losses, damaged));
  cues.push(...crownCues(crowns));
  cues.push(...turnCues(turns, ctx));
  waitForFlips(cues);
  meldIsOneSound(cues, turns);
  cues.sort((a, b) => a.seq - b.seq);
  return { cues, wipes };
}

/**
 * tokenCues gives each created token its pop-in (#4441). Tokens created
 * in one batch are one moment: the first plays the sound and the rest
 * follow it, a stagger apart, in log order. A token with no batch is a
 * moment of its own. A token that entered attached (in `attached`, by
 * arrivalKey) leaves the sound to its attach cue (#4444), so the batch's
 * first other token leads, and a batch of only those plays none here.
 */
function tokenCues(tokens: readonly LogEvent[], attached: ReadonlySet<string>): EffectCue[] {
  const groups = new Map<string, LogEvent[]>();
  for (const e of tokens) {
    const k = e.batch === undefined || e.batch === null ? `q${e.seq}` : String(e.batch);
    const g = groups.get(k);
    if (g) g.push(e);
    else groups.set(k, [e]);
  }
  const out: EffectCue[] = [];
  for (const [, g] of groups) {
    const stagger = Math.min(TOKEN_STAGGER_MS, TOKEN_SPREAD_MS / Math.max(1, g.length - 1));
    const lead = g.findIndex((e) => !attached.has(arrivalKey(e)));
    g.forEach((e, i) => {
      const cue: EffectCue = {
        id: String(e.seq),
        seq: e.seq,
        kind: "token_create",
        cardID: e.card_id ?? null,
        stackItemID: null,
        batch: e.batch ?? null,
        weight: "full",
        delayMs: i * stagger,
      };
      if (i !== lead) cue.follows = true;
      out.push(cue);
    });
  }
  return out;
}

/**
 * counterCues gives each permanent that had counters put on it a pulse
 * (#4442). Only a placement is a cue: the entry's `delta` is above zero
 * (a removal, the CR 704.5q cancel included, is not). An entry with no
 * `delta` (a server that predates it) or no `label` (a card this viewer
 * may not identify) says too little and is skipped. Counters put on in
 * one batch are one moment, so a proliferate or a "put a +1/+1 counter
 * on each creature you control" plays one sound: the first +1/+1 or
 * -1/-1 cue of the batch plays it, and every other cue of the batch
 * follows. The pulses all play at once. One card's entries of one kind
 * in one batch are one pulse.
 */
function counterCues(entries: readonly LogEvent[]): EffectCue[] {
  const groups = new Map<string, EffectCue[]>();
  const seen = new Set<string>();
  for (const e of entries) {
    if (!e.card_id || !e.label || (e.delta ?? 0) <= 0) continue;
    const tone = counterTone(e.label);
    const k = e.batch === undefined || e.batch === null ? `q${e.seq}` : String(e.batch);
    const one = `${k}|${e.card_id}|${e.label}`;
    if (seen.has(one)) continue;
    seen.add(one);
    const cue: EffectCue = {
      id: String(e.seq),
      seq: e.seq,
      kind: "counters",
      cardID: e.card_id,
      stackItemID: null,
      batch: e.batch ?? null,
      weight: "full",
      delayMs: 0,
      counterTone: tone,
      counterLabel: e.label,
    };
    const g = groups.get(k);
    if (g) g.push(cue);
    else groups.set(k, [cue]);
  }
  const out: EffectCue[] = [];
  for (const [, g] of groups) {
    const lead = Math.max(
      0,
      g.findIndex((c) => COUNTER_SOUND[c.counterTone ?? "other"] !== null),
    );
    g.forEach((c, i) => {
      if (i !== lead) c.follows = true;
      out.push(c);
    });
  }
  return out;
}

/**
 * isCommanderCast: a spell cast from the command zone (#4443). The `cast`
 * entry names a non-hand origin in `old_zone`, and only a commander is
 * cast from there (CR 903.8).
 */
export function isCommanderCast(e: LogEvent): boolean {
  return e.kind === "cast" && e.old_zone === "command" && !!e.card_id;
}

/**
 * isCommanderReturn: a commander put into the command zone (#4443), from
 * anywhere: its graveyard or exile by the state-based action (CR
 * 903.9a), or instead of its hand or library by the replacement (CR
 * 903.9b), whose `zone` entry already names the command zone.
 */
export function isCommanderReturn(e: LogEvent, ctx: PlanContext): boolean {
  if (e.kind !== "zone" || e.new_zone !== "command" || e.old_zone === "command") return false;
  return !!e.card_id && ctx.isCommander?.(e.card_id) === true;
}

/**
 * commanderCues gives each commander cast from the command zone its rise
 * to the stack, and each that went home its drift back (#4443). One kind
 * in one batch is one moment (two partners sent home by one bounce): the
 * first plays the sound and the rest follow it.
 */
function commanderCues(entries: readonly LogEvent[]): EffectCue[] {
  const lead = new Set<string>();
  return entries.map((e) => {
    const kind: EffectKind = e.kind === "cast" ? "commander_cast" : "commander_return";
    const cue: EffectCue = {
      id: String(e.seq),
      seq: e.seq,
      kind,
      cardID: e.card_id ?? null,
      // A spell's stack item ID is its card's.
      stackItemID: kind === "commander_cast" ? (e.card_id ?? null) : null,
      batch: e.batch ?? null,
      weight: "full",
      delayMs: 0,
    };
    if (kind === "commander_return") cue.commanderFrom = e.old_zone ?? "";
    const k = `${kind}|${e.batch ?? `q${e.seq}`}`;
    if (lead.has(k)) cue.follows = true;
    else lead.add(k);
    return cue;
  });
}

/**
 * fizzleCue is the cue a `fizzle` entry gets (#4450): a spell or ability
 * the game removed from the stack as it tried to resolve, because every
 * target it had was illegal (CR 608.2b). It dissolves where it sat on the
 * stack. A `fizzle` entry with call "condition" is not one: an optional
 * trigger whose intervening "if" no longer held (CR 603.4) lapsed, and
 * the stack linger says so on its own. An entry that names no stack item
 * has nowhere to dissolve, and still plays its sound.
 */
export function fizzleCue(e: LogEvent): EffectCue | null {
  if (e.kind !== "fizzle" || outcomeOf(e) !== "fizzled") return null;
  return {
    id: String(e.seq),
    seq: e.seq,
    kind: "fizzle",
    cardID: e.card_id ?? null,
    stackItemID: entryItemID(e),
    batch: e.batch ?? null,
    weight: "full",
    delayMs: 0,
  };
}

/**
 * drainCues gives one source's life drain on the table one cue (#4450):
 * `life` entries that lose life (a negative amount), name a source card
 * and share one batch, for two or more different players. That is one
 * resolution making several players lose life at once (Syr Konrad,
 * Zulaport Cutthroat, a "each opponent loses 1 life" trigger), not the
 * separate hits the log shows: the cue is one line from the source to
 * each player and one sound. A source that also dealt damage in the
 * batch is a damage spell, not a drain, and a single player's loss
 * (Blood Artist) is no cue: it stays a life number. Entries with no
 * batch, or no source, can't be grouped and are never one. No wire
 * change: `life` already carries `batch` and `card_id` (the source).
 */
export function drainCues(
  entries: readonly LogEvent[],
  damaged: ReadonlySet<string> = new Set(),
): EffectCue[] {
  const groups = new Map<string, LogEvent[]>();
  for (const e of entries) {
    if ((e.amount ?? 0) >= 0 || e.batch == null || !e.card_id || e.seat < 0) continue;
    const k = `${e.batch}|${e.card_id}`;
    if (damaged.has(k)) continue;
    const g = groups.get(k);
    if (g) g.push(e);
    else groups.set(k, [e]);
  }
  const out: EffectCue[] = [];
  for (const g of groups.values()) {
    const seats = [...new Set(g.map((e) => e.seat))].sort((a, b) => a - b);
    if (seats.length < 2) continue;
    out.push({
      id: `${g[0].seq}:drain`,
      seq: g[0].seq,
      kind: "drain_table",
      cardID: g[0].card_id ?? null,
      stackItemID: null,
      batch: g[0].batch ?? null,
      weight: "full",
      delayMs: 0,
      seat: g[0].seat,
      drainSeats: seats,
    });
  }
  return out;
}

/**
 * crownCues is a glint on the panel of a player who becomes the monarch
 * or takes the initiative (#4450). `seat` is the new holder.
 */
export function crownCues(entries: readonly LogEvent[]): EffectCue[] {
  const out: EffectCue[] = [];
  for (const e of entries) {
    if ((e.kind !== "monarch" && e.kind !== "initiative") || e.seat < 0) continue;
    out.push({
      id: `${e.seq}:${e.kind}`,
      seq: e.seq,
      kind: e.kind,
      cardID: e.card_id ?? null,
      stackItemID: null,
      batch: e.batch ?? null,
      weight: "full",
      delayMs: 0,
      seat: e.seat,
    });
  }
  return out;
}

/**
 * attachKind is the cue an `attach` entry gets (#4444): an Equipment or a
 * Fortification (the same metal clasp) is attach_equip, an Aura
 * attach_aura, read off the entry's `label`, which names what the
 * attachment is as it attaches. Only a permanent host is a cue: a Curse
 * enchanting a player has no tile to land on. An entry with no label (a
 * viewer who may not identify the attachment, or a permanent that is
 * none of the three) is no cue.
 */
export function attachKind(e: LogEvent): EffectKind | null {
  if (e.kind !== "attach" || !e.card_id || !e.target) return null;
  if (e.label === "Aura") return "attach_aura";
  if (e.label === "Equipment" || e.label === "Fortification") return "attach_equip";
  return null;
}

/** One card's arrival in one batch: its batch (or its own seq) and its ID. */
function arrivalKey(e: LogEvent): string {
  return `${e.batch ?? `q${e.seq}`}|${e.card_id ?? ""}`;
}

/**
 * attachCues gives each attachment its slide onto its host (#4444). One
 * kind in one batch is one moment (Armory Automaton's "attach any number
 * of target Equipment to it"): the first plays the sound and the rest
 * follow it. An Equipment and an Aura in one batch are two moments, each
 * with its own sound. An attachment that rose out of a graveyard in the
 * same batch (in `risen`, by arrivalKey) leaves the sound to its
 * reanimate cue (#4446), so the batch's first other attachment leads.
 */
function attachCues(
  entries: readonly LogEvent[],
  risen: ReadonlySet<string> = new Set(),
): EffectCue[] {
  const lead = new Set<string>();
  const out: EffectCue[] = [];
  for (const e of entries) {
    const kind = attachKind(e);
    if (!kind) continue;
    const cue: EffectCue = {
      id: String(e.seq),
      seq: e.seq,
      kind,
      cardID: e.card_id ?? null,
      stackItemID: null,
      batch: e.batch ?? null,
      weight: "full",
      delayMs: 0,
      hostID: e.target,
    };
    const k = `${kind}|${e.batch ?? `q${e.seq}`}`;
    if (risen.has(arrivalKey(e)) || lead.has(k)) cue.follows = true;
    else lead.add(k);
    out.push(cue);
  }
  return out;
}

/**
 * pileKind is the cue an entry about cards going into a graveyard without
 * dying gets (#4445), or null. A `mill` entry (CR 701.17a) and a `zone`
 * entry from a library (surveil's graveyard half, CR 701.25a, an explore's
 * nonland, a sandbox move) are a mill; a `discard` entry (CR 701.9a), a
 * `cycle` entry (its discard, CR 702.29a, which the log tells in its
 * place) and a `zone` entry from a hand are a discard. Only into a
 * graveyard: a madness card discarded into exile (CR 702.35a) is no cue.
 */
export function pileKind(e: LogEvent): "mill" | "discard" | null {
  if (e.kind === "mill") return (e.amount ?? 0) > 0 ? "mill" : null;
  if (e.new_zone !== "graveyard") return null;
  if (e.kind === "discard" || e.kind === "cycle") return e.card_id ? "discard" : null;
  if (e.kind !== "zone" || !e.card_id) return null;
  if (e.old_zone === "library") return "mill";
  if (e.old_zone === "hand") return "discard";
  return null;
}

/**
 * A mill or discard of this many cards or more is one sweep, not one
 * card after another (#4445): a handful moving into the graveyard as one.
 */
export const PILE_SWEEP_MIN = 5;
/** The most cards a mill or discard slides one after another. */
export const PILE_SHOWN_MAX = 4;
/** How far into one card's slide the next one sets off, as a share of it. */
export const PILE_STAGGER = 0.3;
/** The share of a slide at which its card reaches the graveyard. */
export const PILE_LAND = 0.82;

/**
 * How many cards a mill, a discard or a reanimate draws: one by one, or a
 * sweep of three.
 */
export function pileShown(count: number): number {
  return count >= PILE_SWEEP_MIN ? 3 : Math.max(1, Math.min(PILE_SHOWN_MAX, count));
}

/** The cues that move a run of cards to or from a graveyard pile (#4445, #4446). */
export type PileKind = "mill" | "discard" | "reanimate";

export function isPileKind(k: EffectKind): k is PileKind {
  return k === "mill" || k === "discard" || k === "reanimate";
}

/**
 * pileBaseMs is a mill, discard or reanimate cue's length at speed 1,
 * before the cap: one card's slide, plus a stagger for each card after
 * it. A sweep moves as one, a little slower than a single card.
 */
export function pileBaseMs(kind: PileKind, count: number): number {
  const one = EFFECT_BASE_MS[kind];
  if (count >= PILE_SWEEP_MIN) return one * 1.25;
  return one * (1 + PILE_STAGGER * (pileShown(count) - 1));
}

/**
 * pileCues gives each player whose cards went into their graveyard in one
 * batch one cue (#4445): a mill's line already counts its cards, and a
 * discard is one line per card, so one batch's discards from one hand are
 * gathered into one cue, in log order. One kind in one batch is one
 * moment ("each player mills three", a Windfall): the first cue plays
 * the sound and the rest follow it.
 */
function pileCues(entries: readonly { kind: "mill" | "discard"; e: LogEvent }[]): EffectCue[] {
  const byKey = new Map<string, EffectCue>();
  const out: EffectCue[] = [];
  const lead = new Set<string>();
  for (const { kind, e } of entries) {
    const batch = e.batch ?? null;
    const at = batch === null ? `q${e.seq}` : String(batch);
    const k = `${kind}|${at}|${e.seat}`;
    const ids = e.kind === "mill" ? (e.card_ids ?? []) : e.card_id ? [e.card_id] : [];
    const n = e.kind === "mill" ? (e.amount ?? 0) : 1;
    const cue = byKey.get(k);
    if (cue) {
      cue.cardIDs = [...(cue.cardIDs ?? []), ...ids.filter((id) => !cue.cardIDs?.includes(id))];
      cue.count = (cue.count ?? 0) + n;
      continue;
    }
    const c: EffectCue = {
      id: String(e.seq),
      seq: e.seq,
      kind,
      cardID: ids[0] ?? null,
      stackItemID: null,
      batch,
      weight: "full",
      delayMs: 0,
      cardIDs: [...ids],
      count: n,
      seat: e.seat,
    };
    const one = `${kind}|${at}`;
    if (lead.has(one)) c.follows = true;
    else lead.add(one);
    byKey.set(k, c);
    out.push(c);
  }
  return out;
}

/**
 * reanimateTo is where a card that left a graveyard went, for the
 * reanimate cue (#4446), or null. Only a `zone` entry from a graveyard
 * to the battlefield (Reanimate, Living Death, unearth, CR 702.84a,
 * undying, CR 702.93a, a land put back, or one played from a graveyard)
 * or to a hand (Regrowth, Eternal Witness; always its owner's, CR 400.3)
 * is one. A token copy of a card in a graveyard is a `token` entry, not
 * a move (the card stays where it is), and a token that left the
 * battlefield never comes back (CR 111.8), so no token is a reanimation.
 * A commander its owner puts into the command zone from there is
 * commander_return; a spell cast from a graveyard is a `cast` entry.
 */
export function reanimateTo(e: LogEvent): "battlefield" | "hand" | null {
  if (e.kind !== "zone" || e.old_zone !== "graveyard" || !e.card_id) return null;
  if (e.new_zone === "battlefield") return "battlefield";
  if (e.new_zone === "hand") return "hand";
  return null;
}

/**
 * reanimateCues gives one batch's cards that rose out of a graveyard to
 * one destination (the battlefield, or a hand) one cue, in log order:
 * several cards, from one graveyard or more, rise a stagger apart, five
 * or more as one sweep. One batch is one moment (Living Death, "return
 * up to two"): the first cue plays the sound and a batch's other cue
 * follows it.
 */
function reanimateCues(entries: readonly LogEvent[]): EffectCue[] {
  const byKey = new Map<string, EffectCue>();
  const out: EffectCue[] = [];
  const lead = new Set<string>();
  for (const e of entries) {
    const to = reanimateTo(e);
    const id = e.card_id;
    if (!to || !id) continue;
    const batch = e.batch ?? null;
    const at = batch === null ? `q${e.seq}` : String(batch);
    const k = `${at}|${to}`;
    const cue = byKey.get(k);
    if (cue) {
      if (!cue.cardIDs?.includes(id)) {
        cue.cardIDs = [...(cue.cardIDs ?? []), id];
        cue.count = (cue.count ?? 0) + 1;
      }
      continue;
    }
    const c: EffectCue = {
      // A second cue of its entry, beside a planeswalker's arrival.
      id: `${e.seq}:reanimate`,
      seq: e.seq,
      kind: "reanimate",
      cardID: id,
      stackItemID: null,
      batch,
      weight: "full",
      delayMs: 0,
      cardIDs: [id],
      count: 1,
      seat: e.seat,
      reanimateTo: to,
    };
    if (lead.has(at)) c.follows = true;
    else lead.add(at);
    byKey.set(k, c);
    out.push(c);
  }
  return out;
}

/**
 * searchedLibrary is the seat whose library an entry says was searched
 * (#4447, CR 701.23a), or null. The server writes one entry per search,
 * with `library_seat`: its `search` line, or, when the search revealed
 * what it took (a fetch land, a "reveal it" tutor, Bribery), the
 * `reveal` line that already told it in that line's place. A `search`
 * from a server older than `library_seat` reads its `target_seat`. A
 * plain shuffle (`label` "shuffle") is no search, and a reveal that no
 * search made has no `library_seat`. Nothing here says what was found,
 * or whether anything was (CR 701.23b): a library is a face-down pile
 * no one may look at (CR 401.2), and a failed search is the same cue.
 */
export function searchedLibrary(e: LogEvent): number | null {
  if (e.kind !== "search" && e.kind !== "reveal") return null;
  if (e.library_seat !== undefined) return e.library_seat;
  if (e.kind !== "search" || e.label === "shuffle") return null;
  return e.target_seat ?? e.seat;
}

/** How many card backs a library_search cue riffles off the pile (#4447). */
export const RIFFLE_CARDS = 4;

/**
 * searchCues gives each library searched in one batch one riffle
 * (#4447), on that library's pile. One batch is one moment ("each player
 * searches their library"; a search made twice before one shuffle,
 * CR 701.23h): the first cue plays the sound and the rest follow it.
 */
function searchCues(entries: readonly LogEvent[]): EffectCue[] {
  const seen = new Set<string>();
  const lead = new Set<string>();
  const out: EffectCue[] = [];
  for (const e of entries) {
    const lib = searchedLibrary(e);
    if (lib === null || lib < 0) continue;
    const batch = e.batch ?? null;
    const at = batch === null ? `q${e.seq}` : String(batch);
    const k = `${at}|${lib}`;
    if (seen.has(k)) continue;
    seen.add(k);
    const c: EffectCue = {
      id: `${e.seq}:search`,
      seq: e.seq,
      kind: "library_search",
      cardID: null,
      stackItemID: null,
      batch,
      weight: "full",
      delayMs: 0,
      seat: e.seat,
      librarySeat: lib,
    };
    if (lead.has(at)) c.follows = true;
    else lead.add(at);
    out.push(c);
  }
  return out;
}

// ---- A double-faced card transforms, or day and night (#4448) ----

/**
 * In a batch where the game becomes day or night, the permanents it turns
 * over (CR 702.145c, 702.145f) flip this long into the sky's wash, so the
 * table sees night fall and then the werewolves turn. The flip still ends
 * inside the Short cap: 250 + 600 ms.
 */
export const DAY_NIGHT_FLIP_DELAY_MS = 250;
/**
 * The share of a flip at which the card is edge-on and its new face starts
 * to show. Counters put on a permanent that flips in the same batch pulse
 * from here, on the face that has them.
 */
export const TRANSFORM_TURN = 0.5;

/**
 * dayNightOf is what a `day_night` entry says the game became (CR 731.1),
 * or null for anything else.
 */
export function dayNightOf(e: LogEvent): "day" | "night" | null {
  if (e.kind !== "day_night") return null;
  return e.label === "day" || e.label === "night" ? e.label : null;
}

/**
 * turnCues gives each permanent turned over in one batch one flip (#4448):
 * a `transform` entry (a transform, or a convert, CR 701.28a) or a `meld`
 * entry (the melded permanent, CR 701.42a), with its `card_id`; and each
 * batch in which the game became day or night one wash over the table,
 * what it became last. One batch is one moment and one sound: when night
 * falls the wash plays its sound and the werewolves it turns over follow
 * it, a beat later (DAY_NIGHT_FLIP_DELAY_MS); otherwise the batch's first
 * flip plays it (Moonmist turning every Human over). A melded planeswalker
 * (Urza, Planeswalker) is its arrival's cue, pw_enter, and no flip. An
 * entry with no card_id (a viewer who may not identify the card) is no
 * flip. A daybound card entering at night, a disturb cast, a face-down
 * permanent turned face up and "return it transformed" write no such
 * entry: none of them transforms (CR 701.27b, 712.14a).
 */
function turnCues(entries: readonly LogEvent[], ctx: PlanContext): EffectCue[] {
  const skies = new Map<string, EffectCue>();
  for (const e of entries) {
    const sky = dayNightOf(e);
    if (!sky) continue;
    const batch = e.batch ?? null;
    const at = batch === null ? `q${e.seq}` : String(batch);
    const was = skies.get(at);
    if (was) {
      was.sky = sky;
      continue;
    }
    skies.set(at, {
      id: `${e.seq}:sky`,
      seq: e.seq,
      kind: "day_night",
      cardID: null,
      stackItemID: null,
      batch,
      weight: "full",
      delayMs: 0,
      sky,
    });
  }
  const out: EffectCue[] = [...skies.values()];
  const seen = new Set<string>();
  const lead = new Set<string>(skies.keys());
  for (const e of entries) {
    if ((e.kind !== "transform" && e.kind !== "meld") || !e.card_id) continue;
    const melded = e.kind === "meld";
    if (melded && ctx.isPlaneswalker(e.card_id)) continue;
    const batch = e.batch ?? null;
    const at = batch === null ? `q${e.seq}` : String(batch);
    const k = `${at}|${e.card_id}`;
    if (seen.has(k)) continue;
    seen.add(k);
    const c: EffectCue = {
      id: `${e.seq}:turn`,
      seq: e.seq,
      kind: "transform",
      cardID: e.card_id,
      stackItemID: null,
      batch,
      weight: "full",
      delayMs: skies.has(at) ? DAY_NIGHT_FLIP_DELAY_MS : 0,
    };
    if (melded) c.melded = true;
    if (lead.has(at)) c.follows = true;
    else lead.add(at);
    out.push(c);
  }
  return out;
}

/**
 * crackIsOneSound: the permanents one payment sacrifices for their own
 * abilities (#4427) are one moment and one sound. Two Treasures the
 * auto-tapper spends on one spell (CR 601.2g) arrive in one frame and
 * one batch: the first crack plays its sound and the rest follow it.
 * The sound is capped on the payment, the first crack's entry, not on
 * the batch, because a player's separate activations in one priority
 * window share an event batch, and a Treasure cracked a minute after
 * another is a moment of its own.
 */
function crackIsOneSound(cues: EffectCue[]): void {
  const lead = new Map<string, EffectCue>();
  for (const c of cues) {
    if (c.kind !== "crack") continue;
    const at = c.batch === null ? `q${c.seq}` : String(c.batch);
    const first = lead.get(at);
    if (first) {
      c.follows = true;
      c.soundKey = first.soundKey;
      continue;
    }
    c.soundKey = `${at}:crack:${c.seq}`;
    lead.set(at, c);
  }
}

/**
 * waitForFlips holds the pulse of counters put on a permanent that flips
 * in the same batch (#4448) until the flip turns edge-on (TRANSFORM_TURN),
 * so it rings the face the counters are on, not a card mid-turn. Its own
 * sound waits with it.
 */
function waitForFlips(cues: EffectCue[]): void {
  const flips = new Map<string, number>();
  for (const c of cues) {
    if (c.kind !== "transform" || c.batch === null || !c.cardID) continue;
    flips.set(`${c.batch}|${c.cardID}`, c.delayMs);
  }
  if (flips.size === 0) return;
  for (const c of cues) {
    if (c.kind !== "counters" || c.batch === null || !c.cardID) continue;
    const at = flips.get(`${c.batch}|${c.cardID}`);
    if (at === undefined) continue;
    c.delayMs = Math.max(c.delayMs, at + TRANSFORM_TURN * EFFECT_BASE_MS.transform);
  }
}

/**
 * meldIsOneSound: a meld exiles its two cards and puts them back as one
 * (CR 701.42a), in one batch, and that is one moment. The two cards still
 * leave in the exile's white void, which is what happened to them, but
 * play no sound: the melded permanent's flip (or a melded planeswalker's
 * arrival) is the batch's sound.
 */
function meldIsOneSound(cues: EffectCue[], turns: readonly LogEvent[]): void {
  const melds = new Set<number>();
  for (const e of turns) if (e.kind === "meld" && e.card_id && e.batch != null) melds.add(e.batch);
  if (melds.size === 0) return;
  for (const c of cues) {
    if (c.kind === "exile" && c.batch !== null && melds.has(c.batch)) c.follows = true;
  }
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

/**
 * The key one logical sound is capped on: once per batch, per game. A cue
 * with a `soundKey` of its own (a crack's payment, #4427) is capped on
 * that instead.
 */
export function soundBatch(
  gameID: string,
  batch: number | null,
  soundKey?: string,
): string | undefined {
  if (soundKey !== undefined) return `${gameID}:${soundKey}`;
  return batch === null ? undefined : `${gameID}:${batch}`;
}

/**
 * The sound a cue plays, and its volume hint (a wipe's per-card sounds are
 * quieter). A cue that follows another of its batch has none of its own;
 * a counter cue's is its tone's.
 */
export function cueSound(
  cue: Pick<EffectCue, "kind" | "weight" | "follows" | "counterTone" | "reanimateTo" | "crackOf">,
): {
  name: SoundName;
  volume: number;
} | null {
  if (cue.follows) return null;
  // A crack's is its permanent's, a little quieter (#4427).
  if (cue.kind === "crack") {
    return { name: CRACK_SOUND[cue.crackOf ?? "other"], volume: CRACK_VOLUME };
  }
  const name =
    cue.kind === "counters" ? COUNTER_SOUND[cue.counterTone ?? "other"] : EFFECT_SOUND[cue.kind];
  if (!name) return null;
  const volume = cue.weight === "light" ? WIPE_CARD_VOLUME : 1;
  // A card returned to a hand is the quieter variant (#4446).
  if (cue.kind === "reanimate" && cue.reanimateTo === "hand") {
    return { name, volume: volume * REANIMATE_HAND_VOLUME };
  }
  return { name, volume };
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

/** One half of a cracked permanent (#4427). */
export interface CrackHalf {
  /** Its outline: a clip-path polygon in percent of the card's box (unturned). */
  clip: string;
  /**
   * Its centre from the card's, and where it falls to from there, in px
   * on screen (turned with the card: a tapped land splits top and bottom),
   * and its spin as it falls, in degrees.
   */
  ox: number;
  oy: number;
  dx: number;
  dy: number;
  rot: number;
}

export interface CrackPlan {
  /** The crack, from one end of the card to the other (SVG path data, px, unturned). */
  path: string;
  halves: [CrackHalf, CrackHalf];
}

/** How far through a crack the line has run the card's length and it snaps. */
export const CRACK_SNAP = 0.3;
/** Points on the crack line, end to end. */
export const CRACK_JOINTS = 7;

/**
 * crackPlan splits a permanent `w` x `h` (px, unturned) sacrificed for its
 * own ability (#4427): a jagged line runs down its length, like the grain
 * of a snapped branch, and the two halves either side of it fall away from
 * each other and down the screen, turning outward. `rot` is the card's turn
 * on screen (90 for a tapped land), so the halves part across the crack
 * and fall down whichever way the card lies. Seeded by the card, so it
 * always breaks the same way.
 */
export function crackPlan(seed: string, w: number, h: number, rot = 0): CrackPlan {
  const rnd = seeded(`crack:${seed}`);
  const W = Math.max(1, w);
  const H = Math.max(1, h);
  const line: Pt[] = [];
  for (let k = 0; k < CRACK_JOINTS; k++) {
    const t = k / (CRACK_JOINTS - 1);
    // Wider in the middle than at the ends, alternating sides.
    const swing = (k % 2 === 0 ? 1 : -1) * (0.04 + rnd() * 0.08) * Math.sin(Math.PI * t);
    line.push({ x: W * clamp(0.5 + swing + (rnd() - 0.5) * 0.04, 0.3, 0.7), y: H * t });
  }
  const pct = (p: Pt) => `${round1((p.x / W) * 100)}% ${round1((p.y / H) * 100)}%`;
  const left = [{ x: 0, y: 0 }, ...line, { x: 0, y: H }];
  const right = [{ x: W, y: 0 }, { x: W, y: H }, ...[...line].reverse()];
  const mid = line.reduce((t, p) => t + p.x, 0) / line.length;
  const r = (rot * Math.PI) / 180;
  const cos = Math.cos(r);
  const sin = Math.sin(r);
  // From the card's own frame to the screen's.
  const turn = (x: number, y: number) => ({ x: x * cos - y * sin, y: x * sin + y * cos });
  const half = (pts: Pt[], side: -1 | 1): CrackHalf => {
    const cx = (side < 0 ? mid / 2 : (mid + W) / 2) - W / 2;
    const o = turn(cx, 0);
    const out = turn(side * W * (0.22 + rnd() * 0.1), 0);
    const fall = H * (0.38 + rnd() * 0.14);
    return {
      clip: `polygon(${pts.map(pct).join(", ")})`,
      ox: round1(o.x),
      oy: round1(o.y),
      dx: round1(out.x),
      dy: round1(out.y + fall),
      rot: round1(side * (14 + rnd() * 12)),
    };
  };
  const path = `M${line.map((p) => `${round1(p.x)} ${round1(p.y)}`).join(" L")}`;
  return { path, halves: [half(left, -1), half(right, 1)] };
}

/** A speck a fizzling spell comes apart into (#4450). */
export interface FizzleMote {
  /** Where it starts, in percent of the stack item's box. */
  x: number;
  y: number;
  /** Its drift, px: a little sideways, mostly up. */
  dx: number;
  dy: number;
  size: number;
  /** When it lets go, as a share of the cue. */
  at: number;
}

export const FIZZLE_MOTES = 14;

/**
 * fizzleMotes plans a fizzle's dissolve over a stack item `w` x `h`
 * (#4450): specks scattered over the box that let go from the top down,
 * each drifting up and a little sideways as it fades, so the spell seems
 * to come apart into the air. Seeded by the entry.
 */
export function fizzleMotes(seed: string, w: number, h: number, n = FIZZLE_MOTES): FizzleMote[] {
  const rnd = seeded(`fizzle:${seed}`);
  const H = Math.max(1, h);
  const W = Math.max(1, w);
  const out: FizzleMote[] = [];
  for (let i = 0; i < n; i++) {
    const y = 8 + ((i + rnd()) / n) * 84;
    out.push({
      x: round1(8 + rnd() * 84),
      y: round1(y),
      dx: round1((rnd() - 0.5) * W * 0.3),
      dy: round1(-H * (0.18 + rnd() * 0.22)),
      size: round1(2 + rnd() * 3),
      // The top lets go first: the dissolve runs down the card.
      at: round2(0.08 + (y / 100) * 0.4 + rnd() * 0.08),
    });
  }
  return out;
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
 * speed). A full cue keeps its own stagger (a token batch's pop-ins,
 * which fit the Short cap whole). A still cue starts at once and is the
 * badge's hold.
 */
export function cueTiming(
  cue: Pick<EffectCue, "kind" | "weight" | "delayMs" | "count">,
  motion: EffectMotion,
  speed: number,
): { delayMs: number; ms: number } {
  if (motion === "still") return { delayMs: 0, ms: EFFECT_STILL_HOLD_MS };
  const s = speedFactor(speed);
  // A mill, discard or reanimate's length is its run of cards' (#4445, #4446).
  const base = isPileKind(cue.kind)
    ? pileBaseMs(cue.kind, cue.count ?? 1)
    : EFFECT_BASE_MS[cue.kind];
  if (cue.weight === "light") {
    const k = wipeScale(motion);
    return {
      delayMs: cue.delayMs * k * s,
      ms: Math.max(1, Math.min(base, WIPE_CARD_MS) * k * s),
    };
  }
  return { delayMs: cue.delayMs * s, ms: cueMs(base, motion, speed) };
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
 * are not counted: the wipe already fits them inside its cap. Nor is a
 * cue that follows another of its batch: tokens created together count
 * once, and so do the counters one proliferate puts on, two commanders
 * sent home together, the Equipment one effect attaches, a token
 * created attached, whose token cue follows its attach cue, the
 * players one effect makes mill or discard, and the cards one effect
 * returns from graveyards (and a returned planeswalker's or Aura's
 * arrival or attach cue, which follows its rise), the libraries one
 * effect has players search, and the permanents one batch turns over
 * (and, when night falls, the werewolves that follow its wash), and the
 * permanents one payment sacrifices for their own abilities (#4427).
 */
export function headlines(plan: EffectPlan, signatures = 0): number {
  return (
    plan.cues.filter((c) => c.weight === "full" && !c.follows).length +
    plan.wipes.length +
    signatures
  );
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

// ---- The decked-out cue (#4879) ----
//
// A player who loses by drawing from an empty library (CR 704.5b) gets a
// signature-length cue of their own, "the library dies", in place of the
// avatar's rock-shatter (EliminationFx.svelte, #4449): their library slot
// pulses red, cold blue wisps rise from it, and a translucent blue skull
// rises out of the pile, swells and drifts until it fills that player's
// mat, sways once and dissolves upward. As it fills the mat the cards in
// their hand are flung outward, spinning, and vanish. The seat then drains
// to a cold grey-blue and its stone reads "Decked" (Tombstone.svelte, from
// the seat's own `loss_cause`, so a reload still shows it).
//
// The wire's word for the cause is the engine's game.LossCause
// "empty_draw", on the `eliminated` log line's `cause` and on the seat's
// `loss_cause` (docs/protocol.md). Every other cause keeps the shatter.

/** The loss cause the decked-out cue is for: a draw from an empty library (CR 704.5b). */
export const DECKED_CAUSE = "empty_draw";

/** Which cue a player's departure plays, by its cause. */
export type EliminationCueKind = "decked_out" | "eliminated";

/** The decked-out cue for a draw from an empty library; the shatter for every other cause. */
export function eliminationCue(cause: string | null | undefined): EliminationCueKind {
  return cause === DECKED_CAUSE ? "decked_out" : "eliminated";
}

/** The sound a departure plays: player_decked replaces player_eliminated for a decked player. */
export function eliminationSound(cause: string | null | undefined): SoundName {
  return eliminationCue(cause) === "decked_out" ? "player_decked" : "player_eliminated";
}

/** Whether a seat that has left was decked, from its own `loss_cause`. */
export function seatDecked(seat: { eliminated?: boolean; loss_cause?: string }): boolean {
  return seat.eliminated === true && eliminationCue(seat.loss_cause) === "decked_out";
}

/**
 * The seats the log entries say were decked out, each once, in log order:
 * an `eliminated` entry whose cause is a draw from an empty library.
 */
export function deckedSeats(entries: readonly LogEvent[]): number[] {
  const seats: number[] = [];
  for (const e of entries) {
    if (e.kind !== "eliminated" || eliminationCue(e.cause) !== "decked_out") continue;
    if (typeof e.seat !== "number" || e.seat < 0) continue;
    if (!seats.includes(e.seat)) seats.push(e.seat);
  }
  return seats;
}

/** The cue's length at speed 1, at Full (the owner's spec: about 2.2 s). */
export const DECKED_BASE_MS = 2200;
/**
 * A signature's cap (fx/signatures.ts SIGNATURE_CAP_MS, ADR 0209 §4's
 * amendment). Repeated here, not imported, because signatures.ts imports
 * this file; a test holds the two equal.
 */
export const DECKED_CAP_MS = 3500;

/**
 * The beats of the cue, as shares of its length. The pulse and the wisps
 * open it; the skull rises out of the pile and fills the mat at `fill`,
 * which is when the hand is flung; it sways until `sway`, then dissolves
 * upward to the end. The stone settles at `badge`.
 */
export const DECKED_BEATS = {
  pulse: [0, 0.2],
  wisps: [0.02, 0.7],
  rise: [0.08, 0.5],
  fill: 0.5,
  sway: 0.74,
  fling: [0.5, 0.92],
  badge: 0.8,
} as const;

/** Most hand cards flung; a bigger hand flings its first few (the rest are under them). */
export const DECKED_FLING_CAP = 10;
/** How many wisps rise from the library. */
export const DECKED_WISPS = 9;

/**
 * deckedTiming is the cue's length in real ms, or null when it is still
 * (Still mode, reduced motion: the grey seat and its badge only). Full is
 * DECKED_BASE_MS, capped at the signature cap; Short halves it, within
 * Short's cap. Then animations.speed, and a busy table's compression
 * (busyScale), multiplied into `speed` by the caller as every cue's is.
 */
export function deckedTiming(motion: EffectMotion, speed: number): { ms: number } | null {
  if (motion === "still") return null;
  const base = Math.min(DECKED_BASE_MS, DECKED_CAP_MS);
  const ms = motion === "short" ? Math.min(base / 2, EFFECT_SHORT_CAP_MS) : base;
  return { ms: Math.max(1, ms * speedFactor(speed)) };
}

/**
 * When a decked seat's stone settles, in ms after the frame: at the
 * cue's `badge` beat, or at once when the cue is still.
 */
export function deckedBadgeDelay(motion: EffectMotion, speed: number): number {
  const t = deckedTiming(motion, speed);
  return t ? Math.round(t.ms * DECKED_BEATS.badge) : 0;
}

/** One wisp: where it starts across the pile (0..1), how far it drifts and rises (in pile heights), its size and when it starts (a share of the wisp beat). */
export interface DeckedWisp {
  x: number;
  sway: number;
  rise: number;
  size: number;
  at: number;
}

/** The wisps, seeded by the seat, so the same player's library dies the same way. */
export function deckedWisps(seed: string, n = DECKED_WISPS): DeckedWisp[] {
  const rand = seeded(`wisp:${seed}`);
  const out: DeckedWisp[] = [];
  for (let i = 0; i < n; i++) {
    out.push({
      x: round3(0.15 + rand() * 0.7),
      sway: round3((rand() - 0.5) * 0.9),
      rise: round3(1.4 + rand() * 1.6),
      size: round3(0.22 + rand() * 0.22),
      at: round3((i / n) * 0.55),
    });
  }
  return out;
}

/**
 * One flung hand card: where it starts across the hand (0..1), the way it
 * flies (a unit direction; it also falls), how far in board diagonals,
 * and how much it spins, in degrees.
 */
export interface DeckedFling {
  x: number;
  dx: number;
  dy: number;
  reach: number;
  spin: number;
}

/**
 * The hand's cards, flung outward in every direction (#4879): `n` of them,
 * capped at DECKED_FLING_CAP, seeded by the seat. Each goes far enough to
 * leave the board, spinning, and falls a little as it goes.
 */
export function deckedFlings(seed: string, n: number): DeckedFling[] {
  const count = Math.max(0, Math.min(DECKED_FLING_CAP, Math.floor(n)));
  const rand = seeded(`fling:${seed}`);
  const out: DeckedFling[] = [];
  for (let i = 0; i < count; i++) {
    // Spread round the circle, jittered, so no two go the same way.
    const angle = ((i + rand() * 0.8) / Math.max(1, count)) * Math.PI * 2;
    out.push({
      x: count === 1 ? 0.5 : round3(i / (count - 1)),
      dx: round3(Math.cos(angle)),
      dy: round3(Math.sin(angle)),
      reach: round3(0.75 + rand() * 0.45),
      spin: Math.round((rand() < 0.5 ? -1 : 1) * (360 + rand() * 540)),
    });
  }
  return out;
}

function round3(n: number): number {
  return Math.round(n * 1000) / 1000;
}
