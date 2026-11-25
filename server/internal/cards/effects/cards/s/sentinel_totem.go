package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentinel Totem — Artifact {1} (Ixalan, S82 older sets):
//
//	"When this artifact enters, scry 1.
//	 {T}, Exile this artifact: Exile all graveyards."
//
// The scry is a fixed rider (Purpose Scry 1). Exiling the totem is a cost
// (CR 118.3), so the ability resolves even if the totem is destroyed in
// response; every graveyard is exiled (Scavenger Grounds' B02ExileAllGraveyards).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5432bd43-e92c-464b-9ad2-3515f3042799",
		Name:         "Sentinel Totem",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sentinel Totem — scry 1", effects.Do(effects.Scry{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Exile this artifact: Exile all graveyards.",
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Cost:    effects.Plus(effects.TapCost(), effects.ExileThis()),
			Effect:  effects.B02ExileAllGraveyards,
		}},
	})
}
