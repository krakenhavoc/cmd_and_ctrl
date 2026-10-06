package game

import (
	"testing"

	"github.com/google/uuid"
)

// ADR 0115 decision 6 (PR 1): commander tax (CR 903.8) and commander
// damage (CR 903.10a) belong to the commander CARD (CR 903.3), so a
// blink that gives it a new object ID (CR 400.7) keeps both.

// blinkCommander exiles the battlefield commander `id` the way a
// Cloudshift does today, declining the CR 903.9a offer, and returns it
// as a new object. It returns the new instance ID.
func blinkCommander(t *testing.T, g *Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	var owner uuid.UUID
	if c := findBattlefieldCard(g, id); c != nil {
		owner = c.Owner
	} else {
		t.Fatalf("commander %s is not on the battlefield", id)
	}
	var err error
	g.WithWriteLock(func() { err = g.ExileCardForEffect(id) })
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	if len(g.PendingChoices) > 0 {
		answerOnlyCommanderPrompt(t, g, owner, false)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the commander is not in exile after declining the command zone")
	}
	var entered uuid.UUID
	g.WithWriteLock(func() { entered, err = g.ReturnFromExileToBattlefieldForEffect(id, owner, false) })
	if err != nil {
		t.Fatalf("ReturnFromExileToBattlefieldForEffect: %v", err)
	}
	if entered == uuid.Nil || entered == id {
		t.Fatalf("returned as %s, want a new object (old ID %s)", entered, id)
	}
	if findBattlefieldCard(g, entered) == nil {
		t.Fatal("the blinked commander is not on the battlefield")
	}
	return entered
}

// TestABlinkedCommanderKeepsItsTax: a commander cast once, blinked,
// sent home and cast again pays {2} more (CR 903.8 counts casts of
// the card), and its count goes on from 1 to 2.
func TestABlinkedCommanderKeepsItsTax(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	cmdr := NewCommander("Test Commander", p.ID)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.ManaCost = "{1}{W}"
	cmdr.Power, cmdr.Toughness = 2, 2
	g.Battlefield.PushTop(cmdr)
	old := cmdr.InstanceID
	p.CommanderCasts[old] = 1

	id := blinkCommander(t, g, old)

	if got := p.CommanderCasts[id]; got != 1 {
		t.Fatalf("CommanderCasts[new ID] = %d, want 1: the blink reset the tax", got)
	}
	if _, ok := p.CommanderCasts[old]; ok {
		t.Error("the old instance ID still has a CommanderCasts entry")
	}

	// Send it home and cast it again: {1}{W} plus {2} of tax.
	if _, err := MoveCard(g.Battlefield, p.Command, id); err != nil {
		t.Fatalf("move to the command zone: %v", err)
	}
	p.ManaPool = nil
	p.ManaPool.AddMana(ManaToken{Color: "W"}, ManaToken{Color: "C"})
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, FromZone: "command"}); err == nil {
		t.Fatal("the cast paid no tax: a blinked commander was priced as if never cast")
	}
	p.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, FromZone: "command"}); err != nil {
		t.Fatalf("cast with {2} of tax: %v", err)
	}
	if got := p.CommanderCasts[id]; got != 2 {
		t.Errorf("CommanderCasts after the second cast = %d, want 2", got)
	}
}

// TestABlinkedCommanderKeepsItsCommanderDamage: 14 damage before the
// blink and 7 after is 21 from the same commander (CR 903.10a), and
// the defender loses. Before ADR 0115 the blink started a new clock.
func TestABlinkedCommanderKeepsItsCommanderDamage(t *testing.T) {
	g := newActiveGame(t)
	atk, def := g.Seats[0], g.Seats[1]
	cmdr := NewCommander("Test Commander", atk.ID)
	cmdr.TypeLine = "Legendary Creature — Avatar"
	cmdr.Power, cmdr.Toughness = 7, 7
	g.Battlefield.PushTop(cmdr)
	// A partner on the same seat, whose total must not move.
	partner := NewCommander("Test Partner", atk.ID)
	partner.TypeLine = "Legendary Creature — Avatar"
	g.Battlefield.PushTop(partner)

	old := cmdr.InstanceID
	def.CommanderDamage[old] = 14
	def.CommanderDamage[partner.InstanceID] = 5

	id := blinkCommander(t, g, old)

	if got := def.CommanderDamage[id]; got != 14 {
		t.Fatalf("CommanderDamage[new ID] = %d, want 14: the blink started a fresh clock", got)
	}
	if _, ok := def.CommanderDamage[old]; ok {
		t.Error("the old instance ID still has a CommanderDamage entry")
	}
	if got := def.CommanderDamage[partner.InstanceID]; got != 5 {
		t.Errorf("the partner's total = %d, want 5 (untouched)", got)
	}

	// Past this turn, so the returned commander can attack.
	for g.Turn.ActiveSeat == 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for !(g.Turn.Step == StepDeclareAttackers && g.Turn.ActiveSeat == 0) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.DeclareAttacker(id, def.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if got := def.CommanderDamage[id]; got != 21 {
		t.Fatalf("commander damage after the hit = %d, want 21", got)
	}
	if !def.Eliminated {
		t.Error("21 damage from one commander across a blink did not eliminate the defender")
	}
}

// TestABlinkReKeysEveryPlayersTallies: CR 903.8 counts the casts of
// "the player casting it", which the cast path records on the caster,
// so a cast by another player is carried too. A non-commander's blink
// moves nothing.
func TestABlinkReKeysEveryPlayersTallies(t *testing.T) {
	g := newActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	cmdr := NewCommander("Test Commander", owner.ID)
	cmdr.TypeLine = "Legendary Creature — Bear"
	g.Battlefield.PushTop(cmdr)
	old := cmdr.InstanceID
	owner.CommanderCasts[old] = 2
	other.CommanderCasts[old] = 1

	id := blinkCommander(t, g, old)

	if owner.CommanderCasts[id] != 2 || other.CommanderCasts[id] != 1 {
		t.Errorf("casts after the blink: owner %d other %d, want 2 and 1",
			owner.CommanderCasts[id], other.CommanderCasts[id])
	}

	// A non-commander keyed by mistake is left alone: the re-key is a
	// commander's alone (ADR 0115 decision 6).
	bear := NewCard("Test Bear", owner.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)
	owner.CommanderCasts[bear.InstanceID] = 3
	var newBear uuid.UUID
	var err error
	g.WithWriteLock(func() { err = g.ExileCardForEffect(bear.InstanceID) })
	if err != nil {
		t.Fatalf("exile the bear: %v", err)
	}
	g.WithWriteLock(func() { newBear, err = g.ReturnFromExileToBattlefieldForEffect(bear.InstanceID, owner.ID, false) })
	if err != nil || newBear == uuid.Nil {
		t.Fatalf("return the bear: %s %v", newBear, err)
	}
	if _, ok := owner.CommanderCasts[newBear]; ok {
		t.Error("a non-commander's blink moved a tally")
	}
}
