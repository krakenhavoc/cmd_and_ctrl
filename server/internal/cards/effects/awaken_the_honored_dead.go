package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Awaken the Honored Dead — Enchantment — Saga {B}{G}{U} (Tarkir:
// Dragonstorm, #4463):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Destroy target nonland permanent.
//	 II — Mill three cards.
//	 III — You may discard a card. When you do, return target creature or
//	 land card from your graveyard to your hand."
//
// Chapter I is Binding the Old Gods' targeted destroy without the
// opponent restriction. Chapter III is Tether Technician's optional
// discard and reflexive trigger (CR 603.12): the trigger exists only once
// a card was really discarded, and it targets as it goes on the stack, so
// the card just discarded can be the one returned. With an empty hand
// nothing is discarded and nothing returns.
//
// No simplification.
func init() {
	const name = "Awaken the Honored Dead"
	Register(Spec{
		OracleID:     "5f694917-f434-4459-bc86-caf9ac1c80e7",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "destroy target nonland permanent"),
				TargetPermanent("target nonland permanent", Nonland()),
				destroyChosenPermanent),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "mill three cards"), Do(MillCards{N: 3})),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "you may discard a card; when you do, return target creature or land card from your graveyard to your hand"),
				awakenTheHonoredDeadDiscard),
		},
	})
}

// awakenTheHonoredDeadReturnBody is chapter III's reflexive trigger. An
// on-disk key.
var awakenTheHonoredDeadReturnBody = game.ReflexiveBody("awaken-the-honored-dead/return-to-hand",
	simpleBody(returnFirstLegalGraveyardTargetToHand),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature or land card from your graveyard", YouOwn(), Or(Creature(), Land()))
	}))

// awakenTheHonoredDeadDiscard is chapter III: the optional discard.
func awakenTheHonoredDeadDiscard(g *game.Game, item *game.StackItem) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        1,
		UpTo:     true,
		Question: "Awaken the Honored Dead — you may discard a card to return a creature or land card from your graveyard to your hand",
	}, awakenTheHonoredDeadDiscarded, nil)
}

// awakenTheHonoredDeadDiscarded is the discard answered, keyed: a card
// discarded makes the reflexive trigger. An on-disk key.
var awakenTheHonoredDeadDiscarded = game.RegisterDiscardThen("choose-cards/awaken-the-honored-dead-return", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Discarded.Cards()) == 0 {
		return nil
	}
	return WhenYouDo("Awaken the Honored Dead — return target creature or land card from your graveyard to your hand",
		awakenTheHonoredDeadReturnBody).Apply(keyedContext(g, r.Player, r.Source))
})
