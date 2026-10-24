package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Simulacrum — Instant {1}{B}:
//
//	"You gain life equal to the damage dealt to you this turn.
//	 Simulacrum deals damage to target creature you control equal to
//	 the damage dealt to you this turn."
//
// "The damage dealt to you this turn" is the sum of every damage event
// aimed at the caster since the turn began (g.EventsThisTurn, the
// bounded slice of this turn's log), combat and noncombat alike, from
// any source. Damage a shield prevented is never an event, so it does
// not count. It is read once as the spell resolves, so the life gained
// does not change what the creature is dealt, and a Simulacrum cast
// with nothing dealt to you does nothing. The damage to the creature is
// the spell's own, so it is subject to prevention and to the creature's
// protection like any other.
//
// The clause is Harmful (#4314): it damages a creature the caster controls,
// so a bot aims it at its least valuable one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20d69989-7250-40c7-a064-8ed78ccbe556",
		Name:         "Simulacrum",
		Completeness: CompletenessFull,
		Targets:      Harmful(TargetCreature("target creature you control", YouControl())),
		// The damage waits for the life gain's prompt (a CR 616 order; CR 608.2c, #4155). The
		// total is read once, as the spell begins to resolve, and filed for the second step.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				n := damageDealtToPlayerThisTurn(ctx.Game, ctx.Controller())
				ctx.RecordStep(simulacrumRecord, nil, n)
				return GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				n := ctx.StepRecord(simulacrumRecord).N()
				if n <= 0 {
					return nil
				}
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
			}),
		),
	})
}

// simulacrumRecord files the damage dealt to the caster this turn, read once.
const simulacrumRecord = "simulacrum-damage"

// damageDealtToPlayerThisTurn is "the damage dealt to you this turn":
// the total of this turn's damage events naming `player`.
func damageDealtToPlayerThisTurn(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventDealDamage && ev.Target == player && ev.Amount > 0 {
			n += ev.Amount
		}
	}
	return n
}
