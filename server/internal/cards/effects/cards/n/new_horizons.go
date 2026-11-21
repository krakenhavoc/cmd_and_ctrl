package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// New Horizons — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "8834d3c9-cdb5-423b-bc78-94a94add5e74", Name: "New Horizons", Completeness: effects.CompletenessFull,
		Targets:   effects.EnchantLand(),
		Grants:    []effects.AbilityGrant{{Key: "new-horizons/two-of-one-color", Text: "{T}: Add two mana of any one color.", Mana: []effects.ManaAbility{{Cost: effects.ManaAbilityCost{Tap: true}, Produced: effects.OneColorOfAmount(2), Label: "Add two mana of any one color"}}}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached("new-horizons/two-of-one-color")},
		Triggered: []game.TriggeredAbility{effects.Targeting(effects.WhenThisEnters("New Horizons — put a +1/+1 counter on target creature you control", effects.PlusOneCounterOnChosenTarget), effects.TargetCreature("target creature you control", effects.YouControl()))},
	})
}
