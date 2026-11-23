package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// All of History, All at Once — Sorcery {2}{U}{U} (Doctor Who):
//
//	"Time travel. (For each suspended card you own and each permanent you
//	 control with a time counter on it, you may add or remove a time counter.)
//	 Storm (When you cast this spell, copy it for each spell cast before it
//	 this turn.)"
//
// Time travel is the shared keyword action (CR 701.56a); storm is the shared
// cast trigger (CR 702.40a), so each copy time travels too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48f80047-a4f9-4ebb-a97a-cad157b9bb64",
		Name:         "All of History, All at Once",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TimeTravel{}.Apply(ctx)
		},
	})
}
