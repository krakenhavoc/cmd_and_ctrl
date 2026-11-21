package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruric Thar, Biomagus — Legendary Creature — Ogre Crab Wizard
// {4}{U}{U}, 4/6:
//
//	"Flying
//	 Prowess, prowess
//	 Whenever Ruric Thar becomes the target of a spell or ability an
//	 opponent controls, draw a card."
//
// Prowess printed twice is two instances (CR 702.108b): the keyword list
// below names it twice and the engine keeps both, so one noncreature
// spell puts two triggers on the stack. EventBecomesTarget fires once per
// target slot at announce, with the targeting player in Actor; a trigger
// naming it twice in one spell is two draws, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "873fa095-5662-407a-bd24-9a1c4712a428",
		Name:            "Ruric Thar, Biomagus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordProwess, game.KeywordProwess},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor != uuid.Nil && ev.Actor != source.Controller
			}, "Ruric Thar, Biomagus — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
