package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrel Down Sokenzan — Instant — Arcane {2}{R} (Saviors of Kamigawa,
// #5601):
//
//	"Sweep — Return any number of Mountains you control to their owner's
//	 hand. Barrel Down Sokenzan deals damage to target creature equal to
//	 twice the number of Mountains returned this way."
//
// The Mountains are picked as the spell resolves (OldB591AskSweep) and
// returned together; the damage is twice the number that reached a hand,
// dealt by the spell to the target creature if it is still there
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9ecb346-7745-4106-a3e1-c467a3b91ec7",
		Name:         "Barrel Down Sokenzan",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.FirstLegalTarget(ctx); !ok {
				return nil
			}
			return effects.OldB591AskSweep(ctx, "Mountain", "Barrel Down Sokenzan — return any number of Mountains you control to their owner's hand", barrelDownSokenzanPick)
		},
	})
}

// barrelDownSokenzanPick returns the Mountains, then deals the damage. An
// on-disk key.
var barrelDownSokenzanPick = effects.AskedPick("choose-cards/barrel-down-sokenzan-mountains", func(ctx *effects.Context, r game.CardsChosen) error {
	n, err := effects.OldB591ReturnSwept(ctx, r.Picked)
	if err != nil || n <= 0 {
		return err
	}
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2 * n}.Apply(ctx)
})
