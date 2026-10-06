package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Black Breath — Sorcery {2}{B}:
//
//	"Creatures your opponents control get -1/-1 until end of turn. The
//	 Ring tempts you."
//
// The creatures are fixed as the spell resolves (CR 611.2c): one that
// enters later this turn is not shrunk.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0afbb32-dc50-436a-bd24-2990c103a4f8",
		Name:         "The Black Breath",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{
				Match:     And(Creature(), OpponentControls()),
				Power:     -1,
				Toughness: -1,
				Label:     "The Black Breath — -1/-1 until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return TheRingTemptsYou{}.Apply(ctx)
		},
	})
}
