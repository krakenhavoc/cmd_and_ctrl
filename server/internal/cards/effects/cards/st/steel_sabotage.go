package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steel Sabotage — Instant {U} (Mirrodin Besieged, S82 OLD b467, #5576):
//
//	"Choose one —
//	 • Counter target artifact spell.
//	 • Return target artifact to its owner's hand."
//
// One mode is chosen as the spell is cast (CR 700.2), and only that mode's
// target is declared. The counter is Artifact Blast's clause; the return is
// the shared bounce, read again as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e71fc56-0332-4e5e-8658-9257cc0d8a1d",
		Name:         "Steel Sabotage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.ModeDoing("Counter target artifact spell.", effects.TargetSpell("target artifact spell", effects.Artifact()), effects.CounterTheModesTarget),
			effects.ModeDoing("Return target artifact to its owner's hand.", effects.TargetPermanent("target artifact", effects.Artifact()), effects.BounceTheModesTarget),
		),
	})
}
