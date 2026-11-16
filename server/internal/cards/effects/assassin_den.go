package effects

// Assassin Den — Creature — Wall {3}{U}, 2/4 (Assassin's Creed, #5466):
//
//	"Defender (This creature can't attack.)
//	 {3}{U}: Put a +1/+1 counter on target creature you control. It
//	 can't be blocked this turn. Activate only as a sorcery."
//
// Two steps on one target: the counter, then the restriction, so the
// second waits for any prompt the counter asks (a CR 616 order).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "826274fd-9831-4122-a5d9-0d1caa3d3a85",
		Name:            "Assassin Den",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:        "{3}{U}: Put a +1/+1 counter on target creature you control. It can't be blocked this turn. Activate only as a sorcery.",
			Purpose:      ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:         ManaCost("{3}{U}"),
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Effect: Do(
				Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
				Instruction(func(ctx *Context) error {
					return targetCantBeBlockedThisTurn("Assassin Den")(ctx.Game, ctx.Item)
				}),
			),
		}},
	})
}
