// contextMenu.logic — pure menu assembly behind CardContextMenu.
// Lives outside the component (the same split zoneBrowser.logic.ts
// uses) so vitest can pin the option set and the wire payloads for
// every zone without rendering Svelte: the client's test runner is
// node-only, with no jsdom.
//
// Design contract for #170: the menu never invents an action type.
// Every item resolves to one of the verbs the server already
// implements in server/internal/actions/actions.go — move_card,
// add_counter, tap / untap, mark_damage, declare_attacker,
// declare_blocker, clear_combat, set_goaded, sacrifice_permanent.
// That keeps the "escape hatch" honest: anything the menu offers is
// something the engine could already have been driven to do from
// gamecli, just without a discoverable surface.
//
// The same MenuItem tree can render a future forced prompt that needs
// a per-card surface (#164's commander-zone prompt, once named here,
// shipped in #171 as an optional_replacement choice) — an item is a label
// plus either an action, a nested list, or a prompt marker, so a
// caller that wants a two-option "yes / no" menu builds one section
// with two action items and reuses the component verbatim.

import {
  attackAllLabel,
  attackAllParams,
  attackAllTaxLabel,
  attackTaxOn,
  planAttackAll,
  seatLabel,
} from "./attackAll";
import { grantedFromLabel } from "./abilityRef";
import {
  attackersDefendedBy,
  attackTargetHint,
  blockedAttackersOf,
  blockerHasRoom,
  permanentAttackTargets,
} from "./attackTargets";
import { isPlaneswalker } from "./cardTypes";
import { counterCostBlocked } from "./counterCost";
import {
  NO_LEGAL_ACTIONS,
  digestRefusesRow,
  digestRefusesSpecial,
  readyFirst,
  type LegalActions,
} from "./legalActions";
import type {
  ActionType,
  ActivatedAbilityView,
  CardView,
  GameView,
  ManaAbilityView,
} from "./protocol";
import { sacrificeRangeShortfall } from "./sacrificeCost";
import { targetPriceRange } from "./targetPrices";
import {
  canActivateLoyalty,
  canActivateSorcerySpeedAbility,
  canPayLoyaltyCost,
  hasSatisfiableTargets,
  loyaltyOf,
} from "./timing";
import {
  COUNTER_CHARGE,
  COUNTER_DEFENSE,
  COUNTER_LORE,
  COUNTER_LOYALTY,
  COUNTER_MINUS_ONE,
  COUNTER_PLUS_ONE,
  COUNTER_SHIELD,
  COUNTER_STUN,
} from "./counterTypes";

// MenuZone is every zone a card can be right-clicked in. Mirrors
// game.ZoneKind server-side.
export type MenuZone =
  | "battlefield"
  | "hand"
  | "graveyard"
  | "exile"
  | "command"
  | "library"
  | "stack";

export const ZONE_LABELS: Record<MenuZone, string> = {
  battlefield: "battlefield",
  hand: "hand",
  graveyard: "graveyard",
  exile: "exile",
  command: "command zone",
  library: "library",
  stack: "stack",
};

// COMBAT_STEPS gates the attack / block items. Outside combat those
// declarations are meaningless (and the server would have nothing to
// clear), so the section is dropped rather than shown disabled.
export const COMBAT_STEPS: ReadonlySet<string> = new Set([
  "begin_combat",
  "declare_attackers",
  "declare_blockers",
  "first_strike_damage",
  "combat_damage",
  "end_combat",
]);

// QUICK_COUNTERS get their own add / remove rows at the top level.
// Everything else is one level down under "Other counters".
export const QUICK_COUNTERS: readonly string[] = [
  COUNTER_PLUS_ONE,
  COUNTER_MINUS_ONE,
  COUNTER_LOYALTY,
];

// CARD_COUNTER_NAMES is every card-level counter the pip renderer
// knows how to style. Player-level counters (poison, energy,
// experience, rad) are deliberately absent — this is a card menu,
// and those live on the player panel.
export const CARD_COUNTER_NAMES: readonly string[] = [
  COUNTER_PLUS_ONE,
  COUNTER_MINUS_ONE,
  COUNTER_LOYALTY,
  COUNTER_DEFENSE,
  COUNTER_CHARGE,
  COUNTER_STUN,
  COUNTER_SHIELD,
  COUNTER_LORE,
];

// MenuZoneRef mirrors the `src` / `dst` shape of a move_card action.
// Owner is omitted for the shared zones (battlefield, exile, stack)
// exactly as the server's zoneRefWire expects.
export interface MenuZoneRef {
  kind: string;
  owner?: string;
}

// MenuAction is a ready-to-send wire action. `player` is the seat the
// action is attributed to; card-scoped verbs leave it undefined and
// let the connection's own principal stand.
export interface MenuAction {
  type: ActionType;
  params?: Record<string, unknown>;
  player?: string;
}

// MenuPrompt marks an item that needs one more piece of input before
// it can fire. The component renders a tiny inline form for it.
export type MenuPrompt = "custom_counter" | "mark_damage";

// MenuActivate marks an item that hands off to the board's existing
// ability flows instead of sending a bare action. Activating an
// ability can open a sacrifice picker or enter targeting, and both
// of those modals are board-wide, so the menu forwards rather than
// firing itself.
export interface MenuActivate {
  kind: "mana" | "ability";
  index: number;
  // #1443: a mana ability's colours, named up front by the anchored
  // picker so the server queues no mana_pick. Absent from the menu's
  // own rows, which keep the two-step activation.
  colors?: string[];
}

export interface MenuItem {
  // Stable key for the {#each} block and for tests.
  id: string;
  label: string;
  // Tooltip / secondary copy.
  hint?: string;
  // Red styling for destructive overrides.
  danger?: boolean;
  // Rendered but not clickable (e.g. "Remove +1/+1" with none on).
  disabled?: boolean;
  // ADR 0105 (#1789): the server would accept this row right now (its
  // ref is in the legal-action digest). Drawn with the ready accent,
  // and sorted first in its section. Never set on a disabled row.
  ready?: boolean;
  action?: MenuAction;
  prompt?: MenuPrompt;
  activate?: MenuActivate;
  // Leave the menu open after firing. Set on the incremental rows
  // (counters, damage) because "add three +1/+1 counters" is three
  // clicks and re-opening the menu between each would be hostile.
  repeat?: boolean;
  // Non-empty for a submenu row.
  items?: MenuItem[];
}

export interface MenuSection {
  id: string;
  label?: string;
  items: MenuItem[];
}

export interface CardLocation {
  zone: MenuZone;
  // The seat whose zone holds the card. For the shared zones this is
  // the card's own owner, which is what a "back to your hand" move
  // needs anyway.
  ownerID: string;
}

// locateCard finds which zone a card currently sits in. The context
// menu is opened from a Card component that doesn't know its own
// zone, so the lookup happens here against the authoritative
// snapshot rather than being prop-drilled through five components.
export function locateCard(view: GameView, instanceID: string): CardLocation | null {
  for (const c of view.battlefield?.cards ?? []) {
    if (c.instance_id === instanceID) return { zone: "battlefield", ownerID: c.owner };
  }
  for (const c of view.stack?.cards ?? []) {
    if (c.instance_id === instanceID) return { zone: "stack", ownerID: c.owner };
  }
  for (const c of view.exile?.cards ?? []) {
    if (c.instance_id === instanceID) return { zone: "exile", ownerID: c.owner };
  }
  for (const s of view.seats) {
    for (const c of s.hand?.cards ?? []) {
      if (c.instance_id === instanceID) return { zone: "hand", ownerID: s.id };
    }
    for (const c of s.graveyard?.cards ?? []) {
      if (c.instance_id === instanceID) return { zone: "graveyard", ownerID: s.id };
    }
    for (const c of s.command?.cards ?? []) {
      if (c.instance_id === instanceID) return { zone: "command", ownerID: s.id };
    }
    for (const c of s.library?.cards ?? []) {
      if (c.instance_id === instanceID) return { zone: "library", ownerID: s.id };
    }
  }
  return null;
}

// findCard resolves an instance ID against the current snapshot. The
// menu captures the CardView it was opened on, but re-resolving on
// every snapshot keeps counter totals, tapped state and damage live
// while the menu stays open.
export function findCard(view: GameView, instanceID: string): CardView | null {
  const lists: CardView[][] = [
    view.battlefield?.cards ?? [],
    view.stack?.cards ?? [],
    view.exile?.cards ?? [],
  ];
  for (const s of view.seats) {
    lists.push(s.hand?.cards ?? []);
    lists.push(s.graveyard?.cards ?? []);
    lists.push(s.command?.cards ?? []);
    lists.push(s.library?.cards ?? []);
  }
  for (const list of lists) {
    for (const c of list) {
      if (c.instance_id === instanceID) return c;
    }
  }
  return null;
}

// canOverride mirrors the server's requireCardController gate: a
// seated player may drive their own cards, an admin may drive
// anyone's. Showing items the server would bounce is worse than
// showing none, so a non-controller gets an empty menu.
export function canOverride(card: CardView, viewerID: string | null, isAdmin: boolean): boolean {
  if (isAdmin) return true;
  if (!viewerID) return false;
  return card.controller === viewerID;
}

// anyPlayerRows is a permanent's "Any player may activate this
// ability" rows (CR 602.2, ADR 0106 §1). Only `activated_abilities`:
// the declaration is refused on a mana ability, and no printed
// any-player ability functions from a hidden zone.
export function anyPlayerRows(card: CardView): ActivatedAbilityView[] {
  return (card.activated_abilities ?? []).filter((a) => a.any_player === true);
}

// mayActivateAcross is CR 602.2's exception for the client: a SEATED
// viewer who does not control the permanent may still open it for its
// any-player rows (ADR 0106 §1 decision 6). A spectator never may: a
// spectator activates nothing. The viewer's own permanent answers
// false, because its controller already has the whole menu.
export function mayActivateAcross(card: CardView, viewerID: string | null): boolean {
  if (!viewerID) return false;
  if ((card.controller || card.owner) === viewerID) return false;
  return anyPlayerRows(card).length > 0;
}

// menuAbilityRows is the activated-ability rows the ability popover
// lists for this viewer: every row for the permanent's controller, and
// only the any-player rows for anyone else (ADR 0106 §1 decision 6).
// A row a non-controller cannot activate is one the server would
// refuse, and a row that opens a refusal is worse than none.
// `viewerID` undefined means the caller has no viewer in scope (a hand
// card, the zone browser), and the rows pass through unchanged. A
// spectator (null) gets none of a permanent's rows.
export function menuAbilityRows(
  card: CardView,
  viewerID: string | null | undefined,
): ActivatedAbilityView[] {
  const rows = card.activated_abilities ?? card.zone_abilities ?? [];
  if (viewerID === undefined || !card.activated_abilities) return rows;
  if (viewerID === null) return [];
  if ((card.controller || card.owner) === viewerID) return rows;
  return rows.filter((a) => a.any_player === true);
}

// acrossFor says whether `card`'s activated rows are being opened by a
// seated viewer who does not control it (ADR 0106 §1 decision 6): the
// popover then lists only its any-player rows, and greys any the exact
// digest leaves out. Card and the click rule both ask this, so they
// agree on which rows a viewer sees.
export function acrossFor(card: CardView, viewerID: string | null | undefined): boolean {
  return !!viewerID && !!card.activated_abilities && (card.controller || card.owner) !== viewerID;
}

// menuManaRows is menuAbilityRows for the CR 605 list: a permanent's
// mana rows are its controller's alone (a mana ability is never an
// any-player row, ADR 0106 §1), so a viewer who does not control it
// gets none — an opponent's Aura drawn on the viewer's own creature
// included. `viewerID` undefined passes the rows through (a hand card,
// whose list is `zone_mana_abilities`, CR 113.6).
export function menuManaRows(
  card: CardView,
  viewerID: string | null | undefined,
): ManaAbilityView[] {
  const rows = card.mana_abilities ?? card.zone_mana_abilities ?? [];
  if (viewerID === undefined) return rows;
  if (viewerID === null) return [];
  return (card.controller || card.owner) === viewerID ? rows : [];
}

// BattlefieldClickIntent is what a plain left-click on a battlefield
// permanent does once the intercepts (targeting, the combat selects,
// an attack on a listed target, a block) have passed (ADR 0117 §1, as
// amended on 2026-10-04 by #2201).
//
//	"activate" — the card has exactly one usable row, and it is not a
//	             mana row: activate it, through the path the popover's
//	             row takes (#2201)
//	"popover"  — open the card's light ability popover at the card: it
//	             has two or more usable rows, one of them not a mana
//	             row (owner answer 2)
//	"mana"     — tap it FOR mana (#1438): only mana rows are usable
//	"tap"      — Alt-click: a raw tap / untap, the sandbox escape hatch
//	"none"     — nothing on the card can be used right now
export type BattlefieldClickIntent = "activate" | "popover" | "mana" | "tap" | "none";

// LoneAbilityRow is the one usable row an "activate" click acts on, in
// the shape the popover's row hands it on: an activated row (own,
// granted, loyalty or any-player) by its index, for the board's
// activation flow (costs, X, modes, targets); a special action or a
// manual loyalty row by the action its row sends.
export type LoneAbilityRow =
  | { kind: "activated"; index: number }
  | { kind: "special"; action: MenuAction }
  | { kind: "loyalty"; action: MenuAction };

// BattlefieldClickPlan is the click rule's answer: the intent, and for
// "activate" the row it activates.
export type BattlefieldClickPlan =
  | { intent: "activate"; row: LoneAbilityRow }
  | { intent: Exclude<BattlefieldClickIntent, "activate"> };

// BattlefieldClickOptions carries what the click rule cannot read off
// the card. The row inputs are the ones the popover is drawn with
// (PlayerPanel hands both the same values), so the click and the
// popover's greying are one judgement (ADR 0117 §2).
export interface BattlefieldClickOptions {
  // This panel wires mana activations: the popover lists mana rows and
  // the click may tap for mana. Only the viewer's own panel does.
  manaClick?: boolean;
  // This panel wires the special-action and manual-loyalty sender.
  special?: boolean;
  // Alt-click: always "just turn it sideways", where canOverride holds.
  rawTap?: boolean;
  view?: GameView | null;
  payerLife?: number;
  // The panel's words for a window the server shut (`timing_closed`).
  timingWords?: string;
  // The frame's FULL legal-action lookup (the popover's `legalGate`).
  legalGate?: LegalActions;
}

// battlefieldClickPlan is ADR 0117's click rule. The table is
// automated now, so a left-click does what the card does, and a card
// with nothing to do does nothing instead of turning sideways.
//
// History it replaces: a click used to tap, with exceptions grown one
// at a time — a planeswalker opened its menu (#329, Teferi tapping
// when clicked for his abilities), a utility land opened its menu
// (#368), an untapped mana source tapped for mana (#1438), a permanent
// with a granted ability opened its menu (ADR 0093 Decision 8).
//
// The rule:
//   1. Alt-click raw-taps wherever canOverride holds (an admin's too).
//   2. The usable rows are worked out with the one predicate
//      (abilityRowBlocked, through abilityPopoverModel) the popover
//      greys with, and counted the way the popover lists them: special
//      actions, mana rows, activated rows (own, granted, loyalty,
//      any-player) and manual loyalty rows. The Sandbox Tap / Untap
//      row never counts. For a permanent the viewer does not control,
//      only its any-player rows count (CR 602.2, ADR 0106 §1). Who the
//      card belongs to is the viewer's seat, not canOverride, so an
//      admin's plain click on another seat's permanent no longer taps.
//   3. Exactly one usable row, and it is not a mana row: activate it
//      (#2201, the 2026-10-04 amendment). A fetch land fetches,
//      Prodigal Sorcerer starts its targeting, an Equipment's lone
//      Equip starts its creature choice. Its costs are paid as from the
//      popover, at once; the table's Undo covers a misclick (#2201
//      owner answer 2).
//   4. Two or more usable rows, one of them not a mana row: the
//      popover (ADR 0117 owner answer 2).
//   5. Otherwise any usable mana row: the mana path (manaClickPlan),
//      which activates a lone mana ability at once and never chooses
//      silently between two.
//   6. Otherwise nothing.
export function battlefieldClickPlan(
  card: CardView,
  viewerID: string | null,
  isAdmin: boolean,
  opts: BattlefieldClickOptions = {},
): BattlefieldClickPlan {
  if (opts.rawTap && canOverride(card, viewerID, isAdmin)) return { intent: "tap" };
  if (!viewerID) return { intent: "none" };
  const model = abilityPopoverModel({
    card,
    viewerID,
    view: opts.view,
    payerLife: opts.payerLife,
    timingWords: opts.timingWords,
    legalGate: opts.legalGate,
    mana: !!opts.manaClick,
    activated: true,
    special: !!opts.special,
  });
  const nonMana = usableNonManaRows(model);
  const mana = model.mana.filter((r) => !r.blocked).length;
  if (nonMana.length === 1 && mana === 0) return { intent: "activate", row: nonMana[0] };
  if (nonMana.length > 0) return { intent: "popover" };
  if (mana > 0) return { intent: "mana" };
  return { intent: "none" };
}

// battlefieldClickIntent is battlefieldClickPlan's intent alone.
export function battlefieldClickIntent(
  card: CardView,
  viewerID: string | null,
  isAdmin: boolean,
  opts: BattlefieldClickOptions = {},
): BattlefieldClickIntent {
  return battlefieldClickPlan(card, viewerID, isAdmin, opts).intent;
}

// usableNonManaRows: the popover's rows other than its mana rows that
// can be used right now, in the popover's order (special actions,
// activated rows, manual loyalty rows). The sandbox Tap / Untap row is
// never one: it is on every permanent the viewer controls (ADR 0117
// §3).
export function usableNonManaRows(model: AbilityPopoverModel): LoneAbilityRow[] {
  const out: LoneAbilityRow[] = [];
  for (const i of model.special) {
    if (!i.disabled && i.action) out.push({ kind: "special", action: i.action });
  }
  for (const r of model.activated) {
    if (!r.blocked) out.push({ kind: "activated", index: r.a.index });
  }
  for (const i of model.loyalty) {
    if (!i.disabled && i.action) out.push({ kind: "loyalty", action: i.action });
  }
  return out;
}

// popoverHasUsableNonMana: a row in the popover other than a mana row
// can be used right now.
export function popoverHasUsableNonMana(model: AbilityPopoverModel): boolean {
  return usableNonManaRows(model).length > 0;
}

export function zoneRefFor(zone: MenuZone, ownerID: string): MenuZoneRef {
  if (zone === "battlefield" || zone === "exile" || zone === "stack") return { kind: zone };
  return { kind: zone, owner: ownerID };
}

export interface MoveDest {
  id: string;
  label: string;
  zone: MenuZone;
  // Seat the card at the BOTTOM of the destination instead of the
  // top. Only meaningful for the library.
  bottom?: boolean;
}

// MOVE_DESTINATIONS is the full destination list, in the order the
// submenu renders them. The stack is deliberately not a destination:
// a card pushed into the stack zone without a matching StackMeta
// entry would render as an item nobody can resolve. Moving a card
// OFF the stack (a manual fizzle) is supported.
export const MOVE_DESTINATIONS: readonly MoveDest[] = [
  { id: "hand", label: "Hand", zone: "hand" },
  { id: "battlefield", label: "Battlefield", zone: "battlefield" },
  { id: "graveyard", label: "Graveyard", zone: "graveyard" },
  { id: "exile", label: "Exile", zone: "exile" },
  { id: "library-top", label: "Library (top)", zone: "library" },
  { id: "library-bottom", label: "Library (bottom)", zone: "library", bottom: true },
  { id: "command", label: "Command zone", zone: "command" },
];

// moveDestinations drops the zone the card is already in — a
// same-zone move is a server-side no-op and the menu shouldn't
// pretend otherwise.
export function moveDestinations(from: MenuZone): MoveDest[] {
  return MOVE_DESTINATIONS.filter((d) => d.zone !== from);
}

// buildMoveAction assembles a move_card payload. The destination's
// owner is the card's OWNER, not its controller: a stolen creature
// dies to its owner's graveyard (CR 400.3), and the same holds for
// every manual override here.
export function buildMoveAction(
  card: CardView,
  from: CardLocation,
  dest: MoveDest,
  asCommander = false,
): MenuAction {
  const params: Record<string, unknown> = {
    src: zoneRefFor(from.zone, from.ownerID),
    dst: zoneRefFor(dest.zone, card.owner),
    instance_id: card.instance_id,
  };
  if (dest.bottom) params.to_bottom = true;
  if (asCommander) params.as_commander = true;
  return { type: "move_card", params };
}

export function counterAction(card: CardView, name: string, delta: number): MenuAction {
  return {
    type: "add_counter",
    params: { instance_id: card.instance_id, name, delta },
  };
}

export function damageAction(card: CardView, delta: number): MenuAction {
  return {
    type: "mark_damage",
    params: { instance_id: card.instance_id, delta },
  };
}

// AbilityCost is the cost-shaped subset shared by ManaAbilityView and
// ActivatedAbilityView, so one predicate covers both. Exported (#1695)
// so ManaAbilityMenu.svelte can type its own rows against it instead
// of keeping a second, hand-rolled copy of the same shape.
export interface AbilityCost {
  tap_cost?: boolean;
  // #1190: mana_cost is the PRINTED mana component and
  // charged_mana_cost what the engine actually charges right now,
  // after every CR 601.2f cost modifier on the battlefield — see
  // chargedManaCostNote below. Both mana and activated abilities
  // carry the pair under the same two names.
  mana_cost?: string;
  charged_mana_cost?: string;
  sacrifice_label?: string;
  // #747: min / max are the sacrifice count (sacrificeCost.ts).
  // #1213: and they may now differ — an open count ("one or more") is
  // min 1 with no ceiling, and count_from_x means the count IS the
  // announced X.
  sacrifice_options?: {
    players?: string[];
    cards?: string[];
    min?: number;
    max?: number;
    count_from_x?: boolean;
  };
  // #1213: a return-to-hand cost. The row is greyed when nothing the
  // clause admits is on the board, for the same reason a sacrifice
  // cost with no candidates greys one — CR 118.3 refuses the
  // activation, and finding out at the click is worse than seeing it
  // before.
  return_label?: string;
  return_options?: { players?: string[]; cards?: string[]; min?: number; max?: number };
  // #1600: an exile-a-permanent cost ("Exile a creature you control"),
  // greyed on the same terms as the return cost.
  exile_permanent_label?: string;
  exile_permanent_options?: { players?: string[]; cards?: string[]; min?: number; max?: number };
  // #759: a tap-another cost (station). Greyed on the same terms as
  // the return cost: fewer untapped creatures than the clause needs.
  tap_others_label?: string;
  tap_others_options?: { players?: string[]; cards?: string[]; min?: number; max?: number };
  // #1157: `min` is part of the clause and not decoration. "Up to one
  // target creature you control" is min 0, and a clause with min 0 is
  // satisfied by an empty candidate list — see hasSatisfiableTargets.
  legal_targets?: { players?: string[]; cards?: string[]; min?: number };
  // Present, at any value including 0, on a planeswalker's loyalty
  // ability. Mana abilities never carry it.
  loyalty_cost?: number;
  // #1690: a "Pay N life" cost component — Greed's printed one, and
  // since #1688 a computed one (War Room, Murderous Betrayal, priced
  // through the controller's board state). Carried by both mana and
  // activated abilities under the same wire name
  // (ActivatedAbilityView.LifeCost / ManaAbilityView.LifeCost).
  life_cost?: number;
  // S24: "Activate only as a sorcery" (CR 602.5d). Equip is the
  // catalog's first; a loyalty ability gets the same window from its
  // own arm below rather than from this flag.
  sorcery_speed?: boolean;
  // #1208: true when the engine will refuse this activation RIGHT NOW
  // for timing (CR 602.5d, CR 606.3), as modified by any per-player
  // statement on the board — The Wandering Emperor's "you may
  // activate her loyalty abilities any time you could cast an
  // instant", Leonin Shikari's "you may activate equip abilities any
  // time you could cast an instant".
  //
  // It is the ROW's verdict where sorcery_speed is the ability's
  // printed clause, and it is the one to grey on: the client no
  // longer derives the window for a catalogued ability. Absent means
  // the engine has no timing objection, which is every instant-speed
  // ability, always.
  timing_closed?: boolean;
  // #743: the ability's "Activate only if …" condition is false right
  // now. Carried by both mana and activated abilities.
  condition_unmet?: boolean;
  // #1181: an exhaust ability this permanent has already used.
  // #1183: and a MANA ability too — Loot, the Pathfinder's "Exhaust —
  // {G}, {T}: Add three mana of any one color". The server ships one
  // flag under one name for both ability kinds, which is why this
  // predicate needed no sibling.
  exhausted?: boolean;
  // #844: a "in your commander's color identity" mana ability with no
  // identity to narrow to. Mana abilities only.
  adds_no_mana?: boolean;
  // S27: a Vehicle's crew number and the creatures that could pay
  // it. Mana abilities never carry either.
  crew_cost?: number;
  crew_options?: { players?: string[]; cards?: string[] };
  // #625, then #789: the counter components — a "remove N counters"
  // cost's shape and what can pay it right now, and a cost that puts
  // one on. Mirrors ActivatedAbilityView in protocol.ts, and since
  // #789 ManaAbilityView carries the identical fields (Vivid Creek,
  // Ramos, Mage-Ring Network), which is exactly why this type is
  // structural: one predicate, both ability kinds.
  counter_cost_n?: number;
  counter_cost_kind?: string;
  counter_cost_self?: boolean;
  counter_cost_label?: string;
  counter_cost_among?: boolean;
  counter_cost_variable?: boolean;
  counter_cost_max?: number;
  counter_cost_options?: { card_id: string; kinds: { kind: string; count: number }[] }[];
  counter_cost_add?: number;
  counter_cost_add_kind?: string;
  counter_add_blocked?: boolean;
  // #1210: the printed clause of a board-wide "can't be activated"
  // static refusing this row (Cursed Totem). Both ability kinds carry
  // it under the one name; ADR 0117 §2 reads it.
  cant_activate?: string;
}

// ACTIVATION_CONDITION_UNMET is the hint on a row whose
// condition_unmet flag is set (#743). Exported so ManaAbilityMenu's
// popover says the same thing as the context menu.
export const ACTIVATION_CONDITION_UNMET = "activation condition not met";

// ABILITY_EXHAUSTED is the hint on a row the server marked exhausted
// (#1181): "Activate each exhaust ability only once", and this object
// already has. Its own string rather than ACTIVATION_CONDITION_UNMET
// because the two recover differently — a condition may hold again
// next turn, an exhaust only if the permanent becomes a new object.
export const ABILITY_EXHAUSTED = "already activated (exhaust)";

// ADDS_NO_MANA is the hint on a mana row the server marked
// adds_no_mana: it would add nothing right now. That was #844's case
// first (CR 903.4f: "any color in your commander's color identity"
// with no commander, or a colourless one), and since ADR 0117 §5 it is
// also a computed output that comes to nothing (a power-0 Vivi
// Ornitier, an Exotic Orchard with nothing to copy, CR 106.5 and
// 106.7). Generic on purpose: the row's own label already says what
// the output depends on, the way ACTIVATION_CONDITION_UNMET leaves the
// condition to the label. Exported for the same reason.
export const ADDS_NO_MANA = "adds no mana right now";

// chargedManaCostNote is the tooltip fragment for a row whose
// charged_mana_cost differs from its printed mana_cost (#1190) — a
// Boom Scholar-style discount reaching this permanent's ability.
// "" when there is nothing to say: no mana component at all, the
// server didn't price one (a redacted or pre-#1190 view — undefined,
// never a bare ""), or the two already agree, which is nearly every
// ability in the game. Exported so ManaAbilityMenu's chip and this
// menu's hint read the same words.
//
// `charged_mana_cost === undefined` and `charged_mana_cost === ""`
// are different answers and this reads them differently: undefined is
// "not priced" (mana_cost is what the row shows), and "" is "priced
// to nothing" — a real discount worth naming, not a value to treat as
// falsy. A plain `!a.charged_mana_cost` check would conflate the two
// and go silent on the most dramatic discount there is.
export function chargedManaCostNote(a: AbilityCost): string {
  if (!a.mana_cost || a.charged_mana_cost === undefined || a.charged_mana_cost === a.mana_cost) {
    return "";
  }
  return `printed cost ${a.mana_cost}`;
}

// chargedManaCostLabel is what a mana-cost chip displays: the charged
// cost when the server priced one, the printed cost when it did not
// (undefined — a redacted or pre-#1190 view), and "free" for the real,
// distinct case of a discount that emptied the component out
// completely (charged_mana_cost === ""). `a.charged_mana_cost ||
// a.mana_cost` alone would treat that empty string as falsy and show
// the STALE printed cost instead — the one discount worth naming most.
export function chargedManaCostLabel(a: AbilityCost): string {
  if (a.charged_mana_cost === undefined) return a.mana_cost ?? "";
  return a.charged_mana_cost || "free";
}

// ReturnOptionsShape is the part of a LegalTargetsView a return-to-hand
// cost ships (#1213). `min` and `max` are both the clause's count.
export interface ReturnOptionsShape {
  players?: string[];
  cards?: string[];
  min?: number;
  max?: number;
  count_from_x?: boolean;
}

// returnShortfall is the reason a return-to-hand cost can't be paid
// right now, or "" when it can — sacrificeShortfall one verb over.
// The clause's count is 1 on every printed card, so an option list
// shorter than `min` is the whole of "this cannot be paid" (CR 118.3).
//
// ONE definition, two menus (#1227). The right-click menu had it
// inline and the hand / zone-browser popover had nothing at all, which
// is exactly the row ninjutsu needs greyed: a ninjutsu ability is
// unpayable for the whole game except the declare-blockers window with
// an unblocked attacker on the board, so a row that never greys is a
// row that is almost always wrong.
export function returnShortfall(opts: ReturnOptionsShape | undefined, label?: string): string {
  if (!opts) return "";
  const have = opts.cards?.length ?? 0;
  if (have >= (opts.min ?? 1)) return "";
  return `nothing to return (${label ?? "a permanent you control"})`;
}

// exilePermanentShortfall is returnShortfall one destination over
// (#1600): "Exile a creature you control" with no creature of yours to
// exile cannot be paid (CR 118.3).
export function exilePermanentShortfall(
  opts: ReturnOptionsShape | undefined,
  label?: string,
): string {
  if (!opts) return "";
  const have = opts.cards?.length ?? 0;
  if (have >= (opts.min ?? 1)) return "";
  return `nothing to exile (${label ?? "a creature you control"})`;
}

// tapOthersShortfall is returnShortfall for a tap-another cost (#759):
// the reason it can't be paid right now, or "" when it can. The
// server's option list already excludes tapped creatures, other
// players' creatures and — for "another" — the source, so its length
// against `min` is the whole of CR 118.3.
export function tapOthersShortfall(opts: ReturnOptionsShape | undefined, label?: string): string {
  if (!opts) return "";
  const have = opts.cards?.length ?? 0;
  const need = opts.count_from_x ? 1 : (opts.min ?? 1);
  if (have >= need) return "";
  return `nothing to tap (${label ?? "another untapped creature you control"})`;
}

// NOT_ENOUGH_LIFE is the hint on a row whose life_cost is more than
// the paying player has (#1690). Exported so other surfaces that
// render the same cost (ManaAbilityMenu, seatSummary) can say the
// same thing.
export const NOT_ENOUGH_LIFE = "Not enough life";

// notEnoughLife is the reason a "Pay N life" cost can't be paid right
// now, or "" when it can. CR 119.4: paying life equal to your life
// total is legal, so the test is strictly greater, never "at least".
// `life` is undefined when the caller has no life total to check
// against (no snapshot in scope), in which case the row is judged
// unblocked and the server's own CR 119.4 check is the only gate —
// exactly the posture every other advisory-only check here takes.
export function notEnoughLife(lifeCost: number | undefined, life: number | undefined): string {
  if (!lifeCost || life === undefined) return "";
  if (lifeCost <= life) return "";
  return NOT_ENOUGH_LIFE;
}

// abilityBlocked returns the reason an ability can't be activated
// right now, or "" when it can. Advisory only — the server re-checks
// every cost; this just greys the row and explains why.
export function abilityBlocked(
  a: AbilityCost,
  tapped: boolean,
  sick: boolean,
  loyalty?: LoyaltyContext,
  life?: number,
): string {
  if (a.tap_cost && tapped) return "already tapped";
  if (a.tap_cost && sick) return "summoning sickness";
  // #1690: a life cost the paying player can't afford — Greed's fixed
  // one, and since #1688 a computed one (War Room, Murderous
  // Betrayal). Checked early, alongside the other basic cost-shape
  // arms, before the more specific candidate-shortfall reasons below.
  const shortOnLife = notEnoughLife(a.life_cost, life);
  if (shortOnLife) return shortOnLife;
  // #747: fewer options than the clause's count, not just none —
  // "needs three Foods (you have 2)".
  const sacrifice = sacrificeRangeShortfall(
    a.sacrifice_options,
    a.sacrifice_label ?? "a permanent",
  );
  if (sacrifice) return sacrifice;
  // #1213: the same question one verb over.
  const returned = returnShortfall(a.return_options, a.return_label);
  if (returned) return returned;
  // #1600: and one destination over.
  const exiled = exilePermanentShortfall(a.exile_permanent_options, a.exile_permanent_label);
  if (exiled) return exiled;
  // #759: and the tap-another cost.
  const tapOthers = tapOthersShortfall(a.tap_others_options, a.tap_others_label);
  if (tapOthers) return tapOthers;
  // CR 702.122a: a crew cost with no untapped creature to pay it is
  // unpayable. Only the empty case is judged here — whether the
  // creatures that DO exist add up to the crew number is arithmetic
  // the prompt does, with the running total in front of the player,
  // and duplicating the sum in the menu row would put two answers on
  // screen at once.
  if (a.crew_cost && (a.crew_options?.cards?.length ?? 0) === 0) {
    return "no untapped creatures to crew with";
  }
  // #625: a counter-removal cost nothing can pay — Heart of Kiran's
  // alternative crew with no planeswalker holding a loyalty counter,
  // Dragon's Hoard with no gold counter, a Vivid land out of charge
  // counters. Instant speed, like crew, so no timing arm: only the
  // empty pool is judged — except for an "among" cost, where the
  // pool can be non-empty and still short, which the row says
  // because there is no running total anywhere else to read it from.
  // #789 adds the other direction: a permanent that can't have the
  // counter its cost would put on (CR 118.3).
  const counters = counterCostBlocked(a);
  if (counters) return counters;
  // CR 606: a loyalty ability answers to the sorcery-speed window,
  // the once-per-turn flag, and "you have enough counters to pay".
  // The value 0 is a real cost, so this tests for presence.
  //
  // #1208: WHETHER the window is shut is the server's answer
  // (timing_closed, the stamp of game.ActivationTimingOpenLocked);
  // the client only puts it into words. That split is the point: a
  // per-player statement — The Wandering Emperor's "you may activate
  // her loyalty abilities any time you could cast an instant" — is
  // board state the client cannot see, and before this the row stayed
  // greyed on activations the engine accepts.
  if (a.loyalty_cost !== undefined && loyalty) {
    if (a.timing_closed) return timingReason(loyalty);
    if (loyalty.card.loyalty_activated) return "Already activated this turn";
    const unpayable = canPayLoyaltyCost(loyalty.card, a.loyalty_cost);
    if (unpayable) return unpayable;
  }
  // CR 602.5d — "activate only as a sorcery". Equip is the first
  // catalog ability to declare it. Checked after the loyalty arm so
  // a loyalty row keeps its more specific reason.
  //
  // timing_closed outranks sorcery_speed, which is only the ability's
  // PRINTED clause: a Leonin Shikari's controller's equip row carries
  // sorcery_speed, no timing_closed, and is live.
  if (a.timing_closed && a.loyalty_cost === undefined) {
    return loyalty ? timingReason(loyalty) : "sorcery-speed only";
  }
  // #743, CR 602.1b: an "Activate only if …" condition the server says
  // is false. After the timing arms, which is the order the server
  // checks in, so a sorcery-speed row keeps its more specific reason.
  // The row's label already prints the clause, so the reason doesn't.
  // #1181: an exhaust ability already spent. Before condition_unmet,
  // because Bitter Work prints both and "already activated" is the one
  // that will still be true tomorrow.
  if (a.exhausted) return ABILITY_EXHAUSTED;
  if (a.condition_unmet) return ACTIVATION_CONDITION_UNMET;
  // #844 and ADR 0117 §5: the server says this mana ability would add
  // nothing right now. Activating it is legal (CR 605.1a) and
  // pointless: it would tap the source, or spend Vivi's once-per-turn
  // activation, for no mana. So the row is greyed with the reason
  // rather than hidden, and a click never activates it.
  if (a.adds_no_mana) return ADDS_NO_MANA;
  // CR 601.2c, through the same predicate the cast path uses (#1157).
  // Not "is the list empty": a clause needs `min` candidates, and an
  // "up to N" clause needs none. Before this the row for The
  // Aetherspark's "+1: … up to one target creature you control" was
  // greyed on a board with no creature on it, an activation the engine
  // accepts and internal/legal hands to bots — #544's defect with the
  // sign flipped, withholding a move rather than offering a dead one.
  if (!hasSatisfiableTargets(a.legal_targets)) return "no legal target";
  return "";
}

// LoyaltyContext is what abilityBlocked needs to judge a loyalty
// row: the planeswalker itself (for its counters and the server's
// once-per-turn flag) and the snapshot the timing window is read
// from. Absent for callers with no snapshot to hand, in which case
// the loyalty arm is skipped and the server does the gating alone.
export interface LoyaltyContext {
  card: CardView;
  view: GameView | null | undefined;
  viewerID: string | null;
}

// timingReason puts the server's `timing_closed` into words (#1208).
//
// The BIT is the engine's and the SENTENCE is the client's, which is
// the only division of labour that survives a per-player timing
// statement: whether the window is shut depends on board state the
// client cannot see, but WHY it is shut — no priority, split second,
// a non-empty stack, somebody else's turn — is all in the snapshot
// and is what the player wants to read. `canActivateSorcerySpeedAbility`
// can say "legal" here (a restriction shut the window rather than the
// phase), in which case the generic sentence is the honest one.
function timingReason(loyalty: LoyaltyContext): string {
  const timing = canActivateSorcerySpeedAbility(loyalty.view, loyalty.viewerID);
  return timing.reason ?? "Can't activate right now";
}

// withGrantor labels a row another permanent granted (ADR 0093
// Decision 8): "Add one mana of any color (from Cryptolith Rite)".
// One row per grantor, so two Rites read as two labelled rows.
function withGrantor(label: string, row: { granted_by?: { name?: string } }): string {
  const from = grantedFromLabel(row);
  return from ? `${label} (${from})` : label;
}

// ---- ADR 0117 §2: one predicate for "can this row be used" ----------

// AbilityRowKind is which list a row came from. The card-level
// restriction that stops it differs: Arrest's `cant_activate` stops the
// activated rows, `cant_activate_mana` the mana rows (Faith's Fetters
// spares mana abilities, which is why the two are separate bits).
export type AbilityRowKind = "mana" | "activated";

// AbilityRow is what the predicate reads off a row: its cost shape and
// the server's verdict fields (AbilityCost), plus the ref the digest
// is keyed by.
export type AbilityRow = AbilityCost & { ref?: string };

// AbilityRowContext is everything about the card and the frame the
// predicate needs. The click rule (battlefieldClickIntent), the ability
// popover (ManaAbilityMenu), the mana picker (manaSource.ts) and the
// override menu (abilityItems) all build one, so a click never
// disagrees with the menu.
export interface AbilityRowContext {
  // The card the row is on, for its restrictions. Absent in a caller
  // that has only the rows (a popover mounted on its own), in which
  // case no card-level restriction is read.
  card?: CardView;
  // The card's instance ID, for the digest, when there is no `card`.
  cardID?: string;
  tapped: boolean;
  sick: boolean;
  // The PAYING player's life: the controller's on their own permanent,
  // the viewer's on an any-player row (CR 602.1a).
  payerLife?: number;
  // The popover's words for a window the server shut: when defined, a
  // `timing_closed` row is answered with them, ahead of the cost
  // reasons, as the popover always has. Undefined (the override menu)
  // leaves timing to the shared abilityBlocked, in its own order.
  timingWords?: string;
  // The frame's FULL legal-action lookup. Greys a sorcery-speed row,
  // and on another player's permanent any row, that the exact digest
  // leaves out (ADR 0105 §3, ADR 0106 §1 decision 5). The override
  // menu passes none for its own rows: it never greys on the digest.
  legalGate?: LegalActions;
  // The rows are another player's permanent's any-player rows.
  across?: boolean;
  // For a planeswalker's rows: "already activated this turn" and a −N
  // it cannot pay (CR 606.3, 606.6).
  loyalty?: LoyaltyContext;
}

// abilityRowContext builds the context for a card from what every
// caller has: the card, the viewer and the frame. One builder, so the
// popover and the click rule cannot read different fields.
export function abilityRowContext(
  card: CardView,
  opts: {
    viewerID?: string | null;
    view?: GameView | null;
    payerLife?: number;
    timingWords?: string;
    legalGate?: LegalActions;
  } = {},
): AbilityRowContext {
  const viewerID = opts.viewerID ?? null;
  return {
    card,
    tapped: !!card.tapped,
    sick: !!card.summoning_sick,
    payerLife: opts.payerLife,
    timingWords: opts.timingWords,
    legalGate: opts.legalGate,
    across: acrossFor(card, viewerID),
    loyalty: { card, view: opts.view, viewerID },
  };
}

// abilityRowBlocked is ADR 0117 §2's predicate: the reason `a` cannot
// be used right now, or "" when it can. Advisory, as every check here
// is: the server re-checks. In order:
//
//  1. the card's restriction (Arrest's `cant_activate`, Faith's
//     Fetters' `cant_activate_mana`);
//  2. the row's own `cant_activate` clause (#1210, Cursed Totem);
//  3. a {T} cost on a tapped card. CR 106.12 makes this a cost
//     question: the tapped flag matters only when the cost has {T}, so
//     a tapped Vivi Ornitier's "{0}" ability stays usable;
//  4. a {T} cost on a summoning-sick card (CR 302.6);
//  5. the popover's timing words, when it passes them;
//  6. everything the shared abilityBlocked says, with the loyalty
//     context: life, sacrifice and the other costs, a planeswalker's
//     once per turn and its −N, the timing window, exhausted,
//     condition_unmet, adds_no_mana and targets;
//  7. the digest's refusal, for a sorcery-speed row or an any-player
//     row on another player's permanent.
//
// It never judges whether a mana cost can be afforded: the auto-tapper
// pays, and ADR 0118 decides what happens when it cannot.
export function abilityRowBlocked(
  a: AbilityRow,
  kind: AbilityRowKind,
  ctx: AbilityRowContext,
): string {
  const restrictions = ctx.card?.restrictions ?? [];
  if (kind === "activated" && restrictions.includes("cant_activate")) {
    return EFFECT_STOPS_ABILITIES;
  }
  if (kind === "mana" && restrictions.includes("cant_activate_mana")) {
    return EFFECT_STOPS_ABILITIES;
  }
  if (a.cant_activate) return a.cant_activate;
  if (a.tap_cost && ctx.tapped) return "already tapped";
  if (a.tap_cost && ctx.sick) return "summoning sickness";
  if (a.timing_closed && ctx.timingWords !== undefined) {
    return ctx.timingWords || ABILITY_NOT_RIGHT_NOW;
  }
  const fromRow = abilityBlocked(a, ctx.tapped, ctx.sick, ctx.loyalty, ctx.payerLife);
  if (fromRow) return fromRow;
  if (
    kind === "activated" &&
    (a.sorcery_speed || ctx.across) &&
    digestRefusesRow(
      ctx.legalGate ?? NO_LEGAL_ACTIONS,
      ctx.cardID ?? ctx.card?.instance_id ?? "",
      a.ref,
    )
  ) {
    return ABILITY_NOT_RIGHT_NOW;
  }
  return "";
}

// JudgedRow is one row with the predicate's verdict, and whether the
// frame's highlight lookup marks it ready (ADR 0105 §2). A blocked row
// is never ready: an accent on a disabled row would say two things.
export interface JudgedRow<T> {
  a: T;
  blocked: string;
  ready: boolean;
}

// judgeAbilityRows runs the predicate over a list, ready rows first.
export function judgeAbilityRows<T extends AbilityRow>(
  list: readonly T[],
  kind: AbilityRowKind,
  ctx: AbilityRowContext,
  readyRefs: readonly string[] = [],
): JudgedRow<T>[] {
  const out = list.map((a) => {
    const blocked = abilityRowBlocked(a, kind, ctx);
    return { a, blocked, ready: !blocked && !!a.ref && readyRefs.includes(a.ref) };
  });
  return readyFirst(out, (r) => r.ready);
}

// AbilityPopoverModel is what the light ability popover lists for one
// card, judged: its special actions, mana rows, activated rows and, for
// an uncatalogued planeswalker, the manual loyalty rows (ADR 0117 §3).
export interface AbilityPopoverModel {
  special: MenuItem[];
  mana: JudgedRow<ManaAbilityView>[];
  activated: JudgedRow<ActivatedAbilityView>[];
  loyalty: MenuItem[];
}

export interface AbilityPopoverInput {
  card: CardView;
  viewerID: string | null | undefined;
  view?: GameView | null;
  payerLife?: number;
  timingWords?: string;
  legal?: LegalActions;
  legalGate?: LegalActions;
  // Which sections the surface wires: mana activations, activated
  // abilities, and the special-action / manual-loyalty sender.
  mana: boolean;
  activated: boolean;
  special: boolean;
}

// abilityPopoverModel is the popover's rows for a card, as the click
// rule reads them. The popover component builds the same lists from
// the same functions (menuAbilityRows, specialActionItems,
// manualLoyaltyRows) and judges them with abilityRowContext and
// judgeAbilityRows, so the two agree; abilityClick.render.test.ts runs
// both over the same fixtures.
export function abilityPopoverModel(i: AbilityPopoverInput): AbilityPopoverModel {
  const { card } = i;
  const legal = i.legal ?? NO_LEGAL_ACTIONS;
  const ctx = abilityRowContext(card, {
    viewerID: i.viewerID,
    view: i.view,
    payerLife: i.payerLife,
    timingWords: i.timingWords,
    legalGate: i.legalGate,
  });
  const special = i.special
    ? specialActionItems(card, card.controller || card.owner, legal, i.legalGate)
    : [];
  const manaList = i.mana ? menuManaRows(card, i.viewerID) : [];
  const activatedList = i.activated ? menuAbilityRows(card, i.viewerID) : [];
  return {
    special,
    mana: judgeAbilityRows(manaList, "mana", ctx, legal.readyManaRefs(card.instance_id)),
    activated: judgeAbilityRows(
      activatedList,
      "activated",
      ctx,
      legal.readyAbilityRefs(card.instance_id),
    ),
    loyalty: i.special && i.activated ? manualLoyaltyRows(card, i.view, i.viewerID ?? null) : [],
  };
}

// manualLoyaltyRows is the popover's copy of the override menu's
// manual loyalty rows (ADR 0117 §3): only on a planeswalker the viewer
// controls, and only with a frame to judge the window against.
// Without this, a left-click on an uncatalogued planeswalker, which
// used to open the override menu for exactly these rows (#329), would
// leave them reachable only with admin overrides on.
export function manualLoyaltyRows(
  card: CardView,
  view: GameView | null | undefined,
  viewerID: string | null,
): MenuItem[] {
  if (!view || !viewerID) return [];
  if ((card.controller || card.owner) !== viewerID) return [];
  return loyaltyAbilityItems(card, view, viewerID);
}

// abilityItems folds the permanent's mana abilities and CR 602
// activated abilities into the menu. Right-click used to open the
// dedicated ManaAbilityMenu popover; with the admin menu bound to
// the same gesture, these rows keep that surface reachable instead
// of the override menu shadowing it.
//
// ADR 0105 (#1789): `legal` is the frame's highlight lookup. A row whose
// ref is in its digest, and that the row fields do not grey, is marked
// `ready` and sorted to the top of the section. The lookup that knows
// nothing (highlights off, no digest) leaves the menu exactly as it
// was. It never greys a row: the row fields keep that job, because
// they supply the sentence.
function abilityItems(
  card: CardView,
  view: GameView,
  viewerID: string | null,
  legal: LegalActions = NO_LEGAL_ACTIONS,
): MenuItem[] {
  const readyMana = legal.readyManaRefs(card.instance_id);
  const readyAbilities = legal.readyAbilityRefs(card.instance_id);
  const isReady = (blocked: string, refs: readonly string[], ref: string | undefined) =>
    !blocked && !!ref && refs.includes(ref);
  // #1690: the PAYING player's current life, for the life-cost check.
  // That's the card's controller, not necessarily the viewer — an
  // admin override menu can open on a card the viewer doesn't control
  // (canOverride above), and it's still that controller who would pay
  // the cost.
  //
  // ADR 0117 §2: the rows are judged by the one predicate the click
  // and the popover use, with the card's restrictions (S24: Arrest,
  // Faith's Fetters, read off the wire), the loyalty context and the
  // payer's life. No timing words and no digest: this menu answers
  // timing in the shared predicate's own order and never greys its
  // own rows on the digest.
  const ctx: AbilityRowContext = {
    card,
    tapped: !!card.tapped,
    sick: !!card.summoning_sick,
    payerLife: view.seats.find((s) => s.id === (card.controller || card.owner))?.life,
    loyalty: { card, view, viewerID },
  };
  const items: MenuItem[] = [];
  // #1228: a card projects EITHER the battlefield mana list or the
  // in-zone one, never both — the server filters by the zone the card
  // is in (CR 113.6) — so one loop covers a permanent's "{T}: Add {G}"
  // and a Spirit Guide's "Exile this card from your hand: Add {R}",
  // and the index means the same thing to the engine either way. The
  // same shape the activated loop below has had since #660.
  for (const a of card.mana_abilities ?? card.zone_mana_abilities ?? []) {
    const blocked = abilityRowBlocked(a, "mana", ctx);
    // #1190: a discount note when the engine charges less than the
    // printed cost — shown only on an unblocked row, so a "why is
    // this greyed" reason never loses to a price note.
    items.push({
      id: `mana-${a.index}`,
      // ADR 0093: a row another permanent granted says so.
      label: withGrantor(a.label || a.produced || "add mana", a),
      hint: blocked || chargedManaCostNote(a) || undefined,
      disabled: !!blocked,
      ready: isReady(blocked, readyMana, a.ref) || undefined,
      activate: { kind: "mana", index: a.index },
    });
  }
  // #660: a card projects EITHER the battlefield list or the hand
  // list, never both — the server filters by the zone the card is in
  // (CR 113.6) — so one loop covers a permanent's abilities and a
  // hand card's cycling, and the index means the same thing to the
  // engine either way.
  for (const a of card.activated_abilities ?? card.zone_abilities ?? []) {
    const blocked = abilityRowBlocked(a, "activated", ctx);
    items.push(activatedItem(a, blocked, isReady(blocked, readyAbilities, a.ref)));
  }
  return readyFirst(items, (i) => i.ready === true);
}

// EFFECT_STOPS_ABILITIES is the reason on a row an Arrest-style "its
// activated abilities can't be activated" greys (S24, read off the
// wire). It restricts the object, not the activator, so it greys an
// any-player row for every seat (ADR 0106 §1 decision 2).
export const EFFECT_STOPS_ABILITIES = "an effect stops its abilities";

// ABILITY_NOT_RIGHT_NOW is the reason on an any-player row the exact
// digest leaves out (ADR 0106 §1 decision 6): the server would refuse
// it, and the row fields say nothing more specific. The same sentence
// the ability popover uses for a row the digest shuts.
export const ABILITY_NOT_RIGHT_NOW = "Can't activate this right now";

// activatedItem is one CR 602 activated-ability row of the menu.
// #1296: a price that depends on the target (Dragonfire Blade) says
// its range in the hint; the targeting banner names each target's.
function activatedItem(a: ActivatedAbilityView, blocked: string, ready: boolean): MenuItem {
  return {
    id: `ability-${a.index}`,
    label: withGrantor(a.label || "activate", a),
    hint:
      blocked ||
      targetPriceRange(a.target_charged_mana_costs) ||
      chargedManaCostNote(a) ||
      undefined,
    disabled: !!blocked,
    ready: ready || undefined,
    activate: { kind: "ability", index: a.index },
  };
}

// anyPlayerAbilityItems is the whole menu a seated viewer gets on a
// permanent somebody else controls: its "Any player may activate this
// ability" rows and nothing else (ADR 0106 §1 decision 6). The rows are
// the viewer's copy, stamped by the server with the viewer as the
// activator, so their costs, condition and timing verdicts are the
// viewer's (CR 602.1a, CR 109.5).
//
//   - The life check reads the VIEWER's life: the activator pays the
//     cost (CR 602.1a), not the permanent's controller.
//   - `legal` gives the ready accent and order, exactly as on the
//     viewer's own permanents.
//   - `gate`, the frame's full lookup, greys a row the exact digest
//     leaves out. The rows are live exactly when the digest lists their
//     ref. No digest, or the capped fallback, greys nothing new (ADR
//     0105 §3): the row fields keep that job.
function anyPlayerAbilityItems(
  card: CardView,
  view: GameView,
  viewerID: string,
  legal: LegalActions,
  gate: LegalActions,
): MenuItem[] {
  const readyAbilities = legal.readyAbilityRefs(card.instance_id);
  const ctx: AbilityRowContext = {
    card,
    tapped: !!card.tapped,
    sick: !!card.summoning_sick,
    payerLife: view.seats.find((s) => s.id === viewerID)?.life,
    legalGate: gate,
    across: true,
    loyalty: { card, view, viewerID },
  };
  const items = anyPlayerRows(card).map((a) => {
    const blocked = abilityRowBlocked(a, "activated", ctx);
    return activatedItem(a, blocked, !blocked && !!a.ref && readyAbilities.includes(a.ref));
  });
  return readyFirst(items, (i) => i.ready === true);
}

// MAX_MANUAL_MINUS caps the manual minus rows. Karn Liberated's −14
// is the deepest printed cost in Magic; anything past that is a row
// nobody will ever click, and the list is already bounded above by
// the walker's own loyalty (CR 606.6).
const MAX_MANUAL_MINUS = 14;

// MANUAL_PLUS_OFFERS is the non-negative side, which every
// planeswalker can always pay. +2 is the largest printed plus cost
// in the format.
const MANUAL_PLUS_OFFERS: readonly number[] = [2, 1, 0];

// signedLoyalty formats a cost the way the card prints it: "+1",
// "[0]", "−3" (a real minus sign, matching Magic's typography).
function signedLoyalty(n: number): string {
  if (n > 0) return `+${n}`;
  if (n === 0) return "[0]";
  return `−${-n}`;
}

// loyaltyAbilityItems is the manual loyalty-activation section for a
// planeswalker the catalog does not know about — which is still most
// of them, and is exactly what issue #329 was holding: Teferi, Who
// Slows the Sunset, four loyalty counters, `activated_abilities: []`.
//
// It offers the costs the walker can legally pay right now (+2, +1,
// [0], and −1 down to its current loyalty, CR 606.6) as
// `activate_loyalty` actions. The engine charges the counters and
// enforces CR 606.3 — sorcery speed and once per turn — and the
// players resolve the ability's text between themselves, the same
// bargain manual `tap` strikes for every other card the catalog
// can't express.
//
// It is deliberately ABSENT once the catalog knows the card: a
// Teferi, Time Raveler shows his real "+1:" and "−3:" rows, which
// pay the same cost AND run the effect. Offering both would let a
// player spend the turn's activation on the manual row by mistake.
//
// The ungated add / remove loyalty rows under "counters" are
// untouched — those are the admin escape hatch for fixing a mistake,
// and gating them would defeat the point.
function loyaltyAbilityItems(card: CardView, view: GameView, viewerID: string | null): MenuItem[] {
  if (!isPlaneswalker(card)) return [];
  if ((card.activated_abilities ?? []).length > 0) return [];
  const timing = canActivateLoyalty(card, view, viewerID);
  const hint = timing.legal ? undefined : (timing.reason ?? "can't activate right now");
  const deltas = [...MANUAL_PLUS_OFFERS];
  for (let n = 1; n <= Math.min(loyaltyOf(card), MAX_MANUAL_MINUS); n++) {
    deltas.push(-n);
  }
  return deltas.map((delta) => ({
    id: `loyalty-${delta}`,
    label: `${signedLoyalty(delta)}: activate a loyalty ability`,
    hint: hint ?? "pays the loyalty; resolve the ability's text yourselves",
    disabled: !timing.legal,
    action: {
      type: "activate_loyalty" as ActionType,
      params: {
        planeswalker_id: card.instance_id,
        delta,
        label: signedLoyalty(delta),
      },
      player: card.controller,
    },
  }));
}

function tapItems(card: CardView): MenuItem[] {
  const tapped = !!card.tapped;
  // #1438: a left-click on a mana source now taps it FOR mana, so the
  // plain tap here is the one that makes none — and says so.
  const makesMana = (card.mana_abilities?.length ?? 0) > 0;
  return [
    {
      id: "tap",
      label: makesMana ? "Tap (no mana)" : "Tap",
      hint: makesMana ? "turn it sideways without adding mana" : undefined,
      disabled: tapped,
      action: { type: "tap", params: { instance_id: card.instance_id } },
    },
    {
      id: "untap",
      label: "Untap",
      disabled: !tapped,
      action: { type: "untap", params: { instance_id: card.instance_id } },
    },
  ];
}

// goadItems offers the sandbox goad. Since #1598 a creature may carry
// several players' goads at once (CR 701.15c), so "Goad (by you)" stays
// on offer while the viewer is not among its goaders, and "Clear goad"
// (which clears every goad) appears once anyone has goaded it.
// `goaders` is absent from a server before #1598, which held one goader
// in `goaded_by`.
function goadItems(card: CardView, viewerID: string | null): MenuItem[] {
  if (!viewerID) return [];
  const goaders = card.goaders ?? (card.goaded_by ? [card.goaded_by] : []);
  const items: MenuItem[] = [];
  if (!goaders.includes(viewerID)) {
    items.push({
      id: "goad",
      label: "Goad (by you)",
      hint: "until your next turn it attacks each combat, and a player other than you, if able",
      action: {
        type: "set_goaded",
        params: { instance_id: card.instance_id, by: viewerID },
      },
    });
  }
  if (goaders.length > 0) {
    items.push({
      id: "goad-clear",
      label: goaders.length > 1 ? "Clear goads" : "Clear goad",
      action: {
        type: "set_goaded",
        params: { instance_id: card.instance_id, by: "" },
      },
    });
  }
  return items;
}

function otherCounterItems(card: CardView): MenuItem[] {
  const seen = new Set<string>(QUICK_COUNTERS);
  const names: string[] = [];
  for (const n of CARD_COUNTER_NAMES) {
    if (seen.has(n)) continue;
    seen.add(n);
    names.push(n);
  }
  // Anything already on the card that isn't in the styled registry
  // (a homebrew counter placed by an earlier custom prompt) still
  // needs a way back off.
  for (const n of Object.keys(card.counters ?? {})) {
    if (seen.has(n)) continue;
    seen.add(n);
    names.push(n);
  }
  const items: MenuItem[] = [];
  for (const name of names) {
    const have = card.counters?.[name] ?? 0;
    items.push({
      id: `counter-other-add-${name}`,
      label: `${name} +1`,
      repeat: true,
      action: counterAction(card, name, 1),
    });
    items.push({
      id: `counter-other-remove-${name}`,
      label: `${name} −1`,
      disabled: have <= 0,
      repeat: true,
      action: counterAction(card, name, -1),
    });
  }
  return items;
}

function counterItems(card: CardView): MenuItem[] {
  const items: MenuItem[] = [];
  for (const name of QUICK_COUNTERS) {
    const have = card.counters?.[name] ?? 0;
    items.push({
      id: `counter-add-${name}`,
      label: `Add ${name}`,
      hint: have > 0 ? `${have} on this card` : undefined,
      repeat: true,
      action: counterAction(card, name, 1),
    });
    items.push({
      id: `counter-remove-${name}`,
      label: `Remove ${name}`,
      disabled: have <= 0,
      repeat: true,
      action: counterAction(card, name, -1),
    });
  }
  items.push({
    id: "counter-other",
    label: "Other counters",
    items: otherCounterItems(card),
  });
  items.push({
    id: "counter-custom",
    label: "Custom counter…",
    prompt: "custom_counter",
  });
  return items;
}

function damageItems(card: CardView): MenuItem[] {
  const marked = card.damage_marked ?? 0;
  return [
    {
      id: "damage-add",
      label: "Mark 1 damage",
      hint: marked > 0 ? `${marked} marked` : undefined,
      repeat: true,
      action: damageAction(card, 1),
    },
    {
      id: "damage-remove",
      label: "Remove 1 damage",
      disabled: marked <= 0,
      repeat: true,
      action: damageAction(card, -1),
    },
    {
      id: "damage-clear",
      label: "Clear all damage",
      disabled: marked <= 0,
      action: damageAction(card, -marked),
    },
    {
      id: "damage-set",
      label: "Mark damage…",
      prompt: "mark_damage",
    },
  ];
}

function combatItems(view: GameView, card: CardView): MenuItem[] {
  const items: MenuItem[] = [];
  const defenders = view.seats.filter((s) => s.id !== card.controller && !s.eliminated);
  if (defenders.length > 0) {
    items.push({
      id: "combat-attack",
      label: card.attacking_target ? "Re-declare attacker" : "Declare attacker",
      items: [
        ...defenders.map((s) => ({
          id: `combat-attack-${s.id}`,
          label: s.display_name || s.name,
          // ADR 0080 (#1063): the seat's CR 508.1a attack tax, stated
          // on the control that charges it. Read off the server's
          // price, never derived.
          hint: attackTaxOn(view, s.id) ? `costs ${attackTaxOn(view, s.id)}` : undefined,
          action: {
            type: "declare_attacker" as ActionType,
            // auto_tap: the tax may need lands tapped for it. Inert
            // at a table with no attack tax on it.
            params: { attacker: card.instance_id, target: s.id, auto_tap: true },
          },
        })),
        // S27: planeswalkers and battles are attackable too
        // (CR 508.1d). Same verb, same payload — only the id differs,
        // which is what makes the polymorphic target cheap on this
        // side. The set is the server's; see attackTargets.ts.
        ...permanentAttackTargets(view, card.controller).map((t) => ({
          id: `combat-attack-${t.id}`,
          label: t.label,
          hint: attackTargetHint(view, t),
          action: {
            type: "declare_attacker" as ActionType,
            params: { attacker: card.instance_id, target: t.id, auto_tap: true },
          },
        })),
      ],
    });
    // #318: the board-wide sibling of the row above. Same "pick a
    // defender" submenu shape, so the bulk affordance reads as a
    // wider version of the per-card one rather than a new idea — and
    // it lives next to "Clear ALL combat", which is already a
    // board-wide verb parked in this section. Rendered from the
    // card's controller so an admin driving another seat gets that
    // seat's board, not their own.
    const plan = planAttackAll(view, card.controller);
    const n = plan.eligible.length;
    if (n === 0) {
      items.push({
        id: "combat-attack-all",
        label: "Attack with all",
        hint: "none of this seat's creatures can attack right now",
        disabled: true,
      });
    } else {
      items.push({
        id: "combat-attack-all",
        label: `Attack with all ${n}`,
        hint: `declares ${n} creature${n === 1 ? "" : "s"} in one action — one undo takes it all back`,
        items: plan.defenders.flatMap((s) => {
          const params = attackAllParams(plan, s.id);
          if (!params) return [];
          return [
            {
              id: `combat-attack-all-${s.id}`,
              label: seatLabel(s),
              // ADR 0080: the tax clause joins the count, so the
              // price is on the control that commits to it rather
              // than in a rejection toast afterwards.
              hint: [attackAllLabel(plan, s), attackAllTaxLabel(view, plan, s.id)]
                .filter(Boolean)
                .join(" · "),
              action: { type: "declare_attackers" as ActionType, params },
            },
          ];
        }),
      });
    }
  }
  // #1339 (CR 802.4a): only the attackers this card's controller is
  // DEFENDING against — attacking them, a planeswalker they control or
  // a battle they protect. Every other attacker is somebody else's to
  // block, and the server refuses it with not_defending.
  // #1706: a creature that can block more than one attacker and has
  // room is ADDED to another block rather than re-pointed, so it is
  // offered only the attackers it does not block yet.
  const adding = !!card.blocking_target && blockerHasRoom(card);
  const blocked = new Set(blockedAttackersOf(card));
  const attackers = attackersDefendedBy(view, card.controller).filter(
    (a) => !adding || !blocked.has(a.instance_id),
  );
  if (attackers.length > 0) {
    items.push({
      id: "combat-block",
      label: adding
        ? "Also block"
        : card.blocking_target
          ? "Re-declare blocker"
          : "Declare blocker",
      items: attackers.map((a) => ({
        id: `combat-block-${a.instance_id}`,
        label: a.name || "unknown attacker",
        action: {
          type: "declare_blocker" as ActionType,
          params: { blocker: card.instance_id, attacker: a.instance_id },
        },
      })),
    });
  }
  items.push({
    id: "combat-clear",
    label: "Clear ALL combat",
    hint: "wipes every attack and block declaration on the table",
    danger: true,
    action: { type: "clear_combat" },
  });
  return items;
}

function moveItems(card: CardView, location: CardLocation): MenuItem[] {
  const items: MenuItem[] = moveDestinations(location.zone).map((d) => ({
    id: `move-${d.id}`,
    label: d.label,
    action: buildMoveAction(card, location, d),
  }));
  // CR 903.9a (ADR 0115) — the commander dies into its owner's
  // graveyard, so every dies trigger sees it, and `as_commander`
  // answers the state-based action's "put it into the command zone?"
  // with yes in the same click, rather than dropping the card straight
  // into the command zone.
  if (card.is_commander && location.zone === "battlefield") {
    const gy: MoveDest = { id: "graveyard", label: "Graveyard", zone: "graveyard" };
    items.push({
      id: "move-commander-903-9",
      label: "Graveyard → command zone (CR 903.9a)",
      hint: "dies first, so dies triggers see it, then goes to the command zone",
      action: buildMoveAction(card, location, gy, true),
    });
  }
  return items;
}

// specialActionItems is the CR 116.2 special-action rows on one card —
// "Foretell {2}", "Suspend 1—{R}" and "Plot {3}{U}" in the viewer's
// own hand (#658, #659, #1342, ADR 0062 Decision 4), "Turn face up
// {1}{U}" on a face-down permanent they control (#1194, ADR 0082).
//
// A row fires the `special_action` verb DIRECTLY, with no targeting
// or cost picker in between, because no kind has a choice to make:
// the whole payload is the card and the kind, and the server finds
// the mana with auto_tap the way every other menu payment does.
//
// `actor` is WHO takes the action, and it is not the same player for
// every kind: the hand keywords are the owner's (CR 702.143a — a hand
// holds only its owner's cards), and turning a permanent face up is
// its CONTROLLER's (CR 708.6), so a stolen morph is turned up by the
// thief. The caller knows the zone, so it passes the answer rather
// than this function guessing it from two fields.
//
// `available` is the SERVER's per-kind timing answer, never
// re-derived here. The rule the client would get wrong is split
// second — foretell stays legal under it (CR 702.61b), suspend does
// not (CR 702.62c) — and a row that disagreed with the engine would
// be a rejection toast. An unavailable row is greyed rather than
// dropped, so a player can still see the card has the keyword.
//
// #1319: `sa.label` bakes the printed price in by hand ("Foretell
// {2}"), which goes stale the moment a cost modifier reaches the
// action — Ranar's "the first card you foretell each turn costs {0}".
// The hint is the ONE place that discount is visible, reusing the
// same chargedManaCostNote the activated-ability menu shows its own
// discount in, since the two fields are named for the same contract
// one level up (`cost` / `charged_cost` vs. `mana_cost` /
// `charged_mana_cost`). Availability still wins the hint slot when
// the row is greyed — a player needs to know it's not their turn
// more than they need the price note.
//
// ADR 0105 sub-PR 4 (#1789): the same rows are the ability popover's
// special-action section (ManaAbilityMenu), which is how foretell,
// suspend, plot and turn face up stopped being admin-only. Two lookups
// ride along, exactly as for the ability rows:
//
//   - `legal`, the frame's highlight lookup: a live row whose kind is in
//     the digest takes the ready accent and sorts first. The lookup that
//     knows nothing (highlights off, no digest) leaves the rows exactly
//     as they were.
//   - `gate`, the frame's FULL lookup, which the popover passes and the
//     admin menu does not (that menu never greys on the digest, as for
//     its ability rows). An `available` row the exact digest leaves out
//     is greyed: the timing is open, so the server would refuse it on
//     the price. No digest, or the capped fallback, greys nothing new.
export const SPECIAL_NOT_RIGHT_NOW = "Can't do this right now";

export function specialActionItems(
  card: CardView,
  actor: string,
  legal: LegalActions = NO_LEGAL_ACTIONS,
  gate: LegalActions = NO_LEGAL_ACTIONS,
): MenuItem[] {
  const readyKinds = legal.readySpecialActions(card.instance_id);
  const items = (card.special_actions ?? []).map((sa) => {
    const costNote = chargedManaCostNote({
      mana_cost: sa.cost,
      charged_mana_cost: sa.charged_cost,
    });
    const refused = !!sa.available && digestRefusesSpecial(gate, card.instance_id, sa.kind);
    const disabled = !sa.available || refused;
    return {
      // ADR 0103: a Room offers one unlock per door, so the door is
      // part of the id and of the payload.
      id: sa.door ? `special-${sa.kind}-${sa.door}` : `special-${sa.kind}`,
      label: sa.label || sa.kind,
      hint: !sa.available
        ? "not right now"
        : refused
          ? SPECIAL_NOT_RIGHT_NOW
          : costNote || undefined,
      disabled,
      ready: (!disabled && readyKinds.includes(sa.kind)) || undefined,
      action: {
        type: "special_action" as ActionType,
        params: sa.door
          ? {
              card_id: card.instance_id,
              kind: sa.kind,
              door: sa.door,
              strict: true,
              auto_tap: true,
            }
          : { card_id: card.instance_id, kind: sa.kind, strict: true, auto_tap: true },
        player: actor,
      },
    };
  });
  return readyFirst(items, (i) => i.ready === true);
}

// buildMenuSections is the whole menu for one card, in render order.
// An empty result means "the viewer may not override this card" and
// the component says so rather than showing a bare frame.
export function buildMenuSections(
  view: GameView,
  card: CardView,
  viewerID: string | null,
  isAdmin: boolean,
  // ADR 0105: the frame's highlight lookup, for the ability rows'
  // ready accent and order. Omitted: no information, no accent.
  legal: LegalActions = NO_LEGAL_ACTIONS,
  // ADR 0106 §1: the frame's FULL lookup, which greys an any-player row
  // the exact digest leaves out on a permanent the viewer does not
  // control. Read for those rows only; the controller's and the admin's
  // menus never grey on the digest. Omitted: no information.
  gate: LegalActions = NO_LEGAL_ACTIONS,
): MenuSection[] {
  const location = locateCard(view, card.instance_id);
  if (!location) return [];
  // ADR 0106 §1 decision 6 (#1793): a seated non-controller gets one
  // section on a permanent with "Any player may activate this ability"
  // rows: those rows. No state, counters, moves or sacrifice, which
  // stay the controller's (and an admin's, who keeps the full menu).
  if (
    location.zone === "battlefield" &&
    !isAdmin &&
    viewerID &&
    mayActivateAcross(card, viewerID)
  ) {
    const items = anyPlayerAbilityItems(card, view, viewerID, legal, gate);
    return [{ id: "abilities", label: "abilities", items }];
  }
  if (!canOverride(card, viewerID, isAdmin)) return [];

  const sections: MenuSection[] = [];
  if (location.zone === "battlefield") {
    // ADR 0082 decision 9: a FACE-DOWN permanent carries its CR 116.2g
    // "Turn face up {1}{U}" row here, above its abilities — which is
    // also where it reads, because while the permanent is face down it
    // has no abilities at all (CR 708.2a) and this is the only thing
    // its controller can do with it.
    //
    // Same rows, same rule, same verb as the hand's foretell and
    // suspend below: the server decides what is offered and whether it
    // is available, and every other permanent on the board arrives
    // with an empty list.
    const special = specialActionItems(card, card.controller || card.owner, legal);
    if (special.length > 0) {
      sections.push({ id: "special_actions", label: "special actions", items: special });
    }
    const abilities = abilityItems(card, view, viewerID, legal);
    if (abilities.length > 0) {
      sections.push({ id: "abilities", label: "abilities", items: abilities });
    }
    const loyalty = loyaltyAbilityItems(card, view, viewerID);
    if (loyalty.length > 0) {
      sections.push({ id: "loyalty", label: "loyalty abilities", items: loyalty });
    }
    sections.push({
      id: "state",
      label: "state",
      items: [...tapItems(card), ...goadItems(card, viewerID)],
    });
    sections.push({ id: "counters", label: "counters", items: counterItems(card) });
    sections.push({ id: "damage", label: "damage", items: damageItems(card) });
    if (COMBAT_STEPS.has(view.turn?.step ?? "")) {
      sections.push({ id: "combat", label: "combat", items: combatItems(view, card) });
    }
  }

  if (location.zone === "hand") {
    // ADR 0062 Decision 4: the special-action rows sit in the hand
    // card's menu, above "move to". They are only ever present on the
    // viewer's own hand — the server strips `special_actions` from
    // every other seat's, as it strips `zone_abilities`.
    const special = specialActionItems(card, card.owner, legal);
    if (special.length > 0) {
      sections.push({ id: "special_actions", label: "special actions", items: special });
    }
  }

  sections.push({
    id: "move",
    label: "move to",
    items: moveItems(card, location),
  });

  if (location.zone === "battlefield") {
    sections.push({
      id: "extras",
      items: [
        {
          id: "sacrifice",
          label: "Sacrifice",
          hint: "emits a sacrifice event, so aristocrats payoffs fire",
          danger: true,
          action: {
            type: "sacrifice_permanent",
            params: { instance_id: card.instance_id },
            player: card.controller,
          },
        },
      ],
    });
  }
  return sections;
}
