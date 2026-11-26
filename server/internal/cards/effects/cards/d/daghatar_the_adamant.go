package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daghatar the Adamant — {3}{W} Legendary Creature — Human Warrior 0/0 (Fate
// Reforged):
//
//	"Vigilance
//	 Daghatar enters with four +1/+1 counters on it.
//	 {1}{B/G}{B/G}: Move a +1/+1 counter from target creature onto a second
//	 target creature."
//
// The counters are an entry replacement (CR 614.1c), as Yorvo's are. The move
// is Combine Guildmage's: both targets must still be legal and the first must
// still hold a counter, the counter lands on the second creature first and
// comes off the first only if it did (CR 122.5); the second target is a
// different creature.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b303881-ac8d-4459-b04e-768a6804763c",
		Name:            "Daghatar the Adamant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 4, "Daghatar the Adamant: enters with four +1/+1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B/G}{B/G}: Move a +1/+1 counter from target creature onto a second target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{B/G}{B/G}"),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature"),
				effects.Distinct(effects.TargetCreature("a second target creature")),
			),
			Effect: effects.MoveAPlusOneCounterFromFirstToSecond,
		}},
	})
}
