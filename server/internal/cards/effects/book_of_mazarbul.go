package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Book of Mazarbul — Enchantment — Saga {2}{R} (The Lord of the Rings:
// Tales of Middle-earth, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Amass Orcs 1.
//	 II — Amass Orcs 2.
//	 III — Creatures you control get +1/+0 and gain menace until end of
//	 turn."
//
// I and II are the keyword action (CR 701.47). III fixes its set of
// creatures as it resolves (CR 611.2c): one effect, +1/+0 and menace
// together, so a creature that arrives afterwards gets neither.
//
// No simplification.
func init() {
	const name = "Book of Mazarbul"
	Register(Spec{
		OracleID:     "4598c221-b111-4bde-b4dc-673186c4efe6",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Amass: 2}, // the chapter riders are the card's own (#4978)
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "amass Orcs 1"), Do(Amass{Subtype: "Orc", N: 1})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "amass Orcs 2"), Do(Amass{Subtype: "Orc", N: 2})),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "creatures you control get +1/+0 and gain menace until end of turn"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1,
						Label: "Book of Mazarbul — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"menace"},
						Label: "Book of Mazarbul — menace"}.Apply(ctx)
				}),
		},
	})
}
