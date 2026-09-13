package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ad95eac2-5b48-4102-9988-a6d7b1ec30e0",
		Name:         "Patriar's Seal",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Untap target legendary creature you control.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target legendary creature you control", YouControl(), b05Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return UntapTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
