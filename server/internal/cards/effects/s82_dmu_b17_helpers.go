package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dmuB17SoleTargetPermanent is the one permanent a spell on the stack targets
// only (CR 115.9c: every instance of the word "target" chose the same object,
// so a spell that names one creature twice qualifies). It is false for a spell
// with no target, with two different targets, or whose target is a player or
// has left the battlefield (CR 115.9b).
func dmuB17SoleTargetPermanent(g *game.Game, spellID uuid.UUID) (game.Card, bool) {
	t, ok := g.StackItemSoleTargetForEffect(spellID)
	if !ok || t.Kind != game.TargetCard || !g.Battlefield.Contains(t.ID) {
		return game.Card{}, false
	}
	return g.LookupCardForEffect(t.ID)
}

// dmuB17CastSpellTargetsOnly is the condition shared by Ivy, Gleeful Spellthief
// and Vesuvan Duplimancy: a spell was cast (by `byYou` only, or by any player)
// that targets only a single permanent passing `match`.
func dmuB17CastSpellTargetsOnly(byYou bool, match func(target game.Card, source *game.Card) bool) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCast || ev.CardID == uuid.Nil || source == nil {
			return false
		}
		if byYou && ev.Actor != source.Controller {
			return false
		}
		target, ok := dmuB17SoleTargetPermanent(g, ev.CardID)
		return ok && match(target, source)
	}
}
