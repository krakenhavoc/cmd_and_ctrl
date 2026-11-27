package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death Denied — Instant — Arcane {X}{B}{B} (Saviors of Kamigawa, #5601):
//
//	"Return X target creature cards from your graveyard to your hand."
//
// Exactly X targets (CountFromX, as Long Rest and Back in Town), chosen as the
// spell is cast; each one still in the graveyard as it resolves comes back
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94f23a93-cd98-4ee7-9e62-9b7b49ace6fd",
		Name:         "Death Denied",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetsCountedByX(effects.TargetCardInGraveyard("X target creature cards from your graveyard", effects.YouOwn(), effects.Creature())),
		OnResolve:    effects.OldB587ReturnTargetsToHand,
	})
}
