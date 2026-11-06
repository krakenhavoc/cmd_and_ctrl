package effects

// The Art of Tea — Instant — Lesson {1}{G} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Put a +1/+1 counter on up to one target creature you control.
//	 Create a Food token."
//
// The counter step (putPlusOneCounterOnEachLegalTarget) over "up to one", then the Food. With
// no target chosen the spell still resolves and makes the Food; with its
// one target gone illegal it does not resolve at all (CR 608.2b), so no
// Food, as the ruling says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ed4e48f-91bd-45f9-9f31-a2705b330713",
		Name:         "The Art of Tea",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return putPlusOneCounterOnEachLegalTarget(ctx.Game, ctx.Item)
			}),
			CreateToken{Template: FoodToken(), N: 1},
		),
	})
}
