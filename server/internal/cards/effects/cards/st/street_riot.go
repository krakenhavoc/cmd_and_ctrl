package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Street Riot — Enchantment {4}{R} (Guilds of Ravnica, #5525):
//
//	"During your turn, creatures you control get +1/+0 and have trample."
//
// Two statics read live (CR 611.3a): a layer 7c anthem and a layer 6 keyword
// grant, both over the creatures its controller controls and both gated on its
// controller being the active player (Memorial Team Leader's and Bedrock
// Tortoise's gates). A creature that enters later in your turn gets both, and
// neither applies on any other player's turn.
//
// No simplification.
func init() {
	yoursOnYourTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() && effects.IsActivePlayer(g, source.Controller)
	}
	trample := effects.KeywordGrant(yoursOnYourTurn, "trample")
	trample.Label = "During your turn, creatures you control have trample."
	anthem := effects.B16Anthem(yoursOnYourTurn, 1, 0)
	anthem.Label = "During your turn, creatures you control get +1/+0."
	effects.Register(effects.Spec{
		OracleID:     "526db754-abb6-450c-bead-81dd1e1c347c",
		Name:         "Street Riot",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{anthem, trample},
	})
}
