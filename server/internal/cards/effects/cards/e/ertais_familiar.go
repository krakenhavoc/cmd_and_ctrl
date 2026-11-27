package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ertai's Familiar — Creature — Illusion {1}{U}, 2/2 (Weatherlight):
//
//	"Phasing
//	 When this creature phases out or leaves the battlefield, mill three cards.
//	 {U}: Until your next upkeep, this creature can't phase out."
//
// One trigger watching two events (phasing out, CR 603.10b, and leaving the
// battlefield), so a phase-out and a death each mill three and a creature that
// does both in one go does not mill twice for one event. The {U} ability pins a
// "can't phase out" record to the creature until its controller's next upkeep
// (#6201).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "711cd80d-c58d-4c2a-87db-d2e4c75f60d3",
		Name:            "Ertai's Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"phasing"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventPhaseOut, game.EventLTB}, effects.Self,
				"Ertai's Familiar — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Until your next upkeep, this creature can't phase out.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ThisCantPhaseOutUntilYourNextUpkeep("Ertai's Familiar"),
		}},
	})
}
