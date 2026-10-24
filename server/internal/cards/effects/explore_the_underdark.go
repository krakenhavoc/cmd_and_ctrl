package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Explore the Underdark — Sorcery {4}{G}:
//
//	"Search your library for up to two basic land cards and/or Gate
//	 cards, put them onto the battlefield tapped, then shuffle.
//	 You take the initiative."
//
// Circuitous Route's search, then TakeTheInitiative (ADR 0203, #4096).
// The searcher chooses, so the search may queue a prompt; the initiative
// is taken in the search's continuation, after the lands have entered
// and the library is shuffled, which keeps the printed order. Either
// way the venture it causes is the initiative's own trigger, put on the
// stack after the sorcery has finished resolving.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13bad22a-6ba3-4197-8658-e47049a1cf1f",
		Name:         "Explore the Underdark",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			you := item.Controller
			return SearchLibrary{
				Player: you,
				Predicate: func(c game.Card) bool {
					return IsBasicLand(c) || (c.IsLand() && c.HasSubtype("Gate"))
				},
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Explore the Underdark — up to two basic lands and/or Gates, onto the battlefield tapped",
				Source:        item.SourceCardID,
				// #4253 slice 4: the initiative is a registered tail.
				ThenTail: game.KeyedTail{Ref: exploreTheUnderdarkInitiative, Player: you, Source: item.SourceCardID},
			}.Apply(ctx)
		},
	})
}

// exploreTheUnderdarkInitiative is the search's continuation: "You take
// the initiative." An on-disk key.
var exploreTheUnderdarkInitiative = game.RegisterTail("explore-the-underdark/take-the-initiative",
	func(g *game.Game, o game.TailOutcome) error {
		you := o.Tail.Player
		if p := g.PlayerByIDForEffect(you); p == nil || p.Eliminated {
			return nil
		}
		return g.TakeInitiativeForEffect(you, o.Tail.Source)
	})
