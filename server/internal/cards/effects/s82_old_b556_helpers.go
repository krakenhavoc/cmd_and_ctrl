package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB556AGiantSpellWasCast is "whenever a player casts a Giant spell"
// (Elvish Handservant): any player's cast of a spell with the creature type,
// read through the source's text (CR 612.2) and the spell's own types, a
// changeling's included.
func S82OldB556AGiantSpellWasCast(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.HasSubtype(source.TextSubtype("Giant"))
}

// S82OldB556AGoblinYouControlWasPutIntoAGraveyard is "whenever a Goblin you
// control is put into a graveyard from the battlefield" (Knucklebone Witch): the
// permanent, token or not, creature or not, was a Goblin and under the source's
// controller's control as it last existed on the battlefield (CR 603.10a). The
// Witch itself counts.
func S82OldB556AGoblinYouControlWasPutIntoAGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && LeftAsSubtype(ev, c, source.TextSubtype("Goblin")) && LeftUnderControlOf(ev, c) == source.Controller
}

// S82OldB556AGoblinYouControlDealtCombatDamageToAPlayer is "whenever a Goblin you
// control deals combat damage to a player" (Boggart Mob): one trigger for each
// Goblin that dealt it, the Mob included, read through the source's text.
func S82OldB556AGoblinYouControlDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !CombatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.HasSubtype(source.TextSubtype("Goblin"))
}
