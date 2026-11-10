package effects

// cmdrB13CastSpellPower is "that spell's power" for a trigger that watched
// a spell being cast: the power it has where it is now (on the stack, or in
// the graveyard if it was countered).
func cmdrB13CastSpellPower(ctx *Context) int {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return 0
	}
	spell, ok := ctx.Game.LookupCardForEffect(ctx.Item.Trigger.Event.CardID)
	if !ok {
		return 0
	}
	return ctx.Game.PowerAnywhereForEffect(spell)
}
