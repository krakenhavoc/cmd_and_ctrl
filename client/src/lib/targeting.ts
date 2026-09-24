import { type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type {
  ActivatedAbilityView,
  AlternativeCostView,
  CardView,
  LegalTargetsView,
  ModeOptionView,
  OptionalCostView,
  PendingChoiceView,
  TapCostView,
} from "./protocol";
import type { CounterPayment } from "./counterCost";

// targeting.ts is the shared-store plumbing for the S14 "cast a
// catalog card, pick a target" flow. When a player clicks a hand
// card whose CardView declares a non-empty `target_mode`, the
// caller writes a TargetingState here; receiving surfaces (player
// headers, battlefield cards, stack items) watch the store and
// re-route their click handlers to resolve the target prompt.
// Store clears on Escape key or on cast completion.
//
// This is the minimum-viable two-click pick (hand card → click a
// target → cast fires). A richer picker (hover-highlight legal
// targets, validate predicate server-side) is S20 smart-cast
// territory.

// TargetingMode mirrors effects.Spec.TargetMode on the server.
// Values:
//   "any"         — player, creature, planeswalker, or battle
//   "player"      — seated player only
//   "creature"    — battlefield creature only
//   "stack_spell" — a spell currently on the stack
//   "stack_ability" — an activated or triggered ability on the stack
//                     (Stifle, Strionic Resonator) — #1211
//   "stack_item"  — either of those (Disallow, Deflecting Swat,
//                   Bolt Bend, Tale's End) — #1211
//   "card_in_graveyard" — a card in any graveyard (Regrowth, Eternal
//                         Witness)
//
// The three stack modes differ only in the sentence the banner
// writes. They light the same surface (the stack overlay, which draws
// spells and abilities in one list) and obey the same legal set — an
// ability is an ordinary card-kind ref carrying its STACK ITEM's id,
// so nothing downstream had to learn a new shape. A picker that read
// the mode for legality would be reading a hint as a rule.
export type TargetingMode =
  | "any"
  | "player"
  | "creature"
  | "permanent"
  | "stack_spell"
  | "stack_ability"
  | "stack_item"
  | "card_in_graveyard";

// CastChoices bundles every announce-time decision collected before
// targeting opens, so a cast rides one object instead of a growing
// tail of positional parameters (the debt ADR 0021 flagged when
// xValue, discardIDs and sacrificeIDs became three of them; S22's
// alternative cost was the fourth and paid it off).
//
// Every field is optional and every field maps to exactly one
// cast_spell param, so applyCastChoices is the single place that
// knows the wire names.
export interface CastChoices {
  // S20 sub-PR 3: the announced X for an {X} spell.
  xValue?: number;
  // S21 sub-PR 5: instance IDs paid to a "discard a card" additional
  // cost.
  discardIDs?: string[];
  // S21 sub-PR 6: the permanent paid to a "sacrifice a creature"
  // additional cost.
  sacrificeIDs?: string[];
  // S22: the key of the alternative cost being paid INSTEAD of the
  // mana cost ("overload", "evoke", "cleave"). Undefined is the
  // ordinary "pay the printed cost" case.
  altCost?: string;
  // S28: the card paid to the non-mana half of that alternative cost
  // — Force of Will's pitched blue card, Daze's returned Island,
  // Solitude's evoke pitch. Exactly one entry when the chosen offer
  // charges one; undefined otherwise, and the server rejects a
  // non-empty list on an offer that charges nothing.
  altCostIDs?: string[];
  // ADR 0073 (#664): the optional additional costs being paid, as
  // POSITIONS in the card's `optional_costs`, repeated once per
  // payment for a multikicker. Undefined and [] are the same thing
  // to the server: decline them all.
  optionalCosts?: number[];
  // CR 702.174a (#1267): the opponent a gift is promised to — a
  // player ID from the gift offer's `opponent_options`. Set exactly
  // when `optionalCosts` claims the gift offer; the server rejects it
  // on a cast that does not.
  giftOpponent?: string;
  // S22: the untapped permanents tapped to help pay — convoke and
  // waterbend. Undefined and empty are the same thing to the server;
  // tapping nothing is always legal.
  tapIDs?: string[];
  // CR 107.4c/f (#916): how many of the cost's Phyrexian symbols are
  // being paid with 2 life each instead of mana. Collected after the
  // X picker — an {X} cost has to be sized before the rest of it can
  // be priced — and bounded by `phyrexian_symbols` and the caster's
  // life total (CR 119.4). Undefined and 0 are the same thing to the
  // server: pay every symbol with its coloured half.
  phyrexianLife?: number;
  // ADR 0034: which printed face of a modal DFC is being cast or
  // played. Undefined and 0 are both "the front face", which is
  // every single-faced card. The face is chosen FIRST — before the
  // alternative cost, the modes, X and the targets — because it
  // decides what the card even is: Sea Gate Restoration and Sea
  // Gate, Reborn have different types, different costs and, via
  // the composite catalog key, different rules.
  face?: number;
  // S29: the zone the cast comes out of. Undefined is the hand,
  // which is every cast the Board's own surfaces fire.
  //
  // It rides CastChoices rather than being a parameter of its own
  // because a graveyard cast has to walk the SAME prompt chain as a
  // hand cast — flashback picks a cost, escape pays an additional
  // cost, Cackling Counterpart picks a target — and threading a
  // second positional argument through eight `after*` seams is
  // exactly the debt this object was created to pay off. The two
  // pre-S29 senders of `from_zone` (the command zone and the exile
  // impulse button) fire bare payloads with no prompts at all, which
  // is why they never needed it.
  fromZone?: CastSourceZone;
}

// CastSourceZone is the `from_zone` vocabulary the server's
// castZoneFromWire accepts. "hand" is never sent — it is the server
// default and omitting it keeps every pre-S29 client's payload
// byte-identical.
export type CastSourceZone = "command" | "exile" | "graveyard";

// applyCastChoices writes a CastChoices onto a cast_spell payload.
// Undefined fields are omitted rather than sent as null — the server
// distinguishes "no additional cost paid" from "paid nothing".
export function applyCastChoices(
  params: Record<string, unknown>,
  choices: CastChoices | undefined,
): void {
  if (!choices) return;
  if (choices.xValue !== undefined) params.x_value = choices.xValue;
  if (choices.discardIDs !== undefined) params.discard_ids = choices.discardIDs;
  if (choices.sacrificeIDs !== undefined) params.sacrifice_ids = choices.sacrificeIDs;
  if (choices.altCost !== undefined) params.alternative_cost = choices.altCost;
  if (choices.altCostIDs !== undefined && choices.altCostIDs.length > 0)
    params.alt_cost_ids = choices.altCostIDs;
  // ADR 0073: omitted when empty, which is the server default and
  // what every client that predates the kicker toggles sends.
  if (choices.optionalCosts !== undefined && choices.optionalCosts.length > 0)
    params.optional_costs = choices.optionalCosts;
  // #1267: omitted unless a gift was promised — a stray one is an
  // error server-side, not a no-op.
  if (choices.giftOpponent !== undefined && choices.giftOpponent !== "")
    params.gift_opponent = choices.giftOpponent;
  if (choices.tapIDs !== undefined && choices.tapIDs.length > 0) params.tap_ids = choices.tapIDs;
  // #916: omitted at 0, which is the server default and what every
  // client that predates the stepper sends.
  if (choices.phyrexianLife !== undefined && choices.phyrexianLife > 0)
    params.phyrexian_life = choices.phyrexianLife;
  // Face 0 is omitted rather than sent explicitly: it is the server
  // default, and `omitempty` on the Go side means an explicit zero
  // and an absent field are the same byte on the wire anyway.
  if (choices.face !== undefined && choices.face > 0) params.face = choices.face;
  // S29: omitted for a hand cast, for the same reason face 0 is.
  if (choices.fromZone !== undefined) params.from_zone = choices.fromZone;
}

// TargetingState is the active prompt. `card` is the spell being
// cast; `mode` is what the UI should accept as a click. The caller
// is responsible for calling cast_spell with the resolved target
// when a surface-level click matches.
export interface TargetingState {
  card: CardView;
  mode: TargetingMode;
  // S20: the server-computed legal set for this card (from
  // CardView.legal_targets). Undefined for free-form cards, where
  // legality falls back to the mode heuristics below.
  legal?: { players: Set<string>; cards: Set<string> };
  // S20 sub-PR 2: set when the prompt answers a pick_target pending
  // choice (a triggered ability choosing its target) rather than a
  // cast. The click resolves the choice instead of firing
  // cast_spell, and the prompt can't be cancelled — the trigger
  // needs a target.
  choiceID?: string;
  // #1196: which KIND of pending choice `choiceID` names, so the
  // banner can say what is being asked. A pick_target is a trigger
  // waiting for its target; a retarget is a spell already on the
  // stack being pointed somewhere else, and "Deflecting Swat
  // triggered" would be the wrong sentence for it.
  choiceKind?: string;
  // CR 603.2d: attribution for an additional trigger awaiting its
  // target. The server supplies the public doubler metadata.
  doubledBy?: string;
  doubledByName?: string;
  // S22: the announce-time payments and choices collected before
  // this prompt opened — X, additional-cost picks, the alternative
  // cost being paid. They ride the cast_spell payload verbatim via
  // applyCastChoices. Absent for pick_target and ability prompts,
  // which aren't casts.
  choices?: CastChoices;
  // S21 sub-PR 2: set when the prompt collects targets for an
  // ACTIVATED ability rather than a cast. The confirm fires
  // activate_ability with these announce-time choices.
  // #625: `counter` is the chosen payment for a "remove N counters"
  // cost, already shaped as the payload's counter_source_ids /
  // counter_kind.
  ability?: {
    index: number;
    sacrificeIDs: string[];
    crewIDs: string[];
    xValue?: number;
    counter?: CounterPayment;
    // #916, CR 107.4f: the Phyrexian symbols this activation pays
    // with 2 life each. Announced with the rest of the cost, before
    // the targets, and it rides the one activate_ability the confirm
    // sends as `phyrexian_life`.
    phyrexianLife?: number;
    // #1296: the ability's price per legal target, when its price
    // reads the target (target_charged_mana_costs). The banner shows
    // it, because an equip's one click is also its confirm.
    prices?: Record<string, string>;
  };
  // Human-readable clause for the banner ("target artifact or
  // enchantment"); the server's TargetSpec label.
  label?: string;
  // S20 sub-PR 4: the chosen mode indexes of a modal spell; ride the
  // cast_spell payload as `modes`.
  modes?: number[];
  // S20 sub-PR 5: the clause's target count. At max 1 the first
  // click completes the prompt; otherwise clicks toggle into
  // `picked` (in click order — positional clauses read it) and the
  // banner's Done fires once at least `min` are picked. max 0 =
  // unbounded.
  min: number;
  max: number;
  picked: TargetRef[];
  // #764 (ADR 0065 §7): an announcement is a WALK over its target
  // clauses. `steps` is the whole walk — one entry per clause of
  // each chosen mode occurrence — `step` is the cursor, and `done`
  // holds the picks already made for earlier steps, each stamped
  // with the clause it answered. The fields above (`mode`, `legal`,
  // `label`, `min`, `max`, `picked`) always describe the CURRENT
  // step, so every existing consumer — the banner, the highlight
  // helpers, the click handlers — needs no change at all.
  //
  // A one-clause non-modal cast is a one-step walk, which is
  // exactly what it was before.
  steps: TargetStep[];
  step: number;
  done: TargetRef[];
}

// TargetStep is one clause of an announcement: which mode occurrence
// and clause slot it answers, its legal set, its printed wording and
// its count.
export interface TargetStep {
  mode: TargetingMode;
  legal?: { players: Set<string>; cards: Set<string> };
  label?: string;
  min: number;
  max: number;
  // modeIndex is the index into the announced `modes` list (the
  // OCCURRENCE, so a repeated mode's two occurrences differ here);
  // slot is the clause index within that occurrence.
  modeIndex: number;
  slot: number;
  // distinct marks a clause whose picks must differ from every
  // EARLIER clause's ("a second target permanent you control").
  distinct: boolean;
}

export interface TargetRef {
  kind: "player" | "card";
  id: string;
  // slot / mode name the clause this pick answers. Omitted (0) for a
  // single-clause non-modal announcement, which is what the server
  // reads as "the only clause".
  slot?: number;
  mode?: number;
}

export const targeting: Writable<TargetingState | null> = guardedWritable(null, "targeting");

// begin enters a targeting prompt. Overwrites any existing prompt
// — the last cast wins. The caller has already verified the
// card's target_mode is non-empty.
//
// `alt` is the alternative cost being paid, when one is (S22) — or,
// since #1267, the claimed optional cost that rewrites the clause
// (see castTargetOverride): its clause replaces the card's, because the spell's targets are
// whatever the cost it was cast for says they are. Wash Away hard-cast
// can only hit a spell that wasn't cast from its owner's hand;
// cleaved it can hit any spell, and the legal set differs
// accordingly.
export function begin(
  card: CardView,
  mode: TargetingMode,
  choices?: CastChoices,
  alt?: TargetClauseOverride,
): void {
  const lt = alt ? alt.legal_targets : card.legal_targets;
  // #764: a card with more than one clause walks them in printed
  // order. An alternative cost replaces the whole statement, so its
  // own clause list wins when it has one.
  const clauses = alt ? alt.clauses : card.clauses;
  const steps = stepsFor(mode, lt, clauses, 0, choices);
  targeting.set(openWalk(card, steps, { choices }));
}

// stepsFor turns one statement — a single `legal_targets` or a
// multi-entry `clauses` list — into the walk's steps.
export function stepsFor(
  mode: TargetingMode,
  lt: LegalTargetsView | undefined,
  clauses: LegalTargetsView[] | undefined,
  modeIndex: number,
  choices?: CastChoices,
): TargetStep[] {
  const list = clauses && clauses.length > 0 ? clauses : lt ? [lt] : [];
  if (list.length === 0) {
    return [{ mode, min: 1, max: 1, modeIndex, slot: 0, distinct: false }];
  }
  return list.map((c, slot) => ({
    mode,
    legal: { players: new Set(c.players ?? []), cards: new Set(c.cards ?? []) },
    label: c.label,
    ...countOf(c, choices),
    modeIndex,
    slot,
    distinct: c.distinct === true,
  }));
}

// openWalk builds the state for the FIRST step of a walk. Extra
// fields (choices, ability, modes, choiceID …) ride through
// untouched.
export function openWalk(
  card: CardView,
  steps: TargetStep[],
  extra: Partial<TargetingState>,
): TargetingState {
  const first = steps[0];
  return {
    card,
    mode: first.mode,
    legal: first.legal,
    label: first.label,
    min: first.min,
    max: first.max,
    picked: [],
    steps,
    step: 0,
    done: [],
    ...extra,
  };
}

// stamped is the current step's picks with their clause coordinates
// written on, which is what the wire carries.
export function stamped(t: TargetingState): TargetRef[] {
  const cur = t.steps[t.step];
  if (!cur) return t.picked;
  return t.picked.map((p) => ({ ...p, mode: cur.modeIndex, slot: cur.slot }));
}

// allPicks is every pick of the walk so far, in step order.
export function allPicks(t: TargetingState): TargetRef[] {
  return [...t.done, ...stamped(t)];
}

// advance closes the current step and opens the next, or returns
// null when the walk is finished and the caller should fire. A step
// whose legal set is empty and whose min is 0 is skipped — there is
// nothing to ask.
export function advance(t: TargetingState): TargetingState | null {
  const done = allPicks(t);
  let next = t.step + 1;
  while (next < t.steps.length) {
    const s = t.steps[next];
    const n = (s.legal?.players.size ?? 0) + (s.legal?.cards.size ?? 0);
    if (n === 0 && s.min === 0) {
      next++;
      continue;
    }
    break;
  }
  if (next >= t.steps.length) return null;
  const s = t.steps[next];
  return {
    ...t,
    mode: s.mode,
    legal: pruneDistinct(s, done),
    label: s.label,
    min: s.min,
    max: s.max,
    picked: [],
    step: next,
    done,
  };
}

// pruneDistinct drops from a step's legal set everything an earlier
// clause already took, when the clause says its pick must differ.
function pruneDistinct(
  s: TargetStep,
  done: TargetRef[],
): { players: Set<string>; cards: Set<string> } | undefined {
  if (!s.legal || !s.distinct || done.length === 0) return s.legal;
  const taken = new Set(done.map((d) => d.id));
  return {
    players: new Set([...s.legal.players].filter((id) => !taken.has(id))),
    cards: new Set([...s.legal.cards].filter((id) => !taken.has(id))),
  };
}

// countOf reads a clause's min / max off the wire; free-form cards
// (no legal set) are single-target.
//
// S22: a clause counted by X ("Exile X target creatures you
// control") carries no usable min / max — the server can't know X
// when it builds the snapshot — so the count comes from the X the
// caster announced in the cost prompts instead. The server rejects
// any other count, so getting this wrong is a rejected cast rather
// than a wrong one.
function countOf(
  lt: LegalTargetsView | undefined,
  choices?: CastChoices,
): { min: number; max: number } {
  if (!lt) return { min: 1, max: 1 };
  if (lt.count_from_x) {
    const x = choices?.xValue ?? 0;
    return { min: x, max: x };
  }
  return { min: lt.min ?? 1, max: lt.max ?? 1 };
}

// isMultiPick reports whether the prompt accumulates picks rather
// than completing on the first click — which is also what decides
// whether the banner grows a Done button.
//
// `min < 1` is the "up to one target" case (Teferi, Time Raveler's
// −3, The Wandering Emperor's −2). Exactly one pick is allowed, but
// zero is a legal answer, so the prompt cannot complete on the first
// click: the player needs a way to say "none". Done is that way, and
// canConfirm lets it fire at zero.
export function isMultiPick(t: TargetingState): boolean {
  return t.max !== 1 || t.min < 1;
}

// isPicked reports whether a target is already in the pick list.
export function isPicked(t: TargetingState, id: string): boolean {
  return t.picked.some((p) => p.id === id);
}

// togglePick adds a target to (or removes it from) a multi-pick
// prompt. Refuses a pick beyond max. Returns the new state.
export function togglePick(t: TargetingState, ref: TargetRef): TargetingState {
  if (isPicked(t, ref.id)) {
    return { ...t, picked: t.picked.filter((p) => p.id !== ref.id) };
  }
  if (t.max > 0 && t.picked.length >= t.max) return t;
  return { ...t, picked: [...t.picked, ref] };
}

// canConfirm reports whether Done may fire: at least min picked.
export function canConfirm(t: TargetingState): boolean {
  return t.picked.length >= t.min;
}

// The banner (mounted by Game.svelte) and the cast flow (Board)
// are separate components; Board registers the handler that turns
// the pick list into a cast_spell / resolve_choice, and the banner's
// Done calls confirm().
let confirmHandler: (() => void) | null = null;
export function setConfirmHandler(fn: (() => void) | null): void {
  confirmHandler = fn;
}
export function confirm(): void {
  confirmHandler?.();
}

// beginForMode enters a targeting prompt for the targeted option of
// a modal spell: the legal set and banner clause come from the
// option, not the card. `modes` is the full chosen set (the targeted
// option plus any untargeted ones) and rides the cast.
export function beginForModes(card: CardView, modes: number[], choices?: CastChoices): boolean {
  const steps = modeSteps(card, modes, choices);
  if (steps.length === 0) return false;
  targeting.set(openWalk(card, steps, { choices, modes }));
  return true;
}

// modeSteps is the walk a modal announcement asks for: every clause
// of every chosen OCCURRENCE, in the order the modes were chosen
// (CR 608.2c). A repeated mode (CR 700.2d) contributes its clauses
// once per occurrence, each with its own modeIndex, which is what
// gives each occurrence its own targets.
export function modeSteps(card: CardView, modes: number[], choices?: CastChoices): TargetStep[] {
  const options = card.modes?.options ?? [];
  const out: TargetStep[] = [];
  modes.forEach((optionIndex, occurrence) => {
    const option = options[optionIndex];
    if (!option?.legal_targets && !option?.clauses?.length) return;
    const mode = (option.target_mode || "any") as TargetingMode;
    for (const s of stepsFor(mode, option.legal_targets, option.clauses, occurrence, choices)) {
      out.push({ ...s, label: s.label || option.label });
    }
  });
  return out;
}

// isModal reports whether a card needs the mode picker before it
// can be cast.
export function isModal(card: CardView): boolean {
  return (card.modes?.options?.length ?? 0) > 0;
}

// discardCostOf returns how many cards a card demands as an
// additional cost to cast, or 0 for the vast majority that demand
// none.
export function discardCostOf(card: CardView): number {
  return card.additional_cost?.discard_cards ?? 0;
}

// sacrificeCostOptions returns the permanents that may pay a
// "sacrifice a creature" additional cost, or undefined when the card
// charges no such cost. An empty array means the cost is unpayable.
export function sacrificeCostOptions(card: CardView): string[] | undefined {
  const opts = card.additional_cost?.sacrifice_options;
  if (!opts) return undefined;
  return opts.cards ?? [];
}

// castSacrificeClause is the sacrifice clause THIS cast is paying
// (ADR 0073 §4): the card's mandatory one, or — when the card has
// none — the clause of the first optional cost the announcement
// claimed. Undefined when the cast owes no sacrifice at all.
//
// One function so the modal's options, its count and its label all
// come from the same clause. The mandatory cost wins because the
// server walks the flat sacrifice_ids list in that order: mandatory
// first, then each claimed optional cost by index.
//
// NOT a merge of the two. No printed card charges a mandatory
// sacrifice AND an optional one, the server would need both payments
// in one flat list to be in plan order, and a picker that silently
// collected two clauses' worth of permanents under one label would be
// lying about what it was asking. If such a card ever prints, this is
// where it gets a second prompt.
export function castSacrificeClause(
  card: CardView,
  choices: CastChoices | undefined,
): LegalTargetsView | undefined {
  const mandatory = card.additional_cost?.sacrifice_options;
  if (mandatory) return mandatory;
  for (const index of choices?.optionalCosts ?? []) {
    const offer = optionalCostsOf(card)[index];
    if (offer?.sacrifice_options) return offer.sacrifice_options;
  }
  return undefined;
}

// castSacrificeLabel is the prompt copy for that clause, as printed —
// "Sacrifice a creature", "Buyback—Sacrifice a land".
export function castSacrificeLabel(
  card: CardView | null,
  choices: CastChoices | undefined,
): string {
  if (!card) return "a permanent";
  if (card.additional_cost?.sacrifice_options) {
    return card.additional_cost.label ?? "a permanent";
  }
  for (const index of choices?.optionalCosts ?? []) {
    const offer = optionalCostsOf(card)[index];
    if (offer?.sacrifice_options) return offer.label ?? "a permanent";
  }
  return "a permanent";
}

// tapCostOf returns a card's convoke / waterbend clause, or
// undefined for the vast majority of cards that offer none (S22).
export function tapCostOf(card: CardView): TapCostView | undefined {
  return card.tap_cost;
}

// tapCostLimit is how many permanents the picker may accept: the cap
// the server sent, or — for a waterbend {X}, whose size the server
// couldn't know when it built the snapshot — the X the caster just
// announced.
export function tapCostLimit(tc: TapCostView, xValue: number | undefined): number {
  if (tc.max && tc.max > 0) return tc.max;
  return xValue ?? 0;
}

// alternativeCostsOf returns the "cast this for its overload / evoke
// / cleave cost instead" offers on a card, or an empty list for the
// vast majority that have none (S22).
export function alternativeCostsOf(card: CardView): AlternativeCostView[] {
  return card.alternative_costs ?? [];
}

// printedCostClaimable reports whether "its mana cost" is one of the
// prices this cast may claim out of the zone the card is in (#1012).
//
// The server decides this — one list of CR 118.9 prices for the
// engine, the bot and the view — and the client reads the answer. It
// used to infer it: a graveyard card with an offer list was assumed
// to be flashback-only and a hand card was assumed not to be, which
// is right for the two common cards and wrong for a Gravecrawler
// under an Underworld Breach, where the printed cost and the granted
// escape cost are both on the menu.
export function printedCostClaimable(card: CardView): boolean {
  return card.alternative_cost_required !== true;
}

// optionalCostsOf returns the "you may pay an additional cost" offers
// on a card — kicker, multikicker, buyback — or an empty list for the
// vast majority that have none (ADR 0073).
export function optionalCostsOf(card: CardView): OptionalCostView[] {
  return card.optional_costs ?? [];
}

// optionalCostMaxTimes is how many times one offer may be paid: 1 for
// kicker and buyback, the multikicker cap above that. The fallback of
// 1 is for an offer that predates the field rather than a guess — a
// repeatable cost always sends its cap.
export function optionalCostMaxTimes(offer: OptionalCostView): number {
  const n = offer.max_times ?? 1;
  return n < 1 ? 1 : n;
}

// optionalCostSelection turns the picker's per-offer counts into the
// wire's index list: paying an offer N times is naming its index N
// times, which is how multikicker announces its count without a
// second field (ADR 0073 §2).
//
// Ascending, so the list the client sends matches the record the
// server normalises it to — two clients that picked the same costs in
// different orders produce the same announcement.
export function optionalCostSelection(counts: Map<number, number>): number[] {
  const out: number[] = [];
  for (const index of [...counts.keys()].sort((a, b) => a - b)) {
    const n = counts.get(index) ?? 0;
    for (let i = 0; i < n; i++) out.push(index);
  }
  return out;
}

// optionalCostOpponentOptions returns the players a gift offer may be
// promised to, or undefined when the offer is not a gift (#1267). An
// empty array means the offer cannot be taken right now — nobody left
// to promise it to.
export function optionalCostOpponentOptions(offer: OptionalCostView): string[] | undefined {
  if (!offer.chooses_opponent) return undefined;
  return offer.opponent_options ?? [];
}

// TargetClauseOverride is the target-clause trio an offer carries when
// paying it rewrites the spell's clause. AlternativeCostView and
// OptionalCostView both satisfy it.
export type TargetClauseOverride = Pick<
  AlternativeCostView,
  "target_mode" | "legal_targets" | "clauses"
>;

function carriesClause(offer: OptionalCostView): boolean {
  return (
    offer.target_mode !== undefined ||
    offer.legal_targets !== undefined ||
    (offer.clauses?.length ?? 0) > 0
  );
}

// castTargetOverride is the clause that replaces the card's own for
// THIS cast, or undefined when the card's printed clause stands.
//
// The alternative cost wins: it replaces the whole statement, and an
// offer with no target_mode (overload) means "no targets" rather than
// "fall back". Otherwise the first claimed optional cost that carries
// a clause — a promised gift that widens Long River's Pull to any
// spell, or adds a target to a card that prints none. No card has
// both, so precedence is a guard rather than a rule anyone relies on.
export function castTargetOverride(
  card: CardView,
  choices: CastChoices | undefined,
): TargetClauseOverride | undefined {
  const alt = alternativeCostByKey(card, choices?.altCost);
  if (alt) return alt;
  for (const index of choices?.optionalCosts ?? []) {
    const offer = optionalCostsOf(card)[index];
    if (offer && carriesClause(offer)) return offer;
  }
  return undefined;
}

// optionalCostPayOptions returns the permanents that can pay an
// offer's sacrifice half, or undefined when it charges none — which
// is every mana kicker. An empty array means the offer cannot be
// taken right now: a Constant Mists with no land to sacrifice.
export function optionalCostPayOptions(offer: OptionalCostView): string[] | undefined {
  if (!offer.sacrifice_options) return undefined;
  return offer.sacrifice_options.cards ?? [];
}

// castIsForbidden reports whether the server's own cast gate has
// already refused this card from the zone it is in (#760). The client
// greys the card rather than dispatching a cast_spell it knows will
// come back as an error.
export function castIsForbidden(card: CardView): boolean {
  return (card.cant_cast ?? "") !== "";
}

// alternativeCostByKey finds the offer a cast is paying. Undefined
// for the ordinary case — `key` undefined means "paying the printed
// mana cost", which is not an offer and has no view.
export function alternativeCostByKey(
  card: CardView,
  key: string | undefined,
): AlternativeCostView | undefined {
  if (key === undefined) return undefined;
  return alternativeCostsOf(card).find((a) => a.key === key);
}

// altCostPayOptions returns the cards that can pay an offer's
// card-shaped half, or undefined when the offer charges none — which
// is every S22 keyword and most S28 ones. An empty array means the
// offer is unpayable right now: a Force of Will with no other blue
// card in hand.
export function altCostPayOptions(offer: AlternativeCostView | undefined): string[] | undefined {
  if (!offer?.pay_options) return undefined;
  return offer.pay_options.cards ?? [];
}

// altCostPayCount is how many cards the offer's card-shaped half
// demands. One for every S28 shape — Force of Will pitches a card,
// Daze bounces an Island — and N for S29's escape, whose cost is
// "exile five other cards from your graveyard".
//
// The server sends the number as min == max on `pay_options`, so the
// picker sizes itself without knowing which keyword it is paying:
// one is a radio list, more than one is a checklist with a counter.
// The fallback of 1 is for an offer that predates the field rather
// than a guess — a cost with a card component always has a count.
export function altCostPayCount(offer: AlternativeCostView | undefined): number {
  const n = offer?.pay_options?.min ?? offer?.pay_options?.max ?? 1;
  return n > 0 ? n : 1;
}

// modeOptionCastable reports whether an option can be chosen right
// now: untargeted options always can; targeted ones need at least
// one legal target.
export function modeOptionCastable(option: ModeOptionView): boolean {
  const lt = option.legal_targets;
  if (!lt) return true;
  return (lt.players?.length ?? 0) + (lt.cards?.length ?? 0) >= (lt.min ?? 1);
}

// hasXCost reports whether a cast needs the X prompt.
//
// Usually that is an {X} in the printed cost. S22 adds a second
// source: a waterbend {X} cost is a cost of its own, layered on top
// of the card's, so Waterbender's Restoration prints {U}{U} and
// still has an X to announce — and that X is also the number of
// creatures its clause targets.
//
// S23 adds the third, and it is the same shape once more: Toxic
// Deluge's "pay X life" is an additional cost with its own X, the
// printed mana cost is a flat {2}{B}, and the announced X is also
// the -X/-X the spell hands out.
export function hasXCost(card: CardView): boolean {
  if (card.tap_cost?.demands_x) return true;
  if (card.additional_cost?.demands_x) return true;
  return (card.mana_cost ?? "").includes("{X}");
}

// castLocksXAtZero reports CR 107.3b: a spell with {X} in its mana
// cost, cast while paying neither that cost nor an alternative cost
// that includes X, has 0 as its only legal X — so there is nothing to
// ask and the picker must not open.
//
// The rule is NOT re-derived here from the cost strings. The server
// computes it with the same predicate it will judge the cast by
// (game.CastCost.LocksXAtZero) and ships the answer per offer:
// `exile_play.x_locked_at_zero` for a cascade hit or a Siege's free
// cast, `alternative_costs[].x_locked_at_zero` for an offer. Asking
// twice in two languages is how a picker ends up collecting a value
// the announce gate rejects.
//
// The grant wins over the offer, because that is the order the server
// prices a cast in: an exile grant's own price replaces whatever cost
// was chosen.
export function castLocksXAtZero(card: CardView, altCost: string | undefined): boolean {
  if (card.exile_play?.x_locked_at_zero) return true;
  return alternativeCostByKey(card, altCost)?.x_locked_at_zero === true;
}

// beginForAbility enters a targeting prompt for an activated
// ability's target clause. `card` is the source permanent; the
// legal set comes from the ability, not the card.
export function beginForAbility(
  card: CardView,
  ability: ActivatedAbilityView,
  sacrificeIDs: string[],
  crewIDs: string[] = [],
  xValue?: number,
  counter?: CounterPayment,
  modes?: number[],
  phyrexianLife?: number,
): void {
  // #764: a modal activated ability announces its modes WITH its
  // targets (CR 602.2b), so the walk is built out of the chosen
  // bullets' clauses exactly as a modal cast's is. A non-modal
  // ability is the same walk over its own clause list — one step for
  // the overwhelming majority, which is what it always was.
  //
  // The count comes off the wire like every other clause's. This used
  // to be a hard-coded 1 / 1 with a note naming the fix — give the
  // view a LegalTargetsView and emit the count server-side — which
  // #334 needed, because Teferi's "up to one target" is Min 0.
  const steps =
    modes && modes.length > 0
      ? abilityModeSteps(ability, modes)
      : stepsFor(
          (ability.target_mode || "any") as TargetingMode,
          ability.legal_targets,
          ability.clauses,
          0,
        );
  targeting.set(
    openWalk(card, steps, {
      // CR 602.2b: X was announced before the targets were chosen and
      // cannot change now — it rides through to the one
      // activate_ability the confirm sends.
      ability: {
        index: ability.index,
        sacrificeIDs,
        crewIDs,
        xValue,
        counter,
        phyrexianLife,
        prices: ability.target_charged_mana_costs,
      },
      modes,
    }),
  );
}

// abilityModeSteps is modeSteps for an activated ability's own
// ModeSpecView.
export function abilityModeSteps(ability: ActivatedAbilityView, modes: number[]): TargetStep[] {
  const options = ability.modes?.options ?? [];
  const out: TargetStep[] = [];
  modes.forEach((optionIndex, occurrence) => {
    const option = options[optionIndex];
    if (!option?.legal_targets && !option?.clauses?.length) return;
    const mode = (option.target_mode || "any") as TargetingMode;
    for (const st of stepsFor(mode, option.legal_targets, option.clauses, occurrence)) {
      out.push({ ...st, label: st.label || option.label });
    }
  });
  return out;
}

// isLegalCardTarget / isLegalPlayerTarget answer "can I click this
// right now?" for a live prompt. With a server legal set (S20
// structured targeting) it's membership; without one (free-form
// S13.1 cards) it's the mode heuristic and the caller's zone
// routing.
export function isLegalCardTarget(t: TargetingState, instanceID: string): boolean {
  if (t.legal) return t.legal.cards.has(instanceID);
  return isTargetingCreature(t.mode) || isTargetingStack(t.mode) || isTargetingGraveyard(t.mode);
}

export function isLegalPlayerTarget(t: TargetingState, playerID: string): boolean {
  if (t.legal) return t.legal.players.has(playerID);
  return isTargetingPlayer(t.mode);
}

// legalTargetCount is the banner's "N legal targets" figure; -1 when
// the prompt is free-form.
export function legalTargetCount(t: TargetingState): number {
  if (!t.legal) return -1;
  return t.legal.players.size + t.legal.cards.size;
}

// beginChoice enters a targeting prompt for a pick_target pending
// choice. `card` is the trigger's source (for the banner); the
// legal set comes from the choice itself.
export function beginChoice(choice: PendingChoiceView, card: CardView): void {
  const pt = choice.pick_target ?? {};
  // A pick_target prompt is always ONE clause: the server walks a
  // multi-clause trigger one prompt at a time (#764), so the client
  // answers each prompt on its own and never holds a walk here.
  targeting.set(
    openWalk(card, stepsFor("any", pt, undefined, 0), {
      choiceID: choice.id,
      choiceKind: choice.kind,
      label: choice.reason,
      doubledBy: choice.doubled_by,
      doubledByName: choice.doubled_by_name,
    }),
  );
}

// cancel clears the prompt without firing cast_spell. Wired to the
// Escape keybinding in Game.svelte. A pick_target prompt is not
// cancellable — the server is waiting for a target — so cancel is a
// no-op for it (the banner hides its Cancel button too).
export function cancel(): void {
  let current: TargetingState | null = null;
  targeting.update((t) => {
    current = t;
    return t;
  });
  if (current && (current as TargetingState).choiceID) return;
  targeting.set(null);
}

// isTargetingPlayer / isTargetingCard / isTargetingStack are mode-
// membership helpers so consumers don't have to pattern-match.
export function isTargetingPlayer(mode: TargetingMode): boolean {
  return mode === "any" || mode === "player";
}

export function isTargetingCreature(mode: TargetingMode): boolean {
  return mode === "any" || mode === "creature" || mode === "permanent";
}

// opensTargetPicker answers the cast flow's one question about a
// card's `target_mode`: does clicking this card open the picker, or
// does it fire cast_spell straight away?
//
// An ALLOWLIST, and the typed narrowing is the point: a mode the
// client does not recognise falls through to an immediate cast with
// no targets, which the server refuses with nothing on screen to
// explain it. That is exactly what happened to every "counter target
// activated or triggered ability" card before #1211 added
// `stack_ability` and `stack_item` here — so the list lives next to
// the type that declares the vocabulary, where adding a mode and
// forgetting this is one file rather than two.
export function opensTargetPicker(mode: string | undefined | null): mode is TargetingMode {
  switch (mode) {
    case "any":
    case "player":
    case "creature":
    case "permanent":
    case "stack_spell":
    case "stack_ability":
    case "stack_item":
    case "card_in_graveyard":
      return true;
    default:
      return false;
  }
}

export function isTargetingStack(mode: TargetingMode): boolean {
  // #1211: all three stack modes route to the same surface. This is
  // the FREE-FORM fallback ("no server legal set, so guess from the
  // mode"); a structured clause never reaches it, which is why the
  // widening cannot make an illegal target clickable.
  return mode === "stack_spell" || mode === "stack_ability" || mode === "stack_item";
}

export function isTargetingGraveyard(mode: TargetingMode): boolean {
  return mode === "card_in_graveyard";
}
