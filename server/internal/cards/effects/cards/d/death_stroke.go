package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Stroke — Sorcery {B}{B} (Stronghold, #5627):
//
//	"Destroy target tapped creature."
//
// Swift Response's target clause and the shared destroyTheTargetPermanent
// body, at sorcery speed. "Tapped" is part of the clause, so it is checked
// as the spell is cast (CR 601.2c) and again as it resolves (CR 608.2b): a
// creature untapped in response is no longer a legal target and survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ebaa91f-5cbf-4a82-9759-9bd4d93a3e87",
		Name:         "Death Stroke",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
