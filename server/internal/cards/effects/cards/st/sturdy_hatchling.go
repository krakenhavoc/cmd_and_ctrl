package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "08e95b6c-9f4c-42d9-ad64-2d905d781d49",
		Name:         "Sturdy Hatchling",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.EveB1HatchlingCounters("Sturdy Hatchling")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G/U}: This creature gains shroud until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{G/U}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("shroud", "Sturdy Hatchling — shroud until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.EveB1RemoveCounterOnCast("Sturdy Hatchling", "G"),
			effects.EveB1RemoveCounterOnCast("Sturdy Hatchling", "U"),
		},
	})
}
