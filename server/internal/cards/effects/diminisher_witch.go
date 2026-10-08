package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diminisher Witch — Creature — Human Warlock {2}{U}, 3/2:
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as
//	 you cast this spell.)
//	 When this creature enters, if it was bargained, create a Cursed
//	 Role token attached to target creature an opponent controls. (If
//	 you control another Role on it, put that one into the graveyard.
//	 Enchanted creature is 1/1.)"
//
// The intervening if (CR 603.4) reads the permanent's own carried
// record of what was paid (ADR 0073 §5), as offspring does, so an
// unbargained Witch, a reanimated one and a token copy put nothing on
// the stack, and nobody is asked for a target.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "c2283655-f6b8-4399-ba1f-68a7107f9485",
		Name:          "Diminisher Witch",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, diminisherWitchWasBargained),
					"Diminisher Witch — create a Cursed Role token attached to target creature an opponent controls",
					createRoleOnFirstTarget(RoleCursed)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

// diminisherWitchWasBargained is "if it was bargained" for a permanent:
// the bargain cost's index is among those the resolution path carried
// onto it.
func diminisherWitchWasBargained(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return game.OptionalCostTimesPaid(*source, source.Provenance.OptionalCosts, g2BargainKey) > 0
}
