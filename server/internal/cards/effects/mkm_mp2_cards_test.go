package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_mp2_cards_test.go — Murders at Karlov Manor cards M–P, second batch
// (#4469). Each test plays the real catalog entry and asserts what its
// text says.

const (
	mkmMP2MacabreOracle     = "04633ecd-eac0-426a-83a5-8e88262fca30"
	mkmMP2SnufflerOracle    = "8f85088c-256d-4925-a4dd-aafcc2632282"
	mkmMP2MassacreOracle    = "ccb39f4a-2684-4eb7-9066-aad81ae08962"
	mkmMP2YouthsOracle      = "9acb74a9-f372-43ad-a80d-4f8d52236eb8"
	mkmMP2MelekOracle       = "23140c6f-c674-4d13-9b78-bce0ffcb981f"
	mkmMP2SpyOracle         = "87ccb151-a948-4e5a-878b-0d46c921a2b7"
	mkmMP2NoWitnessesOracle = "0639bd82-084f-47d5-b158-05dfc29efa40"
	mkmMP2OfficiousOracle   = "e57e448b-1186-4662-a6b0-01a9ce33b961"
	mkmMP2OutColdOracle     = "2d16d51b-47e0-4cdc-83dd-a805ba50bd83"
	mkmMP2EnforcerOracle    = "be8d3788-709b-4d10-861c-a6f27d4b63a6"
	mkmMP2PersuasiveOracle  = "b401179f-eded-4973-8182-df4627133886"
	mkmMP2GadaboutOracle    = "0361b9b7-9097-4aae-aa06-c63a5fb47a76"
	mkmMP2PrivateEyeOracle  = "20750317-061b-4b65-9fc8-43477fcd7a21"
	mkmMP2ProjektorOracle   = "58b27a79-ba63-45d4-84ee-423ce0d91f6b"
	mkmMP2ThoroughfareOracl = "de5b995c-9691-4555-9070-66bcbc29f955"

	mkmMP2SpyTypeLine = "Creature — Merfolk Detective"
)

func TestMkmMP2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmMP2MacabreOracle:     "Macabre Reconstruction",
		mkmMP2SnufflerOracle:    "Magnetic Snuffler",
		mkmMP2MassacreOracle:    "Massacre Girl, Known Killer",
		mkmMP2YouthsOracle:      "Meddling Youths",
		mkmMP2MelekOracle:       "Melek, Reforged Researcher",
		mkmMP2SpyOracle:         "Mistway Spy",
		mkmMP2NoWitnessesOracle: "No Witnesses",
		mkmMP2OfficiousOracle:   "Officious Interrogation",
		mkmMP2OutColdOracle:     "Out Cold",
		mkmMP2EnforcerOracle:    "Perimeter Enforcer",
		mkmMP2PersuasiveOracle:  "Persuasive Interrogators",
		mkmMP2GadaboutOracle:    "Pompous Gadabout",
		mkmMP2PrivateEyeOracle:  "Private Eye",
		mkmMP2ProjektorOracle:   "Projektor Inspector",
		mkmMP2ThoroughfareOracl: "Public Thoroughfare",
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

// --- shared helpers ------------------------------------------------------------

// mkmMP2Grave puts a plain card into a player's graveyard.
func mkmMP2Grave(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// mkmMP2Clues lists the Clues a player controls.
func mkmMP2Clues(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" && c.Controller == controller {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- Macabre Reconstruction ----------------------------------------------------

func TestMkmMP2MacabreReconstructionCostsTwoLessAfterACreatureCardHitYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := selfPricedMV(t, g, me, mkmMP2MacabreOracle, "Sorcery", "{3}{B}"); got != 4 {
		t.Fatalf("with nothing in the graveyard this turn it costs %d, want 4", got)
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if got := selfPricedMV(t, g, me, mkmMP2MacabreOracle, "Sorcery", "{3}{B}"); got != 4 {
		t.Errorf("an opponent's creature dying made it cost %d, want 4", got)
	}
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if got := selfPricedMV(t, g, me, mkmMP2MacabreOracle, "Sorcery", "{3}{B}"); got != 2 {
		t.Errorf("after my creature card died it costs %d, want 2", got)
	}
}

func TestMkmMP2MacabreReconstructionIgnoresANoncreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	if got := selfPricedMV(t, g, me, mkmMP2MacabreOracle, "Sorcery", "{3}{B}"); got != 4 {
		t.Errorf("an artifact card reaching the graveyard made it cost %d, want 4", got)
	}
}

func TestMkmMP2MacabreReconstructionReturnsTwoCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := mkmMP2Grave(me, "Bear", "Creature — Bear")
	b := mkmMP2Grave(me, "Elk", "Creature — Elk")
	castCatalogSpell(t, g, "Macabre Reconstruction", "Sorcery", mkmMP2MacabreOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not reach the hand")
	}
}

// --- Magnetic Snuffler ---------------------------------------------------------

func TestMkmMP2MagneticSnufflerReturnsAnEquipmentAttachedToItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sword := mkmMP2Grave(me, "Test Sword", "Artifact — Equipment")
	snuffler := castHoldCreature(t, g, "Magnetic Snuffler", "Artifact Creature — Construct", mkmMP2SnufflerOracle, 4, 4)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, sword)
	if !ok {
		t.Fatal("the Equipment did not return to the battlefield")
	}
	if c.AttachedTo.ID != snuffler || c.Controller != me.ID {
		t.Errorf("the Equipment is attached to %v under %v, want the Snuffler under me", c.AttachedTo.ID, c.Controller)
	}
}

func TestMkmMP2MagneticSnufflerGrowsWhenYouSacrificeAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	snuffler := b12Push(g, me.ID, "Magnetic Snuffler", "Artifact Creature — Construct", mkmMP2SnufflerOracle, 4, 4)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, snuffler, game.CounterPlusOne); got != 0 {
		t.Fatalf("sacrificing a creature gave %d counters, want 0", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	passPriorityAroundTable(t, g)
	if got := countersOn(g, snuffler, game.CounterPlusOne); got != 1 {
		t.Errorf("sacrificing an artifact gave %d counters, want 1", got)
	}
}

// --- Massacre Girl, Known Killer -----------------------------------------------

func TestMkmMP2MassacreGirlGivesYourCreaturesWitherAndDrawsOnAZeroToughnessDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	girl := b12Push(g, me.ID, "Massacre Girl, Known Killer", "Legendary Creature — Human Assassin", mkmMP2MassacreOracle, 4, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	assertKeywords(t, g, girl, "menace", "wither")
	if !hasKeywordOnBattlefield(t, g, bear, "wither") {
		t.Fatal("my other creature has no wither")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "wither") {
		t.Fatal("the opponent's creature has wither")
	}

	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, theirs, 2) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("two -1/-1 counters did not kill the 2/2")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card for a creature that died with toughness 0", hand, me.Hand.Size())
	}
}

func TestMkmMP2MassacreGirlDrawsNothingForAnOrdinaryDeath(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Massacre Girl, Known Killer", "Legendary Creature — Human Assassin", mkmMP2MassacreOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d; a creature destroyed with toughness 2 draws nothing", hand, me.Hand.Size())
	}
}

// --- Meddling Youths -----------------------------------------------------------

func TestMkmMP2MeddlingYouthsInvestigatesForThreeAttackers(t *testing.T) {
	for _, attackers := range []int{2, 3} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		youths := b12Push(g, me.ID, "Meddling Youths", "Creature — Human Detective", mkmMP2YouthsOracle, 4, 5)
		assertKeywords(t, g, youths, "haste")
		var ids []uuid.UUID
		for i := 0; i < attackers; i++ {
			ids = append(ids, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		declareAttack(t, g, opp.ID, ids...)
		passPriorityAroundTable(t, g)
		want := 0
		if attackers >= 3 {
			want = 1
		}
		if got := len(mkmMP2Clues(g, me.ID)); got != want {
			t.Errorf("%d attackers: %d Clues, want %d", attackers, got, want)
		}
	}
}

// --- Melek, Reforged Researcher ------------------------------------------------

func TestMkmMP2MelekIsTwiceYourInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mkmMP2Grave(me, "Shock", "Instant")
	mkmMP2Grave(me, "Divination", "Sorcery")
	mkmMP2Grave(me, "Bear", "Creature — Bear")
	mkmMP2Grave(opp, "Their Shock", "Instant")
	melek := b12Push(g, me.ID, "Melek, Reforged Researcher", "Legendary Creature — Weird Detective", mkmMP2MelekOracle, 0, 0)
	g.RunStateChecksForTest()
	if p, tough := mkmMPPT(t, g, melek); p != 4 || tough != 4 {
		t.Errorf("Melek is %d/%d with two instant and sorcery cards, want 4/4", p, tough)
	}
}

func TestMkmMP2MelekDiscountsOnlyTheFirstInstantOrSorceryEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mkmMP2Grave(me, "Shock", "Instant")
	b12Push(g, me.ID, "Melek, Reforged Researcher", "Legendary Creature — Weird Detective", mkmMP2MelekOracle, 0, 0)
	if got := priceInHand(t, g, me, "Bolt", "Instant", "{3}{R}"); got != 1 {
		t.Fatalf("the turn's first instant costs %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("a creature spell costs %d, want 4", got)
	}
	castCatalogSpell(t, g, "First Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if got := priceInHand(t, g, me, "Bolt", "Instant", "{3}{R}"); got != 4 {
		t.Errorf("the turn's second instant or sorcery costs %d, want 4", got)
	}
}

// --- Mistway Spy ---------------------------------------------------------------

func TestMkmMP2MistwaySpyDisguiseIsWired(t *testing.T) {
	ac := game.AlternativeCostByKey(mkmMP2SpyOracle, "disguise")
	if ac == nil || ac.FaceDown == nil || ac.ManaCost != MorphCastCost || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != "{1}{U}" {
		t.Fatalf("Mistway Spy's disguise is %+v", ac)
	}
}

func TestMkmMP2MistwaySpyTurnedUpInvestigatesForEachCreatureThatHits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spy := mkmMPCastFaceDown(t, g, "Mistway Spy", mkmMP2SpyTypeLine, mkmMP2SpyOracle, 1, 1)
	mkmMPTurnFaceUp(t, g, me, spy, "U", "C")
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, spy, "flying")
	if got := len(mkmMP2Clues(g, me.ID)); got != 0 {
		t.Fatalf("%d Clues before any damage, want 0", got)
	}
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	attackWith(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 2 {
		t.Errorf("%d Clues after two creatures dealt combat damage, want 2", got)
	}
}

func TestMkmMP2MistwaySpyEnteringFaceUpDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castHoldCreature(t, g, "Mistway Spy", mkmMP2SpyTypeLine, mkmMP2SpyOracle, 1, 1)
	passPriorityAroundTable(t, g)
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 0 {
		t.Errorf("%d Clues without the Spy turning face up, want 0", got)
	}
}

// --- No Witnesses --------------------------------------------------------------

func TestMkmMP2NoWitnessesCluesTheMostCreaturesThenSweeps(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	var all []uuid.UUID
	for i := 0; i < 2; i++ {
		all = append(all, b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2))
		all = append(all, b12Creature(g, a.ID, "Their Bear", "Creature — Bear", 2, 2))
	}
	all = append(all, b12Creature(g, b.ID, "Lone Bear", "Creature — Bear", 2, 2))

	castCatalogSpell(t, g, "No Witnesses", "Sorcery", mkmMP2NoWitnessesOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range all {
		if g.Battlefield.Contains(id) {
			t.Error("a creature survived")
		}
	}
	for _, c := range []struct {
		p    *game.Player
		want int
	}{{me, 1}, {a, 1}, {b, 0}, {g.Seats[3], 0}} {
		if got := len(mkmMP2Clues(g, c.p.ID)); got != c.want {
			t.Errorf("%s has %d Clues, want %d", c.p.Name, got, c.want)
		}
	}
}

// --- Officious Interrogation ---------------------------------------------------

func TestMkmMP2OfficiousInterrogationCostsWUPerExtraTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	two := append(one, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if got := selfPriced(t, g, me, mkmMP2OfficiousOracle, "Instant", "{W}{U}", game.ZoneHand, one).ManaValue(); got != 2 {
		t.Errorf("one target costs %d, want 2", got)
	}
	if got := selfPriced(t, g, me, mkmMP2OfficiousOracle, "Instant", "{W}{U}", game.ZoneHand, two).ManaValue(); got != 4 {
		t.Errorf("two targets cost %d, want 4", got)
	}
}

func TestMkmMP2OfficiousInterrogationInvestigatesForTheTargetsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, a.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, a.ID, "Their Elk", "Creature — Elk", 3, 3)
	b12Creature(g, b.ID, "Untargeted Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Officious Interrogation", "Instant", mkmMP2OfficiousOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetPlayer, ID: a.ID},
	})
	passPriorityAroundTable(t, g)
	if got := len(mkmMP2Clues(g, me.ID)); got != 3 {
		t.Errorf("%d Clues, want 3 (one of mine and two of theirs)", got)
	}
}

// --- Out Cold ------------------------------------------------------------------

func TestMkmMP2OutColdTapsAndStunsTwoAndInvestigates(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if spec, _ := Lookup(mkmMP2OutColdOracle); !spec.CantBeCountered {
		t.Error("Out Cold can be countered")
	}
	a := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	idle := b12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Out Cold", "Instant", mkmMP2OutColdOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, _ := battlefieldCard(g, id)
		if !c.Tapped || c.Counters[game.CounterStun] != 1 {
			t.Errorf("%s: tapped %v, %d stun counters; want tapped with one", c.Name, c.Tapped, c.Counters[game.CounterStun])
		}
	}
	if c, _ := battlefieldCard(g, idle); c.Tapped || c.Counters[game.CounterStun] != 0 {
		t.Error("the untargeted creature was touched")
	}
	if got := len(mkmMP2Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Perimeter Enforcer --------------------------------------------------------

func TestMkmMP2PerimeterEnforcerGrowsForAnotherDetectiveOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	enforcer := castHoldCreature(t, g, "Perimeter Enforcer", "Creature — Human Detective", mkmMP2EnforcerOracle, 1, 1)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, enforcer, "flying", "lifelink")
	if p, tough := mkmMPPT(t, g, enforcer); p != 1 || tough != 1 {
		t.Fatalf("its own entry pumped it to %d/%d", p, tough)
	}
	castHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if p, _ := mkmMPPT(t, g, enforcer); p != 1 {
		t.Fatalf("a non-Detective entering pumped it to power %d", p)
	}
	// The Inspector's own trigger and the Enforcer's: order them, decline
	// the loot.
	castHoldCreature(t, g, "Projektor Inspector", "Creature — Human Detective", mkmMP2ProjektorOracle, 3, 2)
	b30SettleOrder(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		answerMayChoice(t, g, me.ID, false)
	}
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, enforcer); p != 2 || tough != 2 {
		t.Errorf("another Detective entered and it is %d/%d, want 2/2", p, tough)
	}
}

func TestMkmMP2PerimeterEnforcerGrowsWhenADetectiveTurnsFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	enforcer := b12Push(g, me.ID, "Perimeter Enforcer", "Creature — Human Detective", mkmMP2EnforcerOracle, 1, 1)
	spy := mkmMPCastFaceDown(t, g, "Mistway Spy", mkmMP2SpyTypeLine, mkmMP2SpyOracle, 1, 1)
	if p, _ := mkmMPPT(t, g, enforcer); p != 1 {
		t.Fatalf("a face-down creature entering pumped it to power %d", p)
	}
	// The Spy's own face-up trigger and the Enforcer's: order them.
	mkmMPTurnFaceUp(t, g, me, spy, "U", "C")
	b30SettleOrder(t, g)
	if p, tough := mkmMPPT(t, g, enforcer); p != 2 || tough != 2 {
		t.Errorf("a Detective turned face up and it is %d/%d, want 2/2", p, tough)
	}
}

// --- Persuasive Interrogators --------------------------------------------------

func TestMkmMP2PersuasiveInterrogatorsPoisonsForEachClueSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castHoldCreature(t, g, "Persuasive Interrogators", "Creature — Gorgon Detective", mkmMP2PersuasiveOracle, 5, 6)
	passPriorityAroundTable(t, g)
	clues := mkmMP2Clues(g, me.ID)
	if len(clues) != 1 {
		t.Fatalf("%d Clues on entering, want 1", len(clues))
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(clues[0]) })
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
			t.Fatal("I was accepted as the target opponent")
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("the opponent has %d poison counters, want 2", got)
	}
}

// --- Pompous Gadabout ----------------------------------------------------------

func TestMkmMP2PompousGadaboutHasHexproofOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Push(g, me.ID, "Pompous Gadabout", "Creature — Human Citizen", mkmMP2GadaboutOracle, 4, 2)
	theirs := b12Push(g, opp.ID, "Pompous Gadabout", "Creature — Human Citizen", mkmMP2GadaboutOracle, 4, 2)
	g.RunStateChecksForTest()
	if !hasKeywordOnBattlefield(t, g, mine, "hexproof") {
		t.Error("the active player's Gadabout has no hexproof")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "hexproof") {
		t.Error("the Gadabout has hexproof on its controller's opponent's turn")
	}
}

func TestMkmMP2PompousGadaboutCantBeBlockedByAFaceDownCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gadabout := b12Push(g, me.ID, "Pompous Gadabout", "Creature — Human Citizen", mkmMP2GadaboutOracle, 4, 2)
	hidden := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hidden Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		FaceDown: true, FaceDownKind: game.FaceDownDisguised, Owner: opp.ID, Controller: opp.ID,
	})
	named := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(gadabout, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(hidden, gadabout); err == nil {
		t.Error("a face-down creature, which has no name, blocked the Gadabout")
	}
	if err := g.DeclareBlocker(named, gadabout); err != nil {
		t.Errorf("a named creature could not block the Gadabout: %v", err)
	}
}

// --- Private Eye ---------------------------------------------------------------

func TestMkmMP2PrivateEyePumpsOtherDetectives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eye := b12Push(g, me.ID, "Private Eye", "Creature — Homunculus Detective", mkmMP2PrivateEyeOracle, 3, 3)
	det := b12Creature(g, me.ID, "My Detective", "Creature — Human Detective", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Detective", "Creature — Human Detective", 2, 2)
	g.RunStateChecksForTest()
	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{eye, 3}, {det, 3}, {bear, 2}, {theirs, 2}} {
		if p, _ := mkmMPPT(t, g, c.id); p != c.want {
			t.Errorf("%s has power %d, want %d", tmtB1Card(t, g, c.id).Name, p, c.want)
		}
	}
}

func TestMkmMP2PrivateEyeMakesADetectiveUnblockableOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Private Eye", "Creature — Homunculus Detective", mkmMP2PrivateEyeOracle, 3, 3)
	det := b12Creature(g, me.ID, "My Detective", "Creature — Human Detective", 2, 2)
	blocker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	// The turn's draw step was the first card; this is the second.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, det)
	passPriorityAroundTable(t, g)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(det, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, det); err == nil {
		t.Error("the Detective was blocked")
	}
}

// --- Projektor Inspector -------------------------------------------------------

func TestMkmMP2ProjektorInspectorLootsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 4)
	castHoldCreature(t, g, "Projektor Inspector", "Creature — Human Detective", mkmMP2ProjektorOracle, 3, 2)
	hand := me.Hand.Size()
	answerMayChoice(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

func TestMkmMP2ProjektorInspectorAsksForAnotherDetectiveButNotABear(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Projektor Inspector", "Creature — Human Detective", mkmMP2ProjektorOracle, 3, 2)
	castHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("a non-Detective entering asked to loot")
	}
	castHoldCreature(t, g, "Perimeter Enforcer", "Creature — Human Detective", mkmMP2EnforcerOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) == nil {
		t.Error("another Detective entering asked nothing")
	}
}

// --- Public Thoroughfare -------------------------------------------------------

func TestMkmMP2PublicThoroughfareStaysWhenYouTapALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	land := playLandFromHand(t, g, "Public Thoroughfare", mkmMP2ThoroughfareOracl)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, land); !ok || !c.Tapped {
		t.Fatal("the Thoroughfare did not enter tapped")
	}
	choice := payUnlessFor(g, me.ID)
	if choice == nil || choice.PayAction() == nil || choice.PayAction().Kind != game.PayActionTap {
		t.Fatalf("no tap payment asked: %+v", choice)
	}
	var opts []uuid.UUID
	g.ReadSnapshot(func() { opts = g.PayActionOptionsForEffect(me.ID, choice.PayAction()) })
	if len(opts) != 1 || opts[0] != forest {
		t.Fatalf("the payment offers %v, want only the untapped Forest", opts)
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, true, []uuid.UUID{forest}); err != nil {
		t.Fatalf("tap the Forest: %v", err)
	}
	if c, _ := battlefieldCard(g, forest); !c.Tapped {
		t.Error("the Forest was not tapped")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the Thoroughfare was sacrificed although the payment was made")
	}
}

func TestMkmMP2PublicThoroughfareIsSacrificedWhenYouDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Permanent(g, me.ID, "Rock", "Artifact")
	land := playLandFromHand(t, g, "Public Thoroughfare", mkmMP2ThoroughfareOracl)
	passPriorityAroundTable(t, g)
	choice := payUnlessFor(g, me.ID)
	if choice == nil {
		t.Fatal("no payment asked")
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, false, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the Thoroughfare was not sacrificed")
	}
}
