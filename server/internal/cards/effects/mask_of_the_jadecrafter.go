package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mask of the Jadecrafter — Artifact {2} (The Brothers' War, #5656):
//
//	"{X}, {T}, Sacrifice this artifact: Create an X/X colorless Golem
//	 artifact creature token. Activate only as a sorcery.
//	 Unearth {2}{G}"
//
// X is announced with the cost (CR 602.2b) and read back with ctx.X(), the
// Treasure Vault way. The token card is */*, so the size is set as the token
// is created (s82BroB21XGolem). Unearth is the keyword helper: the Mask
// returns from the graveyard, and the same ability can be activated again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1870fff-b87e-434e-8d4c-d9e60f7f4284",
		Name:         "Mask of the Jadecrafter",
		XMatters:     true,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Activated: []ActivatedAbility{
			{
				Label:        "{X}, {T}, Sacrifice this artifact: Create an X/X colorless Golem artifact creature token. Activate only as a sorcery.",
				Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
				Cost:         Plus(ManaCost("{X}"), TapCost(), SacrificeThis()),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return s82BroB21XGolem(ctx, ctx.X(), false)
				},
			},
			Unearth("{2}{G}"),
		},
	})
}
