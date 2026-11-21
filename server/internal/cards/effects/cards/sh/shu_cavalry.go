package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shu Cavalry — Creature — Human Soldier {2}{W}, 2/2
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
		OracleID:        "d9180f63-c745-41a1-9f01-273e7ee13176",
		Name:            "Shu Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
	})
}
