package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serpent Specialist — Creature — Human Snake Villain {G}, 1/1 (Marvel Super
// Heroes, #4455):
//
//	"Deathtouch
//	 Power-up — {3}{G}: Put two +1/+1 counters on this creature."
//
// The shared PowerUp seam: once per permanent, its mana cost off the price
// on the turn it entered.
//
// No simplification.
func init() {
	// Purpose set by hand: more AnswerPump callers of powerUpAnswering
	// trip unparam (#4571 drops that parameter).
	ab := PowerUp("{3}{G}: Put two +1/+1 counters on this creature.",
		ManaCost("{3}{G}"), plusOneCountersOnThis(2))
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "556f127f-6131-4171-b1ed-11420c21d345",
		Name:            "Serpent Specialist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []ActivatedAbility{ab},
	})
}
