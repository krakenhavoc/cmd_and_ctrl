package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Set Adrift — Sorcery {5}{U}:
//
//	"Delve. Put target nonland permanent on top of its owner's library."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// The tuck is the last instruction on the card, so the fire-and-forget
// form is right: nothing reads where the permanent went. No
// simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f27f51e1-b881-4639-9d36-47cdf2a61117",
		Name:         "Set Adrift",
		Completeness: effects.CompletenessFull,
		Delve:        true,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
