package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// blight_ability_cost_test.go — #4684: the engine rules of
// AbilityCost.Blight. The cards' whole activations (Gristle Glutton,
// Sting-Slinger, Evershrike's Gift, Spiral into Solitude), Doubling
// Season and Vizier of Remedies are pinned in
// cards/effects/blight_ability_cost_test.go.

// TestAbilityBlightValidation is the validator's table: one creature the
// activator controls, any of them; nothing sent for an ability without
// the component; and no creature at all is the counter cost's refusal
// (CR 701.68b).
func TestAbilityBlightValidation(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	put := func(name, typeLine string, owner uuid.UUID) uuid.UUID {
		c := NewCard(name, owner)
		c.TypeLine = typeLine
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	mine := put("Mine", "Creature — Goblin", me.ID)
	other := put("Other", "Creature — Bear", me.ID)
	theirs := put("Theirs", "Creature — Elf", opp.ID)
	rock := put("Rock", "Artifact", me.ID)
	cases := []struct {
		name    string
		player  uuid.UUID
		n       int
		ids     []uuid.UUID
		wantErr error
	}{
		{"one creature you control", me.ID, 1, []uuid.UUID{mine}, nil},
		{"another creature you control", me.ID, 2, []uuid.UUID{other}, nil},
		{"no component, nothing named", me.ID, 0, nil, nil},
		{"no component, an id sent", me.ID, 0, []uuid.UUID{mine}, ErrInvalidParam},
		{"nothing named", me.ID, 1, nil, ErrInvalidParam},
		{"two named", me.ID, 1, []uuid.UUID{mine, other}, ErrInvalidParam},
		{"an opponent's creature", me.ID, 1, []uuid.UUID{theirs}, ErrCardCallerMismatch},
		{"a noncreature you control", me.ID, 1, []uuid.UUID{rock}, ErrNotACreature},
		{"a card that is not on the battlefield", me.ID, 1, []uuid.UUID{uuid.New()}, ErrCardNotFound},
	}
	for _, tc := range cases {
		var err error
		g.ReadSnapshot(func() { err = g.validateAbilityBlightLocked(tc.player, tc.n, tc.ids) })
		if tc.wantErr == nil && err != nil || tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
			t.Errorf("%s: %v, want %v", tc.name, err, tc.wantErr)
		}
	}
	// A player who controls no creature can't blight (CR 701.68b), so
	// the cost can't be paid (CR 118.3), whatever is named.
	var err error
	g.ReadSnapshot(func() { err = g.validateAbilityBlightLocked(g.Seats[2].ID, 1, nil) })
	if !errors.Is(err, ErrCantPayCounterCost) {
		t.Errorf("no creature to blight: %v, want ErrCantPayCounterCost", err)
	}
}

// TestAbilityBlightPaysMinusOneCountersFromACost: the counters land on the
// named creature, and a creature they will kill still pays (CR 701.68b
// asks only that they can be put).
func TestAbilityBlightPaysMinusOneCountersFromACost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := NewCard("Token", me.ID)
	c.TypeLine = "Creature — Goblin"
	c.Power, c.Toughness = 1, 1
	g.Battlefield.PushTop(c)
	var err error
	g.WithWriteLock(func() { err = g.payAbilityBlightLocked(me.ID, 2, []uuid.UUID{c.InstanceID}) })
	if err != nil {
		t.Fatalf("pay: %v", err)
	}
	got := findBattlefieldCard(g, c.InstanceID)
	if got == nil || got.Counters[CounterMinusOne] != 2 {
		t.Fatalf("counters after paying blight 2: %+v", got)
	}
}
