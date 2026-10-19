import { type Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type {
  ActivatedAbilityView,
  AdditionalCostView,
  AlternativeCostView,
  AssistOfferView,
  CardView,
  DivideView,
  GameView,
  LegalTargetsView,
  ModeOptionView,
  OptionalCostView,
  PendingChoiceView,
  PlayerView,
  SpliceOfferView,
  TapCostView,
} from "./protocol";
import type { CounterPayment } from "./counterCost";
import { isBoardPickedCardSetKind } from "./boardAnsweredChoice";

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
//   "card_in_exile" — a card in exile (The Darkness Crystal's "target
//                     creature card exiled with ~"), picked in the zone
//                     browser's exile tab like a graveyard card — #3259
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
  | "card_in_graveyard"
  | "card_in_exile";

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
  // ADR 0109 §9: how many counters the activation's cost is removing —
  // the X a clause bounded by "the counters removed this way" reads
  // (`x_from_counters_removed`). Set only on an ability's walk.
  countersRemoved?: number;
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
  // ADR 0202 (CR 702.47a): the cards in the caster's hand revealed to
  // splice onto this spell — `card_id`s from the card's `splice_offers`,
  // in the order their text happens (the order they were ticked). Each
  // text's clauses join the target walk after the spell's own.
  splices?: string[];
  // CR 702.174a (#1267): the opponent a gift is promised to — a
  // player ID from the gift offer's `opponent_options`. Set exactly
  // when `optionalCosts` claims the gift offer; the server rejects it
  // on a cast that does not.
  giftOpponent?: string;
  // ADR 0150 §5 (CR 702.132a): the player asked to help pay the generic
  // mana of a spell with assist — a player ID from the card's
  // `assist.players`. Undefined casts it alone.
  assistPlayer?: string;
  // ADR 0100 §2: which branch of an either/or additional cost is being
  // paid — an index into `additional_cost.branches`. Required by the
  // server on a card with branches and refused on any other, so it is
  // set only by the cost picker's branch radio.
  costBranch?: number;
  // S22: the untapped permanents tapped to help pay — convoke and
  // waterbend. Undefined and empty are the same thing to the server;
  // tapping nothing is always legal.
  tapIDs?: string[];
  // ADR 0100: the graveyard cards exiled to delve (CR 702.66a), each
  // paying {1} of the generic. Undefined and empty are the same thing
  // to the server; exiling nothing is always legal.
  delveIDs?: string[];
  // #1703: the creatures tapped for a claimed teamwork offer, and the
  // one creature a claimed blight puts its -1/-1 counters on. Set only
  // when `optionalCosts` claims that offer; the server refuses them on
  // a cast that does not.
  teamworkIDs?: string[];
  // #3402: the permanents a mandatory "exile a creature you control"
  // additional cost exiles. Rides cast_spell as `exile_permanent_ids`.
  exilePermanentIDs?: string[];
  // #3175: the hand cards a "you may exile any number of blue cards from
  // your hand" additional cost exiles (March of Swirling Mist). Rides
  // cast_spell as `exile_from_hand_ids`; `[]` is the answer "none".
  exileFromHandIDs?: string[];
  // #3536: the graveyard cards a mandatory "exile N cards from your
  // graveyard" additional cost exiles. Rides cast_spell as `exile_ids`.
  exileIDs?: string[];
  blightIDs?: string[];
  // ADR 0100 amendment 2026-10-07: the one card a chosen reveal / behold
  // branch shows. Set only after the branch is chosen; the server refuses
  // it on a cast that does not pay one.
  revealIDs?: string[];
  // CR 107.4f (#916): how many of the cost's Phyrexian symbols are
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
  // ADR 0103 (CR 702.102a): cast BOTH halves of a split card with fuse,
  // from hand. Sent as `fuse: true`; the face stays 0.
  fuse?: boolean;
  // #2167: the card type a cast through a per-type permission spends
  // (Muldrotha, Aminatou's Augury) — one of the face's
  // `permission_types`. Asked right after the face, because it is a
  // choice about what the chosen face is being cast AS.
  permissionType?: string;
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
  // #1508: the cast was started by dragging the card out of the hand
  // onto the table. It rides CastChoices for the reason fromZone does
  // — a dragged cast walks the SAME prompt chain a click does, and the
  // flag has to survive every hop of it — and applyCastChoices turns
  // it into `strict: true, auto_tap: true` on the wire, whatever the
  // viewer's strictMana setting says (owner decision 3). A click never
  // sets it, so a clicked cast is byte-identical to before.
  viaDrag?: boolean;
  // ADR 0118 §2 (#2188): the cast was started from the "Cast anyway
  // (pay on paper)" row and confirmed. It rides CastChoices for the reason
  // viaDrag does: the face, the costs, X, the modes and the targets are
  // asked exactly as for a click, and the flag must survive every hop.
  // applyCastChoices turns it into `strict: true, force_cast: true`:
  // no mana is spent and the log says so. Life for Phyrexian symbols
  // and every additional cost are still paid.
  forceCast?: boolean;
  // ADR 0131 (#2531): the cast was started from the card menu's "Pay
  // life for {B}…" row, so the Phyrexian stepper opens even when every
  // symbol is a granted one and the mana would have paid. Never sent:
  // it only decides whether the prompt opens.
  askPhyrexianLife?: boolean;
}

// CastSourceZone is the `from_zone` vocabulary the server's
// castZoneFromWire accepts. "hand" is never sent — it is the server
// default and omitting it keeps every pre-S29 client's payload
// byte-identical.
//
// #1440: "library" joins the set for the S42 library-top permissions
// (Bolas's Citadel, Oracle of Mul Daya, Courser of Kruphix) — the same
// wire word `castZoneFromWire` already accepts, sent by the PileBar
// affordance on a visible, playable top card.
//
// #3955: "outside" is the owner's own sideboard (ADR 0198's 2026-10-16
// amendment), castable only under a granted permission (Spawnsire of
// Ulamog's free cast, Wish's play).
export type CastSourceZone = "command" | "exile" | "graveyard" | "library" | "outside";

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
  // ADR 0202: omitted when nothing is spliced, the server default.
  if (choices.splices !== undefined && choices.splices.length > 0) params.splices = choices.splices;
  // #1267: omitted unless a gift was promised — a stray one is an
  // error server-side, not a no-op.
  if (choices.giftOpponent !== undefined && choices.giftOpponent !== "")
    params.gift_opponent = choices.giftOpponent;
  // ADR 0150 §5: omitted unless somebody was asked — the server refuses
  // it on a spell without assist.
  if (choices.assistPlayer !== undefined && choices.assistPlayer !== "")
    params.assist_player = choices.assistPlayer;
  // ADR 0100: sent whenever chosen, branch 0 included — the server
  // never defaults a missing branch.
  if (choices.costBranch !== undefined) params.cost_branch = choices.costBranch;
  if (choices.tapIDs !== undefined && choices.tapIDs.length > 0) params.tap_ids = choices.tapIDs;
  // ADR 0100: omitted when empty, the server default.
  if (choices.delveIDs !== undefined && choices.delveIDs.length > 0)
    params.delve_ids = choices.delveIDs;
  // #1703: omitted unless the offer was claimed and paid.
  if (choices.teamworkIDs !== undefined && choices.teamworkIDs.length > 0)
    params.teamwork_ids = choices.teamworkIDs;
  if (choices.blightIDs !== undefined && choices.blightIDs.length > 0)
    params.blight_ids = choices.blightIDs;
  if (choices.revealIDs !== undefined && choices.revealIDs.length > 0)
    params.reveal_ids = choices.revealIDs;
  if (choices.exilePermanentIDs !== undefined && choices.exilePermanentIDs.length > 0)
    params.exile_permanent_ids = choices.exilePermanentIDs;
  if (choices.exileFromHandIDs !== undefined && choices.exileFromHandIDs.length > 0)
    params.exile_from_hand_ids = choices.exileFromHandIDs;
  if (choices.exileIDs !== undefined && choices.exileIDs.length > 0)
    params.exile_ids = choices.exileIDs;
  // #916: omitted at 0, which is the server default and what every
  // client that predates the stepper sends.
  if (choices.phyrexianLife !== undefined && choices.phyrexianLife > 0)
    params.phyrexian_life = choices.phyrexianLife;
  // Face 0 is omitted rather than sent explicitly: it is the server
  // default, and `omitempty` on the Go side means an explicit zero
  // and an absent field are the same byte on the wire anyway.
  if (choices.face !== undefined && choices.face > 0) params.face = choices.face;
  if (choices.fuse) params.fuse = true;
  // #2167: omitted unless the type picker asked, which it does only when
  // the face has two or more types left; the server settles a single one.
  if (choices.permissionType !== undefined && choices.permissionType !== "")
    params.permission_type = choices.permissionType;
  // S29: omitted for a hand cast, for the same reason face 0 is.
  if (choices.fromZone !== undefined) params.from_zone = choices.fromZone;
  // #1508: a dragged cast always pays strictly and lets the engine tap
  // the lands itself. Written as `strict`, which is also what tells
  // manaEnforcement.ts's stamp to leave the payload alone.
  if (choices.viaDrag) {
    params.strict = true;
    params.auto_tap = true;
  }
  // ADR 0118 §2: a confirmed Cast anyway pays no mana. `strict` keeps
  // manaEnforcement.ts's stamp off the payload, and no auto_tap: the
  // engine must tap nothing for a cast the player chose not to pay.
  if (choices.forceCast) {
    params.strict = true;
    params.force_cast = true;
    delete params.auto_tap;
  }
}

// castChoicesBase is the CastChoices a cast STARTS with, before any
// prompt has asked anything: the zone it comes out of (undefined is the
// hand), whether it was dragged, and whether it is a confirmed Cast
// anyway (ADR 0118 §2). handlePlayCard seeds the chain with it, and the
// face picker stashes it, so none of it is lost when a modal DFC asks
// which face first.
export function castChoicesBase(
  fromZone?: CastSourceZone,
  viaDrag = false,
  forceCast = false,
  askPhyrexianLife = false,
): CastChoices {
  const out: CastChoices = {};
  if (fromZone) out.fromZone = fromZone;
  if (viaDrag) out.viaDrag = true;
  if (forceCast) out.forceCast = true;
  if (askPhyrexianLife) out.askPhyrexianLife = true;
  return out;
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
  // #3817: the pick_target's card-and-ability identity
  // (`shared_target_key`), which groups every copy of the card's trigger.
  sharedKey?: string;
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
  // #1559: the CURRENT step's rule over its chosen set, when it has
  // one — like `legal`, it always describes the step being asked.
  different?: SetRule;
  // #1807: the CURRENT step's sameness rule ("from a single
  // graveyard"), when it has one — every pick must share its key.
  same?: SetRule;
  // #1563: the CURRENT step's divided amount, when its clause is
  // "divided as you choose" — like `legal`, it always describes the
  // step being asked. Already resolved against the X the caster
  // announced.
  divide?: number;
  // #1659: the CURRENT step's divide is X-based and X isn't known
  // yet — see TargetStep.divideFromXUnresolved.
  divideFromXUnresolved?: boolean;
  // #1657: the CURRENT step divides "up to" its amount — see
  // TargetStep.divideUpTo.
  divideUpTo?: boolean;
  // #1563: the division announced so far, target id → share, for
  // every divided step already answered. Rides the action as
  // `distribution`.
  distribution?: Record<string, number>;
}

// SetRule is a clause's rule over the chosen SET of its picks (#1559,
// CR 601.2c): under `different` no two picks may share a key, under
// `same` (#1807) every pick must. `label` completes "those targets
// must …". A card missing from `keys` collides with nothing.
export interface SetRule {
  label: string;
  keys: Record<string, string>;
}

// TargetStep is one clause of an announcement: which mode occurrence
// and clause slot it answers, its legal set, its printed wording and
// its count.
export interface TargetStep {
  // #1943: the second clause is narrowed once its anchor is picked.
  relation?: LegalTargetsView["relation"];
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
  // ADR 0202: the spliced text this step asks for, 1-based in the
  // cast's `splices`; undefined for the spell's own text.
  splice?: number;
  // distinct marks a clause whose picks must differ from every
  // EARLIER clause's ("a second target permanent you control").
  distinct: boolean;
  // #1559: the clause's set rule, when it prints one.
  different?: SetRule;
  // #1807: the clause's sameness rule, when it prints one.
  same?: SetRule;
  // #3860: the extra life naming each card costs (CR 601.2f), and the
  // cards whose tax the viewer cannot pay — already left out of `legal`.
  lifeTaxes?: Record<string, number>;
  lifeTaxesUnpayable?: string[];
  // #1563: the amount the clause divides among its picks, resolved
  // against the announced X. Undefined for a clause that divides
  // nothing.
  divide?: number;
  // #1659: true when the clause divides an X-based amount but X
  // hasn't been collected for this walk — the banner shows "X"
  // rather than the 0 `divide` resolves to in that case (every real
  // walk collects X before targeting starts, so this is a defensive
  // fallback, not a state the client should reach).
  divideFromXUnresolved?: boolean;
  // #1657: "distribute UP TO that many" (Lathiel) — the shares may sum
  // to less than `divide`, each pick still at least 1.
  divideUpTo?: boolean;
}

export interface TargetRef {
  kind: "player" | "card";
  id: string;
  // slot / mode name the clause this pick answers. Omitted (0) for a
  // single-clause non-modal announcement, which is what the server
  // reads as "the only clause".
  slot?: number;
  mode?: number;
  // ADR 0202: the spliced text this pick answers (1-based); omitted for
  // the spell's own text.
  splice?: number;
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
): boolean {
  const lt = alt ? alt.legal_targets : card.legal_targets;
  // #764: a card with more than one clause walks them in printed
  // order. An alternative cost replaces the whole statement, so its
  // own clause list wins when it has one.
  const clauses = alt ? alt.clauses : card.clauses;
  const steps = [...stepsFor(mode, lt, clauses, 0, choices), ...spliceSteps(card, choices)];
  // #2055: every clause is an opponent's choice (Evangelize) and no
  // splice adds one: there is nothing for the caster to pick, and the
  // caller casts at once.
  if (steps.length === 0) return false;
  targeting.set(openWalk(card, steps, { choices }));
  return true;
}

// spliceOffersOf is the cards in the viewer's hand that splice onto
// `card` as it is cast (ADR 0202).
export function spliceOffersOf(card: CardView): SpliceOfferView[] {
  return card.splice_offers ?? [];
}

// spliceSteps is the target walk's steps for the spliced texts of a
// cast (CR 702.47d): each ticked card's clauses, in the order ticked,
// after the spell's own, each stamped with the text it answers. A text
// with no target adds none.
export function spliceSteps(card: CardView, choices: CastChoices | undefined): TargetStep[] {
  const offers = spliceOffersOf(card);
  const out: TargetStep[] = [];
  (choices?.splices ?? []).forEach((id, k) => {
    const offer = offers.find((o) => o.card_id === id);
    if (!offer || (!offer.legal_targets && !(offer.clauses?.length ?? 0))) return;
    const mode = (offer.target_mode || "any") as TargetingMode;
    for (const s of stepsFor(mode, offer.legal_targets, offer.clauses, 0, choices)) {
      out.push({
        ...s,
        splice: k + 1,
        label: s.label ? `${s.label} (${offer.name})` : `${offer.name}'s target`,
      });
    }
  });
  return out;
}

// beginSpliceOnly opens the walk for a cast whose own text takes no
// target but whose spliced text does (Glacial Ray spliced onto Kodama's
// Reach). False, and nothing opened, when no spliced text targets.
export function beginSpliceOnly(card: CardView, choices: CastChoices, modes?: number[]): boolean {
  const steps = spliceSteps(card, choices);
  if (steps.length === 0) return false;
  targeting.set(openWalk(card, steps, modes !== undefined ? { choices, modes } : { choices }));
  return true;
}

// stepsFor turns one statement — a single `legal_targets` or a
// multi-entry `clauses` list — into the walk's steps.
//
// #2055: a clause "of an opponent's choice" is not a step. Its slot is
// skipped (every other step keeps its own slot number), and the server
// asks an opponent for it once the announcement arrives. A statement
// made only of such clauses is an empty walk.
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
  return list.flatMap((c, slot): TargetStep[] =>
    c.chosen_by_opponent
      ? []
      : [
          {
            mode,
            // #3860: a player who can't pay a card's life tax can't target it.
            legal: {
              players: new Set(c.players ?? []),
              cards: new Set(
                withinX(c, choices).filter((id) => !c.life_taxes_unpayable?.includes(id)),
              ),
            },
            lifeTaxes: c.life_taxes,
            lifeTaxesUnpayable: c.life_taxes_unpayable,
            label: c.label,
            ...countOf(c, choices),
            modeIndex,
            slot,
            distinct: c.distinct === true,
            different: c.different
              ? { label: c.different.label, keys: c.different.keys ?? {} }
              : undefined,
            same: c.same ? { label: c.same.label, keys: c.same.keys ?? {} } : undefined,
            relation: c.relation,
            divide: c.divide ? divideTotal(c.divide, choices?.xValue) : undefined,
            divideFromXUnresolved: c.divide?.from_x === true && choices?.xValue === undefined,
            divideUpTo: c.divide?.up_to === true ? true : undefined,
          },
        ],
  );
}

// divideTotal is the amount a divided clause splits (#1563) — the
// client twin of game.(*DivideSpec).TotalFor: a fixed total, or the X
// the caster announced, doubled once X reaches `double_from_x`.
export function divideTotal(d: DivideView, xValue: number | undefined): number {
  if (!d.from_x) return d.total ?? 0;
  const x = Math.max(0, xValue ?? 0);
  if (d.double_from_x && d.double_from_x > 0 && x >= d.double_from_x) return 2 * x;
  return x;
}

// evenSplit is the division the picker opens with, and the one the
// bot announces (game.EvenDistribution): the amount split as evenly as
// possible, the remainder one point at a time to the earliest picks.
export function evenSplit(ids: string[], total: number): Record<string, number> {
  const out: Record<string, number> = {};
  if (ids.length === 0) return out;
  const share = Math.floor(total / ids.length);
  const extra = total % ids.length;
  ids.forEach((id, i) => {
    out[id] = share + (i < extra ? 1 : 0);
  });
  return out;
}

// divisionProblem says why a division would be refused (CR 601.2d),
// or null when the server will take it: every pick at least 1, the
// shares adding up to the amount — or, with `upTo` (#1657), to at most
// the amount. The same rules the engine's settleDistribution enforces,
// so Confirm is enabled exactly when the answer is legal.
export function divisionProblem(
  ids: string[],
  total: number,
  dist: Record<string, number>,
  upTo = false,
): string | null {
  if (ids.length > total) {
    return `${ids.length} targets can't share ${total} — each needs at least 1`;
  }
  let sum = 0;
  for (const id of ids) {
    const v = dist[id] ?? 0;
    if (!Number.isInteger(v) || v < 1) return "each target needs at least 1";
    sum += v;
  }
  if (upTo) {
    // #1657: "up to that many" — less is fine, more is not.
    if (sum > total) return `assign at most ${total} in all (${sum} so far)`;
    return null;
  }
  if (sum !== total) return `assign ${total} in all (${sum} so far)`;
  return null;
}

// needsDivision reports whether completing the CURRENT step has to ask
// for a division first: its clause divides and two or more targets are
// picked. One pick takes the whole amount and needs no question — the
// server fills it in.
export function needsDivision(t: TargetingState): boolean {
  return t.divide !== undefined && t.picked.length >= 2;
}

// withDivision records the current step's division on the state, to
// ride the action with every other divided step's.
export function withDivision(t: TargetingState, dist: Record<string, number>): TargetingState {
  return { ...t, distribution: { ...(t.distribution ?? {}), ...dist } };
}

// distributionOf is the `distribution` a finished walk sends, or
// undefined when nothing was divided among two or more targets.
export function distributionOf(t: TargetingState): Record<string, number> | undefined {
  const d = t.distribution;
  return d && Object.keys(d).length > 0 ? d : undefined;
}

// withinX narrows an X-bounded clause's cards ("with mana value X or
// less", #1559; "with mana value X", "with power X or less", "with
// toughness X or less", ADR 0109 §9) to the ones the announcement
// admits. The server built the legal set before X was chosen, exactly
// as it does for count_from_x, so the bound is applied here from the
// values it shipped; a card with no entry meets no bound. The X is the
// announced one, or the counters the activation is removing when the
// clause says so. Every other clause passes through untouched.
export function withinX(lt: LegalTargetsView, choices?: CastChoices): string[] {
  const cards = lt.cards ?? [];
  let values: Record<string, number> | undefined;
  if (lt.mana_value_at_most_x || lt.mana_value_equals_x) values = lt.mana_values;
  else if (lt.power_at_most_x) values = lt.powers;
  else if (lt.toughness_at_most_x) values = lt.toughnesses;
  else return cards;
  const x = lt.x_from_counters_removed ? (choices?.countersRemoved ?? 0) : (choices?.xValue ?? 0);
  const vals = values ?? {};
  const exact = lt.mana_value_equals_x === true;
  return cards.filter((id) => vals[id] !== undefined && (exact ? vals[id] === x : vals[id] <= x));
}

// countersRemovedOf is how many counters an activation's counter
// payment removes (ADR 0109 §9): the per-permanent counts it sends, or
// the printed count once per permanent when it sends none. Zero with no
// counter component.
export function countersRemovedOf(
  counter: CounterPayment | undefined,
  printedN: number | undefined,
): number {
  if (!counter) return 0;
  if (counter.counter_counts && counter.counter_counts.length > 0) {
    return counter.counter_counts.reduce((a, b) => a + b, 0);
  }
  const n = printedN ?? 0;
  const sources = counter.counter_source_ids?.length ?? (n > 0 ? 1 : 0);
  return n * Math.max(sources, 1);
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
    different: first.different,
    same: first.same,
    divide: first.divide,
    divideUpTo: first.divideUpTo,
    divideFromXUnresolved: first.divideFromXUnresolved,
    ...extra,
  };
}

// stamped is the current step's picks with their clause coordinates
// written on, which is what the wire carries.
export function stamped(t: TargetingState): TargetRef[] {
  const cur = t.steps[t.step];
  if (!cur) return t.picked;
  return t.picked.map((p) =>
    cur.splice
      ? { ...p, mode: cur.modeIndex, slot: cur.slot, splice: cur.splice }
      : { ...p, mode: cur.modeIndex, slot: cur.slot },
  );
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
    const legal = pruneRelation(s, done);
    const n = (legal?.players.size ?? 0) + (legal?.cards.size ?? 0);
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
    legal: pruneRelation(s, done),
    label: s.label,
    min: s.min,
    max: s.max,
    picked: [],
    step: next,
    done,
    different: s.different,
    same: s.same,
    divide: s.divide,
    divideUpTo: s.divideUpTo,
    divideFromXUnresolved: s.divideFromXUnresolved,
  };
}

function pruneRelation(s: TargetStep, done: TargetRef[]): TargetStep["legal"] {
  const legal = pruneDistinct(s, done);
  if (!s.relation || !legal) return legal;
  const anchors = done.filter(
    (p) => (p.mode ?? 0) === s.modeIndex && (p.slot ?? 0) === s.relation?.slot,
  );
  const allowed = new Set(anchors.length === 1 ? (s.relation.pairs[anchors[0].id] ?? []) : []);
  return {
    players: new Set([...legal.players].filter((id) => allowed.has(id))),
    cards: new Set([...legal.cards].filter((id) => allowed.has(id))),
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
    // "Up to X": the announced X is a ceiling and zero picks is legal.
    return { min: lt.up_to_x ? 0 : x, max: x };
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
  // #1563: a divided clause gives each target at least 1, so it takes
  // no more targets than its amount.
  if (t.divide !== undefined && t.picked.length >= t.divide) return t;
  // #1559: the server would refuse the set, so the pick is refused
  // here — the card is already greyed by isLegalCardTarget.
  if (breaksSetRule(t, ref.id)) return t;
  return { ...t, picked: [...t.picked, ref] };
}

// canConfirm reports whether Done may fire: at least min picked.
export function canConfirm(t: TargetingState): boolean {
  return t.picked.length >= t.min;
}

// The dock's targeting prompt (lib/targetingDock.ts, opened by
// Game.svelte; it was TargetingBanner until ADR 0111 PR 4) and the cast
// flow (Board) are separate components; Board registers the handler
// that turns the pick list into a cast_spell / resolve_choice, and the
// dock's Done (or Enter) calls confirm().
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
  // ADR 0202: the spliced texts' clauses follow the chosen modes'.
  const steps = [...modeSteps(card, modes, choices), ...spliceSteps(card, choices)];
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
// none. For an either/or cost pass the cast's choices: the discard is
// the chosen branch's (castAdditionalCost).
export function discardCostOf(card: CardView, choices?: CastChoices): number {
  return castAdditionalCost(card, choices)?.discard_cards ?? 0;
}

// costBranchesOf returns the branches of a card's either/or additional
// cost (ADR 0100 §2), or an empty list for every other card.
export function costBranchesOf(card: CardView): AdditionalCostView[] {
  return card.additional_cost?.branches ?? [];
}

// castAdditionalCost is the mandatory additional cost THIS cast pays:
// the card's own, or — for an either/or cost — the branch the cast has
// chosen. Undefined for a branched card with no branch chosen yet, and
// for a card with no additional cost at all.
export function castAdditionalCost(
  card: CardView,
  choices: CastChoices | undefined,
): AdditionalCostView | undefined {
  const ac = card.additional_cost;
  if (!ac) return undefined;
  const branches = ac.branches ?? [];
  if (branches.length === 0) return ac;
  const i = choices?.costBranch;
  return i === undefined ? undefined : branches[i];
}

// firstPayableBranch is the default the branch radio opens on: the
// first branch the server says the viewer can pay, or undefined when
// none can (the cast is then not offered at all — CR 601.2h).
export function firstPayableBranch(card: CardView): number | undefined {
  const i = costBranchesOf(card).findIndex((b) => b.payable === true);
  return i < 0 ? undefined : i;
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
  // ADR 0100: an either/or cost's chosen branch is the mandatory one.
  const mandatory = castAdditionalCost(card, choices)?.sacrifice_options;
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
  const mandatory = castAdditionalCost(card, choices);
  if (mandatory?.sacrifice_options) {
    return mandatory.label ?? "a permanent";
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

// tapCostFor is the tap cost the cast being built pays: the card's own
// convoke / waterbend clause, or (#3999) the picker of a chosen "you may
// waterbend {N}" offer, which the server makes the cast's tap cost while
// it is chosen. A cast carries one tap cost, so the first one found is
// the one.
export function tapCostFor(card: CardView, choices?: CastChoices): TapCostView | undefined {
  if (card.tap_cost) return card.tap_cost;
  for (const index of choices?.optionalCosts ?? []) {
    const tc = optionalCostsOf(card).find((o) => o.index === index)?.tap_cost;
    if (tc) return tc;
  }
  return undefined;
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

// printedCostCastableNow narrows printedCostClaimable to "and the
// engine would accept it AT THIS MOMENT" (#1686). The two disagree
// only for a card carrying a granted alternative cost with its own
// clock — a live miracle grant is instant-speed for its OWN claim and
// says nothing about the printed one, which stays whatever timing the
// card prints. Drawn on another player's turn, the printed sorcery
// cost is still on the menu (`printedCostClaimable` stays true) but
// not choosable until the caster's own main phase
// (`printed_cost_timing_closed`).
//
// Every other card answers both questions the same way, because
// nothing else on the wire carries a second claim with a different
// clock — this is deliberately a second predicate rather than a
// change to `printedCostClaimable` itself, which the zone browser's
// button label and canCastFromHand's tooltip also read and which must
// keep answering the zone-and-payability question alone (see their
// own call sites for why timing does not belong there).
export function printedCostCastableNow(card: CardView): boolean {
  return printedCostClaimable(card) && card.printed_cost_timing_closed !== true;
}

// castableAlternativeCostsOf narrows alternativeCostsOf to the offers
// the engine would accept RIGHT NOW (#1686) — dropping one whose
// `timing_closed` bit is set. The cost picker is the one reader that
// needs this: it is the only place a card's own claim and a granted
// claim with a different clock are ever offered side by side.
export function castableAlternativeCostsOf(card: CardView): AlternativeCostView[] {
  return alternativeCostsOf(card).filter((o) => o.timing_closed !== true);
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

// assistOf is a spell with assist's offer (ADR 0150 §5): "up to {max}"
// and the players who may be asked, or undefined for a card without
// assist. A forced cast pays no mana, so it has nothing to help with.
export function assistOf(card: CardView, forceCast = false): AssistOfferView | undefined {
  if (forceCast || !card.assist || card.assist.max <= 0) return undefined;
  return card.assist;
}

// assistPlayerLabel names one player in the Assist row. A bot always
// answers 0 (ADR 0150 owner decision 8), so its seat says so.
export function assistPlayerLabel(seat: Pick<PlayerView, "name" | "is_bot"> | undefined): string {
  if (!seat) return "A player";
  return seat.is_bot ? `${seat.name} (always declines)` : seat.name;
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

// modesUnderChoices is the card the mode picker should open on for
// THIS cast (#1655): the card itself, or — when the caster has ticked
// an optional cost and the server said that changes the mode count —
// a copy whose `modes` carries `if_optional_paid`'s bounds. CR 601.2b
// announces the kicker with the modes, so a kicked Inscription of Ruin
// may take any number of bullets and an unkicked one exactly one.
export function modesUnderChoices(card: CardView, choices: CastChoices | undefined): CardView {
  const modes = card.modes;
  const paid = modes?.if_optional_paid;
  if (!modes || !paid || (choices?.optionalCosts?.length ?? 0) === 0) return card;
  return { ...card, modes: { ...modes, min: paid.min, max: paid.max } };
}

// optionalCostPayOptions returns the permanents that can pay an
// offer's sacrifice half, or undefined when it charges none — which
// is every mana kicker. An empty array means the offer cannot be
// taken right now: a Constant Mists with no land to sacrifice.
export function optionalCostPayOptions(offer: OptionalCostView): string[] | undefined {
  // #1703: teamwork's creatures and blight's are the same question —
  // present-and-empty is an offer the board cannot pay right now.
  if (offer.teamwork_options) return offer.teamwork_options.cards ?? [];
  if (offer.blight_options) return offer.blight_options.cards ?? [];
  // #4092: a replicate cost that taps — too few to tap is an offer that
  // cannot be taken now.
  if (offer.tap_others_options) {
    const cards = offer.tap_others_options.cards ?? [];
    return cards.length < (offer.tap_others_options.min ?? 1) ? [] : cards;
  }
  if (!offer.sacrifice_options) return undefined;
  return offer.sacrifice_options.cards ?? [];
}

// ClaimedCreatureCost is a claimed teamwork or blight offer: the
// number the card prints and the creatures the server says could pay
// it (#1703).
export interface ClaimedCreatureCost {
  // The claimed optional offer, or (ADR 0100) the chosen branch of an
  // either/or cost; only its printed label is read.
  offer: Pick<OptionalCostView, "label">;
  n: number;
  options: string[];
}

function claimedOffer(
  card: CardView,
  choices: CastChoices | undefined,
  pick: (o: OptionalCostView) => number | undefined,
  opts: (o: OptionalCostView) => LegalTargetsView | undefined,
): ClaimedCreatureCost | undefined {
  for (const index of choices?.optionalCosts ?? []) {
    const offer = optionalCostsOf(card)[index];
    const n = offer ? pick(offer) : undefined;
    if (offer && n !== undefined && n > 0) {
      return { offer, n, options: opts(offer)?.cards ?? [] };
    }
  }
  return undefined;
}

// castTeamworkOffer is the teamwork offer this cast has claimed, if
// any — the prompt that follows the add-on picker asks which creatures
// to tap (#1703).
export function castTeamworkOffer(
  card: CardView,
  choices: CastChoices | undefined,
): ClaimedCreatureCost | undefined {
  return claimedOffer(
    card,
    choices,
    (o) => o.teamwork,
    (o) => o.teamwork_options,
  );
}

// castBlightOffer is the blight offer this cast has claimed, if any —
// the prompt asks which one creature takes the counters (#1703).
export function castBlightOffer(
  card: CardView,
  choices: CastChoices | undefined,
): ClaimedCreatureCost | undefined {
  // ADR 0100: a chosen either/or branch that blights ("blight 2 or pay
  // {1}") asks the same question with the same picker.
  // #2174: "blight X" asks for the same one-creature pick; X itself is
  // announced at the X prompt that follows (n is 0 here, so the label
  // names X rather than a number).
  const mandatory = card.additional_cost;
  if (mandatory?.blight_x) {
    // No creature means the ceiling is 0: nothing to pick, X is 0.
    if ((mandatory.blight_options?.cards ?? []).length === 0) return undefined;
    return { offer: mandatory, n: 0, options: mandatory.blight_options?.cards ?? [] };
  }
  const branch = costBranchesOf(card).length > 0 ? castAdditionalCost(card, choices) : undefined;
  if (branch?.blight !== undefined && branch.blight > 0) {
    return { offer: branch, n: branch.blight, options: branch.blight_options?.cards ?? [] };
  }
  return claimedOffer(
    card,
    choices,
    (o) => o.blight,
    (o) => o.blight_options,
  );
}

// ClaimedRevealCost is the reveal / behold branch this cast has chosen:
// what the picker should say and which cards could pay it.
export interface ClaimedRevealCost {
  label: string;
  behold: boolean;
  options: string[];
}

// castRevealCost is the reveal / behold cost the chosen either/or branch
// pays, if any (ADR 0100 amendment 2026-10-07) — the prompt after the
// branch asks which card to show.
export function castRevealCost(
  card: CardView,
  choices: CastChoices | undefined,
): ClaimedRevealCost | undefined {
  const branch = costBranchesOf(card).length > 0 ? castAdditionalCost(card, choices) : undefined;
  if (!branch?.reveal) return undefined;
  return {
    label: branch.label ?? (branch.behold ? "Behold" : "Reveal a card from your hand"),
    behold: branch.behold === true,
    options: branch.reveal_options?.cards ?? [],
  };
}

// ClaimedExilePermanentCost is the "exile a creature you control"
// additional cost this cast pays (#3402): the clause as printed, how many
// to pick and which permanents could pay it.
export interface ClaimedExilePermanentCost {
  label: string;
  count: number;
  options: string[];
}

// castExilePermanentCost is the exile-a-permanent cost a cast pays, if
// any: the card's mandatory additional cost, or the chosen either/or
// branch's. The prompt after the sacrifice step asks which to exile.
export function castExilePermanentCost(
  card: CardView,
  choices: CastChoices | undefined,
): ClaimedExilePermanentCost | undefined {
  const cost = castAdditionalCost(card, choices);
  const opts = cost?.exile_permanent_options;
  if (!cost || !opts) return undefined;
  return {
    label: cost.exile_permanent_label ?? "a permanent you control",
    count: opts.min ?? 1,
    options: opts.cards ?? [],
  };
}

// ClaimedTapCost is the "tap an untapped Vampire you control" additional
// cost this cast pays (#3868): the clause as printed, how many to tap and
// which permanents could pay it. `min` is the fewest the cast may name:
// the count for a fixed clause, 0 for "tap any number of untapped
// creatures you control" (#3961, Burn at the Stake), whose `count` is
// then every option.
export interface ClaimedTapCost {
  label: string;
  count: number;
  min: number;
  options: string[];
}

// castTapCost is the tap-permanents cost a cast pays, if any: the card's
// mandatory additional cost, or the chosen either/or branch's (Disruption
// Protocol). The picks ride cast_spell as `teamwork_ids`.
export function castTapCost(
  card: CardView,
  choices: CastChoices | undefined,
): ClaimedTapCost | undefined {
  const cost = castAdditionalCost(card, choices);
  const opts = cost?.tap_others_options;
  if (!cost || !opts) return castReplicateTapCost(card, choices);
  const options = opts.cards ?? [];
  const min = opts.min ?? 1;
  return {
    label: cost.tap_others_label ?? "an untapped permanent you control",
    // #3961: min 0 is "any number": up to every option.
    count: min === 0 ? options.length : min,
    min,
    options,
  };
}

// castReplicateTapCost is a claimed replicate offer that taps (#4092,
// Psionic Ritual): each payment the cast announces taps the offer's
// count, so paying it twice asks for two Horrors on the same picker and
// the same `teamwork_ids` list.
function castReplicateTapCost(
  card: CardView,
  choices: CastChoices | undefined,
): ClaimedTapCost | undefined {
  let offer: OptionalCostView | undefined;
  let times = 0;
  for (const index of choices?.optionalCosts ?? []) {
    const o = optionalCostsOf(card)[index];
    if (o?.tap_others_options) {
      offer = o;
      times++;
    }
  }
  if (!offer?.tap_others_options) return undefined;
  const count = (offer.tap_others_options.min ?? 1) * times;
  return {
    label: offer.tap_others_label ?? "an untapped permanent you control",
    count,
    min: count,
    options: offer.tap_others_options.cards ?? [],
  };
}

// ClaimedHandExileCost is the "you may exile any number of blue cards
// from your hand" additional cost a cast pays (#3175): the clause as
// printed and the hand cards that could pay it. Any number, none included.
export interface ClaimedHandExileCost {
  label: string;
  options: string[];
}

// castHandExileCost is the hand-exile cost a cast pays, if any: the card's
// mandatory additional cost. The prompt after the sacrifice step asks
// which cards to exile, and skips itself when the hand holds none.
export function castHandExileCost(
  card: CardView,
  choices: CastChoices | undefined,
): ClaimedHandExileCost | undefined {
  const cost = castAdditionalCost(card, choices);
  const opts = cost?.exile_from_hand_options;
  if (!cost || !opts) return undefined;
  return {
    label: cost.exile_from_hand_label ?? "any number of cards from your hand",
    options: opts.cards ?? [],
  };
}

// altCostPayDefinesX reports whether the chosen offer's card IS the X
// announcement (#3175): Shining Shoal's "exile a white card with mana
// value X from your hand". The server ships the picker with
// `mana_value_equals_x` and each card's `mana_values`, and the client
// sends the picked card's value as the x_value instead of asking for X.
export function altCostPayDefinesX(offer: AlternativeCostView | undefined): boolean {
  return offer?.pay_options?.mana_value_equals_x === true;
}

// altCostPayX is the X a pitch bound to X announces: the picked card's
// mana value, or undefined when the offer does not bind X or the server
// sent no value for the card.
export function altCostPayX(
  offer: AlternativeCostView | undefined,
  picked: string[],
): number | undefined {
  if (!altCostPayDefinesX(offer) || picked.length !== 1) return undefined;
  return offer?.pay_options?.mana_values?.[picked[0]];
}

// ClaimedExileCardsCost is the "exile N cards from your graveyard"
// additional cost this cast pays (#3536): the clause as printed, how many
// cards to pick and which could pay it. `cost` is the additional-cost view
// itself, so the picker resolves the ids through exileCost.ts exactly as an
// ability's does.
export interface ClaimedExileCardsCost {
  label: string;
  count: number;
  options: string[];
  cost: AdditionalCostView;
  // #1948: the count is the announced X ("exile X cards"), so the picker
  // opens after the X prompt; `count` is the X announced so far (0 until
  // then).
  fromX: boolean;
}

// castExileCardsCost is the exile-cards cost a cast pays, if any: the
// card's mandatory additional cost (or the chosen either/or branch). The
// prompt after the exile-a-permanent step asks which cards to exile; for
// the X form it is asked after the X prompt instead, with X as its count
// and the "X" in its label replaced by the number.
export function castExileCardsCost(
  card: CardView,
  choices: CastChoices | undefined,
): ClaimedExileCardsCost | undefined {
  const cost = castAdditionalCost(card, choices);
  if (!cost) return undefined;
  if (cost.exile_cost_x) {
    const x = choices?.xValue ?? 0;
    return {
      label: (cost.exile_cost_label ?? "X cards").replace(/\bX\b/, String(x)),
      count: x,
      options: cost.exile_cost_options ?? [],
      cost,
      fromX: true,
    };
  }
  if (!cost.exile_cost_n) return undefined;
  return {
    label: cost.exile_cost_label ?? "cards",
    count: cost.exile_cost_n,
    options: cost.exile_cost_options ?? [],
    cost,
    fromX: false,
  };
}

// exileXChoice answers an "exile X cards" cost without a prompt when
// there is nothing to choose (#1948): X = 0 names no card, and a
// graveyard holding exactly X matching cards names all of them. Returns
// the choices with `exileIDs` set, or undefined when the player must pick.
export function exileXChoice(
  cost: ClaimedExileCardsCost,
  choices: CastChoices,
): CastChoices | undefined {
  if (cost.count <= 0) return { ...choices, exileIDs: [] };
  if (cost.options.length <= cost.count) {
    return { ...choices, exileIDs: cost.options.slice(0, cost.count) };
  }
  return undefined;
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

// castsFaceName is the name of the face an offer casts, for the picker
// (ADR 0152 §5): "Goldbug, Scrappy Scout" under More Than Meets the Eye,
// the back face under disturb. Undefined when the offer casts the face
// the card shows, or the card carries no such face.
export function castsFaceName(card: CardView, offer: AlternativeCostView): string | undefined {
  if (!offer.casts_face) return undefined;
  return card.faces?.[offer.casts_face]?.name;
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

// altCostSacrificeClause returns an offer's sacrifice half (#1727) —
// Dread Return's "Flashback—Sacrifice three creatures", Fireblast's two
// Mountains — or undefined when the offer sacrifices nothing. The cast
// flow opens SacrificeCostModal on it rather than AltCostPaymentModal,
// because a sacrifice is the picker every other sacrifice cost uses;
// the picks still ride `alt_cost_ids`.
export function altCostSacrificeClause(
  offer: AlternativeCostView | undefined,
): LegalTargetsView | undefined {
  return offer?.sacrifice_options;
}

// altCostTapClause returns an offer's tap half (ADR 0135 §1) — Orim's
// Cure's "tap an untapped creature you control", Battle Screech's
// "Flashback—Tap three untapped white creatures you control" — or
// undefined when the offer taps nothing. The cast flow opens the tap
// picker on it, even when the board offers exactly the count: tapping a
// blocker is a choice the player should see. The picks ride
// `alt_cost_ids`.
export function altCostTapClause(
  offer: AlternativeCostView | undefined,
): LegalTargetsView | undefined {
  return offer?.tap_options;
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
// now: not one the ability has already used (ADR 0097), untargeted
// options otherwise always can, and targeted ones need at least one
// legal target.
export function modeOptionCastable(option: ModeOptionView): boolean {
  if (option.used) return false;
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
//
// #1657 adds the fourth: an alternative cost priced with an {X} of its
// own on a card whose printed cost has none — Avacyn's Judgment prints
// {1}{R} and its madness cost is {X}{R}, and that X is also the damage
// it divides. `altCost` is the key of the offer being claimed.
//
// #3232 adds the fifth: a kicker whose cost is {X} (Thieving Skydiver,
// Verdeloth the Ancient). The X belongs to the kicker, so it is asked
// only once the kicker is among the optional costs the player chose
// (`optionalCosts`, the indices already on the cast's choices); an
// unkicked cast announces no X.
export function hasXCost(card: CardView, altCost?: string, optionalCosts?: number[]): boolean {
  if (card.tap_cost?.demands_x) return true;
  if (card.additional_cost?.demands_x) return true;
  if ((alternativeCostByKey(card, altCost)?.mana_cost ?? "").includes("{X}")) return true;
  if (chosenOptionalCosts(card, optionalCosts).some((o) => (o.mana_cost ?? "").includes("{X}")))
    return true;
  return (card.mana_cost ?? "").includes("{X}");
}

// chosenOptionalCosts resolves the optional-cost indices a cast has
// announced so far to their offers, skipping an index the card does not
// have.
function chosenOptionalCosts(card: CardView, indices: number[] | undefined): OptionalCostView[] {
  const offers = optionalCostsOf(card);
  const out: OptionalCostView[] = [];
  for (const index of indices ?? []) {
    const offer = offers.find((o) => o.index === index);
    if (offer) out.push(offer);
  }
  return out;
}

// optionalCostXFloor is the smallest X the chosen optional costs allow
// (#3232): the largest `x_min` among them, 0 when none names a floor.
// The server refuses a kicked cast under it, so the X picker opens at
// it rather than letting the player type a value the cast rejects.
export function optionalCostXFloor(card: CardView, optionalCosts: number[] | undefined): number {
  return chosenOptionalCosts(card, optionalCosts).reduce((m, o) => Math.max(m, o.x_min ?? 0), 0);
}

// xCeilingFor is the largest X the viewer may announce for this cast
// (CR 601.2b), or undefined when nothing bounds it. A claimed offer that
// carries its own ceiling (#3225: "as though it had flash if X is 3 or
// less") wins — the server already folded the card's printed `x_max`
// into it — and otherwise the card's own `x_max` applies. The cast for
// the printed cost, which claims no offer, is therefore not bound by an
// offer it is not using.
export function xCeilingFor(card: CardView, altCost: string | undefined): number | undefined {
  const ceiling = alternativeCostByKey(card, altCost)?.x_max ?? card.x_max;
  // #1948: "exile X cards from your graveyard" — X is what the graveyard
  // pays, the same number the server refuses above.
  const ac = card.additional_cost;
  if (ac?.exile_cost_x) {
    const exileMax = ac.exile_cost_x_max ?? 0;
    return ceiling === undefined ? exileMax : Math.min(ceiling, exileMax);
  }
  return ceiling;
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
  // #1659: xValue rides through to stepsFor so a divide-from-X clause
  // (Katilda's activation-cost X, or any future X-cost ability with a
  // "divided as you choose" clause) resolves against the announced X
  // rather than silently landing on 0 — the two call sites used to
  // drop it on the floor, which is invisible until a card exercises
  // both mechanics at once.
  // ADR 0109 §9: the counters this payment removes, for a clause
  // bounded by them (Simic Manipulator). Known here: the counter picker
  // ran before the targeting walk, as CR 602.2b orders it.
  const choices: CastChoices = {};
  if (xValue !== undefined) choices.xValue = xValue;
  if (counter) choices.countersRemoved = countersRemovedOf(counter, ability.counter_cost_n);
  const steps =
    modes && modes.length > 0
      ? abilityModeSteps(ability, modes, xValue, choices.countersRemoved)
      : stepsFor(
          (ability.target_mode || "any") as TargetingMode,
          ability.legal_targets,
          ability.clauses,
          0,
          Object.keys(choices).length > 0 ? choices : undefined,
        );
  // #2055: nothing for the controller to pick (Nova Pentacle's only
  // clause is an opponent's choice); the caller activates at once.
  if (steps.length === 0) {
    targeting.set(null);
    return;
  }
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
// ModeSpecView. `xValue` is the ability's announced X (#1659) — see
// the comment at its call site in beginForAbility.
export function abilityModeSteps(
  ability: ActivatedAbilityView,
  modes: number[],
  xValue?: number,
  countersRemoved?: number,
): TargetStep[] {
  const options = ability.modes?.options ?? [];
  const out: TargetStep[] = [];
  const choices =
    xValue !== undefined || countersRemoved !== undefined ? { xValue, countersRemoved } : undefined;
  modes.forEach((optionIndex, occurrence) => {
    const option = options[optionIndex];
    if (!option?.legal_targets && !option?.clauses?.length) return;
    const mode = (option.target_mode || "any") as TargetingMode;
    for (const st of stepsFor(mode, option.legal_targets, option.clauses, occurrence, choices)) {
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
  if (t.legal) return t.legal.cards.has(instanceID) && !breaksSetRule(t, instanceID);
  return (
    isTargetingCreature(t.mode) ||
    isTargetingStack(t.mode) ||
    isTargetingGraveyard(t.mode) ||
    isTargetingExile(t.mode)
  );
}

export function isLegalPlayerTarget(t: TargetingState, playerID: string): boolean {
  if (t.legal) return t.legal.players.has(playerID);
  return isTargetingPlayer(t.mode);
}

// breaksSetRule reports whether picking `id` would give the current
// step two picks that share its set rule's key (#1559) — "each have a
// different mana value" with a 2-drop already picked, for another
// 2-drop. A card already picked never breaks the rule against itself,
// so it stays clickable to un-pick.
//
// #1807: under a sameness rule it is the other way round — picking
// `id` breaks the rule when an earlier pick's key DIFFERS from its
// own ("from a single graveyard", with a card from another graveyard
// already picked). With no pick yet every candidate is open.
export function breaksSetRule(t: TargetingState, id: string): boolean {
  return breaksDifferent(t, id) || breaksSame(t, id);
}

function breaksDifferent(t: TargetingState, id: string): boolean {
  const rule = t.different;
  if (!rule) return false;
  const key = rule.keys[id];
  if (key === undefined) return false;
  return t.picked.some((p) => p.id !== id && rule.keys[p.id] === key);
}

function breaksSame(t: TargetingState, id: string): boolean {
  const rule = t.same;
  if (!rule) return false;
  const key = rule.keys[id];
  if (key === undefined) return false;
  return t.picked.some((p) => {
    if (p.id === id) return false;
    const other = rule.keys[p.id];
    return other !== undefined && other !== key;
  });
}

// lonePick is the target the current step leaves no choice about
// (#3264), or null when the player is asked as before. It is chosen
// for them only when:
//
//   - the step takes exactly one target and exactly one is legal;
//   - the prompt is a trigger's pick_target (CR 603.3d: a required
//     target with one legal choice is that choice) or an activated
//     ability's walk, which the player started on purpose;
//   - nothing about the pick is still to be read: no price per target
//     (the banner shows an equip's price before its one click).
//
// A spell's targets always wait for the click, so a misclicked removal
// spell whose only legal target is the caster's own creature can still
// be cancelled.
export function lonePick(t: TargetingState): TargetRef | null {
  if (!t.legal || t.min !== 1 || t.max !== 1 || t.picked.length > 0) return null;
  if (t.choiceKind !== "pick_target" && !t.ability) return null;
  if (t.ability?.prices && Object.keys(t.ability.prices).length > 0) return null;
  const players = [...t.legal.players];
  const cards = [...t.legal.cards];
  if (players.length + cards.length !== 1) return null;
  return players.length === 1
    ? { kind: "player", id: players[0]! }
    : { kind: "card", id: cards[0]! };
}

// ---- One target for a repeated trigger (#3265) ----
//
// When one ability triggers several times at once (three creatures
// enter under the same permanent, or a doubler adds an instance), each
// instance queues its own pick_target (CR 603.3d), all at the same time.
// The player may pick once for all of them: the first answer is kept as
// the shared target and every later sibling whose legal set still holds
// it is answered with it. Each instance is still its own ability on the
// stack with its own target; this only fills in the choice.
//
// The player chooses how far the kept target reaches (#3817, ADR 0127's
// 2026-10-15 amendment: the owner chose to offer both):
//
//   - "source": the prompts of the same chooser from the same source
//     with the same clause wording — the same ability on the same
//     permanent, doubled instances included (#3265's scope, the
//     default);
//   - "card": the prompts of the same player for the same card and
//     ability, from every copy of it (twenty Blood Artists and two
//     deaths are forty prompts, one click). The server's
//     `shared_target_key` is that identity; a prompt without one is
//     only ever shared within its source. Never across players and
//     never across abilities: the key carries both;
//   - "off": every instance asks.

export type ShareScope = "source" | "card" | "off";

export interface SharedTarget {
  scope: "source" | "card";
  source: string;
  reason: string;
  /** For the card scope: the player-qualified `sharedKey` it covers. */
  key?: string;
  ref: TargetRef;
}

/** The target kept for the rest of a repeated trigger's instances. */
export const sharedTarget: Writable<SharedTarget | null> = guardedWritable(null, "sharedTarget");

/** The dock's choice of how far one pick reaches; this source's repeats by default. */
export const shareTargetScope: Writable<ShareScope> = guardedWritable<ShareScope>(
  "source",
  "shareTargetScope",
);

/**
 * effectiveShareScope is the scope that applies to one prompt: the card
 * scope falls back to the source scope where it reaches no further (one
 * source, or no card identity), so choosing "all of them" never turns
 * sharing off.
 */
export function effectiveShareScope(scope: ShareScope, cardReaches: boolean): ShareScope {
  return scope === "card" && !cardReaches ? "source" : scope;
}

/** sharedKeyOf is a prompt's card-and-ability identity for its player, or "" for none. */
export function sharedKeyOf(c: PendingChoiceView): string {
  if (!c.shared_target_key) return "";
  return `${c.on_behalf_of ?? c.chooser}\u0000${c.shared_target_key}`;
}

function isRepeatable(t: TargetingState): boolean {
  return t.choiceKind === "pick_target" && t.max === 1 && !!t.choiceID;
}

function siblingsOf(
  t: TargetingState,
  view: GameView | null | undefined,
  same: (mine: PendingChoiceView, c: PendingChoiceView) => boolean,
): PendingChoiceView[] {
  if (!isRepeatable(t) || !view) return [];
  const choices = view.pending_choices ?? [];
  const mine = choices.find((c) => c.id === t.choiceID);
  if (!mine) return [];
  return choices.filter(
    (c) =>
      c.id !== mine.id &&
      c.kind === "pick_target" &&
      c.chooser === mine.chooser &&
      c.reason === mine.reason &&
      same(mine, c),
  );
}

/**
 * repeatsOf counts the OTHER pick_target prompts waiting on the same
 * chooser for the same ability on the same source as `t`'s, or 0 when
 * `t` is not a one-target trigger prompt.
 */
export function repeatsOf(t: TargetingState, view: GameView | null | undefined): number {
  return siblingsOf(t, view, (mine, c) => c.source === mine.source).length;
}

/**
 * cardRepeats is the card scope's reach from `t`: how many OTHER prompts
 * of the same player wait for the same card and ability, and from how
 * many sources in all (this one included). `sources` under 2 means the
 * card scope would reach no further than the source scope.
 */
export function cardRepeats(
  t: TargetingState,
  view: GameView | null | undefined,
): { count: number; sources: number } {
  const key = t.sharedKey;
  if (!key) return { count: 0, sources: t.choiceID ? 1 : 0 };
  const others = siblingsOf(t, view, (_mine, c) => sharedKeyOf(c) === key);
  const sources = new Set([t.card.instance_id, ...others.map((c) => c.source ?? "")]);
  return { count: others.length, sources: sources.size };
}

/** sharedFrom is what to keep after answering `t` with `ref` under `scope`. */
export function sharedFrom(
  t: TargetingState,
  ref: TargetRef,
  scope: "source" | "card" = "source",
): SharedTarget | null {
  if (!isRepeatable(t)) return null;
  if (scope === "card" && !t.sharedKey) return null;
  return {
    scope,
    source: t.card.instance_id,
    reason: t.label ?? "",
    ...(scope === "card" ? { key: t.sharedKey } : {}),
    ref: { kind: ref.kind, id: ref.id },
  };
}

/**
 * sharedAfterAnswer is what to keep after `t` is answered with
 * `targets` under the dock's `scope`: the card scope where it reaches
 * other sources, else the source scope where this source repeats, else
 * nothing (a lone instance, a multi-target pick, or sharing off).
 */
export function sharedAfterAnswer(
  t: TargetingState,
  targets: readonly TargetRef[],
  scope: ShareScope,
  view: GameView | null | undefined,
): SharedTarget | null {
  if (targets.length !== 1 || scope === "off") return null;
  const card = cardRepeats(t, view);
  const eff = effectiveShareScope(scope, card.sources > 1 && card.count > 0);
  if (eff === "card") return sharedFrom(t, targets[0]!, "card");
  if (eff === "source" && repeatsOf(t, view) > 0) return sharedFrom(t, targets[0]!, "source");
  return null;
}

function covers(shared: SharedTarget, source: string, reason: string, key: string): boolean {
  if (reason !== shared.reason) return false;
  if (shared.scope === "card") return !!shared.key && key === shared.key;
  return source === shared.source;
}

/**
 * sharedPick is the kept target, when `t` is within its reach and the
 * target is still legal for it; otherwise null, and the player is asked
 * as usual.
 */
export function sharedPick(t: TargetingState, shared: SharedTarget | null): TargetRef | null {
  if (!shared || !isRepeatable(t) || !t.legal || t.picked.length > 0) return null;
  if (!covers(shared, t.card.instance_id, t.label ?? "", t.sharedKey ?? "")) return null;
  const legal = shared.ref.kind === "player" ? t.legal.players : t.legal.cards;
  return legal.has(shared.ref.id) ? { kind: shared.ref.kind, id: shared.ref.id } : null;
}

/**
 * sharedStillWanted reports whether any prompt in the view could still
 * use the kept target; once none can, the caller forgets it.
 */
export function sharedStillWanted(
  shared: SharedTarget | null,
  view: GameView | null | undefined,
): boolean {
  if (!shared || !view) return false;
  return (view.pending_choices ?? []).some(
    (c) =>
      c.kind === "pick_target" && covers(shared, c.source ?? "", c.reason ?? "", sharedKeyOf(c)),
  );
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
  if (isBoardPickedCardSetKind(choice.kind)) {
    targeting.set(cardSetPickState(choice, card));
    return;
  }
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
      sharedKey: sharedKeyOf(choice) || undefined,
    }),
  );
}

// cardSetPickState is #2394's board pick for a card-set choice over
// permanents on the battlefield (choose_cards, untap_choice): the
// candidates are the legal set, choose_min / choose_max the bounds, and
// the walk is one step. Nothing here is a target (the card says "untap
// up to five lands", not "target lands"); the targeting store is only
// the board's machinery for highlighting and clicking permanents.
export function cardSetPickState(choice: PendingChoiceView, card: CardView): TargetingState {
  const ids = (choice.options ?? []).map((o) => o.instance_id);
  const max = choice.choose_max ?? ids.length;
  const step: TargetStep = {
    mode: "permanent",
    legal: { players: new Set(), cards: new Set(ids) },
    label: choice.reason,
    min: choice.choose_min ?? 0,
    // The targeting store reads max 0 as unbounded; a card set is
    // never bigger than its candidates.
    max: max > 0 ? max : ids.length,
    modeIndex: 0,
    slot: 0,
    distinct: false,
  };
  return openWalk(card, [step], {
    choiceID: choice.id,
    choiceKind: choice.kind,
    label: choice.reason,
  });
}

// isCardSetPick reports whether a targeting state answers a card-set
// choice rather than a target prompt.
export function isCardSetPick(t: TargetingState): boolean {
  return !!t.choiceID && isBoardPickedCardSetKind(t.choiceKind ?? "");
}

// choiceAnswer is the resolve_choice payload for a board-answered
// choice. A target prompt sends `targets`; a card-set pick sends the
// {choice_id, card_ids} answer the modal's grid sends for the same
// kind, so the server sees no difference between the two surfaces.
export function choiceAnswer(t: TargetingState, targets: TargetRef[]): Record<string, unknown> {
  if (isCardSetPick(t)) {
    return { choice_id: t.choiceID, card_ids: targets.map((r) => r.id) };
  }
  return { choice_id: t.choiceID, targets };
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
    case "card_in_exile":
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

// #3259: a card in exile is picked in the zone browser's exile tab, the
// way a graveyard card is picked in its graveyard tab. Like the others
// this is the free-form fallback; a structured clause's legal set, which
// already leaves out a face-down card the viewer has not seen, decides.
export function isTargetingExile(mode: TargetingMode): boolean {
  return mode === "card_in_exile";
}
