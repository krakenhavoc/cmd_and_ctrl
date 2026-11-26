package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nemesis Trap — Instant — Trap {4}{B}{B} (Commander 2011, S82 older sets
// batch 453, #5574):
//
//	"If a white creature is attacking, you may pay {B}{B} rather than pay
//	 this spell's mana cost.
//	 Exile target attacking creature. Create a token that's a copy of that
//	 creature. Exile it at the beginning of the next end step."
//
// Inferno Trap's gated alternative cost. The copy is made only if the exile
// happened, from the creature as it last existed, and is exiled by a delayed
// trigger at the next end step (Mordor on the March's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c6f95fd-f8d0-4d22-aa21-29e5c70b2f85",
		Name:         "Nemesis Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1, CreatureTokens: 1},
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If a white creature is attacking, pay {B}{B}",
			ManaCost:  "{B}{B}",
			Condition: effects.S82OldB453AWhiteCreatureIsAttacking,
		}},
		OnResolve: effects.S82OldB453NemesisTrap,
	})
}
