package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_from_hand_test.go — ADR 0189 decision 3 (#3538): manifesting
// cards from a hand. Kozilek, the Broken Reality manifests two cards
// from each of up to two target players' hands; Scroll of Fate
// manifests one of yours.

const (
	kozilekBrokenRealityOracle = "39f4bf0e-47d5-4ae8-a504-71acda849850"
	scrollOfFateOracle         = "c15c13d2-8db9-4801-a1f0-f087f2551c92"
)

// emptyHandForTest removes every card from a player's hand.
func emptyHandForTest(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { p.Hand.Cards = nil })
}

// fillHandForTest gives a player a hand of n distinct cards and returns
// their IDs. Every third card is an instant, so a manifest of a
// nonpermanent is in the mix.
func fillHandForTest(g *game.Game, p *game.Player, n int) []uuid.UUID {
	emptyHandForTest(g, p)
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		typeLine := "Creature — Bear"
		if i%3 == 2 {
			typeLine = "Instant"
		}
		ids = append(ids, handCardForTest(p, "Hand card "+string(rune('A'+i)), typeLine, ""))
	}
	return ids
}

// castKozilek puts Kozilek in the active player's hand, casts it and
// answers the cast trigger's target prompt with `targets`. It leaves the
// trigger on the stack above the spell.
func castKozilek(t *testing.T, g *game.Game, targets ...uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := putInHand(me, game.Card{Name: "Kozilek, the Broken Reality", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: kozilekBrokenRealityOracle, ManaCost: "{9}", Power: 9, Toughness: 9})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the cast trigger did not ask for its target players")
	}
	refs := make([]game.TargetRef, 0, len(targets))
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	return id
}

// cardInPlay reports whether a card is on the battlefield.
func cardInPlay(g *game.Game, id uuid.UUID) bool {
	_, ok := battlefieldCard(g, id)
	return ok
}

// assertManifestedBy checks a card is a face-down manifest under
// `owner`'s control that only `owner` can look at.
func assertManifestedBy(t *testing.T, g *game.Game, id, owner uuid.UUID) {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	if !c.FaceDownIsPermanent() || c.FaceDownKind != game.FaceDownManifested {
		t.Fatalf("face-down state = (%v, %q), want a manifested permanent", c.FaceDown, c.FaceDownKind)
	}
	if c.Controller != owner {
		t.Errorf("controller = %s, want its owner %s", c.Controller, owner)
	}
	for _, p := range g.Seats {
		if known := c.IsKnownTo(p.ID); known != (p.ID == owner) {
			t.Errorf("seat %s knows the manifest = %v, want only its owner to (CR 708.5)", p.ID, known)
		}
	}
	if eff := c.Effective(); eff.Power != 2 || eff.Toughness != 2 {
		// The anthem tests assert the buffed body; here Kozilek is not
		// on the battlefield, so a manifest is the plain 2/2.
		t.Errorf("body = %d/%d, want a 2/2", eff.Power, eff.Toughness)
	}
}

// The headline: two targets, five cards and one card. The first is
// asked for exactly two; the second is not asked and manifests their
// lone card; both are face down under their owners; the caster draws
// three.
func TestKozilekAsksTheFirstForTwoAndManifestsTheSecondsLoneCard(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	firstHand := fillHandForTest(g, first, 5)
	secondHand := fillHandForTest(g, second, 1)

	castKozilek(t, g, first.ID, second.ID)
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, first.ID)
	if pick == nil {
		t.Fatal("the first target was not asked to choose cards to manifest")
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want exactly two", pick.ChooseMin, pick.ChooseMax)
	}
	if pick.ChooseDestination != game.ChooseManifest {
		t.Errorf("destination %q, want %q", pick.ChooseDestination, game.ChooseManifest)
	}
	if len(pick.ChooseCards) != 5 {
		t.Errorf("%d candidates, want the whole five-card hand", len(pick.ChooseCards))
	}
	if chooseCardsChoiceFor(g, second.ID) != nil {
		t.Error("the one-card hand was asked a question; the ruling is that they manifest that card")
	}
	if cardInPlay(g, secondHand[0]) {
		t.Error("the lone card entered before the first target answered; the run must wait for every seat")
	}

	answerChooseCards(t, g, first.ID, firstHand[1], firstHand[2])
	passPriorityAroundTable(t, g)

	assertManifestedBy(t, g, firstHand[1], first.ID)
	assertManifestedBy(t, g, firstHand[2], first.ID)
	assertManifestedBy(t, g, secondHand[0], second.ID)
	for _, id := range []uuid.UUID{firstHand[0], firstHand[3], firstHand[4]} {
		if cardInPlay(g, id) {
			t.Errorf("card %s was not picked and entered anyway", id)
		}
	}
	if got := first.Hand.Size(); got != 3 {
		t.Errorf("first target holds %d cards, want the 3 unpicked", got)
	}
	if got := second.Hand.Size(); got != 0 {
		t.Errorf("second target holds %d cards, want none", got)
	}
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("caster's hand %d -> %d, want +3 (one draw per manifested card)", handBefore, got)
	}
}

// The trigger is a cast trigger: it resolves above the spell, so a
// countered Kozilek still manifests and draws (the 2024-06-07 ruling).
func TestKozilekStillManifestsWhenCountered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := fillHandForTest(g, opp, 1)

	id := castKozilek(t, g, opp.ID)
	handBefore := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	assertManifestedBy(t, g, hand[0], opp.ID)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("caster's hand %d -> %d, want one card drawn", handBefore, got)
	}
}

// "Up to two": with no target nothing is manifested and nothing drawn.
func TestKozilekWithNoTargetsManifestsAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := fillHandForTest(g, opp, 4)

	castKozilek(t, g)
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)

	for _, id := range hand {
		if cardInPlay(g, id) {
			t.Errorf("card %s was manifested with no target", id)
		}
	}
	if chooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("an untargeted player was asked to manifest")
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("caster's hand %d -> %d, want no draw", handBefore, got)
	}
}

// A target with an empty hand manifests nothing, so the caster draws
// for the other target only.
func TestKozilekSkipsAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	empty, one := g.Seats[2], g.Seats[3]
	emptyHandForTest(g, empty)
	hand := fillHandForTest(g, one, 2)

	castKozilek(t, g, empty.ID, one.ID)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()
	passPriorityAroundTable(t, g)

	if chooseCardsChoiceFor(g, one.ID) != nil {
		t.Error("a two-card hand was asked about two cards")
	}
	assertManifestedBy(t, g, hand[0], one.ID)
	assertManifestedBy(t, g, hand[1], one.ID)
	if got := caster.Hand.Size(); got != handBefore+2 {
		t.Errorf("caster's hand %d -> %d, want +2", handBefore, got)
	}
}

// A replacement that keeps one card off the battlefield lowers the draw
// by one: "for each card manifested this way" counts what arrived.
func TestKozilekDrawsOnlyForTheCardsThatArrive(t *testing.T) {
	const keeper = "test-3538-hand-keeper"
	var kept uuid.UUID
	registerForTest(t, Spec{
		OracleID: keeper,
		Name:     "Test Hand Keeper",
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					ev.OldZone == game.ZoneHand && ev.CardID == kept
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Label: "Test Hand Keeper: one card stays where it is",
		}},
	})
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Test Hand Keeper", "Enchantment", keeper, false)
	hand := fillHandForTest(g, opp, 2)
	kept = hand[0]

	castKozilek(t, g, opp.ID)
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)

	if cardInPlay(g, kept) {
		t.Fatal("the replacement did not keep the card off the battlefield")
	}
	assertManifestedBy(t, g, hand[1], opp.ID)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("caster's hand %d -> %d, want +1: only one card arrived", handBefore, got)
	}
}

// "Other colorless creatures you control get +3/+2": a face-down 2/2
// you control is 5/4, Kozilek itself stays 9/9, and neither a colored
// creature nor an opponent's colorless one is touched.
func TestKozilekPumpsOtherColorlessCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := fillHandForTest(g, me, 1)

	kozilek := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kozilek, the Broken Reality", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: kozilekBrokenRealityOracle, Power: 9, Toughness: 9, Owner: me.ID, Controller: me.ID,
	})
	scion := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Eldrazi Scion", TypeLine: "Creature — Eldrazi Scion",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	green := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Scion", TypeLine: "Creature — Eldrazi Scion",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})

	g.WithWriteLock(func() {
		if err := g.ManifestFromHandsKeyedForEffect(hand, game.KeyedTail{}); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})

	check := func(label string, id uuid.UUID, power, toughness int) {
		t.Helper()
		if p, tg := effectivePower(t, g, id), effectiveToughness(t, g, id); p != power || tg != toughness {
			t.Errorf("%s is %d/%d, want %d/%d", label, p, tg, power, toughness)
		}
	}
	check("Kozilek", kozilek, 9, 9)
	check("a colorless Scion", scion, 4, 3)
	check("a green creature", green, 2, 2)
	check("an opponent's colorless creature", theirs, 1, 1)
	check("the face-down manifest", hand[0], 5, 4)
}

// Scroll of Fate: {T}, manifest a card from your hand, chosen as the
// ability resolves.
func TestScrollOfFateTapsToManifestAChosenCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := fillHandForTest(g, me, 3)
	scroll := pushCatalogPermanent(g, me.ID, "Scroll of Fate", "Artifact", scrollOfFateOracle, false)
	toMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if tapped, _ := battlefieldCardTapped(g, scroll); !tapped {
		t.Error("the scroll was not tapped to pay for its ability")
	}
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Scroll of Fate did not ask which card to manifest")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 || pick.ChooseDestination != game.ChooseManifest {
		t.Errorf("prompt = %d..%d to %q, want exactly one, to manifest", pick.ChooseMin, pick.ChooseMax, pick.ChooseDestination)
	}
	answerChooseCards(t, g, me.ID, hand[1])
	passPriorityAroundTable(t, g)

	assertManifestedBy(t, g, hand[1], me.ID)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand = %d, want the 2 unpicked", got)
	}
}

// A one-card hand is not asked, and an empty hand manifests nothing.
func TestScrollOfFateWithOneCardOrNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := fillHandForTest(g, me, 1)
	scroll := pushCatalogPermanent(g, me.ID, "Scroll of Fate", "Artifact", scrollOfFateOracle, false)
	toMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a one-card hand was asked which card to manifest")
	}
	assertManifestedBy(t, g, hand[0], me.ID)

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == scroll {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("second ActivateAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := me.Hand.Size(); n != 0 {
		t.Errorf("hand = %d after manifesting from an empty hand", n)
	}
}

// CR 701.40g: a manifested instant stays face down. It offers no way
// up (CR 701.40b: only a creature card has one), and an effect that
// turns it up does nothing.
func TestAManifestedInstantFromAHandStaysFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emptyHandForTest(g, me)
	bolt := handCardForTest(me, "Lightning Bolt", "Instant", "")
	scroll := pushCatalogPermanent(g, me.ID, "Scroll of Fate", "Artifact", scrollOfFateOracle, false)
	toMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := findCardAnywhere(t, g, bolt)
	if !c.FaceDownIsPermanent() || !c.IsCreature() {
		t.Fatal("the instant did not become a face-down creature on the battlefield (CR 701.40a)")
	}
	if offer := game.TurnFaceUpOffer(c); offer != nil {
		t.Errorf("a manifested instant offers to turn up for %q", offer.Cost)
	}
	var turned bool
	g.WithWriteLock(func() { turned = g.TurnFaceUpForEffect(scroll, me.ID, bolt) })
	if turned {
		t.Error("an effect turned a manifested instant face up")
	}
	if c := findCardAnywhere(t, g, bolt); !c.FaceDown {
		t.Error("the instant is face up")
	}
}

// A manifested CREATURE card from a hand keeps the turn-face-up route
// for its mana cost (CR 701.40b).
func TestAManifestedCreatureFromAHandCanBeTurnedUpForItsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emptyHandForTest(g, me)
	id := putInHand(me, game.Card{Name: "Serra Angel", TypeLine: "Creature — Angel", ManaCost: "{3}{W}{W}", Power: 4, Toughness: 4})
	g.WithWriteLock(func() {
		if err := g.ManifestFromHandsKeyedForEffect([]uuid.UUID{id}, game.KeyedTail{}); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	offer := game.TurnFaceUpOffer(findCardAnywhere(t, g, id))
	if offer == nil || offer.Cost != "{3}{W}{W}" {
		t.Fatalf("offer = %+v, want the mana cost {3}{W}{W}", offer)
	}
}
