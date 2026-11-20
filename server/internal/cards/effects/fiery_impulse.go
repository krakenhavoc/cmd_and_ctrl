package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiery Impulse — Instant {R} (Magic Origins, #5549):
//
//	"Fiery Impulse deals 2 damage to target creature.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Fiery Impulse deals 3 damage instead."
//
// Spell mastery is read at resolution (spellMastery, Animist's
// Awakening's reader), while the spell is still on the stack, so it does
// not count itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27c84ad1-1491-459b-bf8a-0146062084dc",
		Name:         "Fiery Impulse",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 2
			if spellMastery(ctx.Game, ctx.Controller()) {
				n = 3
			}
			for _, t := range ctx.LegalTargets() {
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
