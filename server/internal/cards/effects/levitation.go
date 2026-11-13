package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Levitation — Enchantment {2}{U}{U} (Urza's Legacy, #5623):
//
//	"Creatures you control have flying."
//
// Fervor's anthem with flying: a layer-6 grant (CR 613.1f) to every creature
// its controller controls, through TribalKeywordGrant with no tribe named. A
// creature that arrives later flies at once, and one that changes control
// stops.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94b703c4-5584-4913-8365-7e9f2f535c2d",
		Name:         "Levitation",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "flying"),
		},
	})
}
