package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syndicate Messenger — Creature — Bird {3}{W}, 2/3 (Ravnica: Clue Edition, #5929): "Flying. Afterlife 1."
//
// Afterlife is the dies trigger it reminds (CR 702.135a), Tithe Taker's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f5f4d78b-ed2b-471b-9ece-237fb55adc13",
		Name:            "Syndicate Messenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Syndicate Messenger — afterlife 1: create a 1/1 white and black Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white and black Spirit with flying"), N: 1})),
		},
	})
}
