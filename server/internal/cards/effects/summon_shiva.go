package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summon: Shiva — Enchantment Creature — Saga Elemental {3}{U}{U}, 4/5
// (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Heavenly Strike — Tap target creature an opponent controls.
//	 Put a stun counter on it.
//	 III — Diamond Dust — Draw a card for each tapped creature your
//	 opponents control."
//
// A Saga creature. Chapters I and II are Summon: Valefor's tap and stun
// with a required target an opponent controls. Chapter III counts the
// tapped creatures your opponents control as it resolves.
//
// No simplification.
func init() {
	const name = "Summon: Shiva"
	strike := finB1TapAndStun
	theirs := func() *game.TargetSpec {
		return TargetCreature("target creature an opponent controls", OpponentControls())
	}
	Register(Spec{
		OracleID:     "5846718c-7d31-4d2c-80a2-b1534e2aebc1",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "Heavenly Strike — tap target creature an opponent controls and put a stun counter on it"), theirs(), strike),
			ChapterTriggerTargeting(2, SagaChapterLabel(name, 2, "Heavenly Strike — tap target creature an opponent controls and put a stun counter on it"), theirs(), strike),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "Diamond Dust — draw a card for each tapped creature your opponents control"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := 0
					for _, opp := range ctx.Opponents() {
						n += countControlled(g, opp, func(c game.Card) bool { return c.IsCreature() && c.Tapped })
					}
					return DrawCards{N: n}.Apply(ctx)
				}),
		},
	})
}
