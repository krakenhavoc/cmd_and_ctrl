package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jennika, Bad Apple Big Sister — {4}{W} Legendary Creature — Mutant
// Ninja Turtle 3/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"When Jennika enters, create a 2/2 red Mutant creature token.
//	 Plainscycling {2}"
//
// The set's 2/2 red Mutant, and Angel of the Ruins' plainscycling: the
// shared Typecycling constructor (CR 702.29e) searching for a Plains card,
// revealed, then a shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b55da4d3-2cad-4c89-ad3f-7d8397b64d52",
		Name:         "Jennika, Bad Apple Big Sister",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Jennika — create a 2/2 red Mutant creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Mutant"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
	})
}
