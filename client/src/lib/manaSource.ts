// manaSource.ts — clicking a mana source FOR mana (#1438).
//
// Ian: "when you click on mana and it taps can you have it tap for that
// mana and put it in the floating mana pool". Until #1438 a left-click
// on a Forest sent a raw `tap`, which turns the card sideways and adds
// nothing (#1296's reporter). Now a left-click on an untapped permanent
// you control that has a mana ability activates it:
//
//   - one fixed output (a Swamp, a Sol Ring) → `activate_mana_ability`
//     straight away.
//   - anything else → a picker anchored at the card, one option per
//     FINAL RESULT. Escape there cancels before anything is sent, so
//     nothing is tapped.
//
// #1443 made the result final. An ability whose output is a choice of
// colours (Birds of Paradise, Command Tower, a painland's "{R|W}")
// publishes `color_options`: the colours each picking slot would
// offer, narrowed (CR 903.4f) and ordered (#843) by the same server
// function the `mana_pick` prompt uses. The picker expands the ability
// into one option per colour, and the pick rides the activation as
// `color` / `colors`, so the server produces it with no second
// question. A painland is {C}, {R} (1 damage) and {W} (1 damage);
// Birds is its five colours; Command Tower the identity's. The client
// still never works out which colours a source can make: an ability
// that publishes no `color_options` (an older server) stays one option
// and the server asks the colour after the tap, as before.
//
// Everything below is pure so it can be tested without a DOM. The
// components are ManaSymbolPicker (the row of symbols, shared with the
// mana_pick prompt) and ManaSourcePicker (the anchored popover).

import { grantedFromLabel } from "./abilityRef";
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
  /**
   * #1443: the colour each picking slot of the ability adds, named up
   * front — one entry per `color_options` list. Sent with the
   * activation, so no `mana_pick` follows.
   */
  colors?: string[];
  /**
   * ADR 0093: set when another permanent GRANTED this ability
   * ("from Cryptolith Rite"). A granted option is never activated
   * silently on click — the picker opens, so the player sees what
   * they are tapping for and where it came from.
   */
  granted?: string;
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
  // ADR 0093: a granted row names its grantor beside the cost rider.
  const granted = grantedFromLabel(a);
  const rider = [manaAbilityRider(a), granted].filter(Boolean).join(" · ");
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
    granted: granted || undefined,
  };
}

// The most options one ability may expand into before the picker gives
// up and falls back to the server's own prompt. Five colours is the
// widest single slot; Orcish Lumberjack's three {R|G} slots make four.
const MAX_COLOR_OPTIONS = 12;

/**
 * colorCombos is every distinct answer to `slots` (one list per
 * picking slot), in the server's order. Mana in a pool is unordered, so
 * {W}{U} and {U}{W} are one answer, kept at its first appearance:
 * Mystic Gate's [[W,U],[W,U]] is WW, WU, UU. Null past the cap.
 */
export function colorCombos(slots: readonly (readonly string[])[]): string[][] | null {
  let combos: string[][] = [[]];
  for (const slot of slots) {
    const next: string[][] = [];
    const seen = new Set<string>();
    for (const combo of combos) {
      for (const c of slot) {
        const out = [...combo, c];
        const key = [...out].sort().join("");
        if (seen.has(key)) continue;
        seen.add(key);
        next.push(out);
        if (next.length > MAX_COLOR_OPTIONS) return null;
      }
    }
    combos = next;
  }
  return combos;
}

// The symbols one answer adds: the ability's printed output with each
// pipe slot replaced by its chosen colour, repeated by that option's
// amount ("{W3|U3|…}" is three of the chosen colour, #742). When the
// output was not published (a computed ability) or does not line up
// with the answer, just the chosen colours.
function symbolsForAnswer(produced: string, colors: readonly string[]): string[] {
  const slots = manaSymbols(produced);
  if (slots.filter((s) => s.includes("|")).length !== colors.length) return [...colors];
  const out: string[] = [];
  let next = 0;
  for (const slot of slots) {
    if (!slot.includes("|")) {
      out.push(slot);
      continue;
    }
    const color = colors[next++];
    const opt = slot.split("|").find((o) => o[0] === color) ?? color;
    const n = Math.min(Math.max(1, Number(opt.slice(1)) || 1), 5);
    for (let i = 0; i < n; i++) out.push(color);
  }
  return out;
}

/**
 * manaAbilityOptionsFor is the picker's options for one ability: the
 * one ability option for a fixed output, or — when the server
 * published `color_options` (#1443) — one option per answer, each
 * naming its colours so the activation needs no second question. A
 * greyed ability stays one greyed option, so its reason is read once.
 */
export function manaAbilityOptionsFor(card: CardView, a: ManaAbilityView): ManaPickOption[] {
  const base = manaAbilityOption(card, a);
  const slots = a.color_options ?? [];
  if (base.disabled || slots.length === 0 || slots.some((s) => s.length === 0)) return [base];
  const combos = colorCombos(slots);
  if (!combos) return [base];
  return combos.map((colors) => {
    const symbols = symbolsForAnswer(a.produced ?? "", colors);
    const adds = `Add ${symbols.map((sym) => `{${sym}}`).join("")}`;
    return {
      key: `ability-${a.index}-${colors.join("")}`,
      symbols,
      caption: captionFor(symbols, false),
      rider: base.rider,
      title: base.rider ? `${adds} — ${base.rider}` : adds,
      abilityIndex: a.index,
      colors,
      granted: base.granted,
    };
  });
}

/** Every final result `card`'s mana abilities offer, in the server's order. */
export function manaAbilityOptions(card: CardView): ManaPickOption[] {
  return (card.mana_abilities ?? []).flatMap((a) => manaAbilityOptionsFor(card, a));
}

/**
 * manaColorParams is the `activate_mana_ability` params for an answer
 * named up front: `color` for the one-slot case (every source but the
 * filter lands), `colors` for several, nothing for none — so a payload
 * with no answer is byte-for-byte what it was before #1443.
 */
export function manaColorParams(colors?: readonly string[]): {
  color?: string;
  colors?: string[];
} {
  if (!colors || colors.length === 0) return {};
  if (colors.length === 1) return { color: colors[0] };
  return { colors: [...colors] };
}

/** Whether `card` publishes at least one battlefield mana ability. */
export function hasManaAbility(card: CardView): boolean {
  return (card.mana_abilities?.length ?? 0) > 0;
}

export type ManaClickPlan =
  /** Send activate_mana_ability for this index now (with its colours). */
  | { kind: "activate"; index: number; colors?: string[] }
  /** Open the anchored picker over these options. */
  | { kind: "pick"; options: ManaPickOption[] };

/**
 * manaClickPlan decides what a left-click on an untapped mana source
 * does. Null when the card has no mana ability (the caller keeps its
 * old click). One live result goes out at once — a Swamp, and also a
 * Command Tower whose identity names one colour. A lone ability the
 * server has greyed still opens the picker, so the player reads WHY
 * instead of seeing nothing happen.
 *
 * ADR 0093 (owner decision, 2026-09-24): a GRANTED option never goes
 * out on its own. A creature under Cryptolith Rite, a Birds with its
 * own ability and the granted one, and a Forest under Chromatic
 * Lantern all open the picker — there is never a silent default
 * between two abilities, and a lone granted one is shown with its
 * grantor before it is used.
 */
export function manaClickPlan(card: CardView): ManaClickPlan | null {
  const options = manaAbilityOptions(card);
  if (options.length === 0) return null;
  const [only] = options;
  if (options.length === 1 && !only.disabled && !only.granted && only.abilityIndex !== undefined) {
    return only.colors
      ? { kind: "activate", index: only.abilityIndex, colors: only.colors }
      : { kind: "activate", index: only.abilityIndex };
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
