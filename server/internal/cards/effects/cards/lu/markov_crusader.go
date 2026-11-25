package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Markov Crusader — Creature — Vampire Knight {4}{B}, 4/3 (Eldritch Moon,
// S82 OLD, #5541):
//
//	"Lifelink
//	 This creature has haste as long as you control another Vampire."
//
// A layer 6 grant to itself while its controller controls another creature
// with the Vampire subtype, read through the layered view (a changeling
// counts) on every recompute.
//
// No simplification.
func init() {
	haste := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID &&
			effects.CountControlled(g, source.Controller, func(c game.Card) bool {
				return c.InstanceID != source.InstanceID && c.HasSubtype(source.TextSubtype("Vampire"))
			}) > 0
	}, "haste")
	haste.Label = "This creature has haste as long as you control another Vampire."
	effects.Register(effects.Spec{
		OracleID:        "0a8511b9-ed7d-4e33-9c2f-96b9a4c535e2",
		Name:            "Markov Crusader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static:          []game.StaticAbility{haste},
	})
}
