package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The headline: two targets, five cards and one card. The first is
// asked for exactly two; the second is not asked and manifests their
// lone card; both are face down under their owners; the caster draws
// three.
func TestKozilekAsksTheFirstForTwoAndManifestsTheSecondsLoneCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	firstHand := effectstest.FillHandForTest(g, first, 5)
	secondHand := effectstest.FillHandForTest(g, second, 1)

	effectstest.CastKozilek(t, g, first.ID, second.ID)
	handBefore := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, first.ID)
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
	if effectstest.ChooseCardsChoiceFor(g, second.ID) != nil {
		t.Error("the one-card hand was asked a question; the ruling is that they manifest that card")
	}
	if effectstest.CardInPlay(g, secondHand[0]) {
		t.Error("the lone card entered before the first target answered; the run must wait for every seat")
	}

	effectstest.AnswerChooseCards(t, g, first.ID, firstHand[1], firstHand[2])
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertManifestedBy(t, g, firstHand[1], first.ID)
	effectstest.AssertManifestedBy(t, g, firstHand[2], first.ID)
	effectstest.AssertManifestedBy(t, g, secondHand[0], second.ID)
	for _, id := range []uuid.UUID{firstHand[0], firstHand[3], firstHand[4]} {
		if effectstest.CardInPlay(g, id) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := effectstest.FillHandForTest(g, opp, 1)

	id := effectstest.CastKozilek(t, g, opp.ID)
	handBefore := me.Hand.Size()
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(id); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertManifestedBy(t, g, hand[0], opp.ID)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("caster's hand %d -> %d, want one card drawn", handBefore, got)
	}
}

// "Up to two": with no target nothing is manifested and nothing drawn.
func TestKozilekWithNoTargetsManifestsAndDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := effectstest.FillHandForTest(g, opp, 4)

	effectstest.CastKozilek(t, g)
	handBefore := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range hand {
		if effectstest.CardInPlay(g, id) {
			t.Errorf("card %s was manifested with no target", id)
		}
	}
	if effectstest.ChooseCardsChoiceFor(g, opp.ID) != nil {
		t.Error("an untargeted player was asked to manifest")
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("caster's hand %d -> %d, want no draw", handBefore, got)
	}
}

// A target with an empty hand manifests nothing, so the caster draws
// for the other target only.
func TestKozilekSkipsAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	empty, one := g.Seats[2], g.Seats[3]
	effectstest.EmptyHandForTest(g, empty)
	hand := effectstest.FillHandForTest(g, one, 2)

	effectstest.CastKozilek(t, g, empty.ID, one.ID)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.ChooseCardsChoiceFor(g, one.ID) != nil {
		t.Error("a two-card hand was asked about two cards")
	}
	effectstest.AssertManifestedBy(t, g, hand[0], one.ID)
	effectstest.AssertManifestedBy(t, g, hand[1], one.ID)
	if got := caster.Hand.Size(); got != handBefore+2 {
		t.Errorf("caster's hand %d -> %d, want +2", handBefore, got)
	}
}

// A replacement that keeps one card off the battlefield lowers the draw
// by one: "for each card manifested this way" counts what arrived.
func TestKozilekDrawsOnlyForTheCardsThatArrive(t *testing.T) {
	const keeper = "test-3538-hand-keeper"
	var kept uuid.UUID
	effectstest.RegisterForTest(t, effects.Spec{
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Test Hand Keeper", "Enchantment", keeper, false)
	hand := effectstest.FillHandForTest(g, opp, 2)
	kept = hand[0]

	effectstest.CastKozilek(t, g, opp.ID)
	handBefore := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.CardInPlay(g, kept) {
		t.Fatal("the replacement did not keep the card off the battlefield")
	}
	effectstest.AssertManifestedBy(t, g, hand[1], opp.ID)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("caster's hand %d -> %d, want +1: only one card arrived", handBefore, got)
	}
}

// "Other colorless creatures you control get +3/+2": a face-down 2/2
// you control is 5/4, Kozilek itself stays 9/9, and neither a colored
// creature nor an opponent's colorless one is touched.
func TestKozilekPumpsOtherColorlessCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := effectstest.FillHandForTest(g, me, 1)

	kozilek := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kozilek, the Broken Reality", TypeLine: "Legendary Creature — Eldrazi",
		OracleID: effectstest.KozilekBrokenRealityOracle, Power: 9, Toughness: 9, Owner: me.ID, Controller: me.ID,
	})
	scion := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Eldrazi Scion", TypeLine: "Creature — Eldrazi Scion",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
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
		if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != power || tg != toughness {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := effectstest.FillHandForTest(g, me, 3)
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Scroll of Fate", "Artifact", effectstest.ScrollOfFateOracle, false)
	effectstest.ToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, scroll); !tapped {
		t.Error("the scroll was not tapped to pay for its ability")
	}
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Scroll of Fate did not ask which card to manifest")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 || pick.ChooseDestination != game.ChooseManifest {
		t.Errorf("prompt = %d..%d to %q, want exactly one, to manifest", pick.ChooseMin, pick.ChooseMax, pick.ChooseDestination)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, hand[1])
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertManifestedBy(t, g, hand[1], me.ID)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand = %d, want the 2 unpicked", got)
	}
}

// A one-card hand is not asked, and an empty hand manifests nothing.
func TestScrollOfFateWithOneCardOrNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := effectstest.FillHandForTest(g, me, 1)
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Scroll of Fate", "Artifact", effectstest.ScrollOfFateOracle, false)
	effectstest.ToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a one-card hand was asked which card to manifest")
	}
	effectstest.AssertManifestedBy(t, g, hand[0], me.ID)

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
	effectstest.PassPriorityAroundTable(t, g)
	if n := me.Hand.Size(); n != 0 {
		t.Errorf("hand = %d after manifesting from an empty hand", n)
	}
}

// CR 701.40g: a manifested instant stays face down. It offers no way
// up (CR 701.40b: only a creature card has one), and an effect that
// turns it up does nothing.
func TestAManifestedInstantFromAHandStaysFaceDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.EmptyHandForTest(g, me)
	bolt := effectstest.HandCardForTest(me, "Lightning Bolt", "Instant", "")
	scroll := effectstest.PushCatalogPermanent(g, me.ID, "Scroll of Fate", "Artifact", effectstest.ScrollOfFateOracle, false)
	effectstest.ToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.FindCardAnywhere(t, g, bolt)
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
	if c := effectstest.FindCardAnywhere(t, g, bolt); !c.FaceDown {
		t.Error("the instant is face up")
	}
}

// A manifested CREATURE card from a hand keeps the turn-face-up route
// for its mana cost (CR 701.40b).
func TestAManifestedCreatureFromAHandCanBeTurnedUpForItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.EmptyHandForTest(g, me)
	id := effectstest.PutInHand(me, game.Card{Name: "Serra Angel", TypeLine: "Creature — Angel", ManaCost: "{3}{W}{W}", Power: 4, Toughness: 4})
	g.WithWriteLock(func() {
		if err := g.ManifestFromHandsKeyedForEffect([]uuid.UUID{id}, game.KeyedTail{}); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	offer := game.TurnFaceUpOffer(effectstest.FindCardAnywhere(t, g, id))
	if offer == nil || offer.Cost != "{3}{W}{W}" {
		t.Fatalf("offer = %+v, want the mana cost {3}{W}{W}", offer)
	}
}
