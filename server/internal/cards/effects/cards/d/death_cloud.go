package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Cloud — Sorcery {X}{B}{B}{B} (The List):
//
//	"Each player loses X life, discards X cards, sacrifices X creatures
//	 of their choice, then sacrifices X lands of their choice."
//
// Smallpox with X for every count: four instructions done to every player,
// one after the other (CR 608.2c). A player with fewer cards or permanents
// than X loses, discards or sacrifices what they have. The continuations
// are registered (s82_bns_b08_helpers.go), so a table waiting on any
// seat's choice is a restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "680f5eca-b98e-4f73-a852-ab629d94297e",
		Name:         "Death Cloud",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		OnResolve:    effects.DeathCloudResolve,
	})
}
