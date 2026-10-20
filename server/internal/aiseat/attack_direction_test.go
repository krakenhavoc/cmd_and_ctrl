package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_direction_test.go — the bot half of ADR 0206 (#4197). Under
// "each player may attack only the nearest opponent in the chosen
// direction" a bot attacks only the player to that side, even when
// another opponent is the obvious target (5 life, nothing to block
// with), and the combat ends. The bot's world is the enumerator, which
// never offers anyone else; this pins that no policy finds a way round.
func TestBotsAttackOnlyTheNearestOpponentInTheChosenDirection(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			s := g.Turn.ActiveSeat
			active, left, right := g.Seats[s], g.Seats[(s+1)%4], g.Seats[(s+3)%4]
			kept := g.Battlefield.Cards[:0]
			for _, c := range g.Battlefield.Cards {
				if c.Controller != right.ID {
					kept = append(kept, c)
				}
			}
			g.Battlefield.Cards = kept
			right.Life = 5
			g.WithWriteLock(func() {
				g.GrantAttackDirectionForEffect(game.AttackDirectionLeft, "Test Tactician", uuid.Nil, g.UntilYourNextTurnDuration(active.ID))
			})

			tally := driveOneCombat(t, g, pol)
			for defender, n := range tally.perDefender {
				if defender != left.ID {
					t.Errorf("%d attack(s) at %s, which is not the player to the left", n, defender)
				}
			}
			if name == "aggressive" && tally.attackers == 0 {
				t.Error("the aggressive policy never attacked the player to the left")
			}
		})
	}
}
