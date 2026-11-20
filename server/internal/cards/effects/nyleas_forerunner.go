package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bf0948b7-f7af-4d99-a8aa-ea7e6a69232d",
		Name:            "Nylea's Forerunner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Others: true, YoursOnly: true}, "trample"),
		},
	})
}
