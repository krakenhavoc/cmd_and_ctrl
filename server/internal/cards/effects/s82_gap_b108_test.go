package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b108_test.go — S82 GAP batch 108 (#5948). Each test plays the real
// catalog entry.

const (
	gapB108SunQuanOracle   = "70503280-6bdf-431d-8213-367a9783463a"
	gapB108SynapseOracle   = "b76f123d-7b78-42a1-a66e-0dee6f773687"
	gapB108TauntingOracle  = "fada8f16-85e8-44f1-9d2c-99da1e27851f"
	gapB108TromokratisOrcl = "9aa2f27b-4c00-44c1-8bc4-014350f05b6d"
	gapB108VizierOracle    = "699fc7cf-1ad3-4c32-b3e4-49f9a2da4b26"
	gapB108WindriderOracle = "78e65286-204d-43ab-8929-b8e2191b1fce"
	gapB108WingedOracle    = "24e96682-14ca-4a62-a920-2d686c8a5dbf"
	gapB108HawkOracle      = "ac8cf699-f67b-44da-bb79-eb95cfc0026d"
)

func TestGapB108CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB108SunQuanOracle:   "Sun Quan, Lord of Wu",
		gapB108SynapseOracle:   "Synapse Sliver",
		gapB108TauntingOracle:  "Taunting Sliver",
		gapB108TromokratisOrcl: "Tromokratis",
		gapB108VizierOracle:    "Vizier of Tumbling Sands",
		gapB108WindriderOracle: "Windrider Wizard",
		gapB108WingedOracle:    "Winged Sliver",
		gapB108HawkOracle:      "Cartographer's Hawk",
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

// gapB108Attack declares attacker at the opponent and moves to the declare
// blockers step, ready for blocks.
func gapB108Attack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
}

// --- Sun Quan, Lord of Wu ----------------------------------------------------

func TestGapB108SunQuanGivesYourCreaturesHorsemanshipAndOnlyYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	quan := b12Push(g, me.ID, "Sun Quan, Lord of Wu", "Legendary Creature — Human Soldier", gapB108SunQuanOracle, 4, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{quan, mine} {
		if !hasKeywordOnBattlefield(t, g, id, "horsemanship") {
			t.Errorf("%s lacks horsemanship", id)
		}
	}
	if hasKeywordOnBattlefield(t, g, theirs, "horsemanship") {
		t.Error("an opponent's creature has horsemanship")
	}
	gapB108Attack(t, g, mine, opp.ID)
	if err := g.DeclareBlocker(theirs, mine); err == nil {
		t.Error("a creature without horsemanship blocked one with it")
	}
}

// --- Winged Sliver -----------------------------------------------------------

func TestGapB108WingedSliverGivesEverySliverFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	winged := b12Push(g, me.ID, "Winged Sliver", "Creature — Sliver", gapB108WingedOracle, 1, 1)
	mine := b12Creature(g, me.ID, "My Sliver", "Creature — Sliver", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Sliver", "Creature — Sliver", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{winged, mine, theirs} {
		if !hasKeywordOnBattlefield(t, g, id, "flying") {
			t.Errorf("Sliver %s lacks flying", id)
		}
	}
	if hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("a Bear has flying")
	}
}

// --- Synapse Sliver ----------------------------------------------------------

func TestGapB108SynapseSliverLetsTheSliversControllerDraw(t *testing.T) {
	g := newCatalogGame(t)
	attacker, victim, owner := g.Seats[0], g.Seats[1], g.Seats[2]
	syn := b12Push(g, owner.ID, "Synapse Sliver", "Creature — Sliver", gapB108SynapseOracle, 3, 3)
	sliver := b12Creature(g, attacker.ID, "Attacking Sliver", "Creature — Sliver", 2, 2)
	seedLibrary(attacker, "A", "B", "C")
	seedLibrary(owner, "A", "B", "C")
	hand, ownerHand := attacker.Hand.Size(), owner.Hand.Size()
	attackWith(t, g, victim.ID, sliver)
	g.RunStateChecksForTest()
	if triggerOnStack(g, syn) == nil {
		t.Fatal("no Synapse Sliver trigger")
	}
	answerLatestTriggerPrompt(t, g, attacker.ID, true)
	passPriorityAroundTable(t, g)
	if attacker.Hand.Size() != hand+1 || owner.Hand.Size() != ownerHand {
		t.Errorf("attacker hand %d -> %d, Synapse's controller %d -> %d, want the Sliver's controller to draw",
			hand, attacker.Hand.Size(), ownerHand, owner.Hand.Size())
	}
}

func TestGapB108SynapseSliverIgnoresNonSlivers(t *testing.T) {
	g := newCatalogGame(t)
	attacker, victim := g.Seats[0], g.Seats[1]
	syn := b12Push(g, attacker.ID, "Synapse Sliver", "Creature — Sliver", gapB108SynapseOracle, 3, 3)
	bear := b12Creature(g, attacker.ID, "My Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, victim.ID, bear)
	g.RunStateChecksForTest()
	if triggerOnStack(g, syn) != nil {
		t.Error("a Bear's combat damage triggered the Synapse Sliver")
	}
}

// --- Taunting Sliver ---------------------------------------------------------

func TestGapB108TauntingSliverGoadsWhenASliverEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Taunting Sliver", "Creature — Sliver", gapB108TauntingOracle, 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	toMain(t, g)
	castCatalogSpell(t, g, "Plain Sliver", "Creature — Sliver", "", nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p == nil || hasID(p.PickTargetCards, mine) {
		t.Fatalf("pick prompt %+v, want one that offers only an opponent's creature", p)
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if b33Goaded(t, g, theirs) != me.ID {
		t.Error("the opponent's creature is not goaded by the Sliver's controller")
	}
}

func TestGapB108TauntingSliverIgnoresANonSliverEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Taunting Sliver", "Creature — Sliver", gapB108TauntingOracle, 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	toMain(t, g)
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if b33Goaded(t, g, theirs) != uuid.Nil {
		t.Error("a Bear entering goaded a creature")
	}
}

// --- Tromokratis -------------------------------------------------------------

func TestGapB108TromokratisIsHexproofUnlessAttackingOrBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tromo := b12Push(g, me.ID, "Tromokratis", "Legendary Creature — Kraken", gapB108TromokratisOrcl, 8, 8)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasKeywordOnBattlefield(t, g, tromo, "hexproof") {
		t.Fatal("an idle Tromokratis lacks hexproof")
	}
	gapB108Attack(t, g, tromo, opp.ID)
	if hasKeywordOnBattlefield(t, g, tromo, "hexproof") {
		t.Error("an attacking Tromokratis still has hexproof")
	}
}

func TestGapB108TromokratisLosesHexproofWhileBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tromo := b12Push(g, opp.ID, "Tromokratis", "Legendary Creature — Kraken", gapB108TromokratisOrcl, 8, 8)
	bear := b12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	if !hasKeywordOnBattlefield(t, g, tromo, "hexproof") {
		t.Fatal("an idle Tromokratis lacks hexproof")
	}
	brAttack(t, g, bear)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: tromo, Attacker: bear}}); err != nil {
		t.Fatalf("Tromokratis blocking: %v", err)
	}
	lockInBlocks(t, g)
	if hasKeywordOnBattlefield(t, g, tromo, "hexproof") {
		t.Error("a blocking Tromokratis still has hexproof")
	}
}

func TestGapB108TromokratisIsBlockableOnlyByEveryDefendingCreatureAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tromo := b12Push(g, me.ID, "Tromokratis", "Legendary Creature — Kraken", gapB108TromokratisOrcl, 8, 8)
	one := b12Creature(g, opp.ID, "Blocker One", "Creature — Bear", 2, 2)
	two := b12Creature(g, opp.ID, "Blocker Two", "Creature — Bear", 2, 2)
	three := b12Creature(g, opp.ID, "Blocker Three", "Creature — Bear", 2, 2)
	brAttack(t, g, tromo)
	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: one, Attacker: tromo}, {Blocker: two, Attacker: tromo},
	}), game.BlockReasonTooFewBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: one, Attacker: tromo}, {Blocker: two, Attacker: tromo}, {Blocker: three, Attacker: tromo},
	}); err != nil {
		t.Fatalf("all three defenders blocking is legal: %v", err)
	}
}

func TestGapB108TromokratisCantBeBlockedWhenADefenderCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tromo := b12Push(g, me.ID, "Tromokratis", "Legendary Creature — Kraken", gapB108TromokratisOrcl, 8, 8)
	one := b12Creature(g, opp.ID, "Blocker One", "Creature — Bear", 2, 2)
	tapped := b12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	brAttack(t, g, tromo)
	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{{Blocker: one, Attacker: tromo}}), game.BlockReasonTooFewBlockers)
}

func TestGapB108TromokratisIsFreelyBlockableByALoneDefender(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tromo := b12Push(g, me.ID, "Tromokratis", "Legendary Creature — Kraken", gapB108TromokratisOrcl, 8, 8)
	only := b12Creature(g, opp.ID, "Only Bear", "Creature — Bear", 2, 2)
	brAttack(t, g, tromo)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: only, Attacker: tromo}}); err != nil {
		t.Fatalf("a lone defender blocking with everything it controls is legal: %v", err)
	}
}

func TestGapB108TromokratisOnlyBindsItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tromokratis", "Legendary Creature — Kraken", gapB108TromokratisOrcl, 8, 8)
	bear := b12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	one := b12Creature(g, opp.ID, "Blocker One", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Blocker Two", "Creature — Bear", 2, 2)
	brAttack(t, g, bear)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: one, Attacker: bear}}); err != nil {
		t.Fatalf("another attacker needs no extra blockers: %v", err)
	}
}

// --- Vizier of Tumbling Sands ------------------------------------------------

func TestGapB108VizierUntapsAnotherPermanentButNotItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vizier := b12Push(g, me.ID, "Vizier of Tumbling Sands", "Creature — Human Cleric", gapB108VizierOracle, 1, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
			if g.Battlefield.Cards[i].InstanceID == vizier {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := ravB2Activate(t, g, me.ID, vizier, 0, game.ActivateAbilityParams{Targets: cardRefs(vizier)}); err == nil {
		t.Error("the Vizier targeted itself")
	}
	ravB2MustActivate(t, g, me.ID, vizier, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if tmtB1Card(t, g, bear).Tapped {
		t.Error("the Bear is still tapped")
	}
}

func TestGapB108VizierCyclingUntapsAnyTargetPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b12Permanent(g, me.ID, "Tapped Land", "Land")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	seedLibrary(me, "A", "B")
	toMain(t, g)
	id := pushCatalogHandCard(me, "Vizier of Tumbling Sands", "Creature — Human Cleric", gapB108VizierOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{U}"); err != nil {
		t.Fatal(err)
	}
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if tmtB1Card(t, g, land).Tapped {
		t.Error("the cycled Vizier did not untap the land")
	}
	if got := len(me.Hand.Cards); got != hand {
		t.Errorf("hand %d -> %d, want the Vizier out and a card in", hand, got)
	}
}

// --- Windrider Wizard --------------------------------------------------------

func TestGapB108WindriderWizardOffersALootOnInstantSorceryAndWizardSpellsOnly(t *testing.T) {
	for _, c := range []struct {
		name, typeLine string
		want           bool
	}{
		{"Quick Instant", "Instant", true},
		{"Slow Sorcery", "Sorcery", true},
		{"Other Wizard", "Creature — Human Wizard", true},
		{"Plain Bear", "Creature — Bear", false},
		{"Plain Artifact", "Artifact", false},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		wiz := b12Push(g, me.ID, "Windrider Wizard", "Creature — Human Wizard", gapB108WindriderOracle, 2, 2)
		if !hasKeywordOnBattlefield(t, g, wiz, "flying") {
			t.Error("Windrider Wizard lacks flying")
		}
		msc18Library(me, 8)
		toMain(t, g)
		id := mayhemHandCard(g, me, c.name, c.typeLine, "{0}", "")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", c.name, err)
		}
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
		ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
		if got := ask != nil && ask.Chooser == me.ID; got != c.want {
			t.Errorf("%s: loot offered = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestGapB108WindriderWizardDrawsThenDiscardsWhenAccepted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Windrider Wizard", "Creature — Human Wizard", gapB108WindriderOracle, 2, 2)
	msc18Library(me, 8)
	toMain(t, g)
	id := mayhemHandCard(g, me, "Quick Instant", "Instant", "{0}", "")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no loot prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

// --- Cartographer's Hawk -----------------------------------------------------

func TestGapB108CartographersHawkReturnsAndFetchesAPlainsWhenTheyHaveMoreLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := b12Push(g, me.ID, "Cartographer's Hawk", "Creature — Bird", gapB108HawkOracle, 2, 1)
	b12Permanent(g, opp.ID, "Their Land One", "Land")
	b12Permanent(g, opp.ID, "Their Land Two", "Land")
	b12Permanent(g, me.ID, "My Land", "Land")
	ids := seedSearchLibrary(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"}, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	attackWith(t, g, opp.ID, hawk)
	g.RunStateChecksForTest()
	if triggerOnStack(g, hawk) == nil {
		t.Fatal("no Hawk trigger although the defender controls more lands")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after the Hawk returned")
	}
	if hasID(c.SearchCards, ids[1]) {
		t.Error("an Island was offered to a search for a Plains card")
	}
	answerSearchByID(t, g, me.ID, ids[0])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(hawk) {
		t.Error("the Hawk is not in its owner's hand")
	}
	plains, ok := g.LookupCardForEffect(ids[0])
	if !ok || !g.Battlefield.Contains(ids[0]) || !plains.Tapped {
		t.Error("the Plains did not enter tapped")
	}
}

func TestGapB108CartographersHawkDoesNothingWithoutMoreLandsOnTheirSide(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := b12Push(g, me.ID, "Cartographer's Hawk", "Creature — Bird", gapB108HawkOracle, 2, 1)
	b12Permanent(g, opp.ID, "Their Land", "Land")
	b12Permanent(g, me.ID, "My Land", "Land")
	attackWith(t, g, opp.ID, hawk)
	g.RunStateChecksForTest()
	if triggerOnStack(g, hawk) != nil {
		t.Fatal("the Hawk triggered with equal land counts")
	}
}
