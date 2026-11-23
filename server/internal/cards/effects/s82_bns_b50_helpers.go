package effects

// bnsB50ImpulseTopUntilYourNextTurnEnds is "exile the top card of your
// library. You may play that card until the end of your next turn": the
// seat-turn window of Alania's Pathmaker (ADR 0063), for a trigger that
// prints the same clause.
func BnsB50ImpulseTopUntilYourNextTurnEnds(ctx *Context) error {
	window := ctx.Game.UntilEndOfYourNextTurnDuration(ctx.Controller())
	return ExileTopWithPermission{
		From:    ctx.Controller(),
		GrantTo: ctx.Controller(),
		N:       1,
		Window:  &window,
	}.Apply(ctx)
}
