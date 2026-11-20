package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bending_cards_test.go — #5276: the four bending events raised by the
// cards and costs that bend, and Avatar Aang // Aang, Master of Elements.

const (
	bendRuinousWaterbendingOracle = "7d3a0216-871b-4c1b-adb1-de99b832f577"
	bendAzulaOnTheHuntOracle      = "89f96032-5cba-404d-82b0-38d2be963e4b"
	bendRokuFirebenderOracle      = "9eb9e4c1-fabf-45ff-a584-fbe4b94fb41c"
)

func bendEvents(g *game.Game, kind game.EventKind, actor uuid.UUID) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.Actor == actor {
			out = append(out, ev)
		}
	}
	return out
}

// avatarAangCard is the card as the importer builds it, face 0 up.
func avatarAangCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   effects.AvatarAangOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Avatar Aang",
				TypeLine:  "Legendary Creature — Human Avatar Ally",
				ManaCost:  "{R}{G}{W}{U}",
				Colors:    []string{"G", "R", "U", "W"},
				Power:     4,
				Toughness: 4,
			},
			{
				Name:      "Aang, Master of Elements",
				TypeLine:  "Legendary Creature — Avatar Ally",
				Colors:    []string{"G", "R", "U", "W"},
				Power:     6,
				Toughness: 6,
			},
		},
	}
	c.SetFace(0)
	return c
}

func TestBendingAvatarAangBothFacesRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effects.AvatarAangOracle:        "Avatar Aang",
		effects.AvatarAangOracle + "#1": "Aang, Master of Elements",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// Each of the four draws a card; the fourth kind transforms him.
// Another player's bend does nothing for him.
func TestBendingAvatarAangDrawsForEachBendAndTransformsOnTheFourth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	aang := avatarAangCard(me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, aang)
	bend := func(kind game.EventKind, who uuid.UUID) {
		t.Helper()
		g.WithWriteLock(func() { g.NoteBendForEffect(kind, who, uuid.Nil) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	hand := me.Hand.Size()
	bend(game.EventAirbend, opp.ID)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("an opponent's airbend drew me %d cards", got-hand)
	}
	bend(game.EventWaterbend, me.ID)
	bend(game.EventEarthbend, me.ID)
	bend(game.EventWaterbend, me.ID)
	bend(game.EventFirebend, me.ID)
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand %d after four of my bends, want %d", got, hand+4)
	}
	if c := effectstest.BattlefieldCardFor(g, aang.InstanceID); c == nil || c.ActiveFace != 0 {
		t.Fatal("Avatar Aang transformed without an airbend")
	}
	bend(game.EventAirbend, me.ID)
	c := effectstest.BattlefieldCardFor(g, aang.InstanceID)
	if c == nil || c.ActiveFace != 1 || c.Name != "Aang, Master of Elements" {
		t.Fatalf("after all four: %+v, want Aang, Master of Elements", c)
	}
	if got := me.Hand.Size(); got != hand+5 {
		t.Errorf("hand %d, want %d: the fourth kind still draws", got, hand+5)
	}
}

// The record counts bends made before Aang arrived: one bend with him
// out finishes the set.
func TestBendingAvatarAangCountsBendsFromEarlierInTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() {
		g.NoteBendForEffect(game.EventWaterbend, me.ID, uuid.Nil)
		g.NoteBendForEffect(game.EventEarthbend, me.ID, uuid.Nil)
		g.NoteBendForEffect(game.EventFirebend, me.ID, uuid.Nil)
	})
	aang := avatarAangCard(me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, aang)
	g.WithWriteLock(func() { g.NoteBendForEffect(game.EventAirbend, me.ID, uuid.Nil) })
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.BattlefieldCardFor(g, aang.InstanceID); c == nil || c.ActiveFace != 1 {
		t.Fatal("Avatar Aang did not transform on the fourth kind of the turn")
	}
}

// The back face: spells cost {W}{U}{B}{R}{G} less, reducing generic too.
func TestBendingAangMasterOfElementsDiscountsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	aang := avatarAangCard(me.ID)
	aang.SetFace(1)
	effectstest.PushBattlefieldCardWithTimestamp(g, aang)
	// {3}{U}{U}: the {U} takes one blue, the other four symbols take
	// four generic (CR 118.7b) — three generic is all there is.
	id := effectstest.HandSpell(me, "Big Spell", "Sorcery", "{3}{U}{U}")
	effectstest.FloatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{3}{U}{U} under Aang, Master of Elements from {U}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left, want 0", me.ManaPool.Total())
	}
}

// The upkeep trigger: "yes" turns him over and gives the four rewards;
// "no" does nothing.
func TestBendingAangMasterOfElementsTransformsBackAtUpkeep(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		aang := avatarAangCard(me.ID)
		aang.SetFace(1)
		effectstest.PushBattlefieldCardWithTimestamp(g, aang)
		effectstest.EndTurn(t, g)
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
		life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.BattlefieldCardFor(g, aang.InstanceID)
		if c == nil {
			t.Fatal("Aang left the battlefield")
		}
		if !yes {
			if c.ActiveFace != 1 || me.Life != life || me.Hand.Size() != hand {
				t.Errorf("declined: face %d, life %d (was %d), hand %d (was %d)", c.ActiveFace, me.Life, life, me.Hand.Size(), hand)
			}
			continue
		}
		if c.ActiveFace != 0 {
			t.Errorf("face %d after yes, want Avatar Aang", c.ActiveFace)
		}
		if me.Life != life+4 {
			t.Errorf("life %d, want %d", me.Life, life+4)
		}
		if got := me.Hand.Size(); got != hand+4 {
			t.Errorf("hand %d, want %d", got, hand+4)
		}
		if n := effectstest.CountersOn(g, aang.InstanceID, game.CounterPlusOne); n != 4 {
			t.Errorf("%d +1/+1 counters, want 4", n)
		}
		if opp.Life != oppLife-4 {
			t.Errorf("opponent life %d, want %d", opp.Life, oppLife-4)
		}
	}
}

// CR 701.67c: paying a waterbend cost on an activation is a waterbend.
func TestBendingWaterbendCostOfAnActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	aang := effectstest.AangSwiftSaviorCard(me.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, aang)
	effectstest.FloatForTest(g, me, "CCCCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, aang.InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Waterbend {8}: %v", err)
	}
	if n := len(bendEvents(g, game.EventWaterbend, me.ID)); n != 1 {
		t.Fatalf("%d waterbend events, want 1", n)
	}
	g.WithWriteLock(func() {
		if n := g.BendsThisTurn(me.ID, game.EventWaterbend); n != 1 {
			t.Errorf("waterbends this turn %d, want 1", n)
		}
	})
}

// An optional waterbend is a waterbend only when it is paid.
func TestBendingOptionalWaterbendOnlyWhenPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ruinous Waterbending", TypeLine: "Sorcery",
				OracleID: bendRuinousWaterbendingOracle, Owner: me.ID, Controller: me.ID})
		})
		params := game.CastSpellParams{}
		if pay {
			effectstest.FloatForTest(g, me, "CCCC")
			params.OptionalCosts = []int{0}
		}
		if err := g.CastSpell(me.ID, id, params); err != nil {
			t.Fatalf("cast (pay %v): %v", pay, err)
		}
		want := 0
		if pay {
			want = 1
		}
		if n := len(bendEvents(g, game.EventWaterbend, me.ID)); n != want {
			t.Errorf("pay %v: %d waterbend events, want %d", pay, n, want)
		}
	}
}

// CR 702.189b: a firebending ability resolving is a firebend, announced
// as it resolves and not as it triggers. Red mana kept through combat
// that is not firebending (Avatar Roku, Firebender) is not one.
func TestBendingFirebendIsTheAbilityResolving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	azula := effectstest.B12Push(g, me.ID, "Azula, On the Hunt", "Legendary Creature — Human Noble", bendAzulaOnTheHuntOracle, 4, 3)
	effectstest.B12Push(g, me.ID, "Avatar Roku, Firebender", "Legendary Creature — Human Avatar", bendRokuFirebenderOracle, 6, 6)
	effectstest.DeclareAttack(t, g, opp.ID, azula)
	if n := len(bendEvents(g, game.EventFirebend, me.ID)); n != 0 {
		t.Fatalf("%d firebend events before the ability resolved", n)
	}
	for i := 0; i < 8 && (g.Stack.Size() > 0 || len(g.PendingChoices) > 0); i++ {
		effectstest.AnswerOpenOrderingIfAny(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := len(bendEvents(g, game.EventFirebend, me.ID)); n != 1 {
		t.Errorf("%d firebend events after the attack, want 1 (Roku's mana is not firebending)", n)
	}
}

// CR 701.65b: "airbend all other creatures" is one airbend, carrying the
// cards it exiled.
func TestBendingAvatarsWrathIsOneAirbend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushCatalogPermanent(g, me.ID, "Bear A", "Creature — Bear", "", false)
	b := effectstest.PushCatalogPermanent(g, opp.ID, "Bear B", "Creature — Bear", "", false)
	c := effectstest.PushCatalogPermanent(g, opp.ID, "Bear C", "Creature — Bear", "", false)
	effectstest.CastCatalogSpell(t, g, "Avatar's Wrath", "Sorcery", avatarsWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	evs := bendEvents(g, game.EventAirbend, me.ID)
	if len(evs) != 1 {
		t.Fatalf("%d airbend events, want 1", len(evs))
	}
	if len(evs[0].CardIDs) != 3 {
		t.Errorf("the airbend carries %d cards, want 3", len(evs[0].CardIDs))
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not airbent", id)
		}
	}
}
