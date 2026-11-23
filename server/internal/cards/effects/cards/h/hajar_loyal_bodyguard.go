package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hajar, Loyal Bodyguard — Legendary Creature — Human Soldier {R}{G}, 3/3
// (The Brothers' War):
//
//	"Sacrifice Hajar: Legendary creatures you control get +1/+0 and gain
//	 indestructible until end of turn."
//
// Sacrificing is the cost, so Hajar is gone before it resolves and is not
// among the creatures (CR 118.3). The legendary creatures are those you
// control as it resolves (CR 611.2c), read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63584001-7f58-42ee-b072-914baa66c78a",
		Name:         "Hajar, Loyal Bodyguard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Sacrifice Hajar: Legendary creatures you control get +1/+0 and gain indestructible until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump | game.AnswerCombatGrant,
				Pump:    &game.Pump{Power: 1, Keywords: []string{"indestructible"}},
			},
			Cost: effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				match := effects.And(effects.Creature(), effects.YouControl(), effects.Legendary())
				if err := (effects.BoostUntilEOT{Match: match, Power: 1, Label: "Hajar — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"indestructible"}, Label: "Hajar — indestructible"}.Apply(ctx)
			},
		}},
	})
}
