package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattered Angel — Creature — Phyrexian Angel {3}{W}{W}, 3/3:
//
//	"Flying
//	 Whenever a land an opponent controls enters, you may gain 3 life."
//
// One trigger per land (the clause says "a land"), each with its own
// optional prompt. Lands that enter under an opponent's control by any
// route count, tokens included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e339c6cb-c75d-4dd4-8c1a-78f844983242",
		Name:            "Shattered Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B43ALandAnOpponentControlsEntered(ev, source, g)
				return ok
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Shattered Angel — gain 3 life?"},
			Key:            "Shattered Angel — you may gain 3 life",
			Effect:         effects.Do(effects.GainLife{Amount: 3}),
		}},
	})
}
