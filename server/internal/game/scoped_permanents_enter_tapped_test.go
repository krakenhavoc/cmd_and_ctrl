package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// scoped_permanents_enter_tapped_test.go — #6147: ModPermanentsEnterTapped,
// "<permanents> enter tapped this turn" (Due Respect, Nahiri's
// Lithoforming). CR 614.1d: "Continuous effects that read "[This
// permanent] enters . . ." or "[Objects] enter [the battlefield] . . ."
// are replacement effects." CR 514.2: "this turn" effects end in cleanup.

func handTypedCardFor(g *Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Seats[g.seatOfForTest(owner)].Hand.PushTop(Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2,
		Owner: owner, Controller: owner,
	})
	return id
}

func permanentsEnterTapped(t *testing.T, g *Game, controller uuid.UUID, types ...string) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.PermanentsEnterTappedThisTurnForEffect(uuid.Nil, controller, types, "test: enter tapped this turn") {
			t.Fatal("the replacement was not registered")
		}
	})
}

// Due Respect: every permanent of every player enters tapped this turn,
// and next turn nothing does.
func TestPermanentsEnterTappedThisTurnForEveryone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	permanentsEnterTapped(t, g, uuid.Nil)
	if !putOntoBattlefieldTapped(t, g, me.ID, handTypedCardFor(g, me.ID, "Forest", "Basic Land — Forest")) {
		t.Error("the caster's land entered untapped")
	}
	if !putOntoBattlefieldTapped(t, g, opp.ID, handTypedCardFor(g, opp.ID, "Bear", "Creature — Bear")) {
		t.Error("an opponent's creature entered untapped")
	}
	if !putOntoBattlefieldTapped(t, g, opp.ID, handTypedCardFor(g, opp.ID, "Rod", "Artifact")) {
		t.Error("an opponent's artifact entered untapped")
	}
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if putOntoBattlefieldTapped(t, g, opp.ID, handTypedCardFor(g, opp.ID, "Next Bear", "Creature — Bear")) {
		t.Error("the replacement outlived its turn")
	}
	if putOntoBattlefieldTapped(t, g, me.ID, handTypedCardFor(g, me.ID, "Next Forest", "Basic Land — Forest")) {
		t.Error("the replacement outlived its turn for the caster")
	}
}

// Nahiri's Lithoforming: only lands, and only its controller's.
func TestLandsYouControlEnterTappedThisTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	permanentsEnterTapped(t, g, me.ID, "land")
	if !putOntoBattlefieldTapped(t, g, me.ID, handTypedCardFor(g, me.ID, "Forest", "Basic Land — Forest")) {
		t.Error("the controller's land entered untapped")
	}
	if !putOntoBattlefieldTapped(t, g, me.ID, handTypedCardFor(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")) {
		t.Error("the controller's land creature entered untapped")
	}
	if putOntoBattlefieldTapped(t, g, me.ID, handTypedCardFor(g, me.ID, "Bear", "Creature — Bear")) {
		t.Error("the controller's nonland creature entered tapped")
	}
	if putOntoBattlefieldTapped(t, g, opp.ID, handTypedCardFor(g, opp.ID, "Opp Forest", "Basic Land — Forest")) {
		t.Error("an opponent's land entered tapped")
	}
}

// The record is a restore point: it survives a round trip, and a mod
// naming a non-permanent type is refused at registration and restore.
func TestPermanentsEnterTappedSurvivesASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	permanentsEnterTapped(t, g, me.ID, "land")
	_, restored := roundTrip(t, g)
	if !putOntoBattlefieldTapped(t, restored, me.ID, handTypedCardFor(restored, me.ID, "Forest", "Basic Land — Forest")) {
		t.Error("the restored replacement did not apply")
	}
	if permanentsEnterTappedModProblem(Mod{Kind: ModPermanentsEnterTapped, Types: []string{"instant"}}) == "" {
		t.Error("a filter naming instant was accepted")
	}
	snap := g.CaptureSnapshot()
	for i := range snap.ScopedEffects {
		if len(snap.ScopedEffects[i].Mods) == 1 && snap.ScopedEffects[i].Mods[0].Kind == ModPermanentsEnterTapped {
			snap.ScopedEffects[i].Mods[0].Types = []string{"sorcery"}
		}
	}
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("a restore point naming sorcery = %v, want ErrUnknownEffectKey", err)
	}
}
