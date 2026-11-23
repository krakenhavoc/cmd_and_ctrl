package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "7b3e74ad-0179-480b-871c-9e3bc30a43ff",
		Name:            "Aethersphere Harvester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Aethersphere Harvester", 2),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Pay {E}: This Vehicle gains lifelink until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.PayEnergy(1),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("lifelink", "Aethersphere Harvester — lifelink until end of turn"),
			},
			{
				Label:   "Crew 1",
				Cost:    effects.CrewCost(1),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Aethersphere Harvester"),
			},
		},
	})
}
