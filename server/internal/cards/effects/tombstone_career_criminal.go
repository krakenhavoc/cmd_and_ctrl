package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tombstone, Career Criminal — Legendary Creature — Human Villain {2}{B},
// 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"When Tombstone enters, return target Villain card from your graveyard
//	 to your hand.
//	 Villain spells you cast cost {1} less to cast."
//
// The enter trigger is Auramancer's targeted graveyard-to-hand shape over a
// Villain card (any card type, so a Villain artifact counts); a graveyard
// with no Villain never puts it on the stack. The discount is Baron
// Strucker's board reduction on the controller's own Villain spells.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ad2cbb6-2291-420c-bfc8-ed3f079b59bf",
		Name:         "Tombstone, Career Criminal",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Villain spells you cast cost {1} less to cast.",
				YourSpell(), OtherSpellOfCreatureType("Villain")),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Tombstone — return target Villain card from your graveyard to your hand",
					returnTargetGraveyardCardToHand),
				TargetCardInGraveyard("target Villain card in your graveyard", YouOwn(), HasSubtype("Villain"))),
		},
	})
}
