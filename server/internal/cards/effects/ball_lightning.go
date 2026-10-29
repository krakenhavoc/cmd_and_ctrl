package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ball Lightning — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "7485cf25-eb41-4397-be13-7f0b8c10c70a", Name: "Ball Lightning", Completeness: CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{On(game.EventBeginEndStep, AnyPlayer, "Ball Lightning — sacrifice this creature", func(g *game.Game, item *game.StackItem) error {
			if !b09SourceStillOnBattlefield(g, item) {
				return nil
			}
			// The trigger's controller cannot sacrifice a permanent
			// another player gained control of in response.
			c, ok := g.LookupCardForEffect(item.SourceCardID)
			if !ok || c.Controller != item.Controller {
				return nil
			}
			return g.SacrificePermanentForEffect(item.SourceCardID)
		})},
	})
}
