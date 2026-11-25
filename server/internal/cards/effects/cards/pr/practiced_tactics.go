package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Practiced Tactics — Instant {W} (Zendikar Rising, S82 OLD #5508):
//
//	"Choose target attacking or blocking creature. Practiced Tactics deals
//	 damage to that creature equal to twice the number of creatures in your
//	 party. (Your party consists of up to one each of Cleric, Rogue,
//	 Warrior, and Wizard.)"
//
// The party is counted as the spell resolves (CR 700.8a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "612c9079-adae-4fe5-9a72-8efc09ddfe5c",
		Name:         "Practiced Tactics",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 2 * effects.FinB9PartySize(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
