package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Night Revelers — Creature — Vampire {4}{R}, 4/4 (Innistrad, S82 OLD
// batch 450, #5572):
//
//	"This creature has haste as long as an opponent controls a Human."
//
// Syr Ginger's conditional keyword grant over Innocent Traveler's
// "opponent controls a Human" test.
//
// No simplification.
func init() {
	// "Human" is a word of the card's text; OpponentControlsAtLeast lends it
	// (CR 612.2, #3984).
	opponentHasHuman := effects.OpponentControlsAtLeast(1, func(c game.Card) bool { return c.HasSubtypeWord("Human") })
	grant := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) && opponentHasHuman(g, source.Controller, source.InstanceID)
	}, "haste")
	grant.Label = "This creature has haste as long as an opponent controls a Human."
	effects.Register(effects.Spec{
		OracleID:     "56697773-8a04-49b6-8327-895e972da7e7",
		Name:         "Night Revelers",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
