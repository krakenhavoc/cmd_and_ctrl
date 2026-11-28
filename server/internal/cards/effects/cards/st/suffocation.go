package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Suffocation — Instant {1}{U} (Alliances, #5634):
//
//	"Cast this spell only if you were dealt damage this turn by a red instant
//	 or sorcery spell.
//	 Suffocation deals 4 damage to the controller of the last red instant or
//	 sorcery spell that dealt damage to you this turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The per-turn record of noncombat damage (TurnTally.NoncombatDamage) keeps,
// for damage a spell dealt, the spell's colours, controller and card types as
// it dealt the damage (CR 608.2h), so the cast condition and the "last such
// spell" are read after the spell has left the stack, even a copy that has
// ceased to exist. The controller is read as the spell resolves (it can't
// change by then: the record is what the spell was). Suffocation resolves
// either way when its condition held on casting, so the delayed draw is
// scheduled whatever the damage does (CR 603.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57aeb106-dc4f-4069-b596-f997263dc0cf",
		Name:         "Suffocation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		CastCondition: func(g *game.Game, caster uuid.UUID, c game.Card) bool {
			_, ok := g.LastSpellDamageToPlayerThisTurnForEffect(caster, c.TextColor("R"), "Instant", "Sorcery")
			return ok
		},
		CastConditionLabel: "Cast this spell only if you were dealt damage this turn by a red instant or sorcery spell.",
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if player, ok := ctx.Game.LastSpellDamageToPlayerThisTurnForEffect(ctx.Controller(), ctx.TextColor("R"), "Instant", "Sorcery"); ok {
				if err := ctx.Game.DealDamageToPlayerForEffect(item.SourceCardID, player, 4); err != nil {
					return err
				}
			}
			return effects.DrawACardAtTheNextUpkeep("Suffocation — draw a card").Apply(ctx)
		},
	})
}
