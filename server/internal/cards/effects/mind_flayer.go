package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Flayer — Creature — Horror {3}{U}{U}, 3/3 (#4092):
//
//	"Dominate Monster — When this creature enters, gain control of
//	 target creature for as long as you control this creature."
//
// Sower of Temptation's shape with the other CR 611.2b condition:
// "for as long as you CONTROL this creature" (DurationWhileYouControlSource,
// Thancred Waters' duration), so the creature goes home when the Flayer
// leaves AND when somebody else takes the Flayer. A Flayer that has
// already left, or changed hands, by the time the trigger resolves
// takes nothing: the effect's condition is false as it would begin, so
// it never begins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8dd473c7-afd9-40f1-bb8e-0a00f4687318",
		Name:         "Mind Flayer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Mind Flayer — Dominate Monster: gain control of target creature for as long as you control this creature", mindFlayerDominate),
				TargetCreature("target creature")),
		},
	})
}

func mindFlayerDominate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		d, ok := DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
		if !ok {
			return nil
		}
		return GainControl{
			Target:     t.ID,
			Controller: item.Controller,
			Duration:   d,
			Label:      "Mind Flayer — control for as long as you control Mind Flayer",
		}.Apply(ctx)
	}
	return nil
}
