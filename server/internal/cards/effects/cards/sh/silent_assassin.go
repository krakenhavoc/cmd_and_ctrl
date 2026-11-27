package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silent Assassin — Creature — Human Mercenary Assassin {B}{B}, 2/1 (Mercadian
// Masques, S82 OLD, #5619):
//
//	"{3}{B}: Destroy target blocking creature at end of combat."
//
// The target must be blocking as the ability is activated. The destruction is a
// delayed trigger (CR 603.7) for the beginning of this combat's end of combat
// step, pinned to the creature as the object it is now (CR 400.7), so a creature
// that has left or changed is not destroyed, and a target that is no longer
// blocking by then is still destroyed (the ability is on the stack, not the creature's state).
// Winter's Chill's delayed destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05b07f1f-42e1-41fb-8858-5e4a311c4ed9",
		Name:         "Silent Assassin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}: Destroy target blocking creature at end of combat.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.ManaCost("{3}{B}"),
			Targets: effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				return effects.ScheduleDelayedTrigger{
					At:     game.StepEndCombat,
					Label:  "Silent Assassin — destroy " + c.Name,
					Cards:  []uuid.UUID{id},
					Body:   effects.DestroyTheObjectBody,
					Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
				}.Apply(ctx)
			},
		}},
	})
}
