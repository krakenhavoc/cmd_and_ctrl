package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ebondeath, Dracolich — Legendary Creature — Zombie Dragon {2}{B}{B}, 5/2
// (Adventures in the Forgotten Realms, S82):
//
//	"Flash
//	 Flying
//	 Ebondeath enters tapped.
//	 You may cast this card from your graveyard if a creature not named
//	 Ebondeath, Dracolich died this turn."
//
// Undead Sprinter's graveyard cast (Spec.CastableZonesIf) over a new
// table-wide tally of the names creatures had as they died
// (TurnTally.DiedNames), so a copy of Ebondeath dying opens nothing but
// any other creature does. The entry is Oathsworn Vampire's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a6fce11-e77c-41f1-9ff7-9a93b18071db",
		Name:            "Ebondeath, Dracolich",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
		CastableZonesIf: []game.ConditionalCastZone{{Zone: game.ZoneGraveyard, Cond: s82AfrB07NotNamedEbondeathDied}},
	})
}
