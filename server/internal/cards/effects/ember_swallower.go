package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ember Swallower — 4/5 Elemental for {2}{R}{R}:
//
//	"{5}{R}{R}: Monstrosity 3. (If this creature isn't monstrous, put
//	 three +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, each player sacrifices
//	 three lands of their choice."
//
// "Each player" includes you. Each seat names three lands, one seat at
// a time in APNAP order, and every named land is sacrificed together
// once the last has chosen (PlayersSacrificeThenKeyedForEffect owes a seat
// one permanent per time it is named; CR 101.4, #3664). A player with
// fewer than three lands sacrifices what they have (CR 701.21a, "if
// you can"). Not targeted, so a hexproof land is as sacrificeable as
// any other.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0cdc86c4-8c68-4ba6-8165-511f32156a4c",
		Name:         "Ember Swallower",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Monstrosity(ManaCost("{5}{R}{R}"), 3)},
		Triggered: []game.TriggeredAbility{
			WhenBecomesMonstrous("Ember Swallower — each player sacrifices three lands", emberSwallowerEdict),
		},
	})
}

func emberSwallowerEdict(g *game.Game, item *game.StackItem) error {
	var asks []uuid.UUID
	if n := len(g.Seats); n > 0 {
		start := g.Turn.ActiveSeat
		for i := 0; i < n; i++ {
			p := g.Seats[(start+i)%n]
			if p == nil || p.Eliminated {
				continue
			}
			asks = append(asks, p.ID, p.ID, p.ID)
		}
	}
	g.PlayersSacrificeForEffect(item.SourceCardID, asks,
		sacrificeSpec("a land", Land()), "Sacrifice a land")
	return nil
}
