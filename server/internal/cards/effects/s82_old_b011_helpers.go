package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b011_helpers.go — shared pieces of the Secret Lair Drop batch 11
// cards (#5462).

// OldB11ManaSpentOnCastTrigger is "the amount of mana spent to cast it" for a
// "When you cast this spell" trigger: the spell is still on the stack when
// the trigger resolves, and its payment record is the spell's (the trigger
// item's own record is empty, CR 400.7d does not apply yet).
func OldB11ManaSpentOnCastTrigger(ctx *Context) int {
	if ctx.Item.Trigger != nil {
		for _, it := range ctx.Game.StackMeta {
			if it != nil && it.Kind == game.StackItemSpell && it.SourceCardID == ctx.Item.Trigger.Event.CardID {
				return NewContext(ctx.Game, it).ManaSpent().Total()
			}
		}
	}
	return ctx.ManaSpentToCastThis().Total()
}

// OldB11OpponentCreatureDied is "a creature an opponent controls dies": the
// permanent left the battlefield for a graveyard as a creature, under the
// control of a player other than the source's controller (last-known
// information, CR 603.10a), a token included.
func OldB11OpponentCreatureDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
		return false
	}
	c, _ := g.LookupCardForEffect(ev.CardID)
	if !LeftAsType(ev, c, "creature") {
		return false
	}
	controller := LeftUnderControlOf(ev, c)
	return controller != uuid.Nil && controller != source.Controller
}

// OldB11GreatestPowerAmongOpponentsCreatures is "creature with the greatest
// power among creatures your opponents control": an opponent's creature whose
// current power is at least that of every other creature the caster's
// opponents control (ties are all legal).
func OldB11GreatestPowerAmongOpponentsCreatures(g *game.Game, caster uuid.UUID, c game.Card) bool {
	if c.Controller == caster || !c.IsCreature() {
		return false
	}
	for _, o := range g.BattlefieldCardsForEffect() {
		if o.Controller != caster && o.IsCreature() && o.CurrentPower() > c.CurrentPower() {
			return false
		}
	}
	return true
}
