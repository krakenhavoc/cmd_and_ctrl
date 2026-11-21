package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "fa98c367-0312-49c6-abef-72e5ead4cc7d",
		Name:         "Gingerbread Cabin",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.SelfEntersTappedUnless(func(g *game.Game, controller uuid.UUID) bool {
				return effects.B08LandsWithSubtypeControlled(g, controller, "Forest") >= 3
			}),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B27SelfEnteredUntapped(ev, source)
			}, "Gingerbread Cabin — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
