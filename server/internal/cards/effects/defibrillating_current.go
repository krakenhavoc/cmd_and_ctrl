package effects

// Defibrillating Current — Sorcery {2/R}{2/W}{2/B} (Tarkir: Dragonstorm,
// #4463):
//
//	"Defibrillating Current deals 4 damage to target creature or
//	 planeswalker and you gain 2 life."
//
// Ribbons of Night's two steps on Obliterating Bolt's target: the damage,
// then the life. Each {2/C} is paid with that colour or two mana of any
// kind by the casting path (CR 107.4e; Beseech the Queen), and the mana
// value is 6 (CR 202.3f). A target that is gone as the spell resolves
// fizzles the whole spell, life included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a96b93b-bae6-48fb-87f5-05f3ffcf7ba9",
		Name:         "Defibrillating Current",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, ref := range ctx.LegalTargets() {
						if err := (DealDamage{Source: ctx.Item.SourceCardID, Target: ref.ID, Amount: 4}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
			Instruction(func(ctx *Context) error {
				return GainLife{Player: ctx.Item.Controller, Amount: 2}.Apply(ctx)
			}),
		),
	})
}
