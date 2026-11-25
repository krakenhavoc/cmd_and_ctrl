package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Gathering — Sorcery {2}{R} (Ravnica Allegiance):
//
//	"Create a number of 1/1 red Goblin creature tokens equal to two plus the
//	 number of cards named Goblin Gathering in your graveyard."
//
// Counted as the spell resolves; it is on the stack, not in the graveyard, so it
// never counts itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01f06642-015f-40b4-a940-af576ad9e82c",
		Name:         "Goblin Gathering",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 2
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Graveyard != nil {
				for _, c := range p.Graveyard.Cards {
					if c.Name == "Goblin Gathering" {
						n++
					}
				}
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin"), N: n}.Apply(ctx)
		},
	})
}
