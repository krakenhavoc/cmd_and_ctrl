package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The World Spell — Enchantment — Saga {5}{G}{G} (Dominaria United):
//
//	"Read ahead (Choose a chapter and start with that many lore
//	 counters. Add one after your draw step. Skipped chapters don't
//	 trigger. Sacrifice after III.)
//	 I, II — Look at the top seven cards of your library. You may reveal a
//	 non-Saga permanent card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order.
//	 III — Put up to two non-Saga permanent cards from your hand onto the
//	 battlefield."
//
// Read ahead is the engine's (CR 702.155). Chapters I and II are one
// printed ability with two chapter numbers, so each is its own registered
// trigger with the same body (Contagious Vorrac's look-and-take): the
// look is at the top seven, the reveal is optional and takes at most one
// card, and everything not taken goes to the bottom in a random order.
// Chapter III puts up to two cards at once (a simultaneous batch), no
// cast, from the hand.
//
// No simplification.
func init() {
	const name = "The World Spell"
	lookTake := func(chapter int) game.TriggeredAbility {
		t := effects.ChapterTrigger(chapter, effects.SagaChapterLabel(name, chapter, "look at the top seven cards; you may put a non-Saga permanent card into your hand"),
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				player := ctx.Controller()
				return effects.TakeFromLibraryToHand{
					Player:   player,
					Cards:    g.LookAtTopOfLibraryForEffect(player, 7),
					Match:    nonSagaPermanentCard,
					Max:      1,
					Optional: true,
					Label:    name + " — you may reveal a non-Saga permanent card and put it into your hand",
					Key:      worldSpellTake,
				}.Apply(ctx)
			})
		t.Purpose = game.Purpose{Tutors: 1}
		return t
	}
	effects.Register(effects.Spec{
		OracleID:        "f5186d3f-52b2-43ef-be11-d183ab765663",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordReadAhead},
		Triggered: []game.TriggeredAbility{
			lookTake(1),
			lookTake(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "put up to two non-Saga permanent cards from your hand onto the battlefield"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromHandOntoBattlefield{
						Match:    nonSagaPermanentCard,
						Optional: true,
						Max:      2,
						Label:    name + " — put up to two non-Saga permanent cards from your hand onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// nonSagaPermanentCard is "non-Saga permanent card".
func nonSagaPermanentCard(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsPermanent() && !game.IsSaga(c)
}

// worldSpellTake is the taken card revealed and the rest on the bottom in a
// random order.
var worldSpellTake = effects.RegisterTake("choose-cards/the-world-spell-take",
	effects.TakeDef{Reveal: true, Then: effects.TakeRestOnBottomInRandomOrder})
