// exileStrip.ts — the pure half of the castable-from-exile strip
// (#1389): which exiled cards sit beside the viewer's hand, in which
// state, and what the corner badge says. ExileStrip.svelte renders
// exactly this and decides nothing on its own, so vitest can pin the
// selection and the badge without a renderer.
//
// NOTHING HERE IS A RULE. Every input is a server answer on the
// viewer's own frame:
//
//   - `castable_here` — "YOU may cast this from here NOW", timing
//     included. Stamped in exile since #1389, per viewer (#1055).
//   - `cast_prices` — what the engine will charge, after every
//     CR 601.2f modifier, cheapest first. Per viewer.
//   - `exile_play` — the grant, public: who holds it, the face it
//     opens, `cast_only`, and warp's / foretell's / plot's
//     `not_before_turn`. The only input a PENDING card has, because
//     the engine prices nothing it would not accept yet.
//
// A card the viewer may not read never arrives with any of these: the
// non-knower redaction strips `exile_play` and the cast surface off a
// face-down card, so an opponent's foretold card cannot reach the
// strip at all.

import { manaSymbols } from "./manaSymbol";
import { castableFaces } from "./faces";
import type { CardView, CastPriceView, GameView } from "./protocol";
import { grantedFace, grantedFaceIndex } from "./zoneBrowser.logic";
import { canCastFromHand, type Legality } from "./timing";

/**
 * now     — castable (or, for a land, playable) this instant.
 * waiting — the viewer holds a live permission, but the window is
 *           shut: a plotted card outside its owner's main phase, a
 *           sorcery in an end step, a card a "can't cast" clause
 *           refuses.
 * later   — a warp, plot or foretell grant whose later turn has not
 *           come. Shown dimmed with a "next turn" hint rather than
 *           hidden: the player just paid to set it up, and a card that
 *           vanished until next turn would read as lost.
 */
export type ExileStripState = "now" | "waiting" | "later";

export interface ExileStripEntry {
  /** The card exactly as the wire sent it — what the cast chain gets. */
  card: CardView;
  /** The one face the grant opens, handed to the cast chain with it. */
  face?: number;
  state: ExileStripState;
  /** "play" for a land under a "you may play it" grant (Breeches). */
  verb: "cast" | "play";
  /** The short caption under a card that is not castable now. */
  hint?: string;
}

// The card and every face a cast may choose: the server stamps the
// per-viewer answer on whichever of them the cast would announce, and
// an adventure card impulse-exiled by Ragavan carries it per face.
function surfaces(card: CardView): CardView[] {
  const faces = castableFaces(card);
  return faces[0] === card ? faces : [card, ...faces];
}

function castableNow(card: CardView): boolean {
  return surfaces(card).some((s) => s.castable_here === true);
}

function hasLivePrice(card: CardView): boolean {
  return surfaces(card).some((s) => (s.cast_prices?.length ?? 0) > 0);
}

/**
 * pendingHint is the caption for a grant whose window opens on a later
 * turn, or undefined when it is open already. `turn` is the round
 * counter the server floors on (CastPermission.NotBeforeTurn).
 */
export function pendingHint(card: CardView, turn: number | undefined): string | undefined {
  const floor = card.exile_play?.not_before_turn;
  if (floor === undefined || turn === undefined || turn >= floor) return undefined;
  return floor === turn + 1 ? "next turn" : `turn ${floor}`;
}

const ORDER: Record<ExileStripState, number> = { now: 0, waiting: 1, later: 2 };

/**
 * exileEntryFor computes the strip's entry for ONE exiled card, or
 * null when the viewer has nothing to do with it — no permission of
 * their own and no live server answer either (a bystander's read of
 * somebody else's grant), or a land a cast-only grant strands (CR
 * 305.1: a land is played, not cast, so there is nothing to offer).
 *
 * Pulled out of `exileStripEntries` (#1406) so the zone browser's
 * exile button can ask the same question about a single card without
 * re-deriving "now / waiting / later" a second way: both surfaces
 * read `castable_here`, `cast_prices` and `exile_play` and must not
 * drift into different opinions about the same card.
 */
export function exileEntryFor(
  card: CardView,
  viewerID: string,
  turn: number | undefined,
): ExileStripEntry | null {
  const grant = card.exile_play;
  const mine = grant?.player === viewerID;
  const live = castableNow(card) || hasLivePrice(card);
  if (!mine && !live) return null;
  const played = grantedFace(card, mine ? (grant ?? null) : null);
  const isLand = (played.type_line ?? "").toLowerCase().includes("land");
  if (isLand && grant?.cast_only) return null;
  const later = pendingHint(card, turn);
  let state: ExileStripState;
  if (castableNow(card)) state = "now";
  else if (later !== undefined && !live) state = "later";
  else state = "waiting";
  return {
    card,
    face: mine ? grantedFaceIndex(grant ?? null) : undefined,
    state,
    verb: isLand ? "play" : "cast",
    hint: state === "later" ? later : undefined,
  };
}

/**
 * exileStripEntries is the strip's contents for `viewerID`, castable
 * cards first, then the ones waiting on their window, then the ones
 * waiting on a later turn — each group in exile order, so a card does
 * not jump about as the frame updates.
 *
 * A card is in the strip when the viewer holds a permission over it:
 * the server stamped their own answer on it (a live permission), or
 * the public grant names them and has not opened yet. It leaves when
 * the permission ends, when it is cast, or when it leaves exile — all
 * three of which the next frame reports by simply not carrying it.
 */
export function exileStripEntries(
  view: GameView | null | undefined,
  viewerID: string | null,
): ExileStripEntry[] {
  if (!view || !viewerID) return [];
  const turn = view.turn?.number;
  const out: ExileStripEntry[] = [];
  for (const card of view.exile?.cards ?? []) {
    const e = exileEntryFor(card, viewerID, turn);
    if (e) out.push(e);
  }
  return out
    .map((e, i) => ({ e, i }))
    .sort((a, b) => ORDER[a.e.state] - ORDER[b.e.state] || a.i - b.i)
    .map(({ e }) => e);
}

/**
 * exileEntryLegality is the verdict a click should obey — the same
 * one the strip reads to grey a card and caption it (#1406, shared
 * with the zone browser's exile button so the two surfaces can't
 * disagree about the same card):
 *
 *   - "later": the grant's own floor hasn't been reached. Named with
 *     the hint rather than treated as a generic denial.
 *   - "waiting": a live permission whose window is shut right now — a
 *     plotted card outside its owner's main phase, a sorcery in an
 *     end step, a card its own `cant_cast` clause refuses. Named with
 *     the printed clause when there is one.
 *   - "now": defers to `canCastFromHand` — the server's own move
 *     list, mana included — which is a NARROWER answer than
 *     `castable_here` alone: the bit says timing is open, the move
 *     list says the viewer can actually afford it right now.
 */
export function exileEntryLegality(
  entry: ExileStripEntry,
  view: GameView,
  viewerID: string | null,
): Legality {
  if (entry.state === "later") {
    return { legal: false, reason: `Castable from exile ${entry.hint}` };
  }
  if (entry.state === "waiting") {
    return { legal: false, reason: entry.card.cant_cast || "Not castable from exile right now" };
  }
  return canCastFromHand(entry.card, view, viewerID);
}

/** The corner badge: the cheapest price, as mana symbols. */
export interface ExileCostBadge {
  /** One entry per mana symbol, braces stripped: "{1}{U}" → ["1", "U"]. */
  symbols: string[];
  /** Life charged on top, when a price has a life half. */
  life?: number;
  /** Hover text: what it costs, what it prints, and any other price. */
  title: string;
  /** Screen-reader label. */
  label: string;
}

// manaSymbols moved to manaSymbol.ts with the symbol component
// (#1438); re-exported so this module's callers and tests keep their
// import.
export { manaSymbols };

// symbolClass (#1406) is gone: the strip and the zone browser's exile
// badge both draw ManaSymbol now (#1438), which is what keeps a price
// tag looking the same wherever it appears.

function priceText(p: CastPriceView): string {
  const life = p.life ? ` + ${p.life} life` : "";
  return `${p.cost === "{0}" ? "free" : p.cost}${life}`;
}

/**
 * exileCostBadge answers "does this card cost something other than
 * what it prints, and what". Null — no badge — when the server has no
 * price for the viewer (a pending grant, an opponent's card, a land)
 * or when the cheapest price IS the printed cost.
 *
 * The prices are the ones on the face the cast would announce: the
 * card's own block (the server stamps the granted face there), and,
 * for a card whose grant leaves the half open, each face's block, the
 * rest listed on hover under the face's name.
 */
export function exileCostBadge(card: CardView): ExileCostBadge | null {
  const own = card.cast_prices ?? [];
  const cheapest = own[0];
  if (!cheapest || cheapest.printed) return null;
  const printed = grantedFace(card, card.exile_play ?? null).mana_cost;
  const lines = [
    `Casts from exile for ${priceText(cheapest)}${cheapest.label ? ` (${cheapest.label})` : ""}`,
  ];
  if (printed !== undefined && printed !== "") lines.push(`printed cost ${printed}`);
  for (const p of own.slice(1)) {
    lines.push(`or ${priceText(p)}${p.label ? ` (${p.label})` : ""}`);
  }
  for (const f of castableFaces(card)) {
    if (f === card || f.name === card.name) continue;
    const fp = f.cast_prices?.[0];
    if (fp) lines.push(`${f.name}: ${priceText(fp)}`);
  }
  const symbols = manaSymbols(cheapest.cost);
  return {
    symbols: symbols.length > 0 ? symbols : ["0"],
    life: cheapest.life,
    title: lines.join("\n"),
    label: `costs ${priceText(cheapest)} to cast from exile`,
  };
}
