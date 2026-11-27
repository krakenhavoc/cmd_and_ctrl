package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mire Kavu — Creature — Kavu {3}{R}, 3/2 (Planeshift, S82 OLD #5615):
//
//	"This creature gets +1/+1 as long as you control a Swamp."
//
// Kird Ape's conditional pump with a land subtype: any permanent you control
// with the Swamp subtype counts, a nonbasic Swamp included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bc58096-7925-4870-bb95-58cce51ea8e7",
		Name:         "Mire Kavu",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Swamp.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Swamp")),
			Apply:     effects.PumpBy(1, 1),
		}},
	})
}
