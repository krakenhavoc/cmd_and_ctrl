package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adept Watershaper — Creature — Merfolk Cleric {2}{W}, 3/4 (Lorwyn
// Eclipsed, #4458):
//
//	"Other tapped creatures you control have indestructible."
//
// Saryth, the Viper's Fang's shape: a layer-6 keyword grant whose
// predicate reads the tapped bit, so a creature has indestructible while
// it is tapped (an attacker without vigilance, a creature tapped for
// convoke or a cost) and loses it the moment it untaps. The Watershaper
// itself never has it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "602cce3e-7033-47d6-aec1-d755f234f270",
		Name:         "Adept Watershaper",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.SarythOtherCreatureYouControl(true), "indestructible"),
		},
	})
}
