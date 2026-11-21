package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Windborn Muse — Creature — Spirit {3}{W}, 2/3 (EDHREC top 1,000, #3654):
//
//	"Flying
//	 Creatures can't attack you unless their controller pays {2} for
//	 each creature they control that's attacking you."
//
// Propaganda's attack tax on a flyer: the tax is paid once per attacker
// as attackers are declared (CR 508.1h), exactly as for Ghostly Prison.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "75e09ac6-d789-4c4d-9aca-845d68f3899f",
		Name:            "Windborn Muse",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AttackTaxes: []game.AttackTax{
			effects.AttackTax("{2}", "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you."),
		},
	})
}
