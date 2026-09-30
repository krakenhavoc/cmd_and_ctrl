package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expedite — Instant {R}:
//
//	"Target creature gains haste until end of turn.
//	 Draw a card."
//
// The haste grant is Act of Treason's own until-end-of-turn clause,
// GrantKeywordUntilEOT, pinned to the chosen target instead of a
// creature just stolen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3501a839-eef5-44e4-8637-b5754780454e",
		Name:         "Expedite",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"haste"},
				Label:    "Expedite — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
