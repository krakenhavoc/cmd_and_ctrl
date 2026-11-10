package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// followup_tdm_lci_dsk_cards_test.go — a follow-up batch for Tarkir:
// Dragonstorm (#4463), The Lost Caverns of Ixalan (#4470) and Duskmourn
// (#4466). Each test plays the real catalog entry and asserts what its
// text says.

const (
	fuKheruOracle      = "aef4ab0e-c81f-428c-ab23-64504a9fb989"
	fuSkimmerOracle    = "a3275d80-112b-4839-a014-e44ef48d2e6b"
	fuLasydOracle      = "f780d6f6-540b-4773-8a48-e56e95c2d39e"
	fuCaveInOracle     = "558135b3-babc-40b6-9d9d-57ab0965ee3d"
	fuCaptivatingCave  = "4c77767a-8133-43bc-b7a5-09a73259d354"
	fuCavernousMaw     = "952ab8fe-f7d3-4673-89de-8c6d3f8a081f"
	fuCorpsesOracle    = "78bac92f-333f-4c43-bccf-03c3772ee99b"
	fuUnableOracle     = "3d240726-5214-4493-8537-29bf65db1e4b"
	fuVeteranSurvivor  = "cc5b1276-d13e-4392-988d-1315bc90a9b1"
	fuCaveLandTypeLine = "Land — Cave"
)

func TestFollowupTdmLciDskEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fuKheruOracle:     "Kheru Goldkeeper",
		fuSkimmerOracle:   "Kishla Skimmer",
		fuLasydOracle:     "Lasyd Prowler",
		fuCaveInOracle:    "Calamitous Cave-In",
		fuCaptivatingCave: "Captivating Cave",
		fuCavernousMaw:    "Cavernous Maw",
		fuCorpsesOracle:   "Corpses of the Lost",
		fuUnableOracle:    "Unable to Scream",
		fuVeteranSurvivor: "Veteran Survivor",
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

// --- fixtures ------------------------------------------------------------------

// fuLand puts an untapped land with the given type line onto the
// battlefield under owner's control.
func fuLand(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
}

// fuSkeletonPirate is the Skeleton Pirate token `owner` controls, or
// uuid.Nil.
func fuSkeletonPirate(g *game.Game, owner uuid.UUID) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Skeleton Pirate" && IsToken(c) {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

// --- Kheru Goldkeeper ----------------------------------------------------------

func TestFollowupKheruGoldkeeperMakesATreasureWhenCardsLeaveOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kheru := b12Push(g, me.ID, "Kheru Goldkeeper", "Creature — Dragon", fuKheruOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, kheru, "flying") {
		t.Error("Kheru Goldkeeper has no flying")
	}
	a := pushGraveyardCard(g, me, "First")
	b := pushGraveyardCard(g, me, "Second")
	advanceToMainOf(t, g, 0)
	tdmExileFromGraveyard(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID); got != 1 {
		t.Fatalf("%d Treasures after a card left on your turn, want 1", got)
	}
	advanceToMainOf(t, g, 1)
	tdmExileFromGraveyard(t, g, opp.ID, b)
	passPriorityAroundTable(t, g)
	if got := countTreasuresControlledBy(g, me.ID); got != 1 {
		t.Errorf("%d Treasures after a card left on an opponent's turn, want still 1", got)
	}
}

func TestFollowupKheruGoldkeeperRenewsTwoCountersAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushBattlefieldCreature(g, me.ID)
	id, _ := activateFromGraveyard(t, g, "Kheru Goldkeeper", "Creature — Dragon",
		fuKheruOracle, 3, 3, "{B}{G}{U}{C}{C}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile Kheru Goldkeeper")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, target, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if got := countersOn(g, target, game.CounterFlying); got != 1 {
		t.Errorf("%d flying counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, target, "flying") {
		t.Error("the target has no flying from its flying counter")
	}
}

// --- Kishla Skimmer ------------------------------------------------------------

func TestFollowupKishlaSkimmerDrawsOnceEachOfYourTurns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skimmer := b12Push(g, me.ID, "Kishla Skimmer", "Creature — Bird Scout", fuSkimmerOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, skimmer, "flying") {
		t.Error("Kishla Skimmer has no flying")
	}
	cards := []uuid.UUID{
		pushGraveyardCard(g, me, "First"), pushGraveyardCard(g, me, "Second"),
		pushGraveyardCard(g, me, "Third"), pushGraveyardCard(g, me, "Fourth"),
	}
	advanceToMainOf(t, g, 0)
	hand := me.Hand.Size()
	tdmExileFromGraveyard(t, g, me.ID, cards[0])
	passPriorityAroundTable(t, g)
	tdmExileFromGraveyard(t, g, me.ID, cards[1])
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("drew %d after two cards left on one turn, want 1", got)
	}

	advanceToMainOf(t, g, 1)
	hand = me.Hand.Size()
	tdmExileFromGraveyard(t, g, opp.ID, cards[2])
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d after a card left on an opponent's turn, want 0", got)
	}

	advanceToMainOf(t, g, 0)
	hand = me.Hand.Size()
	tdmExileFromGraveyard(t, g, me.ID, cards[3])
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on your next turn, want 1 (once each turn)", got)
	}
}

// --- Lasyd Prowler -------------------------------------------------------------

func TestFollowupLasydProwlerMayMillOnePerLand(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		for i := 0; i < 3; i++ {
			fuLand(g, me.ID, "Forest", "Basic Land — Forest")
		}
		var lands int
		g.ReadSnapshot(func() { lands = len(landsControlledByPlayer(g, me.ID)) })
		library := me.Library.Size()
		castHoldCreature(t, g, "Lasyd Prowler", "Creature — Snake Ranger", fuLasydOracle, 5, 5)
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		want := 0
		if accept {
			want = lands
		}
		if got := library - me.Library.Size(); got != want {
			t.Errorf("accept=%v: milled %d cards with %d lands, want %d", accept, got, lands, want)
		}
	}
}

func TestFollowupLasydProwlerRenewsOneCounterPerLandCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushBattlefieldCreature(g, me.ID)
	pushGraveyardCardTyped(me, "Old Forest", "Basic Land — Forest")
	pushGraveyardCardTyped(me, "Old Cave", fuCaveLandTypeLine)
	pushGraveyardCardTyped(me, "Old Sorcery", "Sorcery")
	id, _ := activateFromGraveyard(t, g, "Lasyd Prowler", "Creature — Snake Ranger",
		fuLasydOracle, 5, 5, "{G}{G}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile Lasyd Prowler")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, target, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (two land cards in the graveyard)", got)
	}
}

// --- Calamitous Cave-In --------------------------------------------------------

func TestFollowupCalamitousCaveInDealsOnePerCave(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fuLand(g, me.ID, "Cave One", fuCaveLandTypeLine)
	fuLand(g, me.ID, "Cave Two", fuCaveLandTypeLine)
	fuLand(g, opp.ID, "Their Cave", fuCaveLandTypeLine)
	pushGraveyardCardTyped(me, "Old Cave", fuCaveLandTypeLine)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	castCatalogSpell(t, g, "Calamitous Cave-In", "Sorcery", fuCaveInOracle, nil)
	passPriorityAroundTable(t, g)
	// X is 3: two Caves you control and one Cave card in your graveyard;
	// an opponent's Cave does not count.
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(small) {
		t.Error("a creature with toughness 3 or less survived 3 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 4/4 died; X should be 3")
	}
}

func TestFollowupCalamitousCaveInWithNoCavesDealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	castCatalogSpell(t, g, "Calamitous Cave-In", "Sorcery", fuCaveInOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a 1/1 died with no Caves anywhere")
	}
}

// --- Captivating Cave ----------------------------------------------------------

func TestFollowupCaptivatingCaveFiltersAndSacrificesForTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cave := pushCatalogPermanent(g, me.ID, "Captivating Cave", fuCaveLandTypeLine, fuCaptivatingCave, false)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "R", 1)
	if err := g.ActivateManaAbility(me.ID, cave, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("{1}, {T}: add {G}: %v", err)
	}
	if got := mcPoolCount(me, "G"); got != 1 {
		t.Errorf("green in the pool %d, want 1", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	cave2 := pushCatalogPermanent(g2, me2.ID, "Captivating Cave", fuCaveLandTypeLine, fuCaptivatingCave, false)
	target := pushBattlefieldCreature(g2, me2.ID)
	advanceToMain(t, g2)
	if err := g2.AddManaForEffect(me2.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g2.ActivateCatalogAbility(me2.ID, cave2, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("{4}, {T}, Sacrifice: %v", err)
	}
	if !me2.Graveyard.Contains(cave2) {
		t.Fatal("the Cave was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g2)
	if got := countersOn(g2, target, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
}

func TestFollowupCaptivatingCaveSacrificeIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cave := pushCatalogPermanent(g, me.ID, "Captivating Cave", fuCaveLandTypeLine, fuCaptivatingCave, false)
	target := pushBattlefieldCreature(g, me.ID)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, cave, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Error("the sacrifice ability was activated in combat")
	}
}

// --- Cavernous Maw -------------------------------------------------------------

func TestFollowupCavernousMawAnimatesOnlyWithThreeOtherCaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	maw := pushCatalogPermanent(g, me.ID, "Cavernous Maw", fuCaveLandTypeLine, fuCavernousMaw, false)
	fuLand(g, me.ID, "Cave One", fuCaveLandTypeLine)
	fuLand(g, me.ID, "Cave Two", fuCaveLandTypeLine)
	advanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	// Two other Caves and the Maw itself: the Maw is not counted.
	if err := g.ActivateCatalogAbility(me.ID, maw, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the Maw animated with only two other Caves")
	}
	pushGraveyardCardTyped(me, "Old Cave", fuCaveLandTypeLine)
	if err := g.ActivateCatalogAbility(me.ID, maw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("animate with two other Caves and a Cave card in the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	ch := effectiveOf(t, g, maw)
	if ch.Power != 3 || ch.Toughness != 3 {
		t.Errorf("the Maw is %d/%d, want 3/3", ch.Power, ch.Toughness)
	}
	for _, typ := range []string{"Creature", "Land"} {
		if !slices.Contains(ch.Types, typ) {
			t.Errorf("the animated Maw is not a %s (types %v)", typ, ch.Types)
		}
	}
	for _, sub := range []string{"Elemental", "Cave"} {
		if !slices.Contains(ch.Subtypes, sub) {
			t.Errorf("the animated Maw is not a %s (subtypes %v)", sub, ch.Subtypes)
		}
	}
	if len(ch.Colors) != 0 {
		t.Errorf("the animated Maw has colours %v, want colourless", ch.Colors)
	}
}

// --- Corpses of the Lost -------------------------------------------------------

func TestFollowupCorpsesOfTheLostMakesAPumpedHastySkeleton(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Skeleton", "Creature — Skeleton", 1, 1)
	castCatalogSpell(t, g, "Corpses of the Lost", "Enchantment", fuCorpsesOracle, nil)
	passPriorityAroundTable(t, g)
	pirate := fuSkeletonPirate(g, me.ID)
	if pirate == uuid.Nil {
		t.Fatal("no Skeleton Pirate token was created")
	}
	ch := effectiveOf(t, g, pirate)
	if ch.Power != 3 || ch.Toughness != 2 {
		t.Errorf("the Skeleton Pirate is %d/%d, want 3/2", ch.Power, ch.Toughness)
	}
	if !slices.Contains(ch.Colors, "B") {
		t.Errorf("the Skeleton Pirate's colours are %v, want black", ch.Colors)
	}
	if !hasEffectiveKeyword(t, g, pirate, "haste") {
		t.Error("the Skeleton Pirate has no haste")
	}
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("an opponent's Skeleton has power %d, want 1", got)
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's Skeleton has haste")
	}
}

func TestFollowupCorpsesOfTheLostReturnsForALifeIfYouDescended(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		corpses := castCatalogSpell(t, g, "Corpses of the Lost", "Enchantment", fuCorpsesOracle, nil)
		passPriorityAroundTable(t, g)
		lciDGMillOne(t, g, me, "Creature — Bear")
		advanceToEndStepOf(t, g, 0)
		life := me.Life
		passPriorityAroundTable(t, g)
		if !hasPayUnlessFor(g, me.ID) {
			t.Fatalf("pay %v: the end-step trigger did not offer the life payment", pay)
		}
		answerPayUnless(t, g, me.ID, pay)
		if pay {
			if me.Life != life-1 || !me.Hand.Contains(corpses) {
				t.Errorf("paid: life %d → %d, in hand %v; want 1 life paid and the enchantment in hand",
					life, me.Life, me.Hand.Contains(corpses))
			}
		} else if me.Life != life || !g.Battlefield.Contains(corpses) {
			t.Errorf("declined: life %d → %d, on the battlefield %v; want no change",
				life, me.Life, g.Battlefield.Contains(corpses))
		}
	}
}

func TestFollowupCorpsesOfTheLostDoesNotTriggerWithoutADescent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	corpses := castCatalogSpell(t, g, "Corpses of the Lost", "Enchantment", fuCorpsesOracle, nil)
	passPriorityAroundTable(t, g)
	lciDGMillOne(t, g, me, "Instant")
	advanceToEndStepOf(t, g, 0)
	if triggersOnStackFrom(g, corpses) != 0 {
		t.Fatal("the end-step trigger went on the stack without a descent")
	}
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the life payment was offered without a descent")
	}
}

// --- Unable to Scream ----------------------------------------------------------

func TestFollowupUnableToScreamMakesA0_2ToyArtifactWithNoAbilities(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flying Bear", TypeLine: "Creature — Bear",
		Power: 4, Toughness: 4, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	enchant(t, g, "Unable to Scream", fuUnableOracle, bear)
	ch := effectiveOf(t, g, bear)
	if ch.Power != 0 || ch.Toughness != 2 {
		t.Errorf("the enchanted creature is %d/%d, want 0/2", ch.Power, ch.Toughness)
	}
	for _, typ := range []string{"Artifact", "Creature"} {
		if !slices.Contains(ch.Types, typ) {
			t.Errorf("the enchanted creature is not an %s (types %v)", typ, ch.Types)
		}
	}
	for _, sub := range []string{"Toy", "Bear"} {
		if !slices.Contains(ch.Subtypes, sub) {
			t.Errorf("the enchanted creature is not a %s (subtypes %v); the types are added", sub, ch.Subtypes)
		}
	}
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature kept flying")
	}
}

func TestFollowupUnableToScreamStopsTheEnchantedCreatureTurningFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	enchanted := fdSeatFaceDown(g, me.ID)
	other := fdSeatFaceDown(g, me.ID)
	if !fdTurnUpTiming(g, me.ID, enchanted) {
		t.Fatal("setup: the face-down permanent can't be turned face up before the Aura")
	}
	enchant(t, g, "Unable to Scream", fuUnableOracle, enchanted)
	if fdTurnUpTiming(g, me.ID, enchanted) {
		t.Error("the enchanted face-down creature can still be turned face up")
	}
	var turned bool
	g.WithWriteLock(func() { turned = g.TurnFaceUpForEffect(uuid.Nil, me.ID, enchanted) })
	if turned || !findCardAnywhere(t, g, enchanted).FaceDownIsPermanent() {
		t.Error("an effect turned the enchanted creature face up")
	}
	if !fdTurnUpTiming(g, me.ID, other) {
		t.Error("a face-down creature the Aura is not on can't be turned face up")
	}
}

// --- Veteran Survivor ----------------------------------------------------------

func TestFollowupVeteranSurvivorGrowsAndGainsHexproofAtThreeExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	survivor := b12Push(g, me.ID, "Veteran Survivor", "Creature — Human Survivor", fuVeteranSurvivor, 2, 1)
	cards := []uuid.UUID{
		gyCard(opp, "Old Bear", "Creature — Bear", "{1}{G}", ""),
		gyCard(me, "Old Ritual", "Sorcery", "{B}", ""),
		gyCard(opp, "Old Bolt", "Instant", "{R}", ""),
	}
	for i, card := range cards {
		advanceToMainOf(t, g, 0)
		if ch := effectiveOf(t, g, survivor); ch.Power != 2 || ch.Toughness != 1 {
			t.Fatalf("with %d cards exiled the Survivor is %d/%d, want 2/1", i, ch.Power, ch.Toughness)
		}
		if hasEffectiveKeyword(t, g, survivor, "hexproof") {
			t.Fatalf("with %d cards exiled the Survivor has hexproof", i)
		}
		tdmTap(g, survivor)
		advanceTo(t, g, game.StepPostcombatMain)
		pick := latestPickTarget(g, me.ID)
		if pick == nil {
			t.Fatalf("turn %d: the Survival trigger did not ask for a target", i+1)
		}
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: card}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		passPriorityAroundTable(t, g)
		if !g.Exile.Contains(card) {
			t.Fatalf("turn %d: the target card was not exiled", i+1)
		}
	}
	if ch := effectiveOf(t, g, survivor); ch.Power != 5 || ch.Toughness != 4 {
		t.Errorf("with three cards exiled the Survivor is %d/%d, want 5/4", ch.Power, ch.Toughness)
	}
	if !hasEffectiveKeyword(t, g, survivor, "hexproof") {
		t.Error("with three cards exiled the Survivor has no hexproof")
	}
}

func TestFollowupVeteranSurvivorUntappedDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	survivor := b12Push(g, me.ID, "Veteran Survivor", "Creature — Human Survivor", fuVeteranSurvivor, 2, 1)
	gyCard(opp, "Old Bear", "Creature — Bear", "{1}{G}", "")
	advanceToMainOf(t, g, 0)
	advanceTo(t, g, game.StepPostcombatMain)
	if latestPickTarget(g, me.ID) != nil || triggersOnStackFrom(g, survivor) != 0 {
		t.Error("an untapped Veteran Survivor triggered")
	}
}
