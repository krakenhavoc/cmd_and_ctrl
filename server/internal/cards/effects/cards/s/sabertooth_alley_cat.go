package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sabertooth Alley Cat — Creature — Cat {1}{R}{R}, 2/1 (Ravnica: City of
// Guilds, S82 OLD batch 574, #5600):
//
//	"This creature attacks each combat if able.
//	 {1}{R}: Creatures without defender can't block this creature this turn."
//
// The first line is the engine's attacks-each-combat requirement
// (CR 508.1d). The ability is "can't be blocked this turn except by creatures
// with defender" (CantBeBlockedThisTurnExceptBy): a creature may block the Cat
// only if it has defender, which is the same set as "creatures without
// defender can't block it". It is registered on the Cat as the ability
// resolves, for the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13d301d1-c1e1-4656-8700-3f1c1c7c4485",
		Name:         "Sabertooth Alley Cat",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: Creatures without defender can't block this creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.CantBeBlockedThisTurnExceptBy{
					Target:   item.SourceCardID,
					Keywords: []string{"defender"},
					Text:     "creatures with defender",
					Label:    "Sabertooth Alley Cat — creatures without defender can't block it",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
