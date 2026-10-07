package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aethersphere Harvester — Artifact — Vehicle {3}, 3/5:
//
//	"Flying
//	 When this Vehicle enters, you get {E}{E} (two energy counters).
//	 Pay {E}: This Vehicle gains lifelink until end of turn.
//	 Crew 1 (Tap any number of creatures you control with total power 1
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// ADR 0129 PR 1 (#1995). Lifelink is granted to the Vehicle whether or
// not it is crewed yet, as printed; it matters once it deals damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b3e74ad-0179-480b-871c-9e3bc30a43ff",
		Name:            "Aethersphere Harvester",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Aethersphere Harvester", 2),
		},
		Activated: []ActivatedAbility{
			{
				Label:  "Pay {E}: This Vehicle gains lifelink until end of turn.",
				Cost:   PayEnergy(1),
				Effect: thisGainsKeywordUntilEndOfTurn("lifelink", "Aethersphere Harvester — lifelink until end of turn"),
			},
			{
				Label:  "Crew 1",
				Cost:   CrewCost(1),
				Effect: CrewEffect("Aethersphere Harvester"),
			},
		},
	})
}
