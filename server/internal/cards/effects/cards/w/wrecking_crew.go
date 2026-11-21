package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrecking Crew (Streets of New Capenna, #5491):
//
//	Reach, trample
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "533c4956-87f8-4bae-8924-166fc2e9ec1f",
		Name:            "Wrecking Crew",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
	})
}
