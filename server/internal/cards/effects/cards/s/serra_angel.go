package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra Angel — "Flying, vigilance."
//
// First S18 combat-keyword catalog card. No OnResolve / OnETB /
// Static — the PrintedKeywords slot feeds wire.go's synthesized
// Layer 6 static, which appends "flying" and "vigilance" to this
// card's Characteristic.Abilities on the battlefield. Off-
// battlefield (in hand, flash gating etc.) HasKeyword falls back
// to the catalog hook.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b7ac066-e5c7-43e6-9e7e-2739b24a905d",
		Name:            "Serra Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
