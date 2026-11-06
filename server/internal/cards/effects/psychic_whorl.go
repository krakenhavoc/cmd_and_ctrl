package effects

// Psychic Whorl — Sorcery {2}{B} (Bloomburrow, #4467):
//
//	"Target opponent discards two cards. Then if you control a Rat,
//	 surveil 2. (Look at the top two cards of your library, then put any
//	 number of them into your graveyard and the rest on top of your
//	 library in any order.)"
//
// The discard is the opponent's own choice (CR 701.9a), queued as Mind
// Rot's is. "Then" is a second step (CR 608.2c): it waits for the discard
// to be answered, and the Rat is looked for then, on the battlefield
// through the layers and through the spell's own text (CR 612.2). The
// surveil (CR 701.25a) is the caster's prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3df4eb80-e8dd-4a9e-bd8a-b4b23fb9de6c",
		Name:         "Psychic Whorl",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return targetPlayerDiscards(2)(ctx.Game, ctx.Item) }),
			Instruction(psychicWhorlSurveil),
		),
	})
}

// psychicWhorlSurveil is "if you control a Rat, surveil 2."
func psychicWhorlSurveil(ctx *Context) error {
	if !ControlsA(ctx.TextSubtype("Rat"))(ctx.Game, ctx.Controller()) {
		return nil
	}
	return Surveil{N: 2}.Apply(ctx)
}
