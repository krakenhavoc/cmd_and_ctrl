package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repentant Blacksmith — Creature — Human {1}{W}, 1/2 (Arabian Nights,
// #5641):
//
//	"Protection from red"
//
// Protection from a colour is printed data on PrintedKeywords, enforced
// since #662 / ADR 0072: red sources can't damage, enchant, block or
// target it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a83073a2-e63d-4105-8bad-9612e411fc85",
		Name:            "Repentant Blacksmith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
	})
}
