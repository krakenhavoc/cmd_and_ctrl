package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Hippodrome — Plane — Segovia (Planechase Anthology, #4870):
//
//	"All creatures get -5/-0.
//	 Whenever chaos ensues, you may destroy target creature if its power
//	 is 0 or less."
//
// A plane (ADR 0216 §6). Any creature may be targeted; its power is read
// as the ability resolves (the ruling), and only then is the planar
// controller asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cde8d8c4-5e2b-452f-b3e9-bb9f2e745ec9",
		Name:         "The Hippodrome",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b9AllCreaturesGetPT("All creatures get -5/-0.", func(*game.Card, *game.Game, *game.Card) bool { return true }, -5, 0),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("The Hippodrome — you may destroy target creature if its power is 0 or less",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok || !hippodromePowerLow(g, id) {
						return nil
					}
					return MayChoice{
						Question: "The Hippodrome — destroy the target creature?",
						Then:     hippodromeDestroy,
						Carry:    []uuid.UUID{id},
					}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}

func hippodromePowerLow(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.IsCreature() && c.Effective().Power <= 0
}

// hippodromeDestroy is the keyed answer. An on-disk key.
var hippodromeDestroy = MayYes("confirm/the-hippodrome-destroy", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 || !hippodromePowerLow(ctx.Game, r.Carry[0]) {
		return nil
	}
	return DestroyTarget{Target: r.Carry[0]}.Apply(ctx)
})
