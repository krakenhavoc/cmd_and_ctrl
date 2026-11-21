package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shellshock — {X}{R} Instant (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"For each opponent, choose up to one target creature that player
//	 controls. Shellshock deals X damage to each of those creatures. You
//	 create a Mutagen token for each creature dealt damage this way.
//	 (They're artifacts with "{1}, {T}, Sacrifice this token: Put a
//	 +1/+1 counter on target creature. Activate only as a sorcery.")"
//
// The targets are Age of Ultron's per-opponent clauses: each pick is
// bound to its opponent, so one that changed hands in response is
// illegal (CR 608.2b). Each still-legal pick is dealt X damage, and every
// one actually dealt damage makes one Mutagen token (MutagenToken, CR
// 111.10v) for the caster: X of 0, or damage prevented, makes none. With
// every target gone the spell does not resolve and makes nothing (the
// printed ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f406630b-d9f1-4fc5-b3f3-5327384b2901",
		Name:         "Shellshock",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.PerOpponentClauses(game.MaxPlayers-1, "creature", effects.Creature()),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageXToTarget(0), effects.DamageXToTarget(1), effects.DamageXToTarget(2)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return effects.DealDamageToEachThen(ctx, ids, ctx.X(), effects.ItemRecipient(item, effects.MutagenIfDealtDamage))
		},
	})
}
