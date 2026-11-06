package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Founding of Omashu — Enchantment — Saga {2}{R} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create two 1/1 white Ally creature tokens.
//	 II — You may discard a card. If you do, draw a card.
//	 III — Creatures you control get +1/+0 until end of turn."
//
// Chapter I is the set's 1/1 white Ally token row, two of them (The Cave
// of Two Lovers' chapter I). Chapter II is Rubble Rouser's optional loot:
// the draw only follows a discard that happened. Chapter III pumps the
// creatures you control as it resolves (CR 611.2c), so one that arrives
// later is not pumped. The engine sacrifices the Saga after III (CR 714.4).
//
// No simplification.
func init() {
	const name = "Founding of Omashu"
	lootLabel := SagaChapterLabel(name, 2, "you may discard a card; if you do, draw a card")
	Register(Spec{
		OracleID:     "ea58fb10-3d70-4c25-b317-615dfb20a48d",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create two 1/1 white Ally creature tokens"),
				Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 2})),
			ChapterTrigger(2, lootLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, lootLabel, drawOneIfDiscarded)(NewContext(g, item))
			}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "creatures you control get +1/+0 until end of turn"),
				Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: "Founding of Omashu — creatures you control get +1/+0"})),
		},
	})
}
