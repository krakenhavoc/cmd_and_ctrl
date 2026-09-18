package game

import (
	"testing"

	"github.com/google/uuid"
)

// leave_game_test.go pins CR 800.4a (#769): a player who concedes or
// loses takes their objects out of the game with them, their control
// effects end, and anything they had taken from somebody else is
// exiled.
//
// The four-seat table is the shape that matters. In a two-player game
// the departure always ends the game, and a game that has ended keeps
// its final board on purpose (ADR 0060 Decision 5), so every removal
// assertion below needs a third and fourth seat to have somebody left
// to observe it.

// zoneHoldingCard names the zone that currently holds cardID, or ""
// when the card is in no tracked zone — which, after a departure, is
// what "left the game" looks like.
func zoneHoldingCard(g *Game, cardID uuid.UUID) string {
	if g.Battlefield.Contains(cardID) {
		return "battlefield"
	}
	if g.Stack.Contains(cardID) {
		return "stack"
	}
	if g.Exile.Contains(cardID) {
		return "exile"
	}
	for _, p := range g.Seats {
		switch {
		case p.Library.Contains(cardID):
			return p.Name + "'s library"
		case p.Hand.Contains(cardID):
			return p.Name + "'s hand"
		case p.Graveyard.Contains(cardID):
			return p.Name + "'s graveyard"
		case p.Command.Contains(cardID):
			return p.Name + "'s command zone"
		}
	}
	return ""
}

// seedCardIn drops a fresh card owned by `owner` into `z` and returns
// its instance ID. The zone is whatever the caller names — the point of
// these tests is that ownership, not location, is what CR 800.4a acts
// on.
func seedCardIn(z *Zone, owner uuid.UUID, name string) uuid.UUID {
	c := NewCard(name, owner)
	z.PushTop(c)
	return c.InstanceID
}

// departureProbe is a four-seat table with one seat (b) holding an
// object in every zone the game tracks, plus a survivor's card in
// b's graveyard for the ownership-not-location case.
type departureProbe struct {
	g          *Game
	a, b, c, d *Player

	permanent uuid.UUID // b's creature, on the battlefield
	commander uuid.UUID // b's commander, in b's command zone
	inHand    uuid.UUID
	inLibrary uuid.UUID
	inGrave   uuid.UUID
	inExile   uuid.UUID

	// cSurvivor is owned by c but sitting in b's graveyard: it must
	// NOT leave when b does.
	cSurvivor uuid.UUID
	// cCreature is c's own creature on the battlefield.
	cCreature uuid.UUID
}

func newDepartureProbe(t *testing.T) departureProbe {
	t.Helper()
	g := newFourPlayerActiveGame(t)
	a, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	p := departureProbe{g: g, a: a, b: b, c: c, d: d}
	p.permanent = pushKeywordCreature(t, g, b, 2, 2)
	p.cCreature = pushKeywordCreature(t, g, c, 2, 2)

	cmdr := NewCommander("B's Commander", b.ID)
	b.Command.PushTop(cmdr)
	p.commander = cmdr.InstanceID

	p.inHand = seedCardIn(b.Hand, b.ID, "B in hand")
	p.inLibrary = seedCardIn(b.Library, b.ID, "B in library")
	p.inGrave = seedCardIn(b.Graveyard, b.ID, "B in graveyard")
	p.inExile = seedCardIn(g.Exile, b.ID, "B in exile")
	p.cSurvivor = seedCardIn(b.Graveyard, c.ID, "C's card in B's graveyard")
	return p
}

// assertBHasLeftTheGame is the CR 800.4a column: nothing b owned is in
// any zone, and the survivors' objects are untouched.
func (p departureProbe) assertBHasLeftTheGame(t *testing.T) {
	t.Helper()
	gone := map[string]uuid.UUID{
		"battlefield permanent": p.permanent,
		"commander":             p.commander,
		"hand card":             p.inHand,
		"library card":          p.inLibrary,
		"graveyard card":        p.inGrave,
		"exiled card":           p.inExile,
	}
	for what, id := range gone {
		if zone := zoneHoldingCard(p.g, id); zone != "" {
			t.Errorf("B's %s is still in the game, in %s", what, zone)
		}
	}
	if p.b.Hand.Size() != 0 || p.b.Library.Size() != 0 || p.b.Command.Size() != 0 {
		t.Errorf("B's zones are not empty: hand %d library %d command %d",
			p.b.Hand.Size(), p.b.Library.Size(), p.b.Command.Size())
	}
	if zone := zoneHoldingCard(p.g, p.cSurvivor); zone == "" {
		t.Error("C's card left the game because it was sitting in B's graveyard — ownership is the test, not location")
	}
	if zone := zoneHoldingCard(p.g, p.cCreature); zone != "battlefield" {
		t.Errorf("C's creature is in %q, want the battlefield", zone)
	}
	if p.c.Library.Size() == 0 || p.d.Library.Size() == 0 {
		t.Error("a survivor's library was swept with the departing player's")
	}
}

// TestConcedeTakesEveryOwnedObjectOutOfTheGame is the headline of
// CR 800.4a: concede, and everything you own goes with you.
func TestConcedeTakesEveryOwnedObjectOutOfTheGame(t *testing.T) {
	p := newDepartureProbe(t)
	if err := p.g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if p.g.State != StateActive {
		t.Fatalf("three seats remain; the game should go on: %s", p.g.State)
	}
	p.assertBHasLeftTheGame(t)
}

// TestSBALossTakesEveryOwnedObjectOutOfTheGame: losing to a
// state-based action goes through the same door as conceding.
func TestSBALossTakesEveryOwnedObjectOutOfTheGame(t *testing.T) {
	p := newDepartureProbe(t)
	p.g.WithWriteLock(func() {
		p.b.Life = 0
		p.g.runStateChecksLocked()
	})
	if !p.b.Eliminated {
		t.Fatal("B at 0 life should have lost")
	}
	p.assertBHasLeftTheGame(t)
}

// TestPoisonLossTakesEveryOwnedObjectOutOfTheGame: and so does the
// poison arm of the same SBA loop.
func TestPoisonLossTakesEveryOwnedObjectOutOfTheGame(t *testing.T) {
	p := newDepartureProbe(t)
	p.g.WithWriteLock(func() {
		p.b.Poison = PoisonLethal
		p.g.runStateChecksLocked()
	})
	if !p.b.Eliminated {
		t.Fatal("B at 10 poison should have lost")
	}
	p.assertBHasLeftTheGame(t)
}

// TestLeavingTheGameIsNotAZoneChange — CR 800.4a is not a destruction
// and not a move, so nothing the departing player owned announces a
// zone change, a leaves-the-battlefield or a dies trigger on its way
// out. A Blood Artist does not see a conceding opponent's board go.
func TestLeavingTheGameFiresNoZoneChangeOrDiesTriggers(t *testing.T) {
	p := newDepartureProbe(t)
	before := len(p.g.Events)
	if err := p.g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	owned := map[uuid.UUID]bool{
		p.permanent: true, p.commander: true, p.inHand: true,
		p.inLibrary: true, p.inGrave: true, p.inExile: true,
	}
	for _, ev := range p.g.Events[before:] {
		if !owned[ev.CardID] {
			continue
		}
		switch ev.Kind {
		case EventZoneMove, EventLTB, EventETB, EventSacrifice, EventMill, EventDiscardCard:
			t.Errorf("leaving the game announced %s for a departed player's card: %+v", ev.Kind, ev)
		}
	}
	if len(p.g.PendingTriggers) != 0 {
		t.Errorf("leaving the game queued %d trigger(s)", len(p.g.PendingTriggers))
	}
}

// TestDepartedPlayersStaticAbilitiesStopApplying: the anthem goes with
// its controller, so the survivors' creatures shrink back.
func TestDepartedPlayersStaticAbilitiesStopApplying(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != "anthem-769" {
			return nil
		}
		return []StaticAbility{{
			Layer:     Layer7PT,
			SubLayer:  SubLayer7C_Modify,
			AppliesTo: func(target *Card, _ *Game, _ *Card) bool { return target.IsCreature() },
			Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
				ch.Power++
				ch.Toughness++
			},
		}}
	})
	anthem := NewCard("B's Anthem", p.b.ID)
	anthem.TypeLine = "Enchantment"
	anthem.OracleID = "anthem-769"
	g.Battlefield.PushTop(anthem)
	g.BumpLayerVersionForTest()

	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := findCard(g, p.cCreature); c == nil || c.CurrentPower() != 3 {
		t.Fatalf("setup: C's creature should be a 3/3 under the anthem: %+v", c)
	}

	if err := g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if zone := zoneHoldingCard(g, anthem.InstanceID); zone != "" {
		t.Errorf("the anthem is still in the game, in %s", zone)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := findCard(g, p.cCreature); c == nil || c.CurrentPower() != 2 {
		t.Errorf("C's creature is still anthemed by a player who has left: %+v", c)
	}
}

// TestDepartedPlayersControlEffectEnds is the Mind Control case: the
// Aura leaves the game with its owner, so the creature goes back to
// the player who owns it rather than being exiled. The ORDER of
// CR 800.4a is the whole of this test — owned objects leave first,
// which is what ends the control effect, and only what is STILL
// controlled after that gets exiled.
func TestDepartedPlayersControlEffectEnds(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != "thief-769" {
			return nil
		}
		return []StaticAbility{controlGrabStatic(func(target *Card) bool {
			return target.InstanceID == p.cCreature
		})}
	})
	thief := NewCard("B's Control Magic", p.b.ID)
	thief.TypeLine = "Enchantment — Aura"
	thief.OracleID = "thief-769"
	g.Battlefield.PushTop(thief)
	g.BumpLayerVersionForTest()

	if got := controllerOfCard(t, g, p.cCreature); got != p.b.ID {
		t.Fatalf("setup: C's creature should be under B's control, got %s", got)
	}

	if err := g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if zone := zoneHoldingCard(g, p.cCreature); zone != "battlefield" {
		t.Fatalf("C's creature should still be on the battlefield, not %q — the effect that stole it ended, it was not still controlled by B", zone)
	}
	if got := controllerOfCard(t, g, p.cCreature); got != p.c.ID {
		t.Errorf("C's creature is controlled by %s after B left, want its owner %s", got, p.c.ID)
	}
}

// TestPermanentStillControlledByADepartedPlayerIsExiled is CR 800.4a's
// last clause: another player's card that was under the departing
// player's control with no continuous effect to end — put onto the
// battlefield under their control — has nobody to control it, so it is
// exiled rather than handed back.
func TestPermanentStillControlledByADepartedPlayerIsExiled(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g
	stolen := NewCard("D's Reanimated Bear", p.d.ID)
	stolen.TypeLine = "Creature — Bear"
	stolen.Power, stolen.Toughness = 2, 2
	stolen.Controller = p.b.ID
	g.Battlefield.PushTop(stolen)
	g.BumpLayerVersionForTest()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	if err := g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.Battlefield.Contains(stolen.InstanceID) {
		t.Error("a permanent still controlled by a departed player stayed on the battlefield")
	}
	if !g.Exile.Contains(stolen.InstanceID) {
		t.Errorf("D's card should be exiled (CR 800.4a), it is in %q", zoneHoldingCard(g, stolen.InstanceID))
	}
}

// TestControlEffectEndingIntoADepartedPlayerExiles is CR 800.4c, the
// case CR 800.4a cannot see: at the moment B leaves, D's creature is
// under A's Control Magic, so it is not "still controlled by B" and is
// not exiled. Later A's effect ends and the creature would go back to
// B, who is not there. Nobody can control it, so it is exiled.
func TestControlEffectEndingIntoADepartedPlayerExiles(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g

	// D's card, but it entered the battlefield under B's control, so
	// B is its CR 613.1b baseline.
	victim := NewCard("D's Bear", p.d.ID)
	victim.TypeLine = "Creature — Bear"
	victim.Power, victim.Toughness = 2, 2
	victim.Controller = p.b.ID
	g.Battlefield.PushTop(victim)

	stealing := true
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != "thief-800-4c" || !stealing {
			return nil
		}
		return []StaticAbility{controlGrabStatic(func(target *Card) bool {
			return target.InstanceID == victim.InstanceID
		})}
	})
	thief := NewCard("A's Control Magic", p.a.ID)
	thief.TypeLine = "Enchantment — Aura"
	thief.OracleID = "thief-800-4c"
	g.Battlefield.PushTop(thief)
	g.BumpLayerVersionForTest()

	if got := controllerOfCard(t, g, victim.InstanceID); got != p.a.ID {
		t.Fatalf("setup: the creature should be under A's control, got %s", got)
	}

	if err := g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if !g.Battlefield.Contains(victim.InstanceID) {
		t.Fatal("A still controls the creature when B leaves, so CR 800.4a must leave it alone")
	}

	// A's effect ends. The baseline underneath it is a player who has
	// left the game.
	stealing = false
	g.WithWriteLock(func() {
		g.BumpLayerVersionForTest()
		g.RecomputeLayersIfStaleLocked()
		g.runStateChecksLocked()
	})
	if g.Battlefield.Contains(victim.InstanceID) {
		t.Error("the creature went back to a player who is not in the game (CR 800.4c)")
	}
	if !g.Exile.Contains(victim.InstanceID) {
		t.Errorf("the creature should be exiled, it is in %q", zoneHoldingCard(g, victim.InstanceID))
	}
}

// TestDepartedPlayersBlockerDealsNoCombatDamage is #766's probe from
// the other side: a creature belonging to a player who has left the
// game is not on the battlefield any more, so it neither takes combat
// damage nor deals any. Before CR 800.4a landed, the dead player's
// 4/4 killed the attacker it was blocking.
func TestDepartedPlayersBlockerDealsNoCombatDamage(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, n := g.Seats[0], g.Seats[1]
	attacker := pushKeywordCreature(t, g, a, 2, 2)
	blocker := pushKeywordCreature(t, g, n, 4, 4)

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, n.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.Concede(n.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if zone := zoneHoldingCard(g, blocker); zone != "" {
		t.Fatalf("the departed player's blocker is still in the game, in %s", zone)
	}

	advanceIntoStep(t, g, StepCombatDamage)
	c := findCard(g, attacker)
	if c == nil {
		t.Fatal("the attacker was killed by a creature belonging to a player who had left the game")
	}
	if c.DamageMarked != 0 {
		t.Errorf("the attacker took %d damage from a departed player's blocker", c.DamageMarked)
	}
}

// TestDepartedPlayersDelayedTriggerNeverFires: nothing goes on the
// stack under the control of a player who is not in the game
// (CR 800.4b/d), delayed triggers included.
func TestDepartedPlayersDelayedTriggerNeverFires(t *testing.T) {
	p := newDepartureProbe(t)
	g := p.g
	g.WithWriteLock(func() {
		g.DelayedTriggers = append(g.DelayedTriggers,
			&DelayedTrigger{ID: uuid.New(), Controller: p.b.ID, At: StepEnd, Label: "B's delayed trigger"},
			&DelayedTrigger{ID: uuid.New(), Controller: p.c.ID, At: StepEnd, Label: "C's delayed trigger"},
		)
	})
	if err := g.Concede(p.b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(g.DelayedTriggers) != 1 || g.DelayedTriggers[0].Controller != p.c.ID {
		t.Errorf("delayed triggers after B left: %+v", g.DelayedTriggers)
	}
}

// TestGameEndingDepartureKeepsTheFinalBoard: the departure that ends
// the game does NOT strip the board. Nothing can observe the objects
// leaving, and the last board is what the winner and the replay look
// at — the same reason the turn cursor stays put (ADR 0060 Decision 5).
func TestGameEndingDepartureKeepsTheFinalBoard(t *testing.T) {
	g := newActiveGame(t)
	winner, loser := g.Seats[0], g.Seats[1]
	theirs := pushKeywordCreature(t, g, loser, 2, 2)
	mine := pushKeywordCreature(t, g, winner, 2, 2)

	if err := g.Concede(loser.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if g.State != StateEnded {
		t.Fatalf("state %s, want ended", g.State)
	}
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(mine) {
		t.Error("the board the game ended on was swept")
	}
}
