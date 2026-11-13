package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primal Rage — Enchantment {1}{G} (Stronghold, #5627):
//
//	"Creatures you control have trample."
//
// Ajani's Anguish's layer 6 grant (TribalKeywordGrant over every creature
// you control, CR 613.1f). It is a static ability, so it applies to
// whatever you control at any moment, a creature that arrives later
// included (CR 611.3a), and trample only matters for an attacker's combat
// damage (CR 702.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6464ee4-23fc-4d68-bbda-3b53772015d1",
		Name:         "Primal Rage",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "trample"),
		},
	})
}
