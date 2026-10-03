package protocol

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestLandTypeEffectsAreStampedOnThePermanent is ADR 0109 §1 decision
// 7: a land whose type a resolved effect changed says which effect and
// for how long, to every viewer, and a land with none says nothing.
func TestLandTypeEffectsAreStampedOnThePermanent(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	land, plain, source := uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Tidal Warrior", TypeLine: "Creature — Merfolk Warrior",
			Owner: owner.ID, Controller: owner.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: other.ID, Controller: other.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: plain, Name: "Plains", TypeLine: "Basic Land — Plains",
			Owner: other.ID, Controller: other.ID})
		g.RegisterScopedEffectForEffect(source, g.PinnedObjectsLocked(land),
			[]game.Mod{game.SetBasicLandTypesMod("Island")}, g.UntilEndOfTurnDuration(), "Tidal Warrior — Island")
	})
	for _, viewer := range []string{owner.ID.String(), other.ID.String()} {
		v := ViewOfGameFor(g, viewer)
		var got, none []LandTypeEffectView
		for _, c := range v.Battlefield.Cards {
			switch c.InstanceID {
			case land.String():
				got = c.LandTypeEffects
			case plain.String():
				none = c.LandTypeEffects
			}
		}
		want := []LandTypeEffectView{{Types: []string{"Island"}, Until: "until end of turn", Source: "Tidal Warrior"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("viewer %s: land_type_effects = %+v, want %+v", viewer, got, want)
		}
		if len(none) != 0 {
			t.Errorf("viewer %s: an unchanged land carries %+v", viewer, none)
		}
	}
}

// TestLandTypeLossIsStampedOnThePermanent is ADR 0109 §2 decision 5:
// Ultima's blighted land says it has no land types and no abilities, for
// as long as it has the counter, and the wire carries an empty types
// list rather than null.
func TestLandTypeLossIsStampedOnThePermanent(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	land, source := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Ultima, Origin of Oblivion", TypeLine: "Legendary Creature — God",
			Owner: owner.ID, Controller: owner.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: other.ID, Controller: other.ID})
		if err := g.AddCounterForEffect(land, "blight", 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
		d, ok := g.ForAsLongAsPinnedHasCounterDuration(land, "blight")
		if !ok {
			t.Fatal("the duration never started")
		}
		g.RegisterScopedEffectForEffect(source, g.PinnedObjectsLocked(land),
			[]game.Mod{game.LoseLandTypesMod(), game.LoseAllAbilitiesMod()}, d, "Ultima — blight")
	})
	v := ViewOfGameFor(g, other.ID.String())
	var got []LandTypeEffectView
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == land.String() {
			got = c.LandTypeEffects
		}
	}
	want := []LandTypeEffectView{{Types: []string{}, LosesAll: true, LosesAbilities: true,
		Until: "for as long as it has a blight counter on it", Source: "Ultima, Origin of Oblivion"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("land_type_effects = %+v, want %+v", got, want)
	}
}
