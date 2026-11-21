package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "1a6fce11-e77c-41f1-9ff7-9a93b18071db",
		Name:            "Ebondeath, Dracolich",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		CastableZonesIf: []game.ConditionalCastZone{{Zone: game.ZoneGraveyard, Cond: effects.S82AfrB07NotNamedEbondeathDied}},
	})
}
