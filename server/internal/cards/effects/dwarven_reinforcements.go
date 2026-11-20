package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "b38c567e-fd79-4a74-ab5f-df4a6a8f7e00",
		Name:           "Dwarven Reinforcements",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Tokens: 2, CreatureTokens: 2},
		SpecialActions: []game.SpecialAction{Foretell("{1}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("2/1 red Dwarf Berserker"), N: 2}.Apply(ctx)
		},
	})
}
