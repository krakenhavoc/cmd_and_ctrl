package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wu Light Cavalry — Creature — Human Soldier {1}{U}, 1/2
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
		OracleID:        "c24da532-311b-4b83-82d2-7dc80409ba12",
		Name:            "Wu Light Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
	})
}
