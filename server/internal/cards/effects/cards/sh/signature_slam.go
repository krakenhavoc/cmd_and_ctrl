package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Signature Slam — Instant {2}{G} (Modern Horizons 3, #5948):
//
//	"Put a +1/+1 counter on target creature you control, then each
//	 modified creature you control deals damage equal to its power to
//	 target creature you don't control. (Equipment, Auras you control, and
//	 counters are modifications.)"
//
// Two target clauses read positionally (Bite Down's shape): slot 0 gets the
// counter, slot 1 takes the damage. The set of modified creatures is read
// after the counter lands (CR 700.9, b07IsModified), so the target itself
// counts, and each one's power is read once, as the damage is dealt
// together. If the first target is gone the counter is skipped but the
// damage still happens (CR 608.2b); if the second is gone nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ae23752-47a6-4031-8398-b8f6c77cfaad",
		Name:         "Signature Slam",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // damages a creature
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if mine, ok := ctx.ClauseTarget(0); ok && mine.Kind == game.TargetCard {
				if err := (effects.AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			type hit struct {
				id    uuid.UUID
				power int
			}
			var hits []hit
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.IsCreature() && effects.B07IsModified(ctx.Game, c) {
					hits = append(hits, hit{c.InstanceID, c.CurrentPower()})
				}
			}
			// One printed instruction: all the damage is one instance (CR 615.8).
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, h := range hits {
					if err := (effects.DealDamage{Source: h.id, Target: victim.ID, Amount: h.power}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
