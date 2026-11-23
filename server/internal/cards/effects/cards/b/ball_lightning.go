package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ball Lightning — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "7485cf25-eb41-4397-be13-7f0b8c10c70a", Name: "Ball Lightning", Completeness: effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{effects.On(game.EventBeginEndStep, effects.AnyPlayer, "Ball Lightning — sacrifice this creature", func(g *game.Game, item *game.StackItem) error {
			if !effects.B09SourceStillOnBattlefield(g, item) {
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
