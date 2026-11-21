package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silhana Ledgewalker — Creature — Elf Rogue {1}{G}, 1/1 (Outlaws of
// Thunder Junction Commander, #5368):
//
//	"Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)
//	 This creature can't be blocked except by creatures with flying."
//
// Hexproof is a printed keyword; the evasion is Prowler's Helm's block
// rule over Self: every blocker without flying (read as it is when blocks
// are declared) is refused.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b15dcac-38dc-4957-bc01-40153ea07559",
		Name:            "Silhana Ledgewalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.HasKeyword("flying"), "creatures with flying"),
		},
	})
}
