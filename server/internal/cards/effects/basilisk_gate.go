package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basilisk Gate — Land — Gate:
//
//	"{T}: Add {C}.
//	 {2}, {T}: Target creature gets +X/+X until end of turn, where X is the
//	 number of Gates you control. Activate only as a sorcery."
//
// X is read once as the ability resolves, counting the Gates its controller
// controls then (Basilisk Gate itself included), and the boost is fixed from
// then on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8733a4fc-4068-4af4-9598-dc3d895e8556",
		Name:         "Basilisk Gate",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}: Target creature gets +X/+X until end of turn, where X is the number of Gates you control. " +
				"Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0)),
			Cost:         Plus(ManaCost("{2}"), TapCost()),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && hasSubtype(c, "Gate") {
						x++
					}
				}
				return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Basilisk Gate — +X/+X until end of turn"}.Apply(ctx)
			},
		}},
	})
}
