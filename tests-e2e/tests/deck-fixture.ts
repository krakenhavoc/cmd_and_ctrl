// Minimal legal Commander deck for E2E setup. We use Kenrith, the
// Returned King (WUBRG color identity) so any basic land is a legal
// mainboard card, and fill the other 99 slots with Plains — basic
// lands are exempt from the singleton rule, so 99x Plains is valid.
//
// Both cards are from widely-distributed printings (Throne of
// Eldraine / every Core Set) so any Scryfall bulk dump has them.
//
// The server validates: 100 cards total, ≥1 legal commander,
// singleton on non-basics, color identity subset, commander-format
// legality. This deck passes all six checks with the smallest
// possible text payload.

export const COMMANDER_NAME = "Kenrith, the Returned King";
export const BASIC_LAND = "Plains";

export function makeCommanderDeck(): string {
  const lines: string[] = [];
  lines.push("Commander:");
  lines.push(`1 ${COMMANDER_NAME}`);
  lines.push("");
  lines.push("Mainboard:");
  lines.push(`99 ${BASIC_LAND}`);
  return lines.join("\n") + "\n";
}
