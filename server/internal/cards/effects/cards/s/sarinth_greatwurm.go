package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "5b5b9390-bb3a-4f71-809a-cfe2a6566b59",
		Name:            "Sarinth Greatwurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B39LandEnteredController(ev, g)
				return ok
			},
			Key:    "Sarinth Greatwurm — create a tapped Powerstone token",
			Effect: effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1}),
		}},
	})
}
