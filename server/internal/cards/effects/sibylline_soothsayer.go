package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sibylline Soothsayer — Creature — Human Warlock {2}{R}, 3/2 (Doctor Who):
//
//	"Temporal Foresight — When this creature enters, reveal cards from the top of
//	 your library until you reveal a nonland card with mana value 3 or greater.
//	 Exile that card with three time counters on it. If it doesn't have suspend, it
//	 gains suspend. Put the rest of the revealed cards on the bottom of your library
//	 in a random order."
//
// The reveal is the engine's (revealUntil: a token in a library is revealed but never
// stops the run). The hit is exiled and, once it has landed in exile, gets three time
// counters and suspend (Kang Prime's GainSuspendForEffect, CR 702.62a), then the rest
// of the revealed cards still in the library go to the bottom in a random order
// (nothing to exile when the library runs out: they all go to the bottom).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70718e4a-c3c2-4fcd-96f8-98cd7fee257c",
		Name:         "Sibylline Soothsayer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Temporal Foresight — reveal until a nonland card with mana value 3 or greater; exile it with three time counters; put the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					run, hit := revealUntil(ctx, ctx.Controller(), func(c game.Card) bool { return !c.IsLand() && c.ManaValue() >= 3 },
						"Sibylline Soothsayer — reveal until a nonland card with mana value 3 or greater")
					if hit == uuid.Nil {
						return gapB88BottomRevealed(ctx, run)
					}
					var rest []uuid.UUID
					for _, id := range run {
						if id != hit {
							rest = append(rest, id)
						}
					}
					tail := ctx.TailN(gapB88SuspendedWithTimeCounters, 3)
					tail.Carry = rest
					return ExileTarget{Target: hit, ThenTail: tail}.Apply(ctx)
				}),
		},
	})
}
