package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpine Guide — Snow Creature — Human Scout {2}{R}, 3/3 (Modern Horizons):
//
//	"When this creature enters, you may search your library for a Mountain
//	 card, put that card onto the battlefield tapped, then shuffle.
//	 This creature attacks each combat if able.
//	 When this creature leaves the battlefield, sacrifice a Mountain."
//
// The search is Bedhead Beastie's Mountain test (any land with the Mountain
// type, not only a basic one), the card entering tapped. The requirement is
// Amarant Coral's "attacks each combat if able" (CR 508.1d). The leave trigger
// is mandatory and its controller chooses which Mountain they control to
// sacrifice, read as it resolves; with none it does nothing (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55b12241-2495-45e6-b618-3746d5521291",
		Name:         "Alpine Guide",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Alpine Guide — you may search your library for a Mountain card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsLandWithSubtype("Mountain"),
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Optional:      true,
						Reason:        "Alpine Guide — a Mountain card, tapped",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.WhenThisLeaves("Alpine Guide — sacrifice a Mountain",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificeChoice{
						Player:     item.Controller,
						Candidates: effects.PermanentsControlledByMatching(g, item.Controller, effects.HasSubtype("Mountain")),
						Question:   "Alpine Guide — sacrifice a Mountain",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
