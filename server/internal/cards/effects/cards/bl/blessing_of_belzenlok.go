package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blessing of Belzenlok — Instant {B} (Dominaria, S82 OLD batch 214, #5530):
//
//	"Target creature gets +2/+1 until end of turn. If it's legendary, it also
//	 gains lifelink until end of turn."
//
// The pump and, when the target is legendary as the spell resolves (read
// through the layers, so a creature that was made legendary counts), the
// lifelink grant. A target that is no longer legal does neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4a03986-4db6-4114-9a58-5a8a3a7a2461",
		Name:         "Blessing of Belzenlok",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 1)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 1,
				Label: "Blessing of Belzenlok — +2/+1 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.RecomputeLayersIfStaleLocked()
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !c.IsLegendary() {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, "Blessing of Belzenlok — lifelink until end of turn", game.AddKeywordsMod("lifelink"))
		},
	})
}
