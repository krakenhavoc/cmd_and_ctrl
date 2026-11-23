package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestial Crusader — Creature — Spirit {2}{W}, 2/2 (Time Spiral,
// #5596):
//
//	"Flash
//	 Split second (As long as this spell is on the stack, players can't
//	 cast spells or activate abilities that aren't mana abilities.)
//	 Flying
//	 Other white creatures get +1/+1."
//
// Flash, split second and flying are printed (Quagnoth is the split-
// second creature before it). The anthem reaches every other white
// creature on the battlefield, an opponent's included
// (tspB2OtherWhiteCreatures).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "af06a811-7a57-4f68-9d87-7903e043f36d",
		Name:            "Celestial Crusader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "split second", "flying"},
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.TspB2OtherWhiteCreatures, 1, 1),
		},
	})
}
