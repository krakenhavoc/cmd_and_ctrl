package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honest Rutstein — Legendary Creature — Human Warlock {1}{B}{G}, 3/2
// (Outlaws of Thunder Junction, #4468):
//
//	"When Honest Rutstein enters, return target creature card from your
//	 graveyard to your hand.
//	 Creature spells you cast cost {1} less to cast."
//
// The enters trigger is Hologram Stalker's return with a creature card
// for its target, chosen as the trigger goes on the stack (CR 603.3d):
// with no creature card in your graveyard it is removed. The discount is
// Krallenhorde Howler's cost modifier, generic mana only (CR 601.2f), and
// never Rutstein's own cost: it is not on the battlefield while it is
// being cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "439bf6d1-bd65-474b-a70c-7fdd446de4bb",
		Name:         "Honest Rutstein",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Creature spells you cast cost {1} less to cast.", effects.YourSpell(), effects.CreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Honest Rutstein — return target creature card from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())),
		},
	})
}
