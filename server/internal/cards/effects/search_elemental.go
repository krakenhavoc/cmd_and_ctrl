package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Search Elemental — Creature — Elemental {U}, 1/1 (Mystery Booster 2):
//
//	"Whenever you search your library, scry 1.
//	 Scryfall — Whenever you scry, put a +1/+1 counter on Search Elemental. It
//	 can't be blocked this turn."
//
// Two triggers. The first reads the search event (Archivist of Oghma's
// condition, for your own library): the shuffle flavour of that event is not a
// search, and a search of another player's library by you (Bribery) is not "your
// library". The second is any scry by you, including the one the first trigger
// causes, so a search ends with a counter and an unblockable Elemental. Both
// act on the Elemental only while it is still the permanent the trigger came
// from (CR 400.7); the scry still happens without it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3d41f6d-1792-4a83-a4e8-d1b6d3ea04a8",
		Name:         "Search Elemental",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventSearchLibrary, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Label != "shuffle" && ev.Actor == source.Controller && ev.Target == ev.Actor
			}, "Search Elemental — scry 1", Do(Scry{N: 1})),
			On(game.EventScry, ByYou,
				"Search Elemental — put a +1/+1 counter on it; it can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Search Elemental — can't be blocked"}.Apply(ctx)
				}),
		},
	})
}
