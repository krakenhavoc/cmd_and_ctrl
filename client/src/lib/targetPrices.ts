// targetPrices.ts — #1296. An activated ability whose price reads its
// target ("Equip {4}. This ability costs {1} less to activate for each
// color of the creature it targets" — Dragonfire Blade) has no single
// price, so the server ships one per legal target
// (`target_charged_mana_costs`) from the same function the payment
// charges. These are the words the client shows for it:
//
//   - the menu row's hint, BEFORE the ability is picked: the range
//     ("{2}–{4} depending on the target");
//   - the targeting banner, BEFORE the click that commits the
//     activation: each price with the targets that pay it.
//
// An equip's single-target click is also its confirm, so the banner is
// the last place a player can see the price before paying it.

export type TargetPrices = Record<string, string>;

// priceLabel is how one charged cost reads: the cost string, or "free"
// for a discount that emptied the mana component out (the server's
// "", never undefined here).
export function priceLabel(cost: string): string {
  return cost === "" ? "free" : cost;
}

// manaValueOf is a cost string's mana value — {2}{U} is 3, "" is 0 —
// used only to ORDER prices cheapest first. Hybrid and Phyrexian
// symbols count one each (CR 202.3); X counts zero.
export function manaValueOf(cost: string): number {
  let n = 0;
  for (const m of cost.matchAll(/\{([^}]*)\}/g)) {
    const sym = m[1];
    if (/^\d+$/.test(sym)) n += Number(sym);
    else if (sym !== "X") n += 1;
  }
  return n;
}

// distinctPrices is the set of prices in `prices`, cheapest first.
export function distinctPrices(prices: TargetPrices | undefined): string[] {
  if (!prices) return [];
  const out = [...new Set(Object.values(prices))];
  out.sort((a, b) => manaValueOf(a) - manaValueOf(b) || a.localeCompare(b));
  return out;
}

// targetPriceRange is the menu row's hint: "{2}–{4} depending on the
// target" when the targets price differently, "" when there is one
// price or none (the row's ordinary charged-cost note covers that).
export function targetPriceRange(prices: TargetPrices | undefined): string {
  const ps = distinctPrices(prices);
  if (ps.length < 2) return "";
  return `${priceLabel(ps[0])}–${priceLabel(ps[ps.length - 1])} depending on the target`;
}

// targetPriceSummary is the banner's line: each price, cheapest first,
// with the names of the targets that pay it — "{2}: Vivi Ornitier ·
// {4}: Golem, Ornithopter". `nameOf` resolves an ID (a card or a
// player) and may return undefined for one the viewer cannot name,
// which is then left out of the list rather than shown as a UUID.
// "" when there is nothing worth saying: no prices, or every target
// costing the same (which the row already said).
export function targetPriceSummary(
  prices: TargetPrices | undefined,
  nameOf: (id: string) => string | undefined,
): string {
  const ps = distinctPrices(prices);
  if (!prices || ps.length < 2) return "";
  const parts: string[] = [];
  for (const price of ps) {
    const names = Object.entries(prices)
      .filter(([, p]) => p === price)
      .map(([id]) => nameOf(id))
      .filter((n): n is string => !!n)
      .sort((a, b) => a.localeCompare(b));
    if (names.length === 0) continue;
    parts.push(`${priceLabel(price)}: ${names.join(", ")}`);
  }
  return parts.join(" · ");
}
