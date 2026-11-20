package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6ad2cbb6-2291-420c-bfc8-ed3f079b59bf",
		Name:         "Tombstone, Career Criminal",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Villain spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.OtherSpellOfCreatureType("Villain")),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Tombstone — return target Villain card from your graveyard to your hand",
					effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target Villain card in your graveyard", effects.YouOwn(), effects.HasSubtype("Villain"))),
		},
	})
}
