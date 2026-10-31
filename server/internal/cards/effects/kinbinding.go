package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kinbinding — Enchantment {3}{W}{W} (Lorwyn Eclipsed, #4458):
//
//	"Creatures you control get +X/+X, where X is the number of creatures
//	 that entered the battlefield under your control this turn.
//	 At the beginning of combat on your turn, create a 1/1 green and
//	 white Kithkin creature token."
//
// X is the turn's entry tally (EnteredWithCardTypeThisTurn, "creature"):
// every creature that entered under your control this turn as it entered,
// tokens included, whether or not it is still on the battlefield, and on
// any player's turn. It is read on every layer pass (layer 7c), so it
// grows as creatures enter and drops to zero when the turn ends. The
// combat token counts toward it, so on your turn the team is at least
// +1/+1 by the time attackers are declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da35de40-6fde-4018-bda8-b0ae91ed872a",
		Name:         "Kinbinding",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Creatures you control get +X/+X, where X is the number of creatures that entered the battlefield under your control this turn.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := g.EnteredWithCardTypeThisTurn(source.Controller, "creature")
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Kinbinding — create a 1/1 green and white Kithkin creature token",
				Do(CreateToken{Template: TokenCard("1/1 green and white Kithkin"), N: 1})),
		},
	})
}
