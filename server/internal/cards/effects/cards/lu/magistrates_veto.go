package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magistrate's Veto — Enchantment {2}{R} (Mercadian Masques, #5619):
//
//	"White creatures and blue creatures can't block."
//
// Bedlam's block rule over the blockers that are white, blue or both,
// against every attacker (CR 509.1b). The colours are read as the block is
// declared, so a creature recoloured in response is judged as it stands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f50f8cac-35bf-4d4d-88c0-1c0c581f862c",
		Name:         "Magistrate's Veto",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnMatching(effects.And(effects.Creature(), effects.Or(effects.OfColor("W"), effects.OfColor("U")))), effects.OnMatching(effects.Creature()),
				"Magistrate's Veto: white creatures and blue creatures can't block"),
		},
	})
}
