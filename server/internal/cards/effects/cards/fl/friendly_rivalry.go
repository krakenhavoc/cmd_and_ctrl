package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Friendly Rivalry — Instant {R}{G} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Target creature you control and up to one other target legendary
//	 creature you control each deal damage equal to their power to target
//	 creature you don't control."
//
// Three target clauses. Each of the (one or two) creatures you control that
// is still legal deals damage equal to its power as it resolves, each being
// the source of its own damage (deathtouch, lifelink and prevention see it),
// together as one damage event (Bartz and Boko's DamageInstanceForEffect). The
// victim must still be legal; a damager that has become illegal deals none
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6850561-3c57-4354-8850-c467a6449563",
		Name:         "Friendly Rivalry",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("up to one other target legendary creature you control", effects.YouControl(), effects.Legendary()).WithCount(0, 1),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim, ok := ctx.ClauseTarget(2)
			if !ok || victim.Kind != game.TargetCard || !ctx.IsTargetLegal(victim) {
				return nil
			}
			type hit struct {
				source uuid.UUID
				power  int
			}
			var hits []hit
			for slot := 0; slot < 2; slot++ {
				t, ok := ctx.ClauseTarget(slot)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || c.CurrentPower() <= 0 {
					continue
				}
				hits = append(hits, hit{source: t.ID, power: c.CurrentPower()})
			}
			if len(hits) == 0 {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, h := range hits {
					if err := (effects.DealDamage{Source: h.source, Target: victim.ID, Amount: h.power}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
