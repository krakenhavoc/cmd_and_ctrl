package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aetherize — Instant {3}{U}:
//
//	"Return all attacking creatures to their owner's hand."
//
// A one-sided fog that answers an alpha strike permanently rather
// than for a turn, which is why a flash deck wants it.
//
// Worth noting why this one is expressible while the deck's attack
// TRIGGERS are not: there is no EventAttack in events.go, so nothing
// can fire "whenever ~ attacks". But combat STATE exists —
// DeclareAttacker stamps Card.AttackingTarget, and ClearCombat wipes
// it at end of combat. Aetherize is a spell reading that state at
// resolution, not a trigger waiting on an event, so it needs no new
// plumbing. Untargeted and symmetric: it takes the caster's own
// attackers too.
//
// S23: this is ReturnAllToHand's reason to exist. "All attacking
// creatures" is not a battlefield predicate — attacking-ness lives in
// the combat state, not on the card's characteristics — so the set is
// computed here and handed to the primitive, which bounces it as one
// simultaneous event like any other mass effect.
func init() {
	Register(Spec{
		OracleID: "7c779721-cd1b-4696-9ae9-68ccc284ed2a",
		Name:     "Aetherize",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var attackers []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && c.AttackingTarget != uuid.Nil {
					attackers = append(attackers, c.InstanceID)
				}
			}
			return ReturnAllToHand{Cards: attackers}.Apply(ctx)
		},
	})
}
