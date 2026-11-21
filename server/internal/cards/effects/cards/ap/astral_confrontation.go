package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Astral Confrontation — Instant {4}{W} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"This spell costs {1} less to cast for each opponent you're attacking.
//	 Exile target creature."
//
// An opponent is being attacked when one of the caster's attacking creatures
// is attacking them, or a planeswalker or battle they control (CR 506.2; the
// defending player the engine names for the attacker), counted once however
// many creatures attack them. Read as the spell is cast, so before attackers
// nothing is reduced. The exile is Swords to Plowshares' without the life.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "80788ac7-19d6-42d2-9f6e-676d7b5c912e",
		Name:         "Astral Confrontation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(opponentsYouAreAttacking, "This spell costs {1} less to cast for each opponent you're attacking."),
		},
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return effects.ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}

// opponentsYouAreAttacking counts the distinct opponents of the caster that
// one of the caster's attackers is attacking.
func opponentsYouAreAttacking(q game.CostQuery) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	seen := map[uuid.UUID]bool{}
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller != q.Controller || !c.IsCreature() || c.AttackingTarget == uuid.Nil {
			continue
		}
		if p := q.Game.DefendingPlayerForAttackForEffect(c.AttackingTarget); p != uuid.Nil && p != q.Controller {
			seen[p] = true
		}
	}
	return len(seen)
}
