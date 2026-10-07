package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// All-Fates Scroll — Artifact {3}:
//
//	"{T}: Add one mana of any color.
//	 {7}, {T}, Sacrifice this artifact: Draw X cards, where X is the
//	 number of differently named lands you control."
//
// X is counted when the ability resolves (after the Scroll has been
// sacrificed, which cannot change the lands). Two lands with the same
// name count once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cef5361a-a189-4e05-a8b8-fc764aed73b9",
		Name:         "All-Fates Scroll",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label: "{7}, {T}, Sacrifice this artifact: Draw X cards, where X is the number of differently named lands you control.",
			Cost:  game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{7}"},
			Effect: func(g *game.Game, item *game.StackItem) error {
				names := map[string]bool{}
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsLand() {
						names[c.Name] = true
					}
				}
				return DrawCards{Player: item.Controller, N: len(names)}.Apply(NewContext(g, item))
			},
		}},
	})
}
