// seatSummary derives the dense read-out an opponent's panel shows
// when `settings.display.opponentDetail` is "summary" — life, mana
// that could answer you, creature bodies, and the handful of
// permanents whose presence changes how you play the turn.
//
// WHY A SECOND REPRESENTATION. Card size is a share of the panel's
// height (`--card-h` on .panel, a size container) with a clamp() floor
// at 90px for a flipped opponent. Past the floor a panel stops
// shrinking its cards and starts clipping them, which is #956. Scaling
// has no headroom left, so a small panel gets a DIFFERENT rendering
// rather than a smaller one — and a rendering with no floor to hit
// reads the same in a quadrant as in a third of the top row.
//
// EVERYTHING HERE IS DETERMINISTIC. Nothing ranks, scores or guesses
// which permanents matter. A heuristic that picks "the important
// cards" is wrong at the worst possible moment and is blamed for the
// loss, fairly, because it chose what to hide. The structural rule
// below (planeswalkers, commanders, anything wearing an attachment) is
// explainable and wrong in ways the player can predict.
//
// UNDER-REPORT RATHER THAN OVER-REPORT. A confidently wrong mana read
// is worse than no mana read, because the player acts on it. Anything
// this module cannot pin down lands in `flexible` and is rendered as
// "could be anything", never folded into a colour.

import type { CardView, PlayerView } from "./protocol";
import { isCreature, isPlaneswalker } from "./cardTypes";
import { abilityBlocked } from "./contextMenu.logic";

// The five colours plus colourless, in WUBRG order. Display order only
// — the wire never promises an order and manaPick deliberately does
// not sort, but a read-out the player scans every turn has to sit in
// the same place every time.
export const MANA_ORDER: readonly string[] = ["W", "U", "B", "R", "G", "C"];

export interface ManaAvailability {
  /** Mana this seat could add right now, by colour letter. */
  byColor: Record<string, number>;
  /**
   * Mana from sources whose colour is not determined here — an "any
   * colour" source, a hybrid symbol, a card whose `produced` string is
   * absent. Counted, never attributed: see the under-report note above.
   */
  flexible: number;
  /** Untapped permanents with at least one usable mana ability. */
  sources: number;
}

export interface CreatureStats {
  total: number;
  untapped: number;
  tapped: number;
  /** Summed power over every creature, tapped included. */
  power: number;
  /**
   * Summed power over untapped creatures only — the number that
   * answers "what can block me", which is the one you read on someone
   * else's turn.
   */
  untappedPower: number;
}

export type StructuralKind = "planeswalker" | "commander" | "attached";

export interface StructuralPermanent {
  card: CardView;
  kind: StructuralKind;
  /** Loyalty for a planeswalker, defense for a battle, else undefined. */
  counter?: number;
}

export interface SeatSummaryView {
  seat: PlayerView;
  eliminated: boolean;
  life: number;
  /** Commander damage taken, by commander instance id; zeroes dropped. */
  commanderDamage: Record<string, number>;
  handCount: number;
  libraryCount: number;
  graveyardCount: number;
  mana: ManaAvailability;
  creatures: CreatureStats;
  /** Every creature this seat controls, for the pip row. */
  creatureCards: CardView[];
  structural: StructuralPermanent[];
}

// A mana symbol as it appears inside `produced`: "{G}", "{C}{C}",
// "{W/U}", "{2}". Scryfall's grammar, which is what the server ships.
const SYMBOL = /\{([^}]+)\}/g;

/**
 * usableManaAbilities returns the mana abilities on a card that could
 * be activated right now.
 *
 * Reuses `abilityBlocked` rather than re-deriving the cost checks, so
 * the summary and the card menu can never disagree about whether a
 * source is live. The three-argument form is the one the menu itself
 * uses for a mana row; the optional fourth argument is for the
 * sorcery-speed window, which no mana ability has (CR 605.1a).
 *
 * `cant_activate_mana` is a card-level restriction rather than an
 * ability-level one, so it is checked here.
 */
function usableManaAbilities(c: CardView): NonNullable<CardView["mana_abilities"]> {
  if (c.restrictions?.includes("cant_activate_mana")) return [];
  if (c.restrictions?.includes("cant_activate")) return [];
  const tapped = c.tapped === true;
  const sick = c.summoning_sick === true;
  return (c.mana_abilities ?? []).filter((a) => abilityBlocked(a, tapped, sick) === "");
}

/**
 * manaFromProduced adds one ability's output into an accumulator.
 *
 * A symbol resolves to a colour only when it is exactly one of the six
 * letters. A hybrid ("{W/U}"), a Phyrexian symbol, a generic amount
 * ("{2}") or anything unrecognised is flexible: it is real mana, so it
 * counts, but this module will not say what colour it is.
 *
 * An absent `produced` is flexible for the same reason — the ability
 * adds mana (the server would not have shipped it otherwise) and the
 * label is prose we are not going to parse.
 */
function manaFromProduced(produced: string | undefined, into: ManaAvailability): void {
  if (!produced) {
    into.flexible += 1;
    return;
  }
  let matched = false;
  for (const m of produced.matchAll(SYMBOL)) {
    matched = true;
    const sym = m[1].toUpperCase();
    if (MANA_ORDER.includes(sym)) {
      into.byColor[sym] = (into.byColor[sym] ?? 0) + 1;
      continue;
    }
    // "{2}" on a mana ability means two generic, not two of something
    // known — Cabal Coffers pays out an amount, not a colour.
    const n = Number.parseInt(sym, 10);
    into.flexible += Number.isFinite(n) && n > 0 ? n : 1;
  }
  // A `produced` string with no symbols in it at all ("any color") is
  // one flexible mana rather than none.
  if (!matched) into.flexible += 1;
}

/**
 * manaAvailable reads what this seat could add to their pool right
 * now, across every permanent they control.
 *
 * This is the highest-value line in the summary: it is the "can they
 * respond?" read, and today a player gets it by counting untapped
 * lands by eye and guessing at their colours.
 *
 * It is a FLOOR, not a promise. Mana already floating (`mana_pool`),
 * rituals in hand, and anything requiring a cost this module does not
 * model are all excluded, and any ability whose colour is uncertain
 * lands in `flexible`. A player who reads this as "they have exactly
 * this much" is reading it wrong; a player who reads it as "they have
 * at least this much" is reading it right.
 */
export function manaAvailable(controlledCards: readonly CardView[]): ManaAvailability {
  const out: ManaAvailability = { byColor: {}, flexible: 0, sources: 0 };
  for (const c of controlledCards) {
    const abilities = usableManaAbilities(c);
    if (abilities.length === 0) continue;
    out.sources += 1;
    // One source, one activation. A land with three mana abilities can
    // only be tapped once, so the summary counts its BEST-KNOWN entry
    // rather than summing every printed option — summing would report
    // a Triome as three mana.
    //
    // "Best known" = the first entry whose colour this module can
    // actually name, so a dual land reports a colour rather than
    // falling into `flexible` because its second ability happened to
    // be listed first.
    const named = abilities.find((a) => hasNamedSymbol(a.produced));
    manaFromProduced((named ?? abilities[0]).produced, out);
  }
  return out;
}

function hasNamedSymbol(produced: string | undefined): boolean {
  if (!produced) return false;
  for (const m of produced.matchAll(SYMBOL)) {
    if (MANA_ORDER.includes(m[1].toUpperCase())) return true;
  }
  return false;
}

/**
 * manaLabel is the spoken form of the mana line, for the aria-label
 * on a summary panel.
 *
 * "from N sources" rather than a bare total, because the total alone
 * invites reading it as exact. A screen-reader user should get the
 * same caveat a sighted one reads off the dashed "?" pip.
 */
export function manaLabel(m: ManaAvailability): string {
  const known = Object.values(m.byColor).reduce((a, b) => a + b, 0);
  const total = known + m.flexible;
  if (total === 0) return "no mana";
  return `${total} mana from ${m.sources} ${m.sources === 1 ? "source" : "sources"}`;
}

/**
 * creatureStats counts bodies and power.
 *
 * `power` uses CardView.power, which is the effective value with the
 * combat-damage zero clamp already applied — not `negative_power`,
 * which exists for comparisons like skulk. A creature with no numeric
 * power (a card with non-numeric printed stats omits the field)
 * contributes zero rather than being dropped from the count: the body
 * is on the battlefield and can still block.
 */
export function creatureStats(controlledCards: readonly CardView[]): CreatureStats {
  const out: CreatureStats = { total: 0, untapped: 0, tapped: 0, power: 0, untappedPower: 0 };
  for (const c of controlledCards) {
    if (!isCreature(c)) continue;
    const p = c.power ?? 0;
    out.total += 1;
    out.power += p;
    if (c.tapped === true) {
      out.tapped += 1;
    } else {
      out.untapped += 1;
      out.untappedPower += p;
    }
  }
  return out;
}

/**
 * structuralPermanents picks the non-creature permanents that get a
 * tile of their own, by structure rather than by judgement:
 *
 *   - planeswalkers, because loyalty is a second life total
 *   - the commander, because the game is named after it
 *   - anything wearing an attachment, because an Aura or Equipment
 *     that is drawn nowhere reads as a permanent that isn't there
 *
 * Creatures are excluded — they are already in the pip row — EXCEPT a
 * commander or an enchanted creature, which earn a tile for the reason
 * above and are rendered alongside rather than instead of their pip.
 *
 * `hostIDs` is the set of instance ids that have something attached to
 * them, derived once by the caller from the whole battlefield (an Aura
 * you control can sit on a creature an opponent controls, so it cannot
 * be derived from one seat's slice).
 */
export function structuralPermanents(
  controlledCards: readonly CardView[],
  hostIDs: ReadonlySet<string>,
): StructuralPermanent[] {
  const out: StructuralPermanent[] = [];
  for (const c of controlledCards) {
    if (isPlaneswalker(c)) {
      out.push({ card: c, kind: "planeswalker", counter: c.counters?.loyalty });
      continue;
    }
    if (c.is_commander === true) {
      out.push({ card: c, kind: "commander" });
      continue;
    }
    if (hostIDs.has(c.instance_id)) {
      out.push({ card: c, kind: "attached", counter: c.defense });
    }
  }
  return out;
}

/**
 * attachmentHostIDs is the set of permanents that have at least one
 * Equipment or Aura on them.
 *
 * Takes the WHOLE battlefield, for the cross-controller reason in
 * structuralPermanents. A dangling attachment — the host has left but
 * the state-based action has not swept the relation yet — names an id
 * that is no longer on the battlefield and is simply absent from the
 * set, which is the same thing PlayerPanel does.
 */
export function attachmentHostIDs(battlefield: readonly CardView[]): Set<string> {
  const present = new Set(battlefield.map((c) => c.instance_id));
  const out = new Set<string>();
  for (const c of battlefield) {
    const host = c.attached_to;
    if (host?.kind === "card" && host.id && present.has(host.id)) out.add(host.id);
  }
  return out;
}

/**
 * buildSeatSummary assembles everything one opponent panel renders.
 *
 * An eliminated seat (CR 104.3, or a concession) keeps its identity
 * and its life total and reports nothing else: its permanents have
 * left the battlefield, and a read-out full of zeroes reads as a bug
 * rather than as a player who is out.
 */
export function buildSeatSummary(
  seat: PlayerView,
  controlledCards: readonly CardView[],
  battlefield: readonly CardView[],
): SeatSummaryView {
  const eliminated = seat.eliminated === true;
  const commanderDamage: Record<string, number> = {};
  for (const [id, n] of Object.entries(seat.commander_damage ?? {})) {
    if (n > 0) commanderDamage[id] = n;
  }

  if (eliminated) {
    return {
      seat,
      eliminated,
      life: seat.life,
      commanderDamage,
      handCount: 0,
      libraryCount: seat.library?.count ?? 0,
      graveyardCount: seat.graveyard?.count ?? 0,
      mana: { byColor: {}, flexible: 0, sources: 0 },
      creatures: { total: 0, untapped: 0, tapped: 0, power: 0, untappedPower: 0 },
      creatureCards: [],
      structural: [],
    };
  }

  return {
    seat,
    eliminated,
    life: seat.life,
    commanderDamage,
    handCount: seat.hand?.count ?? 0,
    libraryCount: seat.library?.count ?? 0,
    graveyardCount: seat.graveyard?.count ?? 0,
    mana: manaAvailable(controlledCards),
    creatures: creatureStats(controlledCards),
    creatureCards: controlledCards.filter(isCreature),
    structural: structuralPermanents(controlledCards, attachmentHostIDs(battlefield)),
  };
}
