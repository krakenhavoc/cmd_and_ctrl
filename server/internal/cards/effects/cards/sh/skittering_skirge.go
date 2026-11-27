package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skittering Skirge — Creature — Phyrexian Imp {B}{B} 3/2 (Urza's Saga, #5624):
//
//	"Flying
//	 When you cast a creature spell, sacrifice this creature."
//
// The trigger is on the battlefield only, so casting the Skirge itself does
// nothing. A Skirge that has left the battlefield by resolution is not
// sacrificed (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fae1d5d9-a50e-423e-8692-3e41f39df081",
		Name:            "Skittering Skirge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Creature(), "Skittering Skirge — sacrifice this creature", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
