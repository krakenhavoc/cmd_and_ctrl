package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Choco/Mog — Enchantment Creature — Saga Bird Moogle {2}{W},
// 3/3 (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I, II, III, IV — Stampede! — Other creatures you control get +1/+0
//	 until end of turn."
//
// A Saga creature. Each chapter is Founding of Omashu's chapter III
// without the Saga itself: the creatures you control as it resolves
// (CR 611.2c), other than this object (CR 400.7), get +1/+0.
//
// No simplification.
func init() {
	const name = "Summon: Choco/Mog"
	stampede := func(g *game.Game, item *game.StackItem) error {
		return effects.BoostUntilEOT{
			Match: effects.And(effects.Creature(), effects.YouControl(), effects.TrekB11OtherThanSourceObject(item)),
			Power: 1,
			Label: name + " — other creatures you control get +1/+0",
		}.Apply(effects.NewContext(g, item))
	}
	var chapters []game.TriggeredAbility
	for n := 1; n <= 4; n++ {
		chapters = append(chapters, effects.ChapterTrigger(n,
			effects.SagaChapterLabel(name, n, "Stampede! — other creatures you control get +1/+0 until end of turn"), stampede))
	}
	effects.Register(effects.Spec{
		OracleID:     "8b4bfdc5-8c0a-40c3-abbf-ada4a536dc34",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered:    chapters,
	})
}
