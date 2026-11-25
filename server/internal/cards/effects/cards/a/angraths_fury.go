package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angrath's Fury — Sorcery {3}{B}{R}:
//
//	"Destroy target creature. Angrath's Fury deals 3 damage to target player
//	 or planeswalker. You may search your library and/or graveyard for a card
//	 named Angrath, Minotaur Pirate, reveal it, and put it into your hand. If
//	 you search your library this way, shuffle."
//
// Two target clauses, read by slot as the spell resolves (CR 608.2b): a clause
// whose target is gone is skipped and the rest still happens. The search is
// Liliana's Scorn's named fetch (S82OldB105MayFetchNamed).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61f3a549-77b4-4403-afa2-a0f02603f2e6",
		Name:         "Angrath's Fury",
		Completeness: effects.CompletenessFull,
		Purpose:      s82OldB216FuryPurpose(),
		Targets:      effects.TargetCreature("target creature").Then(effects.TargetPlayerOrPlaneswalker()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID}, game.KeyedTail{})
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(1); ok {
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(effects.S82OldB105MayFetchNamed("Angrath's Fury", "Angrath, Minotaur Pirate")),
		),
	})
}

// s82OldB216FuryPurpose is the removal, the 3 damage and the named fetch.
func s82OldB216FuryPurpose() game.Purpose {
	p := effects.ForTargets(effects.DamageToTarget(1, 3))
	p.Answers = game.AnswerRemove
	p.Tutors = 1
	return p
}
