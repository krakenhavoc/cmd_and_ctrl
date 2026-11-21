package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thelonite Hermit — Creature — Elf Shaman {3}{G}, 1/1:
//
//	"All Saprolings get +1/+1.
//	 Morph {3}{G}{G}
//	 When this creature is turned face up, create four 1/1 green
//	 Saproling creature tokens."
//
// The lord has no controller clause (Verdeloth's TribeFilter without
// YoursOnly). Face down it has no text, so the anthem is off until it
// turns up (CR 708.2a); the four tokens then enter under the anthem.
//
// No simplification.
func init() {
	const label = "Thelonite Hermit — create four 1/1 green Saproling creature tokens"
	effects.Register(effects.Spec{
		OracleID:         "69d2e5f3-2f4d-4a36-bcd5-181164ecd894",
		Name:             "Thelonite Hermit",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Morph("{3}{G}{G}")},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Saproling"}}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisIsTurnedFaceUp(label, effects.Do(effects.CreateToken{Template: effects.B11GreenSaprolingToken(), N: 4})),
		},
	})
}
