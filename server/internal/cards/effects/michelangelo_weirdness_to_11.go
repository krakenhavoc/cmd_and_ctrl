package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Michelangelo, Weirdness to 11 — {1}{G} Legendary Creature — Mutant
// Ninja Turtle 1/1 (Teenage Mutant Ninja Turtles, #4457):
//
//	"When Michelangelo enters, create a Mutagen token. (It's an artifact
//	 with "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")
//	 If one or more +1/+1 counters would be put on a creature you control,
//	 that many plus one +1/+1 counters are put on it instead."
//
// The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676). The replacement is
// Conclave Mentor's (Hardened Scales on a body): +1/+1 counters only, on a
// creature its controller controls, Michelangelo himself included, the
// count bumped by one (CR 614.1a). With two such effects the affected
// player orders them (CR 616.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "170ab932-9d50-4ce7-9a42-08e1edce7e7c",
		Name:         "Michelangelo, Weirdness to 11",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Michelangelo, Weirdness to 11 — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventCounterPlaced},
			AppliesTo: plusOneCountersOnACreatureYouControl,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Michelangelo, Weirdness to 11: +1 +1/+1 counter",
		}},
	})
}
