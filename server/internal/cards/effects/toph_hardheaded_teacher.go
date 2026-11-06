package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Toph, Hardheaded Teacher — Legendary Creature — Human Warrior Ally
// {2}{R}{G}, 3/4 (Avatar: The Last Airbender, #4459):
//
//	"When Toph enters, you may discard a card. If you do, return target
//	 instant or sorcery card from your graveyard to your hand.
//	 Whenever you cast a spell, earthbend 1. If that spell is a Lesson,
//	 put an additional +1/+1 counter on that land."
//
// The enters trigger targets an instant or sorcery card in your graveyard
// as it goes on the stack (CR 603.3d: with none it is removed and asks
// nothing). As it resolves you may discard one card (Formidable Speaker's
// up-to-one discard prompt); only if a card was really discarded is the
// still-legal target returned to your hand (CR 608.2b).
//
// The cast trigger is earthbend 1 on "target land you control" (the
// keyword's target clause). When the spell that triggered it is a Lesson
// (read off the spell, or its last-known record if it has already left the
// stack, CR 608.2h), one more +1/+1 counter goes on that land as a second
// placement after earthbend's own counter has landed, so a counter
// replacement applies to each placement separately.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Toph, Hardheaded Teacher — you may discard a card; if you do, return an instant or sorcery card", tophHardheadedEnters)
	enters.Targets = TargetCardInGraveyard("target instant or sorcery card from your graveyard", YouOwn(), Or(Instant(), Sorcery()))
	Register(Spec{
		OracleID:     "8c1c0117-6fc0-43c5-a4c8-6472de6503d9",
		Name:         "Toph, Hardheaded Teacher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			enters,
			Targeting(WheneverYouCast(nil, "Toph, Hardheaded Teacher — earthbend 1 (an additional counter for a Lesson)", tophHardheadedEarthbend),
				EarthbendTargets()),
		},
	})
}

// tophHardheadedEnters asks for the optional discard.
func tophHardheadedEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   ctx.Controller(),
		Source:   ctx.Source(),
		N:        1,
		UpTo:     true,
		Question: "Toph, Hardheaded Teacher — you may discard a card; if you do, return target instant or sorcery card to your hand",
		Asked:    game.NewAskedItem(item),
	}, tophHardheadedReturn, nil)
}

// tophHardheadedReturn is "If you do, return target instant or sorcery
// card from your graveyard to your hand". An on-disk key.
var tophHardheadedReturn = game.RegisterDiscardThen("choose-cards/toph-hardheaded-teacher-return", func(g *game.Game, r game.DiscardsDone) error {
	if r.Item == nil || r.Discarded.Count() == 0 {
		return nil
	}
	return returnLegalGraveyardTargetsToHand(NewContext(g, r.Item))
})

// tophHardheadedEarthbend is "earthbend 1. If that spell is a Lesson, put
// an additional +1/+1 counter on that land."
func tophHardheadedEarthbend(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	land := FirstLegalBattlefieldTarget(ctx)
	if land == uuid.Nil {
		return nil
	}
	var then game.KeyedTail
	if spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID); ok && spell.HasSubtype("Lesson") {
		then = ctx.ItemTail(tophHardheadedLessonCounter)
		then.Carry = []uuid.UUID{land}
	}
	return Earthbend{Target: land, N: 1, Then: then}.Apply(ctx)
}

// tophHardheadedLessonCounter is the additional +1/+1 counter on that land
// (Carry), once earthbend's counter has landed. An on-disk key.
var tophHardheadedLessonCounter = CardTail("toph-hardheaded-teacher/lesson-counter", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 || !onBattlefield(ctx.Game, o.Tail.Carry[0]) {
		return nil
	}
	return AddCounter{Target: o.Tail.Carry[0], Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
