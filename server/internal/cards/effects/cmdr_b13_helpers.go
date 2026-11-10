package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

// cmdrB13SpellTargetsPermanentOrPlayer reports whether the spell on the
// stack has a target that is a player or a permanent (Shiko and Narset,
// Unified). The cast path announces targets before the cast event
// (CR 601.2c), so the item already carries them. A card target counts only
// while it is on the battlefield, which at the cast is what makes it a
// permanent rather than a card in a graveyard or a spell.
func cmdrB13SpellTargetsPermanentOrPlayer(g *game.Game, spellID uuid.UUID) bool {
	item := g.StackItemForEffect(spellID)
	if item == nil {
		return false
	}
	for _, t := range item.Targets {
		switch t.Kind {
		case game.TargetPlayer, game.TargetSelf:
			return true
		case game.TargetCard:
			if z := g.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneBattlefield {
				return true
			}
		}
	}
	return false
}
