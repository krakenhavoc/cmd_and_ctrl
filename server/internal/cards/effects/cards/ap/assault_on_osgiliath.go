package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assault on Osgiliath — Sorcery {X}{R}{R}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Amass Orcs X, then Goblins and Orcs you control gain double strike and
//	 haste until end of turn."
//
// The amass reads the X announced with the spell (CR 107.3), then the grant
// fixes its set of Goblins and Orcs as it resolves (CR 611.2c), so the Army
// just amassed, an Orc now, is in it. Both keywords are one effect. X of 0
// still amasses 0: it makes a 0/0 Army, which dies at once (CR 704.5f), if you
// had none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2aa50d2-d22d-495e-afc9-23a51c8c8f32",
		Name:         "Assault on Osgiliath",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.Amass{Subtype: "Orc", N: ctx.X()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.YouControl(), effects.Or(effects.HasSubtype("Goblin"), effects.HasSubtype("Orc"))),
					Keywords: []string{"double strike", "haste"},
					Label:    "Assault on Osgiliath — double strike and haste",
				}.Apply(ctx)
			}),
		),
	})
}
