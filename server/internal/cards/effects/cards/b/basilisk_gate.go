package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8733a4fc-4068-4af4-9598-dc3d895e8556",
		Name:         "Basilisk Gate",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}: Target creature gets +X/+X until end of turn, where X is the number of Gates you control. " +
				"Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				x := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && effects.CardHasSubtype(c, "Gate") {
						x++
					}
				}
				return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Basilisk Gate — +X/+X until end of turn"}.Apply(ctx)
			},
		}},
	})
}
