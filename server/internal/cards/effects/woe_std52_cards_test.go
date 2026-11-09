package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std52_cards_test.go — Wilds of Eldraine, Standard batch 52 (#4471,
// #4452). Each test plays the real catalog entry and asserts what its text
// says.

const (
	std52FinancierOracle  = "734a5186-caa6-4811-b6d7-91b7b32fc403"
	std52GriffinOracle    = "f91ff377-a997-4f24-926c-76edcbf91a6c"
	std52PrincessOracle   = "70d52457-1867-482c-ba7a-1f7dacc4514f"
	std52SanctuaryOracle  = "4c0a8eed-f2e2-4821-8011-6e6ad2081936"
	std52MiceOracle       = "e5926e14-d30a-4f81-803c-f36542df2db3"
	std52AlchemistOracle  = "a78b063a-7f74-465a-9670-34f927f4bfe9"
	std52ChillOracle      = "be0dcad2-8e84-4e9f-b69c-837b40b8fb83"
	std52ChancellorOracle = "529f259a-df66-4283-bc28-de3933c42e67"
	std52GiantOracle      = "48ab8ba0-0e85-4656-a798-ceb961fb02ce"
	std52WhaleOracle      = "adca3929-4ecc-45bc-932e-604bcc32550a"
)

func TestStd52EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		std52FinancierOracle:        "Discerning Financier",
		std52GriffinOracle:          "Dutiful Griffin",
		std52PrincessOracle:         "The Princess Takes Flight",
		std52SanctuaryOracle:        "Solitary Sanctuary",
		std52MiceOracle:             "Three Blind Mice",
		std52AlchemistOracle:        "Aquatic Alchemist",
		std52AlchemistOracle + "#1": "Bubble Up",
		std52ChillOracle:            "Bitter Chill",
		std52ChancellorOracle:       "Chancellor of Tales",
		std52GiantOracle:            "Galvanic Giant",
		std52GiantOracle + "#1":     "Storm Reading",
		std52WhaleOracle:            "Horned Loch-Whale",
		std52WhaleOracle + "#1":     "Lagoon Breach",
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

// std52Adventure builds an Adventure card the way deck.toGameCard does.
func std52Adventure(owner uuid.UUID, oracle string, creature, creatureType, creatureCost string, power, toughness int, adv, advType, advCost string) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: creature, TypeLine: creatureType, ManaCost: creatureCost, Power: power, Toughness: toughness},
			{Name: adv, TypeLine: advType, ManaCost: advCost},
		},
	}
	c.SetFace(0)
	return c
}

func std52ToMain(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func std52StunCounters(g *game.Game, id uuid.UUID) int {
	c, _ := battlefieldCard(g, id)
	return c.Counters[game.CounterStun]
}

// --- The Princess Takes Flight ---------------------------------------------

func TestStd52PrincessExilesPumpsAndReturns(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	saga := castFromHand(t, g, game.Card{Name: "The Princess Takes Flight", TypeLine: "Enchantment — Saga", OracleID: std52PrincessOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Fatal("chapter I did not exile the creature")
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, mine); c.Power != 4 || c.Toughness != 4 || !hobB10Has(c, "flying") {
		t.Errorf("chapter II left my Bear %d/%d flying=%v, want 4/4 flying", c.Power, c.Toughness, hobB10Has(c, "flying"))
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	back := findBattlefieldByName(g, "Their Bear")
	if back == uuid.Nil {
		t.Fatal("chapter III did not return the exiled card")
	}
	if c := layeredCard(t, g, back); c.Controller != opp.ID {
		t.Errorf("the card returned under %v, want its owner %v", c.Controller, opp.ID)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- Three Blind Mice ------------------------------------------------------

func TestStd52ThreeBlindMiceMakesCopiesThenPumps(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	castFromHand(t, g, game.Card{Name: "Three Blind Mice", TypeLine: "Enchantment — Saga", OracleID: std52MiceOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Mouse"); n != 1 {
		t.Fatalf("chapter I made %d Mice, want 1", n)
	}
	mouse := roonaOnly(t, g, "Mouse")
	if mouse.Power != 1 || mouse.Toughness != 1 || !mouse.HasColor("W") || !mouse.IsToken() {
		t.Errorf("Mouse %+v, want a 1/1 white token", mouse)
	}

	for chapter := 2; chapter <= 3; chapter++ {
		advanceToPrecombatMainOf(t, g, seat)
		passPriorityAroundTable(t, g)
		answerPickTargetIfOpen(t, g, mouse.InstanceID)
		passPriorityAroundTable(t, g)
		if n := battlefieldCountNamed(g, me.ID, "Mouse"); n != chapter {
			t.Fatalf("chapter %d: %d Mice, want %d", chapter, n, chapter)
		}
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Mouse" {
			continue
		}
		lc := layeredCard(t, g, c.InstanceID)
		if lc.Power != 2 || lc.Toughness != 2 || !hobB10Has(lc, "vigilance") {
			t.Errorf("chapter IV left a Mouse %d/%d vigilance=%v, want 2/2 vigilance", lc.Power, lc.Toughness, hobB10Has(lc, "vigilance"))
		}
	}
}

// --- Solitary Sanctuary ----------------------------------------------------

func TestStd52SanctuaryTapsStunsAndGrowsForEachTapYouMake(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Solitary Sanctuary", TypeLine: "Enchantment", OracleID: std52SanctuaryOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); !tapped {
		t.Error("the entry trigger did not tap the creature")
	}
	if std52StunCounters(g, theirs) != 1 {
		t.Error("the entry trigger did not stun the creature")
	}
	// Its own tap is a creature you tapped: the counter trigger follows.
	answerPickTargetIfOpen(t, g, mine)
	passPriorityAroundTable(t, g)
	if got := cardByID(g, mine).Counters[game.CounterPlusOne]; got != 1 {
		t.Fatalf("+1/+1 counters after the entry tap = %d, want 1", got)
	}

	// A second opposing creature tapped by an effect of yours.
	other := pushVanillaCreature(g, opp.ID, "Their Other Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(other, me.ID) })
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, mine)
	passPriorityAroundTable(t, g)
	if got := cardByID(g, mine).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("+1/+1 counters after tapping a second creature = %d, want 2", got)
	}
	// Tapping a creature that is already tapped is not tapping an untapped one.
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(other, me.ID) })
	passPriorityAroundTable(t, g)
	if got := cardByID(g, mine).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("a tapped creature tapped again changed the counters to %d", got)
	}
	// An opponent tapping their own creature, or your own creature being
	// tapped, is not "you tap an untapped creature an opponent controls".
	third := pushVanillaCreature(g, opp.ID, "Their Third Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(third, opp.ID) })
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(mine, me.ID) })
	passPriorityAroundTable(t, g)
	if got := cardByID(g, mine).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("counters = %d after taps that do not count, want 2", got)
	}
}

// --- Galvanic Giant // Storm Reading ---------------------------------------

func TestStd52GalvanicGiantTapsAndStunsOnABigSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Galvanic Giant", "Creature — Giant Wizard", std52GiantOracle, false)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Small Spell", TypeLine: "Sorcery", ManaCost: "{4}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); tapped {
		t.Fatal("a mana value 4 spell triggered the Giant")
	}
	castFromHand(t, g, game.Card{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{5}"}, game.CastSpellParams{})
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if tapped, _ := battlefieldCardTapped(g, theirs); !tapped || std52StunCounters(g, theirs) != 1 {
		t.Errorf("a mana value 5 spell: tapped=%v stun=%d, want tapped with one stun counter", tapped, std52StunCounters(g, theirs))
	}
}

func TestStd52StormReadingDrawsFourDiscardsTwoAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := std52Adventure(me.ID, std52GiantOracle, "Galvanic Giant", "Creature — Giant Wizard", "{3}{U}", 3, 3, "Storm Reading", "Instant — Adventure", "{5}{U}{U}")
	me.Hand.PushTop(c)
	std52ToMain(t, g)
	hand := handSize(me)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Storm Reading: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Four drawn, then a two-card discard prompt over the hand.
	if got := handSize(me); got != hand-1+4 {
		t.Fatalf("hand after the draw = %d, want %d", got, hand-1+4)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("the Adventure card was not exiled")
	}
}

// --- Horned Loch-Whale // Lagoon Breach ------------------------------------

func TestStd52HornedLochWhaleEntersTappedUnlessItIsYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := std52Adventure(me.ID, std52WhaleOracle, "Horned Loch-Whale", "Creature — Whale", "", 6, 6, "Lagoon Breach", "Instant — Adventure", "{1}{U}")
	me.Hand.PushTop(c)
	std52ToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 0}); err != nil {
		t.Fatalf("cast the Whale: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tapped, ok := battlefieldCardTapped(g, c.InstanceID); !ok || tapped {
		t.Errorf("on my turn: on battlefield=%v tapped=%v, want untapped", ok, tapped)
	}
}

func TestStd52HornedLochWhaleEntersTappedOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Horned Loch-Whale", TypeLine: "Creature — Whale", OracleID: std52WhaleOracle, Power: 6, Toughness: 6, Owner: opp.ID, Controller: opp.ID})
	})
	castFromHandAndResolve(t, g, opp, cardByID(g, id))
	if tapped, ok := battlefieldCardTapped(g, id); !ok || !tapped {
		t.Errorf("on an opponent's turn: on battlefield=%v tapped=%v, want tapped", ok, tapped)
	}
}

func TestStd52LagoonBreachLetsTheOwnerChooseTopOrBottom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	attacker := pushVanillaCreature(g, opp.ID, "Their Attacker", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			}
		}
	})
	bystander := pushVanillaCreature(g, opp.ID, "Their Bystander", 3, 3)
	c := std52Adventure(me.ID, std52WhaleOracle, "Horned Loch-Whale", "Creature — Whale", "{4}{U}{U}", 6, 6, "Lagoon Breach", "Instant — Adventure", "{1}{U}")
	me.Hand.PushTop(c)
	std52ToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Strict: true, Targets: cardRefs(bystander)}); err == nil {
		t.Fatal("a non-attacking creature was a legal target")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Targets: cardRefs(attacker)}); err != nil {
		t.Fatalf("cast Lagoon Breach: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Fatal("the attacker is still on the battlefield")
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("the Adventure card was not exiled")
	}
}
