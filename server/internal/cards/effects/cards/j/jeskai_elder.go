package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeskai Elder — Creature — Human Monk {1}{U}, 1/2 (S82 GAP, #5948):
//
//	"Prowess
//	 Whenever this creature deals combat damage to a player, you may draw a
//	 card. If you do, discard a card."
//
// Stadium Tidalmage's optional loot (MayChoice with Draws, then the discard)
// behind the combat-damage trigger.
func init() {
	const label = "Jeskai Elder — you may draw a card; if you do, discard a card"
	effects.Register(effects.Spec{
		OracleID:        "dfbd5ad1-0f68-4219-95c4-98f674f321be",
		Name:            "Jeskai Elder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(label, func(g *game.Game, item *game.StackItem) error {
				return effects.MayChoice{
					Question: "Jeskai Elder — draw a card, then discard a card?",
					Draws:    1,
					YesLabel: "Draw",
					NoLabel:  "Decline",
					Then:     effects.CelestusAnswer,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
