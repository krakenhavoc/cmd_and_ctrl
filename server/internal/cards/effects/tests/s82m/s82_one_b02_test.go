package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82SkullbombTypeLine = "Artifact"
	s82SphereTypeLine    = "Land — Sphere"
)

func s82Hand(p *game.Player) int { return p.Hand.Size() }

func s82Bears(g *game.Game, p *game.Player, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
	})
}

func s82Skullbomb(g *game.Game, p *game.Player, oracle, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: s82SkullbombTypeLine, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
}

func TestS82OneSkullbombsPlainDrawAbility(t *testing.T) {
	for _, c := range []struct{ oracle, name string }{
		{"43057d84-b424-489c-abec-3fb4cb823d80", "Basilica Skullbomb"},
		{"0c43e9b4-013a-47ee-9805-f41afb0a3bde", "Dross Skullbomb"},
		{"77e0c121-724c-43a0-84e3-1800e2d94a34", "Furnace Skullbomb"},
		{"3cb91e96-285d-4191-97cd-ee7ce5d86eb2", "Maze Skullbomb"},
		{"6d364283-617a-4661-865d-abd9c304b2f9", "Surgical Skullbomb"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := s82Skullbomb(g, me, c.oracle, c.name)
			effectstest.AdvanceToMain(t, g)
			before := s82Hand(me)
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("activated with no mana")
			}
			effectstest.B06AddMana(me, "C")
			effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
			effectstest.DrcSettle(t, g, me.ID)
			if _, ok := effectstest.BattlefieldCard(g, id); ok {
				t.Error("the Skullbomb stayed on the battlefield")
			}
			if got := s82Hand(me); got != before+1 {
				t.Errorf("hand %d, want %d", got, before+1)
			}
		})
	}
}

func TestS82OneBasilicaSkullbombPumpsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := s82Skullbomb(g, me, "43057d84-b424-489c-abec-3fb4cb823d80", "Basilica Skullbomb")
	bear := s82Bears(g, me, "Bear")
	foe := s82Bears(g, g.Seats[1], "Foe")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
	before := s82Hand(me)
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.DrcSettle(t, g, me.ID)
	c := effectstest.BattlefieldCardCopy(t, g, bear)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "flying") {
		t.Errorf("bear is %d/%d flying=%v, want 4/4 flying", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
	}
	if got := s82Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestS82OneMazeSkullbombPumpsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := s82Skullbomb(g, me, "3cb91e96-285d-4191-97cd-ee7ce5d86eb2", "Maze Skullbomb")
	bear := s82Bears(g, me, "Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "G")
	before := s82Hand(me)
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.DrcSettle(t, g, me.ID)
	c := effectstest.BattlefieldCardCopy(t, g, bear)
	if c.CurrentPower() != 5 || c.CurrentToughness() != 5 || !game.HasKeyword(&c, "trample") {
		t.Errorf("bear is %d/%d, want 5/5 trample", c.CurrentPower(), c.CurrentToughness())
	}
	if got := s82Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestS82OneDrossSkullbombReturnsACreatureCardAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := s82Skullbomb(g, me, "0c43e9b4-013a-47ee-9805-f41afb0a3bde", "Dross Skullbomb")
	dead := effectstest.DrcGraveyardCard(me, "Dead Bear", effectstest.DrcCreatureTypeLine)
	theirs := effectstest.DrcGraveyardCard(g.Seats[1], "Their Bear", effectstest.DrcCreatureTypeLine)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "B")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("targeted a card in an opponent's graveyard")
	}
	before := s82Hand(me)
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}})
	effectstest.DrcSettle(t, g, me.ID)
	if got := effectstest.B12ZoneOf(g, dead); got != game.ZoneHand {
		t.Errorf("the creature card is in %q, want hand", got)
	}
	if got := s82Hand(me); got != before+2 {
		t.Errorf("hand %d, want %d (the card and the draw)", got, before+2)
	}
}

func TestS82OneSurgicalSkullbombBouncesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := s82Skullbomb(g, me, "6d364283-617a-4661-865d-abd9c304b2f9", "Surgical Skullbomb")
	foe := s82Bears(g, g.Seats[1], "Foe")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "U")
	before := s82Hand(me)
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)})
	effectstest.DrcSettle(t, g, me.ID)
	if got := effectstest.B12ZoneOf(g, foe); got != game.ZoneHand {
		t.Errorf("the creature is in %q, want hand", got)
	}
	if got := s82Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestS82OneFurnaceSkullbombPutsTwoOilCountersAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := s82Skullbomb(g, me, "77e0c121-724c-43a0-84e3-1800e2d94a34", "Furnace Skullbomb")
	bear := s82Bears(g, me, "Bear")
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land)}); err == nil {
		t.Fatal("targeted a land that is neither an artifact nor a creature")
	}
	before := s82Hand(me)
	effectstest.B16Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.DrcOilOn(t, g, bear); n != 2 {
		t.Errorf("%d oil counters, want 2", n)
	}
	if got := s82Hand(me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestS82OneSkullbombSecondAbilityIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := s82Skullbomb(g, me, "6d364283-617a-4661-865d-abd9c304b2f9", "Surgical Skullbomb")
	foe := s82Bears(g, g.Seats[1], "Foe")
	effectstest.B06AddMana(me, "C", "C", "U")
	// The harness opens on the draw step: not a main phase.
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err == nil {
		t.Fatal("activated at instant speed")
	}
}

func TestS82OneSphereLands(t *testing.T) {
	for _, c := range []struct{ oracle, name, colour string }{
		{"e2dd05d5-312e-47f1-873e-c0741ee6ef4a", "The Autonomous Furnace", "R"},
		{"8110fe69-c66c-4e2c-86ee-dcc8dc9a13d1", "The Dross Pits", "B"},
		{"8618f8e8-fd86-47e3-905d-e6624b599b9b", "The Fair Basilica", "W"},
		{"b0ea4975-a944-4585-8711-60f203cffa4a", "The Hunter Maze", "G"},
		{"db9c155f-b342-41bb-9e1a-50358fb9f40e", "The Surgical Bay", "U"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.B31PlayLand(t, g, c.name, s82SphereTypeLine, c.oracle)
			if !effectstest.B31Tapped(t, g, id) {
				t.Fatal("it did not enter tapped")
			}
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
				t.Fatal("tapped a tapped land for mana")
			}
			g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = false })
			if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			if n := me.ManaPool.CountColor(c.colour); n != 1 {
				t.Errorf("pool has %d %s, want 1", n, c.colour)
			}
			g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = false })
			me.ManaPool = nil
			effectstest.B06AddMana(me, "C")
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("drew without paying the coloured mana")
			}
			effectstest.B06AddMana(me, c.colour)
			before := s82Hand(me)
			effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
			effectstest.DrcSettle(t, g, me.ID)
			if _, ok := effectstest.BattlefieldCard(g, id); ok {
				t.Error("the land was not sacrificed")
			}
			if got := s82Hand(me); got != before+1 {
				t.Errorf("hand %d, want %d", got, before+1)
			}
		})
	}
}

func TestS82OneMonumentalFacade(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B31PlayLand(t, g, "The Monumental Facade", s82SphereTypeLine, "63e8d282-d038-4a8c-a0cb-f51ddf87d8ea")
	if n := effectstest.DrcOilOn(t, g, id); n != 2 {
		t.Fatalf("entered with %d oil counters, want 2", n)
	}
	if effectstest.B31Tapped(t, g, id) {
		t.Fatal("it entered tapped")
	}
	bear := s82Bears(g, me, "Bear")
	foe := s82Bears(g, g.Seats[1], "Foe")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err == nil {
		t.Fatal("targeted an opponent's creature")
	}
	if n := effectstest.DrcOilOn(t, g, id); n != 2 {
		t.Fatalf("a refused activation spent counters: %d", n)
	}
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.DrcOilOn(t, g, bear); n != 1 {
		t.Errorf("bear has %d oil counters, want 1", n)
	}
	if n := effectstest.DrcOilOn(t, g, id); n != 1 {
		t.Errorf("the land has %d oil counters, want 1", n)
	}
	// Tapped now: the second counter needs an untap.
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = false })
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.DrcSettle(t, g, me.ID)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = false })
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("activated with no oil counters left")
	}
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = false })
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} ability: %v", err)
	}
	if n := me.ManaPool.CountColor("C"); n != 1 {
		t.Errorf("pool has %d {C}, want 1", n)
	}
}

func TestS82OneMirrex(t *testing.T) {
	const oracle = "5502741a-e3b9-454e-8121-4360a6db6750"
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B31PlayLand(t, g, "Mirrex", s82SphereTypeLine, oracle)
	if err := g.ActivateManaAbility(me.ID, id, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("any-colour mana the turn it entered: %v", err)
	}
	s82AnswerColour(t, g, me, "U")
	if me.ManaPool.Total() != 1 || me.ManaPool.CountColor("U") != 1 {
		t.Errorf("pool has %d mana, want one {U}", me.ManaPool.Total())
	}

	// A Mirrex that did not enter this turn only makes {C}.
	old := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: old, Name: "Mirrex", TypeLine: s82SphereTypeLine, OracleID: oracle,
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(me.ID, old, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("any-colour mana from a land that did not enter this turn")
	}
	if err := g.ActivateManaAbility(me.ID, old, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the {C} ability: %v", err)
	}
}

func TestS82OneMirrexMakesAMite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mirrex", TypeLine: s82SphereTypeLine, OracleID: "5502741a-e3b9-454e-8121-4360a6db6750",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for two mana")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.B16CountNamed(g, "Phyrexian Mite"); n != 1 {
		t.Fatalf("%d Phyrexian Mites, want 1", n)
	}
	if !effectstest.B16Tapped(t, g, id) {
		t.Error("Mirrex did not tap")
	}
	var mite game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Mite" {
			mite = c
		}
	}
	if game.ToxicTotal(&mite) != 1 || mite.CurrentPower() != 1 || mite.CurrentToughness() != 1 {
		t.Errorf("the Mite is %d/%d toxic %d", mite.CurrentPower(), mite.CurrentToughness(), game.ToxicTotal(&mite))
	}
}

func TestS82OneMyrConvertPaysTwoLifeForAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Myr Convert", TypeLine: "Artifact Creature — Phyrexian Myr",
		OracleID: "6abf433e-27da-4abb-b262-3af792a1db23", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	if got := game.ToxicTotal(ptrCard(effectstest.BattlefieldCardCopy(t, g, id))); got != 1 {
		t.Errorf("toxic %d, want 1", got)
	}
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).SummonedThisTurn = false })
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	s82AnswerColour(t, g, me, "G")
	if me.ManaPool.Total() != 1 || me.ManaPool.CountColor("G") != 1 {
		t.Errorf("pool has %d mana, want 1", me.ManaPool.Total())
	}
	if !effectstest.B16Tapped(t, g, id) {
		t.Error("it did not tap")
	}

	// At 1 life the cost can't be paid (CR 119.4).
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Tapped = false })
	me.Life = 1
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("paid 2 life at 1 life")
	}
	if me.Life != 1 {
		t.Errorf("a refused activation cost life: %d", me.Life)
	}
}

func ptrCard(c game.Card) *game.Card { return &c }

// s82AnswerColour answers the any-colour pick an activation just opened.
func s82AnswerColour(t *testing.T, g *game.Game, p *game.Player, colour string) {
	t.Helper()
	pick := effectstest.RiderLatestManaPick(g, p.ID)
	if pick == nil {
		t.Fatal("no colour pick was opened")
	}
	if err := g.ResolveManaChoice(pick.ID, p.ID, colour); err != nil {
		t.Fatalf("ResolveManaChoice %s: %v", colour, err)
	}
}
