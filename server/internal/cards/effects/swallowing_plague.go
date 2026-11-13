package effects

// Swallowing Plague — Sorcery — Arcane {X}{B}{B} (Champions of Kamigawa,
// #5603):
//
//	"Swallowing Plague deals X damage to target creature and you gain X
//	 life."
//
// Tendrils of Corruption's two steps with X as paid (CR 107.3a): the
// damage, then the life, both X whatever the creature's toughness. With
// its only target illegal the spell does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91cac048-90ca-41bc-9177-56c78453c4a2",
		Name:         "Swallowing Plague",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if len(ctx.Item.Targets) == 0 || ctx.X() <= 0 {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: ctx.Item.Targets[0].ID, Amount: ctx.X()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error { return GainLife{Amount: ctx.X()}.Apply(ctx) }),
		),
	})
}
