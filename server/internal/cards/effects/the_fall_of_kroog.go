package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Fall of Kroog — Sorcery {4}{R}{R} (The Brothers' War, #5482):
//
//	"Choose target opponent. Destroy target land that player controls.
//	 The Fall of Kroog deals 3 damage to that player and 1 damage to each
//	 creature they control."
//
// Two target clauses, the land's tied to the opponent (Ravager of the
// Fells' "that player controls", ControlledByTargetsController with a player
// anchor): the land must belong to the chosen opponent as the spell is cast
// (CR 601.2c) and again as it resolves (CR 608.2b). If the land is gone the
// spell still deals its damage to the opponent, and if the opponent is
// gone it does nothing to them. The damage is dealt all at once: the creature set is read before any is dealt and a damage
// replacement on a creature sees it with the player's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bfd8d447-7253-40d0-a153-78e9700045d7",
		Name:         "The Fall of Kroog",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true, OpponentsOnly: true},
			Answers: game.AnswerRemove,
		},
		Targets: Clauses(
			TargetPlayer("target opponent", Opponent()),
			ControlledByTargetsController(TargetPermanent("target land that player controls", Land()), 0),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				land, ok := ctx.ClauseTarget(1)
				if !ok || land.Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: land.ID}.Apply(ctx)
			}),
			Instruction(theFallOfKroogDamage),
		),
	})
}

// theFallOfKroogDamage is "deals 3 damage to that player and 1 damage to
// each creature they control", dealt together.
func theFallOfKroogDamage(ctx *Context) error {
	who, ok := ctx.ClauseTarget(0)
	if !ok || who.Kind != game.TargetPlayer {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == who.ID && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	src := ctx.Source()
	return ctx.Game.DamageInstanceForEffect(func() error {
		if err := (DealDamage{Source: src, Target: who.ID, Amount: 3}).Apply(ctx); err != nil {
			return err
		}
		for _, id := range ids {
			if err := (DealDamage{Source: src, Target: id, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
