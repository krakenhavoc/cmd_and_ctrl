package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Masterful Replication — Instant {5}{U}:
//
//	"Choose one —
//	 • Create two 3/3 colorless Golem artifact creature tokens.
//	 • Choose target artifact you control. Each other artifact you
//	   control becomes a copy of that artifact until end of turn."
//
// The first bullet is Mass Production's tokens. The second is Brudiclad's
// BecomeCopy over the artifacts you control as the spell resolves, the
// target excluded: the copiable values are the target's as the effect
// begins (CR 707.2), and the effect ends at cleanup (CR 611.2). A target
// that is illegal as the spell resolves does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc6e6e16-da2b-4fc0-91ad-b76f0a101306",
		Name:         "Masterful Replication",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Create two 3/3 colorless Golem artifact creature tokens.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: TokenCard("3/3 colorless Golem artifact creature"), N: 2}.Apply(ctx)
				}),
			ModeDoing("Choose target artifact you control. Each other artifact you control becomes a copy of that artifact until end of turn.",
				TargetPermanent("target artifact you control", Artifact(), YouControl()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || !onBattlefield(ctx.Game, t.ID) {
						return nil
					}
					var others []uuid.UUID
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Controller == ctx.Controller() && c.IsArtifact() && c.InstanceID != t.ID {
							others = append(others, c.InstanceID)
						}
					}
					if len(others) == 0 {
						return nil
					}
					return BecomeCopy{
						Targets:  others,
						Of:       t.ID,
						Duration: CopyUntilEndOfTurn,
						Label:    "Masterful Replication — becomes a copy until end of turn",
					}.Apply(ctx)
				}),
		),
	})
}
