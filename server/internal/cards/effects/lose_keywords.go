package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lose_keywords.go — the battlefield static "<permanents> lose
// <keywords>" (Gravity Sphere's "All creatures lose flying", Mystic
// Decree's "lose flying and islandwalk").
//
// An ordinary layer-6 removal (CR 613.1f), and deliberately NOT
// LoseAndCantHave: a card that prints only "lose" is ordered against
// every grant by CR 613.7 timestamp, so a flying grant with a later
// timestamp than the static gives the keyword back. Only a card that
// prints "can't have" strips after the whole layer-6 bucket
// (cant_have.go).
//
// Append-only: add a helper below the last one.

// LoseKeywords is the static "<permanents> lose <keywords>". `applies`
// is the ordinary StaticAbility predicate, read live on every pass (a
// static locks no set, CR 611.3a). `keywords` are canonical lowercase
// tokens (game/keywords.go).
func LoseKeywords(applies func(target *game.Card, g *game.Game, source *game.Card) bool, keywords ...string) game.StaticAbility {
	lost := append([]string(nil), keywords...)
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			kept := c.Abilities[:0:0]
			for _, a := range c.Abilities {
				if !keywordIn(a, lost) {
					kept = append(kept, a)
				}
			}
			c.Abilities = kept
		},
	}
}

// keywordIn reports whether the ability token `a` is one of `keywords`.
func keywordIn(a string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.EqualFold(a, kw) {
			return true
		}
	}
	return false
}

// AllCreatures is the static scope "all creatures" — every creature on
// the battlefield, every player's.
func AllCreatures(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsCreature()
}
