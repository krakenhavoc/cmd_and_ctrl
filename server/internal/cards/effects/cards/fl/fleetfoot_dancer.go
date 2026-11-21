package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleetfoot Dancer (Streets of New Capenna, #5491):
//
//	Trample, lifelink, haste
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8f606acf-db16-42cd-8907-dad88c4a1d54",
		Name:            "Fleetfoot Dancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "lifelink", "haste"},
	})
}
