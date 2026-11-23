package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fb8f80cc-6214-4ab9-a9a9-1873ab9feb0c",
		Name:         "Xenic Poltergeist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Until your next upkeep, target noncreature artifact becomes an artifact creature with power and toughness each equal to its mana value.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target noncreature artifact", effects.And(effects.Artifact(), effects.Noncreature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ArtifactBecomesCreatureOfItsManaValue(ctx, effects.DurationUntilYourNextUpkeep(ctx, item.Controller),
					"Xenic Poltergeist — an artifact creature until your next upkeep")
			},
		}},
	})
}
