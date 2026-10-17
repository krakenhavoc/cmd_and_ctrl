// stackAbilityCard — what the C2 ability card says about an ability on
// the stack (ADR 0119, amendment 2026-10-16 §B and §E).
//
// An ability has no card of its own (CR 405.1), so the pile draws one:
// its source's art crop, a kind line naming what the rules call it, the
// verbatim text, and a cost badge in the corner. This file decides the
// kind line and the badge; StackAbilityCard.svelte only draws them.
//
// THE KIND. The wire says "triggered" or "activated" (CR 113.3b, 113.3c).
// A loyalty ability is an activated ability with a loyalty symbol in its
// cost (CR 606.2). Nothing on the stack item says so, so the leading cost
// of its verbatim text decides it: "+1: …", "−2: …", "0: …". An engine
// label written "<source> — <cost>: …" is read after its dash.
//
// THE COST. Read only from the text's leading cost, everything before the
// colon (CR 602.1a), never from the effect:
// "{T}: Add {G}." costs {T}, not {G}. A keyword ability printed without a
// colon ("Equip {2}", "Ninjutsu {1}{U}") costs the symbols after its
// keyword. A cost with no symbols ("Sacrifice a creature: …") has no
// badge; the text right under it says the whole cost. A triggered ability
// and a spell never have one (call 23).

import type { StackLaneItem } from "./stackLane";

/** What the kind line says. */
export type AbilityCardKind = "triggered" | "activated" | "loyalty";

/** The cost badge: a signed loyalty cost, or brace-notation symbols. */
export type AbilityCost = { kind: "loyalty"; value: string } | { kind: "mana"; cost: string };

/** U+2212, the minus sign a printed loyalty cost uses. */
export const MINUS = "−";

// "+1:", "−2:", "0:", "−X:", and the bracketed "[+1]:" some labels use.
// A hyphen or an en dash is read as the minus sign.
const LOYALTY_LEAD = /^\s*\[?\s*([+−–-]?)\s*(\d+|X)\s*\]?\s*:/;

/**
 * loyaltyCostOf reads a loyalty ability's cost off the start of its
 * text, signed as printed with a real minus sign: "+1", "−2", "0", "−X".
 * Null when the text does not start with one.
 */
export function loyaltyCostOf(text: string): string | null {
  const m = LOYALTY_LEAD.exec(text) ?? LOYALTY_LEAD.exec(afterDash(text));
  if (!m) return null;
  const [, sign, amount] = m;
  if (amount === "0") return "0";
  if (sign === "+") return `+${amount}`;
  // A printed loyalty cost is signed unless it is 0; "3: …" is not one.
  if (sign === "") return null;
  return `${MINUS}${amount}`;
}

/** The kind of ability a stack item is, or null for a spell or a pending row. */
export function abilityCardKind(
  item: Pick<StackLaneItem, "kind" | "text">,
): AbilityCardKind | null {
  if (item.kind === "triggered") return "triggered";
  if (item.kind !== "activated") return null;
  return loyaltyCostOf(item.text) !== null ? "loyalty" : "activated";
}

/** The kind line's words. The card draws them in mono capitals. */
export function abilityKindLine(kind: AbilityCardKind): string {
  switch (kind) {
    case "triggered":
      return "triggered ability";
    case "activated":
      return "activated ability";
    case "loyalty":
      return "loyalty ability";
  }
}

const SYMBOL = /\{[^}]+\}/g;

// The part of an engine label after "<source> — ", or "" without one.
function afterDash(text: string): string {
  const at = text.indexOf(" \u2014 ");
  return at < 0 ? "" : text.slice(at + 3);
}

/**
 * activatedCostOf is the brace-notation symbols of an activated
 * ability's leading cost: "{2}{B}, {T}, Sacrifice a creature: …" gives
 * "{2}{B}{T}". "" when the cost has no symbols or no cost can be found.
 */
export function activatedCostOf(text: string): string {
  const colon = text.indexOf(":");
  const quote = text.indexOf('"');
  let costPart: string | null = null;
  if (colon > 0 && (quote < 0 || quote > colon)) {
    costPart = text.slice(0, colon);
  } else {
    // A keyword ability with its cost after the keyword, and nothing
    // else: "Equip {2}", "Equip legendary creature {3}", "Ninjutsu {1}{U}".
    const kw = /^\s*[A-Z][A-Za-z' ]*?\s((?:\{[^}]+\})+)/.exec(text);
    if (kw) costPart = kw[1];
  }
  if (costPart === null) return "";
  return (costPart.match(SYMBOL) ?? []).join("");
}

/**
 * abilityCostOf is the badge for a stack item: a loyalty ability's
 * signed cost, an activated ability's symbols, or null (a triggered
 * ability, a spell, a pending trigger, a cost with no symbols).
 */
export function abilityCostOf(item: Pick<StackLaneItem, "kind" | "text">): AbilityCost | null {
  const kind = abilityCardKind(item);
  if (kind === "loyalty") {
    const value = loyaltyCostOf(item.text);
    return value === null ? null : { kind: "loyalty", value };
  }
  if (kind === "activated") {
    const cost = activatedCostOf(item.text);
    return cost ? { kind: "mana", cost } : null;
  }
  return null;
}
