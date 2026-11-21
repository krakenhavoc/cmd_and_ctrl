package cardsst

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swallowed by Leviathan — Instant {2}{U} (Final Fantasy, #4462):
//
//	"Choose target spell. Surveil 2, then counter the chosen spell unless
//	 its controller pays {1} for each card in your graveyard."
//
// Two steps in printed order (ResolveSteps, CR 608.2c): the surveil asks
// first, and the price is counted only after it is answered, so cards
// the surveil put into the graveyard raise it. The counter is Logic
// Knot's CounterUnlessPaid, which halts the table until the spell's
// controller answers and asks nothing if the spell has already left the
// stack. An illegal target fizzles the whole spell, surveil included
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a8185ed-636b-49ad-bf24-2ab02c7fb91e",
		Name:         "Swallowed by Leviathan",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.ResolveSteps(effects.Surveil{N: 2}, effects.Instruction(swallowedByLeviathanCounter)),
	})
}

// swallowedByLeviathanCounter is "counter the chosen spell unless its
// controller pays {1} for each card in your graveyard", counted now.
func swallowedByLeviathanCounter(ctx *effects.Context) error {
	if len(ctx.Item.Targets) == 0 {
		return nil
	}
	n := 0
	if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Graveyard != nil {
		n = p.Graveyard.Size()
	}
	cost := "{" + strconv.Itoa(n) + "}"
	return effects.CounterUnlessPaid{
		StackID:  ctx.Item.Targets[0].ID,
		Cost:     cost,
		Question: "Swallowed by Leviathan — pay " + cost + " or your spell is countered",
	}.Apply(ctx)
}
