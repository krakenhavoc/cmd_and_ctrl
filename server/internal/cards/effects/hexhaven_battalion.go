package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hexhaven Battalion — Sorcery {4}{W}{W} (Reality Fracture, tracker
// #2795):
//
//	"Create three 2/2 colorless Wizard Soldier creature tokens named
//	 Cadet. Empower Jace 2.
//	 Basic landcycling {2}"
//
// Three Cadets, then the keyword action (ADR 0139); basic landcycling is
// the ordinary typecycling activation from hand (#660).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "158823ed-6015-437a-82b0-8dadb1e9aab2",
		Name:         "Hexhaven Battalion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		// Empower Jace waits for a prompt the tokens' creation asks (a CR 616 order; CR 608.2c,
		// #4155).
		OnResolve: ResolveSteps(
			CreateToken{
				Template: TokenCard("2/2 colorless Wizard Soldier named Cadet"),
				N:        3,
			},
			EmpowerJace{N: 2},
		),
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
