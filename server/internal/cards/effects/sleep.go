package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sleep —
//
// "Tap all creatures target player controls. Those creatures don't untap
// during that player's next untap step."
func init() {
	Register(Spec{
		OracleID:     "9b93ff69-f195-4d72-8e1d-574c3e53bca8",
		Name:         "Sleep",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			ids := b751CreatureIDs(ctx.Game, ts[0].ID)
			return (TapAndFreeze{Targets: ids, Player: ts[0].ID, Label: "Sleep"}).Apply(ctx)
		},
	})
}
