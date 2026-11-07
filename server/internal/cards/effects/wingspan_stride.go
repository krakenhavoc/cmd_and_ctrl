package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e75cee5d-ab02-4227-b36e-61c5d2f564a8",
		Name:         "Wingspan Stride",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Return this Aura to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{2}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
