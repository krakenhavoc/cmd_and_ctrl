// playerKeywordBadges.ts — #1201, the client half of #1197's player
// protection and hexproof, and since #1200 of the life-total lock
// that rides the same seat tile.
//
// PlayerView.keywords (docs/protocol.md, server/internal/protocol/
// view.go) carries the bare engine tokens a seat has right now:
// "hexproof", or a general "protection from <quality>" (CR 702.16i /
// CR 702.11d). Unlike CardView.protection, the wire does not parse
// the quality into a structured ProtectionView for a seat — there is
// exactly one production quality today ("everything", from Teferi's
// Protection, The One Ring, Leyline of Sanctity and Aegis of the
// Gods), so a second structured field for one value wasn't worth
// shipping. This title-cases the parsed quality instead of switching
// on the two spellings a catalogued card produces today, so a future
// card granting a player "protection from red" or "protection from
// Demons" reaches a badge without a wire change.
//
// The visual language is #979's card protection badge
// (KeywordBadgeRow.svelte): the same shield abbreviations (ALL for
// everything, PLR for the chosen player, else the first three
// letters of the quality) and the same hexproof icon
// (keywordIcons.ts), reused here rather than re-derived — a second
// copy of either would be free to drift from the card badge's.

import { KEYWORD_ICONS } from "./keywordIcons";
import type { GameEndGateView } from "./protocol";

export interface PlayerKeywordBadge {
  // The raw wire token — also the {#each} key, since two Leylines
  // both granting "hexproof" collapse to one badge (see dedupe
  // below), same posture as KeywordBadgeRow's protection dedupe
  // (CR 702.16m, two Swords of Fire and Ice on one creature).
  key: string;
  // The badge face when there's no icon for this token.
  short: string;
  // Tooltip / aria-label text.
  title: string;
  // Present only for a token KEYWORD_ICONS recognizes (hexproof
  // today); the badge renders this instead of `short`.
  icon?: string;
  // Which of KeywordBadgeRow's badge shells this is — "protection"
  // gets the blue shield tint, "plain" the neutral dark one (an icon
  // badge like hexproof is plain, same as a card's non-protection
  // keyword icons).
  kind: "plain" | "protection";
}

const PROTECTION_PREFIX = "protection from ";

// CR 702.16j / CR 702.16k: the two qualities the server's closed
// grammar (server/internal/game/protection.go) treats as words rather
// than abbreviating by their first three letters — the same two
// KeywordBadgeRow's protectionShort special-cases for a card.
const EVERYTHING = "everything";
const CHOSEN_PLAYER = "the chosen player";

function titleCase(s: string): string {
  return s.replace(/\S+/g, (word) => word[0].toUpperCase() + word.slice(1).toLowerCase());
}

function protectionShort(quality: string): string {
  if (quality === EVERYTHING) return "ALL";
  if (quality === CHOSEN_PLAYER) return "PLR";
  return quality.slice(0, 3).toUpperCase();
}

// #1200 (CR 119.7, CR 119.8): "your life total can't change". Not a
// token in PlayerView.keywords — that list is engine ABILITY tokens
// and this is not an ability the player has (ADR 0085 Decision 7) —
// so it arrives as its own bool and gets its own badge here, in the
// same visual language, rather than a second badge row.
//
// Last in the list, because the keyword tokens are the ones a reader
// is scanning for when they are wondering why their spell found no
// target, and this one answers a different question.
const LIFE_LOCK_BADGE: PlayerKeywordBadge = {
  key: "life-total-locked",
  short: "LIFE",
  title: "Life total can't change — no gain, no loss, and no paying life",
  kind: "protection",
};

// ADR 0057 (#749, CR 104.3): "can't lose the game" / "can't win the
// game" on a seat, decided on 2026-09-17 (option (b)): a small badge,
// with the sources in its tooltip. A player at -8 life who is still in
// the game is confusing unless the seat says why.
//
// The gates arrive as PlayerView.cant_lose (the causes that can't make
// this seat lose), cant_win, and end_gates (the sources). Like the
// life-total lock they are not ability tokens, so they get their own
// badges here in the same visual language, after everything else.
export interface SeatEndGates {
  cant_lose?: string[];
  cant_win?: boolean;
  end_gates?: GameEndGateView[];
}

const LOSS_CAUSE_TEXT: Record<string, string> = {
  life: "0 or less life",
  empty_draw: "drawing from an empty library",
  poison: "poison",
  commander_damage: "commander damage",
  effect: "an effect",
};

// sourcesText names the gates' sources for a tooltip, marking a
// until-end-of-turn grant: "Platinum Angel, Angel's Grace (this turn)".
function sourcesText(gates: GameEndGateView[], pick: (g: GameEndGateView) => boolean): string {
  const names: string[] = [];
  for (const g of gates) {
    if (!pick(g)) continue;
    const name = g.this_turn ? `${g.source_name} (this turn)` : g.source_name;
    if (!names.includes(name)) names.push(name);
  }
  return names.join(", ");
}

/**
 * The "can't lose" / "can't win" badges for one seat, in that order,
 * or none. A seat whose gate stops only some causes (Phyrexian Unlife's
 * "0 or less life") gets a tooltip that says which.
 */
export function endGateBadges(gates?: SeatEndGates): PlayerKeywordBadge[] {
  if (!gates) return [];
  const all = gates.end_gates ?? [];
  const badges: PlayerKeywordBadge[] = [];
  const causes = gates.cant_lose ?? [];
  if (causes.length > 0) {
    const whole = Object.keys(LOSS_CAUSE_TEXT).every((c) => causes.includes(c));
    const what = whole
      ? "Can't lose the game"
      : `Can't lose the game to ${causes.map((c) => LOSS_CAUSE_TEXT[c] ?? c).join(", ")}`;
    const from = sourcesText(all, (g) => (g.cant_lose?.length ?? 0) > 0);
    const title = from
      ? `${what} — ${from}. Conceding still loses.`
      : `${what}. Conceding still loses.`;
    badges.push({ key: "cant-lose", short: "CAN'T LOSE", title, kind: "protection" });
  }
  if (gates.cant_win) {
    const from = sourcesText(all, (g) => g.cant_win === true);
    const title = from ? `Can't win the game — ${from}` : "Can't win the game";
    badges.push({ key: "cant-win", short: "CAN'T WIN", title, kind: "plain" });
  }
  return badges;
}

/**
 * Turns PlayerView.keywords (and #1200's life_total_locked) into the
 * badges the seat tile renders, one per distinct token, in wire order
 * (derived grants first, then durationed ones — see
 * PlayerView.keywords in protocol.ts).
 */
export function playerKeywordBadges(
  keywords?: string[],
  lifeTotalLocked?: boolean,
  endGates?: SeatEndGates,
): PlayerKeywordBadge[] {
  const seen = new Set<string>();
  const badges: PlayerKeywordBadge[] = [];
  for (const token of keywords ?? []) {
    if (seen.has(token)) continue;
    seen.add(token);

    if (token === "hexproof") {
      badges.push({
        key: token,
        short: "HEX",
        title: "Hexproof",
        icon: KEYWORD_ICONS.hexproof,
        kind: "plain",
      });
      continue;
    }

    if (token.startsWith(PROTECTION_PREFIX)) {
      const quality = token.slice(PROTECTION_PREFIX.length);
      badges.push({
        key: token,
        short: protectionShort(quality),
        title: `Protection from ${titleCase(quality)}`,
        kind: "protection",
      });
      continue;
    }

    // A token this module doesn't recognize yet — shown rather than
    // silently dropped, the same posture KeywordBadgeRow takes for an
    // ability token with no icon registered.
    badges.push({
      key: token,
      short: token.slice(0, 3).toUpperCase(),
      title: titleCase(token),
      kind: "plain",
    });
  }
  if (lifeTotalLocked) badges.push(LIFE_LOCK_BADGE);
  badges.push(...endGateBadges(endGates));
  return badges;
}
