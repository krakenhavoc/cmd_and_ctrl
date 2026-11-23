package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sister Repentia — Creature — Human Warrior {3}{W}{B}, 5/1:
//
//	"Martyrdom — When this creature dies, you gain 2 life and draw two
//	 cards.
//	 Miracle {W}{B}"
//
// Miracle is the alternative cost offered when it is the first card drawn
// this turn (CR 702.94); the dies trigger is a plain leaves-the-battlefield
// trigger on itself.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "24e9a1cb-b58b-491f-9d08-3f039a86851e",
		Name:             "Sister Repentia",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{W}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Sister Repentia — you gain 2 life and draw two cards",
				effects.Do(effects.GainLife{Amount: 2}, effects.DrawCards{N: 2})),
		},
	})
}
