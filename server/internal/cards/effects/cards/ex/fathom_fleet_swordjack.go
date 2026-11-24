package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fathom Fleet Swordjack — {3}{R} Creature — Orc Pirate 4/3, Commander Legends (#5507):
//
//	"Whenever this creature attacks, it deals damage to the player or
//	 planeswalker it's attacking equal to the number of artifacts you control.
//	 Encore {5}{R}"
//
// The attack trigger counts your artifacts as it resolves and hits the player
// the attack was declared against (the engine has no planeswalker defenders, so
// the attacked object is always the player). Encore is the shared constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77828ebf-58d2-4aca-ae6f-0efa50d5d640",
		Name:         "Fathom Fleet Swordjack",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Encore("{5}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Fathom Fleet Swordjack — damage to the attacked player equal to the number of artifacts you control",
				effects.S82OldB082DamageAttackedPlayerPerArtifact),
		},
	})
}
