package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grim Lavamancer: two graveyard cards, named at announce, in exile
// before the ability resolves; then the 2 damage.
func TestGrimLavamancerExilesTwoGraveyardCardsToDealTwo(t *testing.T) {
	g, me, opp := effectstest.ExileCostTable(t)
	lava := effectstest.PushCatalogPermanent(g, me.ID, "Grim Lavamancer", "Creature — Human Wizard", effectstest.GrimLavamancerOracle, false)
	a := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	b := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	c := effectstest.PushGraveyardCardTyped(me, "Old Land", "Land")
	inHand := effectstest.PushCatalogHandCard(me, "Held Card", "Instant", "")
	effectstest.FloatMana(t, g, me, "{R}")
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
	effectstest.NoDiscardEvents(t, g)
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

	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lifeBefore-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, lifeBefore-2)
	}
}

// Moorland Haunt: "a creature card" is paid with a creature card and
// nothing else, and the Spirit it makes flies.
func TestMoorlandHauntExilesACreatureCardForASpirit(t *testing.T) {
	g, me, _ := effectstest.ExileCostTable(t)
	haunt := effectstest.PushCatalogPermanent(g, me.ID, "Moorland Haunt", "Land", effectstest.MoorlandHauntOracle, false)
	spell := effectstest.PushGraveyardCardTyped(me, "Spent Spell", "Sorcery")
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.FloatMana(t, g, me, "{W}{U}")

	if err := g.ActivateCatalogAbility(me.ID, haunt, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery paid \"exile a creature card\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, haunt, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(bear) || !me.Graveyard.Contains(spell) {
		t.Fatal("the payment did not exile exactly the creature card")
	}
	effectstest.PassPriorityAroundTable(t, g)
	spirits := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && effects.IsToken(c) && c.HasSubtype("Spirit") {
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
	g, me, _ := effectstest.ExileCostTable(t)
	shredder := effectstest.PushCatalogPermanent(g, me.ID, "Tome Shredder", "Creature — Wolf", effectstest.TomeShredderOracle, false)
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	bolt := effectstest.PushGraveyardCardTyped(me, "Spent Bolt", "Instant")

	if err := g.ActivateCatalogAbility(me.ID, shredder, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature card paid \"an instant or sorcery card\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, shredder, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bolt}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	spec, ok := effects.Lookup(effectstest.MinesOfMoriaOracle)
	if !ok {
		t.Fatal("Mines of Moria is not in the catalog")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Mines of Moria: completeness %v, caveats %v — the Treasure ability is implemented", spec.Completeness, spec.Caveats)
	}

	g, me, _ := effectstest.ExileCostTable(t)
	mines := effectstest.PushCatalogPermanent(g, me.ID, "Mines of Moria", "Legendary Land", effectstest.MinesOfMoriaOracle, false)
	fuel := []uuid.UUID{
		effectstest.PushGraveyardCardTyped(me, "One", "Sorcery"),
		effectstest.PushGraveyardCardTyped(me, "Two", "Instant"),
	}
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, mines, 0, game.ActivateAbilityParams{ExileIDs: fuel}); err == nil {
		t.Fatal("two graveyard cards paid \"exile three cards\"")
	}
	fuel = append(fuel, effectstest.PushGraveyardCardTyped(me, "Three", "Land"))
	if err := g.ActivateCatalogAbility(me.ID, mines, 0, game.ActivateAbilityParams{ExileIDs: fuel}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	treasures := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && effects.IsToken(c) && c.Name == "Treasure" {
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
	g, me, _ := effectstest.ExileCostTable(t)
	wisdom := effectstest.PushCatalogPermanent(g, me.ID, "Holistic Wisdom", "Enchantment", effectstest.HolisticWisdomOracle, false)
	deadBear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	handBeast := effectstest.PushCatalogHandCard(me, "Spare Beast", "Creature — Beast", "")
	handLand := effectstest.PushCatalogHandCard(me, "Spare Land", "Land", "")

	// A graveyard card is not a HAND payment.
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{deadBear}, Targets: effectstest.CardRefs(deadBear),
	}); err == nil {
		t.Fatal("a graveyard card paid \"exile a card from your hand\"")
	}

	// A land shares no card type with a creature card: the land is
	// exiled (the cost was paid) and the Bear stays put.
	if err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{handLand}, Targets: effectstest.CardRefs(deadBear),
	}); err != nil {
		t.Fatalf("activate (land): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(handLand) {
		t.Fatal("the land was not exiled")
	}
	if !me.Graveyard.Contains(deadBear) {
		t.Fatal("a creature card came back for an exiled land")
	}
	effectstest.NoDiscardEvents(t, g)

	// A creature card does: the Bear returns.
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{handBeast}, Targets: effectstest.CardRefs(deadBear),
	}); err != nil {
		t.Fatalf("activate (creature): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g, me, _ := effectstest.ExileCostTable(t)
	wisdom := effectstest.PushCatalogPermanent(g, me.ID, "Holistic Wisdom", "Enchantment", effectstest.HolisticWisdomOracle, false)
	mountain := effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	guide := effectstest.PushCatalogHandCard(me, "Simian Spirit Guide", "Creature — Ape Spirit", effectstest.SimianSpiritGuideOr)
	deadBear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")

	err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{guide}, Targets: effectstest.CardRefs(deadBear),
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
	effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	if err := g.ActivateCatalogAbility(me.ID, wisdom, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{guide}, Targets: effectstest.CardRefs(deadBear),
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
	cost := effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.ExileFromGraveyard(2, "two cards", nil))
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
	if hand := effects.ExileFromHand(1, "a card", nil).ExileCards; hand.Zone() != game.ZoneHand {
		t.Errorf("ExileFromHand reads the %s", hand.Zone())
	}
}

// Register refuses an exile clause that exiles nothing (a free ability)
// or reads a pile the component does not walk (an ability that can
// never be paid), on BOTH owners — one check, shared.
func TestRegisterRejectsABadExileCardsClause(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	effectstest.MustPanic(t, "exiles 0 cards", func() {
		effects.Register(effects.Spec{
			OracleID: "exile-cost-test-zero",
			Name:     "Zero Exile",
			Activated: []effects.ActivatedAbility{{
				Label: "Exile no cards: do nothing", Cost: effects.ExileFromGraveyard(0, "no cards", nil), Effect: noop,
			}},
		})
	})
	effectstest.MustPanic(t, "from the library", func() {
		effects.Register(effects.Spec{
			OracleID: "exile-cost-test-library",
			Name:     "Library Exile",
			Activated: []effects.ActivatedAbility{{
				Label:  "Exile a card from your library: do nothing",
				Cost:   game.AbilityCost{ExileCards: &game.ExileCost{N: 1, Label: "a card", From: game.ZoneLibrary}},
				Effect: noop,
			}},
		})
	})
	effectstest.MustPanic(t, "exiles 0 cards", func() {
		effects.Register(effects.Spec{
			OracleID: "exile-cost-test-mana-zero",
			Name:     "Zero Exile Mana",
			ManaAbilities: []effects.ManaAbility{{
				Cost: effects.ManaAbilityCost{Tap: true, ExileCards: effects.ExileCardsFromHand(0, "no cards", nil)}, Produced: "{C}", Label: "Add {C}",
			}},
		})
	})
}
