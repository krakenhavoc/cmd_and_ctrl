package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// That's Rough Buddy — Instant — Lesson {1}{W} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Put a +1/+1 counter on target creature. Put two +1/+1 counters on
//	 that creature instead if a creature left the battlefield under your
//	 control this turn.
//	 Draw a card."
//
// The "instead" is read as the spell resolves, off the per-player
// CreaturesLeft tally (avB28CreatureLeftUnderYourControl). With its only
// target gone the spell does not resolve (CR 608.2b), draw included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc3ee37a-f676-44af-80d0-774ff271ce11",
		Name:         "That's Rough Buddy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				n := 1
				if avB28CreatureLeftUnderYourControl(ctx.Game, ctx.Controller()) {
					n = 2
				}
				if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
