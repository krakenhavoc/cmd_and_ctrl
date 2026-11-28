package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harvest Mage — Creature — Human Spellshaper {G}, 1/1 (Nemesis, S82 OLD batch
// 679, #5618):
//
//	"{G}, {T}, Discard a card: Until end of turn, if you tap a land for mana,
//	 it produces one mana of a color of your choice instead of any other type
//	 and amount."
//
// The cost is a Spellshaper's: {G}, {T} (so summoning sickness applies, CR
// 302.6) and a discard. The effect is the #6210 seam, a replacement on the
// mana a land of yours makes while it is tapped for mana, until end of turn
// (CR 614.1a, 106.12b): one mana of a colour chosen at the tap in place of
// whatever and however much it would make. It is not a mana ability itself,
// so it uses the stack.
//
// No simplification.
func init() {
	const label = "Harvest Mage — lands you tap for mana this turn produce one mana of a color of your choice"
	effects.Register(effects.Spec{
		OracleID:     "6f82b01d-8975-42a1-87e2-99782272bbea",
		Name:         "Harvest Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Discard a card: Until end of turn, if you tap a land for mana, it produces one mana of a color of your choice instead of any other type and amount.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.LandsYouTapProduceOneManaOfAColorOfYourChoiceThisTurnForEffect(item.SourceCardID, item.Controller, label)
				return nil
			},
		}},
	})
}
