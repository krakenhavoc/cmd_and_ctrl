package game

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// damage_to_player_by_test.go — #5674. TurnTally.DamagedPlayers records
// how much damage each source object dealt to each player this turn.

func damageDealtToPlayerBy(g *Game, source, victim uuid.UUID) int {
	n := 0
	g.WithWriteLock(func() {
		c := g.findCardByIDLocked(source)
		if c == nil {
			return
		}
		n = g.DamageDealtToPlayerThisTurnByObjectForEffect(ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}, victim)
	})
	return n
}

func TestDamagedPlayersSumsCombatAndNoncombatPerSourceAndPlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp, third := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	ogre := pushPermanent(g, Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: me, Controller: me})
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me, Controller: me})
	rock := pushPermanent(g, Card{Name: "Rock", TypeLine: "Artifact", Owner: me, Controller: me})
	emit(g,
		Event{Kind: EventDealDamage, Actor: me, Source: ogre, Target: opp, Amount: 3, Combat: true},
		Event{Kind: EventDealDamage, Actor: me, Source: ogre, Target: opp, Amount: 4},
		Event{Kind: EventDealDamage, Actor: me, Source: ogre, Target: third, Amount: 2},
		Event{Kind: EventDealDamage, Actor: me, Source: bear, Target: opp, Amount: 2, Combat: true},
		Event{Kind: EventDealDamage, Actor: me, Source: rock, Target: opp, Amount: 1},
		// Damage to a permanent, and damage prevented to 0, are not
		// damage dealt to a player.
		Event{Kind: EventDealDamage, Actor: me, Source: ogre, Target: bear, Amount: 5},
		Event{Kind: EventDealDamage, Actor: me, Source: ogre, Target: opp, Amount: 0},
	)
	for _, tc := range []struct {
		name           string
		source, victim uuid.UUID
		want           int
	}{
		{"combat and noncombat to one player", ogre, opp, 7},
		{"the record is per victim", ogre, third, 2},
		{"the record is per source", bear, opp, 2},
		{"a noncreature source is recorded too", rock, opp, 1},
		{"nothing dealt", bear, third, 0},
	} {
		if got := damageDealtToPlayerBy(g, tc.source, tc.victim); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, got, tc.want)
		}
	}
}

// CR 400.7: a source that leaves and returns is a new object that has
// dealt nothing; the old object's record stays for a reader that holds
// its ref (CR 608.2h).
func TestDamagedPlayersDoesNotCountAFlickeredSourceAndResetsNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	ogre := pushPermanent(g, Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: me, Controller: me})
	var old ObjectRef
	g.WithWriteLock(func() {
		c := g.findCardByIDLocked(ogre)
		old = ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
	})
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: ogre, Target: opp, Amount: 10, Combat: true})
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(ogre); err != nil {
			t.Fatal(err)
		}
		back, err := g.ReturnFromExileToBattlefieldForEffect(ogre, uuid.Nil, false)
		if err != nil {
			t.Fatal(err)
		}
		c := g.findCardByIDLocked(back)
		if c == nil {
			t.Fatal("the ogre did not come back")
		}
		if got := g.DamageDealtToPlayerThisTurnByObjectForEffect(ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}, opp); got != 0 {
			t.Errorf("a flickered source is a new object that dealt nothing, got %d", got)
		}
		if got := g.DamageDealtToPlayerThisTurnByObjectForEffect(old, opp); got != 10 {
			t.Errorf("the old object's record = %d, want 10", got)
		}
	})
	seat := g.Turn.ActiveSeat
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() {
		if got := g.DamageDealtToPlayerThisTurnByObjectForEffect(old, opp); got != 0 {
			t.Errorf("last turn's damage is not this turn's: %d", got)
		}
		if n := len(g.TurnTally.DamagedPlayers); n != 0 {
			t.Errorf("the turn boundary left %d records", n)
		}
	})
}

func TestDamagedPlayersSurviveCloneRestoreAndTheSnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	ogre := pushPermanent(g, Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: me, Controller: me})
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: ogre, Target: opp, Amount: 6, Combat: true})

	snap := g.Clone()
	emit(g, Event{Kind: EventDealDamage, Actor: me, Source: ogre, Target: opp, Amount: 4})
	if got := damageDealtToPlayerBy(snap, ogre, opp); got != 6 {
		t.Errorf("the clone shares its record with the live game: %d", got)
	}
	if got := damageDealtToPlayerBy(g, ogre, opp); got != 10 {
		t.Errorf("live total = %d, want 10", got)
	}
	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if got := damageDealtToPlayerBy(g, ogre, opp); got != 6 {
		t.Errorf("restore did not bring the record back as it was: %d", got)
	}

	// The snapshot round trip: a restored table still knows what the
	// ogre dealt this turn.
	_, restored := roundTrip(t, g)
	if got := damageDealtToPlayerBy(restored, ogre, opp); got != 6 {
		t.Errorf("the snapshot lost the record: %d", got)
	}

	var empty TurnTally
	raw, _ := json.Marshal(empty)
	if strings.Contains(string(raw), "damagedPlayers") {
		t.Errorf("an empty tally must stay additive (omitempty): %s", raw)
	}
}
