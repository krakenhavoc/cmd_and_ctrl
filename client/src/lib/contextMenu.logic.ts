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

import { attackAllLabel, attackAllParams, planAttackAll, seatLabel } from "./attackAll";
import { attackTargetHint, permanentAttackTargets } from "./attackTargets";
import { isCreature, isLand, isPlaneswalker } from "./cardTypes";
import { counterCostBlocked } from "./counterCost";
import type { ActionType, CardView, GameView } from "./protocol";
import { sacrificeShortfall } from "./sacrificeCost";
import {
  canActivateLoyalty,
  canActivateSorcerySpeedAbility,
  canPayLoyaltyCost,
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

// BattlefieldClickIntent is what a plain left-click on a
// battlefield permanent should do.
//
//	"abilities" — open this card's menu so the player can pick one
//	"tap"       — the historic default: toggle tapped / untapped
//	"none"      — the viewer may not drive this card at all
export type BattlefieldClickIntent = "abilities" | "tap" | "none";

// battlefieldClickIntent routes a left-click. Issue #329: "I cast
// teferi and when I click on him to choose one of his abilities it
// just tapped him."
//
// He was right that it was wrong, and the replay shows it happening
// — the `tapped` bit on his Teferi flips true / false across six
// consecutive snapshots while he clicks. PlayerPanel's click handler
// fell through every branch (targeting, combat select, block) to
// `onTapToggle`, because tap/untap is the only thing a permanent
// "does" in a sandbox.
//
// Tapping a planeswalker is close to meaningless: no loyalty ability
// has a {T} component, nothing in the rules taps one in normal play,
// and the loyalty abilities are the entire reason to click him. So
// the planeswalker branch goes to the menu instead — and it does so
// whether or not the card has catalog abilities, because a
// planeswalker with none still has the manual loyalty +/− rows
// there, which beats a meaningless tap. Tap and untap remain in that
// same menu for the rare effect that wants them.
export function battlefieldClickIntent(
  card: CardView,
  viewerID: string | null,
  isAdmin: boolean,
): BattlefieldClickIntent {
  if (!canOverride(card, viewerID, isAdmin)) return "none";
  if (isPlaneswalker(card)) return "abilities";
  // #368, the same shape one rung down. Fabled Passage's only act is
  // "{T}, Sacrifice this land: search for a basic" — a CR 602
  // activated ability, not a mana ability — and left-clicking it
  // just tapped it, leaving the fetch reachable only by right-click.
  // A land that carries a non-mana activated ability is clicked FOR
  // that ability, the way a planeswalker is clicked for its loyalty.
  // Deliberately narrow:
  //   - only lands, so a creature keeps its left-click tap (that is
  //     how the sandbox marks one tapped) and its abilities stay on
  //     right-click;
  //   - only NON-creature lands, so an animated manland is still
  //     tappable and selectable in combat;
  //   - only `activated_abilities`, so a Forest — mana abilities
  //     only — still taps on click. Utility lands are the whole
  //     affected set.
  // Tap and untap remain in the menu this opens, so nothing is lost.
  if (isLand(card) && !isCreature(card) && (card.activated_abilities?.length ?? 0) > 0) {
    return "abilities";
  }
  return "tap";
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
// ActivatedAbilityView, so one predicate covers both. Mirrors the
// identically-shaped local type in ManaAbilityMenu.svelte.
interface AbilityCost {
  tap_cost?: boolean;
  sacrifice_label?: string;
  // #747: min / max are the sacrifice count (sacrificeCost.ts).
  sacrifice_options?: { players?: string[]; cards?: string[]; min?: number; max?: number };
  legal_targets?: { players?: string[]; cards?: string[] };
  // Present, at any value including 0, on a planeswalker's loyalty
  // ability. Mana abilities never carry it.
  loyalty_cost?: number;
  // S24: "Activate only as a sorcery" (CR 602.5d). Equip is the
  // catalog's first; a loyalty ability gets the same window from its
  // own arm below rather than from this flag.
  sorcery_speed?: boolean;
  // S27: a Vehicle's crew number and the creatures that could pay
  // it. Mana abilities never carry either.
  crew_cost?: number;
  crew_options?: { players?: string[]; cards?: string[] };
  // #625: a "remove N counters" cost — its shape and what can pay it
  // right now. Mirrors ActivatedAbilityView in protocol.ts; mana
  // abilities never carry it.
  counter_cost_n?: number;
  counter_cost_kind?: string;
  counter_cost_self?: boolean;
  counter_cost_label?: string;
  counter_cost_options?: { card_id: string; kinds: { kind: string; count: number }[] }[];
}

// abilityBlocked returns the reason an ability can't be activated
// right now, or "" when it can. Advisory only — the server re-checks
// every cost; this just greys the row and explains why.
export function abilityBlocked(
  a: AbilityCost,
  tapped: boolean,
  sick: boolean,
  loyalty?: LoyaltyContext,
): string {
  if (a.tap_cost && tapped) return "already tapped";
  if (a.tap_cost && sick) return "summoning sickness";
  // #747: fewer options than the clause's count, not just none —
  // "needs three Foods (you have 2)".
  const sacrifice = sacrificeShortfall(a.sacrifice_options, a.sacrifice_label ?? "a permanent");
  if (sacrifice) return sacrifice;
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
  // Dragon's Hoard with no gold counter. Instant speed, like crew, so
  // no timing arm: only the empty pool is judged.
  const counters = counterCostBlocked(a);
  if (counters) return counters;
  // CR 606: a loyalty ability answers to the sorcery-speed window,
  // the once-per-turn flag, and "you have enough counters to pay".
  // The value 0 is a real cost, so this tests for presence.
  if (a.loyalty_cost !== undefined && loyalty) {
    const timing = canActivateLoyalty(loyalty.card, loyalty.view, loyalty.viewerID);
    if (!timing.legal) return timing.reason ?? "can't activate right now";
    const unpayable = canPayLoyaltyCost(loyalty.card, a.loyalty_cost);
    if (unpayable) return unpayable;
  }
  // CR 602.5d — "activate only as a sorcery". Equip is the first
  // catalog ability to declare it. Checked after the loyalty arm so
  // a loyalty row keeps its more specific reason.
  if (a.sorcery_speed && a.loyalty_cost === undefined && loyalty) {
    const timing = canActivateSorcerySpeedAbility(loyalty.view, loyalty.viewerID);
    if (!timing.legal) return timing.reason ?? "sorcery-speed only";
  }
  if (a.legal_targets) {
    const n = (a.legal_targets.players?.length ?? 0) + (a.legal_targets.cards?.length ?? 0);
    if (n === 0) return "no legal target";
  }
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

// abilityItems folds the permanent's mana abilities and CR 602
// activated abilities into the menu. Right-click used to open the
// dedicated ManaAbilityMenu popover; with the admin menu bound to
// the same gesture, these rows keep that surface reachable instead
// of the override menu shadowing it.
function abilityItems(card: CardView, view: GameView, viewerID: string | null): MenuItem[] {
  const tapped = !!card.tapped;
  const sick = !!card.summoning_sick;
  const loyalty: LoyaltyContext = { card, view, viewerID };
  // S24: "its activated abilities can't be activated" (Arrest,
  // Faith's Fetters). Read off the wire, not derived — the server
  // refuses these activations outright, and a row that opens a
  // rejection toast is worse than a row that says why. Faith's
  // Fetters spares mana abilities, which is why the two bits are
  // checked separately rather than as one "restricted" flag.
  const restrictions = card.restrictions ?? [];
  const restricted = restrictions.includes("cant_activate") ? "an effect stops its abilities" : "";
  const manaRestricted = restrictions.includes("cant_activate_mana")
    ? "an effect stops its abilities"
    : "";
  const items: MenuItem[] = [];
  for (const a of card.mana_abilities ?? []) {
    // Mana abilities never carry a loyalty cost, so the context is
    // inert for them — passed anyway to keep one call shape.
    const blocked = manaRestricted || abilityBlocked(a, tapped, sick, loyalty);
    items.push({
      id: `mana-${a.index}`,
      label: a.label || a.produced || "add mana",
      hint: blocked || undefined,
      disabled: !!blocked,
      activate: { kind: "mana", index: a.index },
    });
  }
  for (const a of card.activated_abilities ?? []) {
    const blocked = restricted || abilityBlocked(a, tapped, sick, loyalty);
    items.push({
      id: `ability-${a.index}`,
      label: a.label || "activate",
      hint: blocked || undefined,
      disabled: !!blocked,
      activate: { kind: "ability", index: a.index },
    });
  }
  return items;
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
  return [
    {
      id: "tap",
      label: "Tap",
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

function goadItems(card: CardView, viewerID: string | null): MenuItem[] {
  if (!viewerID) return [];
  if (card.goaded_by) {
    return [
      {
        id: "goad-clear",
        label: "Clear goad",
        action: {
          type: "set_goaded",
          params: { instance_id: card.instance_id, by: "" },
        },
      },
    ];
  }
  return [
    {
      id: "goad",
      label: "Goad (by you)",
      hint: "sandbox marker — must-attack is not enforced",
      action: {
        type: "set_goaded",
        params: { instance_id: card.instance_id, by: viewerID },
      },
    },
  ];
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
          action: {
            type: "declare_attacker" as ActionType,
            params: { attacker: card.instance_id, target: s.id },
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
            params: { attacker: card.instance_id, target: t.id },
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
              hint: attackAllLabel(plan, s),
              action: { type: "declare_attackers" as ActionType, params },
            },
          ];
        }),
      });
    }
  }
  const attackers = (view.battlefield?.cards ?? []).filter(
    (c) => !!c.attacking_target && c.controller !== card.controller,
  );
  if (attackers.length > 0) {
    items.push({
      id: "combat-block",
      label: card.blocking_target ? "Re-declare blocker" : "Declare blocker",
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
  // CR 903.9 — route a commander leaving the battlefield through the
  // replacement pipeline instead of dropping it straight in the
  // command zone. #164 (shipped in #171) made that replacement fire
  // on every path, and this manual trigger keeps it testable by hand.
  if (card.is_commander && location.zone === "battlefield") {
    const gy: MoveDest = { id: "graveyard", label: "Graveyard", zone: "graveyard" };
    items.push({
      id: "move-commander-903-9",
      label: "Graveyard → command zone (CR 903.9)",
      hint: "fires the commander-zone replacement rather than moving directly",
      action: buildMoveAction(card, location, gy, true),
    });
  }
  return items;
}

// buildMenuSections is the whole menu for one card, in render order.
// An empty result means "the viewer may not override this card" and
// the component says so rather than showing a bare frame.
export function buildMenuSections(
  view: GameView,
  card: CardView,
  viewerID: string | null,
  isAdmin: boolean,
): MenuSection[] {
  const location = locateCard(view, card.instance_id);
  if (!location) return [];
  if (!canOverride(card, viewerID, isAdmin)) return [];

  const sections: MenuSection[] = [];
  if (location.zone === "battlefield") {
    const abilities = abilityItems(card, view, viewerID);
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
