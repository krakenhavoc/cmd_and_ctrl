package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bd7454f3-57bc-44f4-bd58-54e883c699fd",
		Name:         "Stolen Vitality",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 1)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !onBattlefield(ctx.Game, t.ID) {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 1, Label: "Stolen Vitality — +3/+1"}).Apply(ctx); err != nil {
				return err
			}
			kw := "first strike"
			if IsYourTurn(ctx.Game, ctx.Controller()) {
				kw = "trample"
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{kw}, Label: "Stolen Vitality — " + kw}.Apply(ctx)
		},
	})
}
