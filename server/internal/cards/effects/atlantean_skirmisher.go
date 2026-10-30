package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Atlantean Skirmisher — Creature — Merfolk Rogue {1}{U}, 1/2
// (Marvel Super Heroes, #4455):
//
//	"Whenever this creature attacks, it connives. (Draw a card, then
//	 discard a card. If you discarded a nonland card, put a +1/+1
//	 counter on this creature.)"
//
// The attack-side twin of A.I.M. Scientists' enters connive: the same
// ConniveEach for the one creature, here the attacker itself. The
// trigger is the ThisAttacked shape, and the counter goes on the
// Skirmisher only when it is still there (CR 701.50c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f1c21d6-bf85-4e31-a978-6423c40650d2",
		Name:         "Atlantean Skirmisher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Atlantean Skirmisher — it connives",
				Do(Instruction(sourceConnives))),
		},
	})
}

// sourceConnives makes the ability's source connive (CR 701.50): shared
// by A.I.M. Scientists' enter trigger and the Skirmisher's attack trigger.
func sourceConnives(ctx *Context) error {
	return ConniveEach{Controller: ctx.Controller(), Creatures: []uuid.UUID{ctx.Item.SourceCardID}}.Apply(ctx)
}
