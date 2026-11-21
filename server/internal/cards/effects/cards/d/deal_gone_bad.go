package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deal Gone Bad — Instant {3}{B} (Streets of New Capenna, #5491):
//
//	"Target creature gets -3/-3 until end of turn. Target player mills three
//	 cards."
//
// Two target clauses read by slot: the creature, then the player. One going
// illegal leaves the other half to resolve (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9ede348-f1dc-45ba-9e31-3129ccc182d1",
		Name:         "Deal Gone Bad",
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPlayer("target player"),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: -3, Toughness: -3, Label: "Deal Gone Bad — -3/-3"}).Apply(ctx); err != nil {
					return err
				}
			}
			if p, ok := ctx.ClauseTarget(1); ok && p.Kind == game.TargetPlayer {
				return effects.MillCards{Player: p.ID, N: 3}.Apply(ctx)
			}
			return nil
		},
	})
}
