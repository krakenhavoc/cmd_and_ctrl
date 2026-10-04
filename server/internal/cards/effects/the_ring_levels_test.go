package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_levels_test.go — ADR 0114 PR 3: the Ring's lines 2 to 4
// (CR 701.54c), driven through real combat. the_ring_test.go covers
// the tempt itself and line 1.

// ringEmblemID is the instance ID of `p`'s Ring emblem, or uuid.Nil.
func ringEmblemID(p *game.Player) uuid.UUID {
	if p.Emblems == nil {
		return uuid.Nil
	}
	for _, c := range p.Emblems.Cards {
		if c.IsRingEmblem() {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// ringTemptTimes has the Ring tempt `player` n times. Each time must be
// a forced choice or no choice: the test arranges at most one creature.
func ringTemptTimes(t *testing.T, g *game.Game, player uuid.UUID, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		ringTempt(t, g, player)
		if ringPrompt(g, player) != nil {
			t.Fatal("setup: the tempt asked for a Ring-bearer; arrange one creature")
		}
	}
}

// ringSettle passes priority until the stack is empty, ordering any
// simultaneous triggers in the order offered (CR 603.3b). It stops at
// any other open prompt, which the caller answers.
func ringSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		passPriorityAroundTable(t, g)
		order := latestChoiceOfKind(g, game.PendingChoiceTriggerOrder)
		if order == nil {
			return
		}
		if err := g.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	t.Fatal("the table never settled")
}

// ringAttack declares `attackers` against `defender` and settles the
// declaration's triggers, answering a loot's discard off the top of the
// hand, then parks the cursor on declare blockers.
func ringAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	declareAttack(t, g, defender, attackers...)
	ringSettleLooting(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
}

// ringSettleLooting is ringSettle that also answers every loot's discard.
func ringSettleLooting(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		ringSettle(t, g)
		answered := false
		for _, p := range g.Seats {
			if discardChoiceFor(g, p.ID) != nil {
				discardFromHand(t, g, p.ID)
				answered = true
			}
		}
		if !answered {
			return
		}
	}
}

// ringTriggersFrom counts the Ring's triggered items on the stack or
// waiting to go there.
func ringTriggersFrom(g *game.Game, emblem uuid.UUID, label string) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.SourceCardID == emblem && item.Label == label {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == emblem && item.Label == label {
			n++
		}
	}
	return n
}

func ringOnBattlefield(g *game.Game, id uuid.UUID) bool {
	ok := false
	g.ReadSnapshot(func() { ok = onBattlefield(g, id) })
	return ok
}

// The Ring has each line from the temptation its number names, and
// keeps it (CR 701.54c): the triggered lines exist only from the second
// temptation on, one more at each, and the chip's text follows.
func TestTheRingGainsItsLinesInOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	want := []string{theRingLine1, theRingLine2, theRingLine3, theRingLine4}
	for n := 1; n <= 5; n++ {
		ringTempt(t, g, me.ID)
		var emblem game.Card
		for _, c := range me.Emblems.Cards {
			if c.IsRingEmblem() {
				emblem = c
			}
		}
		lines := n
		if lines > 4 {
			lines = 4
		}
		if got := len(game.TriggersForCard(emblem)); got != lines-1 {
			t.Errorf("after %d temptations the Ring has %d triggered abilities, want %d", n, got, lines-1)
		}
		var text string
		g.ReadSnapshot(func() { text = g.EmblemsForPlayer(me.ID)[0].Text })
		if text != strings.Join(want[:lines], "\n") {
			t.Errorf("after %d temptations the Ring reads:\n%s", n, text)
		}
	}
}

// Line 2: "Whenever your Ring-bearer attacks, draw a card, then discard
// a card." Not before the second temptation, once per attack, and only
// for the Ring-bearer.
func TestTheRingLevel2LootsWhenTheBearerAttacks(t *testing.T) {
	for _, temptations := range []int{1, 2} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
		ringTemptTimes(t, g, me.ID, temptations)
		other := b12Creature(g, me.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
		emblem := ringEmblemID(me)

		declareAttack(t, g, opp.ID, bear, other)
		loots := ringTriggersFrom(g, emblem, theRingLootLabel)
		if temptations == 1 {
			if loots != 0 {
				t.Fatalf("the Ring looted after one temptation (%d triggers)", loots)
			}
			continue
		}
		if loots != 1 {
			t.Fatalf("two attackers, one of them the Ring-bearer: %d loot triggers, want 1", loots)
		}
		hand := me.Hand.Size()
		ringSettle(t, g)
		if me.Hand.Size() != hand+1 {
			t.Fatalf("hand %d → %d before the discard, want one card drawn", hand, me.Hand.Size())
		}
		if discardOwed(g, me.ID) != 1 {
			t.Fatalf("the loot asks for %d discards, want 1", discardOwed(g, me.ID))
		}
		discardFromHand(t, g, me.ID)
		if me.Hand.Size() != hand {
			t.Fatalf("hand %d after the loot, want %d", me.Hand.Size(), hand)
		}
	}
}

// Line 3: "Whenever your Ring-bearer becomes blocked by a creature, the
// blocking creature's controller sacrifices it at end of combat." Two
// blockers trigger it twice (CR 509.3d); each is sacrificed as the end
// of combat step begins, by whoever controls it then — here, a third
// player who took one of them mid-combat.
func TestTheRingLevel3SacrificesEachBlockerAtEndOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTemptTimes(t, g, me.ID, 3)
	emblem := ringEmblemID(me)
	wallA := b12Creature(g, opp.ID, "Wall A", "Creature — Wall", 0, 4)
	wallB := b12Creature(g, opp.ID, "Wall B", "Creature — Wall", 0, 4)

	ringAttack(t, g, opp.ID, bear)
	for _, w := range []uuid.UUID{wallA, wallB} {
		if err := g.DeclareBlocker(w, bear); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	lockInBlocks(t, g)
	if n := ringTriggersFrom(g, emblem, theRingSacrificeLabel); n != 2 {
		t.Fatalf("two blockers: %d sacrifice triggers, want 2", n)
	}
	ringSettle(t, g)
	if !ringOnBattlefield(g, wallA) || !ringOnBattlefield(g, wallB) {
		t.Fatal("a blocker was sacrificed before end of combat")
	}

	// The third player takes Wall A before end of combat.
	g.WithWriteLock(func() {
		if err := (GainControl{Target: wallA, Controller: third.ID}).Apply(ctxFor(g, &game.StackItem{Controller: third.ID})); err != nil {
			t.Fatalf("GainControl: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})

	advanceTo(t, g, game.StepEndCombat)
	ringSettle(t, g)
	if ringOnBattlefield(g, wallA) || ringOnBattlefield(g, wallB) {
		t.Fatal("a blocker survived end of combat")
	}
	if !opp.Graveyard.Contains(wallA) || !opp.Graveyard.Contains(wallB) {
		t.Fatal("the sacrificed blockers are not in their owner's graveyard")
	}
	sacrificedBy := map[uuid.UUID]uuid.UUID{}
	for _, ev := range eventsOfKind(g, game.EventSacrifice) {
		sacrificedBy[ev.CardID] = ev.Actor
	}
	if sacrificedBy[wallA] != third.ID || sacrificedBy[wallB] != opp.ID {
		t.Fatalf("sacrificed by %v and %v, want the third player (who took Wall A) and the defender",
			sacrificedBy[wallA], sacrificedBy[wallB])
	}
	if !ringOnBattlefield(g, bear) {
		t.Fatal("the Ring-bearer died to two 0-power walls")
	}
}

// "The blocking creature" is the object that blocked: one that leaves
// the battlefield and comes back is a new object (CR 400.7) and is not
// sacrificed.
func TestTheRingLevel3SparesABlockerThatLeftAndCameBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTemptTimes(t, g, me.ID, 3)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)

	ringAttack(t, g, opp.ID, bear)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	ringSettle(t, g)

	// It leaves and comes back: a new object with the same instance ID.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, wall); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneBattlefield}, wall); err != nil {
		t.Fatalf("return: %v", err)
	}
	if !ringOnBattlefield(g, wall) {
		t.Fatal("setup: the blocker did not come back")
	}
	advanceTo(t, g, game.StepEndCombat)
	ringSettle(t, g)
	if !ringOnBattlefield(g, wall) {
		t.Fatal("the blocker came back as a new object and was sacrificed anyway")
	}
}

// Line 3 is about the Ring-bearer becoming blocked. The Ring-bearer
// blocking is not that, and does not trigger it.
func TestTheRingLevel3DoesNotTriggerWhenTheBearerBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringTemptTimes(t, g, opp.ID, 3)
	attacker := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	ringAttack(t, g, opp.ID, attacker)
	if err := g.DeclareBlocker(theirs, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	if n := ringTriggersFrom(g, ringEmblemID(opp), theRingSacrificeLabel); n != 0 {
		t.Fatalf("the Ring-bearer blocking triggered line 3 %d times", n)
	}
	ringSettle(t, g)
	for _, d := range g.DelayedTriggers {
		if d.Body == theRingSacrificeBlockerBody {
			t.Fatal("a delayed sacrifice was scheduled when the Ring-bearer blocked")
		}
	}
}

// Line 4: "Whenever your Ring-bearer deals combat damage to a player,
// each opponent loses 3 life" — every opponent, not only the one dealt
// the damage, and the Ring's owner loses nothing.
func TestTheRingLevel4EachOpponentLosesThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTemptTimes(t, g, me.ID, 4)
	other := b12Creature(g, me.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}

	ringAttack(t, g, g.Seats[1].ID, bear, other)
	advanceTo(t, g, game.StepCombatDamage)
	if n := ringTriggersFrom(g, ringEmblemID(me), theRingLoseLifeLabel); n != 1 {
		t.Fatalf("two attackers connected, one of them the Ring-bearer: %d triggers, want 1", n)
	}
	ringSettle(t, g)
	want := map[uuid.UUID]int{
		me.ID:         life[me.ID],
		g.Seats[1].ID: life[g.Seats[1].ID] - 4 - 3,
		g.Seats[2].ID: life[g.Seats[2].ID] - 3,
		g.Seats[3].ID: life[g.Seats[3].ID] - 3,
	}
	for _, p := range g.Seats {
		if p.Life != want[p.ID] {
			t.Errorf("%s: life %d, want %d", p.Name, p.Life, want[p.ID])
		}
	}
}

// Line 4 sees the damage as it is dealt: a trampling Ring-bearer that
// the blocker kills in the same combat damage step still triggers it,
// because state-based actions come after the damage event.
func TestTheRingLevel4TriggersForABearerThatDiesInTheSameStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bearer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Trampling Ox", TypeLine: "Creature — Ox",
		Power: 3, Toughness: 1, Keywords: []string{"trample"}, Owner: me.ID, Controller: me.ID,
	})
	ringTemptTimes(t, g, me.ID, 4)
	blocker := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs, others := opp.Life, g.Seats[2].Life

	ringAttack(t, g, opp.ID, bearer)
	if err := g.DeclareBlocker(blocker, bearer); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	ringSettle(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	for i := 0; i < 4; i++ {
		c := latestChoiceOfKind(g, game.PendingChoiceDamageAssignment)
		if c == nil {
			break
		}
		if err := g.ResolveDamageAssignment(c.ID, c.Chooser, []game.DamageAssignmentEntry{{BlockerID: blocker, Amount: 2}}, 1); err != nil {
			t.Fatalf("assign: %v", err)
		}
	}
	ringSettle(t, g)
	if ringOnBattlefield(g, bearer) {
		t.Fatal("setup: the Ring-bearer survived the blocker's damage")
	}
	if opp.Life != theirs-1-3 {
		t.Errorf("the defender's life %d → %d, want 1 trample damage and 3 from the Ring", theirs, opp.Life)
	}
	if g.Seats[2].Life != others-3 {
		t.Errorf("another opponent's life %d → %d, want -3", others, g.Seats[2].Life)
	}
}

// Several temptations in one turn: the Ring reaches its fourth line in
// one main phase, there is still one emblem, and every "whenever the
// Ring tempts you" ability triggers once per temptation (a Nazgûl puts
// a counter on itself four times).
func TestTheRingSeveralTemptsInOneTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	nazgul := b12Push(g, me.ID, "Nazgûl", "Creature — Wraith Knight", nazgulOracle, 1, 2)
	turn := g.Turn.Seq
	for i := 0; i < 4; i++ {
		ringTempt(t, g, me.ID)
		ringSettle(t, g)
	}
	if g.Turn.Seq != turn {
		t.Fatal("setup: the turn moved on")
	}
	if ringEmblems(me) != 1 || ringCount(g, me.ID) != 4 {
		t.Fatalf("%d emblems at count %d, want one at 4", ringEmblems(me), ringCount(g, me.ID))
	}
	if got := ringBFCard(t, g, nazgul).Counters["+1/+1"]; got != 4 {
		t.Fatalf("Nazgûl has %d +1/+1 counters after four temptations, want 4", got)
	}
	var emblem game.Card
	for _, c := range me.Emblems.Cards {
		if c.IsRingEmblem() {
			emblem = c
		}
	}
	if got := len(game.TriggersForCard(emblem)); got != 3 {
		t.Fatalf("the Ring has %d triggered lines after four temptations in one turn, want 3", got)
	}
}

// A line gained while the Ring-bearer is already in combat applies from
// then on, and never to what already happened: the attack is not looted
// after the fact, but the block that follows triggers line 3.
func TestTheRingALevelGainedMidCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTemptTimes(t, g, me.ID, 1)
	emblem := ringEmblemID(me)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)

	declareAttack(t, g, opp.ID, bear)
	ringSettle(t, g)
	// The attack is declared; now the Ring tempts twice more, choosing
	// the attacking Bears both times (the only creature I control).
	ringTemptTimes(t, g, me.ID, 2)
	if ringBearerOf(g, me.ID) != bear || ringCount(g, me.ID) != 3 {
		t.Fatal("setup: the attacking Bears are not the Ring-bearer at count 3")
	}
	if n := ringTriggersFrom(g, emblem, theRingLootLabel); n != 0 {
		t.Fatalf("line 2, gained after the attack, looted for it (%d triggers)", n)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	if n := ringTriggersFrom(g, emblem, theRingSacrificeLabel); n != 1 {
		t.Fatalf("line 3, gained mid-combat, did not see the block (%d triggers)", n)
	}
	ringSettle(t, g)
	advanceTo(t, g, game.StepEndCombat)
	ringSettle(t, g)
	if ringOnBattlefield(g, wall) {
		t.Fatal("the blocker survived end of combat")
	}
}

// ADR 0114 §8: the Ring's triggered lines are catalog rows, so a table
// with any of them on the stack — and with line 3's delayed sacrifice
// waiting — is a restore point, and the restored table finishes them.
func TestTheRingTriggersSurviveARestore(t *testing.T) {
	// The blocked path: the loot, the sacrifice trigger, the delayed
	// sacrifice.
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTemptTimes(t, g, me.ID, 4)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)

	declareAttack(t, g, opp.ID, bear)
	if ringTriggersFrom(g, ringEmblemID(me), theRingLootLabel) != 1 {
		t.Fatal("setup: no loot trigger")
	}
	r := restoreRoundTrip(t, g, true)
	rme := r.Seats[0]
	hand := rme.Hand.Size()
	ringSettle(t, r)
	if discardOwed(r, rme.ID) != 1 || rme.Hand.Size() != hand+1 {
		t.Fatalf("restored loot: hand %d → %d, discard owed %d", hand, rme.Hand.Size(), discardOwed(r, rme.ID))
	}
	discardFromHand(t, r, rme.ID)

	advanceTo(t, r, game.StepDeclareBlockers)
	if err := r.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, r)
	if ringTriggersFrom(r, ringEmblemID(rme), theRingSacrificeLabel) != 1 {
		t.Fatal("setup: no sacrifice trigger")
	}
	r = restoreRoundTrip(t, r, true)
	ringSettle(t, r)
	waiting := 0
	for _, d := range r.DelayedTriggers {
		if d.Body == theRingSacrificeBlockerBody {
			waiting++
		}
	}
	if waiting != 1 {
		t.Fatalf("%d delayed sacrifices waiting after the restored trigger resolved, want 1", waiting)
	}
	r = restoreRoundTrip(t, r, true)
	advanceTo(t, r, game.StepEndCombat)
	ringSettle(t, r)
	if ringOnBattlefield(r, wall) {
		t.Fatal("the restored delayed sacrifice did not sacrifice the blocker")
	}

	// The unblocked path: the life loss.
	g = newCatalogGame(t)
	me = g.Seats[0]
	bear = b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringTemptTimes(t, g, me.ID, 4)
	ringAttack(t, g, g.Seats[1].ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	if ringTriggersFrom(g, ringEmblemID(me), theRingLoseLifeLabel) != 1 {
		t.Fatal("setup: no life-loss trigger")
	}
	r = restoreRoundTrip(t, g, true)
	before := r.Seats[2].Life
	ringSettle(t, r)
	if r.Seats[2].Life != before-3 {
		t.Fatalf("restored life loss: %d → %d, want -3", before, r.Seats[2].Life)
	}
}
