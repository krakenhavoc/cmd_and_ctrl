package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "f406630b-d9f1-4fc5-b3f3-5327384b2901",
		Name:         "Shellshock",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      perOpponentClauses(game.MaxPlayers-1, "creature", Creature()),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageXToTarget(0), DamageXToTarget(1), DamageXToTarget(2)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return DealDamageToEachThen(ctx, ids, ctx.X(), itemRecipient(item, mutagenIfDealtDamage))
		},
	})
}
