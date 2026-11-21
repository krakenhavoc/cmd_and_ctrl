package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Depart the Realm — Instant {1}{U} (Kaldheim, #5505):
//
//	"Return target nonland permanent to its owner's hand.
//	 Foretell {U}"
//
// Void Squall's bounce (bounceTheTarget) on a nonland permanent, plus
// the shared foretell special action (CR 702.143).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ae069ba4-8ca6-4eda-8f5d-fc82072d7d40",
		Name:           "Depart the Realm",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Targets:        effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		SpecialActions: []game.SpecialAction{effects.Foretell("{U}")},
		OnResolve:      effects.BounceTheTarget,
	})
}
