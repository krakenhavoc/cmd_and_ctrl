// matchLength.ts — the match length picker's two choices (ADR 0222 §3).
//
// A table of a format that plays matches is best-of-one unless its host
// asks for best-of-three. The server (game/match.go) is the authority
// on which formats play matches; `playsMatches` mirrors it so the
// picker is not offered where the answer would be a 400.

import { formatOf } from "./formats";
import { L } from "./labels";

export interface MatchLength {
  bestOf: 1 | 3;
  name: string;
  blurb: string;
}

export const MATCH_LENGTHS: readonly MatchLength[] = [
  { bestOf: 1, name: L.matchBestOfOne, blurb: "one game decides it" },
  {
    bestOf: 3,
    name: L.matchBestOfThree,
    blurb: "first to two games, with sideboarding between them",
  },
];

/** Whether a table of the format plays a match. Commander does not. */
export function playsMatches(format: string | undefined | null): boolean {
  return formatOf(format).id === "standard";
}

/** The chip a table of this match length wears in a list, or "" for none. */
export function matchLengthChip(bestOf: number | undefined | null): string {
  return bestOf === 3 ? L.matchBestOfThree : "";
}
