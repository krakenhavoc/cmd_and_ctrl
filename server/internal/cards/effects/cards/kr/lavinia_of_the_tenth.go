package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lavinia of the Tenth — Legendary Creature — Human Soldier {3}{W}{U},
// 4/4:
//
//	"Protection from red
//	 When Lavinia enters, detain each nonland permanent your opponents
//	 control with mana value 4 or less. (Until your next turn, those
//	 permanents can't attack or block and their activated abilities
//	 can't be activated.)"
//
// Protection from red is a printed keyword on the engine's protection
// seam (ADR 0072): damage, enchanting, blocking and targeting by red
// sources all follow from the token.
//
// The detain is the Detain primitive: a restriction record pinned to
// the permanents that matched as the trigger resolved, ending as
// Lavinia's controller's next turn begins. The set is decided at
// resolution, not when the trigger goes on the stack, so a permanent an
// opponent plays in response with a mana value of 4 or less is
// detained, and one that left is not. Tokens have mana value 0 and are
// detained. The trigger targets nothing, so hexproof and protection do
// not stop it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "55df6fa3-3d24-46d5-b325-099336c3f0fd",
		Name:            "Lavinia of the Tenth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lavinia of the Tenth — detain each nonland permanent your opponents control with mana value 4 or less",
				func(g *game.Game, item *game.StackItem) error {
					return effects.Detain{
						Match: effects.And(effects.OpponentControls(), effects.Nonland(), effects.ManaValueLE(4)),
						Label: "Lavinia of the Tenth — detain",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
