package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thieving Magpie — Creature — Bird {2}{U}{U}, 1/3 (Urza's Destiny,
// #5621):
//
//	"Flying
//	 Whenever this creature deals damage to an opponent, draw a card."
//
// Thieving Otter's trigger on a 1/3 flier. Any damage counts, combat or
// not, but only to an opponent: damage to a creature or planeswalker
// draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f074636-6a36-46a3-b872-add79e19aa8b",
		Name:            "Thieving Magpie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent("Thieving Magpie — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
