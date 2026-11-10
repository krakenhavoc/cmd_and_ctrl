package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Measure of a Man — Enchantment — Saga {2}{W} (Star Trek, #4973):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Exile up to one target nontoken artifact or creature.
//	 II — Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")
//	 III — Return the exiled card to the battlefield under its owner's
//	 control. You gain 2 life."
//
// Chapter I's exile is linked to this Saga OBJECT (ExileTarget.Linked,
// CR 607.2a), so chapter III's "the exiled card" is the card in exile
// that carries that link (ExiledWithForEffect on the Saga's {instance,
// epoch}, The Creation of Avacyn's reading). A Saga that left and came
// back is a new object with no claim on the card, which then stays in
// exile. Investigating is creating a Clue (Eliminate the Impossible).
// Chapter III returns the card under its owner's control, then you gain
// 2 life whether or not anything returned; the life waits for any
// question the returning permanent asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "727e98cc-b295-47f9-9d3d-bcf38c9f56d2",
		Name:         "The Measure of a Man",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("The Measure of a Man", 1, "exile up to one target nontoken artifact or creature"),
				TargetPermanent("up to one target nontoken artifact or creature", Or(Artifact(), Creature()), Not(IsTokenPredicate())).WithCount(0, 1),
				measureOfAManExile),
			ChapterTrigger(2, SagaChapterLabel("The Measure of a Man", 2, "investigate"),
				Do(Investigate{N: 1})),
			ChapterTrigger(3, SagaChapterLabel("The Measure of a Man", 3, "return the exiled card to the battlefield under its owner's control; you gain 2 life"),
				Do(Instruction(measureOfAManReturn), GainLife{Amount: 2})),
		},
	})
}

// measureOfAManExile is chapter I: the target, if still legal, exiled
// and linked to this Saga.
func measureOfAManExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ExileTarget{Target: id, Linked: true}.Apply(ctx)
}

// measureOfAManReturn is chapter III's return: every card in exile
// linked to this Saga object (at most one), under its owner's control.
func measureOfAManReturn(ctx *Context) error {
	ref, ok := ctx.Game.SourceObjectForEffect(ctx.Item)
	if !ok {
		return nil
	}
	// Copied out of the exile zone first, so the return may move them.
	type exiled struct{ id, owner uuid.UUID }
	var cards []exiled
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		cards = append(cards, exiled{c.InstanceID, c.Owner})
	}
	for _, c := range cards {
		if err := (ReturnFromExile{Target: c.id, Controller: c.owner}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
