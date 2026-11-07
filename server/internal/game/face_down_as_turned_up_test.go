package game

import (
	"testing"

	"github.com/google/uuid"
)

// face_down_as_turned_up_test.go — #5237: a card's own "As this creature
// is turned face up, put N +1/+1 counters on it" (CR 614.1e,
// CR 708.11), declared on its face-down cast as
// FaceDownCast.AsTurnedFaceUp. Bubble Smuggler's fixed four and
// Crowd-Control Warden's computed X, in the game package's own terms.

// asTurnedUpOffer is a disguise declaration with an "as this is turned
// face up" clause, the shape effects.AsTurnedFaceUp builds.
func asTurnedUpOffer(up AsTurnedFaceUpCounters) AlternativeCost {
	offer := morphOffer(FaceDownDisguised, morphCost, false)
	offer.Key = "disguise"
	offer.FaceDown.AsTurnedFaceUp = &up
	return offer
}

// castAndLandFaceDown casts the morphGame card face down under the
// offer's key and resolves it onto the battlefield.
func castAndLandFaceDown(t *testing.T, g *Game, me *Player, id uuid.UUID, key string) {
	t.Helper()
	if err := castFaceDown(t, g, me, id, key); err != nil {
		t.Fatalf("cast face down: %v", err)
	}
	resolveTop(t, g)
	if c, _, _ := cardAnywhere(g, id); !c.FaceDownIsPermanent() {
		t.Fatal("fixture: the card did not land face down")
	}
}

// turnUpBySpecialAction tops the pool up and turns the permanent face
// up by its own special action.
func turnUpBySpecialAction(t *testing.T, g *Game, me *Player, id uuid.UUID) {
	t.Helper()
	fillPool(g, me, 4)
	if err := g.PerformSpecialAction(me.ID, id, SpecialActionTurnFaceUp, SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
}

func plusOneCounters(g *Game, id uuid.UUID) int {
	c, _, _ := cardAnywhere(g, id)
	return c.Counters["+1/+1"]
}

// TestAsTurnedFaceUpCountersArriveBeforeTheTurnedUpTrigger is the
// headline: the counters go on as the permanent turns face up, so a
// "when this is turned face up" trigger already sees them (CR 708.11:
// "applied while that permanent is being turned face up, not
// afterward").
func TestAsTurnedFaceUpCountersArriveBeforeTheTurnedUpTrigger(t *testing.T) {
	g, me, id := morphGame(t, asTurnedUpOffer(AsTurnedFaceUpCounters{N: 4}))
	seen := -1
	withCatalogTriggers(t, func(oracleID string) []TriggeredAbility {
		if oracleID != morphOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventTurnedFaceUp},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				if ev.CardID != source.InstanceID {
					return false
				}
				seen = source.Counters["+1/+1"]
				return true
			},
			Key: "turned face up",
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "turned face up")
			},
		}}
	})
	castAndLandFaceDown(t, g, me, id, "disguise")
	if n := plusOneCounters(g, id); n != 0 {
		t.Fatalf("a face-down permanent already has %d +1/+1 counters", n)
	}

	turnUpBySpecialAction(t, g, me, id)

	if n := plusOneCounters(g, id); n != 4 {
		t.Errorf("+1/+1 counters after turning face up = %d, want 4", n)
	}
	if seen != 4 {
		t.Errorf("the turned-face-up trigger saw %d counters, want all 4 already on (CR 708.11)", seen)
	}
	c, _, _ := cardAnywhere(g, id)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 6 {
		t.Errorf("turned up = %d/%d, want the 1/2 card plus four counters, 5/6", c.CurrentPower(), c.CurrentToughness())
	}
}

// TestAsTurnedFaceUpCountersNotWhenCastFaceUp: the clause is about
// turning face up, and a card cast for its mana cost never does.
func TestAsTurnedFaceUpCountersNotWhenCastFaceUp(t *testing.T) {
	g, me, id := morphGame(t, asTurnedUpOffer(AsTurnedFaceUpCounters{N: 4}))
	g.WithWriteLock(func() { g.Turn.Step = StepPrecombatMain })
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "hand", Strict: true}); err != nil {
		t.Fatalf("cast face up: %v", err)
	}
	resolveTop(t, g)
	c, zone, _ := cardAnywhere(g, id)
	if zone != ZoneBattlefield || c.FaceDown {
		t.Fatalf("fixture: card is in %q face down=%v, want face up on the battlefield", zone, c.FaceDown)
	}
	if n := c.Counters["+1/+1"]; n != 0 {
		t.Errorf("a card cast face up got %d +1/+1 counters; it was never turned face up", n)
	}
}

// TestAsTurnedFaceUpComputedCountIsReadAsItTurnsUp is Crowd-Control
// Warden's X: counted at the moment of turning, against the board as it
// is then, with the permanent already face up.
func TestAsTurnedFaceUpComputedCountIsReadAsItTurnsUp(t *testing.T) {
	sawFaceDown := false
	count := func(g *Game, c *Card) int {
		if c.FaceDown {
			sawFaceDown = true
		}
		n := 0
		for i := range g.Battlefield.Cards {
			o := &g.Battlefield.Cards[i]
			if o.InstanceID != c.InstanceID && o.Controller == c.Controller && o.IsCreature() {
				n++
			}
		}
		return n
	}
	g, me, id := morphGame(t, asTurnedUpOffer(AsTurnedFaceUpCounters{N: 99, Count: count}))
	castAndLandFaceDown(t, g, me, id, "disguise")
	// The board grows after the cast and before the turn: the count must
	// read the three creatures there now, not the empty board it was
	// cast onto.
	for _, name := range []string{"Bear A", "Bear B", "Bear C"} {
		turnableCreature(t, g, me, name)
	}
	turnableCreature(t, g, g.Seats[1], "Their Bear")

	turnUpBySpecialAction(t, g, me, id)

	if n := plusOneCounters(g, id); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 (the other creatures you control as it turned up; Count wins over N)", n)
	}
	if sawFaceDown {
		t.Error("Count was asked about the face-down object; it must see the face-up card")
	}
}

// TestAsTurnedFaceUpAppliesWhenAnEffectTurnsItUp: unlike megamorph's
// counter, the clause is the card's ability and not a rider on a cost,
// so an effect that turns the permanent face up still puts them on.
// Megamorph on the same door still gets none (CR 702.37b).
func TestAsTurnedFaceUpAppliesWhenAnEffectTurnsItUp(t *testing.T) {
	for _, tc := range []struct {
		name  string
		offer AlternativeCost
		kind  FaceDownKind
		want  int
	}{
		{"as turned face up", asTurnedUpOffer(AsTurnedFaceUpCounters{N: 4}), FaceDownDisguised, 4},
		{"megamorph", morphOffer(FaceDownMorphed, morphCost, true), FaceDownMorphed, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withCatalogAlternativeCosts(t, altCostFor(morphOracle, tc.offer))
			g := newActiveGame(t)
			me := g.Seats[0]
			id := turnableCreature(t, g, me, "Smuggler")
			findBattlefieldCard(g, id).OracleID = morphOracle
			g.applyFaceDownLandingLocked(g.Battlefield, id, tc.kind, nil)
			if !g.TurnFaceUpForEffect(uuid.Nil, me.ID, id) {
				t.Fatal("the effect door refused")
			}
			if n := findBattlefieldCard(g, id).Counters["+1/+1"]; n != tc.want {
				t.Errorf("+1/+1 counters after an effect turned it up = %d, want %d", n, tc.want)
			}
		})
	}
}

// TestMegamorphIsUnchangedByAsTurnedFaceUp: a megamorph card with no
// clause gets exactly its one counter from the special action, and the
// offer still carries it as before.
func TestMegamorphIsUnchangedByAsTurnedFaceUp(t *testing.T) {
	g, me, id := morphGame(t, morphOffer(FaceDownMorphed, morphCost, true))
	castAndLandFaceDown(t, g, me, id, "morph")
	c, _, _ := cardAnywhere(g, id)
	offer := TurnFaceUpOffer(c)
	if offer == nil || !offer.FaceUpCounter {
		t.Fatalf("megamorph offer = %+v, want one carrying FaceUpCounter", offer)
	}
	turnUpBySpecialAction(t, g, me, id)
	if n := plusOneCounters(g, id); n != 1 {
		t.Errorf("megamorph +1/+1 counters = %d, want 1 (CR 702.37b)", n)
	}
}

// TestAsTurnedFaceUpCountersSurviveARestore: the counters are ordinary
// counters once on, and a face-down permanent restored from a snapshot
// still gets them when it turns up, because they are read off the
// catalog row and nothing was stored for them.
func TestAsTurnedFaceUpCountersSurviveARestore(t *testing.T) {
	t.Run("turned up, then restored", func(t *testing.T) {
		g, me, id := morphGame(t, asTurnedUpOffer(AsTurnedFaceUpCounters{N: 4}))
		castAndLandFaceDown(t, g, me, id, "disguise")
		turnUpBySpecialAction(t, g, me, id)
		_, restored := roundTrip(t, g)
		c := findBattlefieldCard(restored, id)
		if c == nil {
			t.Fatal("the permanent did not survive the snapshot")
		}
		if n := c.Counters["+1/+1"]; n != 4 {
			t.Errorf("restored +1/+1 counters = %d, want 4", n)
		}
	})
	t.Run("restored face down, then turned up", func(t *testing.T) {
		g, me, id := morphGame(t, asTurnedUpOffer(AsTurnedFaceUpCounters{N: 4}))
		castAndLandFaceDown(t, g, me, id, "disguise")
		_, restored := roundTrip(t, g)
		rme := restored.playerByIDLocked(me.ID)
		if rme == nil {
			t.Fatal("seat lost in the restore")
		}
		restored.WithWriteLock(func() { restored.Turn.Step = StepPrecombatMain })
		turnUpBySpecialAction(t, restored, rme, id)
		if n := plusOneCounters(restored, id); n != 4 {
			t.Errorf("+1/+1 counters after a restored permanent turned up = %d, want 4", n)
		}
	})
}
