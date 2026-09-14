package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arahbo, the First Fang — Legendary Creature — Cat Avatar {2}{W},
// 2/2 (EDHREC rank 3793):
//
//	"Other Cats you control get +1/+1.
//	 Whenever Arahbo or another nontoken Cat you control enters,
//	 create a 1/1 white Cat creature token."
//
// The Cat lord that brings a friend. The anthem is the shared
// TribalAnthem builder with "other" and "you control"; the trigger
// fires for Arahbo's own entry and for every other NONTOKEN Cat that
// enters under his controller's control — the tokens it makes are
// Cats, and are lifted by the anthem, but never re-fire it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dfec65a0-6fb6-463d-aa62-e84a427db217",
		Name:         "Arahbo, the First Fang",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Cat"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b36SelfOrAnotherNontokenCatYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Arahbo, the First Fang — create a 1/1 white Cat",
					b34CreateTokens(b36WhiteCatToken, 1))
			},
		}},
	})
}
