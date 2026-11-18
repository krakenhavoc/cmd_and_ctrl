package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Valiant Endeavor — Sorcery {4}{W}{W} (Adventures in the Forgotten Realms):
//
//	"Roll two d6 and choose one result. Destroy each creature with power
//	 greater than or equal to that result. Then create a number of 2/2
//	 white Knight creature tokens with vigilance equal to the other
//	 result."
//
// The shared two-dice shape is afrEndeavorTail. The Knights are the
// destruction's own tail, so they arrive after the destroyed creatures
// have settled (a prompt in the sweep waits for its answer first).
func init() {
	Register(Spec{
		OracleID: "4ba1f2c5-6456-4809-a563-fd5bd311869a", Name: "Valiant Endeavor", Completeness: CompletenessFull,
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return RollDice{Sides: 6, N: 2, Then: valiantEndeavorRolled}.Apply(ctx)
		},
	})
}

var valiantEndeavorRolled = afrEndeavorTail("dice/valiant-endeavor", "Valiant Endeavor",
	func(chosen, other int) string {
		return "Destroy power " + strconv.Itoa(chosen) + " or more, " + strconv.Itoa(other) + " Knights"
	},
	func(ctx *Context, chosen, other int) error {
		return DestroyAllMatching{
			Match:    And(Creature(), PowerGE(chosen)),
			ThenTail: withN(ctx.ItemTail(valiantEndeavorKnights), other),
		}.Apply(ctx)
	})

// valiantEndeavorKnights is "then create a number of 2/2 white Knight
// creature tokens with vigilance equal to the other result": the number is
// the tail's N.
var valiantEndeavorKnights = CardTail("then/valiant-endeavor-knights", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return CreateToken{Controller: o.Tail.Player, Template: TokenCard("2/2 white Knight with vigilance"), N: o.Tail.N}.Apply(ctx)
})
