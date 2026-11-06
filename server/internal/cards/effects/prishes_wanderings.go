package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prishe's Wanderings — Instant {2}{G} (Final Fantasy, #4462):
//
//	"Search your library for a basic land card or Town card, put it onto
//	 the battlefield tapped, then shuffle. When you search your library
//	 this way, put a +1/+1 counter on target creature you control."
//
// Rampant Growth's search with a wider net: a basic land card, or any
// card with the Town subtype. A search for a stated quality may find
// nothing (CR 701.23b), but it is still a search, so the reflexive
// trigger (CR 603.12) is made from the search's continuation whatever
// was found. Its target is chosen as it goes on the stack, after the
// spell has finished resolving, so the land found is already on the
// battlefield; Teo, Spirited Glider's counter body puts the +1/+1
// counter on the target as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc92988c-d628-45ea-8606-516a1211cef3",
		Name:         "Prishe's Wanderings",
		Purpose:      game.Purpose{Lands: 1},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     prishesWanderingsFinds,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Prishe's Wanderings — a basic land card or Town card, tapped",
				ThenTail:      ctx.ItemTail(prishesWanderingsSearched),
			}.Apply(ctx)
		},
	})
}

// prishesWanderingsFinds is "a basic land card or Town card".
func prishesWanderingsFinds(c game.Card) bool {
	return IsBasicLand(c) || c.HasSubtype("Town")
}

// prishesWanderingsSearched is "When you search your library this way":
// the search has happened, so the reflexive trigger is made. An on-disk
// key.
var prishesWanderingsSearched = CardTail("prishes-wanderings/searched", func(ctx *Context, _ game.TailOutcome) error {
	return WhenYouDo("Prishe's Wanderings — put a +1/+1 counter on target creature you control",
		prishesWanderingsCounterBody).Apply(ctx)
})

// prishesWanderingsCounterBody is the reflexive half: a +1/+1 counter on
// the target. An on-disk key.
var prishesWanderingsCounterBody = game.ReflexiveBody("prishes-wanderings/counter", simpleBody(PlusOneCounterOnTheTarget),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target creature you control", YouControl())
	}))
