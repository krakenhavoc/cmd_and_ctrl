package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// caster_loses_life.go — "whenever a player casts <a spell>, that player
// loses N life": Mai, Scornful Striker and Soot Imp. Kambal's drain
// without the opponent filter and without the life gain.
//
// The loss is life loss, not damage (a negative life change), so
// nothing that prevents or redirects damage touches it. It happens on
// the CAST, so a spell that is later countered has already cost its
// caster the life. The caster is read at resolution off the item's
// carried trigger context (item.Trigger.Event.Actor, #1223).
//
// Append-only.

// castersLoseLife is a triggered ability that fires when any player
// (the controller included) casts a spell `spellOK` accepts and makes
// that player lose `n` life.
func castersLoseLife(label string, n int, spellOK func(spell game.Card) bool) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventCast},
		AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			spell, ok := g.LookupCardForEffect(ev.CardID)
			return ok && spellOK(spell)
		},
		Key: label,
		Effect: func(g *game.Game, item *game.StackItem) error {
			caster := item.Trigger.Event.Actor
			if p := g.PlayerByIDForEffect(caster); p == nil || p.Eliminated {
				return nil
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, caster, -n)
		},
	}
}
