package effects

// Shamanic Revelation — Sorcery {3}{G}{G} (EDHREC rank 368):
//
//	"Draw a card for each creature you control.
//	 Ferocious — You gain 4 life for each creature you control with
//	 power 4 or greater."
//
// The go-wide deck's refill. Both counts are taken once, as the
// spell resolves, over the creatures the caster controls: the draw
// is per creature, the lifegain per creature with power 4 or
// greater (CurrentPower — effective power plus counters, so a
// pumped or counter-laden 3/3 counts). Ferocious's "if you control a
// creature with power 4 or greater" gate is implied by a count of
// zero gaining zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1d171de-1c6d-4fb9-817a-9c689c709f3d",
		Name:         "Shamanic Revelation",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The life gain waits for a prompt the draw asks (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				creatures, _ := shamanicRevelationCounts(ctx)
				return DrawCards{N: creatures}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				_, big := shamanicRevelationCounts(ctx)
				return GainLife{Amount: 4 * big}.Apply(ctx)
			}),
		),
	})
}

// shamanicRevelationCounts is the creatures you control and how many of
// them have power 4 or greater. Each of the spell's steps reads it from the
// game, since a piece cannot hand it to the next (#4155).
func shamanicRevelationCounts(ctx *Context) (creatures, big int) {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != ctx.Controller() || !c.IsCreature() {
			continue
		}
		creatures++
		if c.CurrentPower() >= 4 {
			big++
		}
	}
	return creatures, big
}
