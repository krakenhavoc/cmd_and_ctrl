package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Budoka Pupil // Ichiga, Who Topples Oaks — {1}{G}{G} Creature — Human
// Monk 2/2 // Legendary Creature — Spirit 4/3, a flip card (#3216, ADR
// 0151):
//
//	Top:     "Whenever you cast a Spirit or Arcane spell, you may put a
//	          ki counter on this creature.
//	          At the beginning of the end step, if there are two or more
//	          ki counters on this creature, you may flip it."
//	Flipped: "Trample
//	          Remove a ki counter from Ichiga: Target creature gets
//	          +2/+2 until end of turn."
//
// The ki cycle's shared triggers are kiFlipTriggers. The counters stay
// on the permanent through the flip (it is the same object), so Ichiga
// spends the ones Budoka Pupil collected. Ichiga is green and {1}{G}{G}
// (CR 710.1c).
//
// No simplification.
const BudokaPupilOracleID = "6f571092-7244-4223-aa14-f5133cc000d8"

func init() {
	effects.Register(effects.Spec{
		OracleID:     BudokaPupilOracleID,
		Name:         "Budoka Pupil",
		Completeness: effects.CompletenessFull,
		Triggered:    effects.KiFlipTriggers("Budoka Pupil"),
	})
	effects.Register(effects.Spec{
		OracleID:        BudokaPupilOracleID + "#1",
		Name:            "Ichiga, Who Topples Oaks",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a ki counter from Ichiga: Target creature gets +2/+2 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
			Cost:    effects.RemoveAKiCounter(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Ichiga, Who Topples Oaks — +2/+2"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
