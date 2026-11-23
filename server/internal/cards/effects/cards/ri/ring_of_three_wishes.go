package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ring of Three Wishes — Artifact {5} (Magic 2014, #5563):
//
//	"This artifact enters with three wish counters on it.
//	 {5}, {T}, Remove a wish counter from this artifact: Search your library for a card, put that card into your hand, then shuffle."
//
// Wishclaw Talisman's counters and tutor without the hand-over: three wish
// counters as it enters, one removed as each activation's cost, so it can
// tutor three times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "677c3616-295f-493a-8b60-1e22b0224cfa",
		Name:         "Ring of Three Wishes",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("wish", 3, "Ring of Three Wishes: enters with three wish counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Remove a wish counter from this artifact: Search your library for a card, put that card into your hand, then shuffle.",
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.RemoveCountersFromThis("wish", 1)),
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Effect:  effects.RingOfThreeWishesTutor,
		}},
	})
}
