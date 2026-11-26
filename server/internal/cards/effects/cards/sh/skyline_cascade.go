package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyline Cascade — Land (Battle for Zendikar):
//
//	"This land enters tapped.
//	 When this land enters, target creature an opponent controls doesn't untap
//	 during its controller's next untap step.
//	 {T}: Add {U}."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79301ae1-8c9c-4723-be21-dc27e1646f35",
		Name:         "Skyline Cascade",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Skyline Cascade — target creature an opponent controls doesn't untap during its controller's next untap step",
				effects.S82OldB311DoesntUntapTheTarget("Skyline Cascade — it doesn't untap during its controller's next untap step")),
				effects.Harmful(effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()))),
		},
	})
}
