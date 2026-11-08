package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "4b803bf2-025a-4bf6-9d3c-a28d714b4781",
		Name:          "Kellan's Lightblades",
		Completeness:  CompletenessFull,
		Targets:       TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		Purpose:       game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if g2WasBargained(ctx) {
				return DestroyTarget{Target: id}.Apply(ctx)
			}
			return DealDamage{Source: item.SourceCardID, Target: id, Amount: 3}.Apply(ctx)
		},
	})
}
