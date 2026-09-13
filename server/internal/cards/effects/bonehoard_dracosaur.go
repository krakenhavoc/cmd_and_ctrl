package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bonehoard Dracosaur — Creature — Dinosaur Dragon {3}{R}{R}, 5/5
// (EDHREC rank 1375):
//
//	"Flying, first strike
//	 At the beginning of your upkeep, exile the top two cards of your
//	 library. You may play them this turn. If you exiled a land card
//	 this way, create a 3/1 red Dinosaur creature token. If you exiled
//	 a nonland card this way, create a Treasure token."
//
// Two cards a turn plus a body or a Treasure for each kind, on a 5/5
// flying first striker. The exile is the impulse-exile primitive
// with a "play" grant (so the land can be the turn's drop), and the
// riders read the exiled cards where they now sit — the printed
// type line, which for "land card" / "nonland card" is exactly the
// question. Each rider fires at most once ("a land card", not "each
// land card"): two lands is one Dinosaur, one land and one spell is
// one of each, and an empty library exiles nothing and makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7ba550a5-81fc-42c2-8df8-ad455d938605",
		Name:            "Bonehoard Dracosaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Bonehoard Dracosaur — exile the top two cards, play them this turn",
					func(g *game.Game, item *game.StackItem) error {
						exiled, err := b12ImpulseExileForTurn(g, item, 2)
						if err != nil {
							return err
						}
						land, nonland := false, false
						for _, id := range exiled {
							c, ok := g.LookupCardForEffect(id)
							if !ok {
								continue
							}
							if c.IsLand() {
								land = true
							} else {
								nonland = true
							}
						}
						ctx := NewContext(g, item)
						if land {
							if err := (CreateToken{Controller: item.Controller, Template: b12RedDinosaurToken(), N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						if nonland {
							if err := (CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
			},
		}},
	})
}
