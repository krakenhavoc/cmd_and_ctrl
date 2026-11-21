package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wose Pathfinder — Creature — Human Shaman {1}{G}, 1/1 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"{T}: Add one mana of any color.
//	 {6}{G}, {T}: Another target creature gets +3/+3 and gains trample
//	 until end of turn."
//
// An any-color tap (Abundant Growth's mana row) and a targeted pump that
// taps the Pathfinder as part of its cost; "another" keeps it from
// targeting itself. A target that left in response gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb2ff7f2-b7fd-4984-a4e8-cf57b619c1a5",
		Name:         "Wose Pathfinder",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{6}{G}, {T}: Another target creature gets +3/+3 and gains trample until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
			Cost:    Plus(ManaCost("{6}{G}"), game.AbilityCost{Tap: true}),
			Targets: Another(TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return gapB125PumpTrample(ctx, id, 3, 3, "Wose Pathfinder")
			},
		}},
	})
}
