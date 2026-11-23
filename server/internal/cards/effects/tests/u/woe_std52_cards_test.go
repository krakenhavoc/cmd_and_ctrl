package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.Counters[game.CounterStun]
}

// --- The Princess Takes Flight ---------------------------------------------

func TestStd52PrincessExilesPumpsAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	saga := effectstest.CastFromHand(t, g, game.Card{Name: "The Princess Takes Flight", TypeLine: "Enchantment — Saga", OracleID: std52PrincessOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Fatal("chapter I did not exile the creature")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if c, p, th := effectstest.LayeredCard(t, g, mine), effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 4 || th != 4 || !effectstest.HobB10Has(c, "flying") {
		t.Errorf("chapter II left my Bear %d/%d flying=%v, want 4/4 flying", p, th, effectstest.HobB10Has(c, "flying"))
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.FindBattlefieldByName(g, "Their Bear")
	if back == uuid.Nil {
		t.Fatal("chapter III did not return the exiled card")
	}
	if c := effectstest.LayeredCard(t, g, back); c.Controller != opp.ID {
		t.Errorf("the card returned under %v, want its owner %v", c.Controller, opp.ID)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- Three Blind Mice ------------------------------------------------------

func TestStd52ThreeBlindMiceMakesCopiesThenPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.CastFromHand(t, g, game.Card{Name: "Three Blind Mice", TypeLine: "Enchantment — Saga", OracleID: std52MiceOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Mouse"); n != 1 {
		t.Fatalf("chapter I made %d Mice, want 1", n)
	}
	mouse := effectstest.RoonaOnly(t, g, "Mouse")
	if mouse.Power != 1 || mouse.Toughness != 1 || !mouse.HasColor("W") || !mouse.IsToken() {
		t.Errorf("Mouse %+v, want a 1/1 white token", mouse)
	}

	for chapter := 2; chapter <= 3; chapter++ {
		effectstest.AdvanceToPrecombatMainOf(t, g, seat)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPickTargetIfOpen(t, g, mouse.InstanceID)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.BattlefieldCountNamed(g, me.ID, "Mouse"); n != chapter {
			t.Fatalf("chapter %d: %d Mice, want %d", chapter, n, chapter)
		}
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Mouse" {
			continue
		}
		lc := effectstest.LayeredCard(t, g, c.InstanceID)
		if p, th := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID); p != 2 || th != 2 || !effectstest.HobB10Has(lc, "vigilance") {
			t.Errorf("chapter IV left a Mouse %d/%d vigilance=%v, want 2/2 vigilance", p, th, effectstest.HobB10Has(lc, "vigilance"))
		}
	}
}

// --- Solitary Sanctuary ----------------------------------------------------

func TestStd52SanctuaryTapsStunsAndGrowsForEachTapYouMake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.CastFromHand(t, g, game.Card{Name: "Solitary Sanctuary", TypeLine: "Enchantment", OracleID: std52SanctuaryOracle, ManaCost: "{2}{W}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, theirs); !tapped {
		t.Error("the entry trigger did not tap the creature")
	}
	if std52StunCounters(g, theirs) != 1 {
		t.Error("the entry trigger did not stun the creature")
	}
	// Its own tap is a creature you tapped: the counter trigger follows.
	effectstest.AnswerPickTargetIfOpen(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardByID(g, mine).Counters[game.CounterPlusOne]; got != 1 {
		t.Fatalf("+1/+1 counters after the entry tap = %d, want 1", got)
	}

	// A second opposing creature tapped by an effect of yours.
	other := effectstest.PushVanillaCreature(g, opp.ID, "Their Other Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(other, me.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardByID(g, mine).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("+1/+1 counters after tapping a second creature = %d, want 2", got)
	}
	// Tapping a creature that is already tapped is not tapping an untapped one.
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(other, me.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardByID(g, mine).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("a tapped creature tapped again changed the counters to %d", got)
	}
	// An opponent tapping their own creature, or your own creature being
	// tapped, is not "you tap an untapped creature an opponent controls".
	third := effectstest.PushVanillaCreature(g, opp.ID, "Their Third Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(third, opp.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffectBy(mine, me.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardByID(g, mine).Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("counters = %d after taps that do not count, want 2", got)
	}
}

// --- Galvanic Giant // Storm Reading ---------------------------------------

func TestStd52GalvanicGiantTapsAndStunsOnABigSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Galvanic Giant", "Creature — Giant Wizard", std52GiantOracle, false)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.CastFromHand(t, g, game.Card{Name: "Small Spell", TypeLine: "Sorcery", ManaCost: "{4}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, theirs); tapped {
		t.Fatal("a mana value 4 spell triggered the Giant")
	}
	effectstest.CastFromHand(t, g, game.Card{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{5}"}, game.CastSpellParams{})
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, theirs); !tapped || std52StunCounters(g, theirs) != 1 {
		t.Errorf("a mana value 5 spell: tapped=%v stun=%d, want tapped with one stun counter", tapped, std52StunCounters(g, theirs))
	}
}

func TestStd52StormReadingDrawsFourDiscardsTwoAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := std52Adventure(me.ID, std52GiantOracle, "Galvanic Giant", "Creature — Giant Wizard", "{3}{U}", 3, 3, "Storm Reading", "Instant — Adventure", "{5}{U}{U}")
	me.Hand.PushTop(c)
	std52ToMain(t, g)
	hand := effectstest.HandSize(me)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Storm Reading: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Four drawn, then a two-card discard prompt over the hand.
	if got := effectstest.HandSize(me); got != hand-1+4 {
		t.Fatalf("hand after the draw = %d, want %d", got, hand-1+4)
	}
	var pitch []uuid.UUID
	g.ReadSnapshot(func() {
		for _, h := range me.Hand.Cards[:2] {
			pitch = append(pitch, h.InstanceID)
		}
	})
	effectstest.AnswerDiscard(t, g, me.ID, pitch...)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != hand-1+4-2 {
		t.Fatalf("hand after the discard = %d, want %d", got, hand-1+4-2)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("the Adventure card was not exiled")
	}
}

// --- Horned Loch-Whale // Lagoon Breach ------------------------------------

func TestStd52HornedLochWhaleEntersTappedUnlessItIsYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := std52Adventure(me.ID, std52WhaleOracle, "Horned Loch-Whale", "Creature — Whale", "{4}{U}{U}", 6, 6, "Lagoon Breach", "Instant — Adventure", "{1}{U}")
	std52ToMain(t, g)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if tapped, ok := effectstest.BattlefieldCardTapped(g, c.InstanceID); !ok || tapped {
		t.Errorf("on my turn: on battlefield=%v tapped=%v, want untapped", ok, tapped)
	}
}

func TestStd52HornedLochWhaleEntersTappedOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Horned Loch-Whale", TypeLine: "Creature — Whale", OracleID: std52WhaleOracle, Power: 6, Toughness: 6, Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.CastFromHandAndResolve(t, g, opp, effectstest.CardByID(g, id))
	if tapped, ok := effectstest.BattlefieldCardTapped(g, id); !ok || !tapped {
		t.Errorf("on an opponent's turn: on battlefield=%v tapped=%v, want tapped", ok, tapped)
	}
}

func TestStd52LagoonBreachLetsTheOwnerChooseTopOrBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, opp.ID, "Their Attacker", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attacker {
				g.Battlefield.Cards[i].AttackingTarget = me.ID
			}
		}
	})
	bystander := effectstest.PushVanillaCreature(g, opp.ID, "Their Bystander", 3, 3)
	c := std52Adventure(me.ID, std52WhaleOracle, "Horned Loch-Whale", "Creature — Whale", "{4}{U}{U}", 6, 6, "Lagoon Breach", "Instant — Adventure", "{1}{U}")
	me.Hand.PushTop(c)
	std52ToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Strict: true, Targets: effectstest.CardRefs(bystander)}); err == nil {
		t.Fatal("a non-attacking creature was a legal target")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.CardRefs(attacker)}); err != nil {
		t.Fatalf("cast Lagoon Breach: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if pc == nil || pc.LibraryPlacement != game.LibraryPlaceTopOrBottom {
		t.Fatalf("the owner was not asked top or bottom: %+v", pc)
	}
	if err := g.ResolvePutInLibrary(pc.ID, opp.ID, []uuid.UUID{attacker}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(attacker) || !opp.Library.Contains(attacker) {
		t.Fatal("the attacker is not in its owner's library")
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("the Adventure card was not exiled")
	}
}
