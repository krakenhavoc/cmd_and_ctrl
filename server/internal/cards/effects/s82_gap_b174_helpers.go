package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// DamageToFirstLegalTarget is "<source> deals N damage to <the target>" as
// an ability's whole body: the first still-legal target takes N, and a
// target that left leaves nothing to do (CR 608.2b). Chandra, Flame's
// Fury's +1 and Sarkhan, the Dragonspeaker's −3.
func DamageToFirstLegalTarget(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		t, ok := FirstLegalTarget(ctx)
		if !ok {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
	}
}

// PlusOneCounterOnFirstLegalTarget is "put a +1/+1 counter on <the
// target>" as an ability's whole body — Jace, Arcane Strategist's
// second-draw trigger.
func PlusOneCounterOnFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
