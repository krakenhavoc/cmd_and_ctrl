package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gingerbread Cabin — Land — Forest (EDHREC rank 2847):
//
//	"({T}: Add {G}.)
//	 This land enters tapped unless you control three or more other
//	 Forests.
//	 When this land enters untapped, create a Food token. (It's an
//	 artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")"
//
// The Eldraine Forest that pays out a Food once the deck is mostly
// Forests. It IS a Forest — the mana ability is in reminder text
// because the land type supplies it, and a fetchland can find it —
// so the {G} is declared explicitly for one click in the client and
// the intrinsic ability the type would give is covered by it. The
// tapped entry is SelfEntersTappedUnless counting the OTHER Forests
// the controller controls: the Cabin is not on the battlefield while
// its own entry is being replaced, so "other" costs nothing extra.
// The Food trigger reads the Cabin's tapped flag as it enters (the
// replacement has already run by the time the harvester sees the
// ETB), so an untapped entry — three Forests already, or a Horizon
// Explorer's "lands you control enter untapped" — makes the Food and
// a tapped one does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa98c367-0312-49c6-abef-72e5ead4cc7d",
		Name:         "Gingerbread Cabin",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			SelfEntersTappedUnless(func(g *game.Game, controller uuid.UUID) bool {
				return b08LandsWithSubtypeControlled(g, controller, "Forest") >= 3
			}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b27SelfEnteredUntapped(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Gingerbread Cabin — create a Food token",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
