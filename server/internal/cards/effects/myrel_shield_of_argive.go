package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "503094aa-6f8f-42c1-ad5a-9305934e5f0c",
		Name:         "Myrel, Shield of Argive",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			OpponentsCantCastDuringYourTurn(
				"During your turn, your opponents can't cast spells.", nil),
		},
		ActivationRestrictions: []game.ActivationRestriction{
			OpponentsSourcesCantActivateDuringYourTurn(
				"During your turn, your opponents can't activate abilities of artifacts, creatures, or enchantments.",
				Or(Artifact(), Creature(), Enchantment()), false),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Myrel — create X 1/1 colorless Soldier artifact creature tokens, where X is the number of Soldiers you control",
				func(g *game.Game, item *game.StackItem) error {
					x := trekPermanentsWithSubtypeYouControl(g, item, item.Controller, "Soldier")
					if x <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 colorless Soldier artifact"), N: x}.Apply(NewContext(g, item))
				}),
		},
	})
}
