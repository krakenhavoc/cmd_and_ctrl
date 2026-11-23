package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chill of Foreboding — Sorcery {2}{U} (Dark Ascension, #5571):
//
//	"Each player mills five cards.
//	 Flashback {7}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Each player mills in turn order from the active player
// (b02EachPlayerMills, Eye Collector's body); a library with fewer than
// five cards mills what it has. Flashback is the shared alternative cost
// (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "cd1677e4-c356-43cd-9766-5fb8a1e64249",
		Name:             "Chill of Foreboding",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{7}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B02EachPlayerMills(ctx.Game, item, 5)
		},
	})
}
