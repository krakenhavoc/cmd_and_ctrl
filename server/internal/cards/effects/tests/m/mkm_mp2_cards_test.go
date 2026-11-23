package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMkmMP2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.MkmMP2MacabreOracle:     "Macabre Reconstruction",
		effectstest.MkmMP2SnufflerOracle:    "Magnetic Snuffler",
		effectstest.MkmMP2MassacreOracle:    "Massacre Girl, Known Killer",
		effectstest.MkmMP2YouthsOracle:      "Meddling Youths",
		effectstest.MkmMP2MelekOracle:       "Melek, Reforged Researcher",
		effectstest.MkmMP2SpyOracle:         "Mistway Spy",
		effectstest.MkmMP2NoWitnessesOracle: "No Witnesses",
		effectstest.MkmMP2OfficiousOracle:   "Officious Interrogation",
		effectstest.MkmMP2OutColdOracle:     "Out Cold",
		effectstest.MkmMP2EnforcerOracle:    "Perimeter Enforcer",
		effectstest.MkmMP2PersuasiveOracle:  "Persuasive Interrogators",
		effectstest.MkmMP2GadaboutOracle:    "Pompous Gadabout",
		effectstest.MkmMP2PrivateEyeOracle:  "Private Eye",
		effectstest.MkmMP2ProjektorOracle:   "Projektor Inspector",
		effectstest.MkmMP2ThoroughfareOracl: "Public Thoroughfare",
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

// --- Macabre Reconstruction ----------------------------------------------------

func TestMkmMP2MacabreReconstructionCostsTwoLessAfterACreatureCardHitYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.MkmMP2MacabreOracle, "Sorcery", "{3}{B}"); got != 4 {
		t.Fatalf("with nothing in the graveyard this turn it costs %d, want 4", got)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.MkmMP2MacabreOracle, "Sorcery", "{3}{B}"); got != 4 {
		t.Errorf("an opponent's creature dying made it cost %d, want 4", got)
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.MkmMP2MacabreOracle, "Sorcery", "{3}{B}"); got != 2 {
		t.Errorf("after my creature card died it costs %d, want 2", got)
	}
}

func TestMkmMP2MacabreReconstructionIgnoresANoncreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.MkmMP2MacabreOracle, "Sorcery", "{3}{B}"); got != 4 {
		t.Errorf("an artifact card reaching the graveyard made it cost %d, want 4", got)
	}
}

func TestMkmMP2MacabreReconstructionReturnsTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.MkmMP2Grave(me, "Bear", "Creature — Bear")
	b := effectstest.MkmMP2Grave(me, "Elk", "Creature — Elk")
	effectstest.CastCatalogSpell(t, g, "Macabre Reconstruction", "Sorcery", effectstest.MkmMP2MacabreOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two creature cards did not reach the hand")
	}
}

// --- Magnetic Snuffler ---------------------------------------------------------

func TestMkmMP2MagneticSnufflerReturnsAnEquipmentAttachedToItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sword := effectstest.MkmMP2Grave(me, "Test Sword", "Artifact — Equipment")
	snuffler := effectstest.CastHoldCreature(t, g, "Magnetic Snuffler", "Artifact Creature — Construct", effectstest.MkmMP2SnufflerOracle, 4, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, sword)
	if !ok {
		t.Fatal("the Equipment did not return to the battlefield")
	}
	if c.AttachedTo.ID != snuffler || c.Controller != me.ID {
		t.Errorf("the Equipment is attached to %v under %v, want the Snuffler under me", c.AttachedTo.ID, c.Controller)
	}
}

func TestMkmMP2MagneticSnufflerGrowsWhenYouSacrificeAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	snuffler := effectstest.B12Push(g, me.ID, "Magnetic Snuffler", "Artifact Creature — Construct", effectstest.MkmMP2SnufflerOracle, 4, 4)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, snuffler, game.CounterPlusOne); got != 0 {
		t.Fatalf("sacrificing a creature gave %d counters, want 0", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, snuffler, game.CounterPlusOne); got != 1 {
		t.Errorf("sacrificing an artifact gave %d counters, want 1", got)
	}
}

// --- Massacre Girl, Known Killer -----------------------------------------------

func TestMkmMP2MassacreGirlGivesYourCreaturesWitherAndDrawsOnAZeroToughnessDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	girl := effectstest.B12Push(g, me.ID, "Massacre Girl, Known Killer", "Legendary Creature — Human Assassin", effectstest.MkmMP2MassacreOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AssertKeywords(t, g, girl, "menace", "wither")
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "wither") {
		t.Fatal("my other creature has no wither")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "wither") {
		t.Fatal("the opponent's creature has wither")
	}

	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(bear, theirs, 2) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("two -1/-1 counters did not kill the 2/2")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card for a creature that died with toughness 0", hand, me.Hand.Size())
	}
}

func TestMkmMP2MassacreGirlDrawsNothingForAnOrdinaryDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Massacre Girl, Known Killer", "Legendary Creature — Human Assassin", effectstest.MkmMP2MassacreOracle, 4, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d; a creature destroyed with toughness 2 draws nothing", hand, me.Hand.Size())
	}
}

// --- Meddling Youths -----------------------------------------------------------

func TestMkmMP2MeddlingYouthsInvestigatesForThreeAttackers(t *testing.T) {
	for _, attackers := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		youths := effectstest.B12Push(g, me.ID, "Meddling Youths", "Creature — Human Detective", effectstest.MkmMP2YouthsOracle, 4, 5)
		effectstest.AssertKeywords(t, g, youths, "haste")
		var ids []uuid.UUID
		for i := 0; i < attackers; i++ {
			ids = append(ids, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		effectstest.DeclareAttack(t, g, opp.ID, ids...)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if attackers >= 3 {
			want = 1
		}
		if got := len(effectstest.MkmMP2Clues(g, me.ID)); got != want {
			t.Errorf("%d attackers: %d Clues, want %d", attackers, got, want)
		}
	}
}

// --- Melek, Reforged Researcher ------------------------------------------------

func TestMkmMP2MelekIsTwiceYourInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.MkmMP2Grave(me, "Shock", "Instant")
	effectstest.MkmMP2Grave(me, "Divination", "Sorcery")
	effectstest.MkmMP2Grave(me, "Bear", "Creature — Bear")
	effectstest.MkmMP2Grave(opp, "Their Shock", "Instant")
	melek := effectstest.B12Push(g, me.ID, "Melek, Reforged Researcher", "Legendary Creature — Weird Detective", effectstest.MkmMP2MelekOracle, 0, 0)
	g.RunStateChecksForTest()
	if p, tough := effectstest.MkmMPPT(t, g, melek); p != 4 || tough != 4 {
		t.Errorf("Melek is %d/%d with two instant and sorcery cards, want 4/4", p, tough)
	}
}

func TestMkmMP2MelekDiscountsOnlyTheFirstInstantOrSorceryEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.MkmMP2Grave(me, "Shock", "Instant")
	effectstest.B12Push(g, me.ID, "Melek, Reforged Researcher", "Legendary Creature — Weird Detective", effectstest.MkmMP2MelekOracle, 0, 0)
	if got := effectstest.PriceInHand(t, g, me, "Bolt", "Instant", "{3}{R}"); got != 1 {
		t.Fatalf("the turn's first instant costs %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{3}{G}"); got != 4 {
		t.Errorf("a creature spell costs %d, want 4", got)
	}
	effectstest.CastCatalogSpell(t, g, "First Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PriceInHand(t, g, me, "Bolt", "Instant", "{3}{R}"); got != 4 {
		t.Errorf("the turn's second instant or sorcery costs %d, want 4", got)
	}
}

// --- Mistway Spy ---------------------------------------------------------------

func TestMkmMP2MistwaySpyDisguiseIsWired(t *testing.T) {
	ac := game.AlternativeCostByKey(effectstest.MkmMP2SpyOracle, "disguise")
	if ac == nil || ac.FaceDown == nil || ac.ManaCost != effects.MorphCastCost || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != "{1}{U}" {
		t.Fatalf("Mistway Spy's disguise is %+v", ac)
	}
}

func TestMkmMP2MistwaySpyTurnedUpInvestigatesForEachCreatureThatHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spy := effectstest.MkmMPCastFaceDown(t, g, "Mistway Spy", effectstest.MkmMP2SpyTypeLine, effectstest.MkmMP2SpyOracle, 1, 1)
	effectstest.MkmMPTurnFaceUp(t, g, me, spy, "U", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, spy, "flying")
	if got := len(effectstest.MkmMP2Clues(g, me.ID)); got != 0 {
		t.Fatalf("%d Clues before any damage, want 0", got)
	}
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	effectstest.AttackWith(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.MkmMP2Clues(g, me.ID)); got != 2 {
		t.Errorf("%d Clues after two creatures dealt combat damage, want 2", got)
	}
}

func TestMkmMP2MistwaySpyEnteringFaceUpDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastHoldCreature(t, g, "Mistway Spy", effectstest.MkmMP2SpyTypeLine, effectstest.MkmMP2SpyOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.MkmMP2Clues(g, me.ID)); got != 0 {
		t.Errorf("%d Clues without the Spy turning face up, want 0", got)
	}
}

// --- No Witnesses --------------------------------------------------------------

func TestMkmMP2NoWitnessesCluesTheMostCreaturesThenSweeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	var all []uuid.UUID
	for i := 0; i < 2; i++ {
		all = append(all, effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2))
		all = append(all, effectstest.B12Creature(g, a.ID, "Their Bear", "Creature — Bear", 2, 2))
	}
	all = append(all, effectstest.B12Creature(g, b.ID, "Lone Bear", "Creature — Bear", 2, 2))

	effectstest.CastCatalogSpell(t, g, "No Witnesses", "Sorcery", effectstest.MkmMP2NoWitnessesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range all {
		if g.Battlefield.Contains(id) {
			t.Error("a creature survived")
		}
	}
	for _, c := range []struct {
		p    *game.Player
		want int
	}{{me, 1}, {a, 1}, {b, 0}, {g.Seats[3], 0}} {
		if got := len(effectstest.MkmMP2Clues(g, c.p.ID)); got != c.want {
			t.Errorf("%s has %d Clues, want %d", c.p.Name, got, c.want)
		}
	}
}

// --- Officious Interrogation ---------------------------------------------------

func TestMkmMP2OfficiousInterrogationCostsWUPerExtraTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	two := append(one, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if got := effectstest.SelfPriced(t, g, me, effectstest.MkmMP2OfficiousOracle, "Instant", "{W}{U}", game.ZoneHand, one).ManaValue(); got != 2 {
		t.Errorf("one target costs %d, want 2", got)
	}
	if got := effectstest.SelfPriced(t, g, me, effectstest.MkmMP2OfficiousOracle, "Instant", "{W}{U}", game.ZoneHand, two).ManaValue(); got != 4 {
		t.Errorf("two targets cost %d, want 4", got)
	}
}

func TestMkmMP2OfficiousInterrogationInvestigatesForTheTargetsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, a.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, a.ID, "Their Elk", "Creature — Elk", 3, 3)
	effectstest.B12Creature(g, b.ID, "Untargeted Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Officious Interrogation", "Instant", effectstest.MkmMP2OfficiousOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetPlayer, ID: a.ID},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.MkmMP2Clues(g, me.ID)); got != 3 {
		t.Errorf("%d Clues, want 3 (one of mine and two of theirs)", got)
	}
}

// --- Out Cold ------------------------------------------------------------------

func TestMkmMP2OutColdTapsAndStunsTwoAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if spec, _ := effects.Lookup(effectstest.MkmMP2OutColdOracle); !spec.CantBeCountered {
		t.Error("Out Cold can be countered")
	}
	a := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Out Cold", "Instant", effectstest.MkmMP2OutColdOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c, _ := effectstest.BattlefieldCard(g, id)
		if !c.Tapped || c.Counters[game.CounterStun] != 1 {
			t.Errorf("%s: tapped %v, %d stun counters; want tapped with one", c.Name, c.Tapped, c.Counters[game.CounterStun])
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, idle); c.Tapped || c.Counters[game.CounterStun] != 0 {
		t.Error("the untargeted creature was touched")
	}
	if got := len(effectstest.MkmMP2Clues(g, me.ID)); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Perimeter Enforcer --------------------------------------------------------

func TestMkmMP2PerimeterEnforcerGrowsForAnotherDetectiveOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	enforcer := effectstest.CastHoldCreature(t, g, "Perimeter Enforcer", "Creature — Human Detective", effectstest.MkmMP2EnforcerOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, enforcer, "flying", "lifelink")
	if p, tough := effectstest.MkmMPPT(t, g, enforcer); p != 1 || tough != 1 {
		t.Fatalf("its own entry pumped it to %d/%d", p, tough)
	}
	effectstest.CastHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.MkmMPPT(t, g, enforcer); p != 1 {
		t.Fatalf("a non-Detective entering pumped it to power %d", p)
	}
	// The Inspector's own trigger and the Enforcer's: order them, decline
	// the loot.
	effectstest.CastHoldCreature(t, g, "Projektor Inspector", "Creature — Human Detective", effectstest.MkmMP2ProjektorOracle, 3, 2)
	effectstest.B30SettleOrder(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		effectstest.AnswerMayChoice(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, enforcer); p != 2 || tough != 2 {
		t.Errorf("another Detective entered and it is %d/%d, want 2/2", p, tough)
	}
}

func TestMkmMP2PerimeterEnforcerGrowsWhenADetectiveTurnsFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	enforcer := effectstest.B12Push(g, me.ID, "Perimeter Enforcer", "Creature — Human Detective", effectstest.MkmMP2EnforcerOracle, 1, 1)
	spy := effectstest.MkmMPCastFaceDown(t, g, "Mistway Spy", effectstest.MkmMP2SpyTypeLine, effectstest.MkmMP2SpyOracle, 1, 1)
	if p, _ := effectstest.MkmMPPT(t, g, enforcer); p != 1 {
		t.Fatalf("a face-down creature entering pumped it to power %d", p)
	}
	// The Spy's own face-up trigger and the Enforcer's: order them.
	effectstest.MkmMPTurnFaceUp(t, g, me, spy, "U", "C")
	effectstest.B30SettleOrder(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, enforcer); p != 2 || tough != 2 {
		t.Errorf("a Detective turned face up and it is %d/%d, want 2/2", p, tough)
	}
}

// --- Persuasive Interrogators --------------------------------------------------

func TestMkmMP2PersuasiveInterrogatorsPoisonsForEachClueSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastHoldCreature(t, g, "Persuasive Interrogators", "Creature — Gorgon Detective", effectstest.MkmMP2PersuasiveOracle, 5, 6)
	effectstest.PassPriorityAroundTable(t, g)
	clues := effectstest.MkmMP2Clues(g, me.ID)
	if len(clues) != 1 {
		t.Fatalf("%d Clues on entering, want 1", len(clues))
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(clues[0]) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
			t.Fatal("I was accepted as the target opponent")
		}
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("the opponent has %d poison counters, want 2", got)
	}
}

// --- Pompous Gadabout ----------------------------------------------------------

func TestMkmMP2PompousGadaboutHasHexproofOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Pompous Gadabout", "Creature — Human Citizen", effectstest.MkmMP2GadaboutOracle, 4, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Pompous Gadabout", "Creature — Human Citizen", effectstest.MkmMP2GadaboutOracle, 4, 2)
	g.RunStateChecksForTest()
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "hexproof") {
		t.Error("the active player's Gadabout has no hexproof")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "hexproof") {
		t.Error("the Gadabout has hexproof on its controller's opponent's turn")
	}
}

func TestMkmMP2PompousGadaboutCantBeBlockedByAFaceDownCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gadabout := effectstest.B12Push(g, me.ID, "Pompous Gadabout", "Creature — Human Citizen", effectstest.MkmMP2GadaboutOracle, 4, 2)
	hidden := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hidden Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		FaceDown: true, FaceDownKind: game.FaceDownDisguised, Owner: opp.ID, Controller: opp.ID,
	})
	named := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(gadabout, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(hidden, gadabout); err == nil {
		t.Error("a face-down creature, which has no name, blocked the Gadabout")
	}
	if err := g.DeclareBlocker(named, gadabout); err != nil {
		t.Errorf("a named creature could not block the Gadabout: %v", err)
	}
}

// --- Private Eye ---------------------------------------------------------------

func TestMkmMP2PrivateEyePumpsOtherDetectives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eye := effectstest.B12Push(g, me.ID, "Private Eye", "Creature — Homunculus Detective", effectstest.MkmMP2PrivateEyeOracle, 3, 3)
	det := effectstest.B12Creature(g, me.ID, "My Detective", "Creature — Human Detective", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Detective", "Creature — Human Detective", 2, 2)
	g.RunStateChecksForTest()
	for _, c := range []struct {
		id   uuid.UUID
		want int
	}{{eye, 3}, {det, 3}, {bear, 2}, {theirs, 2}} {
		if p, _ := effectstest.MkmMPPT(t, g, c.id); p != c.want {
			t.Errorf("%s has power %d, want %d", effectstest.TmtB1Card(t, g, c.id).Name, p, c.want)
		}
	}
}

func TestMkmMP2PrivateEyeMakesADetectiveUnblockableOnTheSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Private Eye", "Creature — Homunculus Detective", effectstest.MkmMP2PrivateEyeOracle, 3, 3)
	det := effectstest.B12Creature(g, me.ID, "My Detective", "Creature — Human Detective", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// The turn's draw step was the first card; this is the second.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, det)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(det, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, det); err == nil {
		t.Error("the Detective was blocked")
	}
}

// --- Projektor Inspector -------------------------------------------------------

func TestMkmMP2ProjektorInspectorLootsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 4)
	effectstest.CastHoldCreature(t, g, "Projektor Inspector", "Creature — Human Detective", effectstest.MkmMP2ProjektorOracle, 3, 2)
	hand := me.Hand.Size()
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

func TestMkmMP2ProjektorInspectorAsksForAnotherDetectiveButNotABear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Projektor Inspector", "Creature — Human Detective", effectstest.MkmMP2ProjektorOracle, 3, 2)
	effectstest.CastHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("a non-Detective entering asked to loot")
	}
	effectstest.CastHoldCreature(t, g, "Perimeter Enforcer", "Creature — Human Detective", effectstest.MkmMP2EnforcerOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) == nil {
		t.Error("another Detective entering asked nothing")
	}
}

// --- Public Thoroughfare -------------------------------------------------------

func TestMkmMP2PublicThoroughfareStaysWhenYouTapALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	land := effectstest.PlayLandFromHand(t, g, "Public Thoroughfare", effectstest.MkmMP2ThoroughfareOracl)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, land); !ok || !c.Tapped {
		t.Fatal("the Thoroughfare did not enter tapped")
	}
	choice := effectstest.PayUnlessFor(g, me.ID)
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
	if c, _ := effectstest.BattlefieldCard(g, forest); !c.Tapped {
		t.Error("the Forest was not tapped")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the Thoroughfare was sacrificed although the payment was made")
	}
}

func TestMkmMP2PublicThoroughfareIsSacrificedWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	land := effectstest.PlayLandFromHand(t, g, "Public Thoroughfare", effectstest.MkmMP2ThoroughfareOracl)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.PayUnlessFor(g, me.ID)
	if choice == nil {
		t.Fatal("no payment asked")
	}
	if err := g.ResolvePayUnlessWithCards(choice.ID, me.ID, false, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the Thoroughfare was not sacrificed")
	}
}
