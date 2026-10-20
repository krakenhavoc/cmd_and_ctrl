package effects

// Lazotep Plating — Instant {1}{U}:
//
//	"Amass Zombies 1.
//	 You and permanents you control gain hexproof until end of turn."
//
// The amass runs first, so the Army it makes or grows is a permanent
// you control when the hexproof grant is applied and is covered by it.
// The grant is Dawn's Truce's: a player keyword for you and a
// keyword grant for the permanents you control, both until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ba0082fb-2d4c-489e-8140-93a6fa693fd0",
		Name:         "Lazotep Plating",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The hexproof waits for a prompt the amass asks (CR 608.2c, #4155).
			Amass{Subtype: "Zombie", N: 1},
			Instruction(func(ctx *Context) error {
				if err := (GainPlayerKeyword{
					Player:   ctx.Controller(),
					Keyword:  KeywordHexproof,
					Label:    "Lazotep Plating — hexproof",
					Duration: DurationUntilEndOfTurn(ctx),
				}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Match:    YouControl(),
					Keywords: []string{"hexproof"},
					Label:    "Lazotep Plating — permanents you control gain hexproof",
				}.Apply(ctx)
			}),
		),
	})
}
