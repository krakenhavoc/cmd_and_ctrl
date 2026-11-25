package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Briarblade Adept — {4}{B} Creature — Elf Assassin 3/4, Commander Legends (#5507):
//
//	"Whenever this creature attacks, target creature an opponent controls gets
//	 -1/-1 until end of turn.
//	 Encore {3}{B}"
//
// A targeted attack trigger (CR 603.3d) and the shared Encore constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6680a5f7-3ac1-4550-a29a-9553ad7ae465",
		Name:         "Briarblade Adept",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Encore("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventAttack, effects.Self, "Briarblade Adept — target creature an opponent controls gets -1/-1 until end of turn",
				effects.BoostTargetUntilEOT("Briarblade Adept — -1/-1 until end of turn", -1, -1)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
