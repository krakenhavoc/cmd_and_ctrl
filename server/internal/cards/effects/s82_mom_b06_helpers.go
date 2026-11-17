package effects

// drawXCards is "draw X cards": the controller draws the X announced for
// the spell (CR 107.3); X = 0 draws nothing. Named once so a draw-X spell
// calls it instead of copying the closure (TestNoNewExactClonesInTheCatalog).
func drawXCards(ctx *Context) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	return ctx.Game.DrawNForEffect(ctx.Controller(), x)
}
