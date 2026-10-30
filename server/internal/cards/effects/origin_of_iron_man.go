package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6abb489e-cc62-42d7-b898-3e5095037e75",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "tap up to one target creature; it doesn't untap while this Saga remains"),
				TargetCreature("up to one target creature").WithCount(0, 1), originOfIronManHold),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "draw two cards"),
				Do(DrawCards{N: 2})),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "you may put an artifact card with mana value 5 or less from your hand onto the battlefield"),
				func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Match:    And(Artifact(), ManaValueLE(5)),
						Optional: true,
						Label:    "Origin of Iron Man — you may put an artifact card with mana value 5 or less from your hand onto the battlefield",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// originOfIronManHold taps the chosen creature and holds it tapped for as
// long as the Saga remains on the battlefield. The tap always happens; the
// hold is recorded only if the Saga is still the object that made the
// chapter.
func originOfIronManHold(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := holdTargetIDs(ctx)
	if err := tapEach(ctx, ids); err != nil {
		return err
	}
	d, ok := DurationWhileSourceRemains(ctx, ctx.Source())
	if !ok {
		return nil
	}
	return DoesntUntapWhile{Targets: ids, Duration: d}.Apply(ctx)
}
