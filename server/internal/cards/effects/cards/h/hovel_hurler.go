package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hovel Hurler — Creature — Giant Warrior {3}{R/W}{R/W} (Lorwyn Eclipsed, #4458):
//
//	"This creature enters with two -1/-1 counters on it.
//	 {R/W}{R/W}, Remove a counter from this creature: Another target
//	 creature you control gets +1/+0 and gains flying until end of turn.
//	 Activate only as a sorcery."
//
// "A counter" is a counter of any kind (the cost names none).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb9cd9d0-8eea-4aa8-bb9c-bb0aac847421",
		Name:         "Hovel Hurler",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterMinusOne, 2, "Hovel Hurler: enters with two -1/-1 counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{R/W}{R/W}, Remove a counter from this creature: Another target creature you control gets +1/+0 and gains flying until end of turn. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{R/W}{R/W}"), effects.RemoveCountersFromThis("", 1)),
			Targets:      effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			SorcerySpeed: true,
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Hovel Hurler — +1/+0 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"},
					Label: "Hovel Hurler — flying until end of turn"}.Apply(ctx)
			},
		}},
	})
}
