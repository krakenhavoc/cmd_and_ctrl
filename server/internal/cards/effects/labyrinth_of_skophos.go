package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Labyrinth of Skophos — Land:
//
//	"{T}: Add {C}.
//	 {4}, {T}: Remove target attacking or blocking creature from combat."
//
// Reconnaissance's removal (RemoveFromCombatForEffect, CR 506.4) with a
// target clause that accepts an attacker or a blocker of any controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "9ec5a487-d8ed-459a-8f58-56f6e9a2dfe8",
		Name:          "Labyrinth of Skophos",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{cmdr2B01TapForColorless()},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Remove target attacking or blocking creature from combat.",
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Targets: TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				id, ok := b16FirstLegalTargetCard(NewContext(g, item))
				if !ok {
					return nil
				}
				// A target that has already left combat is not an error.
				_ = g.RemoveFromCombatForEffect(id)
				return nil
			},
		}},
	})
}
