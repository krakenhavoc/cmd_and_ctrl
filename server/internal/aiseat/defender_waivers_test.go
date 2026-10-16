package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// defender_waivers_test.go — the bot half of ADR 0106's 2026-10-16
// amendment (#1936). A bot's Weathered Sentinels attacks only a player who
// attacked the bot during their last turn, even with a far more tempting
// opponent at 5 life and nothing to block with, and the combat ends. The
// bot's world is the enumerator, so this pins that no policy finds a way
// round it.

const botSentinelsOracle = "0903b941-82f1-4d3e-9717-29ab7d0c8cef"

func TestBotsAttackWithWeatheredSentinelsOnlyWhereItMay(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			s := g.Turn.ActiveSeat
			active, hitMe, tempting := g.Seats[s], g.Seats[(s+1)%4], g.Seats[(s+2)%4]
			kept := g.Battlefield.Cards[:0]
			for _, c := range g.Battlefield.Cards {
				if c.Controller != tempting.ID {
					kept = append(kept, c)
				}
			}
			g.Battlefield.Cards = kept
			tempting.Life = 5
			hitMe.LastTurnAttacks = []game.AttackRecord{{Attacker: uuid.New(), Defender: active.ID}}

			sentinels := limitPush(g, active, "Weathered Sentinels", "Artifact Creature — Wall", botSentinelsOracle, 2, 5)
			g.WithWriteLock(func() {
				g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: sentinels, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; c.InstanceID == sentinels {
						c.SummonedThisTurn = false
					}
				}
			})

			attackedAt := map[uuid.UUID]bool{}
			driveOneCombatWatching(t, g, pol, func() {
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; c.InstanceID == sentinels && c.AttackingTarget != uuid.Nil {
						attackedAt[c.AttackingTarget] = true
					}
				}
			})
			for target := range attackedAt {
				if target != hitMe.ID {
					t.Errorf("the Sentinels attacked %s, who did not attack the bot: %v", target, attackedAt)
				}
			}
			if name == "aggressive" && !attackedAt[hitMe.ID] {
				t.Error("the aggressive policy was offered the attack at the player who attacked it and did not take it")
			}
		})
	}
}
