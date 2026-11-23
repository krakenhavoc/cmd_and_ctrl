package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Origin of Iron Man — {4}{U} Enchantment — Saga (Marvel Super Heroes
// Commander, #4455):
//
//	"I — Tap up to one target creature. It doesn't untap during its
//	     controller's untap step for as long as this Saga remains on the
//	     battlefield.
//	 II — Draw two cards.
//	 III — You may put an artifact card with mana value 5 or less from your
//	       hand onto the battlefield."
//
// Chapter I is the tap-and-hold of Rust Tick timed by the Saga's presence
// (Old Fat Spider Can't See Me's duration): a Saga that was flickered while
// the chapter waited is not the object that made it, and holds nothing
// (CR 611.2b, CR 400.7). "Up to one" with no target does nothing. Chapter
// III is a put, not a cast.
//
// No simplification.
func init() {
	const name = "Origin of Iron Man"
	effects.Register(effects.Spec{
		OracleID:     "6abb489e-cc62-42d7-b898-3e5095037e75",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "tap up to one target creature; it doesn't untap while this Saga remains"),
				effects.TargetCreature("up to one target creature").WithCount(0, 1), originOfIronManHold),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "draw two cards"),
				effects.Do(effects.DrawCards{N: 2})),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "you may put an artifact card with mana value 5 or less from your hand onto the battlefield"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromHandOntoBattlefield{
						Match:    effects.And(effects.Artifact(), effects.ManaValueLE(5)),
						Optional: true,
						Label:    "Origin of Iron Man — you may put an artifact card with mana value 5 or less from your hand onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// originOfIronManHold taps the chosen creature and holds it tapped for as
// long as the Saga remains on the battlefield. The tap always happens; the
// hold is recorded only if the Saga is still the object that made the
// chapter.
func originOfIronManHold(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ids := effects.HoldTargetIDs(ctx)
	if err := effects.TapEach(ctx, ids); err != nil {
		return err
	}
	d, ok := effects.DurationWhileSourceRemains(ctx, ctx.Source())
	if !ok {
		return nil
	}
	return effects.DoesntUntapWhile{Targets: ids, Duration: d}.Apply(ctx)
}
