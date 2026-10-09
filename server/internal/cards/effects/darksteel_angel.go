package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Darksteel Angel — Artifact Creature — Angel {9}, 4/4:
//
//	"Flying, indestructible
//	 You can't lose the game and your opponents can't win the game.
//	 Creatures you control can't have -1/-1 counters put on them."
//
// The first static is the derived game-end gate Platinum Angel and
// Herald of Eternal Dawn use (ADR 0057 Decision 4). The second is a
// pure-cancel counter replacement over the controller's creatures: a
// -1/-1 counter placement on one of them simply does not happen, so
// wither and infect damage deal their damage without the counters
// (CR 614.17). Removing counters, and +1/+1 counters, are untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d0259f7d-9bd7-4722-8fb9-439768308516",
		Name:            "Darksteel Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "indestructible"},
		GameEndGates:    YouCantLoseOpponentsCantWin(),
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterName != game.CounterMinusOne || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.IsCreature() && target.Controller == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			PureCancel: true,
			Label:      "Darksteel Angel: creatures you control can't have -1/-1 counters put on them",
		}},
	})
}
