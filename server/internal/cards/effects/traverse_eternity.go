package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Traverse Eternity — Sorcery {2}{U}{U} (Doctor Who):
//
//	"Draw cards equal to the greatest mana value among historic permanents you
//	 control. (Artifacts, legendaries, and Sagas are historic.)"
//
// The greatest mana value is read as the spell resolves, over the historic
// permanents its controller controls (b09IsHistoric: an artifact, a legendary
// or a Saga, by the layered view). With none, or all of mana value 0, nothing is
// drawn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac5c2975-6a56-42c7-97c5-6a0e75c000bb",
		Name:         "Traverse Eternity",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			most := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && b09IsHistoric(c) {
					most = max(most, c.ManaValue())
				}
			}
			if most <= 0 {
				return nil
			}
			return DrawCards{N: most}.Apply(ctx)
		},
	})
}
