package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sovereign Okinec Ahau — Legendary Creature — Cat Noble {2}{G}{W}, 3/4
// (The Lost Caverns of Ixalan, #4470):
//
//	"Ward {2}
//	 Whenever Sovereign Okinec Ahau attacks, for each creature you control
//	 with power greater than that creature's base power, put a number of
//	 +1/+1 counters on that creature equal to the difference."
//
// Ward {2} is the shared ward trigger (CR 702.21a). Base power is Kutzil,
// Malamet Exemplar's comparison (CR 208.4b): Card.BasePower, the power
// after layer 7b, against PowerForComparison, which adds pumps, anthems
// and counters without clamping. Every difference is read as the trigger
// resolves, before any counter is placed, so the counters one creature gets
// never change what another gets; then each creature gets its own number of
// +1/+1 counters through the counter window (a Doubling Season doubles
// them).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "087b2490-e415-44e2-a674-3b5619a0fad5",
		Name:         "Sovereign Okinec Ahau",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Sovereign Okinec Ahau — ward {2}"),
			WheneverThisAttacks("Sovereign Okinec Ahau — +1/+1 counters equal to each creature's power above its base power",
				sovereignOkinecAhauCounters),
		},
	})
}

// sovereignOkinecAhauCounters is the attack trigger's body: every
// difference first, then the counters.
func sovereignOkinecAhauCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	type grow struct {
		id uuid.UUID
		n  int
	}
	var todo []grow
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		if diff := c.PowerForComparison() - c.BasePower(); diff > 0 {
			todo = append(todo, grow{c.InstanceID, diff})
		}
	}
	for _, t := range todo {
		if err := (AddCounter{Target: t.id, Kind: game.CounterPlusOne, N: t.n}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
