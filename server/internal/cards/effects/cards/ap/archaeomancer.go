package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archaeomancer — Creature — Human Wizard, {2}{U}{U}, 1/2
// (EDHREC rank 959):
//
//	"When this creature enters, return target instant or sorcery card
//	 from your graveyard to your hand."
//
// The blue Eternal Witness, and the flicker deck's engine — every
// blink is a spell back. One targeted ETB trigger over the graveyard
// clause; mandatory, so with an instant or sorcery in the graveyard
// the controller picks one, and with none the trigger is removed
// without a prompt (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a91a3266-cadd-47a0-9b20-160307f14c07",
		Name:         "Archaeomancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets:   effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
			Key:       "Archaeomancer — return an instant or sorcery card to hand",
			Effect:    effects.ReturnTargetCardToHand,
		}},
	})
}
