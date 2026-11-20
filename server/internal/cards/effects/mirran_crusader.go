package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirran Crusader — Creature — Human Knight {1}{W}{W}, 2/2 (Mirrodin Besieged,
// #5576):
//
//	"Double strike, protection from black and from green"
//
// Double strike and two protections, written as Akroma, Angel of Fury's
// are: protection from black and protection from green are separate
// qualities (CR 702.16a), each refusing that colour's targets, blocks,
// damage and attachments.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fe69d9bd-2a60-4b33-a0d8-1ca18c2b6705",
		Name:            "Mirran Crusader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", game.ProtectionFromColor("B"), game.ProtectionFromColor("G")},
	})
}
