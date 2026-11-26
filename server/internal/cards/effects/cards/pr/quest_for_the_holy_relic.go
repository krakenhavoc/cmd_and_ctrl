package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quest for the Holy Relic — Enchantment {W} (Zendikar, S82 OLD b493):
//
//	"Whenever you cast a creature spell, you may put a quest counter on this
//	 enchantment.
//	 Remove five quest counters from this enchantment and sacrifice it: Search
//	 your library for an Equipment card, put it onto the battlefield, attach it
//	 to a creature you control, then shuffle."
//
// Khalni Heart Expedition's shape: an optional counter trigger and an
// activation that pays its counters and the sacrifice as costs. The search puts
// the Equipment onto the battlefield and the controller then picks the creature
// it is attached to (Old493SearchEquipmentOntoBattlefieldAttach); with no
// creature to carry it, it stays unattached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fa5803b-3604-46c7-9774-bf70e4a2b925",
		Name:         "Quest for the Holy Relic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.YouCast(effects.Creature()), "Quest for the Holy Relic — put a quest counter",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})), "Quest for the Holy Relic — put a quest counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove five quest counters from this enchantment and sacrifice it: Search your library for an Equipment card, put it onto the battlefield, attach it to a creature you control, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 5), effects.SacrificeThis()),
			Effect:  effects.Old493SearchEquipmentOntoBattlefieldAttach,
		}},
	})
}
