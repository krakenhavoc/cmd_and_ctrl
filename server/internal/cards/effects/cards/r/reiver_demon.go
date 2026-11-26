package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reiver Demon — Creature — Demon {4}{B}{B}{B}{B}, 6/6 (Commander 2011, S82
// older sets batch 453, #5574):
//
//	"Flying
//	 When this creature enters, if you cast it from your hand, destroy all
//	 nonartifact, nonblack creatures. They can't be regenerated."
//
// Wakening Sun's Avatar's cast-from-hand trigger with the other filter; the
// Demon is itself black so survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7dcb95a9-c114-4333-8b45-4380f92d9695",
		Name:            "Reiver Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.B06SelfETB(ev, source, lki, g) && effects.B30CastFromHand(g, source.InstanceID)
			}, "Reiver Demon — destroy all nonartifact, nonblack creatures",
				effects.S82OldB453DestroyNonartifactNonblack),
		},
	})
}
