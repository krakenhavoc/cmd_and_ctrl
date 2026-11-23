package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flinthoof Boar — Creature — Boar {1}{G}, 2/2 (Magic 2013, #5568):
//
//	"This creature gets +1/+1 as long as you control a Mountain.
//	 {R}: This creature gains haste until end of turn. (It can attack and
//	 {T} this turn.)"
//
// Kird Ape's land-type pump and Thornling's haste for mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acb58839-2b26-4724-9e9b-b49b8217b13f",
		Name:         "Flinthoof Boar",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control a Mountain.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfWhileYouControlA(effects.Subtype("Mountain")),
			Apply:     effects.PumpBy(1, 1),
		}},
		Activated: []effects.ActivatedAbility{effects.ConB1ThisGainsRow("{R}", "haste", "Flinthoof Boar", game.AnswerCombatGrant)},
	})
}
