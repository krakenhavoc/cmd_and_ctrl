package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// New Horizons — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "8834d3c9-cdb5-423b-bc78-94a94add5e74", Name: "New Horizons", Completeness: CompletenessFull,
		Targets:   EnchantLand(),
		Grants:    []AbilityGrant{{Key: "new-horizons/two-of-one-color", Text: "{T}: Add two mana of any one color.", Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: OneColorOfAmount(2), Label: "Add two mana of any one color"}}}},
		Static:    []game.StaticAbility{GrantAbilitiesToAttached("new-horizons/two-of-one-color")},
		Triggered: []game.TriggeredAbility{Targeting(WhenThisEnters("New Horizons — put a +1/+1 counter on target creature you control", PlusOneCounterOnChosenTarget), TargetCreature("target creature you control", YouControl()))},
	})
}
