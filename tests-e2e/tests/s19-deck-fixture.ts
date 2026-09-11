// S19 e2e deck fixtures. Two 100-card Commander decks designed
// around the trigger-harvester surface:
//
//   - The caster deck holds one copy of every S19 sub-PR 3 card
//     (Mulldrifter, Reclamation Sage, Acidic Slime, Eternal Witness,
//     Solemn Simulacrum) plus support cards that seed the graveyard
//     (Lightning Bolt) and a few generic permanents. Filler is
//     Forest — Kenrith's WUBRG identity legalises any basic — with
//     exactly one Island so that a "search your library for a basic
//     land" effect has a choice that can be named from outside.
//
//   - The opponent deck holds the destroy-target permanents
//     (Sol Ring artifact, Glorious Anthem enchantment) plus filler
//     basics. Acidic Slime's land target comes from the basic spread.
//
// Both decks resolve through the live Scryfall index, so the
// embedded card names match the canonical oracle entries. If the
// index is missing any of these names the upload step will reject
// the deck with a clear error — the test fails at setup rather
// than mid-scenario.

export const COMMANDER_NAME = "Kenrith, the Returned King";

// Caster deck — Player 1 admin-moves these cards onto the
// battlefield to drive each S19 trigger. Library order doesn't
// matter for the move-card path, but the card has to live on
// the seat at game start.
const CASTER_NON_BASICS = [
  "Mulldrifter",
  "Reclamation Sage",
  "Acidic Slime",
  "Eternal Witness",
  "Solemn Simulacrum",
  "Lightning Bolt",
  "Sol Ring",
  "Glorious Anthem",
  // S22 — one lone Island among ninety Forests. It is a basic, not
  // a non-basic, but it belongs in this list because it is a SINGLE
  // copy rather than filler.
  //
  // It exists so the search chooser has an answer that is visible
  // from the outside. The engine used to take the first match in
  // library order; it now asks the searcher which basic they want.
  // A fetched Forest cannot tell those two behaviours apart — there
  // are ninety of them and library order would produce one anyway.
  // A fetched Island can: it is proof the PICK decided it.
  "Island",
];

// Opponent deck — these permanents are the destroy targets for
// Reclamation Sage / Acidic Slime. Sol Ring and Glorious Anthem
// double as artifact + enchantment; Plains supplies the land row.
// Smothering Tithe (sub-PR 6) sits here too: it's the opponent's
// card so the CASTER is the one taxed on a draw.
const OPPONENT_NON_BASICS = ["Sol Ring", "Glorious Anthem", "Plains", "Smothering Tithe"];

function buildDeck(nonBasics: string[], filler: string, fillerCount: number): string {
  const lines: string[] = [];
  lines.push("Commander:");
  lines.push(`1 ${COMMANDER_NAME}`);
  lines.push("");
  lines.push("Mainboard:");
  for (const name of nonBasics) {
    lines.push(`1 ${name}`);
  }
  lines.push(`${fillerCount} ${filler}`);
  return lines.join("\n") + "\n";
}

// makeS19CasterDeck builds the 100-card caster deck. 1 commander +
// 8 non-basics + 1 Island + 90 Forest = 100 mainboard. Forest is the filler
// because every S19 caster card has a green color identity (Kenrith
// is WUBRG so any basic is legal, but Forest is the most thematic
// for the green-heavy S19 pool — Reclamation Sage / Acidic Slime /
// Eternal Witness / Solemn Simulacrum).
export function makeS19CasterDeck(): string {
  return buildDeck(CASTER_NON_BASICS, "Forest", 90);
}

// makeS19OpponentDeck builds the 100-card opponent deck. 1
// commander + 4 non-basics + 95 Plains = 100 mainboard. Plains
// gives Acidic Slime a default land target without needing the
// other player to admin-place a basic.
export function makeS19OpponentDeck(): string {
  return buildDeck(OPPONENT_NON_BASICS, "Plains", 95);
}

// CARDS surfaces the named cards we'll look up by name from the
// game state in the spec. Putting the strings here keeps the deck
// shape and the test expectations on the same source of truth.
export const CARDS = {
  Commander: COMMANDER_NAME,
  Mulldrifter: "Mulldrifter",
  ReclamationSage: "Reclamation Sage",
  AcidicSlime: "Acidic Slime",
  EternalWitness: "Eternal Witness",
  SolemnSimulacrum: "Solemn Simulacrum",
  LightningBolt: "Lightning Bolt",
  SolRing: "Sol Ring",
  GloriousAnthem: "Glorious Anthem",
  SmotheringTithe: "Smothering Tithe",
  Treasure: "Treasure",
  Plains: "Plains",
  Forest: "Forest",
  Island: "Island",
} as const;
