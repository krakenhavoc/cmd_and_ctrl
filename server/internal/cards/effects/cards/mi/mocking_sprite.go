package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mocking Sprite — Creature — Faerie Rogue {2}{U}, 2/1 (Foundations,
// #4465):
//
//	"Flying
//	 Instant and sorcery spells you cast cost {1} less to cast."
//
// Goblin Electromancer's reduction (generic only) with flying, which
// rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "013ca531-1ff7-4efc-a280-4bc3ab574072",
		Name:            "Mocking Sprite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
	})
}
