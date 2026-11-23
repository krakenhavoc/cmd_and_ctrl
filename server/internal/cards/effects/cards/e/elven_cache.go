package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elven Cache — Sorcery {2}{G}{G} (Visions, #5631):
//
//	"Return target card from your graveyard to your hand."
//
// Regrowth's text at a higher cost, and Regrowth's body
// (returnTargetFromGraveyardToHand, regrowth.go). A target that left the
// graveyard in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a922366-ee2a-4ddb-a904-683dcf4f191a",
		Name:         "Elven Cache",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
	})
}
