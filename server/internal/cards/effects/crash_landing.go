package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crash Landing — Instant {2}{G} (Guildpact, #5599):
//
//	"Target creature with flying loses flying until end of turn. Crash
//	 Landing deals damage to that creature equal to the number of Forests
//	 you control."
//
// The target must have flying as the spell is cast and again as it
// resolves (CR 608.2b), so a creature that lost flying in response is an
// illegal target and nothing happens. Then, in printed order, a layer 6
// removal with its own timestamp (a later grant of flying gives it back)
// and the damage. The Forests are counted as it resolves, by land type,
// not only basic Forests (rockalancheForests).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d36e5ee-1e23-48fe-ad13-081cc153342b",
		Name:         "Crash Landing",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, 0, 0)),
		Targets:      TargetCreature("target creature with flying", HasKeyword("flying")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := untilEndOfTurn(ctx, id, nil, "Crash Landing — loses flying", game.RemoveKeywordsMod("flying")); err != nil {
				return err
			}
			return DealDamage{Source: ctx.Source(), Target: id, Amount: rockalancheForests(ctx)}.Apply(ctx)
		},
	})
}
