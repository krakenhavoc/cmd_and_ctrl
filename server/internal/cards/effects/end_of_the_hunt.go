package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// End of the Hunt — Sorcery {1}{B} (Secrets of Strixhaven, #4456):
//
//	"Target opponent exiles a creature or planeswalker they control with
//	 the greatest mana value among creatures and planeswalkers they
//	 control."
//
// Break Under Pressure's edict with exile as the verb: the TARGETED
// opponent is the one asked, and chooses among the tied greatest (the
// permanent is not targeted, so hexproof and protection are irrelevant).
// The comparison reads the opponent's creatures and planeswalkers as the
// spell resolves, mana value taken on the battlefield (X is zero there, CR
// 202.3e; a token with no mana cost is zero). The pick is the exile edict's
// (msc17ExileEdict, Doomfall): a permanent that left before the answer is
// withdrawn, and an opponent with nothing to exile is not asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74b8182e-54be-444a-97b1-84fc5fd3c290",
		Name:         "End of the Hunt",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			return msc17ExileEdict(ctx, t.ID,
				"End of the Hunt — exile a creature or planeswalker with the greatest mana value",
				b10GreatestManaValueCreatureOrPlaneswalkerYouControl(), doomfallExile)
		},
	})
}
