package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slaughter — Instant {2}{B}{B} (Exodus, #5782):
//
//	"Buyback—Pay 4 life. (You may pay 4 life in addition to any other
//	 costs as you cast this spell. If you do, put this card into your
//	 hand as it resolves.)
//	 Destroy target nonblack creature. It can't be regenerated."
//
// The buyback pays a fixed amount of life (BuybackPayLife, ADR 0100's
// 2026-11-20 amendment): it is paid with the rest of the total cost
// (CR 601.2h), and it can be announced only by a caster whose life total
// is at least 4 (CR 119.4). The return to hand is the engine's (CR
// 702.27a). The target is judged at announce and again at resolution
// (CR 608.2b), so a creature that turned black in response is spared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "37de575d-15d4-42a2-a3d2-3855fb6ab110",
		Name:          "Slaughter",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		Targets:       TargetCreature("target nonblack creature", NonBlack()),
		OptionalCosts: []game.AdditionalCost{BuybackPayLife(4)},
		OnResolve:     destroyTheTargetPermanentNoRegen,
	})
}
