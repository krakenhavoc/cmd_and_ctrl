package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_ac2_cards_test.go — The Lost Caverns of Ixalan, cards A–C, second
// batch (#4470). Each test plays the real catalog entry and asserts what
// its text says.

const (
	lciAC2AbueloOracle     = "eaa9baac-72f2-4bde-9146-860878d9860b"
	lciAC2GemguardOracle   = "2fdc0d15-ecb4-4a6d-be77-7a3fca188c9c"
	lciAC2AkalPakalOracle  = "53bfe54e-9b97-4113-97d9-3c5dbc85e7cf"
	lciAC2AkawalliOracle   = "0679dcef-e1c3-4739-a769-18a4ffdcfde7"
	lciAC2CapybaraOracle   = "0ce678fa-6495-4857-b178-231c283b4514"
	lciAC2MycoidOracle     = "eeba4706-b4f2-48f9-bcb1-e42a5fc68c40"
	lciAC2CavalryOracle    = "9da7e8b5-887d-4174-b7aa-982fd052fcd1"
	lciAC2CanonizedOracle  = "d491995f-e9b2-4187-852c-8a4c2495eaca"
	lciAC2MineCartOracle   = "c3775748-ff62-469f-916b-8e551ccce23c"
	lciAC2VolcanoOracle    = "701c8624-70d8-49df-84b1-d7b2a8e83552"
	lciAC2ChupacabraOracle = "68c56db4-c97a-4c1b-a6f6-096392488e21"
	lciAC2CoatiOracle      = "535f9bc6-9a07-4850-91eb-c00d06633e7e"
	lciAC2CompassOracle    = "b64051c0-e791-4fcb-b79f-4287e1b06468"
	lciAC2RiddleOracle     = "3df49c4d-a285-4826-824e-b742b0deed3a"
	lciAC2CouncilOracle    = "7b513bd0-27df-45f3-a85f-1f0aba3cae48"
)

func TestLciAC2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciAC2AbueloOracle:     "Abuelo, Ancestral Echo",
		lciAC2GemguardOracle:   "Adaptive Gemguard",
		lciAC2AkalPakalOracle:  "Akal Pakal, First Among Equals",
		lciAC2AkawalliOracle:   "Akawalli, the Seething Tower",
		lciAC2CapybaraOracle:   "Basking Capybara",
		lciAC2MycoidOracle:     "Broodrage Mycoid",
		lciAC2CavalryOracle:    "Burning Sun Cavalry",
		lciAC2CanonizedOracle:  "Canonized in Blood",
		lciAC2MineCartOracle:   "Careening Mine Cart",
		lciAC2VolcanoOracle:    "Child of the Volcano",
		lciAC2ChupacabraOracle: "Chupacabra Echo",
		lciAC2CoatiOracle:      "Coati Scavenger",
		lciAC2CompassOracle:    "Compass Gnome",
		lciAC2RiddleOracle:     "Confounding Riddle",
		lciAC2CouncilOracle:    "Council of Echoes",
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

// lciAC2SeedGraveyard puts n permanent cards and a sorcery into p's
// graveyard and refreshes the layers, returning the last permanent card.
func lciAC2SeedGraveyard(g *game.Game, p *game.Player, n int) uuid.UUID {
	var last uuid.UUID
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			last = pushGraveyardCardTyped(p, "Old Land", "Land")
		}
		pushGraveyardCardTyped(p, "Old Sorcery", "Sorcery")
		g.BumpLayerVersionForTest()
	})
	return last
}

// lciAC2CastAndPick casts a creature with an enters trigger, picks the
// target if asked, and resolves the trigger.
func lciAC2CastAndPick(t *testing.T, g *game.Game, me *game.Player, c game.Card, target uuid.UUID) uuid.UUID {
	t.Helper()
	id, err := msc17Cast(t, g, me, c, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	passPriorityAroundTable(t, g)
	if target != uuid.Nil {
		pickIfAsked(t, g, me.ID, target)
	}
	passPriorityAroundTable(t, g)
	return id
}

// --- Abuelo, Ancestral Echo ------------------------------------------------------

// The artifact leaves now and comes back at the next end step as a new
// object.
func TestLciAC2AbueloBlinksAnArtifactYouControlUntilTheEndStep(t *testing.T) {
	g, me, _ := spendTable(t)
	abuelo := pushCatalogPermanent(g, me.ID, "Abuelo, Ancestral Echo", "Legendary Creature — Spirit", lciAC2AbueloOracle, false)
	relic := b12Permanent(g, me.ID, "My Relic", "Artifact")
	if !hasEffectiveKeyword(t, g, abuelo, "flying") {
		t.Error("Abuelo has no flying")
	}
	apaMana(me, "C", "W", "U")
	p7Activate(t, g, me, abuelo, 0, game.ActivateAbilityParams{Targets: rfCardTarget(relic), Strict: true})
	if onBattlefieldByName(g, "My Relic") {
		t.Fatal("the artifact is still on the battlefield before the end step")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	back := roonaOnly(t, g, "My Relic")
	if back.InstanceID == relic {
		t.Error("the artifact came back as the same object (CR 400.7)")
	}
}

// "Another target creature or artifact you control": not Abuelo itself,
// and not an opponent's creature.
func TestLciAC2AbueloRefusesItselfAndAnOpponentsCreature(t *testing.T) {
	g, me, opp := spendTable(t)
	abuelo := pushCatalogPermanent(g, me.ID, "Abuelo, Ancestral Echo", "Legendary Creature — Spirit", lciAC2AbueloOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, target := range []uuid.UUID{abuelo, theirs} {
		apaMana(me, "C", "W", "U")
		if err := g.ActivateCatalogAbility(me.ID, abuelo, 0, game.ActivateAbilityParams{Targets: rfCardTarget(target)}); err == nil {
			t.Errorf("Abuelo's ability accepted an illegal target %v", target)
		}
	}
}

// --- Adaptive Gemguard -----------------------------------------------------------

func TestLciAC2AdaptiveGemguardTapsTwoForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	gem := b12Push(g, me.ID, "Adaptive Gemguard", "Artifact Creature — Gnome", lciAC2GemguardOracle, 3, 3)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	land := pushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	if err := g.ActivateCatalogAbility(me.ID, gem, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{gem, land}, Strict: true}); err == nil {
		t.Fatal("a land paid for \"artifacts and/or creatures\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, gem, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{gem, relic}, Strict: true}); err != nil {
		t.Fatalf("tapping the Gemguard and an artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, gem, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestLciAC2AdaptiveGemguardIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gem := b12Push(g, me.ID, "Adaptive Gemguard", "Artifact Creature — Gnome", lciAC2GemguardOracle, 3, 3)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, gem, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{gem, relic}, Strict: true}); err == nil {
		t.Error("the ability was activated at the beginning of combat")
	}
}

// --- Akal Pakal, First Among Equals ---------------------------------------------

func TestLciAC2AkalPakalLooksAtTwoAfterAnArtifactEnteredUnderYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Akal Pakal, First Among Equals", "Legendary Creature — Human Advisor", lciAC2AkalPakalOracle, 1, 5)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	lower := topLibraryCard(me, "Lower", "Sorcery")
	upper := topLibraryCard(me, "Upper", "Sorcery")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick between the top two cards at the end step")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top two", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{upper}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(upper) || !me.Graveyard.Contains(lower) {
		t.Error("the chosen card should be in hand and the other in the graveyard")
	}
}

func TestLciAC2AkalPakalIgnoresAnOpponentsArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	akal := b12Push(g, me.ID, "Akal Pakal, First Among Equals", "Legendary Creature — Human Advisor", lciAC2AkalPakalOracle, 1, 5)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1) })
	advanceToEndStepOf(t, g, 0)
	if triggersOnStackFrom(g, akal) != 0 || latestChooseCardsFor(g, me.ID) != nil {
		t.Error("Akal Pakal triggered with only an opponent's artifact having entered")
	}
}

// --- Akawalli, the Seething Tower ------------------------------------------------

func TestLciAC2AkawalliGrowsWithDescendFourAndEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ak := b12Push(g, me.ID, "Akawalli, the Seething Tower", "Legendary Creature — Fungus", lciAC2AkawalliOracle, 3, 3)
	lciAC2SeedGraveyard(g, me, 3)
	if p, tough := effectivePower(t, g, ak), effectiveToughness(t, g, ak); p != 3 || tough != 3 {
		t.Errorf("with three permanent cards Akawalli is %d/%d, want 3/3", p, tough)
	}
	if hasEffectiveKeyword(t, g, ak, "trample") {
		t.Error("trample with three permanent cards")
	}
	lciAC2SeedGraveyard(g, me, 1)
	if p, tough := effectivePower(t, g, ak), effectiveToughness(t, g, ak); p != 5 || tough != 5 {
		t.Errorf("with four permanent cards Akawalli is %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, ak, "trample") {
		t.Error("no trample with four permanent cards")
	}
	lciAC2SeedGraveyard(g, me, 4)
	if p, tough := effectivePower(t, g, ak), effectiveToughness(t, g, ak); p != 7 || tough != 7 {
		t.Errorf("with eight permanent cards Akawalli is %d/%d, want 7/7", p, tough)
	}
}

func TestLciAC2AkawalliCantBeDoubleBlockedWithDescendEight(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cards   int
		refused bool
	}{
		{"four permanent cards", 4, false},
		{"eight permanent cards", 8, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			ak := b12Push(g, me.ID, "Akawalli, the Seething Tower", "Legendary Creature — Fungus", lciAC2AkawalliOracle, 3, 3)
			a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
			b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
			lciAC2SeedGraveyard(g, me, tc.cards)
			brAttack(t, g, ak)
			err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: a, Attacker: ak}, {Blocker: b, Attacker: ak}})
			if tc.refused {
				brRefusal(t, err, game.BlockReasonTooManyBlockers)
			} else if err != nil {
				t.Errorf("two blockers refused with only four permanent cards: %v", err)
			}
		})
	}
}

// --- Basking Capybara ------------------------------------------------------------

func TestLciAC2BaskingCapybaraGetsPlusThreeWithDescendFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	capy := b12Push(g, me.ID, "Basking Capybara", "Creature — Capybara", lciAC2CapybaraOracle, 1, 3)
	lciDGSeedDescend(g, me)
	if p, tough := effectivePower(t, g, capy), effectiveToughness(t, g, capy); p != 1 || tough != 3 {
		t.Errorf("with three permanent cards the Capybara is %d/%d, want 1/3", p, tough)
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	if p, tough := effectivePower(t, g, capy), effectiveToughness(t, g, capy); p != 4 || tough != 3 {
		t.Errorf("with four permanent cards the Capybara is %d/%d, want 4/3", p, tough)
	}
}

// --- Broodrage Mycoid ------------------------------------------------------------

func TestLciAC2BroodrageMycoidMakesAFungusThatCantBlockIfYouDescended(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		want     int
	}{
		{"an artifact card milled", "Artifact", 1},
		{"only an instant card milled", "Instant", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			b12Push(g, me.ID, "Broodrage Mycoid", "Creature — Fungus", lciAC2MycoidOracle, 4, 3)
			lciDGMillOne(t, g, me, tc.typeLine)
			advanceToEndStepOf(t, g, 0)
			passPriorityAroundTable(t, g)
			fungi := trekB33Tokens(g, me.ID, "Fungus")
			if len(fungi) != tc.want {
				t.Fatalf("%d Fungus tokens, want %d", len(fungi), tc.want)
			}
			if tc.want == 1 {
				if f := fungi[0]; f.Power != 1 || f.Toughness != 1 || !f.HasSubtype("Fungus") {
					t.Errorf("the token is a %d/%d %q, want a 1/1 Fungus", f.Power, f.Toughness, f.TypeLine)
				}
			}
		})
	}
}

// --- Burning Sun Cavalry ---------------------------------------------------------

func TestLciAC2BurningSunCavalryGrowsWhenItAttacksWhileYouControlADinosaur(t *testing.T) {
	for _, dino := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cav := b12Push(g, me.ID, "Burning Sun Cavalry", "Creature — Human Knight", lciAC2CavalryOracle, 2, 2)
		if dino {
			b12Creature(g, me.ID, "My Dinosaur", "Creature — Dinosaur", 3, 3)
		} else {
			b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
		}
		advanceToMain(t, g)
		lciB1Attack(t, g, cav, opp.ID)
		want := 2
		if dino {
			want = 3
		}
		if p, tough := effectivePower(t, g, cav), effectiveToughness(t, g, cav); p != want || tough != want {
			t.Errorf("dinosaur=%v: the attacking Cavalry is %d/%d, want %d/%d", dino, p, tough, want, want)
		}
	}
}

// --- Canonized in Blood ----------------------------------------------------------

func TestLciAC2CanonizedInBloodCountersYourCreatureIfYouDescended(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Canonized in Blood", "Enchantment", lciAC2CanonizedOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	lciDGMillOne(t, g, me, "Creature — Bear")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	pickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the creature you control, want 1", got)
	}
}

func TestLciAC2CanonizedInBloodSkipsTheEndStepWithoutADescent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	canon := pushCatalogPermanent(g, me.ID, "Canonized in Blood", "Enchantment", lciAC2CanonizedOracle, false)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	lciDGMillOne(t, g, me, "Sorcery")
	advanceToEndStepOf(t, g, 0)
	if triggersOnStackFrom(g, canon) != 0 || latestPickTarget(g, me.ID) != nil {
		t.Error("the end-step trigger went on the stack without a descent")
	}
}

func TestLciAC2CanonizedInBloodSacrificesForAVampireDemon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	canon := pushCatalogPermanent(g, me.ID, "Canonized in Blood", "Enchantment", lciAC2CanonizedOracle, false)
	advanceToMain(t, g)
	fillPoolColored(me, "B", 6)
	if err := g.ActivateCatalogAbility(me.ID, canon, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with six mana for {5}{B}{B}")
	}
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, canon, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(canon) {
		t.Error("the enchantment was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Vampire Demon", me.ID); got != 1 {
		t.Fatalf("%d Vampire Demons, want 1", got)
	}
	demon := trekB33Tokens(g, me.ID, "Vampire Demon")[0]
	if demon.Power != 4 || demon.Toughness != 3 || !hasEffectiveKeyword(t, g, demon.InstanceID, "flying") {
		t.Errorf("the token is a %d/%d, want a 4/3 with flying", demon.Power, demon.Toughness)
	}
}

// --- Careening Mine Cart ---------------------------------------------------------

func TestLciAC2CareeningMineCartMakesATreasureWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cart := dftABPush(g, me.ID, "Careening Mine Cart", dftABVehicle, lciAC2MineCartOracle, 3, 3)
	pilot := b12Creature(g, me.ID, "Pilot", "Creature — Gnome", 1, 1)
	advanceToMain(t, g)
	dftABCrew(t, g, me, cart, 0, pilot)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 0 {
		t.Fatalf("%d Treasures before the attack, want 0", n)
	}
	lciB1Attack(t, g, cart, opp.ID)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after the cart attacked, want 1", n)
	}
}

// --- Child of the Volcano --------------------------------------------------------

func TestLciAC2ChildOfTheVolcanoGrowsAtYourEndStepIfYouDescended(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		want     int
	}{
		{"a permanent card milled", "Creature — Bear", 1},
		{"only a sorcery card milled", "Sorcery", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			child := b12Push(g, me.ID, "Child of the Volcano", "Creature — Elemental", lciAC2VolcanoOracle, 3, 3)
			if !hasEffectiveKeyword(t, g, child, "trample") {
				t.Error("Child of the Volcano has no trample")
			}
			lciDGMillOne(t, g, me, tc.typeLine)
			advanceToEndStepOf(t, g, 0)
			passPriorityAroundTable(t, g)
			if got := countersOn(g, child, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

// --- Chupacabra Echo -------------------------------------------------------------

func TestLciAC2ChupacabraEchoShrinksByPermanentCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	lciDGSeedDescend(g, me)
	advanceToMain(t, g)
	if _, err := msc17Cast(t, g, me, game.Card{Name: "Chupacabra Echo", TypeLine: "Creature — Beast Horror Spirit", OracleID: lciAC2ChupacabraOracle, Power: 3, Toughness: 2}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Chupacabra Echo: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, c := range p.PickTargetCards {
			if c == mine {
				t.Error("the controller's own creature is offered")
			}
		}
	}
	pickIfAsked(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, victim), effectiveToughness(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("with three permanent cards and a sorcery the target is %d/%d, want 1/1", p, tough)
	}
}

// --- Coati Scavenger -------------------------------------------------------------

func TestLciAC2CoatiScavengerReturnsAPermanentCardWithDescendFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := lciAC2SeedGraveyard(g, me, 4)
	lciAC2CastAndPick(t, g, me, game.Card{Name: "Coati Scavenger", TypeLine: "Creature — Raccoon", OracleID: lciAC2CoatiOracle, Power: 3, Toughness: 2}, land)
	if !me.Hand.Contains(land) || me.Graveyard.Contains(land) {
		t.Error("the targeted permanent card did not return to hand")
	}
}

func TestLciAC2CoatiScavengerDoesNothingWithThreePermanentCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := lciAC2SeedGraveyard(g, me, 3)
	coati := lciAC2CastAndPick(t, g, me, game.Card{Name: "Coati Scavenger", TypeLine: "Creature — Raccoon", OracleID: lciAC2CoatiOracle, Power: 3, Toughness: 2}, uuid.Nil)
	if triggersOnStackFrom(g, coati) != 0 || latestPickTarget(g, me.ID) != nil {
		t.Error("the enters trigger went on the stack with three permanent cards")
	}
	if !me.Graveyard.Contains(land) {
		t.Error("a card left the graveyard without descend 4")
	}
}

// --- Compass Gnome ---------------------------------------------------------------

func TestLciAC2CompassGnomePutsACaveOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Echoing Deeps", "Land — Cave"),
		searchTestLand("Grizzly Bears", "Creature — Bear"),
		searchTestLand("Mystic Sanctuary", "Land — Island"),
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	fdnCCast(t, g, "Compass Gnome", "Artifact Creature — Gnome", lciAC2CompassOracle, 2, 1)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger did not offer a search")
	}
	if searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil || searchOptionNamed(g, c, "Mystic Sanctuary") != uuid.Nil {
		t.Error("only basic land cards and Cave cards are legal finds")
	}
	if searchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Error("a basic land card is not offered")
	}
	answerSearchNamed(t, g, me.ID, "Echoing Deeps")
	passPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.Name != "Echoing Deeps" {
		t.Fatalf("the top card is %q, want Echoing Deeps", top.Name)
	}
	if me.Library.Size() != 4 {
		t.Errorf("library = %d, the card never left it", me.Library.Size())
	}
}

// --- Confounding Riddle ----------------------------------------------------------

func TestLciAC2ConfoundingRiddleTakesOneOfFourAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var top []uuid.UUID
	for _, name := range []string{"D", "C", "B", "A"} {
		top = append(top, topLibraryCard(me, name, "Sorcery"))
	}
	if err := rfCastModal(t, g, "Confounding Riddle", lciAC2RiddleOracle, []int{0}, nil); err != nil {
		t.Fatalf("cast Confounding Riddle: %v", err)
	}
	lib := me.Library.Size()
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top four cards")
	}
	if len(pick.ChooseCards) != 4 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top four", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) {
		t.Error("the chosen card is not in hand")
	}
	for _, id := range []uuid.UUID{top[0], top[2], top[3]} {
		if !me.Graveyard.Contains(id) {
			t.Error("one of the other three is not in the graveyard")
		}
	}
	if me.Library.Size() != lib-4 {
		t.Errorf("library %d -> %d, want exactly the top four gone", lib, me.Library.Size())
	}
}

func TestLciAC2ConfoundingRiddleCountersUnlessFourIsPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := hobA7BearSpell(t, g, me)
		if err := rfCastModal(t, g, "Confounding Riddle", lciAC2RiddleOracle, []int{1}, rfCardTarget(bear)); err != nil {
			t.Fatalf("cast Confounding Riddle: %v", err)
		}
		passUntilTaxed(t, g, me.ID)
		if pay {
			b06AddMana(me, "C", "C", "C", "C")
		}
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		if resolved := g.Battlefield.Contains(bear); resolved != pay {
			t.Errorf("pay=%v: the spell resolved=%v", pay, resolved)
		}
	}
}

// --- Council of Echoes -----------------------------------------------------------

func TestLciAC2CouncilOfEchoesBouncesWithDescendFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	lciAC2SeedGraveyard(g, me, 4)
	council := lciAC2CastAndPick(t, g, me, game.Card{Name: "Council of Echoes", TypeLine: "Creature — Spirit Advisor", OracleID: lciAC2CouncilOracle, Power: 4, Toughness: 4}, victim)
	if !opp.Hand.Contains(victim) || g.Battlefield.Contains(victim) {
		t.Error("the targeted nonland permanent did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(council) {
		t.Error("the Council left the battlefield")
	}
	if !hasEffectiveKeyword(t, g, council, "flying") {
		t.Error("Council of Echoes has no flying")
	}
}

func TestLciAC2CouncilOfEchoesDoesNothingWithThreePermanentCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	lciAC2SeedGraveyard(g, me, 3)
	council := lciAC2CastAndPick(t, g, me, game.Card{Name: "Council of Echoes", TypeLine: "Creature — Spirit Advisor", OracleID: lciAC2CouncilOracle, Power: 4, Toughness: 4}, uuid.Nil)
	if triggersOnStackFrom(g, council) != 0 || latestPickTarget(g, me.ID) != nil {
		t.Error("the enters trigger went on the stack with three permanent cards")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("a permanent was bounced without descend 4")
	}
}
