package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82g22GangLeader = "79d99305-b3dd-4687-af69-bf4b28704aae"
	s82g22Lackey     = "213fbbdd-f1bc-4bba-9124-7b26e9a706c8"
	s82g22Guide      = "51d9564b-44fc-4de1-9119-09d7b4089378"
	s82g22Rootwalla  = "d43f5977-0532-4653-bc57-ce29f9a531c9"
	s82g22Coatl      = "f4b1bbd1-f1b2-4c7d-bafd-9d2857f92a89"
	s82g22Bolshack   = "c8714939-aaf9-4ede-a04f-61f0df5ea4d4"
	s82g22Gobland    = "74900080-80a6-4fd8-9557-5f773c0e4bec"
)

func TestS82GapB22EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82g22GangLeader: "Goblin Gang Leader", s82g22Lackey: "Goblin Lackey", s82g22Guide: "Goblin Guide",
		s82g22Rootwalla: "Blazing Rootwalla", s82g22Coatl: "Ice-Fang Coatl", s82g22Gobland: "Gobland", s82g22Bolshack: "Bolshack Dragon",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB22GoblinGangLeaderMakesTwoGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Goblin Gang Leader", "Creature — Goblin Warrior", s82g22GangLeader, nil)
	passPriorityAroundTable(t, g) // the creature spell
	passPriorityAroundTable(t, g) // the trigger
	if got := battlefieldCountNamed(g, me.ID, "Goblin"); got != 2 {
		t.Errorf("%d Goblin tokens, want 2", got)
	}
}

func TestS82GapB22GoblinLackeyPutsAGoblinFromHandOnAnyDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lackey := iceB1Push(g, me.ID, "Goblin Lackey", "Creature — Goblin", s82g22Lackey, 1, 1)
	goblin := handCardForTest(me, "Goblin Piker", "Creature — Goblin", "")
	bear := handCardForTest(me, "Bear", "Creature — Bear", "")
	sorcery := handCardForTest(me, "Goblin Charge", "Sorcery — Goblin", "")

	// Damage to a creature does nothing.
	other := iceB1Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	dealNoncombatDamage(g, me.ID, lackey, other)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked after damage to a creature")
	}

	// Noncombat damage to a player triggers it; only the Goblin permanent card is offered.
	dealNoncombatDamage(g, me.ID, lackey, opp.ID)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("not asked: %+v", g.PendingChoices)
	}
	for _, id := range pick.ChooseCards {
		if id == bear || id == sorcery {
			t.Errorf("offered %s, which is not a Goblin permanent card", id)
		}
	}
	answerChooseCards(t, g, me.ID, goblin)
	if !g.Battlefield.Contains(goblin) {
		t.Error("the Goblin did not enter")
	}
}

func TestS82GapB22GoblinGuideGivesTheDefenderTheirTopLand(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		wantHand int
	}{{"land", "Basic Land — Forest", 1}, {"nonland", "Creature — Bear", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			guide := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin Guide",
				TypeLine: "Creature — Goblin Scout", OracleID: s82g22Guide, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
				Keywords: []string{"haste"}})
			top := uuid.New()
			opp.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: tc.typeLine, Owner: opp.ID, Controller: opp.ID})
			handBefore, mine := len(opp.Hand.Cards), len(me.Hand.Cards)
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(guide, opp.ID); err != nil {
				t.Fatal(err)
			}
			lockInAttacks(t, g)
			passPriorityAroundTable(t, g)
			if got := len(opp.Hand.Cards) - handBefore; got != tc.wantHand {
				t.Errorf("defender's hand changed by %d, want %d", got, tc.wantHand)
			}
			if topCard, err := opp.Library.Top(); tc.wantHand == 0 && (err != nil || topCard.InstanceID != top) {
				t.Error("a nonland card did not stay on top of the library")
			}
			if len(me.Hand.Cards) != mine {
				t.Error("the attacker's controller gained a card")
			}
		})
	}
}

func TestS82GapB22BlazingRootwallaPumpsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	walla := iceB1Push(g, me.ID, "Blazing Rootwalla", "Creature — Lizard", s82g22Rootwalla, 1, 1)
	dcMana(t, g, me.ID, "{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, walla); p != 3 || tough != 1 {
		t.Errorf("Rootwalla is %d/%d, want 3/1", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, walla, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time in one turn")
	}
	if spec, _ := Lookup(s82g22Rootwalla); spec.Madness != "{0}" {
		t.Errorf("madness %q, want {0}", spec.Madness)
	}
}

func TestS82GapB22IceFangCoatlHasDeathtouchWithThreeOtherSnowPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	coatl := iceB1Push(g, me.ID, "Ice-Fang Coatl", "Snow Creature — Snake", s82g22Coatl, 1, 1, "flash", "flying")
	if hasEffectiveKeyword(t, g, coatl, "deathtouch") {
		t.Fatal("deathtouch with no other snow permanent")
	}
	var snow []uuid.UUID
	for i := 0; i < 3; i++ {
		snow = append(snow, iceB1Push(g, me.ID, "Snow Land", "Snow Land", "", 0, 0))
		want := i == 2
		if got := hasEffectiveKeyword(t, g, coatl, "deathtouch"); got != want {
			t.Errorf("with %d other snow permanents deathtouch = %v, want %v", i+1, got, want)
		}
	}
	// A nonsnow permanent doesn't count; losing one ends it.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(snow[0]) })
	if hasEffectiveKeyword(t, g, coatl, "deathtouch") {
		t.Error("deathtouch with only two other snow permanents")
	}
}

func TestS82GapB22IceFangCoatlDrawsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	before := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Ice-Fang Coatl", "Snow Creature — Snake", s82g22Coatl, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 1 {
		t.Errorf("hand changed by %d, want 1 (cast from hand then draw)", got)
	}
}

func TestS82GapB22GoblandCantBlockAndIsASummoningSickManaSource(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Gobland", "Land Creature — Mountain Goblin", s82g22Gobland, true)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Gobland tapped for mana")
	}
	ready := pushCatalogPermanent(g, me.ID, "Gobland", "Land Creature — Mountain Goblin", s82g22Gobland, false)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want R", got)
	}
}

func TestS82GapB22BolshackDragonGainsPowerFromRedGraveyardCardsOnlyWhileAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := iceB1Push(g, me.ID, "Bolshack Dragon", "Creature — Armored Dragon", s82g22Bolshack, 6, 6, "double strike", "haste")
	for _, colors := range [][]string{{"R"}, {"R", "G"}, {"U"}, nil} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Instant", Colors: colors, Owner: me.ID, Controller: me.ID})
	}
	// An opponent's red card doesn't count.
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Instant", Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	effPower := func() int { return layeredCard(t, g, dragon).Effective().Power }
	if p := effPower(); p != 6 {
		t.Errorf("power %d while not attacking, want 6", p)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dragon, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	if p := effPower(); p != 8 {
		t.Errorf("attacking Dragon has power %d, want 8 (two red cards in my graveyard)", p)
	}
	more := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: more, Name: "More", TypeLine: "Instant", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, more); err != nil {
		t.Fatal(err)
	}
	if p := effPower(); p != 9 {
		t.Errorf("attacking Dragon has power %d after a third red card, want 9", p)
	}
}

func TestS82GapB22GoblandCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Keywords: []string{"haste"}})
	land := pushCatalogPermanent(g, opp.ID, "Gobland", "Land Creature — Mountain Goblin", s82g22Gobland, false)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(land, attacker); err == nil {
		t.Error("Gobland blocked")
	}
}
