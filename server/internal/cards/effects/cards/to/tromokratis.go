package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tromokratis — Legendary Creature — Kraken {5}{U}{U}, 8/8 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"Tromokratis has hexproof unless it's attacking or blocking.
//	 Tromokratis can't be blocked unless all creatures defending player
//	 controls block it. (If any creature that player controls doesn't block
//	 this creature, it can't be blocked.)"
//
// Hexproof is Drownyard Behemoth's conditional keyword grant, read live in
// layer 6: on unless Tromokratis is an attacking creature or is blocking
// something. The block clause is a minimum-blockers rule (Rampaging Ceratops'
// MinBlockers, #750) whose minimum is the number of creatures the defending
// player controls, tapped or not (CR 509.1b): a defender's block that leaves
// any creature out is refused whole, and a defender with a creature that
// can't block can't block it at all, as the reminder text says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9aa2f27b-4c00-44c1-8bc4-014350f05b6d",
		Name:         "Tromokratis",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{tromokratisHexproof()},
		BlockRules: []game.BlockRule{{
			Count: func(g *game.Game, attacker, source *game.Card) (int, int) {
				if source == nil || attacker.InstanceID != source.InstanceID {
					return 0, 0
				}
				defender := g.DefendingPlayerForAttackerForEffect(attacker.InstanceID)
				if defender == uuid.Nil {
					return 0, 0
				}
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsCreature() && c.Controller == defender {
						n++
					}
				}
				return n, 0
			},
		}},
	})
}

// tromokratisIsAttackingOrBlocking is "it's attacking or blocking": an
// attacking creature, or one that blocks any attacker (CR 506.4).
func tromokratisIsAttackingOrBlocking(c *game.Card) bool {
	return c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil
}

// tromokratisHexproof is the conditional keyword grant. Attacking and blocking
// are not on the battlefield's shape, so the layers are told to re-read them
// when a creature is declared as an attacker (DependsOnAttackingStatus, which
// also covers leaving combat) or a blocker (DependsOnEvents: EventBlock).
func tromokratisHexproof() game.StaticAbility {
	ab := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && !tromokratisIsAttackingOrBlocking(target)
	}, "hexproof")
	ab.Label = "Tromokratis has hexproof unless it's attacking or blocking."
	ab.DependsOnAttackingStatus = true
	ab.DependsOnEvents = []game.EventKind{game.EventBlock}
	return ab
}
