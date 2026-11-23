package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viewpoint Synchronization — Sorcery {4}{G}:
//
//	"Freerunning {2}{G} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)
//	 Search your library for up to three basic land cards and reveal them.
//	 Put two of them onto the battlefield tapped and the other into your
//	 hand, then shuffle."
//
// Cultivate's search stretched to three cards, one search per card: the
// first two basic lands go onto the battlefield tapped and the third into
// hand, each revealed as it is found, and the library shuffles once at the
// end. The searches run in order, and a search that finds nothing ends the
// run (with the shuffle): with one land found it goes onto the battlefield,
// with two found both do, and only a third goes into hand, as the printed
// "two of them … and the other" has it. Each step is a registered tail, so
// a table waiting on an entry question is a restore point. Freerunning is
// Eagle Vision's alternative cost.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2e338e09-5758-4cc2-9143-21f050a28677",
		Name:             "Viewpoint Synchronization",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Lands: 3, Tutors: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Freerunning("{2}{G}", "Assassin")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Source:        ctx.Source(),
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				TappedOnEntry: true,
				Reason:        "Viewpoint Synchronization — first basic land onto the battlefield tapped",
				ThenTail:      game.KeyedTail{Ref: viewpointSynchronizationSecond, Player: ctx.Controller(), Source: ctx.Source()},
			}.Apply(ctx)
		},
	})
}

// viewpointSynchronizationSecond is the second basic land, onto the
// battlefield tapped, once the first was found. An on-disk key.
var viewpointSynchronizationSecond = game.RegisterTail("then/viewpoint-synchronization-second", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return g.ShuffleLibraryForEffect(o.Tail.Player)
	}
	return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:        o.Tail.Player,
		Source:        o.Tail.Source,
		Pred:          effects.IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Reveal:        true,
		TappedOnEntry: true,
		Reason:        "Viewpoint Synchronization — second basic land onto the battlefield tapped",
		ThenTail:      game.KeyedTail{Ref: viewpointSynchronizationThird, Player: o.Tail.Player, Source: o.Tail.Source},
	})
})

// viewpointSynchronizationThird is the last basic land, into hand, then the
// shuffle. An on-disk key.
var viewpointSynchronizationThird = game.RegisterTail("then/viewpoint-synchronization-third", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return g.ShuffleLibraryForEffect(o.Tail.Player)
	}
	return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:  o.Tail.Player,
		Source:  o.Tail.Source,
		Pred:    effects.IsBasicLand,
		Dest:    game.ZoneHand,
		Limit:   1,
		Reveal:  true,
		Shuffle: true,
		Reason:  "Viewpoint Synchronization — basic land into your hand",
	})
})
