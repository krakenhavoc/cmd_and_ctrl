package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myrel, Shield of Argive — Legendary Creature — Human Soldier {3}{W}, 3/4
// (The Brothers' War):
//
//	"During your turn, your opponents can't cast spells or activate
//	 abilities of artifacts, creatures, or enchantments.
//	 Whenever Myrel attacks, create X 1/1 colorless Soldier artifact
//	 creature tokens, where X is the number of Soldiers you control."
//
// The static is Grand Abolisher's two turn-scoped restrictions. X is counted
// as the trigger resolves, Myrel herself included (CR 608.2h); a count of zero
// makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "503094aa-6f8f-42c1-ad5a-9305934e5f0c",
		Name:         "Myrel, Shield of Argive",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.OpponentsCantCastDuringYourTurn(
				"During your turn, your opponents can't cast spells.", nil),
		},
		ActivationRestrictions: []game.ActivationRestriction{
			effects.OpponentsSourcesCantActivateDuringYourTurn(
				"During your turn, your opponents can't activate abilities of artifacts, creatures, or enchantments.",
				effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment()), false),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Myrel — create X 1/1 colorless Soldier artifact creature tokens, where X is the number of Soldiers you control",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.TrekPermanentsWithSubtypeYouControl(g, item, item.Controller, "Soldier")
					if x <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Soldier artifact"), N: x}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
