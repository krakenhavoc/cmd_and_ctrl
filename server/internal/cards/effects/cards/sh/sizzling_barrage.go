package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sizzling Barrage — Instant {1}{R} (Zendikar Rising, S82 OLD #5508):
//
//	"Sizzling Barrage deals 4 damage to target creature that blocked this
//	 turn."
//
// "Blocked this turn" is read from the turn's block events (Demonic Stench's
// reading): a creature that blocked stays a legal target after it has been
// removed from combat, and one that has left the battlefield is not there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26a55f91-91a1-4a1e-86fe-6551fd64ea87",
		Name:         "Sizzling Barrage",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		Targets: effects.TargetCreature("target creature that blocked this turn",
			func(g *game.Game, _ uuid.UUID, c game.Card) bool {
				return effects.GapB213CreaturesThatBlockedThisTurn(g)[c.InstanceID]
			}),
		OnResolve: effects.DamageToFirstTargetStep(4),
	})
}
