package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "439bf6d1-bd65-474b-a70c-7fdd446de4bb",
		Name:         "Honest Rutstein",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Creature spells you cast cost {1} less to cast.", YourSpell(), CreatureSpell()),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Honest Rutstein — return target creature card from your graveyard to your hand",
				returnTargetGraveyardCardToHand),
				TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())),
		},
	})
}
