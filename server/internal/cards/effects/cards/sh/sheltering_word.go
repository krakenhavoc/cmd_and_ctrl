package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sheltering Word — {1}{G} Instant: "Target creature you control gains
// hexproof until end of turn. You gain life equal to that creature's
// toughness."
//
// The life is the creature's toughness as the spell resolves; a creature
// that is no longer a legal target makes the whole spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77aaaad5-7960-453d-b024-6dacbca3d606",
		Name:         "Sheltering Word",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if !found {
				return nil
			}
			toughness := c.CurrentToughness()
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Sheltering Word — hexproof"}).Apply(ctx); err != nil {
				return err
			}
			if toughness <= 0 {
				return nil
			}
			return effects.GainLife{Player: item.Controller, Amount: toughness}.Apply(ctx)
		},
	})
}
