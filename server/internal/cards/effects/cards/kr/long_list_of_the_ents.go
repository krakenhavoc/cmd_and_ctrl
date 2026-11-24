package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Long List of the Ents — Enchantment — Saga {G}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after VI.)
//	 I, II, III, IV, V, VI — Note a creature type that hasn't been noted
//	 for this Saga. When you next cast a creature spell of that type this
//	 turn, that creature enters with an additional +1/+1 counter on it."
//
// Six chapter abilities with one text. Each asks for a creature type as it
// resolves, with the types already noted for this Saga left out (the
// prompt's Except), and notes it (game.Card.Notes, #6097): the list belongs
// to this object, so a Long List that leaves and returns starts over
// (CR 400.7). The rest is Kumano Faces Kakkazan's next-spell promise
// narrowed to that creature type: spent by the next creature spell of the
// type its controller casts this turn, and lapsing at end of turn (CR
// 603.7). A Saga that has left before its chapter resolves notes nothing,
// but the promise is still made for the type chosen.
func init() {
	const name = "Long List of the Ents"
	text := "note a creature type; when you next cast a creature spell of that type this turn, it enters with an additional +1/+1 counter"
	var chapters []game.TriggeredAbility
	for n := 1; n <= 6; n++ {
		chapters = append(chapters, effects.ChapterTrigger(n, effects.SagaChapterLabel(name, n, text), longListOfTheEntsChapter))
	}
	effects.Register(effects.Spec{
		OracleID:     "01376449-5cd2-4079-8aaa-5128634ef20b",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered:    chapters,
	})
}

func longListOfTheEntsChapter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	effects.NoteACreatureTypeThenTail(g, ctx.Controller(), ctx.Source(),
		"Long List of the Ents — note a creature type that hasn't been noted for this Saga", ctx.Tail(longListOfTheEntsNoted))
	return nil
}

// longListOfTheEntsNoted is the rest of a chapter, told the type. An
// on-disk key: never renamed, never reused.
var longListOfTheEntsNoted = effects.CreatureTypeChosen("creature-type/long-list-of-the-ents-note", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	ctx.Game.AddNoteForEffect(ctx.Source(), t)
	return effects.GrantNextSpellPromise{
		From: "Long List of the Ents",
		Promise: game.NextSpellPromise{
			Filter:      game.PermissionFilter{CreatureOnly: true, CreatureType: t},
			CounterKind: game.CounterPlusOne,
			Counters:    1,
			Text:        "When you next cast a " + t + " creature spell this turn, that creature enters with an additional +1/+1 counter on it.",
		},
	}.Apply(ctx)
})
