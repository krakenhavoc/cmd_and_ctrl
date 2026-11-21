package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Waverider — Creature — Otter Wizard {4}{U}{U}, 4/4 (Bloomburrow,
// #4467):
//
//	"When this creature enters, you may cast target instant or sorcery
//	 card with mana value 4 or less from your graveyard without paying
//	 its mana cost. If that spell would be put into your graveyard, exile
//	 it instead."
//
// Victor Timely's entry trigger over instant and sorcery cards only:
// Torrential Gearhulk's free cast from the graveyard
// (freeCastTargetFromYourGraveyardExilingIt), offered as the trigger
// resolves (CR 608.2g), with the spell exiled whether it resolves, is
// countered or fizzles. The mana value is the card's in the graveyard (X
// is 0 there, CR 202.3e). Declining leaves the card where it is.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Daring Waverider — you may cast target instant or sorcery card with mana value 4 or less from your graveyard without paying its mana cost",
		effects.FreeCastTargetFromYourGraveyardExilingIt("Daring Waverider"))
	enters.Targets = effects.TargetCardInGraveyard("target instant or sorcery card with mana value 4 or less from your graveyard",
		effects.Or(effects.Instant(), effects.Sorcery()), effects.ManaValueLE(4), effects.YouOwn())
	effects.Register(effects.Spec{
		OracleID:     "e4d52559-7624-48b2-96ba-e52ada6c507a",
		Name:         "Daring Waverider",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
