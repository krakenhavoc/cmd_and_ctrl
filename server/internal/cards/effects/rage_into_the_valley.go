package effects

// Rage into the Valley — Sorcery {2}{B} (The Hobbit, HOB):
//
//	"You draw a card and lose 1 life.
//	 Amass Goblins 2."
//
// Profane Insight's draw and life loss (life loss, not damage, CR 119.3),
// then the amass; each waits for any prompt the one before asked (a dredge
// offer; CR 608.2c, #4155).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c7641b6-3a6c-4441-a98c-3b5d61d91f76",
		Name:         "Rage into the Valley",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			DrawCards{N: 1},
			Instruction(func(ctx *Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Item.Controller, -1)
			}),
			Amass{Subtype: "Goblin", N: 2},
		),
	})
}
