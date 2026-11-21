package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nylea's Forerunner — Enchantment Creature — Beast {4}{G}, 5/3 (Theros Beyond Death, #5514):
//
//	"Trample
//	 Other creatures you control have trample."
//
// Trample is the printed keyword; the grant to the other creatures is
// Aggressive Mammoth's TribalKeywordGrant, which leaves the Forerunner
// itself out (it has its own).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bf0948b7-f7af-4d99-a8aa-ea7e6a69232d",
		Name:            "Nylea's Forerunner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Others: true, YoursOnly: true}, "trample"),
		},
	})
}
