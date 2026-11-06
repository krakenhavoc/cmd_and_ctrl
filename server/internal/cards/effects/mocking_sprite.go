package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "013ca531-1ff7-4efc-a280-4bc3ab574072",
		Name:            "Mocking Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
	})
}
