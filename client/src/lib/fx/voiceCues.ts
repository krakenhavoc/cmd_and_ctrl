// Which planeswalker moments a game-log frame holds (ADR 0209 §5, #4398).
//
//   enter     a planeswalker card moved onto the battlefield (a `zone` entry
//             whose new_zone is "battlefield").
//   leave     a planeswalker card moved off it.
//   ultimate  an `activate` entry for a loyalty ability that is the walker's
//             most expensive minus ability.
//
// The log never says which ability was a loyalty ability, or its cost: an
// `activate` entry carries the source card and the ability's `label`. So
// "ultimate" is decided here by looking the label up among the walker's
// activated_abilities (which carry loyalty_cost): it is the ultimate when it
// matches exactly one loyalty ability, that ability's cost is below zero, and
// it is strictly more negative than every other fixed-cost loyalty ability of
// the walker. A -X ability is not counted as the ultimate, and a walker with a
// single minus ability, or a tie, has none. When the label does not match
// (an older server, a granted ability) nothing plays: silent beats wrong.
//
// The walker is often gone from the frame that records its ultimate (its
// loyalty hit zero), so the tracker remembers each walker's loyalty abilities
// from the frames it saw it on the battlefield.

import type { ActivatedAbilityView, CardView, GameView, LogEvent } from "../protocol";
import type { VoiceMoment } from "./voices";

export interface VoiceCue {
  /** The walker's printed name, for the catalog lookup. */
  name: string;
  moment: VoiceMoment;
}

interface Remembered {
  name: string;
  abilities: ActivatedAbilityView[];
}

export function isPlaneswalker(card: CardView | null | undefined): boolean {
  return !!card && !card.is_token && /\bplaneswalker\b/i.test(card.type_line ?? "");
}

function allCards(view: GameView): CardView[] {
  const out: CardView[] = [...(view.battlefield?.cards ?? []), ...(view.exile?.cards ?? [])];
  for (const s of view.seats) {
    out.push(...(s.graveyard?.cards ?? []), ...(s.hand?.cards ?? []), ...(s.command?.cards ?? []));
  }
  return out;
}

function loyaltyAbilities(card: CardView): ActivatedAbilityView[] {
  return (card.activated_abilities ?? []).filter((a) => a.loyalty_cost !== undefined);
}

/**
 * isUltimate says whether the ability labelled `label` is the strictly largest
 * minus cost among `abilities`' fixed-cost loyalty abilities.
 */
export function isUltimate(
  abilities: readonly ActivatedAbilityView[],
  label: string | undefined,
): boolean {
  if (!label) return false;
  const loyalty = abilities.filter((a) => a.loyalty_cost !== undefined);
  const matches = loyalty.filter((a) => a.label === label);
  if (matches.length !== 1) return false;
  const m = matches[0]!;
  const cost = m.loyalty_cost ?? 0;
  if (m.loyalty_cost_x || cost >= 0) return false;
  return loyalty.every((a) => a === m || a.loyalty_cost_x || (a.loyalty_cost ?? 0) > cost);
}

function cueFor(
  e: LogEvent,
  card: CardView | undefined,
  remembered: Remembered | undefined,
): VoiceCue | null {
  const walker = card && isPlaneswalker(card) ? card : undefined;
  if (e.kind === "zone") {
    const entering = e.new_zone === "battlefield" && e.old_zone !== "battlefield";
    const leaving = e.old_zone === "battlefield" && e.new_zone !== "battlefield";
    if (!entering && !leaving) return null;
    if (walker) return { name: walker.name, moment: entering ? "enter" : "leave" };
    // A walker that left may be hidden from this viewer now (library, hand).
    if (leaving && remembered) return { name: remembered.name, moment: "leave" };
    return null;
  }
  if (e.kind === "activate") {
    // Prefer the remembered list when the walker's own is empty (it left
    // the battlefield and its view carries no abilities).
    const own = walker ? loyaltyAbilities(walker) : [];
    const list = own.length > 0 ? own : (remembered?.abilities ?? []);
    const name = walker?.name ?? remembered?.name;
    if (name && isUltimate(list, e.label)) return { name, moment: "ultimate" };
  }
  return null;
}

export function createVoiceCueTracker() {
  let lastSeq: number | null = null;
  // instance_id -> the walker's name and loyalty abilities, from the last
  // frame it stood on the battlefield.
  const seen = new Map<string, Remembered>();

  return {
    /**
     * observe takes one frame and returns the cues for the log entries new
     * since the last. The first frame is the baseline (reloading into a game
     * is silent) and a frame that rewinds (an undo) drops the watermark and
     * cues nothing.
     */
    observe(view: GameView | null | undefined): VoiceCue[] {
      if (!view) return [];
      const log = view.log ?? [];
      const top = log.reduce((m, e) => Math.max(m, e.seq), 0);
      const cues: VoiceCue[] = [];
      if (lastSeq !== null && top > lastSeq) {
        const cards = allCards(view);
        for (const e of log) {
          if (e.seq <= lastSeq || !e.card_id) continue;
          const id = e.card_id;
          const cue = cueFor(
            e,
            cards.find((c) => c.instance_id === id),
            seen.get(id),
          );
          if (cue) cues.push(cue);
        }
      }
      lastSeq = top;
      for (const c of view.battlefield?.cards ?? []) {
        if (isPlaneswalker(c)) {
          seen.set(c.instance_id, { name: c.name, abilities: loyaltyAbilities(c) });
        }
      }
      return cues;
    },
    reset(): void {
      lastSeq = null;
      seen.clear();
    },
  };
}
