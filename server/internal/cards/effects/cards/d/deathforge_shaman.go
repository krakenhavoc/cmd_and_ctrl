package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathforge Shaman — Creature — Ogre Shaman {4}{R}, 4/3 (Worldwake, #5582):
//
//	"Multikicker {R} (You may pay an additional {R} any number of times as
//	 you cast this spell.)
//	 When this creature enters, it deals damage to target player or
//	 planeswalker equal to twice the number of times it was kicked."
//
// The count is Stronghold Arena's: read off the record the resolution
// carried onto the permanent (CR 400.7d) as the trigger is put on the
// stack, and stamped on the item (wwkWhenEntersPerKick), so it survives
// the creature leaving in response. Unkicked, the trigger still goes on
// the stack and does nothing, as printed (there is no intervening if).
// The target is Lava Axe's "target player or planeswalker".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1d3cfdd3-e16f-4955-9b81-0d6fc981c168",
		Name:          "Deathforge Shaman",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Multikicker("{R}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.WwkWhenEntersPerKick("Deathforge Shaman — damage to target player or planeswalker equal to twice the number of times it was kicked", 2,
				effects.TargetPlayerOrPlaneswalker(), deathforgeShamanDamage),
		},
	})
}

// deathforgeShamanDamage deals the stamped amount to the first target
// still legal as the trigger resolves (CR 608.2b).
func deathforgeShamanDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || item.Params.Amount <= 0 {
		return nil
	}
	return effects.DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: item.Params.Amount}.Apply(ctx)
}
