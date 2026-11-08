package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summon: Esper Ramuh — Enchantment Creature — Saga Wizard {2}{R}{R},
// 3/3 (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Judgment Bolt — This creature deals damage equal to the number
//	 of noncreature, nonland cards in your graveyard to target creature
//	 an opponent controls.
//	 II, III — Wizards you control get +1/+0 until end of turn."
//
// A Saga creature. Chapter I counts the graveyard as it resolves (Flame
// On!'s count) and the Saga deals the damage. Chapters II and III pump
// the Wizards you control as each resolves (CR 611.2c), the Saga itself
// among them; "Wizards" is read through the card's text (CR 612.2).
//
// No simplification.
func init() {
	const name = "Summon: Esper Ramuh"
	wizards := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return BoostUntilEOT{
			Match: And(Creature(), YouControl(), HasSubtype(ctx.TextSubtype("Wizard"))),
			Power: 1,
			Label: name + " — Wizards you control get +1/+0",
		}.Apply(ctx)
	}
	Register(Spec{
		OracleID:     "0d054301-8465-4a4c-91e7-9a28dda92e57",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "Judgment Bolt — damage equal to the noncreature, nonland cards in your graveyard to target creature an opponent controls"),
				TargetCreature("target creature an opponent controls", OpponentControls()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					n := len(graveyardIDs(g, ctx.Controller(), func(c game.Card) bool {
						return !c.IsCreature() && !c.IsLand()
					}))
					if n == 0 {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: id, Amount: n}.Apply(ctx)
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "Wizards you control get +1/+0 until end of turn"), wizards),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "Wizards you control get +1/+0 until end of turn"), wizards),
		},
	})
}
