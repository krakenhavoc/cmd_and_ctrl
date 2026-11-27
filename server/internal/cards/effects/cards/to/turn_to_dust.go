package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turn to Dust — Instant {G} (Mirrodin, S82 OLD, #5606):
//
//	"Destroy target Equipment. Add {G}."
//
// Two steps in order, each waiting on the one before: the Equipment is
// destroyed, then its caster adds {G}, which goes into their pool as the
// spell resolves (and empties at the end of the step, CR 500.5). The
// target is judged on resolution (CR 608.2b): with the Equipment gone
// the spell does nothing, and no mana is added.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56828166-eaa3-4711-91b6-401a3e3b733f",
		Name:         "Turn to Dust",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target Equipment", effects.HasSubtype("Equipment")),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.DestroyTarget{Target: id}.Apply(ctx) }),
			effects.AddMana{Produced: "{G}"},
		),
	})
}
