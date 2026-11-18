package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_rpr_b01_helpers.go — shared pieces for S82 reprint pass, batch 1 (#5908).

// anotherNontokenCreatureEnteredUnderYourControl is "whenever another
// nontoken creature you control enters" (Anafenza, Kin-Tree Spirit).
func anotherNontokenCreatureEnteredUnderYourControl(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !AnotherCreatureEnteredUnderYourControl(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// creatureWithPower2OrLessEnteredUnderYourControl is "whenever a creature
// you control with power 2 or less enters", the source itself included
// (Ezuri, Claw of Progress). The power is the entering creature's as it
// enters.
func creatureWithPower2OrLessEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && c.CurrentPower() <= 2
}

// condPlusOneCounterPutOnYourPermanent is "if a +1/+1 counter was put on a
// permanent under your control this turn" as an intervening-if Condition
// (Fairgrounds Trumpeter), read for the item's controller as it resolves.
type condPlusOneCounterPutOnYourPermanent struct{}

func (condPlusOneCounterPutOnYourPermanent) Holds(ctx *Context) bool {
	return ctx.Game.PlusOneCounterPutOnYourPermanentThisTurnForEffect(ctx.Controller())
}

func (condPlusOneCounterPutOnYourPermanent) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondPlusOneCounterPutOnYourPermanent, 0), true
}

// condDamagedByGiantWizardOrSpell is "if a Giant, Wizard, or spell you
// controlled dealt damage to it this turn", "it" being the permanent the
// trigger's damage event names (Aegar, the Freezing Flame).
type condDamagedByGiantWizardOrSpell struct{}

func (condDamagedByGiantWizardOrSpell) Holds(ctx *Context) bool {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return false
	}
	return ctx.Game.PermanentDealtDamageThisTurnByKindForEffect(ctx.Item.Trigger.Event.Target, ctx.Controller(), true, "Giant", "Wizard")
}

func (condDamagedByGiantWizardOrSpell) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondDamagedByGiantWizardOrSpell, 0), true
}
