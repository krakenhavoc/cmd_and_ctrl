// castStrip.ts — the pure half of the castable-from-other-zones strip
// beside the viewer's hand (#2202). The strip was the castable-from-
// exile strip (#1389); it now also holds the viewer's commanders, so a
// commander can be cast, or dragged onto the table, from beside the
// hand like an exiled card. ExileStrip.svelte renders exactly this and
// decides nothing on its own.
//
// NOTHING HERE IS A RULE. Every input is a server answer on the
// viewer's own frame:
//
//   - the viewer's own seat's `command` zone — what the commanders are.
//     Only the viewer's seat is read, so an opponent's commander never
//     reaches the strip.
//   - `cast_prices` on a command-zone card (#2202) — the engine's price
//     for a cast from the command zone: the printed cost, the commander
//     tax (CR 903.8) and every cost modifier, already added up.
//   - the move list (canCastFromHand, the cast gate the command zone
//     panel and the hand read) — whether the commander is castable now.
//   - `commander_casts` — how many times the viewer has cast it from
//     the command zone, for the hover text only. The tag is the price.
//
// The exile half is exileStrip.ts, unchanged; this module puts the two
// side by side.

import type { CardView, GameView } from "./protocol";
import {
  exileCostBadge,
  exileEntryLegality,
  exileStripEntries,
  priceSymbols,
  priceText,
  type ExileCostBadge,
  type ExileStripEntry,
} from "./exileStrip";
import { canCastFromHand, castAnywayBlocked, castAnywayOffered, type Legality } from "./timing";

/** The `from_zone` a strip card is cast from. */
export type CastStripZone = "exile" | "command";

export interface CastStripEntry extends ExileStripEntry {
  zone: CastStripZone;
  /** A commander's earlier casts from the command zone, for the hover. */
  casts?: number;
}

/**
 * commanderTax is the extra generic mana CR 903.8 charges after `casts`
 * earlier casts from the command zone: {2} for each. Display only — the
 * command zone panel's "+N" badge and the strip's hover line. What a
 * cast costs is the server's `cast_prices`, never this.
 */
export function commanderTax(casts: number | undefined): number {
  return Math.max(0, casts ?? 0) * 2;
}

/**
 * commanderStripEntries is every commander in the VIEWER'S own command
 * zone, in zone order. Each one is in the strip whether or not it is
 * castable this instant, as a hand card is in the hand: the cast gate
 * (castStripLegality) lights or greys it.
 */
export function commanderStripEntries(
  view: GameView | null | undefined,
  viewerID: string | null,
): CastStripEntry[] {
  if (!view || !viewerID) return [];
  const seat = view.seats?.find((s) => s.id === viewerID);
  if (!seat) return [];
  return (seat.command?.cards ?? []).map((card) => ({
    card,
    zone: "command" as const,
    state: "now" as const,
    verb: "cast" as const,
    casts: seat.commander_casts?.[card.instance_id] ?? 0,
  }));
}

/**
 * castStripEntries is the strip's contents: the viewer's commanders
 * first, nearest the hand, then the exile cards in exileStripEntries'
 * order. Commanders lead because they are always there (until cast),
 * so an exile card arriving or leaving never moves them.
 */
export function castStripEntries(
  view: GameView | null | undefined,
  viewerID: string | null,
): CastStripEntry[] {
  const exile = exileStripEntries(view, viewerID).map(
    (e): CastStripEntry => ({ ...e, zone: "exile" }),
  );
  return [...commanderStripEntries(view, viewerID), ...exile];
}

/**
 * castStripLegality is the verdict a click obeys and the greying reads.
 * A commander asks the same gate the command zone panel's click asks
 * (canCastFromHand: the server's move list, mana included), so the two
 * surfaces never disagree about one card; an exile card keeps
 * exileEntryLegality.
 */
export function castStripLegality(
  entry: CastStripEntry,
  view: GameView,
  viewerID: string | null,
): Legality {
  if (entry.zone === "command") return canCastFromHand(entry.card, view, viewerID);
  return exileEntryLegality(entry, view, viewerID);
}

/**
 * castStripOffersCastAnyway says whether a strip card gets the "Cast
 * anyway (don't pay)" row (ADR 0118 §2): a commander, or an exile entry
 * whose verb is "cast" (a land under a "you may play it" grant is
 * played, never cast).
 */
export function castStripOffersCastAnyway(entry: CastStripEntry): boolean {
  return entry.verb === "cast" && castAnywayOffered(entry.card, entry.zone);
}

/**
 * castStripCastAnywayBlocked is why a strip card's "Cast anyway (don't
 * pay)" row is greyed, or "" when it is live. An exile entry whose grant
 * waits on a later turn or a closed window says so first, in the strip's
 * own words; everything else is timing.ts castAnywayBlocked, which reads
 * the entry's `castable_here` but never the mana.
 */
export function castStripCastAnywayBlocked(
  entry: CastStripEntry,
  view: GameView,
  viewerID: string | null,
): string {
  if (entry.zone === "exile" && entry.state !== "now") {
    return exileEntryLegality(entry, view, viewerID).reason ?? "Not castable from exile right now";
  }
  return castAnywayBlocked(entry.card, view, viewerID, entry.zone);
}

/**
 * commanderCostBadge is a commander's price tag: the cheapest
 * `cast_prices` entry, shown only when it is NOT the printed cost —
 * so no tag before the first cast, and none for a frame that carries
 * no price (an older server). Hovering lists the printed cost and the
 * commander tax; a cost modifier is already inside the total.
 */
export function commanderCostBadge(card: CardView, casts = 0): ExileCostBadge | null {
  const prices = card.cast_prices ?? [];
  const cheapest = prices[0];
  if (!cheapest || cheapest.printed) return null;
  const lines = [
    `Casts from the command zone for ${priceText(cheapest)}${cheapest.label ? ` (${cheapest.label})` : ""}`,
  ];
  if (card.mana_cost) lines.push(`printed cost ${card.mana_cost}`);
  const tax = commanderTax(casts);
  if (tax > 0) {
    lines.push(`commander tax +{${tax}} (cast ${casts} ${casts === 1 ? "time" : "times"} before)`);
  }
  for (const p of prices.slice(1)) {
    lines.push(`or ${priceText(p)}${p.label ? ` (${p.label})` : ""}`);
  }
  const symbols = priceSymbols(cheapest);
  return {
    symbols: symbols.length > 0 ? symbols : ["0"],
    life: cheapest.life,
    title: lines.join("\n"),
    label: `costs ${priceText(cheapest)} to cast from the command zone`,
  };
}

/** The strip's price tag for any entry, or null for none. */
export function castStripBadge(entry: CastStripEntry): ExileCostBadge | null {
  return entry.zone === "command"
    ? commanderCostBadge(entry.card, entry.casts)
    : exileCostBadge(entry.card);
}
