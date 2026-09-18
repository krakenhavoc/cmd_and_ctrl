// counterCost.ts — #625, then #789: the choices behind a counter
// activation cost, kept out of the modal so they can be tested without
// a component harness.
//
// Five printed shapes arrive on ActivatedAbilityView — and, since
// #789, on ManaAbilityView too, with the same field names, because
// the server carries ONE component with two owners:
//
//   self      "Remove a gold counter from this artifact"
//             counter_cost_self; nothing to send, the source pays.
//   other     "remove a loyalty counter from a planeswalker you control"
//             counter_cost_label; send the chosen permanent.
//   any kind  "Remove a counter from a creature you control"
//             no counter_cost_kind; send the permanent AND the kind.
//   among     "Remove two +1/+1 counters from among artifacts you control"
//             counter_cost_among; send several permanents AND a count
//             for each, totalling counter_cost_n.
//   variable  "Remove any number of storage counters from this land"
//             counter_cost_variable; counter_cost_n is the FLOOR and
//             counter_cost_max the ceiling, and the count is sent.
//
// The server lists what can pay right now in counter_cost_options —
// permanents the viewer controls holding enough counters, each with the
// kinds that could pay, most counters first — from the same candidate
// walk it validates against. The client only has to choose among them.
//
// The sixth shape has no choice in it at all: counter_cost_add is a
// cost that PUTS a counter on the source (Devoted Druid). There is
// nothing to pick, so it never opens a prompt; it only greys the menu
// row when the server says the permanent can't have the counter.

// CounterCostShape is the subset of ActivatedAbilityView / ManaAbilityView
// this file reads. Its own type rather than the protocol one so
// contextMenu.logic.ts's structural AbilityCost can pass through it too.
export interface CounterCostShape {
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
}

// CounterChoice is one permanent's share of a payment: this permanent,
// this kind, this many. A single-permanent payment is a one-entry
// list, which is what lets one payload builder serve all five shapes.
export interface CounterChoice {
  cardID: string;
  kind: string;
  count: number;
  // n is how many counters this part pays. Defaults to the printed
  // count for a fixed cost; set explicitly by the among and variable
  // pickers.
  n?: number;
}

// hasCounterCost reports whether the ability carries the removal
// component.
export function hasCounterCost(a: CounterCostShape): boolean {
  return (a.counter_cost_n ?? 0) > 0 || !!a.counter_cost_variable;
}

// hasCounterAddCost reports the other direction — a cost that puts a
// counter on the source.
export function hasCounterAddCost(a: CounterCostShape): boolean {
  return (a.counter_cost_add ?? 0) > 0;
}

// isMultiCounterCost reports a cost the player pays with more than one
// decision — which permanents AND how many from each (among), or how
// many (variable). These open the many-pick modal; the rest open the
// single-pick one, or none at all.
export function isMultiCounterCost(a: CounterCostShape): boolean {
  return hasCounterCost(a) && (!!a.counter_cost_among || !!a.counter_cost_variable);
}

// counterChoices flattens the server's options into one choice per
// (permanent, kind), most counters first across the whole set — the
// same order the bots spend their budget in.
export function counterChoices(a: CounterCostShape): CounterChoice[] {
  const out: CounterChoice[] = [];
  for (const o of a.counter_cost_options ?? []) {
    for (const k of o.kinds ?? []) {
      out.push({ cardID: o.card_id, kind: k.kind, count: k.count });
    }
  }
  // Array.prototype.sort is stable, so ties keep the server's order.
  return out.sort((x, y) => y.count - x.count);
}

// autoCounterChoice is the choice to make without asking: when there
// is exactly one permanent that can pay and exactly one kind on it.
// Heart of Kiran with one planeswalker out, Dragon's Hoard, Mikaeus,
// every Vivid land — the common case — never open a prompt.
//
// A multi-pick cost is never auto-chosen even with one option: an
// among payment still has to say how many, and a variable one is the
// whole question. Null when there is a real choice (or none at all).
export function autoCounterChoice(a: CounterCostShape): CounterChoice | null {
  if (isMultiCounterCost(a)) return null;
  const opts = a.counter_cost_options ?? [];
  if (opts.length !== 1 || (opts[0].kinds ?? []).length !== 1) return null;
  const k = opts[0].kinds[0];
  return { cardID: opts[0].card_id, kind: k.kind, count: k.count };
}

// counterCostNeedsPrompt reports whether paying this cost still needs
// an answer from the player: more than one permanent or kind could
// pay, or the count is the player's to name. False for the common
// self-with-a-printed-kind-and-count case — every Vivid land, Ramos,
// Dragon's Hoard — which fires straight from the menu, and false for
// a cost nothing can pay (the row is greyed instead).
export function counterCostNeedsPrompt(a: CounterCostShape): boolean {
  if (!hasCounterCost(a)) return false;
  if ((a.counter_cost_options ?? []).length === 0) return false;
  return isMultiCounterCost(a) || autoCounterChoice(a) === null;
}

// counterCostTotal is how many counters a payment must total: the
// printed count, or the announced count for a variable cost.
export function counterCostTotal(a: CounterCostShape, chosen: CounterChoice[]): number {
  if (a.counter_cost_variable) return chosen.reduce((sum, c) => sum + (c.n ?? 0), 0);
  return a.counter_cost_n ?? 0;
}

// counterPaymentReady reports whether the picks add up to a payment
// the server will accept: exactly N for an among cost, at least the
// floor for a variable one, one pick for everything else.
export function counterPaymentReady(a: CounterCostShape, chosen: CounterChoice[]): boolean {
  if (!hasCounterCost(a)) return true;
  const paid = chosen.reduce((sum, c) => sum + (c.n ?? 0), 0);
  if (a.counter_cost_variable) return paid >= (a.counter_cost_n ?? 0) && paid > 0;
  if (a.counter_cost_among) return paid === (a.counter_cost_n ?? 0);
  return chosen.length === 1;
}

// counterCostBlocked is the menu row's reason when the cost cannot be
// paid, or "" when it can. Advisory: the server re-checks.
export function counterCostBlocked(a: CounterCostShape): string {
  if (hasCounterAddCost(a) && a.counter_add_blocked) {
    const kind = a.counter_cost_add_kind ? `${a.counter_cost_add_kind} ` : "";
    return `it can't have ${kind}counters put on it`;
  }
  if (!hasCounterCost(a)) return "";
  const opts = a.counter_cost_options ?? [];
  const n = a.counter_cost_n ?? 1;
  const kind = a.counter_cost_kind ? `${a.counter_cost_kind} ` : "";
  if (opts.length === 0) {
    if (a.counter_cost_variable || a.counter_cost_self) {
      return n <= 1 ? `no ${kind}counter to remove` : `fewer than ${n} ${kind}counters`;
    }
    const what = n === 1 ? `a ${kind}counter` : `${n} ${kind}counters`;
    return `nothing to remove ${what} from (${a.counter_cost_label ?? "a permanent you control"})`;
  }
  if (a.counter_cost_among) {
    // The options exist but may not add up — the one question the
    // option list does not answer on its own.
    let available = 0;
    for (const o of opts) {
      for (const k of o.kinds ?? []) {
        if (!a.counter_cost_kind || k.kind === a.counter_cost_kind) available += k.count;
      }
    }
    if (available < n) {
      return `only ${available} ${kind}counters among ${a.counter_cost_label ?? "those permanents"}`;
    }
  }
  return "";
}

// counterPaymentParams is the activate_ability / activate_mana_ability
// payload half for a chosen payment: `counter_source_ids` for every
// form but the self one, `counter_counts` when the engine cannot
// assume the printed count (among and variable), and `counter_kind`
// for the any-kind form.
export interface CounterPayment {
  counter_source_ids?: string[];
  counter_counts?: number[];
  counter_kind?: string;
}

export function counterPaymentParams(
  a: CounterCostShape,
  chosen: CounterChoice | CounterChoice[] | undefined,
): CounterPayment {
  if (!hasCounterCost(a) || !chosen) return {};
  const picks = Array.isArray(chosen) ? chosen : [chosen];
  if (picks.length === 0) return {};
  const out: CounterPayment = {};
  if (!a.counter_cost_self) out.counter_source_ids = picks.map((c) => c.cardID);
  if (a.counter_cost_among || a.counter_cost_variable) {
    out.counter_counts = picks.map((c) => c.n ?? a.counter_cost_n ?? 1);
  }
  if (!a.counter_cost_kind) out.counter_kind = picks[0].kind;
  return out;
}

// counterChoiceKey identifies a choice for the picker's selection.
export function counterChoiceKey(c: CounterChoice): string {
  return `${c.cardID}\u0000${c.kind}`;
}
