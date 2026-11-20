package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b60_test.go — GAP batch 60 (#5948): twelve Fallout artifacts.

const (
	s82GapAgilityOracle    = "83a3897c-c245-4fad-b253-3f8bbe6445fb"
	s82GapVertibirdOracle  = "3b000071-3768-44cf-8957-180d0e94c988"
	s82GapCampOracle       = "9bf498ce-6d6d-462f-9ec3-b36a2184aff7"
	s82GapEnduranceOracle  = "6a69116b-3d2e-4eac-9d02-902373de3cb7"
	s82GapSafeOracle       = "cf9129a3-25fe-4787-b41c-0b6968ba5aee"
	s82GapLuckOracle       = "d870a93a-8399-4004-9b45-2b49b001f165"
	s82GapPerceptionOracle = "f0bcdc43-c43d-4ce7-a283-b8dcab4fc497"
	s82GapPipBoyOracle     = "1e2ee630-9d37-4396-ae18-697c1a3a4e62"
	s82GapReconOracle      = "8fba247e-9849-4a06-84ff-242894b0503a"
	s82GapCostumeOracle    = "b40138c3-48bc-4825-ae48-5fea95304088"
	s82GapPowerFistOracle  = "bdfcf7df-8c06-4abe-af41-eccabb9fc2c8"
	s82GapJunkJetOracle    = "0347b8b7-0553-4a1f-9677-ad730470c345"
)

func TestS82GapB60EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82GapAgilityOracle:    "Agility Bobblehead",
		s82GapVertibirdOracle:  "Brotherhood Vertibird",
		s82GapCampOracle:       "C.A.M.P.",
		s82GapEnduranceOracle:  "Endurance Bobblehead",
		s82GapSafeOracle:       "Expert-Level Safe",
		s82GapLuckOracle:       "Luck Bobblehead",
		s82GapPerceptionOracle: "Perception Bobblehead",
		s82GapPipBoyOracle:     "Pip-Boy 3000",
		s82GapReconOracle:      "Recon Craft Theta",
		s82GapCostumeOracle:    "Silver Shroud Costume",
		s82GapPowerFistOracle:  "Power Fist",
		s82GapJunkJetOracle:    "Junk Jet",
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

// s82GapHasted puts an untapped, unsick creature with the given keywords on
// the battlefield.
func s82GapKeywordCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Human", Power: power, Toughness: toughness,
		Keywords: keywords, Owner: owner, Controller: owner,
	})
}

// --- Agility Bobblehead ---------------------------------------------------

func TestS82GapB60AgilityBobbleheadHastesUpToXAndLimitsBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bob := pipb1Bobblehead(g, me.ID, "Agility Bobblehead", s82GapAgilityOracle)
	pipb1Bobblehead(g, me.ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
	pipb1Bobblehead(g, opp.ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
	a := b12Creature(g, me.ID, "Ape", "Creature — Ape", 2, 2)
	b := b12Creature(g, me.ID, "Bat", "Creature — Bat", 2, 2)
	c := b12Creature(g, me.ID, "Cat", "Creature — Cat", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	hasty := s82GapKeywordCreature(g, opp.ID, "Raider", 1, 1, "haste")

	// X is 2: two Bobbleheads you control (the opponent's does not count).
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bob, 0, game.ActivateAbilityParams{Targets: cardRefs(a, b, c), Strict: true}); err == nil {
		t.Fatal("three targets were accepted with two Bobbleheads")
	}
	pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}", game.ActivateAbilityParams{Targets: cardRefs(a, b)})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("target %s lacks haste", id)
		}
	}
	if hasEffectiveKeyword(t, g, c, "haste") {
		t.Error("the third creature gained haste")
	}
	declareAttack(t, g, opp.ID, a, c)
	advanceTo(t, g, game.StepDeclareBlockers)
	if blockRefusal(t, g, wall, a) == "" {
		t.Error("a creature without haste blocked the chosen attacker")
	}
	if blockRefusal(t, g, hasty, a) != "" {
		t.Error("a creature with haste could not block the chosen attacker")
	}
	if blockRefusal(t, g, wall, c) != "" {
		t.Error("the unchosen attacker became unblockable")
	}
}

// --- Endurance Bobblehead -------------------------------------------------

func TestS82GapB60EnduranceBobbleheadReadsXOnceAndRunsAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bob := pipb1Bobblehead(g, me.ID, "Endurance Bobblehead", s82GapEnduranceOracle)
	other := pipb1Bobblehead(g, me.ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
	a := b12Creature(g, me.ID, "Ape", "Creature — Ape", 2, 2)
	b := b12Creature(g, me.ID, "Bat", "Creature — Bat", 2, 2)

	// Not a main phase: activate only as a sorcery.
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bob, 0, game.ActivateAbilityParams{Targets: cardRefs(a), Strict: true}); err == nil {
		t.Fatal("Endurance Bobblehead was activated outside a main phase")
	}
	advanceToMain(t, g)
	pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}", game.ActivateAbilityParams{Targets: cardRefs(a, b)})
	// X was read as the ability was activated: the other Bobblehead leaving in
	// response does not take a target away.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, _ := ptOf(t, g, id); p != 3 {
			t.Errorf("target %s has power %d, want 3", id, p)
		}
		if !hasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("target %s is not indestructible", id)
		}
	}
}

func TestS82GapB60EnduranceBobbleheadAloneTargetsOnlyOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bob := pipb1Bobblehead(g, me.ID, "Endurance Bobblehead", s82GapEnduranceOracle)
	a := b12Creature(g, me.ID, "Ape", "Creature — Ape", 2, 2)
	b := b12Creature(g, me.ID, "Bat", "Creature — Bat", 2, 2)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bob, 0, game.ActivateAbilityParams{Targets: cardRefs(a, b), Strict: true}); err == nil {
		t.Fatal("two targets were accepted with one Bobblehead")
	}
	pipb1Activate(t, g, me, bob, 0, "", game.ActivateAbilityParams{Targets: cardRefs(a)})
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, a); p != 3 {
		t.Errorf("the target has power %d, want 3", p)
	}
	if p, _ := ptOf(t, g, b); p != 2 {
		t.Errorf("the other creature has power %d, want 2", p)
	}
}

// --- Luck Bobblehead ------------------------------------------------------

func TestS82GapB60LuckBobbleheadRollsXAndMakesATappedTreasurePerEvenResult(t *testing.T) {
	sawEven := false
	for k := 0; k < 6; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		me := g.Seats[0]
		advanceToMain(t, g)
		bob := pipb1Bobblehead(g, me.ID, "Luck Bobblehead", s82GapLuckOracle)
		pipb1Bobblehead(g, me.ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
		pipb1Bobblehead(g, me.ID, "Charisma Bobblehead", pipb1CharismaBobbleheadOracle)
		pipb1Bobblehead(g, g.Seats[1].ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
		from := len(g.Events)
		pipb1Activate(t, g, me, bob, 0, "{C}", game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
		rolls := s5682Rolls(t, g, from, me.ID, bob, 6)
		if len(rolls) != 3 {
			t.Fatalf("key %d: rolled %v, want three d6 (three Bobbleheads you control)", k, rolls)
		}
		even := 0
		for _, r := range rolls {
			if r%2 == 0 {
				even++
			}
		}
		sawEven = sawEven || even > 0
		treasures, tapped := 0, 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Treasure" {
				treasures++
				if c.Tapped {
					tapped++
				}
			}
		}
		if treasures != even || tapped != even {
			t.Errorf("key %d: rolled %v: %d Treasures, %d tapped; want %d, all tapped", k, rolls, treasures, tapped, even)
		}
		if g.Outcome != nil {
			t.Errorf("key %d: the game ended after rolling %v", k, rolls)
		}
	}
	if !sawEven {
		t.Error("no key rolled an even result: the Treasure half went untested")
	}
}

// The win is the tail told the results: exactly seven sixes wins, six or
// eight do not. Rolling seven sixes through real dice is a 1-in-280,000
// event, so the tail is run on hand-made results.
func TestS82GapB60LuckBobbleheadWinsOnExactlySevenSixes(t *testing.T) {
	for _, tc := range []struct {
		rolls []int
		win   bool
	}{
		{[]int{6, 6, 6, 6, 6, 6, 6}, true},
		{[]int{6, 6, 6, 6, 6, 6, 5}, false},
		{[]int{6, 6, 6, 6, 6, 6, 6, 6}, false},
		{[]int{6, 6, 6, 6, 6, 6, 6, 1, 3}, true},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bob := pipb1Bobblehead(g, me.ID, "Luck Bobblehead", s82GapLuckOracle)
		g.WithWriteLock(func() {
			ctx := NewContext(g, &game.StackItem{Kind: game.StackItemActivated, Controller: me.ID, Owner: me.ID, SourceCardID: bob})
			// A win stops the resolution with an error: the game is over.
			if err := luckBobbleheadApply(ctx, tc.rolls); err != nil && !tc.win {
				t.Errorf("rolls %v: %v", tc.rolls, err)
			}
		})
		won := g.Outcome != nil && g.Outcome.Winner == me.ID
		if won != tc.win {
			t.Errorf("rolls %v: won = %v, want %v", tc.rolls, won, tc.win)
		}
	}
}

// --- Perception Bobblehead ------------------------------------------------

func TestS82GapB60PerceptionBobbleheadLooksAtXAndCastsOneCheapSpellFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bob := pipb1Bobblehead(g, me.ID, "Perception Bobblehead", s82GapPerceptionOracle)
	pipb1Bobblehead(g, me.ID, "Strength Bobblehead", pipb1StrengthBobbleheadOracle)
	pipb1Bobblehead(g, me.ID, "Charisma Bobblehead", pipb1CharismaBobbleheadOracle)
	deep := sunbirdLibrary(me, game.Card{Name: "Deep Card", TypeLine: "Basic Land — Plains"})[0]
	ids := sunbirdLibrary(me,
		game.Card{Name: "Mid Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3},
		game.Card{Name: "Huge Wurm", TypeLine: "Creature — Wurm", ManaCost: "{7}{G}", Power: 8, Toughness: 8},
		game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Fourth Elk", TypeLine: "Creature — Elk", ManaCost: "{G}", Power: 1, Toughness: 1},
	)
	bear, wurm, forest, elk := ids[0], ids[1], ids[2], ids[3]
	pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}", game.ActivateAbilityParams{})
	passPriorityUntilChoice(t, g)
	c := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast)
	if !hasID(c.DiscardOptions, bear) {
		t.Error("the mana value 3 spell was not offered")
	}
	for name, id := range map[string]uuid.UUID{"the mana value 8 spell": wurm, "the land": forest, "the fourth card (only three are looked at)": elk} {
		if hasID(c.DiscardOptions, id) {
			t.Errorf("%s was offered", name)
		}
	}
	if err := castThroughResolutionCast(t, g, me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	castInsideTheResolution(t, g, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the free spell did not resolve")
	}
	// The rest of the three looked-at cards went under the deep card; the
	// fourth card was not looked at and is now the top.
	for _, id := range []uuid.UUID{wurm, forest} {
		if i := libraryIndex(me, id); i < 0 || i > 2 {
			t.Errorf("a looked-at card is at index %d, want it among the bottom three", i)
		}
	}
	if i := libraryIndex(me, deep); i < 2 {
		t.Errorf("the deep card is at index %d: the looked-at rest go under it", i)
	}
	if i := libraryIndex(me, elk); i != len(me.Library.Cards)-1 {
		t.Errorf("the unlooked fourth card is at %d, want the top", i)
	}
}

// --- Brotherhood Vertibird ------------------------------------------------

func TestS82GapB60BrotherhoodVertibirdPowerCountsYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bird := b12Push(g, me.ID, "Brotherhood Vertibird", "Artifact — Vehicle", s82GapVertibirdOracle, 0, 4)
	if p, tt := ptOf(t, g, bird); p != 1 || tt != 4 {
		t.Fatalf("alone it is %d/%d, want 1/4 (it counts itself)", p, tt)
	}
	b12Permanent(g, me.ID, "Rock", "Artifact")
	b12Permanent(g, me.ID, "Pebble", "Artifact — Treasure")
	b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p, _ := ptOf(t, g, bird); p != 3 {
		t.Errorf("with two more artifacts power is %d, want 3 (an opponent's artifact does not count)", p)
	}
	b12Permanent(g, me.ID, "Shield", "Artifact — Equipment")
	if p, _ := ptOf(t, g, bird); p != 4 {
		t.Errorf("power is %d after another artifact, want 4", p)
	}
	crew := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if cardByID(g, bird).IsCreature() {
		t.Fatal("a Vehicle is not a creature before it is crewed")
	}
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !cardByID(g, bird).IsCreature() || !hasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("the crewed Vertibird is not a flying creature")
	}
}

// --- Recon Craft Theta ----------------------------------------------------

func TestS82GapB60ReconCraftThetaMakesAnAlienWithACounterAndProliferatesOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	craft := castAndResolveCreature(t, g, "Recon Craft Theta", "Artifact — Vehicle", s82GapReconOracle)
	passPriorityAroundTable(t, g)
	alien := avB1Named(g, me.ID, "Alien")
	if alien == uuid.Nil {
		t.Fatal("no Alien token")
	}
	if p, tt := ptOf(t, g, alien); p != 1 || tt != 1 {
		t.Fatalf("the Alien is %d/%d, want 1/1 (a 0/0 with a +1/+1 counter)", p, tt)
	}
	if c := cardByID(g, alien); !c.IsCreature() || len(c.Colors) != 1 || c.Colors[0] != "U" {
		t.Errorf("the Alien is %+v, want a blue creature token", c)
	}
	_ = craft

	// Attack: proliferate adds another counter to the Alien (and the opponent's
	// -1/-1 counters would grow too; none here).
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	car := b12Push(g2, me2.ID, "Recon Craft Theta", "Artifact — Vehicle", s82GapReconOracle, 4, 4)
	marked := b12Creature(g2, me2.ID, "Marked", "Creature — Bear", 2, 2)
	theirs := b12Creature(g2, opp2.ID, "Theirs", "Creature — Bear", 2, 2)
	g2.WithWriteLock(func() {
		findBattlefieldCardForTest(g2, marked).Counters = map[string]int{game.CounterPlusOne: 1}
		findBattlefieldCardForTest(g2, theirs).Counters = map[string]int{game.CounterPlusOne: 1}
	})
	if err := g2.ActivateCatalogAbility(me2.ID, car, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{marked}}); err != nil {
		t.Fatalf("crew 2: %v", err)
	}
	passPriorityAroundTable(t, g2)
	declareAttack(t, g2, opp2.ID, car)
	if n := triggersOnStackFrom(g2, car); n != 1 {
		t.Fatalf("%d proliferate triggers, want 1", n)
	}
	_ = opp
	passPriorityAroundTable(t, g2)
	answerProliferate(t, g2, marked)
	passPriorityAroundTable(t, g2)
	if n := countersOn(g2, marked, game.CounterPlusOne); n != 2 {
		t.Errorf("the chosen permanent has %d +1/+1 counters after proliferate, want 2", n)
	}
	if n := countersOn(g2, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("a permanent that was not chosen has %d counters, want 1", n)
	}
}

// --- Power Fist -----------------------------------------------------------

func TestS82GapB60PowerFistGrantsTrampleAndCountersEqualToCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	fist := pushCatalogPermanent(g, me.ID, "Power Fist", "Artifact — Equipment", s82GapPowerFistOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	pipb1Activate(t, g, me, fist, 0, "{C}{C}", game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, fist); host.ID != bear {
		t.Fatalf("Power Fist is on %v, want the Bear", host)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") || hasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("only the equipped creature has trample")
	}
	life := opp.Life
	attackWith(t, g, opp.ID, bear, plain)
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-5)
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 3 {
		t.Errorf("the equipped Bear has %d +1/+1 counters, want 3 (the damage it dealt)", n)
	}
	if n := countersOn(g, plain, game.CounterPlusOne); n != 0 {
		t.Errorf("the unequipped Bear has %d counters", n)
	}
}

// --- Silver Shroud Costume ------------------------------------------------

func TestS82GapB60SilverShroudCostumeAttachesOnEntryGivesShroudAndCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	if spec, _ := Lookup(s82GapCostumeOracle); !slicesContains(spec.PrintedKeywords, "flash") {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
	costume := castCatalogSpell(t, g, "Silver Shroud Costume", "Artifact — Equipment", s82GapCostumeOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if latestPickTarget(g, me.ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, costume); host.ID != bear {
		t.Fatalf("the Costume is on %v, want the Bear", host)
	}
	if !hasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("the creature did not gain shroud until end of turn")
	}
	declareAttack(t, g, opp.ID, bear, other)
	advanceTo(t, g, game.StepDeclareBlockers)
	if blockRefusal(t, g, wall, bear) == "" {
		t.Error("the equipped creature was blockable")
	}
	if blockRefusal(t, g, wall, other) != "" {
		t.Error("the unequipped creature became unblockable")
	}
}

// --- Junk Jet -------------------------------------------------------------

func TestS82GapB60JunkJetMakesAJunkAndDoublesTheEquippedPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	jet := castCatalogSpell(t, g, "Junk Jet", "Artifact — Equipment", s82GapJunkJetOracle, nil)
	passPriorityAroundTable(t, g)
	junk := avB1Named(g, me.ID, "Junk")
	if junk == uuid.Nil {
		t.Fatal("no Junk token when the Equipment entered")
	}
	// Doubling with nothing equipped does nothing; the sacrifice is still paid.
	spare := b12Permanent(g, me.ID, "Spare Rock", "Artifact")
	pipb1Activate(t, g, me, jet, 1, "{C}", game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, jet); host.ID != bear {
		t.Fatalf("Junk Jet is on %v, want the Bear", host)
	}
	if err := g.ActivateCatalogAbility(me.ID, jet, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{jet}}); err == nil {
		t.Fatal("Junk Jet sacrificed itself to its own \"another artifact\" cost")
	}
	pipb1Activate(t, g, me, jet, 0, "{C}{C}{C}", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{spare}})
	if g.Battlefield.Contains(spare) {
		t.Error("the sacrificed artifact is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, bear); p != 6 {
		t.Errorf("the equipped Bear has power %d, want 6", p)
	}
}

func slicesContains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// --- Pip-Boy 3000 ---------------------------------------------------------

func s82GapPipBoyAttack(t *testing.T, mode int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	boy := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pip-Boy 3000", TypeLine: "Artifact — Equipment", OracleID: s82GapPipBoyOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	_ = boy
	declareAttack(t, g, opp.ID, bear)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, mode)
	return g, me, bear, boy
}

func TestS82GapB60PipBoyPickAPerkPutsACounterOnTheAttacker(t *testing.T) {
	g, _, bear, _ := s82GapPipBoyAttack(t, 1)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the attacker has %d +1/+1 counters, want 1", n)
	}
}

func TestS82GapB60PipBoySortInventoryDrawsThenDiscards(t *testing.T) {
	g, me, _, _ := s82GapPipBoyAttack(t, 0)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	// The draw is in; the discard is the player's choice, asked next.
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	discard := me.Hand.Cards[0].InstanceID
	answerDiscard(t, g, me.ID, discard)
	if me.Hand.Size() != hand || !me.Graveyard.Contains(discard) {
		t.Errorf("hand %d, want %d after the discard", me.Hand.Size(), hand)
	}
}

func TestS82GapB60PipBoyCheckMapUntapsUpToTwoTargetLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pip-Boy 3000", TypeLine: "Artifact — Equipment", OracleID: s82GapPipBoyOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		id := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
		g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = true })
		lands = append(lands, id)
	}
	declareAttack(t, g, opp.ID, bear)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the untap mode")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: lands[0]}, {Kind: game.TargetCard, ID: lands[1]},
	}); err != nil {
		t.Fatalf("two lands: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tappedNow(g, lands[0]) || tappedNow(g, lands[1]) {
		t.Error("a chosen land stayed tapped")
	}
	if !tappedNow(g, lands[2]) {
		t.Error("the unchosen land untapped")
	}
}

// --- C.A.M.P. -------------------------------------------------------------

func s82GapCampTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	camp := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "C.A.M.P.", TypeLine: "Artifact — Fortification", OracleID: s82GapCampOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: forest},
	})
	return g, me, forest, camp
}

func TestS82GapB60CampFortifiesALandAndStaysAttached(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	forest := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	camp := pushCatalogPermanent(g, me.ID, "C.A.M.P.", "Artifact — Fortification", s82GapCampOracle, false)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, camp, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("Fortify targeted a creature")
	}
	pipb1Activate(t, g, me, camp, 0, "", game.ActivateAbilityParams{Targets: cardRefs(forest)})
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, camp); host.ID != forest {
		t.Fatalf("C.A.M.P. is on %v, want the Forest", host)
	}
	// A state-based check must not take it off a land (CR 301.6).
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, camp); host.ID != forest {
		t.Errorf("C.A.M.P. fell off its land: %v", host)
	}
}

func TestS82GapB60CampCountersATargetAndMakesJunkOnASharedColour(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors []string
		junk   int
	}{
		{"a green creature shares the Forest's colour", []string{"G"}, 1},
		{"a red creature does not", []string{"R"}, 0},
		{"a colourless creature does not", nil, 0},
	} {
		g, me, forest, _ := s82GapCampTable(t)
		creature := pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Colors: tc.colors, Owner: me.ID, Controller: me.ID,
		})
		if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: tap: %v", tc.name, err)
		}
		pickCard(t, g, me.ID, creature)
		passPriorityAroundTable(t, g)
		if n := countersOn(g, creature, game.CounterPlusOne); n != 1 {
			t.Errorf("%s: %d +1/+1 counters, want 1", tc.name, n)
		}
		junk := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Junk" {
				junk++
			}
		}
		if junk != tc.junk {
			t.Errorf("%s: %d Junk tokens, want %d", tc.name, junk, tc.junk)
		}
	}
}

func TestS82GapB60CampOnlyTriggersForTheFortifiedLand(t *testing.T) {
	g, me, _, camp := s82GapCampTable(t)
	other := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, other, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if n := triggersOnStackFrom(g, camp); n != 0 {
		t.Errorf("%d triggers for an unfortified land", n)
	}
}

// --- Expert-Level Safe ----------------------------------------------------

func s82GapSafeTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ids := sunbirdLibrary(me,
		game.Card{Name: "First", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Second", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Third", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "Fourth", TypeLine: "Basic Land — Island"},
	)
	safe := castAndResolveCreature(t, g, "Expert-Level Safe", "Artifact", s82GapSafeOracle)
	passPriorityAroundTable(t, g)
	return g, me, opp, safe, ids
}

func s82GapExiledWith(g *game.Game, safe uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(safe)
		if !ok {
			return
		}
		for _, e := range g.ExiledWithForEffect(game.PermissionCardRef{ID: safe, Epoch: c.ObjectEpoch}) {
			out = append(out, e.InstanceID)
		}
	})
	return out
}

func TestS82GapB60ExpertLevelSafeExilesTheTopTwoFaceDown(t *testing.T) {
	g, me, opp, safe, ids := s82GapSafeTable(t)
	if got := s82GapExiledWith(g, safe); len(got) != 2 || !hasID(got, ids[0]) || !hasID(got, ids[1]) {
		t.Fatalf("exiled with the Safe: %v, want the top two cards", got)
	}
	if libraryIndex(me, ids[2]) != len(me.Library.Cards)-1 {
		t.Error("the third card is not the top of the library")
	}
	for _, id := range ids[:2] {
		var found *game.Card
		for i := range g.Exile.Cards {
			if g.Exile.Cards[i].InstanceID == id {
				found = &g.Exile.Cards[i]
			}
		}
		if found == nil || !found.FaceDown || found.FaceDownKind != game.FaceDownExiled {
			t.Errorf("card %v is not face down in exile: %+v", id, found)
			continue
		}
		if found.IsKnownTo(me.ID) || found.IsKnownTo(opp.ID) {
			t.Errorf("card %v is known to a player", id)
		}
	}
}

func s82GapSafeChoose(t *testing.T, g *game.Game, me, opp *game.Player, safe uuid.UUID, mine, theirs string) {
	t.Helper()
	pipb1Activate(t, g, me, safe, 0, "{C}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	voteFor(t, g, me.ID, mine)
	voteFor(t, g, opp.ID, theirs)
}

func TestS82GapB60ExpertLevelSafeMatchingChoicesSacrificeItAndReturnTheCards(t *testing.T) {
	g, me, opp, safe, ids := s82GapSafeTable(t)
	hand := me.Hand.Size()
	s82GapSafeChoose(t, g, me, opp, safe, "2", "2")
	if g.Battlefield.Contains(safe) {
		t.Error("the Safe was not sacrificed")
	}
	if me.Hand.Size() != hand+2 || !handHas(me, ids[0]) || !handHas(me, ids[1]) {
		t.Errorf("the two exiled cards are not in hand (hand %d -> %d)", hand, me.Hand.Size())
	}
}

func TestS82GapB60ExpertLevelSafeMismatchExilesAnotherCard(t *testing.T) {
	g, me, opp, safe, ids := s82GapSafeTable(t)
	s82GapSafeChoose(t, g, me, opp, safe, "1", "3")
	if !g.Battlefield.Contains(safe) {
		t.Error("the Safe left on a mismatch")
	}
	got := s82GapExiledWith(g, safe)
	if len(got) != 3 || !hasID(got, ids[2]) {
		t.Errorf("exiled with the Safe: %v, want the third card added", got)
	}
}
