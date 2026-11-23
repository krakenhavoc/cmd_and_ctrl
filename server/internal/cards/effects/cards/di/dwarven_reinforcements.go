package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Reinforcements — Sorcery {3}{R} (Kaldheim, #5505):
//
//	"Create two 2/1 red Dwarf Berserker creature tokens.
//	 Foretell {1}{R}"
//
// Two of Fearless Liberator's tokens, plus the shared foretell special
// action (CR 702.143).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "b38c567e-fd79-4a74-ab5f-df4a6a8f7e00",
		Name:           "Dwarven Reinforcements",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Tokens: 2, CreatureTokens: 2},
		SpecialActions: []game.SpecialAction{effects.Foretell("{1}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("2/1 red Dwarf Berserker"), N: 2}.Apply(ctx)
		},
	})
}
