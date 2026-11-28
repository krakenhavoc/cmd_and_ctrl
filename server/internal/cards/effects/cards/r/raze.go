package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raze — Sorcery {R} (Urza's Saga, S82 OLD batch 731, #5624):
//
//	"As an additional cost to cast this spell, sacrifice a land.
//	 Destroy target land."
//
// The land is sacrificed as the spell is cast (CR 601.2h), so it can't be the
// target and a response can't spend it twice; the target is a land on the
// battlefield judged at announce and again at resolution (CR 608.2b), any
// player's. Shard Volley's cost, Befoul's destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "02fde433-955d-49df-bd72-270b6feac9e7",
		Name:           "Raze",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a land", effects.Land()),
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Targets:        effects.TargetPermanent("target land", effects.Land()),
		OnResolve:      effects.DestroyEachLegalTarget,
	})
}
