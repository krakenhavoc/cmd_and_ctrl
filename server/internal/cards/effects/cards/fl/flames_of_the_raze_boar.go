package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flames of the Raze-Boar — Instant {5}{R} (Jumpstart, #5510):
//
//	"Flames of the Raze-Boar deals 4 damage to target creature an opponent
//	 controls. Then Flames of the Raze-Boar deals 2 damage to each other
//	 creature that player controls if you control a creature with power 4 or
//	 greater."
//
// "That player" is the target's controller, read before the damage (CR
// 608.2h), so the second part still reaches their other creatures if the
// target died. The power condition is checked after the first damage (the
// "Then"), against the creatures you control now, through the layers. The
// second part is a separate damage instance from the first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da8dd2ee-0ce3-4ce6-abb9-7444cefcbe83",
		Name:         "Flames of the Raze-Boar",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.DamageToTarget(0, 4)),
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, OpponentsOnly: true, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			owner := victim.Controller
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 4}).Apply(ctx); err != nil {
				return err
			}
			if !effects.S82OldB099ControlsPowerAtLeast(ctx.Game, ctx.Controller(), 4) {
				return nil
			}
			return effects.DamageEachMatching(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsCreature() && c.Controller == owner && c.InstanceID != id
			}, 2)
		},
	})
}
