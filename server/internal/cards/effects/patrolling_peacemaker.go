package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Patrolling Peacemaker — Artifact Creature — Robot Soldier {2}{W}, 0/0
// (Foundations Commander):
//
//	"This creature enters with two +1/+1 counters on it.
//	 Whenever an opponent commits a crime, proliferate."
//
// Afiya Grove's entry counters and Flux Channeler's proliferate. A crime
// is the once-per-announcement event (CR 700.13, YouCommittedCrime's
// reading) by a player who is not the Peacemaker's controller; each
// crime is one proliferate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa44171d-1c76-40a7-82fc-5ef2abbca647",
		Name:         "Patrolling Peacemaker",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Patrolling Peacemaker: enters with two +1/+1 counters on it"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBecomesTarget && ev.CrimeCommitted && ev.Actor != source.Controller
			}, "Patrolling Peacemaker — proliferate", Do(Proliferate{})),
		},
	})
}
