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
