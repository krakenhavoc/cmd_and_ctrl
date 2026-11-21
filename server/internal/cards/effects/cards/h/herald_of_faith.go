package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of Faith — Creature — Angel {3}{W}{W}, 4/3 (Foundations, #4465):
//
//	"Flying
//	 Whenever this creature attacks, you gain 2 life."
//
// Shopkeeper's Bane's attack trigger on a flier: WheneverThisAttacks with
// a 2-life GainLife, which a life-gain replacement sees as usual.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cb97da84-1d13-4795-b68a-2bf111a50067",
		Name:            "Herald of Faith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Herald of Faith — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
