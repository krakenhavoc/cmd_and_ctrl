package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shu Elite Companions — Creature — Human Soldier {4}{W}, 3/3
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
		OracleID:        "edf304b8-523e-4233-8d77-1d45d33cbb86",
		Name:            "Shu Elite Companions",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
	})
}
