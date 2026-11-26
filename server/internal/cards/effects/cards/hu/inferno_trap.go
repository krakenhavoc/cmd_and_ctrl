package cardshu

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inferno Trap — Instant — Trap {3}{R} (Duel Decks: Speed vs. Cunning, #5555):
//
//	"If you've been dealt damage by two or more creatures this turn, you may
//	 pay {R} rather than pay this spell's mana cost.
//	 Inferno Trap deals 4 damage to target creature."
//
// The trap clause is an alternative cost (CR 118.9) offered only once two
// different creature objects have dealt damage to the caster this turn
// (damage of any kind; a creature that has since died still counts, and one
// that left and came back is a new object, CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83a71000-942b-4472-9521-0e1070a3343b",
		Name:         "Inferno Trap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.DamageToTarget(0, 4)),
		},
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If you've been dealt damage by two or more creatures this turn, pay {R}",
			ManaCost:  "{R}",
			Condition: dealtDamageByTwoOrMoreCreaturesThisTurn,
		}},
		OnResolve: effects.DamageToFirstTargetStep(4),
	})
}

// dealtDamageByTwoOrMoreCreaturesThisTurn is the trap's condition.
func dealtDamageByTwoOrMoreCreaturesThisTurn(g *game.Game, controller uuid.UUID) bool {
	seen := map[uuid.UUID]bool{}
	for _, ref := range g.CreaturesThatDealtDamageToThisTurn(controller) {
		seen[ref.ID] = true
	}
	return len(seen) >= 2
}
