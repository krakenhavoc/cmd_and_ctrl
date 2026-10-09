package game

import (
	"errors"
	"testing"
)

// #2811: "This mana can't be spent to cast spells from your hand"
// (Heartwood Crafter, Karolina Dean). The spend context carries the
// zone a spell is cast from.
func TestNotFromHandRestriction(t *testing.T) {
	pool := ManaPool{{Color: "C", Restrictions: []string{ManaRestrictNotFromHand}}}
	cost, _ := ParseCost("{C}")
	spell := creatureSpendContext()

	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"an activation", ManaSpendContext{Purpose: SpendPurposeActivate}, true},
		{"an unlock", ManaSpendContext{Purpose: SpendPurposeUnlock}, true},
		{"a cast from the graveyard", withCastFrom(spell, ZoneGraveyard), true},
		{"a cast from exile", withCastFrom(spell, ZoneExile), true},
		{"a cast from the command zone", withCastFrom(spell, ZoneCommand), true},
		{"a cast from the library", withCastFrom(spell, ZoneLibrary), true},
		{"a cast from hand", withCastFrom(spell, ZoneHand), false},
		{"a cast from an unknown zone", spell, false},
		{"an unknown purpose", ManaSpendContext{}, false},
	} {
		if got := pool.CanPayFor(cost, 0, tc.ctx); got != tc.want {
			t.Errorf("%s: can pay = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func withCastFrom(ctx ManaSpendContext, z ZoneKind) ManaSpendContext {
	ctx.CastFrom = z
	return ctx
}

func TestManaSpendForCastParamsReadsTheZone(t *testing.T) {
	c := Card{Name: "Bear", TypeLine: "Creature — Bear"}
	for wire, want := range map[string]ZoneKind{"": ZoneHand, "hand": ZoneHand, "command": ZoneCommand, "graveyard": ZoneGraveyard, "nowhere": ""} {
		if got := ManaSpendForCastParams(c, CastSpellParams{FromZone: wire}).CastFrom; got != want {
			t.Errorf("FromZone %q: CastFrom = %q, want %q", wire, got, want)
		}
	}
}

// Through CastSpell in strict mode: the restricted {C} can't cast a
// creature from hand, and casts the same creature from the command zone.
func TestNotFromHandManaThroughTheCastPath(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepPrecombatMain)
	restricted := func() {
		me.ManaPool.EmptyPool()
		me.ManaPool.AddMana(ManaToken{Color: "C", Restrictions: []string{ManaRestrictNotFromHand}})
	}

	inHand := NewCard("Hand Bear", me.ID)
	inHand.TypeLine, inHand.ManaCost, inHand.Power, inHand.Toughness = "Creature — Bear", "{1}", 2, 2
	me.Hand.PushTop(inHand)
	restricted()
	err := g.CastSpell(me.ID, inHand.InstanceID, CastSpellParams{Strict: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("cast from hand with only not-from-hand mana: err = %v, want InsufficientManaError", err)
	}

	commander := NewCard("Commander Bear", me.ID)
	commander.TypeLine, commander.ManaCost, commander.Power, commander.Toughness = "Legendary Creature — Bear", "{1}", 2, 2
	commander.IsCommander = true
	me.Command.PushTop(commander)
	restricted()
	if err := g.CastSpell(me.ID, commander.InstanceID, CastSpellParams{Strict: true, FromZone: "command"}); err != nil {
		t.Fatalf("cast from the command zone with not-from-hand mana: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the restricted mana was not spent: %d left", len(me.ManaPool))
	}
}
