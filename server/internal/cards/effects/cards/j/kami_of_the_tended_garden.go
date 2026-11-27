package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of the Tended Garden — Creature — Spirit {3}{G}, 4/4 (Saviors of
// Kamigawa, #5601):
//
//	"At the beginning of your upkeep, sacrifice this creature unless you pay
//	 {G}.
//	 Soulshift 3 (When this creature dies, you may return target Spirit card
//	 with mana value 3 or less from your graveyard to your hand.)"
//
// Krosan Cloudscraper's upkeep payment: pay {G} or sacrifice it. The
// sacrificed Kami dies, so its Soulshift triggers (CR 702.46a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2770ebf-24e3-4787-956b-ca01aaa50a6f",
		Name:         "Kami of the Tended Garden",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Kami of the Tended Garden — sacrifice unless you pay {G}", func(g *game.Game, item *game.StackItem) error {
				return effects.UpkeepPayUnless{
					Chooser:  item.Controller,
					Cost:     "{G}",
					Question: "Kami of the Tended Garden — pay {G} or sacrifice Kami of the Tended Garden?",
					Else:     effects.UpkeepSacrificeSource,
				}.Apply(effects.NewContext(g, item))
			}),
			effects.Soulshift(3, "Kami of the Tended Garden"),
		},
	})
}
