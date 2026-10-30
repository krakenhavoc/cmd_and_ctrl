package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mshB4ControlsAnotherHero reports whether `controller` controls a Hero other
// than `self` ("if you control another Hero"). It reads effective subtypes, so
// a changeling counts. `hero` is the word as the source reads it
// (source.TextSubtype("Hero")), so a text-changing effect reaches it.
func mshB4ControlsAnotherHero(g *game.Game, controller, self uuid.UUID, hero string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.InstanceID != self && c.HasSubtype(hero) {
			return true
		}
	}
	return false
}

// mshB4CastSpellTargetingYourCreature is "whenever you cast a spell that
// targets a creature you control" (Colleen Wing, Iron Fist): the cast is by
// the source's controller, and at least one of the spell's targets is a
// creature on the battlefield under that player's control.
func mshB4CastSpellTargetingYourCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil || item.Kind != game.StackItemSpell {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() && c.Controller == source.Controller && onBattlefield(g, t.ID) {
			return true
		}
	}
	return false
}
