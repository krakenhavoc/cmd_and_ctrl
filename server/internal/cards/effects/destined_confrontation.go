package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Destined Confrontation — Sorcery {2}{W}{W} (Avatar: The Last
// Airbender, #4459):
//
//	"Each player chooses any number of creatures they control with
//	 total power 4 or less, then sacrifices all other creatures they
//	 control."
//
// Slaughter the Strong's text at a different cost, so it runs the same
// keep-set (keepCreaturesThenSacrificeTheRest) under the same registered
// run and set rule (slaughterKeep): the clause is one rule, shared by
// every card that prints it. Each player is asked once, in APNAP order
// (the printed ruling: the active player first, then each other player
// in turn order, knowing the choices made before them), for the creatures
// it keeps, total power 4 or less read after layers (a negative power
// subtracts from the total, as the ruling says); then every other
// creature goes in one simultaneous sacrifice. A player whose whole board
// already fits is not asked. Nothing is targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "9703140f-36ee-4427-a390-c1dbdb9e4ee9",
		Name:     "Destined Confrontation",
		// ADR 0126 §6: each player keeps creatures with total power 4 or less.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return keepCreaturesThenSacrificeTheRest(ctx.Game, item, apnapPlayers(ctx.Game),
				"Destined Confrontation — choose any number of creatures with total power 4 or less to keep; the rest are sacrificed",
				0, slaughterKeep)
		},
	})
}
