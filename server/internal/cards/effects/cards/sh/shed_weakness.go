package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shed Weakness — {G} Instant (Amonkhet, S82 old batch 246):
//
//	"Target creature gets +2/+2 until end of turn. You may remove a -1/-1
//	 counter from it."
//
// The pump is applied first; then, if the creature has a -1/-1 counter, the
// spell's controller is asked whether to remove one ("you may"), so a creature
// without one is not offered the question. A target that is no longer legal
// gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3390776-66cc-4e0a-a034-701793b05db8",
		Name:         "Shed Weakness",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Shed Weakness — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || c.Counters[game.CounterMinusOne] <= 0 {
				return nil
			}
			return effects.MayChoice{
				Question: "Shed Weakness — remove a -1/-1 counter from the creature?",
				Then:     effects.MayRemoveAMinusOneCounterFromCarried,
				Carry:    []uuid.UUID{id},
			}.Apply(ctx)
		},
	})
}
