package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Component Pouch — Artifact {3}:
//
//	"{T}, Remove a component counter from this artifact: Add two mana
//	 of different colors.
//	 {T}: Roll a d20.
//	 1—9 | Put a component counter on this artifact.
//	 10—20 | Put two component counters on this artifact."
//
// Two tap abilities, so one or the other each turn: roll to stock the
// pouch, then spend a counter for two mana of DIFFERENT colours (#2558,
// DifferentColors(2)). The counter comes off as part of the mana
// ability's cost (#789), so a pouch with none cannot be tapped for
// mana. The roll is an ordinary activated ability on the stack; the
// counters land on the pouch only if it is still the same object when
// it resolves (#1432).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bb1863e-50f3-4e08-884e-1f58ee55817f",
		Name:         "Component Pouch",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				RemoveCounters: RemoveCountersFromThis("component", 1).RemoveCounters,
			},
			Produced: DifferentColors(2),
			Label:    "{T}, Remove a component counter from this artifact: Add two mana of different colors",
		}},
		Activated: []ActivatedAbility{{
			Label: "{T}: Roll a d20.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				rolls, err := rollDice(ctx, 20, 1)
				if err != nil || len(rolls) == 0 {
					return err
				}
				if zone := g.FindCardZoneForEffect(item.SourceCardID); zone == nil || zone.Kind != game.ZoneBattlefield || sourceIsNewObject(g, item) {
					return nil
				}
				n := 1
				if rolls[0] >= 10 {
					n = 2
				}
				return g.AddCounterForEffect(item.SourceCardID, "component", n)
			},
		}},
	})
}
