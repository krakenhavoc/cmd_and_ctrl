package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func DealCombatDamage(g *game.Game, owner, source, victim uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: owner, Source: source, Target: victim, Amount: 2, Combat: true})
	})
}

func DealNoncombatDamage(g *game.Game, owner, source, victim uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: owner, Source: source, Target: victim, Amount: 2})
	})
}

func EffectiveHasKeyword(c *game.Card, kw string) bool {
	for _, a := range c.Effective().Abilities {
		if a == kw {
			return true
		}
	}
	return false
}
