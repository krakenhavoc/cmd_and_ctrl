package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hazardous Blast — Sorcery {3}{R}:
//
//	"Hazardous Blast deals 1 damage to each creature your opponents
//	 control. Creatures your opponents control can't block this turn."
//
// Cosmotronic Wave's text under another name. The damage is one-shot
// and the "can't block" is read live until cleanup (#1650); see
// cosmotronic_wave.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffd666c7-a8ca-4467-8733-878efb127268",
		Name:         "Hazardous Blast",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		Completeness: CompletenessFull,
		OnResolve:    pingThenOpponentsCantBlock("Hazardous Blast"),
	})
}
