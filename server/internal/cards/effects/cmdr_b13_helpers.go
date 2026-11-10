package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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

// cmdrB13CreatureFromGraveyard is "a creature you control entered from a
// graveyard or was cast from a graveyard" for one entering creature
// (Celes, Rune Knight's reading, the source itself included).
func cmdrB13CreatureFromGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	return ev.EnteredFrom == game.ZoneGraveyard || c.Provenance.FromZone == game.ZoneGraveyard
}
