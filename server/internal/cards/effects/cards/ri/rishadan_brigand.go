package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rishadan Brigand — Creature — Human Pirate {4}{U}, 3/2 (Mercadian Masques, #5619):
//
//	"Flying
//	 When this creature enters, each opponent sacrifices a permanent of
//	 their choice unless they pay {3}.
//	 This creature can block only creatures with flying."
//
// Flying is the printed keyword and the last line is Welkin Tern's block rule
// (CR 509.1b). Each opponent is asked for {3} as the trigger resolves; one who does not
// pay sacrifices a permanent of their choice. Shared body:
// RishadanEachOpponentSacrificesUnlessTheyPay (s82_old_b696_helpers.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae2db472-c699-4c39-aa35-c359ac42df4e",
		Name:            "Rishadan Brigand",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rishadan Brigand — each opponent sacrifices a permanent unless they pay {3}",
				effects.RishadanEachOpponentSacrificesUnlessTheyPay("Rishadan Brigand", "{3}")),
		},
	})
}
