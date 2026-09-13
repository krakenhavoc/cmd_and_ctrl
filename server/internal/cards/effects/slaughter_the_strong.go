package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slaughter the Strong — Sorcery {1}{W}{W} (EDHREC rank 2355):
//
//	"Each player chooses any number of creatures they control with
//	 total power 4 or less, then sacrifices all other creatures they
//	 control."
//
// The go-wide deck's one-sided wrath. The engine has no prompt for
// "choose a set of your permanents under a sum constraint" — the
// sacrifice prompt picks one permanent and carries no continuation,
// so it cannot be asked repeatedly until the survivors fit — and the
// choice is made FOR each player instead: they keep their creatures
// lowest power first (battlefield order on a tie) while the total
// stays at four or less, which keeps as many as possible, and
// sacrifice the rest. Every keep set is fixed before anything is
// sacrificed, as the printed "then" requires, and the sacrifices go
// through the ordinary path so Blood Artist sees each one.
//
// Declared simplification (weaker than printed — a player who would
// rather keep one 4-power commander than four 1/1 tokens is not
// asked): the auto-pick. The seam is a resolution-time "choose any
// number of your permanents matching a constraint" prompt.
func init() {
	Register(Spec{
		OracleID:     "7a3569a0-a55f-40c9-9588-92db35e26567",
		Name:         "Slaughter the Strong",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You don't choose which creatures to keep — each player keeps their lowest-power creatures while the total power stays at 4 or less, and sacrifices the rest."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b22EachPlayerKeepsWithinPowerAndSacrificesTheRest(ctx.Game, item, 4)
		},
	})
}
