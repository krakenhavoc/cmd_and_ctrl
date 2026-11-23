package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Far Traveler — Legendary Enchantment — Background {2}{W} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "At the beginning of your end step, exile
//	 up to one target tapped creature you control, then return it to the
//	 battlefield under its owner's control.""
//
// A granted trigger (ADR 0093), the commander creature's own: "your" is
// whoever controls it. The blink is Icewind Stalwart's (the card returns as a
// new object under its owner's control, CR 400.7, re-triggering what it has,
// and a token does not come back, CR 111.8) over "up to one" tapped creature
// the controller controls, the commander itself included, chosen as the
// trigger goes on the stack and still tapped as it resolves (CR 608.2b).
// Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const farTravelerGrant = "far-traveler/blink-tapped"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9432430-df01-497d-b025-7576dce830e4",
		Name:         "Far Traveler",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: farTravelerGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.AtYourEndStep("Far Traveler — exile up to one target tapped creature you control, then return it to the battlefield under its owner's control",
					effects.FlickerFirstLegalTargetUnderOwner),
					effects.TargetCreature("up to one target tapped creature you control", effects.YouControl(), effects.TappedPermanent()).WithCount(0, 1)),
			},
			Text: "At the beginning of your end step, exile up to one target tapped creature you control, then return it to the battlefield under its owner's control.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(farTravelerGrant)},
	})
}
