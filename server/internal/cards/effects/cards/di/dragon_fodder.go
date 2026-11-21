package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon Fodder — Sorcery {1}{R}:
//
//	"Create two 1/1 red Goblin creature tokens."
//
// Two bodies for two mana with no rider — the cheapest way an
// aristocrats deck turns a card into sacrifice fuel, and the reason
// it shows up in every Goblin and every token list.
//
// No simplification: the tokens are Krenko's, so they count for
// Krenko, Mob Boss's own ability and for anything else that counts
// Goblins.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0d2c45b-b6e3-4999-bdab-976e8f0d6617",
		Name:         "Dragon Fodder",
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{
				Controller: ctx.Controller(),
				Template:   effects.RedGoblinToken(),
				N:          2,
			}.Apply(ctx)
		},
	})
}
