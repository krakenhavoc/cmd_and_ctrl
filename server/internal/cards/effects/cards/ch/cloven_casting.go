package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloven Casting — Enchantment {5}{U}{R} (Alara Reborn):
//
//	"Whenever you cast a multicolored instant or sorcery spell, you may pay
//	 {1}. If you do, copy that spell. You may choose new targets for the
//	 copy."
//
// Reflective Golem's copy-after-pay with a different cast condition: the
// spell you cast is an instant or sorcery with two or more colours (read
// through the effective colours of the spell on the stack, CR 105.4). The
// {1} is the casting player's choice as the ability resolves; the copy is of
// the spell as last known if it has already resolved (CR 707.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b08e5f7-c771-4e03-8560-15de6a8bec71",
		Name:         "Cloven Casting",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && (c.IsInstant() || c.IsSorcery()) && len(c.EffectiveColors()) >= 2
			}, "Cloven Casting — you may pay {1} to copy that spell",
				effects.S82OldB455MayPayToCopyThatSpell("{1}", "Cloven Casting — pay {1} to copy that spell?")),
		},
	})
}
