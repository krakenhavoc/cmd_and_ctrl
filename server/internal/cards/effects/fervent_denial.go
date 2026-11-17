package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fervent Denial — Instant {3}{U}{U} (Odyssey, #5613):
//
//	"Counter target spell.
//	 Flashback {5}{U}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Counterspell's sentence (CR 701.6a) on the shared counterTheTargetSpell
// body. Flashed back, it is still an instant, so it answers a spell from
// the graveyard at instant speed.
//
// Flashback is the shared constructor: the graveyard cast, the price,
// and CR 702.34a's exile as the spell leaves the stack (resolved,
// countered, or with its target gone).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2e56bc34-1e35-4b95-9525-cb58ac8ba060",
		Name:             "Fervent Denial",
		Completeness:     CompletenessFull,
		Targets:          TargetSpell("target spell"),
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{U}{U}")},
		OnResolve:        counterTheTargetSpell,
	})
}
