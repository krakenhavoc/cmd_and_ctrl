package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID

	mine := effectstest.P1Creature(g, me, 3, 3)
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.CastCatalogSpell(t, g, "Unnatural Aggression", "Instant", p1G4UnnaturalOracle, p1G4Fighters(mine, bear))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear the fight killed")
	if got := effectstest.DamageMarkedOn(g, mine); got != effectstest.P1BearPower {
		t.Errorf("my creature has %d damage, want %d from the fight", got, effectstest.P1BearPower)
	}

	big := effectstest.P1Creature(g, opp, 6, 6)
	effectstest.CastCatalogSpell(t, g, "Unnatural Aggression", "Instant", p1G4UnnaturalOracle, p1G4Fighters(mine, big))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, big, game.ZoneBattlefield, "the 6/6 after the fight")
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 6/6 destroyed later this turn")
}

// The rider is the spell's: with my creature gone, no fight happens, and
// the opponent's creature is still marked (the 2015-08-25 ruling).
func TestP1G4UnnaturalAggressionMarksEvenWithoutAFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	mine := effectstest.P1Creature(g, me, 3, 3)
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	effectstest.CastCatalogSpell(t, g, "Unnatural Aggression", "Instant", p1G4UnnaturalOracle, p1G4Fighters(mine, bear))
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
		t.Fatalf("the bear took %d damage with no fight", got)
	}
	effectstest.P1Destroy(t, g, bear)
	effectstest.P1WantZone(t, g, bear, game.ZoneExile, "the bear destroyed later")
}

// Mawloc: X +1/+1 counters, the fight, and the exile on the creature it
// fought — including one that survived the fight and dies later.
func TestP1G4MawlocFightsAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	big := effectstest.P1Creature(g, opp, 6, 6)
	hand := len(me.Hand.Cards)

	mawloc := effectstest.CastXSpell(t, g, "Mawloc", "Creature — Tyranid", p1G4MawlocOracle, "{X}{R}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, mawloc, game.CounterPlusOne); got != 2 {
		t.Fatalf("Mawloc entered with %d +1/+1 counters, want 2", got)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	// The test card has no printed power, so Mawloc's is its counters.
	if got := effectstest.DamageMarkedOn(g, big); got != 2 {
		t.Errorf("the 6/6 has %d damage, want 2 from Mawloc's counters", got)
	}
	if len(me.Hand.Cards) != hand {
		t.Error("X = 2 drew a card")
	}
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 6/6 Mawloc fought, destroyed later")
}

// Ravenous: X of 5 or more draws a card when it enters.
func TestP1G4MawlocRavenousDrawsAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := len(me.Hand.Cards)
	mawloc := effectstest.CastXSpell(t, g, "Mawloc", "Creature — Tyranid", p1G4MawlocOracle, "{X}{R}{G}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, mawloc, game.CounterPlusOne); got != 5 {
		t.Fatalf("Mawloc entered with %d +1/+1 counters, want 5", got)
	}
	// Both enters triggers are queued (Terror from the Deep with no
	// target picked), so their controller orders them.
	if !effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("no ordering prompt for the two enters triggers")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn at X = 5", hand, got)
	}
}

// Cry of the Carnarium: the -2/-2 kills and the replacement exiles; a
// creature card put into a graveyard from the battlefield this turn is
// exiled; one already there, or whose latest arrival was from a
// library, stays; a creature that dies later this turn is exiled.
func TestP1G4CryOfTheCarnarium(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oppSeat := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp := oppSeat.ID

	died := effectstest.P1Creature(g, opp, 5, 5)
	effectstest.P1Destroy(t, g, died)
	effectstest.P1WantZone(t, g, died, game.ZoneGraveyard, "the creature destroyed before the spell")
	old := effectstest.PushGraveyardCardForTest(oppSeat, "Already Dead")
	milled := effectstest.P1Creature(g, opp, 5, 5)
	effectstest.P1Destroy(t, g, milled)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(milled, game.ZoneLibrary); err != nil {
			t.Fatalf("return to library: %v", err)
		}
		if err := g.MillNForEffect(opp, 1); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	effectstest.P1WantZone(t, g, milled, game.ZoneGraveyard, "the creature that died, went to the library and was milled")
	small := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	big := effectstest.P1Creature(g, opp, 3, 3)

	effectstest.CastCatalogSpell(t, g, "Cry of the Carnarium", "Sorcery", p1G4CryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.P1WantZone(t, g, died, game.ZoneExile, "the creature card put there from the battlefield this turn")
	effectstest.P1WantZone(t, g, old, game.ZoneGraveyard, "the creature card already in the graveyard")
	effectstest.P1WantZone(t, g, milled, game.ZoneGraveyard, "the card whose latest arrival was from the library")
	effectstest.P1WantZone(t, g, small, game.ZoneExile, "the 2/2 the -2/-2 killed")
	effectstest.P1WantZone(t, g, big, game.ZoneBattlefield, "the 3/3 as a 1/1")
	effectstest.P1Destroy(t, g, big)
	effectstest.P1WantZone(t, g, big, game.ZoneExile, "the 3/3 destroyed later this turn")
	late := effectstest.P1Creature(g, opp, 4, 4)
	effectstest.P1Destroy(t, g, late)
	effectstest.P1WantZone(t, g, late, game.ZoneExile, "a creature that arrived after the spell")
}

func p1G4TimeCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.B12Counter(t, g, id, game.CounterTime)
}

// The War Doctor: one time counter per batch of other cards put into
// exile, from anywhere; a token is not a card; its own exile does not
// count.
func TestP1G4WarDoctorCountsCardsPutIntoExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oppSeat := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doc := effectstest.B12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)

	a := effectstest.PushGraveyardCardForTest(oppSeat, "Dead A")
	b := effectstest.PushGraveyardCardForTest(me, "Dead B")
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(a)
		_ = g.ExileCardForEffect(b)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 1 {
		t.Fatalf("two cards exiled in one batch: %d time counters, want 1", got)
	}

	creature := effectstest.P1Creature(g, oppSeat.ID, 2, 2)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(creature) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 2 {
		t.Fatalf("a permanent exiled from the battlefield: %d time counters, want 2", got)
	}

	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: oppSeat.ID, Controller: oppSeat.ID,
	})
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 2 {
		t.Fatalf("a token exiled: %d time counters, want 2", got)
	}
}

// A spell countered into exile is a card put into exile from the stack,
// though the counter announces no destination.
func TestP1G4WarDoctorCountsASpellCounteredIntoExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	doc := effectstest.B12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)
	bear := effectstest.P1Creature(g, opp, effectstest.P1BearPower, effectstest.P1BearTo)
	spell := effectstest.CastCatalogSpell(t, g, "Lava Coil", "Sorcery", effectstest.P1LavaCoilOracle, effectstest.Pr6Card(bear))
	g.WithWriteLock(func() {
		if err := g.CounterTargetToZoneForEffect(spell, game.ZoneRef{Kind: game.ZoneExile}); err != nil {
			t.Fatalf("counter into exile: %v", err)
		}
	})
	effectstest.P1WantZone(t, g, spell, game.ZoneExile, "the countered Lava Coil")
	effectstest.PassPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 1 {
		t.Fatalf("a spell countered into exile: %d time counters, want 1", got)
	}
}

// Phasing out: one counter per batch of other permanents.
func TestP1G4WarDoctorCountsPermanentsPhasingOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	doc := effectstest.B12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)
	x := effectstest.P1Creature(g, opp, 2, 2)
	y := effectstest.P1Creature(g, opp, 2, 2)
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, x, y); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := p1G4TimeCounters(t, g, doc); got != 1 {
		t.Fatalf("two permanents phased out together: %d time counters, want 1", got)
	}
}

// The attack trigger deals damage equal to the time counters, and a
// creature dealt damage this way is exiled when it dies.
func TestP1G4WarDoctorAttackDealsItsTimeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doc := effectstest.B12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(doc, game.CounterTime, 3); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	victim := effectstest.P1Creature(g, opp.ID, 3, 3)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(doc, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1WantZone(t, g, victim, game.ZoneExile, "the 3/3 dealt 3 by The War Doctor")
}

// "A creature dealt damage this way": a target whose damage was all
// prevented is not marked.
func TestP1G4WarDoctorPreventedDamageMarksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	doc := effectstest.B12Push(g, me.ID, "The War Doctor", "Legendary Creature — Time Lord Doctor", p1G4WarDoctorOracle, 3, 5)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(doc, game.CounterTime, 3); err != nil {
			t.Fatalf("counters: %v", err)
		}
	})
	shielded := effectstest.P1Creature(g, opp.ID, 3, 3)
	effectstest.P1ShieldCreature(g, shielded)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(doc, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, shielded)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, shielded)
	effectstest.P1WantZone(t, g, shielded, game.ZoneGraveyard, "a creature The War Doctor dealt no damage")
}
