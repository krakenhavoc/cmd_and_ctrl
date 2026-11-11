package effects

// Become the Avalanche — Sorcery {4}{G}{G} (Foundations Commander):
//
//	"Draw a card for each creature you control with power 4 or greater. Then
//	 creatures you control get +X/+X until end of turn, where X is the number
//	 of cards in your hand."
//
// Both counts are read as the spell resolves, in printed order: the creatures
// with power 4 or greater are counted before anything is drawn, then the draw
// settles (a replacement's prompt is waited out, CR 608.2c), and X is the hand
// size after the draw. The set of creatures that get the bonus is fixed then
// (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5fcbbaf-9431-4fa8-ac11-12ace1f2054c",
		Name:         "Become the Avalanche",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller == ctx.Controller() && c.IsCreature() && b31CurrentPowerOf(ctx.Game, c.InstanceID) >= 4 {
						n++
					}
				}
				if n <= 0 {
					return nil
				}
				return DrawCards{N: n}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				x := 0
				if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Hand != nil {
					x = p.Hand.Size()
				}
				if x <= 0 {
					return nil
				}
				return BoostUntilEOT{
					Match:     And(Creature(), YouControl()),
					Power:     x,
					Toughness: x,
					Label:     "Become the Avalanche — +X/+X",
				}.Apply(ctx)
			}),
		),
	})
}
