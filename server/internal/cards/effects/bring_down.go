package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bring Down — Sorcery {3}{W}:
//
//	"Assist (Another player can pay up to {3} of this spell's cost.)
//	 Destroy target creature with power 4 or greater."
//
// Killing Glare's shape with the power bound a fixed 4 instead of X:
// the pick is a creature with power 4 or greater (PowerGE), and the
// destroy is the plain DestroyTarget. A creature that has lost power
// below 4 in response is an illegal target (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:        "05d2b89b-acc4-4289-95a3-a09ebb23a5d6",
		Name:            "Bring Down",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist},
		Targets:         TargetCreature("target creature with power 4 or greater", PowerGE(4)),
		OnResolve:       destroyTheTargetPermanent,
	})
}
