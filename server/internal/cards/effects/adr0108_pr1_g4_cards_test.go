package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr1_g4_cards_test.go — ADR 0108 Delivery PR 1 (#1886), group
// 4: Unnatural Aggression, Mawloc, Cry of the Carnarium and The War
// Doctor, each against the board that tells its wording apart.

const (
	p1G4UnnaturalOracle = "597d1dce-67e8-4c37-9781-eeaeb7c2d7b7"
	p1G4MawlocOracle    = "e5d928dc-b465-4cf7-ab11-d5bd3328f8e7"
	p1G4CryOracle       = "1b9a5170-39c0-4cbf-a041-f3c15f1359ae"
	p1G4WarDoctorOracle = "d97553bf-6763-4a7b-8d82-1b438a22aa62"
)

func p1G4Fighters(mine, theirs uuid.UUID) []game.TargetRef {
	return []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Slot: 1},
	}
}

// Unnatural Aggression: the opponent's creature the fight kills is
// exiled; one that survives the fight and dies later is exiled too.
func TestP1G4UnnaturalAggressionExilesTheOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID

	mine := p1Creature(g, me, 3, 3)
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	castCatalogSpell(t, g, "Unnatural Aggression", "Instant", p1G4UnnaturalOracle, p1G4Fighters(mine, bear))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear the fight killed")
	if got := damageMarkedOn(g, mine); got != p1BearPower {
		t.Errorf("my creature has %d damage, want %d from the fight", got, p1BearPower)
	}

	big := p1Creature(g, opp, 6, 6)
	castCatalogSpell(t, g, "Unnatural Aggression", "Instant", p1G4UnnaturalOracle, p1G4Fighters(mine, big))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, big, game.ZoneBattlefield, "the 6/6 after the fight")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 6/6 destroyed later this turn")
}

// The rider is the spell's: with my creature gone, no fight happens, and
// the opponent's creature is still marked (the 2015-08-25 ruling).
func TestP1G4UnnaturalAggressionMarksEvenWithoutAFight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	mine := p1Creature(g, me, 3, 3)
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	castCatalogSpell(t, g, "Unnatural Aggression", "Instant", p1G4UnnaturalOracle, p1G4Fighters(mine, bear))
	p1Destroy(t, g, mine)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, bear); got != 0 {
		t.Fatalf("the bear took %d damage with no fight", got)
	}
	p1Destroy(t, g, bear)
	p1WantZone(t, g, bear, game.ZoneExile, "the bear destroyed later")
}

// Mawloc: X +1/+1 counters, the fight, and the exile on the creature it
// fought — including one that survived the fight and dies later.
func TestP1G4MawlocFightsAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	big := p1Creature(g, opp, 6, 6)
	hand := len(me.Hand.Cards)

	mawloc := castXSpell(t, g, "Mawloc", "Creature — Tyranid", p1G4MawlocOracle, "{X}{R}{G}", 2, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mawloc, game.CounterPlusOne); got != 2 {
		t.Fatalf("Mawloc entered with %d +1/+1 counters, want 2", got)
	}
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	// The test card has no printed power, so Mawloc's is its counters.
	if got := damageMarkedOn(g, big); got != 2 {
		t.Errorf("the 6/6 has %d damage, want 2 from Mawloc's counters", got)
	}
	if len(me.Hand.Cards) != hand {
		t.Error("X = 2 drew a card")
	}
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 6/6 Mawloc fought, destroyed later")
}

// Ravenous: X of 5 or more draws a card when it enters.
func TestP1G4MawlocRavenousDrawsAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := len(me.Hand.Cards)
	mawloc := castXSpell(t, g, "Mawloc", "Creature — Tyranid", p1G4MawlocOracle, "{X}{R}{G}", 5, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, mawloc, game.CounterPlusOne); got != 5 {
		t.Fatalf("Mawloc entered with %d +1/+1 counters, want 5", got)
	}
	// Both enters triggers are queued (Terror from the Deep with no
	// target picked), so their controller orders them.
	if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("no ordering prompt for the two enters triggers")
	}
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn at X = 5", hand, got)
	}
}

// Cry of the Carnarium: the -2/-2 kills and the replacement exiles; a
// creature card put into a graveyard from the battlefield this turn is
// exiled; one already there, or whose latest arrival was from a
// library, stays; a creature that dies later this turn is exiled.
func TestP1G4CryOfTheCarnarium(t *testing.T) {
	g := newCatalogGame(t)
	oppSeat := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp := oppSeat.ID

	died := p1Creature(g, opp, 5, 5)
	p1Destroy(t, g, died)
	p1WantZone(t, g, died, game.ZoneGraveyard, "the creature destroyed before the spell")
	old := pushGraveyardCardForTest(oppSeat, "Already Dead")
	milled := p1Creature(g, opp, 5, 5)
	p1Destroy(t, g, milled)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(milled, game.ZoneLibrary); err != nil {
			t.Fatalf("return to library: %v", err)
		}
		if err := g.MillNForEffect(opp, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	p1WantZone(t, g, milled, game.ZoneGraveyard, "the creature that died, went to the library and was milled")
	small := p1Creature(g, opp, p1BearPower, p1BearTo)
	big := p1Creature(g, opp, 3, 3)

	castCatalogSpell(t, g, "Cry of the Carnarium", "Sorcery", p1G4CryOracle, nil)
	passPriorityAroundTable(t, g)

	p1WantZone(t, g, died, game.ZoneExile, "the creature card put there from the battlefield this turn")
	p1WantZone(t, g, old, game.ZoneGraveyard, "the creature card already in the graveyard")
	p1WantZone(t, g, milled, game.ZoneGraveyard, "the card whose latest arrival was from the library")
	p1WantZone(t, g, small, game.ZoneExile, "the 2/2 the -2/-2 killed")
	p1WantZone(t, g, big, game.ZoneBattlefield, "the 3/3 as a 1/1")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the 3/3 destroyed later this turn")
	late := p1Creature(g, opp, 4, 4)
	p1Destroy(t, g, late)
	p1WantZone(t, g, late, game.ZoneExile, "a creature that arrived after the spell")
}

func p1G4TimeCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return b12Counter(t, g, id, game.CounterTime)
}

// The War Doctor: one time counter per batch of other cards put into
// exile, from anywhere; a token is not a card; its own exile does not
// count.
func TestP1G4WarDoctorCountsCardsPutIntoExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oppSeat := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doc := b12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)

	a := pushGraveyardCardForTest(oppSeat, "Dead A")
	b := pushGraveyardCardForTest(me, "Dead B")
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(a)
		_ = g.ExileCardForEffect(b)
	})
	passPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 1 {
		t.Fatalf("two cards exiled in one batch: %d time counters, want 1", got)
	}

	creature := p1Creature(g, oppSeat.ID, 2, 2)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(creature) })
	passPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 2 {
		t.Fatalf("a permanent exiled from the battlefield: %d time counters, want 2", got)
	}

	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: oppSeat.ID, Controller: oppSeat.ID,
	})
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(token) })
	passPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 2 {
		t.Fatalf("a token exiled: %d time counters, want 2", got)
	}
}

// A spell countered into exile is a card put into exile from the stack,
// though the counter announces no destination.
func TestP1G4WarDoctorCountsASpellCounteredIntoExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	doc := b12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)
	bear := p1Creature(g, opp, p1BearPower, p1BearTo)
	spell := castCatalogSpell(t, g, "Lava Coil", "Sorcery", p1LavaCoilOracle, pr6Card(bear))
	g.WithWriteLock(func() {
		if err := g.CounterTargetToZoneForEffect(spell, game.ZoneRef{Kind: game.ZoneExile}); err != nil {
			t.Fatalf("counter into exile: %v", err)
		}
	})
	p1WantZone(t, g, spell, game.ZoneExile, "the countered Lava Coil")
	passPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 1 {
		t.Fatalf("a spell countered into exile: %d time counters, want 1", got)
	}
}

// Phasing out: one counter per batch of other permanents.
func TestP1G4WarDoctorCountsPermanentsPhasingOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	doc := b12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)
	x := p1Creature(g, opp, 2, 2)
	y := p1Creature(g, opp, 2, 2)
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, x, y); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 1 {
		t.Fatalf("two permanents phased out together: %d time counters, want 1", got)
	}
}

// The attack trigger deals damage equal to the time counters, and a
// creature dealt damage this way is exiled when it dies.
func TestP1G4WarDoctorAttackDealsItsTimeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doc := b12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(doc, game.CounterTime, 3); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	victim := p1Creature(g, opp.ID, 3, 3)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(doc, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, victim, game.ZoneExile, "the 3/3 dealt 3 by The War Doctor")
}

// "A creature dealt damage this way": a target whose damage was all
// prevented is not marked.
func TestP1G4WarDoctorPreventedDamageMarksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doc := b12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(doc, game.CounterTime, 3); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	shielded := p1Creature(g, opp.ID, 3, 3)
	p1ShieldCreature(g, shielded)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(doc, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, shielded)
	passPriorityAroundTable(t, g)
	p1Destroy(t, g, shielded)
	p1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature The War Doctor dealt no damage")
}
