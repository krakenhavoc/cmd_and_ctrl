// autoTapPlan.ts — #1285. What the auto-tap preview shows for each
// planned source, and the one sentence that sums the payment up.
//
// The preview used to look each planned ID up on the BATTLEFIELD and
// print "these N permanent(s) tap". Two engine changes made that
// wrong:
//
//   - #1228 lets the planner spend a Spirit Guide out of the HAND. The
//     battlefield lookup found nothing, so the modal said "ok" and
//     showed fewer sources than the payment would spend — and the one
//     it hid is the one that does not come back next turn.
//   - #1242 lets it CRACK a Gold or an Eldrazi Spawn, which is
//     sacrificed and never tapped.
//
// So each row says what paying with it costs, and the summary names
// every card that leaves the hand. "Taps 2 lands" and "taps 2 lands
// and exiles Simian Spirit Guide from your hand" are not the same
// offer.

import type { AutoTapPreview, AutoTapPreviewSource } from "./api";

export type PlanPayment = "tap" | "sacrifice" | "tap_sacrifice" | "exile";

export interface PlanRow {
  id: string;
  name: string;
  payment: PlanPayment;
  // fromHand: the source is a card in the viewer's hand, not a
  // permanent.
  fromHand: boolean;
  // gone: paying spends the source for good — a sacrifice or an
  // exile — rather than a tap that untaps next turn.
  gone: boolean;
}

// paymentOf reads a described source's cost components. A source the
// server did not describe (an older server) is a tap, which is what
// every plan entry was before #1242.
export function paymentOf(src: AutoTapPreviewSource | undefined): PlanPayment {
  if (!src) return "tap";
  if (src.exile) return "exile";
  if (src.sacrifice) return src.tap ? "tap_sacrifice" : "sacrifice";
  return "tap";
}

// paymentVerb is the short badge a row shows.
export function paymentVerb(p: PlanPayment): string {
  switch (p) {
    case "exile":
      return "exile from hand";
    case "sacrifice":
      return "sacrifice";
    case "tap_sacrifice":
      return "tap + sacrifice";
    default:
      return "tap";
  }
}

// planRows projects a preview onto the rows the modal renders, in plan
// order. `nameOf` resolves an ID the server did not name (a board
// lookup); the server's own name wins when it sent one, because it is
// the only reader that can see which pile a planned card is in.
export function planRows(
  preview: AutoTapPreview | null,
  nameOf: (id: string) => string | undefined,
): PlanRow[] {
  if (!preview?.plan) return [];
  const described = new Map<string, AutoTapPreviewSource>();
  for (const s of preview.sources ?? []) described.set(s.card_id, s);
  return preview.plan.map((id) => {
    const src = described.get(id);
    const payment = paymentOf(src);
    return {
      id,
      name: src?.name || nameOf(id) || id.slice(0, 8),
      payment,
      fromHand: src?.zone === "hand",
      gone: payment !== "tap",
    };
  });
}

function joinNames(names: string[]): string {
  if (names.length <= 1) return names.join("");
  return `${names.slice(0, -1).join(", ")} and ${names[names.length - 1]}`;
}

function plural(n: number, one: string, many: string): string {
  return `${n} ${n === 1 ? one : many}`;
}

// planSummary is the modal's one-line description of the payment.
// Taps are counted; anything that spends a source for good is NAMED,
// and a card out of the hand is named with where it comes from.
export function planSummary(rows: PlanRow[]): string {
  if (rows.length === 0) return "Nothing needs to be spent.";
  const taps = rows.filter((r) => r.payment === "tap").length;
  const crackedAfterTap = rows.filter((r) => r.payment === "tap_sacrifice").map((r) => r.name);
  const sacrificed = rows.filter((r) => r.payment === "sacrifice").map((r) => r.name);
  const exiled = rows.filter((r) => r.payment === "exile").map((r) => r.name);
  const parts: string[] = [];
  if (taps > 0) parts.push(`taps ${plural(taps, "permanent", "permanents")}`);
  const eaten = [...crackedAfterTap, ...sacrificed];
  if (eaten.length > 0) parts.push(`sacrifices ${joinNames(eaten)}`);
  if (exiled.length > 0) parts.push(`exiles ${joinNames(exiled)} from your hand`);
  const sentence = joinNames(parts);
  return sentence.charAt(0).toUpperCase() + sentence.slice(1) + ".";
}
