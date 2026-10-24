package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invert // Invent — split card, oracle 9a378964-2c04-4d60-a905-c819d37ed4c3:
//
//	Invert — Instant {U/R}: "Switch the power and toughness of each of up
//	         to two target creatures until end of turn."
//	Invent — Instant {4}{U}{R}: "Search your library for an instant card
//	         and/or a sorcery card, reveal them, put them into your hand,
//	         then shuffle."
//
// ADR 0103's split card: each half registers as an ADR 0034 face, Invert
// under the bare oracle ID and Invent under "#1". Invert's switch is
// layer 7d (CR 613.4d, #3996), after every set, pump and counter.
// Invent is Axgard Armory's two-step search: up to one instant card,
// then up to one sorcery card, both revealed, one shuffle after the
// second.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9a378964-2c04-4d60-a905-c819d37ed4c3",
		Name:         "Invert",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return switchEachTargetPT(ctx, "Invert — switch power and toughness")
		},
	})
	Register(Spec{
		OracleID:     "9a378964-2c04-4d60-a905-c819d37ed4c3#1",
		Name:         "Invent",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller, source := item.Controller, item.SourceCardID
			return SearchLibrary{
				Player:    controller,
				Predicate: func(c game.Card) bool { return c.IsInstant() },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Reason:    "Invent — choose up to one instant card to put into your hand",
				ThenTail:  game.KeyedTail{Ref: inventSorcerySearch, Player: controller, Source: source},
			}.Apply(ctx)
		},
	})
}

// inventSorcerySearch is the sorcery half of Invent's search, run once the
// instant search has settled. An on-disk key.
var inventSorcerySearch = game.RegisterTail("invent/sorcery-search", func(g *game.Game, o game.TailOutcome) error {
	return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:  o.Tail.Player,
		Source:  o.Tail.Source,
		Pred:    func(c game.Card) bool { return c.IsSorcery() },
		Dest:    game.ZoneHand,
		Limit:   1,
		Reveal:  true,
		Shuffle: true,
		Reason:  "Invent — choose up to one sorcery card to put into your hand",
	})
})
