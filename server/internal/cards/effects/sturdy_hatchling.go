package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sturdy Hatchling — Creature — Elemental {3}{G/U}, 6/6 (Eventide, #5590):
//
//	"This creature enters with four -1/-1 counters on it.
//	 {G/U}: This creature gains shroud until end of turn. (It can't be the
//	 target of spells or abilities.)
//	 Whenever you cast a green spell, remove a -1/-1 counter from this
//	 creature.
//	 Whenever you cast a blue spell, remove a -1/-1 counter from this
//	 creature."
//
// The four counters are an entry replacement (Moonshadow's shape), and each
// colour is its own trigger: a spell of both colours removes two. A
// creature with no -1/-1 counter left removes nothing.
// The shroud row is Advanced Hoverguard's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08e95b6c-9f4c-42d9-ad64-2d905d781d49",
		Name:         "Sturdy Hatchling",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{eveB1HatchlingCounters("Sturdy Hatchling")},
		Activated: []ActivatedAbility{{
			Label:   "{G/U}: This creature gains shroud until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{G/U}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("shroud", "Sturdy Hatchling — shroud until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{
			eveB1RemoveCounterOnCast("Sturdy Hatchling", "G"),
			eveB1RemoveCounterOnCast("Sturdy Hatchling", "U"),
		},
	})
}
