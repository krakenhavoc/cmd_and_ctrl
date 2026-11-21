package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardspr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/pr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prowl_stoic_strategist_test.go — #2651: the ability half of CR 607.2a's
// "exiled with" link, through Prowl, Stoic Strategist.

func prowlRow() cards.Card {
	return effectstest.TransformerRow(cardspr.ProwlOracleID,
		effectstest.DisturbFace{Name: "Prowl, Stoic Strategist", TypeLine: "Legendary Artifact Creature — Robot", Cost: "{3}{W}", Power: "3", Toughness: "3"},
		effectstest.DisturbFace{Name: "Prowl, Pursuit Vehicle", TypeLine: "Legendary Artifact — Vehicle", Cost: "", Power: "2", Toughness: "3"},
		"More Than Meets the Eye {2}{W}\nWhenever Prowl attacks, exile up to one other target tapped creature or Vehicle. For as long as that card remains exiled, its owner may play it.\nWhenever a player plays a card exiled with Prowl, you draw a card and convert Prowl.",
		"Living metal\nWhenever another creature or Vehicle you control enters, put a +1/+1 counter on Prowl. If this is the second time this ability has resolved this turn, convert Prowl.",
		"Living metal", "Convert")
}

// prowlTapped puts a tapped creature on the battlefield for `owner`.
func prowlTapped(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, owner, name, typeLine, "", false)
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, id); c != nil {
			c.Tapped = true
		}
	})
	return id
}

// prowlExiledCard reads a card in exile by instance ID.
func prowlExiledCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("%s is not in exile", id)
	return game.Card{}
}

// prowlAttacksExiling attacks with Prowl and answers its trigger with
// `target`, then lets everything resolve.
func prowlAttacksExiling(t *testing.T, g *game.Game, prowl, defender, target uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(prowl, defender); err != nil {
		t.Fatalf("Prowl can't attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTarget(t, g, target)
	effectstest.TfSettle(t, g)
}

// The attack trigger exiles the tapped creature linked to this Prowl
// OBJECT and to the attack trigger's catalog row, and its owner may play
// it while it remains exiled.
func TestProwlExilesATappedCreatureLinkedToItsAttackTrigger(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	prowl := effectstest.TransformerIn(t, g, prowlRow(), me)
	victim := prowlTapped(g, opp.ID, "Tapped Bear", "Creature — Bear")

	prowlAttacksExiling(t, g, prowl, opp.ID, victim)

	c := prowlExiledCard(t, g, victim)
	src := effectstest.FlipState(t, g, prowl)
	if want := (game.PermissionCardRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}); c.ExiledWith != want {
		t.Errorf("ExiledWith = %+v, want %+v", c.ExiledWith, want)
	}
	if !c.ExiledWithAbility.IsRow(game.AbilitySlotTriggered, cardspr.ProwlExileLabel) {
		t.Errorf("ExiledWithAbility = %+v, want the attack trigger's row", c.ExiledWithAbility)
	}
	g.WithWriteLock(func() {
		got := g.ExiledWithRowForEffect(c.ExiledWith, game.AbilitySlotTriggered, cardspr.ProwlExileLabel)
		if len(got) != 1 || got[0].InstanceID != victim {
			t.Errorf("ExiledWithRowForEffect = %d cards, want the victim", len(got))
		}
		if other := g.ExiledWithRowForEffect(c.ExiledWith, game.AbilitySlotTriggered, cardspr.ProwlPlayLabel); len(other) != 0 {
			t.Errorf("another row of the same object reads %d cards, want none (CR 607.2a)", len(other))
		}
	})
	perm, ok := effectstest.AirbendGrantOn(g, victim)
	if !ok || perm.Player != opp.ID || perm.CastOnly {
		t.Errorf("permission = %+v, want the owner may PLAY it", perm)
	}
}

// "Other": Prowl can't exile itself, and an untapped creature is not a
// target.
func TestProwlTargetsOnlyAnotherTappedCreatureOrVehicle(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	prowl := effectstest.TransformerIn(t, g, prowlRow(), me)
	untapped := effectstest.PushCatalogPermanent(g, opp.ID, "Untapped Bear", "Creature — Bear", "", false)
	vehicle := prowlTapped(g, opp.ID, "Tapped Cart", "Artifact — Vehicle")

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(prowl, opp.ID); err != nil {
		t.Fatalf("Prowl can't attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{prowl, untapped} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as Prowl's target", bad)
		}
	}
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: vehicle}); err != nil {
		t.Fatalf("a tapped noncreature Vehicle is a legal target: %v", err)
	}
	effectstest.TfSettle(t, g)
	prowlExiledCard(t, g, vehicle)
}

// Casting the exiled card from exile is playing a card exiled with
// Prowl: its controller draws a card and Prowl converts.
func TestProwlDrawsAndConvertsWhenTheExiledCardIsCast(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	prowl := effectstest.TransformerIn(t, g, prowlRow(), me)
	mine := prowlTapped(g, me.ID, "My Tapped Bear", "Creature — Bear")
	effectstest.SeedLibrary(me, "Card A", "Card B", "Card C")

	prowlAttacksExiling(t, g, prowl, opp.ID, mine)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	hand := effectstest.HandSize(me)

	if err := effectstest.CastFromExile(t, g, me.ID, mine); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.TfSettle(t, g)

	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand = %d, want %d (Prowl draws a card)", got, hand+1)
	}
	effectstest.WantFace(t, g, prowl, 1, "Prowl, Pursuit Vehicle")
	if !g.Battlefield.Contains(mine) {
		t.Error("the cast creature never resolved")
	}
}

// A land played from exile is played too.
func TestProwlTriggersOnALandPlayedFromExile(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	prowl := effectstest.TransformerIn(t, g, prowlRow(), me)
	arbor := prowlTapped(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad")
	effectstest.SeedLibrary(me, "Card A", "Card B", "Card C")

	prowlAttacksExiling(t, g, prowl, opp.ID, arbor)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	hand := effectstest.HandSize(me)

	if err := g.CastSpell(me.ID, arbor, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	effectstest.TfSettle(t, g)

	if !g.Battlefield.Contains(arbor) {
		t.Fatal("the land was not played")
	}
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand = %d, want %d (Prowl draws a card)", got, hand+1)
	}
	effectstest.WantFace(t, g, prowl, 1, "Prowl, Pursuit Vehicle")
}

// CR 607.2a: a card linked to Prowl by anything but its attack trigger,
// or to another object, is not "exiled with Prowl" for this ability.
func TestProwlIgnoresCardsItsAttackTriggerDidNotExile(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	prowl := effectstest.TransformerIn(t, g, prowlRow(), me)
	effectstest.SeedLibrary(me, "Card A", "Card B", "Card C")
	src := effectstest.FlipState(t, g, prowl)
	self := game.PermissionCardRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}

	cases := []struct {
		name    string
		with    game.PermissionCardRef
		ability game.AbilityRef
	}{
		{"a replacement's link (no ability)", self, game.AbilityRef{}},
		{"another row of Prowl", self, game.AbilityRef{Key: cardspr.ProwlOracleID, Slot: game.AbilitySlotTriggered, Ref: "own:1", Name: cardspr.ProwlPlayLabel}},
		{"another Prowl object (CR 400.7)", game.PermissionCardRef{ID: self.ID, Epoch: self.Epoch + 1}, game.AbilityRef{Key: cardspr.ProwlOracleID, Slot: game.AbilitySlotTriggered, Ref: "own:0", Name: cardspr.ProwlExileLabel}},
		{"another permanent", game.PermissionCardRef{ID: uuid.New(), Epoch: 1}, game.AbilityRef{Key: cardspr.ProwlOracleID, Slot: game.AbilitySlotTriggered, Ref: "own:0", Name: cardspr.ProwlExileLabel}},
	}
	for _, tc := range cases {
		id := uuid.New()
		g.WithWriteLock(func() {
			g.Exile.PushTop(game.Card{
				InstanceID: id, Name: "Exiled Instant", TypeLine: "Instant",
				Owner: me.ID, Controller: me.ID, ExiledWith: tc.with, ExiledWithAbility: tc.ability,
			})
			g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{Player: me.ID, Duration: game.WhileInZoneDuration()})
		})
		hand := effectstest.HandSize(me)
		if err := effectstest.CastFromExile(t, g, me.ID, id); err != nil {
			t.Fatalf("%s: cast from exile: %v", tc.name, err)
		}
		effectstest.TfSettle(t, g)
		if got := effectstest.HandSize(me); got != hand {
			t.Errorf("%s: hand = %d, want %d (no trigger)", tc.name, got, hand)
		}
		effectstest.WantFace(t, g, prowl, 0, "Prowl, Stoic Strategist")
	}
}

// Back face: a +1/+1 counter for each other creature or Vehicle you
// control that enters, and the second resolution this turn converts it.
func TestProwlPursuitVehicleCountsTwoEntriesThenConverts(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	prowl := effectstest.TransformerConverted(t, g, prowlRow(), me)

	enter := func(name, typeLine string) {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() {
			if _, err := g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{}); err != nil {
				t.Fatalf("put %s: %v", name, err)
			}
		})
		effectstest.TfSettle(t, g)
	}

	enter("A Rock", "Artifact")
	if n := effectstest.CounterCount(g, prowl, "+1/+1"); n != 0 {
		t.Fatalf("a noncreature artifact put %d counters on Prowl", n)
	}
	enter("First Bear", "Creature — Bear")
	if n := effectstest.CounterCount(g, prowl, "+1/+1"); n != 1 {
		t.Fatalf("+1/+1 counters = %d after one creature, want 1", n)
	}
	effectstest.WantFace(t, g, prowl, 1, "Prowl, Pursuit Vehicle")
	enter("A Cart", "Artifact — Vehicle")
	if n := effectstest.CounterCount(g, prowl, "+1/+1"); n != 2 {
		t.Fatalf("+1/+1 counters = %d after a Vehicle, want 2", n)
	}
	effectstest.WantFace(t, g, prowl, 0, "Prowl, Stoic Strategist")
}
