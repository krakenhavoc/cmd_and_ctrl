package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// command_zone_put_test.go — #1278: PutFromCommandZoneOntoBattlefieldForEffect,
// the door commander ninjutsu (CR 702.49c) puts its card through. The
// card-level keyword is cards/effects/commander_ninjutsu_test.go.

// pushCommandZoneCommander seeds a creature card into its owner's
// command zone as their commander.
func pushCommandZoneCommander(owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Legendary Creature — Human Ninja"
	c.Power, c.Toughness = 1, 3
	c.IsCommander = true
	owner.Command.PushTop(c)
	return c.InstanceID
}

// A card PUT from the command zone arrives tapped and attacking under
// the same instance ID, keeps its designation, is not a cast (the
// CR 903.8 tally is untouched) and asks no CR 903.9 question — that
// replacement is about leaving for a library, hand, graveyard or
// exile, and the battlefield is none of them.
func TestPutFromCommandZoneOntoBattlefieldIsNotACast(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	cmd := pushCommandZoneCommander(me, "Yuriko")
	me.CommanderCasts[cmd] = 1
	declareAttacks(t, g, attacker)

	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.PutFromCommandZoneOntoBattlefieldForEffect(cmd, ZoneEntryOptions{
			Controller: me.ID,
			Tapped:     true,
			Attacking:  opp.ID,
		})
	})
	if err != nil {
		t.Fatalf("PutFromCommandZoneOntoBattlefieldForEffect: %v", err)
	}
	if entered != cmd {
		t.Fatalf("entered %s, want the same instance %s — CommanderCasts is keyed by it", entered, cmd)
	}
	if me.Command.Contains(cmd) {
		t.Error("the commander is still in the command zone")
	}
	c := findBattlefieldCard(g, cmd)
	if c == nil {
		t.Fatal("the commander did not reach the battlefield")
	}
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("tapped=%v attacking=%s, want tapped and attacking %s", c.Tapped, c.AttackingTarget, opp.ID)
	}
	if !c.IsCommander {
		t.Error("the designation did not survive the exit")
	}
	if got := me.CommanderCasts[cmd]; got != 1 {
		t.Errorf("CommanderCasts = %d, want 1 — a put is not a cast (CR 903.8)", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d pending choices; an arrival on the battlefield is not a CR 903.9 destination", len(g.PendingChoices))
	}

	// And the designation is live on the way OUT: bounced, the owner is
	// offered CR 903.9's command zone by the shared exit primitive.
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(cmd); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d pending choices after bouncing the commander, want the one CR 903.9 prompt", len(g.PendingChoices))
	}
}

// The door names its zone: a card anywhere else is refused and nothing
// moves, the posture the hand and library doors take.
func TestPutFromCommandZoneRefusesACardNotInACommandZone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	inHand := pushHandCreature(t, g, me, "Ninja", 2, 2)
	g.WithWriteLock(func() {
		if _, err := g.PutFromCommandZoneOntoBattlefieldForEffect(inHand, ZoneEntryOptions{}); !errors.Is(err, ErrCardNotFound) {
			t.Errorf("a hand card through the command-zone door: %v, want ErrCardNotFound", err)
		}
	})
	if !me.Hand.Contains(inHand) {
		t.Error("the refused put moved the card")
	}
}
