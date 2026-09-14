package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tiamat — Legendary Creature — Dragon God {2}{W}{U}{B}{R}{G}, 7/7
// (EDHREC rank 2995):
//
//	"Flying
//	 When Tiamat enters, if you cast it, search your library for up
//	 to five Dragon cards not named Tiamat that each have different
//	 names, reveal them, put them into your hand, then shuffle."
//
// The five-colour Dragon tutor. Flying rides PrintedKeywords. "If
// you cast it" is read off the event log (b16EnteredFromStack —
// Zacama's shape): a Tiamat reanimated, blinked or fetched onto the
// battlefield does not tutor, as printed. It is checked as the
// trigger would fire and not again at resolution, which for "if you
// cast it" can never change between the two.
//
// The search is one SearchLibrary: Dragon cards not named Tiamat, up
// to five, revealed, to hand, then shuffle — with "each have
// different names" as the Validate over the picked set
// (b28DifferentNames, Myriad Landscape's shape), because no per-card
// predicate can say it. The searcher chooses (S22) and may fail to
// find.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f00e4df1-13fb-4514-8abb-92954068689a",
		Name:            "Tiamat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && b16EnteredFromStack(g, source.InstanceID)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Tiamat — search for up to five differently named Dragon cards",
					func(g *game.Game, item *game.StackItem) error {
						return b28SearchDragonsNotNamed(g, item, "Tiamat", "Tiamat: up to five Dragon cards not named Tiamat with different names")
					})
			},
		}},
	})
}
