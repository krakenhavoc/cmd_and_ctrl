// playerKeywordBadges.ts — #1201, the client half of #1197's player
// protection and hexproof.
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

/**
 * Turns PlayerView.keywords into the badges the seat tile renders,
 * one per distinct token, in wire order (derived grants first, then
 * durationed ones — see PlayerView.keywords in protocol.ts).
 */
export function playerKeywordBadges(keywords?: string[]): PlayerKeywordBadge[] {
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
  return badges;
}
