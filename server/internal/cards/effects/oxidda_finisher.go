package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oxidda Finisher — Creature — Ogre Rebel {5}{R}{R}, 7/5 (Phyrexia: All
// Will Be One, #5479):
//
//	"Affinity for Equipment (This spell costs {1} less to cast for each
//	 Equipment you control.)
//	 Trample"
//
// Affinity is a self cost modifier counting the Equipment you control
// (CR 702.41a); only generic mana is reduced.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f308a1de-94a6-4f9f-8f46-6f2018959b40",
		Name:            "Oxidda Finisher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Equipment", Subtype("Equipment")),
		},
	})
}
