package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impromptu Raid — Enchantment {3}{R/G} (Shadowmoor, S82 OLD b531, #5591):
//
//	"{2}{R/G}: Reveal the top card of your library. If it isn't a creature card,
//	 put it into your graveyard. Otherwise, put that card onto the battlefield.
//	 That creature gains haste. Sacrifice it at the beginning of the next end
//	 step."
//
// The card is revealed (CR 701.20), so the table sees it. A noncreature card goes
// to the graveyard; a creature enters through the keyed library entry, and haste
// (with no duration, as printed) and a delayed sacrifice bound to that object
// (CR 603.7a, Apprentice Necromancer's) follow once it has arrived. The ability
// may be activated more than once, each time with its own delayed sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f46fe37f-c552-44e6-9657-ff63dbbd0ccf",
		Name:         "Impromptu Raid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R/G}: Reveal the top card of your library. If it isn't a creature card, put it into your graveyard. Otherwise, put that card onto the battlefield. That creature gains haste. Sacrifice it at the beginning of the next end step.",
			Cost:    effects.ManaCost("{2}{R/G}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect:  effects.OldB531ImpromptuRaid,
		}},
	})
}
