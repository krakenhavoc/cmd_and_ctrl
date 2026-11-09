package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hovel Hurler — Creature — Giant Warrior {3}{R/W}{R/W} (Lorwyn Eclipsed, #4458):
//
//	"This creature enters with two -1/-1 counters on it.
//	 {R/W}{R/W}, Remove a counter from this creature: Another target
//	 creature you control gets +1/+0 and gains flying until end of turn.
//	 Activate only as a sorcery."
//
// "A counter" is a counter of any kind (the cost names none).
func init() {
	Register(Spec{
		OracleID:     "fb9cd9d0-8eea-4aa8-bb9c-bb0aac847421",
		Name:         "Hovel Hurler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Hovel Hurler: enters with two -1/-1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{R/W}{R/W}, Remove a counter from this creature: Another target creature you control gets +1/+0 and gains flying until end of turn. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{R/W}{R/W}"), RemoveCountersFromThis("", 1)),
			Targets:      Another(TargetCreature("another target creature you control", YouControl())),
			SorcerySpeed: true,
			Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (BoostUntilEOT{Target: id, Power: 1, Label: "Hovel Hurler — +1/+0 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"},
					Label: "Hovel Hurler — flying until end of turn"}.Apply(ctx)
			},
		}},
	})
}
