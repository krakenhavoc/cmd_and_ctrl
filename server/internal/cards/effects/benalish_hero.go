package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benalish Hero — Creature — Human Soldier {W}, 1/1 (Limited Edition
// Alpha):
//
//	"Banding"
//
// Banding is the engine's (game/banding.go, ADR 0224): the attack
// declaration's bands, a blocker of one member blocking the band, and
// the damage division handed to its controller.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "4c81cfb7-8765-4e28-ae33-4287fa9a86cc",
		Name:            "Benalish Hero",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordBanding},
	})
}
