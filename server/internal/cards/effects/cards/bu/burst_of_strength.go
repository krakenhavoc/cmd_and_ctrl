package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burst of Strength — Instant {G} (Gatecrash, S82 OLD #5565):
//
//	"Put a +1/+1 counter on target creature and untap it."
//
// Battlegrowth's counter, then Provoke's untap, in printed order (CR 608.2c) on
// the one target: a target gone by resolution makes the whole spell do nothing
// (CR 608.2b). An already-untapped creature stays untapped and still gets the
// counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ab592fa-adfb-4879-af6c-0204eab323fb",
		Name:         "Burst of Strength",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.UntapTarget{Target: id}.Apply(ctx)
			}),
		),
	})
}
