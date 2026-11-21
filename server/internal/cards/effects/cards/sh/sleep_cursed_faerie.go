package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sleep-Cursed Faerie — Creature — Faerie Wizard {U}, 3/3 (Wilds of
// Eldraine, #4471):
//
//	"Flying, ward {2}
//	 This creature enters tapped with three stun counters on it. (If it
//	 would become untapped, remove a stun counter from it instead.)
//	 {1}{U}: Untap this creature."
//
// The entry is Tonberry's SelfEntersTappedWithCounters with three stun
// counters, ward {2} is Hulking Raptor's WardMana, and the untap is
// Devoted Druid's UntapTarget on itself. The untap goes through the
// engine's one untap primitive, so while a stun counter remains, the
// activation removes one instead of untapping it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a17633aa-b1a7-4ae5-b051-6747975d8a2f",
		Name:            "Sleep-Cursed Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTappedWithCounters(game.CounterStun, 3)},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Sleep-Cursed Faerie — ward {2}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Untap this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
