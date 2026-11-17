package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarinth Greatwurm — Creature — Wurm {4}{R}{G}, 7/6 (The Brothers' War):
//
//	"Trample
//	 Whenever a land enters, create a tapped Powerstone token."
//
// Any land, under any player's control, so the trigger is Zo-Zu's: it
// watches the ETB event and the Wurm's controller gets the token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5b5b9390-bb3a-4f71-809a-cfe2a6566b59",
		Name:            "Sarinth Greatwurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b39LandEnteredController(ev, g)
				return ok
			},
			Key:    "Sarinth Greatwurm — create a tapped Powerstone token",
			Effect: Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1}),
		}},
	})
}
