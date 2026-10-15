package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Vandal — Creature — Goblin Rogue {R}, 1/1:
//
//	"Whenever this creature attacks and isn't blocked, you may pay {R}. If you do, destroy target
//	 artifact defending player controls and this creature assigns no combat damage this turn."
//
// #3531: the artifact is targeted as the trigger goes on the stack; the {R} is the "if you do",
// paid from the pool or by auto-tap, and the destruction and the gate follow it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "50bebbb9-01b7-4eb6-8efd-307c9bcd2517",
		Name:         "Goblin Vandal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(WhenAttacksAndIsNotBlockedEffect("Goblin Vandal — you may pay {R}: destroy target artifact defending player controls",
				func(g *game.Game, item *game.StackItem, _ uuid.UUID) error {
					ctx := NewContext(g, item)
					target := firstLegalCardTarget(ctx)
					return MayPay{
						Chooser:  ctx.Controller(),
						Cost:     "{R}",
						Question: "Goblin Vandal — pay {R} to destroy the artifact? (it then assigns no combat damage this turn)",
						OnPay: func(ctx *Context) error {
							if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
								return err
							}
							assignsNoCombatDamageThisTurn(ctx, "Goblin Vandal — assigns no combat damage this turn")
							return nil
						},
					}.Apply(ctx)
				}), TargetOfDefendingPlayer("artifact", "target artifact defending player controls", Artifact())),
		},
	})
}
