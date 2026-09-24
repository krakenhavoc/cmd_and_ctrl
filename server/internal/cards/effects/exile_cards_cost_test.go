package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_cards_cost_test.go — #1297: "Exile N cards from your graveyard"
// and "Exile a card from your hand" as a CR 602 activated ability's
// cost (AbilityCost.ExileCards), on the five proof cards.
//
// What is pinned, card by card:
//
//   - Grim Lavamancer: the count, the pile, the distinctness, and that
//     the payment happens at ANNOUNCE — the cards are in exile while
//     the ability is still on the stack — through the exit primitive,
//     never the discard door.
//   - Moorland Haunt, Tome Shredder: the clause's predicate.
//   - Mines of Moria: the caveat is gone and the ability is real.
//   - Holistic Wisdom: the hand form, the paid-cost record the effect
//     reads, and the auto-tapper keeping its hands off the named card.

const (
	grimLavamancerOracle = "37445e06-88a1-4e2e-a432-383736c9b977"
	moorlandHauntOracle  = "5324192b-6687-41e4-8e56-326b21a5dbf3"
	tomeShredderOracle   = "b145952b-52e3-4a66-b47e-f08a489f9443"
	minesOfMoriaOracle   = "583cdebe-0195-45be-bd2e-5765f07cb902"
	holisticWisdomOracle = "7e108285-52da-473c-accd-d48e646a49c0"
	simianSpiritGuideOr  = "44e0ffa3-8915-4c1f-8f1a-4aeea1365f07"
)

// exileCostTable is a catalog game in the active seat's main phase with
// an empty hand and graveyard.
func exileCostTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		me.Graveyard.Cards = nil
	})
	return g, me, opp
}

func floatMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect %s: %v", mana, err)
	}
}

func noDiscardEvents(t *testing.T, g *game.Game) {
	t.Helper()
	for _, ev := range g.Events {
		if ev.Kind == game.EventDiscardCard {
			t.Fatalf("an EventDiscardCard fired (%+v) — the cost exiles, it does not discard", ev)
		}
	}
}

// Grim Lavamancer: two graveyard cards, named at announce, in exile
// before the ability resolves; then the 2 damage.
func TestGrimLavamancerExilesTwoGraveyardCardsToDealTwo(t *testing.T) {
	g, me, opp := exileCostTable(t)
	lava := pushCatalogPermanent(g, me.ID, "Grim Lavamancer", "Creature — Human Wizard", grimLavamancerOracle, false)
	a := pushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	b := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	c := pushGraveyardCardTyped(me, "Old Land", "Land")
	inHand := pushCatalogHandCard(me, "Held Card", "Instant", "")
	floatMana(t, g, me, "{R}")
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	for _, bad := range []struct {
		why string
		ids []uuid.UUID
	}{
		{"no exile_ids", nil},
		{"one card for \"two cards\"", []uuid.UUID{a}},
		{"three cards for \"two cards\"", []uuid.UUID{a, b, c}},
		{"the same card twice", []uuid.UUID{a, a}},
		{"a card from the HAND for a graveyard cost", []uuid.UUID{a, inHand}},
		{"the source itself", []uuid.UUID{a, lava}},
	} {
		if err := g.ActivateCatalogAbility(me.ID, lava, 0, game.ActivateAbilityParams{
			ExileIDs: bad.ids, Targets: target,
		}); err == nil {
			t.Fatalf("%s paid the cost", bad.why)
		}
	}
	// Nothing moved on any refusal (validate-all-then-pay).
	if len(me.Graveyard.Cards) != 3 || !me.Hand.Contains(inHand) {
		t.Fatalf("a refused activation paid part of its cost: graveyard %d", len(me.Graveyard.Cards))
	}
	// Nor is a graveyard card a DISCARD payment for it.
	if err := g.ActivateCatalogAbility(me.ID, lava, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{a, b}, Targets: target,
	}); err == nil {
		t.Fatal("discard_ids paid an exile cost")
	}

	lifeBefore := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, lava, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{a, b}, Targets: target,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// CR 602.2b: paid at announce.
	for _, id := range []uuid.UUID{a, b} {
		if me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
			t.Errorf("%v is not in exile while the ability is on the stack", id)
		}
	}
	if !me.Graveyard.Contains(c) {
		t.Error("the card that was not named left the graveyard")
	}
	noDiscardEvents(t, g)
	var item *game.StackItem
	for _, it := range g.StackMeta {
		if it != nil && it.SourceCardID == lava {
			item = it
		}
	}
	if item == nil {
		t.Fatal("no ability on the stack")
	}
	if got := item.Paid.Exiled; len(got) != 2 || got[0] != a || got[1] != b {
		t.Errorf("Paid.Exiled = %v, want [%v %v]", got, a, b)
	}

	passPriorityAroundTable(t, g)
	if opp.Life != lifeBefore-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, lifeBefore-2)
	}
}

// Moorland Haunt: "a creature card" is paid with a creature card and
// nothing else, and the Spirit it makes flies.
func TestMoorlandHauntExilesACreatureCardForASpirit(t *testing.T) {
	g, me, _ := exileCostTable(t)
	haunt := pushCatalogPermanent(g, me.ID, "Moorland Haunt", "Land", moorlandHauntOracle, false)
	spell := pushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	bear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	floatMana(t, g, me, "{W}{U}")

	if err := g.ActivateCatalogAbility(me.ID, haunt, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery paid \"exile a creature card\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, haunt, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(bear) || !me.Graveyard.Contains(spell) {
		t.Fatal("the payment did not exile exactly the creature card")
	}
	passPriorityAroundTable(t, g)
	spirits := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && IsToken(c) && c.HasSubtype("Spirit") {
			spirits++
			if !game.HasKeyword(&c, "flying") {
				t.Error("the Spirit token has no flying")
			}
		}
	}
	if spirits != 1 {
		t.Errorf("%d Spirit tokens, want 1", spirits)
	}
}

// Tome Shredder: "an instant or sorcery card", and the counter lands on
// the Shredder.
func TestTomeShredderExilesAnInstantOrSorceryForACounter(t *testing.T) {
	g, me, _ := exileCostTable(t)
	shredder := pushCatalogPermanent(g, me.ID, "Tome Shredder", "Creature — Wolf", tomeShredderOracle, false)
	bear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	bolt := pushGraveyardCardTyped(me, "Spent Bolt", "Instant")

	if err := g.ActivateCatalogAbility(me.ID, shredder, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature card paid \"an instant or sorcery card\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, shredder, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bolt}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	var counters int
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == shredder {
				counters = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if counters != 1 {
		t.Errorf("Tome Shredder has %d +1/+1 counters, want 1", counters)
	}
}

// Mines of Moria: the ability the caveat used to name is real — three
// graveyard cards, two Treasures — and the card is complete.
func TestMinesOfMoriaExilesThreeCardsForTwoTreasures(t *testing.T) {
	spec, ok := Lookup(minesOfMoriaOracle)
	if !ok {
		t.Fatal("Mines of Moria is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Mines of Moria: completeness %v, caveats %v — the Treasure ability is implemented", spec.Completeness, spec.Caveats)
	}

	g, me, _ := exileCostTable(t)
	mines := pushCatalogPermanent(g, me.ID, "Mines of Moria", "Legendary Land", minesOfMoriaOracle, false)
	fuel := []uuid.UUID{
		pushGraveyardCardTyped(me, "One", "Sorcery"),
		pushGraveyardCardTyped(me, "Two", "Instant"),
	}
	floatMana(t, g, me, "{R}{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, mines, 0, game.ActivateAbilityParams{ExileIDs: fuel}); err == nil {
		t.Fatal("two graveyard cards paid \"exile three cards\"")
	}
	fuel = append(fuel, pushGraveyardCardTyped(me, "Three", "Land"))
	if err := g.ActivateCatalogAbility(me.ID, mines, 0, game.ActivateAbilityParams{ExileIDs: fuel}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	treasures := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && IsToken(c) && c.Name == "Treasure" {
			treasures++
		}
	}
	if treasures != 2 {
		t.Errorf("%d Treasures, want 2", treasures)
	}
}

// Holistic Wisdom: the hand form. The exiled card's card type decides
// whether the target comes back — the effect reads the paid-cost
// record, since by resolution the card is in exile and nothing on the
// board says which one paid.
func TestHolisticWisdomReturnsTheTargetOnlyIfItSharesAType(t *testing.T) {
	g, me, _ := exileCostTable(t)
	wisdom := pushCatalogPermanent(g, me.ID, "Holistic Wisdom", "Enchantment", holisticWisdomOracle, false)
	deadBear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	handBeast := pushCatalogHandCard(me, "Spare Beast", "Creature — Beast", "")
	handLand := pushCatalogHandCard(me, "Spare Land", "Land", "")

	// A graveyard card is not a HAND payment.
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{deadBear}, Targets: cardRefs(deadBear),
	}); err == nil {
		t.Fatal("a graveyard card paid \"exile a card from your hand\"")
	}

	// A land shares no card type with a creature card: the land is
	// exiled (the cost was paid) and the Bear stays put.
	if err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{handLand}, Targets: cardRefs(deadBear),
	}); err != nil {
		t.Fatalf("activate (land): %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(handLand) {
		t.Fatal("the land was not exiled")
	}
	if !me.Graveyard.Contains(deadBear) {
		t.Fatal("a creature card came back for an exiled land")
	}
	noDiscardEvents(t, g)

	// A creature card does: the Bear returns.
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{handBeast}, Targets: cardRefs(deadBear),
	}); err != nil {
		t.Fatalf("activate (creature): %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(deadBear) {
		t.Error("the Bear did not return for an exiled creature card")
	}
	if !g.Exile.Contains(handBeast) {
		t.Error("the Beast was not exiled")
	}
}

// CR 118.3: a Simian Spirit Guide named to Holistic Wisdom's exile cost
// is spent on that cost, so the auto-tapper may not also exile it for
// the {2}. With one Mountain and nothing else the activation is simply
// unpayable — refused, with the Guide still in hand and the Mountain
// untapped — rather than a Guide spent twice.
func TestHolisticWisdomDoesNotSpendItsExiledSpiritGuideOnMana(t *testing.T) {
	g, me, _ := exileCostTable(t)
	wisdom := pushCatalogPermanent(g, me.ID, "Holistic Wisdom", "Enchantment", holisticWisdomOracle, false)
	mountain := pushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	guide := pushCatalogHandCard(me, "Simian Spirit Guide", "Creature — Ape Spirit", simianSpiritGuideOr)
	deadBear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")

	err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{guide}, Targets: cardRefs(deadBear),
		Strict: true, AutoTap: true,
	})
	if err == nil {
		t.Fatal("one Mountain paid {2} — the named Spirit Guide was spent on the mana as well as the cost")
	}
	if !me.Hand.Contains(guide) {
		t.Error("the refused activation moved the Spirit Guide")
	}
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == mountain && c.Tapped {
				t.Error("the refused activation tapped the Mountain")
			}
		}
	})

	// With a second Mountain the same announcement is payable, and the
	// Guide pays the exile cost — not the mana.
	pushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	if err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{guide}, Targets: cardRefs(deadBear),
		Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("activate with two Mountains: %v", err)
	}
	if !g.Exile.Contains(guide) {
		t.Error("the Spirit Guide did not pay the exile cost")
	}
}

// Plus has to carry the component through. Every graveyard-exile card
// composes it with mana and usually {T}, and a Plus that dropped it
// would ship Grim Lavamancer pinging for {R} forever — stronger than
// printed, the #259 direction, and silent.
func TestPlusKeepsTheExileCardsComponent(t *testing.T) {
	cost := Plus(ManaCost("{R}"), TapCost(), ExileFromGraveyard(2, "two cards", nil))
	if !cost.Tap || cost.Mana != "{R}" {
		t.Fatalf("the other components were lost: %+v", cost)
	}
	ec := cost.ExileCards
	if ec == nil {
		t.Fatal("Plus dropped the exile-cards component")
	}
	if ec.N != 2 || ec.Label != "two cards" || ec.Zone() != game.ZoneGraveyard {
		t.Errorf("ExileCards = %+v, want two cards from the graveyard", *ec)
	}
	if hand := ExileFromHand(1, "a card", nil).ExileCards; hand.Zone() != game.ZoneHand {
		t.Errorf("ExileFromHand reads the %s", hand.Zone())
	}
}

// Register refuses an exile clause that exiles nothing (a free ability)
// or reads a pile the component does not walk (an ability that can
// never be paid), on BOTH owners — one check, shared.
func TestRegisterRejectsABadExileCardsClause(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	mustPanic(t, "exiles 0 cards", func() {
		Register(Spec{
			OracleID: "exile-cost-test-zero",
			Name:     "Zero Exile",
			Activated: []ActivatedAbility{{
				Label: "Exile no cards: do nothing", Cost: ExileFromGraveyard(0, "no cards", nil), Effect: noop,
			}},
		})
	})
	mustPanic(t, "from the library", func() {
		Register(Spec{
			OracleID: "exile-cost-test-library",
			Name:     "Library Exile",
			Activated: []ActivatedAbility{{
				Label:  "Exile a card from your library: do nothing",
				Cost:   game.AbilityCost{ExileCards: &game.ExileCost{N: 1, Label: "a card", From: game.ZoneLibrary}},
				Effect: noop,
			}},
		})
	})
	mustPanic(t, "exiles 0 cards", func() {
		Register(Spec{
			OracleID: "exile-cost-test-mana-zero",
			Name:     "Zero Exile Mana",
			ManaAbilities: []ManaAbility{{
				Cost: ManaAbilityCost{Tap: true, ExileCards: ExileCardsFromHand(0, "no cards", nil)}, Produced: "{C}", Label: "Add {C}",
			}},
		})
	})
}
