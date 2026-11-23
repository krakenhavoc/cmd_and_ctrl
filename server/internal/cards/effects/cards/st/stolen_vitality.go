package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stolen Vitality — Instant {1}{R} (Innistrad: Midnight Hunt, #5497):
//
//	"Target creature gets +3/+1 until end of turn. If it's your turn, that
//	 creature gains trample until end of turn. Otherwise, it gains first
//	 strike until end of turn."
//
// "Your turn" is read as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd7454f3-57bc-44f4-bd58-54e883c699fd",
		Name:         "Stolen Vitality",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 1)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !effects.OnBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 1, Label: "Stolen Vitality — +3/+1"}).Apply(ctx); err != nil {
				return err
			}
			kw := "first strike"
			if effects.IsYourTurn(ctx.Game, ctx.Controller()) {
				kw = "trample"
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{kw}, Label: "Stolen Vitality — " + kw}.Apply(ctx)
		},
	})
}
