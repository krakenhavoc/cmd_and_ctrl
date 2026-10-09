package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Candlekeep Sage — Legendary Enchantment — Background {2}{U} (EDHREC
// rank 5612):
//
//	"Commander creatures you own have "When this creature enters or
//	 leaves the battlefield, draw a card.""
//
// An ADR 0093 grant to each commander creature you own, under anyone's
// control (choose_a_background.go). The creature has the ability as it
// enters while the Background is on the battlefield, so its entering
// triggers it (CR 603.6a), and its leaving triggers it from its
// last-known abilities (CR 603.10a), so dying, a bounce and a trip to
// the command zone each draw. "You" is the creature's controller then.
// The Background entering after the commander does not draw.
//
// No simplification.
const candlekeepSageGrant = "candlekeep-sage/draw"

func init() {
	Register(Spec{
		OracleID:     "f9516b60-36dd-4f8a-bebb-ac1bc0961f07",
		Name:         "Candlekeep Sage",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: candlekeepSageGrant,
			Triggered: []game.TriggeredAbility{
				TriggerWithPurpose(WhenThisEntersOrLeaves("Candlekeep Sage — draw a card", func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				}), game.Purpose{Draws: 1}),
			},
			Text: "When this creature enters or leaves the battlefield, draw a card.",
		}},
		Static: []game.StaticAbility{grantToCommanderCreaturesYouOwn(candlekeepSageGrant)},
	})
}
