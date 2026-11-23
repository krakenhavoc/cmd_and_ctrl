package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Patriar's Seal — Artifact {3} (EDHREC rank 1201):
//
//	"{T}: Add one mana of any color.
//	 {1}, {T}: Untap target legendary creature you control."
//
// The legends deck's rock that doubles as a pseudo-vigilance or a
// second mana-dork activation. The mana ability is the five-way pick
// with no identity narrowing; the untap is a mana-plus-tap activated
// ability with a target clause — legendary read off the effective
// supertypes (b05Legendary), so a layer effect that adds or removes
// Legendary composes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad95eac2-5b48-4102-9988-a6d7b1ec30e0",
		Name:         "Patriar's Seal",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Untap target legendary creature you control.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target legendary creature you control", effects.YouControl(), effects.B05Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return effects.UntapTarget{Target: item.Targets[0].ID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
