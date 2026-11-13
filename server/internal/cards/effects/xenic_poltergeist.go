package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Xenic Poltergeist — Creature — Spirit (1/1) for {1}{B}{B}:
//
//	"{T}: Until your next upkeep, target noncreature artifact becomes an
//	 artifact creature with power and toughness each equal to its mana
//	 value."
//
// Karn, the Great Creator's +1 with another duration: a creature type
// added in layer 4 and a base power and toughness set in layer 7b at the
// artifact's mana value as the ability resolves, timed by
// game.UntilYourNextUpkeep (#5737). The artifact is still a creature
// through its controller's untap step and stops being one as the
// Poltergeist's controller's next upkeep begins. The {T} cost is
// subject to summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb8f80cc-6214-4ab9-a9a9-1873ab9feb0c",
		Name:         "Xenic Poltergeist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Until your next upkeep, target noncreature artifact becomes an artifact creature with power and toughness each equal to its mana value.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target noncreature artifact", And(Artifact(), Noncreature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return artifactBecomesCreatureOfItsManaValue(ctx, DurationUntilYourNextUpkeep(ctx, item.Controller),
					"Xenic Poltergeist — an artifact creature until your next upkeep")
			},
		}},
	})
}
