package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Order of Sacred Dusk — Creature — Vampire Knight {6}{W}{B}, 5/5
// (Commander, #5368):
//
//	"Convoke
//	 Flying, lifelink, haste
//	 Exalted
//	 Other Vampires you control have exalted."
//
// Convoke is the spell's tap cost (Convoke()); flying, lifelink, haste
// and exalted are printed keywords (exalted's trigger is derived by the
// engine, game/exalted.go, once per instance). The static grants exalted
// to every OTHER Vampire its controller controls (TribeFilter Others +
// YoursOnly), read through the layers, so a changeling counts. Each
// granted instance triggers on its own (CR 702.83b), so two exalted
// permanents make +2/+2 for a creature attacking alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0cb41c3f-5e6a-4c63-80b5-c9ab84b4e4bf",
		Name:            "Order of Sacred Dusk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink", "haste", "exalted"},
		TapCost:         Convoke(),
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Vampire"}, Others: true, YoursOnly: true}, game.KeywordExalted),
		},
	})
}
