package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kataki, War's Wage — Legendary Creature — Spirit {1}{W}, 2/1:
//
//	"All artifacts have 'At the beginning of your upkeep, sacrifice this
//	 artifact unless you pay {1}.'"
//
// A layer-6 grant of a catalog bundle (ADR 0093) to every artifact on the
// battlefield, whoever controls it: the recipient is "this artifact" and its
// controller is the "you" of its upkeep trigger, so each player pays for
// the artifacts they control, on their own upkeep. The sacrifice is the
// shared keyed upkeepSacrificeSource.
func init() {
	const key = "kataki-wars-wage/artifact-upkeep"
	Register(Spec{
		OracleID:     "27ffa6e9-c161-4f53-8b41-7fae5f199f41",
		Name:         "Kataki, War's Wage",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       key,
			Triggered: []game.TriggeredAbility{UpkeepSacrificeUnlessYouPay("This artifact", "{1}")},
			Text:      "At the beginning of your upkeep, sacrifice this artifact unless you pay {1}.",
		}},
		Static: []game.StaticAbility{GrantAbilities(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsArtifact()
		}, key)},
	})
}
