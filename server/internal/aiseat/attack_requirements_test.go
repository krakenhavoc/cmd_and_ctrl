package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
)

// attack_requirements_test.go — the bot half of #1571: a goaded
// creature on a bot's side of a four-seat table attacks, attacks a
// player other than the goader, and the combat ends — whatever the
// policy wants. driveOneCombat (combat_limits_test.go) fails on a
// refused enumerated move, on a priority holder that declines, and on
// a combat that never ends, which are the #544 wedges a withheld pass
// could cause.
func TestBotsObeyGoadInFourSeats(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			active := g.Seats[g.Turn.ActiveSeat]
			goader := g.Seats[(g.Turn.ActiveSeat+1)%4]
			// A 1/1 into three untapped 2/2 boards: a creature the
			// heuristic would rather keep home.
			goaded := limitPush(g, active, "Goaded Ogre", "Creature — Ogre", "", 1, 1)
			if err := g.SetGoaded(goaded, goader.ID); err != nil {
				t.Fatal(err)
			}
			attackedAt := map[uuid.UUID]bool{}
			driveOneCombatWatching(t, g, pol, func() {
				for i := range g.Battlefield.Cards {
					c := &g.Battlefield.Cards[i]
					if c.InstanceID == goaded && c.AttackingTarget != uuid.Nil {
						attackedAt[c.AttackingTarget] = true
					}
				}
			})
			if len(attackedAt) == 0 {
				t.Fatal("the goaded creature never attacked")
			}
			if attackedAt[goader.ID] {
				t.Error("the goaded creature attacked its goader while other opponents were open")
			}
		})
	}
}
