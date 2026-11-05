package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b20_helpers.go — shared pieces for Star Trek batch 20 (#4973).

// trekB20IsBasicLandOrPlanetCard is "a basic land card or Planet card":
// a basic land, or a land card with the Planet land type (CR 205.3i).
func trekB20IsBasicLandOrPlanetCard(c game.Card) bool {
	return b30IsBasicLandCard(c) || (c.IsLand() && c.HasSubtype("Planet"))
}

// trekB20SearchBasicOrPlanetTapped is "search your library for a basic
// land card or Planet card, put it onto the battlefield tapped, then
// shuffle": Final Frontier's spell and Genesis Device's enters trigger.
func trekB20SearchBasicOrPlanetTapped(ctx *Context, reason string) error {
	return SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     trekB20IsBasicLandOrPlanetCard,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		TappedOnEntry: true,
		Shuffle:       true,
		Reason:        reason,
	}.Apply(ctx)
}

// trekB20OtherCreatureYouControlWithACounter is M'Benga's "other
// creatures you control with counters on them", any kind of counter.
func trekB20OtherCreatureYouControlWithACounter(target *game.Card, g *game.Game, source *game.Card) bool {
	return source != nil && target.InstanceID != source.InstanceID && socB9CreatureYouControlWithACounter(target, g, source)
}

// trekB20IsFirstSpellCastThisGame reports whether the resolving spell is
// the first spell its controller cast this game (First Contact): the very
// object stamped as it became cast (Player.FirstSpellCast, CR 601.2i). A
// copy is not cast (CR 707.10) and is a different object, so it never is.
func trekB20IsFirstSpellCastThisGame(ctx *Context) bool {
	item := ctx.Item
	if item == nil || item.IsCopy {
		return false
	}
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.FirstSpellCast.ID == uuid.Nil || p.FirstSpellCast.ID != item.SourceCardID {
		return false
	}
	c, ok := ctx.Game.LookupCardForEffect(item.SourceCardID)
	return ok && c.ObjectEpoch == p.FirstSpellCast.Epoch
}

// trekB20McCoyDiesBody is Dr. McCoy's delayed "when it dies this turn"
// firing: the card that went to the graveyard, as the new object it is
// there (CR 400.7), is followed to the next end step.
var trekB20McCoyDiesBody = game.SimpleDelayedBody("dr-leonard-mccoy/when-it-dies-this-turn", trekB20McCoyDied)

// trekB20McCoyReturnBody is the second delayed trigger, "at the
// beginning of the next end step": the card returns under its owner's
// control with a +1/+1 counter on it, if it is still that card in a
// graveyard (CR 603.7c).
var trekB20McCoyReturnBody = game.DelayedBody("dr-leonard-mccoy/return-at-next-end-step", trekB20McCoyReturn)

// trekB20McCoyDied schedules the end-step return for the creature card
// that died. A token ceases to exist (CR 111.7) and a card that is not in
// a graveyard (a commander its owner moved) schedules nothing.
func trekB20McCoyDied(g *game.Game, item *game.StackItem) error {
	obj, ok := payloadObject(NewContext(g, item))
	if !ok {
		return nil
	}
	z := g.FindCardZoneForEffect(obj.ID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(obj.ID)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Dr. Leonard \"Bones\" McCoy — return that card to the battlefield with a +1/+1 counter on it",
		Cards:  []uuid.UUID{obj.ID},
		Body:   trekB20McCoyReturnBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: obj.ID, Epoch: c.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// trekB20McCoyReturn is trekB20McCoyReturnBody: under its OWNER's control
// (uuid.Nil), with one +1/+1 counter.
func trekB20McCoyReturn(g *game.Game, _ *game.StackItem, p game.EffectParams) error {
	if !objectStillInAGraveyard(g, p.Object) {
		return nil
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(p.Object.ID, uuid.Nil, false,
		map[string]int{game.CounterPlusOne: 1})
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}

// trekB20FetchThatManyLandsTapped is "search your library for up to that
// many land cards, put them onto the battlefield tapped, then shuffle",
// run once the sacrificed lands have gone (Genesis Device). Zero is a
// search for nothing, and the shuffle still happens.
func trekB20FetchThatManyLandsTapped(ctx *Context, n int, reason string) error {
	if n <= 0 {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:        ctx.Controller(),
		Source:        ctx.Source(),
		Pred:          b03IsLandCard,
		Dest:          game.ZoneBattlefield,
		Limit:         n,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        reason,
	})
}
