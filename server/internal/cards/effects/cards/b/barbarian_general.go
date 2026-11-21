package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barbarian General — Creature — Human Barbarian Soldier {4}{R}, 3/2
// (Portal Three Kingdoms, #5622):
//
//	"Horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d4f2e305-bec0-48c8-8040-f85618c71cc9",
		Name:            "Barbarian General",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
	})
}
