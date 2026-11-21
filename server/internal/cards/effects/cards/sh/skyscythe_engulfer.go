package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyscythe Engulfer — Creature — Phyrexian Beast {5}{G}, 6/5 (Phyrexia:
// All Will Be One, #5479):
//
//	"Reach, trample
//	 This creature can't be blocked by creatures with flying."
//
// The evasion is a block rule on itself (Azure Beastbinder's shape): a
// flying blocker, read through the layered view as the block is declared
// (CR 509.1b), is refused; a reach creature without flying may block.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7cc831c-c1a1-4ea4-ae78-5be2b6173407",
		Name:            "Skyscythe Engulfer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.HasKeyword("flying"), "creatures with flying"),
		},
	})
}
