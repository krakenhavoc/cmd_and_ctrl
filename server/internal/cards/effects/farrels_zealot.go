package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Farrel's Zealot — Creature — Human {1}{W}{W}, 2/2:
//
//	"Whenever this creature attacks and isn't blocked, you may have it deal 3 damage to
//	 target creature. If you do, this creature assigns no combat damage this turn."
//
// #3531: the target is chosen when the trigger goes on the stack and "you may" is asked as it
// resolves. On yes the damage is dealt, then the CR 510.1a gate stops the Zealot's own combat
// damage (not a prevention shield, so "damage can't be prevented" does not undo it).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0d063c0e-66bc-4a9c-98aa-a381b9cea120",
		Name:         "Farrel's Zealot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(unblockedAttackerMay("Farrel's Zealot", "deal 3 damage to target creature",
				func(ctx *Context, target uuid.UUID) error {
					return DealDamage{Source: ctx.Source(), Target: target, Amount: 3}.Apply(ctx)
				}), TargetCreature("target creature")),
		},
	})
}
