package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kellan's Lightblades — Instant {1}{W} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Kellan's Lightblades deals 3 damage to target attacking or blocking
//	 creature. If this spell was bargained, destroy that creature
//	 instead."
//
// The target clause is Cosmium Blast's (AttackingOrBlocking, read off the
// combat designations at cast and again as it resolves, CR 608.2b).
// Bargain is the shared optional cost (g2Bargain), read back at
// resolution as Candy Grapple reads it: a bargained cast destroys in
// place of the damage, never both. The declared purpose is the
// unbargained 3 damage, the least the spell does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "4b803bf2-025a-4bf6-9d3c-a28d714b4781",
		Name:          "Kellan's Lightblades",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:       game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{effects.G2Bargain()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if effects.G2WasBargained(ctx) {
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			}
			return effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 3}.Apply(ctx)
		},
	})
}
