package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Regrower — Creature — Elf Druid {2}{G}{G}, 4/3 (Foundations,
// #4465):
//
//	"When this creature enters, return target permanent card from your
//	 graveyard to your hand."
//
// Hologram Stalker's entry trigger. A permanent card is an artifact,
// battle, creature, enchantment, land or planeswalker card (CR 110.4b);
// an instant or sorcery card isn't one. The card returns if it is still
// in your graveyard as the trigger resolves (CR 608.2b), and with no
// permanent card there the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "adac526d-1a85-431a-a302-0a4b8f4a0c44",
		Name:         "Elvish Regrower",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Elvish Regrower — return target permanent card from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target permanent card from your graveyard", effects.Permanent(), effects.YouOwn())),
		},
	})
}
