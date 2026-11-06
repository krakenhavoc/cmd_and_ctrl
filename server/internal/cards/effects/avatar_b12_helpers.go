package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b12_helpers.go — shared helpers for Avatar: The Last Airbender,
// batch 12 (#4459).

// avB12YouNextCastALessonCondition is the event condition of "when you
// next cast a Lesson spell this turn" (Jeong Jeong, the Deserter): an
// EventCast by the delayed trigger's controller whose spell has the
// Lesson subtype. game.CastFilter names card types only (CR 205.2a), and
// Lesson is a spell subtype (CR 205.3k), so the subtype gets a condition
// of its own rather than a filter. An on-disk key: never renamed, never
// reused.
var avB12YouNextCastALessonCondition = game.DelayedCondition("cast/you-next-cast-a-lesson",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventCast || ev.Actor != dt.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && isLessonCard(c)
	})

// avB12DragonOrLessonCardsInGraveyard counts the Dragon and/or Lesson
// cards in a player's graveyard (Ran and Shaw): a card that is both
// counts once. `dragon` is the creature type as the card's text reads
// it (CR 612.2).
func avB12DragonOrLessonCardsInGraveyard(g *game.Game, player uuid.UUID, dragon string) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.HasSubtype(dragon) || isLessonCard(c) {
			n++
		}
	}
	return n
}

// destroyThenItsControllerMaySearchForATappedBasic is "Destroy target
// <permanent>. Its controller may search their library for a basic land
// card, put it onto the battlefield tapped, then shuffle." (Avengers
// Disassembled's land mode, Price of Freedom). The controller is read
// before the destroy, since the permanent is gone afterwards; a target
// that is no longer on the battlefield does nothing.
func destroyThenItsControllerMaySearchForATappedBasic(ctx *Context, target uuid.UUID, reason string) error {
	victim, ok := controllerOfTarget(ctx, target)
	if !ok {
		return nil
	}
	if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return SearchLibrary{
		Player:        victim,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: true,
		Optional:      true,
		Reason:        reason,
	}.Apply(ctx)
}
