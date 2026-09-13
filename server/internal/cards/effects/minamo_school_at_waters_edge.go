package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Minamo, School at Water's Edge — Legendary Land (EDHREC rank 603):
//
//	"{T}: Add {U}.
//	 {U}, {T}: Untap target legendary permanent."
//
// An Island that untaps your commander. The second ability targets,
// so it is a CR 602 activated ability on the stack, with a mana and
// tap cost — Buried Ruin's shape without the sacrifice. "Legendary"
// is read off the effective supertypes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17784f90-89a1-47a5-83ef-ae60dfc30bd1",
		Name:         "Minamo, School at Water's Edge",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}: Untap target legendary permanent.",
			Cost:    Plus(ManaCost("{U}"), TapCost()),
			Targets: TargetPermanent("target legendary permanent", b05Legendary()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return UntapTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
