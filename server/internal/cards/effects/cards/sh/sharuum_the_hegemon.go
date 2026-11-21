package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sharuum the Hegemon — Legendary Artifact Creature — Sphinx {3}{W}{U}{B},
// 5/5:
//
//	"Flying
//	 When Sharuum enters, you may return target artifact card from your
//	 graveyard to the battlefield."
//
// Sun Titan's optional graveyard-to-battlefield trigger on the enter
// event alone, narrowed to artifact cards you own. It returns under your
// control (your graveyard) and untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "037e7fc9-3aa6-484c-a2c8-43009e45f1d8",
		Name:            "Sharuum the Hegemon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: effects.TargetCardInGraveyard("target artifact card in your graveyard", effects.YouOwn(), effects.Artifact()),
			Key:     "Sharuum the Hegemon — return an artifact card to the battlefield",
			Effect:  effects.HobB8ReturnTargetGraveyardCardToBattlefield,
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Sharuum the Hegemon — return an artifact card from your graveyard to the battlefield?",
			},
		}},
	})
}
