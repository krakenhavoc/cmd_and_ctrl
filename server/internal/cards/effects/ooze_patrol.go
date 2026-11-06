package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ooze Patrol — Creature — Ooze {3}{G}, 2/2 (Aetherdrift, #4464):
//
//	"When this creature enters, mill two cards, then put a +1/+1 counter
//	 on this creature for each artifact and/or creature card in your
//	 graveyard."
//
// The mill (CR 701.17a), then Aatchik's count of the artifact and/or
// creature cards in your graveyard, read after the mill so the cards it
// just milled count (CR 608.2c): an artifact creature card counts once.
// The count waits for any prompt the mill asks. If the Ooze has left the
// battlefield by then, the counters have nothing to go on (CR 400.7), but
// the mill still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9e3e8b0-4276-419d-9326-be56af3c41be",
		Name:         "Ooze Patrol",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ooze Patrol — mill two cards, then a +1/+1 counter for each artifact and/or creature card in your graveyard",
				Do(MillCards{N: 2}, Instruction(func(ctx *Context) error {
					n := dftABArtifactOrCreatureCardsInGraveyard(ctx.Game, ctx.Controller())
					if n == 0 {
						return nil
					}
					return CounterOnThis{Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}))),
		},
	})
}
