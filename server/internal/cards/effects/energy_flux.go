package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Energy Flux — Enchantment {2}{U}:
//
//	"All artifacts have 'At the beginning of your upkeep, sacrifice this
//	 artifact unless you pay {2}.'"
//
// A layer-6 grant of a catalog bundle (ADR 0093) to every artifact on the
// battlefield, whoever controls it: the recipient is "this artifact" and its
// controller is the "you" of its upkeep trigger, so each player pays for the
// artifacts they control on their own upkeep. The sacrifice is the shared
// keyed upkeepSacrificeSource.
func init() {
	const key = "energy-flux/artifact-upkeep"
	Register(Spec{
		OracleID:     "7a756cd1-29a8-4edf-bb74-fbb5b4020022",
		Name:         "Energy Flux",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       key,
			Triggered: []game.TriggeredAbility{UpkeepSacrificeUnlessYouPay("This artifact", "{2}")},
			Text:      "At the beginning of your upkeep, sacrifice this artifact unless you pay {2}.",
		}},
		Static: []game.StaticAbility{GrantAbilities(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsArtifact()
		}, key)},
	})
}
