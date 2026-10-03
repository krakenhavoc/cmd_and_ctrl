package effects

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_hand_cards_test.go — #1600's proof cards: Lion's Eye Diamond
// and Diamond Lion (the mana-ability owner of "Discard your hand", and
// "Activate only as an instant"), Null Brooch and Slate of Ancestry
// (the CR 602 owner). The engine half is pinned in
// game/discard_hand_cost_test.go.

const (
	lionsEyeDiamondOracle = "ee6099b0-fb1f-42f1-b862-7708c6e36d05"
	diamondLionOracle     = "7ea9bb3b-76e9-4150-9e4f-1cd3abfe8ad7"
	nullBroochOracle      = "6f885041-3e57-4a69-84f2-fd207ff9f31b"
	slateOfAncestryOracle = "a07483b4-c04f-42a4-b979-8b77c11fa8f5"
)

func TestDiscardHandCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lionsEyeDiamondOracle: "Lion's Eye Diamond",
		diamondLionOracle:     "Diamond Lion",
		nullBroochOracle:      "Null Brooch",
		slateOfAncestryOracle: "Slate of Ancestry",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered as %q, completeness %v", name, spec.Name, spec.Completeness)
		}
	}
}

// End to end: the whole hand is discarded, the Diamond is sacrificed,
// and ONE colour pick adds three mana of that colour.
func TestLionsEyeDiamondAddsThreeManaOfOneChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if me.Hand.Size() == 0 {
		t.Fatal("test premise: a full hand")
	}
	hand := make([]uuid.UUID, 0, me.Hand.Size())
	for _, c := range me.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	led := b31Push(g, me.ID, "Lion's Eye Diamond", "Artifact", lionsEyeDiamondOracle, "{0}", 0, 0)

	if err := g.ActivateManaAbility(me.ID, led, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Hand.Size() != 0 {
		t.Errorf("%d cards left in hand", me.Hand.Size())
	}
	for _, id := range hand {
		if !me.Graveyard.Contains(id) {
			t.Errorf("hand card %s was not discarded", id)
		}
	}
	if g.Battlefield.Contains(led) || !me.Graveyard.Contains(led) {
		t.Error("the Diamond was not sacrificed")
	}
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || len(pick.ColorOptions) != 5 || pick.ManaAmounts["U"] != 3 {
		t.Fatalf("pick = %+v, want one five-colour pick of three", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"U", "U", "U"}) {
		t.Errorf("pool = %v, want three {U}", got)
	}
}

// CR 118.3: an empty hand pays "Discard your hand". The colour can be
// named up front (#1443), which skips the pick.
func TestLionsEyeDiamondWithAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	led := b31Push(g, me.ID, "Lion's Eye Diamond", "Artifact", lionsEyeDiamondOracle, "{0}", 0, 0)

	if err := g.ActivateManaAbility(me.ID, led, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("an empty hand must pay the cost: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"R", "R", "R"}) {
		t.Errorf("pool = %v, want three {R}", got)
	}
}

// The hand is discarded as a COST, so Library of Leng — "If an EFFECT
// causes you to discard a card" — does not apply: every card goes to
// the graveyard and no "put it on top instead?" prompt is offered.
func TestLionsEyeDiamondIsNotReplacedByLibraryOfLeng(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b31Push(g, me.ID, "Library of Leng", "Artifact", "867def48-4be8-4056-bcf1-d6b00450b9a3", "{1}", 0, 0)
	hand := make([]uuid.UUID, 0, me.Hand.Size())
	for _, c := range me.Hand.Cards {
		hand = append(hand, c.InstanceID)
	}
	led := b31Push(g, me.ID, "Lion's Eye Diamond", "Artifact", lionsEyeDiamondOracle, "{0}", 0, 0)

	if err := g.ActivateManaAbility(me.ID, led, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	for _, id := range hand {
		if !me.Graveyard.Contains(id) {
			t.Errorf("hand card %s is not in the graveyard; a cost discard is not Library of Leng's", id)
		}
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("%d prompts open after a cost discard, want none: %+v", n, g.PendingChoices)
	}
}

// "Activate only as an instant": not while another player holds
// priority, and not while the controller is answering a prompt — the
// Diamond's own colour pick included, so a second Diamond waits for the
// first to finish resolving. Allowed in response to the controller's
// own spell, which is how the card is played.
func TestLionsEyeDiamondOnlyAsAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	first := b31Push(g, me.ID, "Lion's Eye Diamond", "Artifact", lionsEyeDiamondOracle, "{0}", 0, 0)
	second := b31Push(g, me.ID, "Lion's Eye Diamond", "Artifact", lionsEyeDiamondOracle, "{0}", 0, 0)

	// My own sorcery on the stack, my priority: the window is open.
	castCatalogSpell(t, g, "Divination", "Sorcery", "no-such-oracle-divination", nil)
	size := me.Hand.Size()
	if err := g.ActivateManaAbility(me.ID, first, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("in response to my own spell: %v", err)
	}
	if me.Hand.Size() != 0 || size == 0 {
		t.Fatalf("hand %d → %d, want it emptied", size, me.Hand.Size())
	}
	// The first Diamond's colour pick is open and mine to answer.
	if err := g.ActivateManaAbility(me.ID, second, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("while answering a pick: err = %v, want ErrConditionNotMet", err)
	}
	if !g.Battlefield.Contains(second) {
		t.Fatal("a refused activation sacrificed the Diamond")
	}
	if n := b10ResolveAllManaPicks(t, g, me.ID, "B"); n != 1 {
		t.Fatalf("answered %d picks, want 1", n)
	}

	// Another player holds priority: shut.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, second, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("without priority: err = %v, want ErrConditionNotMet", err)
	}
	if !g.Battlefield.Contains(second) {
		t.Error("a refused activation sacrificed the Diamond")
	}
}

// The auto-tapper never cracks a Diamond to pay for a spell — the
// mid-cast window "Activate only as an instant" forbids (Diamond Lion's
// 2021-06-18 ruling) — not even with an empty hand.
func TestLionsEyeDiamondIsNeverAutoTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	led := b31Push(g, me.ID, "Lion's Eye Diamond", "Artifact", lionsEyeDiamondOracle, "{0}", 0, 0)
	lion := b31Push(g, me.ID, "Diamond Lion", "Artifact Creature — Cat", diamondLionOracle, "{2}", 2, 2)

	cost, err := game.ParseCost("{R}")
	if err != nil {
		t.Fatal(err)
	}
	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Errorf("auto-tap planned %v for {R}; neither Diamond may be a planned source", plan)
	}
	if !g.Battlefield.Contains(led) || !g.Battlefield.Contains(lion) {
		t.Error("planning spent a Diamond")
	}
}

// Diamond Lion: the {T} is a creature's, so summoning sickness stops it
// (CR 302.6) with nothing paid; a Lion that has been around since the
// turn began taps, discards, sacrifices and adds three.
func TestDiamondLion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sick := pushCatalogPermanent(g, me.ID, "Diamond Lion", "Artifact Creature — Cat", diamondLionOracle, true)
	size := me.Hand.Size()
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrSummoningSick) {
		t.Fatalf("summoning-sick Lion: err = %v, want ErrSummoningSick", err)
	}
	if me.Hand.Size() != size || !g.Battlefield.Contains(sick) {
		t.Fatal("a refused activation paid part of its cost")
	}

	lion := pushCatalogPermanent(g, me.ID, "Diamond Lion", "Artifact Creature — Cat", diamondLionOracle, false)
	if err := g.ActivateManaAbility(me.ID, lion, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Size() != 0 || g.Battlefield.Contains(lion) {
		t.Error("the hand was not discarded or the Lion not sacrificed")
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"G", "G", "G"}) {
		t.Errorf("pool = %v, want three {G}", got)
	}
}

// Null Brooch: the hand goes at announce, before anyone can respond,
// and the ability counters a noncreature spell when it resolves.
func TestNullBroochCountersANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brooch := b31Push(g, me.ID, "Null Brooch", "Artifact", nullBroochOracle, "{4}", 0, 0)
	bolt := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	life := me.Life
	size := me.Hand.Size()
	if size == 0 {
		t.Fatal("test premise: a full hand")
	}

	// A Brooch that is already tapped pays nothing: every component is
	// validated before any is paid.
	setTapped := func(v bool) {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == brooch {
					g.Battlefield.Cards[i].Tapped = v
				}
			}
		})
	}
	setTapped(true)
	if err := g.ActivateCatalogAbility(me.ID, brooch, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err == nil {
		t.Fatal("a tapped Brooch was activated")
	}
	if me.Hand.Size() != size {
		t.Fatal("a refused activation discarded the hand")
	}
	setTapped(false)

	b10AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, brooch, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Size() != 0 {
		t.Errorf("%d cards in hand after the announce; the cost discards them all", me.Hand.Size())
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("the Bolt resolved: life %d → %d", life, me.Life)
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
}

// Slate of Ancestry: an empty hand pays, and the draw counts the
// creatures the activator controls when the ability resolves.
func TestSlateOfAncestryDrawsACardPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slate := b31Push(g, me.ID, "Slate of Ancestry", "Artifact", slateOfAncestryOracle, "{4}", 0, 0)
	for i := 0; i < 3; i++ {
		b31Push(g, me.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	}
	b31Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2)
	me.Hand.Cards = nil
	b10AddMana(me, "C", "C", "C", "C")

	if err := g.ActivateCatalogAbility(me.ID, slate, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("an empty hand must pay the cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand has %d cards, want 3 — one per creature you control", got)
	}
}

// The boot-time rules for the clause (discard_hand_cost.go).
func TestRegisterHoldsTheDiscardYourHandClause(t *testing.T) {
	expectRegisterPanic(t, "build it with DiscardYourHand", func() {
		checkDiscardClause("Fixture", "ability 0", &game.DiscardCost{Hand: true, N: 1})
	})
	expectRegisterPanic(t, "build it with DiscardYourHand", func() {
		checkDiscardClause("Fixture", "ability 0", &game.DiscardCost{Hand: true, Random: true})
	})
	expectRegisterPanic(t, "build it with DiscardYourHand", func() {
		checkDiscardClause("Fixture", "ability 0", &game.DiscardCost{Hand: true, Match: func(game.Card) bool { return true }})
	})
	// The count form keeps its old refusal.
	expectRegisterPanic(t, "discards at least one", func() {
		checkDiscardClause("Fixture", "mana ability 0", &game.DiscardCost{N: 0, Label: "a card"})
	})
	checkDiscardClause("Fixture", "ability 0", DiscardYourHand().DiscardCards)

	expectRegisterPanic(t, "also spends a card from it", func() {
		checkDiscardHandBesideHandCosts("Fixture", "ability 0", Plus(DiscardYourHand(), PutACardFromHandOnTop()))
	})
	expectRegisterPanic(t, "also spends a card from it", func() {
		checkDiscardHandBesideHandCosts("Fixture", "ability 0", Plus(DiscardYourHand(), DiscardThis()))
	})
	expectRegisterPanic(t, "also spends a card from it", func() {
		checkManaDiscardHandBesideHandCosts("Fixture", "mana ability 0",
			ManaAbilityCost{ExileSelf: true, DiscardCards: DiscardYourHand().DiscardCards})
	})
	checkManaDiscardHandBesideHandCosts("Fixture", "mana ability 0",
		ManaAbilityCost{Sacrifice: true, DiscardCards: DiscardYourHand().DiscardCards})
}
