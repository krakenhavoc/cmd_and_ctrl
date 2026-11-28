package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zirilan of the Claw — Legendary Creature — Lizard Shaman {3}{R}{R}, 3/4
// (Mirage, S82 OLD b800, #5633):
//
//	"{1}{R}{R}, {T}: Search your library for a Dragon permanent card, put that
//	 card onto the battlefield, then shuffle. That Dragon gains haste until end
//	 of turn. Exile it at the beginning of the next end step."
//
// Postmortem Lunge's haste-and-exile rider over a search for one Dragon
// permanent card: the haste lasts until end of turn, and the exile is a
// delayed trigger naming that card (it does nothing if the Dragon has left by
// then). Finding nothing is allowed (CR 701.23b) and still shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3aeb28ac-8ed2-4a90-a631-d84361ac6acc",
		Name:         "Zirilan of the Claw",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}, {T}: Search your library for a Dragon permanent card, put that card onto the battlefield, then shuffle. That Dragon gains haste until end of turn. Exile it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}{R}"), effects.TapCost()),
			Effect:  effects.OldB800ZirilanSearch,
		}},
	})
}
