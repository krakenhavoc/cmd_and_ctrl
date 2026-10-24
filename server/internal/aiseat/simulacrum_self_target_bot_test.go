package aiseat_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// simulacrum_self_target_bot_test.go — #4314, whole-engine: the real
// catalog Simulacrum, the real enumerator, the real view and the real
// heuristic.
//
// Simulacrum "deals damage to target creature you control". The clause is
// declared Harmful, the view carries it on the card's legal_targets, and
// the heuristic prices a pick of its own permanent by what losing it
// costs. Before, targetsValue gave every creature the bot controls the
// same flat OwnPermanentTarget, so the bot aimed the damage at whichever
// creature the enumerator listed first, its best one included.
//
// Not behind AISEAT_GAME_TESTS: one decision, milliseconds.

const botSimulacrumOracle = "20d69989-7250-40c7-a064-8ed78ccbe556"

// simulacrumTable parks a two-seat game on the active seat's first main
// phase with two Swamps, a Simulacrum in hand and two creatures, a 5/5 and
// a 1/1, pushed in the given order so a pick that falls back on "first
// listed" is caught both ways round.
func simulacrumTable(t *testing.T, seed uint64, bestFirst bool) (g *game.Game, me *game.Player, best, least uuid.UUID) {
	t.Helper()
	g = newSettledTable(t, seed)
	advanceToStep(t, g, game.StepPrecombatMain)
	me = g.Seats[g.Turn.ActiveSeat]
	everyone := seatIDs(g)
	mk := func(name string, p, tough int) game.Card {
		c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
			ManaCost: "{2}{G}", Power: p, Toughness: tough, Owner: me.ID, Controller: me.ID}
		c.AddKnowersAll(everyone)
		return c
	}
	big, small := mk("Test Titan", 5, 5), mk("Test Mouse", 1, 1)
	best, least = big.InstanceID, small.InstanceID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Command.Cards = nil
		if bestFirst {
			g.Battlefield.PushTop(big)
			g.Battlefield.PushTop(small)
		} else {
			g.Battlefield.PushTop(small)
			g.Battlefield.PushTop(big)
		}
		for i := 0; i < 2; i++ {
			s := game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
				Owner: me.ID, Controller: me.ID}
			s.AddKnowersAll(everyone)
			g.Battlefield.PushTop(s)
		}
		sim := game.Card{InstanceID: uuid.New(), Name: "Simulacrum", OracleID: botSimulacrumOracle,
			TypeLine: "Instant", ManaCost: "{1}{B}", Owner: me.ID, Controller: me.ID}
		sim.AddKnowersAll(everyone)
		me.Hand.Cards = append(me.Hand.Cards, sim)
	})
	return g, me, best, least
}

// TestHeuristicAimsSimulacrumAtItsLeastCreature: given only the casts of
// Simulacrum, one per legal creature, the heuristic picks the one aimed at
// the lesser creature, whichever the enumerator lists first.
func TestHeuristicAimsSimulacrumAtItsLeastCreature(t *testing.T) {
	for _, bestFirst := range []bool{true, false} {
		name := "mouse listed first"
		if bestFirst {
			name = "titan listed first"
		}
		t.Run(name, func(t *testing.T) {
			g, me, best, least := simulacrumTable(t, 4314, bestFirst)
			var casts []legal.Move
			for _, m := range legal.EnumerateFor(g, me.ID) {
				if m.Kind == legal.KindCast && strings.Contains(m.Label, "Simulacrum") {
					casts = append(casts, m)
				}
			}
			aimed := map[uuid.UUID]bool{}
			for _, m := range casts {
				for _, id := range targetIDs(t, m) {
					aimed[id] = true
				}
			}
			if !aimed[best] || !aimed[least] {
				t.Fatalf("the enumerator should offer Simulacrum at both creatures; got %d casts aimed at %v", len(casts), aimed)
			}
			// Rank, not Decide: the bot may decline the cast outright (a
			// self-inflicted loss rarely beats keeping the card), and the
			// question here is only which target it prefers.
			ranked := heuristic.New().Rank(context.Background(), aiseat.Input{
				View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: casts,
			})
			if len(ranked) != len(casts) {
				t.Fatalf("ranked %d of %d casts", len(ranked), len(casts))
			}
			d := ranked[0]
			picked := targetIDs(t, casts[d.Index])
			if len(picked) != 1 || picked[0] != least {
				t.Errorf("Simulacrum aimed at %v, want the 1/1 %s (the 5/5 is %s): %s", picked, least, best, d.Reason)
			}
		})
	}
}

const botPlagueBelcherOracle = "f086a82d-8dbe-4d02-a076-03789704e9d4"

// TestHeuristicPutsPlagueBelchersCountersOnItsLeastCreature: the other
// self-harm clause the catalog declares Harmful. Plague Belcher's enters
// trigger puts two -1/-1 counters on "target creature you control"; the
// bot casts it and puts them on the 1/1, never on the 5/5 or the Belcher.
func TestHeuristicPutsPlagueBelchersCountersOnItsLeastCreature(t *testing.T) {
	for _, bestFirst := range []bool{true, false} {
		name := "mouse listed first"
		if bestFirst {
			name = "titan listed first"
		}
		t.Run(name, func(t *testing.T) {
			g, me, best, least := simulacrumTable(t, 4315, bestFirst)
			var belcher uuid.UUID
			g.WithWriteLock(func() {
				c := game.Card{InstanceID: uuid.New(), Name: "Plague Belcher", OracleID: botPlagueBelcherOracle,
					TypeLine: "Creature — Zombie Beast", ManaCost: "{2}{B}", Power: 5, Toughness: 4,
					Owner: me.ID, Controller: me.ID}
				c.AddKnowersAll(seatIDs(g))
				me.Hand.Cards = []game.Card{c}
				belcher = c.InstanceID
				s := game.Card{InstanceID: uuid.New(), Name: "Swamp", TypeLine: "Basic Land — Swamp",
					Owner: me.ID, Controller: me.ID}
				s.AddKnowersAll(seatIDs(g))
				g.Battlefield.PushTop(s)
			})
			pol := heuristic.New()
			playHeliodUntil(t, g, me, pol, func() bool { return g.Turn.Step != game.StepPrecombatMain })
			took := driveChoices(t, g, pol, me.ID, 10)
			playHeliodUntil(t, g, me, pol, func() bool { return g.Turn.Step != game.StepPrecombatMain })
			if !onBattlefield(g, belcher) {
				t.Fatalf("the bot did not cast Plague Belcher; took %v", took)
			}
			var counters map[uuid.UUID]int
			g.ReadSnapshot(func() {
				counters = map[uuid.UUID]int{}
				for _, c := range g.Battlefield.Cards {
					counters[c.InstanceID] = c.Counters[game.CounterMinusOne]
				}
			})
			if counters[best] != 0 || counters[belcher] != 0 {
				t.Errorf("the bot put -1/-1 counters on its Titan (%d) or its Belcher (%d); took %v", counters[best], counters[belcher], took)
			}
			if onBattlefield(g, least) && counters[least] == 0 {
				t.Errorf("the bot put the counters nowhere near its Mouse; took %v", took)
			}
		})
	}
}
