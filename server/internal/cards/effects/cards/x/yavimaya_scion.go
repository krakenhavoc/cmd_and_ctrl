package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yavimaya Scion — Creature — Treefolk {4}{G} 4/4: "Protection from artifacts".
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6e06b5cc-444c-49c9-b08a-81ef8d97c9ab",
		Name:            "Yavimaya Scion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from artifacts"},
	})
}
