package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gadwick, the Wizened — Legendary Creature — Human Wizard {X}{U}{U}{U},
// 3/3 (Throne of Eldraine, #5516):
//
//	"When Gadwick enters, draw X cards.
//	 Whenever you cast a blue spell, tap target nonland permanent an
//	 opponent controls."
//
// X is the value announced for the spell (CR 107.3m), carried on the enters
// trigger as it is built (WhenThisEntersWithCastX). The cast trigger targets
// a nonland permanent an opponent controls and taps it as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7530ecd-e22f-41e7-a76b-de32809901c8",
		Name:         "Gadwick, the Wizened",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersWithCastX("Gadwick — draw X cards", func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{N: item.Params.Amount}.Apply(effects.NewContext(g, item))
			}),
			effects.Targeting(
				effects.WheneverYouCast(effects.OfColor("U"), "Gadwick — tap target nonland permanent an opponent controls",
					effects.B36TapChosenCreature),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls())),
		},
	})
}
