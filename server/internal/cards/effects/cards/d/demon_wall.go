package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demon Wall — Artifact Creature — Demon Wall {1}{B}, 3/3 (Final
// Fantasy, #4462):
//
//	"Defender
//	 Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 As long as this creature has a counter on it, it can attack as
//	 though it didn't have defender.
//	 {5}{B}: Put two +1/+1 counters on this creature."
//
// The waiver is Platypus-Bear's self-only layer 6 static
// (avB18CanAttackAsThoughNoDefenderWhile), read at declaration: a
// counter of any kind turns it on, and defender stays (CR 702.3b is
// waived, not removed). The activation is Afterburner Expert's two
// counters on its own source, without exhaust, so the counters it puts
// on also let it attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "077b5ed2-05c2-4a16-b5eb-794884ee1c21",
		Name:            "Demon Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "menace"},
		Static: []game.StaticAbility{
			effects.AvB18CanAttackAsThoughNoDefenderWhile(
				"As long as this creature has a counter on it, it can attack as though it didn't have defender.",
				func(_ *game.Game, source *game.Card) bool {
					return source != nil && effects.SocB9CountersOn(*source) > 0
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{B}: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, SourceCounters: 2},
			Cost:    effects.ManaCost("{5}{B}"),
			Effect:  effects.PlusOneCountersOnThis(2),
		}},
	})
}
