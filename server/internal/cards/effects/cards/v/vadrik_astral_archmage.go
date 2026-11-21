package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vadrik, Astral Archmage — {1}{U}{R} Legendary Creature — Human Wizard
// 1/2 (#2586, ADR 0132):
//
//	"If it's neither day nor night, it becomes day as Vadrik enters.
//	 Instant and sorcery spells you cast cost {X} less to cast, where X
//	 is Vadrik's power.
//	 Whenever day becomes night or night becomes day, put a +1/+1
//	 counter on Vadrik."
//
// The discount is Animar's computed reduction, read off Vadrik's current
// power (the cost query's Source is the live permanent), so each flip's
// counter makes the next spell cheaper. A reduction spends generic mana
// only (CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9ab17ce-ba07-4ce8-99f9-9c4c633957f7",
		Name:         "Vadrik, Astral Archmage",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.BecomesDayAsEnters(),
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int { return q.Source.CurrentPower() },
				"Instant and sorcery spells you cast cost {X} less to cast, where X is Vadrik's power.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverDayBecomesNightOrNightBecomesDay("Vadrik — put a +1/+1 counter on Vadrik", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
