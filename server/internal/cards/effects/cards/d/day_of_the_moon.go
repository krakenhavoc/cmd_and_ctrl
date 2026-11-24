package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Day of the Moon — Enchantment — Saga {2}{R}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II, III — Choose a creature card name, then goad all creatures
//	 with a name chosen for this enchantment. (Until your next turn, they
//	 attack each combat if able and attack a player other than you if
//	 able.)"
//
// Three chapter abilities with one text. Each asks its controller for a
// card name as it resolves and notes it for this Saga (game.Card.Notes,
// #6097), so chapter III goads the creatures with any of the three names.
// The names belong to this object: a Day of the Moon that leaves and
// returns starts over (CR 400.7), and a chapter whose Saga has left in
// response still reads the names it had (CR 608.2h). "A creature card
// name" is kept where the name is read (CreatureWithChosenCreatureCardName).
// The goad is the ordinary resolving goad, until its controller's next
// turn (CR 701.15a).
func init() {
	const name = "Day of the Moon"
	text := "choose a creature card name, then goad all creatures with a name chosen for this enchantment"
	effects.Register(effects.Spec{
		OracleID:     "7ec3aaa4-8b6a-4046-b599-654f9844ca6c",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, text), dayOfTheMoonChapter),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, text), dayOfTheMoonChapter),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, text), dayOfTheMoonChapter),
		},
	})
}

func dayOfTheMoonChapter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
		"Day of the Moon — choose a creature card name", ctx.Tail(dayOfTheMoonNamed))
	return nil
}

// dayOfTheMoonNamed is the rest of a chapter, told the chosen name. An
// on-disk key: never renamed, never reused.
var dayOfTheMoonNamed = effects.CardTail("card-name/day-of-the-moon-goad", func(ctx *effects.Context, o game.TailOutcome) error {
	chosen := o.Tail.Word
	ctx.Game.AddNoteForEffect(ctx.Source(), chosen)
	names := ctx.Game.NotesForEffect(ctx.Source())
	if chosen != "" && !game.NotedIn(names, chosen) {
		names = append(names, chosen)
	}
	if len(names) == 0 {
		return nil
	}
	return effects.GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool {
		return effects.CreatureWithChosenCreatureCardName(c, names)
	})
})
