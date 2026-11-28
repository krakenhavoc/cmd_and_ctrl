package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jungle Wurm — Creature — Wurm {3}{G}{G}, 5/5 (Mirage, #5633):
//
//	"Whenever this creature becomes blocked, it gets -1/-1 until end of turn for
//	 each creature blocking it beyond the first."
//
// The mirror of rampage: the blockers are counted as the trigger resolves
// (CR 702.23b), and a single blocker costs it nothing.
//
// No simplification.
func init() {
	const label = "Jungle Wurm — -1/-1 for each creature blocking it beyond the first"
	effects.Register(effects.Spec{
		OracleID:     "457705c9-5b0c-4c3a-89dd-80286fac6a6c",
		Name:         "Jungle Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesBlocked && ev.CardID == source.InstanceID
			}, label, effects.OldB796ThisGetsMinusPerExtraBlocker(label)),
		},
	})
}
