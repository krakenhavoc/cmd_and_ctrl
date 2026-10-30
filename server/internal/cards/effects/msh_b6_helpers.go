package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// msh_b6_helpers.go — shared pieces for Marvel Super Heroes, half B, sixth
// batch (#4455). Every name carries the mshB6 prefix because the other
// Marvel batches land beside this one.

// mshB6CastTargetingYourCreature is the When of "whenever you cast a spell
// that targets a creature you control": the cast event is yours and the
// spell on the stack has a card target that is a creature on the
// battlefield under your control.
func mshB6CastTargetingYourCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
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
