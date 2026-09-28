package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// multi_block_view_test.go — #1706, the wire half of a creature that
// blocks more than one attacker: its capacity (block_capacity /
// blocks_any_number), every attacker it blocks (blocking_targets, first
// entry = blocking_target), and the CR 510.1d division prompt
// (damage_assignment.blocker_divides).
func TestMultiBlockIsOnTheWire(t *testing.T) {
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%3]
	push := func(owner *game.Player, name, oracle string, power, tough int) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Creature — Test", OracleID: oracle,
			Power: power, Toughness: tough, Owner: owner.ID, Controller: owner.ID,
		})
		if oracle != "" {
			// The zone move a real entry emits, so the layer pass
			// reaches its statics. Only the blockers: it also makes a
			// creature summoning sick, which an attacker cannot be.
			g.WithWriteLock(func() {
				g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
			})
		}
		return id
	}
	a := push(active, "Bear A", "", 2, 2)
	b := push(active, "Bear B", "", 2, 2)
	giant := push(def, "Two-Headed Giant of Foriys", "38aa31bd-7145-43b9-9409-463d9ad6cd69", 4, 4)
	guard := push(def, "Palace Guard", "5c92f375-ae6a-4be5-a499-ab87d6bdc49b", 1, 4)
	plain := push(def, "Plain Bear", "", 2, 2)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, def.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	view := func(id uuid.UUID) CardView {
		t.Helper()
		for _, c := range ViewOfGame(g).Battlefield.Cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
		t.Fatalf("%s is not on the wire", id)
		return CardView{}
	}
	if v := view(giant); v.BlockCapacity != 2 || v.BlocksAnyNumber {
		t.Errorf("giant: capacity %d, any %v — want 2", v.BlockCapacity, v.BlocksAnyNumber)
	}
	if v := view(guard); v.BlockCapacity != 0 || !v.BlocksAnyNumber {
		t.Errorf("guard: capacity %d, any %v — want any number", v.BlockCapacity, v.BlocksAnyNumber)
	}
	if v := view(plain); v.BlockCapacity != 0 || v.BlocksAnyNumber {
		t.Errorf("an ordinary creature carries capacity %d / any %v", v.BlockCapacity, v.BlocksAnyNumber)
	}

	if err := g.DeclareBlocker(giant, a); err != nil {
		t.Fatal(err)
	}
	if v := view(giant); v.BlockingTarget != a.String() || v.BlockingTargets != nil {
		t.Errorf("one block: target %q, targets %v — want only blocking_target", v.BlockingTarget, v.BlockingTargets)
	}
	if err := g.DeclareBlocker(giant, b); err != nil {
		t.Fatal(err)
	}
	v := view(giant)
	if v.BlockingTarget != a.String() || len(v.BlockingTargets) != 2 || v.BlockingTargets[0] != a.String() || v.BlockingTargets[1] != b.String() {
		t.Fatalf("two blocks: target %q, targets %v", v.BlockingTarget, v.BlockingTargets)
	}

	// The division prompt.
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		if g.Turn.Step == game.StepCombatDamage {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	var da *DamageAssignmentView
	for _, pc := range ViewOfGameFor(g, def.ID.String()).PendingChoices {
		if pc.DamageAssignment != nil {
			da = pc.DamageAssignment
		}
	}
	if da == nil || !da.BlockerDivides || da.AttackerCardID != giant.String() || len(da.BlockerCardIDs) != 2 || da.AllowTrample {
		t.Fatalf("division prompt on the wire = %+v", da)
	}
}
