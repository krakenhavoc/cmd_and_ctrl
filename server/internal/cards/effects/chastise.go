package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chastise — Instant {3}{W} (Judgment, #5611):
//
//	"Destroy target attacking creature. You gain life equal to its
//	 power."
//
// The target is judged against combat at announce and at resolution
// (AttackingCreature, CR 608.2b), so a creature removed from combat in
// response counters the spell. The power is read before the
// destruction, its last state on the battlefield (CR 608.2h), and the
// caster gains it whether or not the creature was destroyed: an
// indestructible attacker still pays out.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b7553f3f-5de1-409c-a184-e12e40f017ab",
		Name:         "Chastise",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				card, ok := ctx.Game.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				power := card.CurrentPower()
				if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				return GainLife{Amount: power}.Apply(ctx)
			}),
		),
	})
}
