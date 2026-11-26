package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b338_helpers.go — Fate Reforged, batch 338 (#5552).

// S82OldB338DamageEachClauseTarget is "<source> deals N damage to <the first
// target> and N damage to <the second>": one printed instruction, so one damage
// instance (CR 615.8), dealt to whichever of the two clauses still has a legal
// target (CR 608.2b).
func S82OldB338DamageEachClauseTarget(ctx *Context, clauses, n int) error {
	return ctx.Game.DamageInstanceForEffect(func() error {
		for slot := 0; slot < clauses; slot++ {
			if t, ok := ctx.ClauseTarget(slot); ok {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// S82OldB338ControlsGreatestToughness is "you control the creature with the
// greatest toughness or tied for the greatest toughness" (Abzan Beastmaster):
// `controller` controls a creature whose toughness is at least every other
// creature's, read on the layered board.
func S82OldB338ControlsGreatestToughness(g *game.Game, controller uuid.UUID) bool {
	g.RecomputeLayersIfStaleLocked()
	best, mine, any := 0, 0, false
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsCreature() {
			continue
		}
		t := c.CurrentToughness()
		if !any || t > best {
			best = t
		}
		any = true
		if c.Controller == controller && t > mine {
			mine = t
		}
	}
	if !any {
		return false
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.CurrentToughness() == best {
			return true
		}
	}
	return false
}
