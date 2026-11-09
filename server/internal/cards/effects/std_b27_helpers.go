package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b27_helpers.go — shared pieces of Standard-set batches 27–30
// (Bloomburrow and Duskmourn leftovers, umbrella #4452).

// The delayed "return it at the beginning of the next end step" bodies
// with one counter each (Parting Gust's +1/+1, Salvation Swan's flying).
var (
	returnExiledWithPlusOneCreatureBody = game.SimpleDelayedBody(
		"flicker/return-exiled-with-plus-one-counter", returnExiledCardsWithPlusOneCounter)
	returnExiledWithFlyingCounterBody = game.SimpleDelayedBody(
		"flicker/return-exiled-with-flying-counter", returnExiledCardsWithFlyingCounter)
)

func exiledCardTargets(item *game.StackItem) []uuid.UUID {
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

// returnExiledCardsWithPlusOneCounter returns the listed cards together
// under their owners' control, each entering with a +1/+1 counter on it
// (CR 614.12; Parting Gust).
func returnExiledCardsWithPlusOneCounter(g *game.Game, item *game.StackItem) error {
	return ReturnFromExileTogether{
		Targets:        exiledCardTargets(item),
		CountersIfType: []game.EntryCountersIfType{{CardType: "creature", Counter: game.CounterPlusOne, N: 1}},
	}.Apply(NewContext(g, item))
}

// returnExiledCardsWithFlyingCounter returns the listed cards together
// under their owners' control, each entering with a flying counter on it
// (Salvation Swan).
func returnExiledCardsWithFlyingCounter(g *game.Game, item *game.StackItem) error {
	return ReturnFromExileTogether{
		Targets:        exiledCardTargets(item),
		CountersIfType: []game.EntryCountersIfType{{CardType: "creature", Counter: game.CounterFlying, N: 1}},
	}.Apply(NewContext(g, item))
}

// splashPortalDrawIfTribe is Splash Portal's "if that creature is a Bird,
// Frog, Otter, or Rat, draw a card", asked of the permanent that actually
// came back (the ruling: its types after returning). A token never
// returns, so entered is nil and nothing is drawn. An on-disk key.
var splashPortalDrawIfTribe = EnteredTail("blb/splash-portal-draw-if-tribe", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(entered)
	if !ok {
		return nil
	}
	for _, sub := range []string{"Bird", "Frog", "Otter", "Rat"} {
		if c.HasSubtype(sub) {
			return DrawCards{N: 1}.Apply(ctx)
		}
	}
	return nil
})

// splashPortalReturn returns the exiled card under its owner's control
// and then asks the tribe question of it. An on-disk key.
var splashPortalReturn = CardTail("blb/splash-portal-return", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	return ReturnFromExile{Target: o.Tail.Carry[0], ThenTail: ctx.ItemTail(splashPortalDrawIfTribe)}.Apply(ctx)
})

// driftgloomCoyoteCounter is "if that creature had power 2 or less, put a
// +1/+1 counter on this creature": N carries the answer, read off the
// creature before it was exiled (CR 608.2h), and it runs only when the
// creature was exiled this way. An on-disk key.
var driftgloomCoyoteCounter = CardTail("blb/driftgloom-coyote-counter", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N != 1 {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(o.Tail.Source); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return ctx.Game.AddCounterForEffect(o.Tail.Source, game.CounterPlusOne, 1)
})

// coilingRebirthCopy is Coiling Rebirth's "then if the gift was promised and
// that creature isn't legendary, create a token that's a copy of that
// creature, except it's 1/1": N is 1 when the gift was promised, and the
// question is asked of the creature that actually came back. An on-disk key.
var coilingRebirthCopy = EnteredTail("blb/coiling-rebirth-copy", func(ctx *Context, entered uuid.UUID, o game.TailOutcome) error {
	if entered == uuid.Nil || o.Tail.N != 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(entered)
	if !ok || c.IsLegendary() {
		return nil
	}
	return CreateTokenCopy{
		Copy: entered,
		N:    1,
		Except: func(tok *game.Card) {
			tok.Power, tok.Toughness = 1, 1
			tok.VariableToughness = false
			tok.PrintedPTKnown = true
		},
	}.Apply(ctx)
})

// TapCreaturesForManaCost is "{cost}. For each mana in this ability's
// activation cost, you may tap an untapped creature you control rather
// than pay that mana" (Heirloom Epic): the whole mana cost may be paid by
// tapping untapped creatures, one for each {1}. It is waterbend's
// machinery over creatures alone and under its own key, so it never counts
// as waterbending (CR 701.67) and an artifact cannot pay.
func TapCreaturesForManaCost(cost string) game.AbilityCost {
	return game.AbilityCost{Mana: cost, Waterbend: &game.TapPermanentsCost{
		Key:   "tap-creatures-for-mana",
		Label: "Tap creatures for {1} each",
		Extra: cost,
		Spec:  TargetCreature("an untapped creature you control", YouControl(), Untapped()),
	}}
}
