package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unholy Grotto — Land:
//
//	"{T}: Add {C}.
//	 {B}, {T}: Put target Zombie card from your graveyard on top of your
//	 library."
//
// Academy Ruins' shape with the cost and the card filter swapped (Lord of
// the Undead's "Zombie card"): the put goes through the shared exit
// primitive, and a target that left the graveyard in response makes the
// ability fizzle (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c28211c6-a5ee-40c3-bb6a-da3e7e73fd95",
		Name:         "Unholy Grotto",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}: Put target Zombie card from your graveyard on top of your library.",
			Cost:    Plus(ManaCost("{B}"), TapCost()),
			Targets: TargetCardInGraveyard("target Zombie card from your graveyard", YouOwn(), Subtype("Zombie")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutChosenTargetOnTopOfLibrary(g, item)
			},
		}},
	})
}
