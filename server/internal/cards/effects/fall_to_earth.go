package effects

// Fall to Earth — Instant {4}{W} (Marvel Super Heroes Commander, #4455):
//
//	"Exile target creature. Each player gains 3 life.
//	 Basic landcycling {2}"
//
// The life comes after the exile (CR 608.2c), in APNAP order. With its
// only target illegal the spell does not resolve (CR 608.2b), so nobody
// gains life. Basic landcycling is the cycling family's BasicLandcycling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "223db021-979f-4312-8cb5-169ae4bf17bc",
		Name:         "Fall to Earth",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(exileTheFirstTargetIfLegal),
			Instruction(func(ctx *Context) error {
				for _, player := range apnapPlayers(ctx.Game) {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, 3); err != nil {
						return err
					}
				}
				return nil
			}),
		),
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
