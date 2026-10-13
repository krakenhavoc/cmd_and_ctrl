package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slaughter the Strong — Sorcery {1}{W}{W} (EDHREC rank 2355):
//
//	"Each player chooses any number of creatures they control with
//	 total power 4 or less, then sacrifices all other creatures they
//	 control."
//
// The go-wide deck's one-sided wrath. Each player is asked once, over
// their own creatures, for the set they keep (any number, with a rule on
// the set: total power 4 or less, read after layers), and only when every
// seat has answered does everything else go in one simultaneous
// sacrifice (keepCreaturesThenSacrificeTheRest, CR 608.2c, 701.21), so a
// Blood Artist swept by the same spell sees every death. A player whose
// whole board already fits is not asked. Nothing is targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "7a3569a0-a55f-40c9-9588-92db35e26567",
		Name:     "Slaughter the Strong",
		// ADR 0126 §6: each player keeps creatures with total power 4 or less.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return keepCreaturesThenSacrificeTheRest(ctx.Game, item, apnapPlayers(ctx.Game),
				"Slaughter the Strong — choose any number of creatures with total power 4 or less to keep; the rest are sacrificed",
				0, totalPowerAtMost(4))
		},
	})
}
