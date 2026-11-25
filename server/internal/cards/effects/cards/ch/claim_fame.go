package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Claim // Fame — Sorcery // Sorcery with aftermath (Hour of Devastation):
//
//	Claim {B}: "Return target creature card with mana value 2 or less from
//	 your graveyard to the battlefield."
//	Fame {1}{R}: "Aftermath (Cast this spell only from your graveyard. Then
//	 exile it.) Target creature gets +2/+0 and gains haste until end of turn."
//
// Two faces under the oracle ID and its "#1" suffix; aftermath is the engine's
// (CR 702.127a).
func init() {
	const oracle = "16da9e4b-c6ee-4e70-8653-31ca147f3b01"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Claim",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Fame",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "haste")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Fame — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"}, Label: "Fame — haste"}.Apply(ctx)
		},
	})
}
