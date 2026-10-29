// faceHit — the sound of combat damage landing on a player (#4432,
// #4543, ADR 0209).
//
// A creature that hits a player's face should sound like something
// smashed into them, not like a drain, and a bigger blow should sound
// bigger. The sound is one of three tiers by how much combat damage the
// player takes (faceHitSound): light for 1 (a deflection), medium for 2
// to 3 (a blunt punch) and heavy for 4 or more (a crush). Two surfaces
// could play a sound for the same hit, so the choice is made here, once,
// and both read it:
//
//   - the strike layer (CombatStrikes.svelte) plays the tier at the
//     beat's first contact (`tl.contactMs`), once per player the beat
//     hit, so the sound lands with the lunge. The tier is picked by all
//     the combat damage that player takes in that beat (two attackers of
//     2 are one blow of 4);
//   - the life popup (PlayerIdentity.svelte) plays `damage` for any
//     loss, so it stands down for a combat hit and lets the tier own
//     it. When no strike is scheduled (motion off, reduced motion, a
//     hidden tab) nothing lunges, so the popup plays the tier itself
//     the moment the loss shows, the same split combat_resolve has
//     between the step change and the strike layer.
//
// combat_resolve is a long war cry in the same category, so on a beat that
// hits a player it follows the hit(s) by WAR_CRY_AFTER_HIT_MS at a lower
// volume on both surfaces (warCryPlan, #4560); a creature-only beat keeps it
// at contact.
//
// Commander damage is combat damage, so it is covered here and gets a
// tier like any other blow. Non-combat life loss (a drain, life paid,
// noncombat damage) is not a combat hit and keeps `damage` from the popup.
//
// Plain functions over plain data, unit-tested without a component.

import type { GameView, LogEvent } from "./protocol";
import type { SoundName } from "./sounds";

// FACE_HIT_STAGGER_MS separates the blows when one beat hits several
// players. sounds.ts drops a repeat of a name inside its 40 ms rate limit,
// so simultaneous plays of one tier would collapse into one.
export const FACE_HIT_STAGGER_MS = 90;

// WAR_CRY_AFTER_HIT_MS (#4560): combat_resolve is a 5-7 s war cry in the
// same category as the face hits, so played together it buries them. When a
// beat hits a player the war cry starts this long after the LAST face hit
// (warCryPlan), at WAR_CRY_AFTER_HIT_VOLUME of its normal level.
export const WAR_CRY_AFTER_HIT_MS = 300;
export const WAR_CRY_AFTER_HIT_VOLUME = 0.6;

/**
 * When combat_resolve plays, as ms after first contact, and at what caller
 * volume hint. A beat that hits no player plays it at contact, at full
 * volume. One that hits `hitCount` players plays it WAR_CRY_AFTER_HIT_MS
 * after the last of the (FACE_HIT_STAGGER_MS apart) hits, quieter.
 */
export function warCryPlan(hitCount: number): { delayMs: number; volume: number } {
  if (hitCount <= 0) return { delayMs: 0, volume: 1 };
  return {
    delayMs: (hitCount - 1) * FACE_HIT_STAGGER_MS + WAR_CRY_AFTER_HIT_MS,
    volume: WAR_CRY_AFTER_HIT_VOLUME,
  };
}

/**
 * The players the newest step of `log` hit with combat damage: the entries
 * after the last `step` entry. For the step-change fallback, which has no
 * cue to read.
 */
export function stepFaceHits(log: readonly LogEvent[] | undefined): FaceHit[] {
  const all = log ?? [];
  let from = 0;
  for (let i = all.length - 1; i >= 0; i--) {
    if (all[i].kind === "step") {
      from = i + 1;
      break;
    }
  }
  return faceHits(all.slice(from));
}

/**
 * The step-change fallback's combat_resolve (motion off, reduced motion,
 * a hidden tab): plays it now when the step hit no player, else after the
 * hit(s) the life popup plays as the loss shows, quieter. Returns the
 * timer of a delayed cry, for the caller to clear on teardown.
 */
export function playStepWarCry(
  log: readonly LogEvent[] | undefined,
  play: (name: "combat_resolve", opts?: { volume: number }) => void,
  onFired?: (t: ReturnType<typeof setTimeout>) => void,
): ReturnType<typeof setTimeout> | null {
  const cry = warCryPlan(stepFaceHits(log).length);
  if (cry.delayMs === 0) {
    play("combat_resolve");
    return null;
  }
  const t = setTimeout(() => {
    onFired?.(t);
    play("combat_resolve", { volume: cry.volume });
  }, cry.delayMs);
  return t;
}

/** The three sounds a combat hit on a player can make. */
export type FaceHitSound = Extract<
  SoundName,
  "face_hit_light" | "face_hit_medium" | "face_hit_heavy"
>;

/**
 * The tier of a blow of `amount` combat damage: 1 is light, 2 to 3
 * medium, 4 or more heavy. Null for no damage (nothing hit), so a fully
 * prevented blow is silent.
 */
export function faceHitSound(amount: number): FaceHitSound | null {
  if (!Number.isFinite(amount) || amount <= 0) return null;
  if (amount >= 4) return "face_hit_heavy";
  if (amount >= 2) return "face_hit_medium";
  return "face_hit_light";
}

/** One player a beat hit, and the combat damage they took in all. */
export interface FaceHit {
  seat: number;
  amount: number;
}

/**
 * The players a set of log entries hit with combat damage, each once, in
 * log order, with the damage they took summed. A `damage` entry names a
 * player in `target_seat` (a card it hit is `target`). Zero damage (fully
 * prevented) is not a hit.
 */
export function faceHits(entries: readonly LogEvent[]): FaceHit[] {
  const hits: FaceHit[] = [];
  for (const e of entries) {
    if (e.kind !== "damage" || !e.combat) continue;
    if (e.target_seat === undefined || e.target_seat === null || e.target_seat < 0) continue;
    const amount = e.amount ?? 0;
    if (amount <= 0) continue;
    const seen = hits.find((h) => h.seat === e.target_seat);
    if (seen) seen.amount += amount;
    else hits.push({ seat: e.target_seat, amount });
  }
  return hits;
}

/** The seats a set of log entries hit with combat damage, each once, in log order. */
export function faceHitSeats(entries: readonly LogEvent[]): number[] {
  return faceHits(entries).map((h) => h.seat);
}

/**
 * The combat damage a popup's life changes hold for `seatIndex`, summed;
 * 0 when there is none. `deltas` is the popup's, oldest first; they are
 * matched against the seat's newest logged changes one for one, as
 * lifeCause.ts does, and 0 when they do not line up (then the popup keeps
 * `damage`).
 */
export function popupCombatDamage(
  view: GameView,
  seatIndex: number,
  deltas: readonly number[],
): number {
  if (seatIndex < 0 || deltas.length === 0) return 0;
  const mine: { delta: number; combat: boolean }[] = [];
  for (const e of view.log ?? []) {
    const amount = e.amount ?? 0;
    if (e.kind === "life" && e.seat === seatIndex && amount !== 0) {
      mine.push({ delta: amount, combat: false });
    } else if (e.kind === "damage" && e.target_seat === seatIndex && amount > 0) {
      mine.push({ delta: -amount, combat: e.combat === true });
    }
  }
  if (mine.length < deltas.length) return 0;
  const newest = mine.slice(mine.length - deltas.length);
  if (newest.some((c, k) => c.delta !== deltas[k])) return 0;
  return newest.reduce((sum, c) => (c.combat ? sum - c.delta : sum), 0);
}

/**
 * The sound a life popup plays, or null for none. `base` is lifePopup's
 * own choice ("damage" or "heal"), `combatDamage` the combat damage to the
 * seat that the popup holds (0 for none), and `strikesPlay` whether the
 * strike layer is scheduled to play the tier at contact (strikesScheduled).
 * A combat hit is either left to the strike layer (silent here) or, when
 * no strike plays, heard now as its tier; every other loss keeps `damage`.
 */
export function popupSound(
  base: "damage" | "heal",
  combatDamage: number,
  strikesPlay: boolean,
): "damage" | "heal" | FaceHitSound | null {
  const tier = faceHitSound(combatDamage);
  if (base === "heal" || !tier) return base;
  return strikesPlay ? null : tier;
}
