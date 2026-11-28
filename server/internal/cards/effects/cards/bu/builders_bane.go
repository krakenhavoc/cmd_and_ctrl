package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Builder's Bane — Sorcery {X}{X}{R} (Mirage, #5633):
//
//	"Destroy X target artifacts. Builder's Bane deals damage to each player equal
//	 to the number of artifacts they controlled that were put into a graveyard
//	 this way."
//
// The target count is the announced X; the destruction is one simultaneous event
// and each player's count is read from the pre-move copies (OldB794BuildersBane).
//
// No simplification.
func init() {
	spec := effects.TargetPermanent("X target artifacts", effects.Artifact())
	spec.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "814adb99-6100-4c36-96d9-763bc20dca5d",
		Name:         "Builder's Bane",
		Completeness: effects.CompletenessFull,
		Targets:      spec,
		OnResolve:    effects.OldB794BuildersBane,
	})
}
