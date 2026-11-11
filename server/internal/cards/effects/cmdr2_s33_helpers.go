package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2S33YourCommanderAttacked is "whenever your commander attacks": a
// commander its controller owns is declared as an attacker, by that
// controller. An attack on a planeswalker or battle counts, since the
// clause names no player (CR 508.1m, CR 508.3).
func cmdr2S33YourCommanderAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCommander && c.Owner == source.Controller
}

// cmdr2S33CondThisPowerAtLeast is an intervening "if its power is N or
// greater" about the ability's own source creature (CR 603.4): its
// current power while it is on the battlefield, its last-known power once
// it has left (CR 608.2h). Angelic Sell-Sword's attack trigger.
type cmdr2S33CondThisPowerAtLeast struct{ N int }

func (c cmdr2S33CondThisPowerAtLeast) Holds(ctx *Context) bool {
	info, ok := ctx.SourcePermanent()
	return ok && info.Power >= c.N
}

func (c cmdr2S33CondThisPowerAtLeast) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondSource, c.N), c.N > 0
}

// cmdr2S33ThisOrAnotherNontokenCreatureYouControlEntered is "whenever this
// creature or another nontoken creature you control enters": the source
// itself always counts, token or not; any other creature must be a
// nontoken one under the source's controller.
func cmdr2S33ThisOrAnotherNontokenCreatureYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() {
		return false
	}
	return ev.CardID == source.InstanceID || !IsToken(c)
}
