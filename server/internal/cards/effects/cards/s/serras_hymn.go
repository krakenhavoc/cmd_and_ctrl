package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra's Hymn — Enchantment {W} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	Sacrifice this enchantment: Prevent the next X damage that would be dealt this turn to any number of targets, divided as you choose, where X is the number of verse counters on this enchantment.
//
// The division is announced with the targets (CR 601.2d), sized from the verse counters read as the ability is announced; each target gets a shield for its share (Remedy's shape). There is no mana cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66a3ffde-4c81-4798-b99b-a6cc3a93c3e3",
		Name:         "Serra's Hymn",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("Serra's Hymn")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Prevent the next X damage that would be dealt this turn to any number of targets, divided as you choose, where X is the number of verse counters on this enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.S82OldB724SacrificeCost(""),
			Targets: effects.TargetAny().WithCount(0, 0).Dividing(effects.DivideBy(effects.S82OldB724VerseCountOnAnnounce)),
			Effect:  effects.RprB04PreventDividedDamage("Serra's Hymn — prevent the next damage this turn"),
		}},
	})
}
