package cardsex

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extravagant Spirit — Creature — Spirit {3}{U}, 4/4 (Mercadian Masques, S82
// OLD b683, #5619):
//
//	"Flying
//	 At the beginning of your upkeep, sacrifice this creature unless you pay
//	 {1} for each card in your hand."
//
// Flying is a printed keyword. Fettergeist's upkeep payment with the amount
// counted as the trigger resolves: your hand size, and with an empty hand
// there is nothing to pay and nothing is sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6470afc0-2f67-454a-8a5b-d8dfea4b84ce",
		Name:            "Extravagant Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Extravagant Spirit — sacrifice this creature unless you pay {1} for each card in your hand",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
						return nil
					}
					cost := "{" + strconv.Itoa(p.Hand.Size()) + "}"
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     cost,
						Question: "Extravagant Spirit — pay " + cost + " or sacrifice Extravagant Spirit?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
