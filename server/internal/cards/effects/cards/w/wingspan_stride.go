package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wingspan Stride — Enchantment — Aura {U} (Tarkir: Dragonstorm, #4463):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flying.
//	 {2}{U}: Return this Aura to its owner's hand."
//
// The two statics are PumpAttached and GrantToAttached. The return is
// Flickering Ward's: it may be activated at instant speed by the Aura's
// controller, and the Aura returns to its owner's hand if it is still on
// the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e75cee5d-ab02-4227-b36e-61c5d2f564a8",
		Name:         "Wingspan Stride",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
