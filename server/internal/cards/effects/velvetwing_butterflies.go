package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Velvetwing Butterflies // Gaze in Wonder — an Adventure card (CR 715,
// The Hobbit, #4454), oracle 70a702fe…:
//
//	Velvetwing Butterflies — Creature — Insect {2}{W}, 2/2
//	  "Flying"
//	Gaze in Wonder — Instant — Adventure {1}{W}
//	  "Tap one or two target creatures. (Then exile this card. You may
//	   cast the creature later from exile.)"
//
// No simplification.

// velvetwingButterfliesOracleID is shared by both faces of the card.
const velvetwingButterfliesOracleID = "70a702fe-1dd6-4d7b-a85a-45744ecad741"

func init() {
	Register(Spec{
		OracleID:        velvetwingButterfliesOracleID,
		Name:            "Velvetwing Butterflies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	Register(Spec{
		OracleID:     velvetwingButterfliesOracleID + "#1",
		Name:         "Gaze in Wonder",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("one or two target creatures").WithCount(1, 2),
		OnResolve:    func(item *game.StackItem, ctx *Context) error { return tapEachLegalTarget(ctx.Game, item) },
	})
}
