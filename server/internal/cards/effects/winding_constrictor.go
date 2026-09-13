package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winding Constrictor — Creature — Snake {B}{G}, 2/3 (EDHREC rank
// 2069):
//
//	"If one or more counters would be put on an artifact or creature
//	 you control, that many plus one of each of those kinds of
//	 counters are put on that permanent instead.
//	 If you would get one or more counters, you get that many plus
//	 one of each of those kinds of counters instead."
//
// Hardened Scales for every kind of counter on every artifact and
// creature. A CR 614 replacement on the counter-placement event:
// a positive delta of any kind aimed at an artifact or creature the
// Constrictor's controller controls (post-layer types, so an
// animated Treasure counts) gets one more. The pipeline fires once
// per kind, which is what "of each of those kinds" means; "enters
// with" counters run through the same pipeline, so a Walking
// Ballista or Hangarback Walker enters one larger, and two
// Constrictors add two. A removal (a negative delta) is not counters
// being "put on" and passes through untouched.
//
// SANDBOX GAP, weaker than printed: the second sentence — counters
// YOU get (poison, energy, experience) — is not modelled. Player
// counters have no replacement pipeline (AddPlayerCounterForEffect
// notes it: RepEventCounter is card-targeted), so there is nothing
// for the second replacement to watch. Never stronger.
func init() {
	Register(Spec{
		OracleID:     "c9404d7d-a026-4082-9fcb-1ab571a136b5",
		Name:         "Winding Constrictor",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Counters you get yourself (poison, energy, experience) aren't increased — only counters put on your artifacts and creatures are."},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventCounter || ev.CounterDelta <= 0 {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				if !ok || target.Controller != src.Controller {
					return false
				}
				return target.IsArtifact() || target.IsCreature()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Winding Constrictor: one more counter of that kind",
		}},
	})
}
