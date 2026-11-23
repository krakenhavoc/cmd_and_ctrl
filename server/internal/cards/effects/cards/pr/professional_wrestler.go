package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Professional Wrestler — Creature — Human Warrior Performer {3}{G}, 4/4
// (Marvel's Spider-Man):
//
//	"When this creature enters, create a Treasure token.
//	 This creature can't be blocked by more than one creature."
//
// The block limit is MaxBlockers on itself (Outland Colossus); the
// Treasure is the shared TreasureToken template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3dd92eea-02ca-4e55-8f69-8c8059dd7ee6",
		Name:         "Professional Wrestler",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Professional Wrestler — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
