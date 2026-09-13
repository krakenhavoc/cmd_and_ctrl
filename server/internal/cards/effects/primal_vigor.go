package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primal Vigor — Enchantment {4}{G} (EDHREC rank 1214):
//
//	"If one or more tokens would be created, twice that many of those
//	 tokens are created instead.
//	 If one or more +1/+1 counters would be put on a creature, twice
//	 that many +1/+1 counters are put on that creature instead."
//
// The symmetrical Doubling Season: EVERY player's +1/+1 counters and
// tokens, not just the controller's. The counter half is Branching
// Evolution's replacement with the controller check removed — any
// creature, anyone's — and it composes with the other doublers
// through the CR 616 order prompt exactly as they compose with each
// other.
//
// Sandbox simplification, declared — Doubling Season's own posture:
// the TOKEN half is not implemented. There is no replacement-event
// kind for token creation (the Academy Manufactor / Anointed
// Procession gap), so tokens are made at their printed count. Weaker
// than printed for everyone at the table, never stronger.
func init() {
	Register(Spec{
		OracleID:     "c665544f-557b-4631-a1dc-39571470ca2e",
		Name:         "Primal Vigor",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Token doubling isn't implemented — only the +1/+1 counter half works."},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterName != "+1/+1" || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && target.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Primal Vigor: double +1/+1 counters",
		}},
	})
}
