package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Break Under Pressure — Instant {2}{B}:
//
//	"Target opponent sacrifices a creature or planeswalker with the
//	 greatest mana value among creatures and planeswalkers they control.
//	 You gain 2 life."
//
// Soul Shatter's edict aimed at one opponent: the sacrifice is THEIR
// choice among the tied greatest (so hexproof is irrelevant to the
// permanent), and the targeted opponent is the only one asked. The
// life is gained whether or not they had anything to sacrifice.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "8dd23c50-cc84-4064-89cd-8d52fbfbd4ab",
		Name:         "Break Under Pressure",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := firstLegalTarget(ctx); ok && t.Kind == game.TargetPlayer {
				ctx.Game.PlayerSacrificesForEffect(ctx.Source(), t.ID,
					sacrificeSpec("a creature or planeswalker with the greatest mana value",
						b10GreatestManaValueCreatureOrPlaneswalkerYouControl()),
					"Break Under Pressure — sacrifice a creature or planeswalker with the greatest mana value")
			}
			return GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
