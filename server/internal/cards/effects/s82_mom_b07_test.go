package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b07_test.go — S82, March of the Machine batch 7: the convoke
// cards that read the creatures that convoked them.

const (
	s82AncientImperiosaurOracle = "3fe95820-cd7f-4f28-82bf-eb4350ea4179"
	s82KnightErrantOracle       = "4bd20d79-0861-494a-a8f2-f6ddb5088004"
	s82ZephyrSingerOracle       = "1800ae74-0b54-4547-ab8f-ea61418d73e8"
	s82WildfireAwakenerOracle   = "d22196b5-b111-45ce-ae51-a06ca02533b6"
	s82CompleteTheCircuitOracle = "402506fa-6bd5-4814-8353-90c148dd576c"
	s82CutShortOracle           = "428d131c-b0fa-4e96-9c02-e2e6714357b9"
	s82WickedSlumberOracle      = "5b27b794-580c-472d-b63e-99b271eb115d"
	s82ArtisticRefusalOracle    = "ff192482-3808-4291-a31e-3dfaeb49cda2"
)

func TestS82MomB07EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		s82AncientImperiosaurOracle: "Ancient Imperiosaur",
		s82KnightErrantOracle:       "Knight-Errant of Eos",
		s82ZephyrSingerOracle:       "Zephyr Singer",
		s82WildfireAwakenerOracle:   "Wildfire Awakener",
		s82CompleteTheCircuitOracle: "Complete the Circuit",
		s82CutShortOracle:           "Cut Short",
		s82WickedSlumberOracle:      "Wicked Slumber",
		s82ArtisticRefusalOracle:    "Artistic Refusal",
	}
	for oracle, name := range want {
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

// s82Cast casts a card from the active seat's hand under the strict mana
// gate, paying with the floated mana and the creatures in taps.
func s82Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, colors []string,
	mana string, taps []uuid.UUID, targets []game.TargetRef) (*game.Player, uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	id := handCardFull(me, name, typeLine, cost, oracle, colors)
	floatForTest(g, me, mana)
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: taps, Targets: targets})
	return me, id, err
}

func s82Creatures(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, pushVanillaCreature(g, owner, "Convoker", 1, 1))
	}
	return ids
}

// --- Ancient Imperiosaur ---

func TestAncientImperiosaurEntersWithTwoCountersPerConvoker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	taps := s82Creatures(g, me.ID, 3)
	bystander := pushVanillaCreature(g, me.ID, "Bystander", 1, 1)
	_, id, err := s82Cast(t, g, "Ancient Imperiosaur", "Creature — Dinosaur", "{5}{G}{G}",
		s82AncientImperiosaurOracle, []string{"G"}, "GGGG", taps, nil)
	if err != nil {
		t.Fatalf("cast with three convokers: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 6 {
		t.Errorf("counters with three convokers: %d, want 6", got)
	}
	if counterCount(g, bystander, game.CounterPlusOne) != 0 {
		t.Error("a creature that did not convoke gained counters")
	}
}

func TestAncientImperiosaurWithoutConvokeEntersPlain(t *testing.T) {
	g := newCatalogGame(t)
	_, id, err := s82Cast(t, g, "Ancient Imperiosaur", "Creature — Dinosaur", "{5}{G}{G}",
		s82AncientImperiosaurOracle, []string{"G"}, "GGGGGGG", nil, nil)
	if err != nil {
		t.Fatalf("cast for mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 0 {
		t.Errorf("counters with no convoker: %d, want 0", got)
	}
}

// --- Zephyr Singer ---

func TestZephyrSingerGivesFlyingCountersToItsConvokers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	taps := s82Creatures(g, me.ID, 2)
	bystander := pushVanillaCreature(g, me.ID, "Bystander", 1, 1)
	_, singer, err := s82Cast(t, g, "Zephyr Singer", "Creature — Siren Pirate", "{2}{U}{U}",
		s82ZephyrSingerOracle, []string{"U"}, "UU", taps, nil)
	if err != nil {
		t.Fatalf("cast with two convokers: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range taps {
		if got := counterCount(g, id, game.CounterFlying); got != 1 {
			t.Errorf("convoker flying counters: %d, want 1", got)
		}
	}
	if counterCount(g, bystander, game.CounterFlying) != 0 || counterCount(g, singer, game.CounterFlying) != 0 {
		t.Error("a creature that did not convoke gained a flying counter")
	}
}

// --- Knight-Errant of Eos ---

func TestKnightErrantOfEosFindsCreaturesUpToTheConvokerCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	taps := s82Creatures(g, me.ID, 2)
	mk := func(name, tl, cost string) uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		})
		return id
	}
	big := mk("Big", "Creature — Giant", "{4}{G}")
	spell := mk("Spell", "Instant", "{W}")
	small := mk("Small", "Creature — Elf", "{1}{G}")
	_, _, err := s82Cast(t, g, "Knight-Errant of Eos", "Creature — Human Knight", "{4}{W}",
		s82KnightErrantOracle, []string{"W"}, "WWW", taps, nil)
	if err != nil {
		t.Fatalf("cast with two convokers: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no pick offered")
	}
	offered := map[uuid.UUID]bool{}
	for _, c := range pick.ChooseCards {
		offered[c] = true
	}
	if !offered[small] || offered[big] || offered[spell] {
		t.Fatalf("offered small=%v big=%v spell=%v; want only the creature with mana value 2 or less", offered[small], offered[big], offered[spell])
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{small}); err != nil {
		t.Fatalf("take: %v", err)
	}
	if !me.Hand.Contains(small) {
		t.Error("the chosen creature is not in hand")
	}
	if !me.Library.Contains(big) {
		t.Error("the unchosen cards left the library")
	}
}

// --- Wildfire Awakener ---

func TestWildfireAwakenerMakesXPingingElementals(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	taps := s82Creatures(g, me.ID, 1)
	toMainForCost(t, g)
	id := handCardFull(me, "Wildfire Awakener", "Creature — Human Wizard", "{X}{1}{R}{W}", s82WildfireAwakenerOracle, []string{"R", "W"})
	floatForTest(g, me, "RWRR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: taps, XValue: 2}); err != nil {
		t.Fatalf("cast with X=2 and a convoker: %v", err)
	}
	passPriorityAroundTable(t, g)
	var tokens []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			tokens = append(tokens, c.InstanceID)
			if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "R" {
				t.Errorf("token is %d/%d %v, want 1/1 red", c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("Elementals: %d, want 2", len(tokens))
	}
	life := opp.Life
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tokens[0]) })
	g.RunStateChecksForTest()
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := opp.Life; got != life-1 {
		t.Errorf("target player life %d, want %d (1 damage when the Elemental became tapped)", got, life-1)
	}
}

// --- Cut Short ---

func TestCutShortTargetsAnActivatedPlaneswalkerOrATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	untapped := pushVanillaCreature(g, opp.ID, "Untapped Bear", 2, 2)
	tapped := pushVanillaCreature(g, opp.ID, "Tapped Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	idle := pushPermanentForTest(g, opp.ID, "Idle Walker", "", "Legendary Planeswalker — Test")
	busy := pushPermanentForTest(g, opp.ID, "Busy Walker", "", "Legendary Planeswalker — Test")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].IsPlaneswalker() {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 3}
			}
		}
		if g.LoyaltyActivatedThisTurn == nil {
			g.LoyaltyActivatedThisTurn = map[uuid.UUID]bool{}
		}
		g.LoyaltyActivatedThisTurn[busy] = true
	})
	for _, bad := range []uuid.UUID{untapped, idle} {
		if _, _, err := s82Cast(t, g, "Cut Short", "Instant", "{2}{W}", s82CutShortOracle, []string{"W"}, "WWW", nil, cardTarget(bad)); err == nil {
			t.Fatalf("Cut Short accepted an untapped creature or an idle planeswalker")
		}
	}
	_, _, err := s82Cast(t, g, "Cut Short", "Instant", "{2}{W}", s82CutShortOracle, []string{"W"}, "WWW", nil, cardTarget(busy))
	if err != nil {
		t.Fatalf("an activated planeswalker is a legal target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, busy) {
		t.Error("the activated planeswalker was not destroyed")
	}
	if _, _, err := s82Cast(t, g, "Cut Short", "Instant", "{2}{W}", s82CutShortOracle, []string{"W"}, "WWW", nil, cardTarget(tapped)); err != nil {
		t.Fatalf("a tapped creature is a legal target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, tapped) || !onBattlefieldByID(g, untapped) {
		t.Error("Cut Short should destroy only the tapped creature")
	}
}

func TestCutShortCanBeConvoked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := pushVanillaCreature(g, opp.ID, "Tapped Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(victim) })
	taps := s82Creatures(g, me.ID, 2)
	if _, _, err := s82Cast(t, g, "Cut Short", "Instant", "{2}{W}", s82CutShortOracle, []string{"W"}, "W", taps, cardTarget(victim)); err != nil {
		t.Fatalf("cast with two convokers: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, victim) {
		t.Error("the tapped creature survived")
	}
}

// --- Wicked Slumber ---

func TestWickedSlumberTapsTwoAndPlacesTwoStunCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
	_, _, err := s82Cast(t, g, "Wicked Slumber", "Instant", "{3}{U}", s82WickedSlumberOracle, []string{"U"}, "UUUU", nil,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isTapped(g, a) || !isTapped(g, b) {
		t.Fatal("both targets should be tapped")
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no first choice")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("first pick: %v", err)
	}
	pick = latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no second choice")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("second pick: %v", err)
	}
	if got := counterCount(g, a, game.CounterStun); got != 2 {
		t.Errorf("stun counters on the chosen creature: %d, want 2 (both counters may go on one)", got)
	}
	if got := counterCount(g, b, game.CounterStun); got != 0 {
		t.Errorf("stun counters on the other: %d, want 0", got)
	}
}

func TestWickedSlumberWithOneTargetPutsBothCountersOnIt(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	if _, _, err := s82Cast(t, g, "Wicked Slumber", "Instant", "{3}{U}", s82WickedSlumberOracle, []string{"U"}, "UUUU", nil, cardTarget(a)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, a, game.CounterStun); got != 2 || !isTapped(g, a) {
		t.Errorf("stun counters %d tapped %v; want 2 and tapped", got, isTapped(g, a))
	}
}

// --- Artistic Refusal ---

func TestArtisticRefusalCountersAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	toMainForCost(t, g)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() { opp.ManaPool.AddMana(game.ManaToken{Color: "R", Source: uuid.New()}) })
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("Bolt: %v", err)
	}
	taps := s82Creatures(g, me.ID, 3)
	id := handCardFull(me, "Artistic Refusal", "Instant", "{4}{U}{U}", s82ArtisticRefusalOracle, []string{"U"})
	floatForTest(g, me, "UUU")
	handBefore := me.Hand.Size()
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{0, 1}, TapIDs: taps,
		Targets: []game.TargetRef{p1g3ModeCard(bolt, 0)}})
	if err != nil {
		t.Fatalf("cast with both modes: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the Bolt was not countered")
	}
	if me.Hand.Size() != handBefore-1+2 {
		t.Errorf("hand %d after the draw two, discard pending: want %d", me.Hand.Size(), handBefore+1)
	}
}

// --- Complete the Circuit ---

func TestCompleteTheCircuitCopiesTheNextSpellTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%4].ID
	taps := s82Creatures(g, me.ID, 3)
	if _, _, err := s82Cast(t, g, "Complete the Circuit", "Instant", "{5}{U}", s82CompleteTheCircuitOracle, []string{"U"}, "UUU", taps, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("Complete the Circuit left %d delayed triggers, want 1", len(g.DelayedTriggers))
	}
	before := lifeOf(g, victim)
	castBoltAtAndKeepTheCopysTarget(t, g, me.ID, victim, 2)
	if got := lifeOf(g, victim); got != before-9 {
		t.Errorf("victim life = %d, want %d (the Bolt and two copies)", got, before-9)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("the trigger is for the next spell only; %d left", len(g.DelayedTriggers))
	}
}

func TestCompleteTheCircuitLetsSorceriesBeCastWithFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	taps := s82Creatures(g, me.ID, 3)
	if _, _, err := s82Cast(t, g, "Complete the Circuit", "Instant", "{5}{U}", s82CompleteTheCircuitOracle, []string{"U"}, "UUU", taps, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	sorc := handCardFull(me, "Test Sorcery", "Sorcery", "{U}", "", []string{"U"})
	floatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, sorc, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a sorcery should have flash this turn: %v", err)
	}
}

func TestSorceriesCannotBeCastInTheEndStepWithoutIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	sorc := handCardFull(me, "Test Sorcery", "Sorcery", "{U}", "", []string{"U"})
	floatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, sorc, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a sorcery was cast in the end step with no flash")
	}
}
