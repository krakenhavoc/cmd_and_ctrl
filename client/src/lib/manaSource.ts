// manaSource.ts — clicking a mana source FOR mana (#1438).
//
// Ian: "when you click on mana and it taps can you have it tap for that
// mana and put it in the floating mana pool". Until #1438 a left-click
// on a Forest sent a raw `tap`, which turns the card sideways and adds
// nothing (#1296's reporter). Now a left-click on an untapped permanent
// you control that has a mana ability activates it:
//
//   - one mana ability → `activate_mana_ability` straight away. That
//     covers a Swamp and a Sol Ring, and ALSO a Birds of Paradise or a
//     Command Tower: their one ability produces a choice, and the
//     server asks it as a `mana_pick` prompt once the ability has been
//     activated, with the colours it computed (commander identity
//     narrowing, CR 903.4f) in the order it chose (#843). The client
//     never works out which colours a source can make.
//   - several mana abilities (a painland's {C} and its coloured
//     ability; Tarnished Citadel) → a picker anchored at the card, one
//     option per ABILITY, built from the server's own
//     `mana_abilities` list. Escape there cancels before anything is
//     sent, so nothing is tapped.
//
// Everything below is pure so it can be tested without a DOM. The
// components are ManaSymbolPicker (the row of symbols, shared with the
// mana_pick prompt) and ManaSourcePicker (the anchored popover).

import {
  ABILITY_EXHAUSTED,
  ACTIVATION_CONDITION_UNMET,
  NO_COMMANDER_IDENTITY,
} from "./contextMenu.logic";
import type { ColorButton } from "./manaPick";
import { manaSymbolMeta, manaSymbols } from "./manaSymbol";
import type { CardView, ManaAbilityView } from "./protocol";

/** One choice in a ManaSymbolPicker. */
export interface ManaPickOption {
  /** Stable {#each} key. */
  key: string;
  /** Symbols drawn on the button, braces stripped: ["C", "C"]. */
  symbols: string[];
  /**
   * True when the symbols are alternatives rather than a total: a
   * painland's "{R} or {W}" ability. The picker draws them smaller
   * and says the colour is chosen next.
   */
  choice?: boolean;
  /** Short words under the symbols: "Red", "2 Colorless". */
  caption: string;
  /** What else happens: "deals 1 damage to you", "pay 1 life". */
  rider?: string;
  /** Hover text and accessible name. */
  title: string;
  /** Non-empty when the server says this option cannot be used. */
  disabled?: string;
  /** Which mana ability this option activates (ability pickers). */
  abilityIndex?: number;
  /** Which colour this option answers (mana_pick / choose_color). */
  color?: string;
}

// The server's own "can't" flags on a mana ability, in words. These
// are verdicts the server published, not rules the client derives:
// summoning sickness and unpayable costs are deliberately NOT here —
// the activation goes out and the server's refusal is shown.
function blockedReason(card: CardView, a: ManaAbilityView): string {
  if ((card.restrictions ?? []).includes("cant_activate_mana")) {
    return "an effect stops its abilities";
  }
  if (a.exhausted) return ABILITY_EXHAUSTED;
  if (a.condition_unmet) return ACTIVATION_CONDITION_UNMET;
  if (a.adds_no_mana) return NO_COMMANDER_IDENTITY;
  return "";
}

// The damage rider the server spells out in the label rather than as
// a cost field: "Add {R} or {W}. This land deals 1 damage to you." →
// "deals 1 damage to you". Only the sentence AFTER the "Add …" one is
// rider; a label with nothing after it has none.
export function labelRider(label: string | undefined): string {
  if (!label) return "";
  const m = /Add\b[^.]*\.\s*(.+)$/.exec(label);
  if (!m) return "";
  return m[1]
    .replace(/\.\s*$/, "")
    .replace(/^(This|That|The)\s+(land|creature|artifact|permanent|enchantment)\s+/i, "")
    .trim();
}

/**
 * riders lists every part of activating `a` that is not "add mana":
 * the printed rider plus each cost component other than {T}.
 */
export function manaAbilityRider(a: ManaAbilityView): string {
  const parts: string[] = [];
  const printed = labelRider(a.label);
  if (printed) parts.push(printed);
  const mana = a.charged_mana_cost ?? a.mana_cost;
  if (mana) parts.push(`pay ${mana}`);
  if (a.life_cost) parts.push(`pay ${a.life_cost} life`);
  if (a.sacrifice_cost) parts.push("sacrifice it");
  if (a.sacrifice_options) parts.push(`sacrifice ${a.sacrifice_label || "a permanent"}`);
  if (a.exile_self) parts.push("exile it");
  if (a.discard_cost_n) parts.push(a.discard_cost_label || `discard ${a.discard_cost_n}`);
  if (a.exile_cost_n) parts.push(a.exile_cost_label || `exile ${a.exile_cost_n}`);
  if (a.counter_cost_label) parts.push(a.counter_cost_label);
  return parts.join(", ");
}

// "Red", "2 Colorless", "Blue and Black", "Red or White", "Any color".
function captionFor(symbols: string[], choice: boolean): string {
  if (symbols.length === 0) return "";
  const names = symbols.map((s) => manaSymbolMeta(s).name);
  if (choice) {
    const colors = new Set(symbols.filter((s) => "WUBRG".includes(s)));
    return colors.size === 5 ? "Any color" : names.join(" or ");
  }
  const counts = new Map<string, number>();
  for (const n of names) counts.set(n, (counts.get(n) ?? 0) + 1);
  return Array.from(counts, ([n, k]) => (k > 1 ? `${k} ${n}` : n)).join(" and ");
}

/**
 * manaAbilityOption turns one server-published mana ability into a
 * picker option. `produced` is read, never completed: an ability whose
 * output is computed at activation (Cabal Coffers) publishes none, and
 * its option shows the server's label instead of symbols.
 */
export function manaAbilityOption(card: CardView, a: ManaAbilityView): ManaPickOption {
  const produced = a.produced ?? "";
  const slots = manaSymbols(produced);
  const choice = slots.some((s) => s.includes("|"));
  const symbols = slots.flatMap((s) => s.split("|"));
  const rider = manaAbilityRider(a);
  const caption = captionFor(symbols, choice) || a.label || "Add mana";
  const disabled = blockedReason(card, a);
  const what = a.label || (produced ? `Add ${produced}` : "Add mana");
  return {
    key: `ability-${a.index}`,
    symbols,
    choice,
    caption,
    rider: rider || undefined,
    title: disabled ? `${what} — ${disabled}` : what,
    disabled: disabled || undefined,
    abilityIndex: a.index,
  };
}

/** Every mana ability on `card`, in the server's order. */
export function manaAbilityOptions(card: CardView): ManaPickOption[] {
  return (card.mana_abilities ?? []).map((a) => manaAbilityOption(card, a));
}

/** Whether `card` publishes at least one battlefield mana ability. */
export function hasManaAbility(card: CardView): boolean {
  return (card.mana_abilities?.length ?? 0) > 0;
}

export type ManaClickPlan =
  /** Send activate_mana_ability for this index now. */
  | { kind: "activate"; index: number }
  /** Open the anchored picker over these options. */
  | { kind: "pick"; options: ManaPickOption[] };

/**
 * manaClickPlan decides what a left-click on an untapped mana source
 * does. Null when the card has no mana ability (the caller keeps its
 * old click). A lone ability the server has greyed still opens the
 * picker, so the player reads WHY instead of seeing nothing happen.
 */
export function manaClickPlan(card: CardView): ManaClickPlan | null {
  const options = manaAbilityOptions(card);
  if (options.length === 0) return null;
  if (options.length === 1 && !options[0].disabled && options[0].abilityIndex !== undefined) {
    return { kind: "activate", index: options[0].abilityIndex };
  }
  return { kind: "pick", options };
}

/**
 * colorPickOptions adapts a mana_pick / choose_color prompt's buttons
 * (manaPick.colorButtons, already in the server's order) to picker
 * options. The order is kept exactly; nothing is sorted.
 */
export function colorPickOptions(
  buttons: readonly ColorButton[],
  verb: "add" | "choose",
): ManaPickOption[] {
  return buttons.map((b) => {
    const amount = Math.max(1, b.amount);
    const symbols = Array.from({ length: Math.min(amount, 5) }, () => b.color);
    const title =
      verb === "add"
        ? amount > 1
          ? `Add ${amount} ${b.label} mana`
          : `Add ${b.label} mana`
        : `Choose ${b.label}`;
    return {
      key: `color-${b.color}`,
      symbols,
      caption: amount > 1 ? `${amount} ${b.label}` : b.label,
      title,
      color: b.color,
    };
  });
}

// ---- anchoring ---------------------------------------------------

export interface AnchorRect {
  left: number;
  top: number;
  right: number;
  bottom: number;
}

export interface PopoverPlacement {
  left: number;
  top: number;
  side: "above" | "below";
}

/**
 * placePopover puts a `width`×`height` box next to `anchor` inside a
 * `vw`×`vh` viewport, keeping a `gutter` (16 px, the mobile side
 * margin) clear on every edge. Above the card is preferred — your own
 * lands sit at the bottom of the screen — then below, then whichever
 * side has more room, clamped. Horizontally it centres on the card
 * and slides to stay inside, so a 390 px phone never scrolls sideways.
 */
export function placePopover(
  anchor: AnchorRect,
  width: number,
  height: number,
  vw: number,
  vh: number,
  gutter = 16,
  gap = 8,
): PopoverPlacement {
  const centre = (anchor.left + anchor.right) / 2;
  const maxLeft = Math.max(gutter, vw - gutter - width);
  const left = Math.min(Math.max(gutter, centre - width / 2), maxLeft);

  const aboveTop = anchor.top - gap - height;
  const belowTop = anchor.bottom + gap;
  const maxTop = Math.max(gutter, vh - gutter - height);
  if (aboveTop >= gutter) return { left, top: aboveTop, side: "above" };
  if (belowTop <= maxTop) return { left, top: belowTop, side: "below" };
  const roomAbove = anchor.top;
  const roomBelow = vh - anchor.bottom;
  if (roomAbove >= roomBelow) return { left, top: gutter, side: "above" };
  return { left, top: maxTop, side: "below" };
}
