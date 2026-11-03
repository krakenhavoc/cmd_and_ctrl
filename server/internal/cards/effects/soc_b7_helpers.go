package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b7_helpers.go — shared readers for Secrets of Strixhaven
// Commander, batch 7 (#4834).

// socB7ControlsADecayedCreature reports whether `player` controls a
// creature with decayed, read through the layered view (Jadar).
func socB7ControlsADecayedCreature(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && game.HasKeyword(&c, game.KeywordDecayed) {
			return true
		}
	}
	return false
}

// socB7SourceHasPlusOneCounter reports whether the card is on the
// battlefield with one or more +1/+1 counters (Ingenious Prodigy).
func socB7SourceHasPlusOneCounter(g *game.Game, id uuid.UUID) bool {
	if !onBattlefield(g, id) {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	return ok && c.Counters[game.CounterPlusOne] > 0
}

// socB7CreaturePermanent is "a creature" as a spell's target: a creature
// on the battlefield, not a creature card in another zone (Killian, Ink
// Duelist).
func socB7CreaturePermanent() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && onBattlefield(g, c.InstanceID)
	}
}

// socB7EnchantedByAuraOf reports whether the battlefield card `host` is
// enchanted by an Aura `player` controls (Killian, Decisive Mentor).
func socB7EnchantedByAuraOf(g *game.Game, host, player uuid.UUID) bool {
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.Controller == player && a.IsAura() && a.IsAttachedTo(host) {
			return true
		}
	}
	return false
}
