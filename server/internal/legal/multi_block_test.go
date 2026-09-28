package legal_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// multi_block_test.go — #1706. A creature that can block more than one
// attacker is offered each attacker it does not block yet while it has
// room, and nothing once it is full; every offered move is accepted,
// and the move it is no longer offered really is refused.

const (
	oracleTwoHeadedGiant = "38aa31bd-7145-43b9-9409-463d9ad6cd69" // can block an additional creature
	oraclePalaceGuard    = "5c92f375-ae6a-4be5-a499-ab87d6bdc49b" // can block any number
)

// enteredCard is battlefieldCard plus the zone move a real entry emits,
// so the layer pass reaches the card's statics.
func enteredCard(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	id := battlefieldCard(g, p, c)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	return id
}

// blockTargetsOffered lists the attackers the moves offer `blocker`.
func blockTargetsOffered(t *testing.T, moves []legal.Move, blocker uuid.UUID) map[uuid.UUID]bool {
	t.Helper()
	out := map[uuid.UUID]bool{}
	for _, m := range moves {
		if m.Kind != legal.KindBlock {
			continue
		}
		var one struct{ Blocker, Attacker string }
		var set struct {
			Blocks []struct{ Blocker, Attacker string }
		}
		pairs := set.Blocks
		if m.Type == legal.TypeDeclareBlocker {
			if err := json.Unmarshal(m.Params, &one); err != nil {
				t.Fatal(err)
			}
			pairs = append(pairs, one)
		} else {
			if err := json.Unmarshal(m.Params, &set); err != nil {
				t.Fatal(err)
			}
			pairs = set.Blocks
		}
		for _, p := range pairs {
			if p.Blocker == blocker.String() {
				out[uuid.MustParse(p.Attacker)] = true
			}
		}
	}
	return out
}

func TestMultiBlockMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	a := freshCreature(g, me, "Bear A")
	b := freshCreature(g, me, "Bear B")
	c := freshCreature(g, me, "Bear C")
	giant := enteredCard(g, def, game.Card{
		Name: "Two-Headed Giant of Foriys", TypeLine: "Creature — Giant", Power: 4, Toughness: 4,
		OracleID: oracleTwoHeadedGiant,
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b, c} {
		if err := g.DeclareAttacker(id, def.ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("active player's pass: %v", err)
	}

	step := func(want ...uuid.UUID) {
		t.Helper()
		moves := legal.EnumerateFor(g, def.ID)
		dispatchAll(t, g, def.ID, moves)
		got := blockTargetsOffered(t, moves, giant)
		if len(got) != len(want) {
			t.Fatalf("giant offered %v, want %v", got, want)
		}
		for _, w := range want {
			if !got[w] {
				t.Fatalf("giant offered %v, want %v", got, want)
			}
		}
	}
	step(a, b, c)
	if err := g.DeclareBlocker(giant, a); err != nil {
		t.Fatal(err)
	}
	step(b, c)
	if err := g.DeclareBlocker(giant, b); err != nil {
		t.Fatal(err)
	}
	step()
	// The withheld move is refused, not merely unoffered.
	var br *game.BlockRefusedError
	if err := g.Clone().DeclareBlocker(giant, c); !errors.As(err, &br) || br.Reason != game.BlockReasonBlockerCapacity {
		t.Fatalf("a third block: %v, want blocker_capacity", err)
	}
}

// TestAnyNumberBlockerIsOfferedEveryAttacker — Palace Guard keeps being
// offered until it blocks everything.
func TestAnyNumberBlockerIsOfferedEveryAttacker(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	clearHand(def)
	var atks []uuid.UUID
	for _, n := range []string{"A", "B", "C", "D"} {
		atks = append(atks, freshCreature(g, me, "Bear "+n))
	}
	guard := enteredCard(g, def, game.Card{
		Name: "Palace Guard", TypeLine: "Creature — Human Soldier", Power: 1, Toughness: 4,
		OracleID: oraclePalaceGuard,
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range atks {
		if err := g.DeclareAttacker(id, def.ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	for i, atk := range atks {
		moves := legal.EnumerateFor(g, def.ID)
		dispatchAll(t, g, def.ID, moves)
		if got := blockTargetsOffered(t, moves, guard); len(got) != len(atks)-i {
			t.Fatalf("round %d: guard offered %d attackers, want %d", i, len(got), len(atks)-i)
		}
		if err := g.DeclareBlocker(guard, atk); err != nil {
			t.Fatal(err)
		}
	}
	if got := blockTargetsOffered(t, legal.EnumerateFor(g, def.ID), guard); len(got) != 0 {
		t.Fatalf("a guard blocking everything is still offered %v", got)
	}
}
