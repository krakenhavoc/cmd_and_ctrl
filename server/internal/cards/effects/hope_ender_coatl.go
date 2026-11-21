package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hope-Ender Coatl — Creature — Eldrazi Snake {2}{U}, 2/2 (Modern
// Horizons 3):
//
//	"Devoid (This card has no color.)
//	 Flash
//	 When you cast this spell, counter target spell an opponent controls
//	 unless they pay {1}.
//	 Flying"
//
// Spell Pierce's pay-unless (CounterUnlessPaid: the question goes to the
// spell's controller and the spell cannot resolve while it is open) as
// World Breaker's cast trigger, so it resolves above the Coatl and works
// even if the Coatl is itself countered.
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell("Hope-Ender Coatl — counter target spell an opponent controls unless they pay {1}",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  id,
				Cost:     "{1}",
				Question: "Hope-Ender Coatl — pay {1} or your spell is countered",
			}.Apply(ctx)
		})
	cast.Targets = TargetSpell("target spell an opponent controls", OpponentControls())
	Register(Spec{
		OracleID:        "57cf576b-5a03-4c60-a531-544df2638323",
		Name:            "Hope-Ender Coatl",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flash", "flying"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}
