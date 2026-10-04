package aiseat_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ring_bearer_test.go — ADR 0114 §7: a bot seat answers "choose your
// Ring-bearer". The prompt goes through the real enumerator, the real
// heuristic and the real resolver (driveChoices), so an answer the
// engine would refuse fails here as #544 would. Not behind
// AISEAT_GAME_TESTS: one prompt, milliseconds.

// ringCreature puts a public creature onto the battlefield under
// `owner`, able to attack this turn.
func ringCreature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	}
	c.AddKnowersAll(seatIDs(g))
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func ringBearer(g *game.Game, seat uuid.UUID) uuid.UUID {
	var id uuid.UUID
	g.ReadSnapshot(func() { id = game.RingBearerOf(g, seat) })
	return id
}

func temptForBot(t *testing.T, g *game.Game, seat uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.RingTemptsForEffect(seat, uuid.Nil, nil); err != nil {
			t.Fatalf("RingTemptsForEffect: %v", err)
		}
	})
	if !owesChoice(g, seat) {
		t.Fatal("setup: no ring_bearer prompt")
	}
}

// The heuristic gives the Ring to the hardest hitter, and the engine
// accepts the answer it was offered.
func TestHeuristicSeatChoosesTheHardestHitterAsRingBearer(t *testing.T) {
	g := newSettledTable(t, 2076)
	me := g.Seats[0]
	ringCreature(g, me.ID, "Elvish Mystic", "Creature — Elf Druid", 1, 1)
	big := ringCreature(g, me.ID, "Craw Wurm", "Creature — Wurm", 6, 4)
	ringCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	temptForBot(t, g, me.ID)

	taken := driveChoices(t, g, heuristic.New(), me.ID, 4)
	if got := ringBearer(g, me.ID); got != big {
		t.Fatalf("the bot chose %v (took %v), want the 6-power Craw Wurm", got, taken)
	}
}

// It avoids the creature the Ring's "is legendary" would put into the
// legend rule against a same-named legend the seat already has.
func TestHeuristicSeatKeepsTheRingOffALegendRuleVictim(t *testing.T) {
	g := newSettledTable(t, 2077)
	me := g.Seats[0]
	ringCreature(g, me.ID, "Gollum", "Legendary Creature — Halfling Horror", 1, 1)
	twin := ringCreature(g, me.ID, "Gollum", "Creature — Halfling Horror", 5, 5)
	bear := ringCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	temptForBot(t, g, me.ID)

	driveChoices(t, g, heuristic.New(), me.ID, 4)
	switch got := ringBearer(g, me.ID); got {
	case twin:
		t.Fatal("the bot made a second Gollum legendary and lost one to the legend rule")
	case bear:
	default:
		t.Fatalf("the bot chose %v, want the Bears: the hardest hitter that is not a legend-rule victim", got)
	}
}

// Options.OrderRingBearer: on a board wider than the enumerator's cap,
// the best Ring-bearer survives it only because the heuristic ordered
// the pool. Without the hook the battlefield order spends the cap.
func TestRingBearerOrderKeepsTheBestCandidateInsideTheCap(t *testing.T) {
	g := newSettledTable(t, 2078)
	me := g.Seats[0]
	for i := 0; i < 14; i++ {
		ringCreature(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", 1, 1)
	}
	best := ringCreature(g, me.ID, "Craw Wurm", "Creature — Wurm", 6, 4)
	temptForBot(t, g, me.ID)

	offers := func(opts legal.Options) bool {
		for _, m := range legal.EnumerateForWithOptions(g, me.ID, opts) {
			var p struct {
				CardIDs []string `json:"card_ids"`
			}
			if json.Unmarshal(m.Params, &p) == nil && len(p.CardIDs) == 1 && p.CardIDs[0] == best.String() {
				return true
			}
		}
		return false
	}
	if offers(legal.Options{}) {
		t.Skip("the cap no longer bites on this board; widen it")
	}
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID}
	if !offers(legal.Options{OrderRingBearer: heuristic.New().RingBearerOrder(in)}) {
		t.Fatal("with the heuristic's order the best Ring-bearer is still cut by the cap")
	}
}
