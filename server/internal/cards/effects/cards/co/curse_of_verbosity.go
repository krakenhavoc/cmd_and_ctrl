package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of Verbosity — Enchantment — Aura Curse {2}{U}:
//
//	"Enchant player
//	 Whenever enchanted player is attacked, you draw a card. Each
//	 opponent attacking that player does the same."
//
// Curse of Opulence's shape with a draw in place of a Gold token (see that
// card for the batching: one declaration is one "is attacked", and each
// attacking opponent draws once however many creatures they send). The
// shared bodies are in s82_rpr_b33_helpers.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6f76fa7-095e-4bfe-a38c-5c4531880880",
		Name:         "Curse of Verbosity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered:    effects.CurseOfVerbosityTriggers(),
	})
}
