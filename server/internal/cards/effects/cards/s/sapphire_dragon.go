package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sapphire Dragon // Psionic Pulse — an Adventure creature (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	Sapphire Dragon {5}{U}{U}, Creature — Dragon, 5/6
//	  "Flying
//	   Whenever this creature attacks or blocks, scry 2."
//	Psionic Pulse {2}{U}, Instant — Adventure
//	  "Counter target noncreature spell."
//
// Morkrut Necropod's attacks-or-blocks trigger (one trigger per declaration,
// SelfBlocksOnce for the block side) with a scry 2; Psionic Pulse is Negate.
//
// No simplification.
func init() {
	const oracleID = "53015c41-af2d-43e2-a690-d8877537b8dd"
	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Sapphire Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 2},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclared(ev, source) || effects.SelfBlocksOnce(ev, source)
				},
				"Sapphire Dragon — scry 2",
				effects.Do(effects.Scry{N: 2})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Psionic Pulse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
