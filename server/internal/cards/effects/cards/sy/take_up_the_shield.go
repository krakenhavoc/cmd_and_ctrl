package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Take Up the Shield — {1}{W} Instant:
//
//	"Put a +1/+1 counter on target creature. It gains lifelink and
//	 indestructible until end of turn."
//
// The counter is permanent; the keywords last until end of turn
// (CR 611.2a). If the target is gone as the spell resolves, nothing happens
// (CR 608.2b).
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8d0eafe-540a-4b2e-a989-e98ff3c31105",
		Name:         "Take Up the Shield",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink", "indestructible")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"lifelink", "indestructible"},
					Label:    "Take Up the Shield — lifelink and indestructible",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
