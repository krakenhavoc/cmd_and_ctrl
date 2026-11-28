package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Sentry — Creature — Goblin Warrior {R} 1/1 (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"Whenever an opponent casts a spell, this creature gets +2/+2 until end of
//	 turn."
//
// Bloodvial Purveyor's opponent-casts trigger with a self pump. Each spell is
// its own trigger, so two spells make it +4/+4.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac6f490b-6cf7-4afc-910e-4c8b51fcb2de",
		Name:         "Mogg Sentry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B15OpponentCastSpell(ev, source)
			},
			Key:    "Mogg Sentry — it gets +2/+2 until end of turn",
			Effect: effects.ThisGetsUntilEndOfTurn(2, 2, "Mogg Sentry — +2/+2 until end of turn"),
		}},
	})
}
