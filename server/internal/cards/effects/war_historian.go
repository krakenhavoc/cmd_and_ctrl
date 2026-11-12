package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War Historian — Creature — Human Monk {2}{G}, 3/3 (March of the
// Machine, #5649):
//
//	"Reach
//	 This creature has indestructible as long as it attacked a battle
//	 this turn."
//
// The static reads the turn's attack records: it is on from the moment
// War Historian is declared as an attacker attacking a battle (CR 508.1b)
// and stays on until the turn ends, after the combat is over and whether
// or not the battle is still on the battlefield. Being put onto the
// battlefield attacking a battle is not attacking one (CR 508.4), and a
// War Historian that left and came back is a new object that attacked
// nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2676d0c3-0a24-4e93-ae80-8f3fe03da483",
		Name:            "War Historian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static:          []game.StaticAbility{SelfKeywordWhileItAttackedABattleThisTurn("indestructible")},
	})
}
