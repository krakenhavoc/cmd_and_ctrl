package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Solarium Sentry — Creature — Cat Soldier {G}{W}, 3/3:
//
//	"Whenever an opponent casts a spell with mana value 2 or less, you
//	 gain 2 life."
//
// The mana value is the spell's as it was cast (an X spell counts X as
// 0, CR 202.3e), read through ManaValueLE.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "afddba56-f9a0-4917-8167-c9fd17ca9954",
		Name:         "Solarium Sentry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.ManaValueLE(2)),
				"Solarium Sentry — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
