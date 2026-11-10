package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tocasia's Dig Site — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}: Surveil 1."
//
// Broodheart Engine's plain surveil as an activated ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a91f93fd-e428-4a36-b1b3-604b47a34287",
		Name:          "Tocasia's Dig Site",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{cmdr2B01TapForColorless()},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}: Surveil 1.",
			Cost:  Plus(ManaCost("{3}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Surveil{N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
