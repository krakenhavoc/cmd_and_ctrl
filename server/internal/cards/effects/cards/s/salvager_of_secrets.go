package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvager of Secrets — Creature — Merfolk Wizard {3}{U}{U}, 2/2
// (Core Set 2019, #5527):
//
//	"When this creature enters, return target instant or sorcery card from your graveyard to your hand."
//
// Archaeomancer's text: the enters trigger returns an instant or
// sorcery card from your graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e400e13-6cae-45e6-b5cc-5b5f9d74dec5",
		Name:         "Salvager of Secrets",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets:   effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
			Key:       "Salvager of Secrets — return an instant or sorcery card to hand",
			Effect:    effects.ReturnTargetCardToHand,
		}},
	})
}
