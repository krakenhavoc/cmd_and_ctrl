package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// s82GapHasted puts an untapped, unsick creature with the given keywords on
// the battlefield.
func s82GapKeywordCreature(g *game.Game, owner uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Human", Power: power, Toughness: toughness,
		Keywords: keywords, Owner: owner, Controller: owner,
	})
}

// --- Agility Bobblehead ---------------------------------------------------

func TestS82GapB60AgilityBobbleheadHastesUpToXAndLimitsBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bob := effectstest.Pipb1Bobblehead(g, me.ID, "Agility Bobblehead", s82GapAgilityOracle)
	effectstest.Pipb1Bobblehead(g, me.ID, "Strength Bobblehead", effectstest.Pipb1StrengthBobbleheadOracle)
	effectstest.Pipb1Bobblehead(g, opp.ID, "Strength Bobblehead", effectstest.Pipb1StrengthBobbleheadOracle)
	a := effectstest.B12Creature(g, me.ID, "Ape", "Creature — Ape", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bat", "Creature — Bat", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Cat", "Creature — Cat", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	hasty := s82GapKeywordCreature(g, opp.ID, "Raider", 1, 1, "haste")

	// X is 2: two Bobbleheads you control (the opponent's does not count).
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bob, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b, c), Strict: true}); err == nil {
		t.Fatal("three targets were accepted with two Bobbleheads")
	}
	effectstest.Pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}", game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("target %s lacks haste", id)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, c, "haste") {
		t.Error("the third creature gained haste")
	}
	effectstest.DeclareAttack(t, g, opp.ID, a, c)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if effectstest.BlockRefusal(t, g, wall, a) == "" {
		t.Error("a creature without haste blocked the chosen attacker")
	}
	if effectstest.BlockRefusal(t, g, hasty, a) != "" {
		t.Error("a creature with haste could not block the chosen attacker")
	}
	if effectstest.BlockRefusal(t, g, wall, c) != "" {
		t.Error("the unchosen attacker became unblockable")
	}
}

// --- Endurance Bobblehead -------------------------------------------------

func TestS82GapB60EnduranceBobbleheadReadsXOnceAndRunsAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bob := effectstest.Pipb1Bobblehead(g, me.ID, "Endurance Bobblehead", s82GapEnduranceOracle)
	other := effectstest.Pipb1Bobblehead(g, me.ID, "Strength Bobblehead", effectstest.Pipb1StrengthBobbleheadOracle)
	a := effectstest.B12Creature(g, me.ID, "Ape", "Creature — Ape", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bat", "Creature — Bat", 2, 2)

	// Not a main phase: activate only as a sorcery.
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bob, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a), Strict: true}); err == nil {
		t.Fatal("Endurance Bobblehead was activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.Pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}", game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b)})
	// X was read as the ability was activated: the other Bobblehead leaving in
	// response does not take a target away.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, _ := effectstest.PtOf(t, g, id); p != 3 {
			t.Errorf("target %s has power %d, want 3", id, p)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("target %s is not indestructible", id)
		}
	}
}

func TestS82GapB60EnduranceBobbleheadAloneTargetsOnlyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bob := effectstest.Pipb1Bobblehead(g, me.ID, "Endurance Bobblehead", s82GapEnduranceOracle)
	a := effectstest.B12Creature(g, me.ID, "Ape", "Creature — Ape", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bat", "Creature — Bat", 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bob, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b), Strict: true}); err == nil {
		t.Fatal("two targets were accepted with one Bobblehead")
	}
	effectstest.Pipb1Activate(t, g, me, bob, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardRefs(a)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 3 {
		t.Errorf("the target has power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, b); p != 2 {
		t.Errorf("the other creature has power %d, want 2", p)
	}
}

// --- Luck Bobblehead ------------------------------------------------------

func TestS82GapB60LuckBobbleheadRollsXAndMakesATappedTreasurePerEvenResult(t *testing.T) {
	sawEven := false
	for k := 0; k < 6; k++ {
		g := effectstest.NewCatalogGame(t)
		g.SetRNGKeyForTest(effectstest.RandomModesKey(byte(k)))
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		bob := effectstest.Pipb1Bobblehead(g, me.ID, "Luck Bobblehead", s82GapLuckOracle)
		effectstest.Pipb1Bobblehead(g, me.ID, "Strength Bobblehead", effectstest.Pipb1StrengthBobbleheadOracle)
		effectstest.Pipb1Bobblehead(g, me.ID, "Charisma Bobblehead", effectstest.Pipb1CharismaBobbleheadOracle)
		effectstest.Pipb1Bobblehead(g, g.Seats[1].ID, "Strength Bobblehead", effectstest.Pipb1StrengthBobbleheadOracle)
		from := len(g.Events)
		effectstest.Pipb1Activate(t, g, me, bob, 0, "{C}", game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		rolls := effectstest.S5682Rolls(t, g, from, me.ID, bob, 6)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bob := effectstest.Pipb1Bobblehead(g, me.ID, "Luck Bobblehead", s82GapLuckOracle)
		g.WithWriteLock(func() {
			ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemActivated, Controller: me.ID, Owner: me.ID, SourceCardID: bob})
			// A win stops the resolution with an error: the game is over.
			if err := cardslu.LuckBobbleheadApply(ctx, tc.rolls); err != nil && !tc.win {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bob := effectstest.Pipb1Bobblehead(g, me.ID, "Perception Bobblehead", s82GapPerceptionOracle)
	effectstest.Pipb1Bobblehead(g, me.ID, "Strength Bobblehead", effectstest.Pipb1StrengthBobbleheadOracle)
	effectstest.Pipb1Bobblehead(g, me.ID, "Charisma Bobblehead", effectstest.Pipb1CharismaBobbleheadOracle)
	deep := effectstest.SunbirdLibrary(me, game.Card{Name: "Deep Card", TypeLine: "Basic Land — Plains"})[0]
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Mid Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3},
		game.Card{Name: "Huge Wurm", TypeLine: "Creature — Wurm", ManaCost: "{7}{G}", Power: 8, Toughness: 8},
		game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Fourth Elk", TypeLine: "Creature — Elk", ManaCost: "{G}", Power: 1, Toughness: 1},
	)
	bear, wurm, forest, elk := ids[0], ids[1], ids[2], ids[3]
	effectstest.Pipb1Activate(t, g, me, bob, 0, "{C}{C}{C}", game.ActivateAbilityParams{})
	effectstest.PassPriorityUntilChoice(t, g)
	c := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast)
	if !effectstest.HasID(c.DiscardOptions, bear) {
		t.Error("the mana value 3 spell was not offered")
	}
	for name, id := range map[string]uuid.UUID{"the mana value 8 spell": wurm, "the land": forest, "the fourth card (only three are looked at)": elk} {
		if effectstest.HasID(c.DiscardOptions, id) {
			t.Errorf("%s was offered", name)
		}
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the free spell did not resolve")
	}
	// The rest of the three looked-at cards went under the deep card; the
	// fourth card was not looked at and is now the top.
	for _, id := range []uuid.UUID{wurm, forest} {
		if i := effectstest.LibraryIndex(me, id); i < 0 || i > 2 {
			t.Errorf("a looked-at card is at index %d, want it among the bottom three", i)
		}
	}
	if i := effectstest.LibraryIndex(me, deep); i < 2 {
		t.Errorf("the deep card is at index %d: the looked-at rest go under it", i)
	}
	if i := effectstest.LibraryIndex(me, elk); i != len(me.Library.Cards)-1 {
		t.Errorf("the unlooked fourth card is at %d, want the top", i)
	}
}

// --- Brotherhood Vertibird ------------------------------------------------

func TestS82GapB60BrotherhoodVertibirdPowerCountsYourArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bird := effectstest.B12Push(g, me.ID, "Brotherhood Vertibird", "Artifact — Vehicle", s82GapVertibirdOracle, 0, 4)
	if p, tt := effectstest.PtOf(t, g, bird); p != 1 || tt != 4 {
		t.Fatalf("alone it is %d/%d, want 1/4 (it counts itself)", p, tt)
	}
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Pebble", "Artifact — Treasure")
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p, _ := effectstest.PtOf(t, g, bird); p != 3 {
		t.Errorf("with two more artifacts power is %d, want 3 (an opponent's artifact does not count)", p)
	}
	effectstest.B12Permanent(g, me.ID, "Shield", "Artifact — Equipment")
	if p, _ := effectstest.PtOf(t, g, bird); p != 4 {
		t.Errorf("power is %d after another artifact, want 4", p)
	}
	crew := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.CardByID(g, bird).IsCreature() {
		t.Fatal("a Vehicle is not a creature before it is crewed")
	}
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.CardByID(g, bird).IsCreature() || !effectstest.HasEffectiveKeyword(t, g, bird, "flying") {
		t.Error("the crewed Vertibird is not a flying creature")
	}
}

// --- Recon Craft Theta ----------------------------------------------------

func TestS82GapB60ReconCraftThetaMakesAnAlienWithACounterAndProliferatesOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	craft := effectstest.CastAndResolveCreature(t, g, "Recon Craft Theta", "Artifact — Vehicle", s82GapReconOracle)
	effectstest.PassPriorityAroundTable(t, g)
	alien := effectstest.AvB1Named(g, me.ID, "Alien")
	if alien == uuid.Nil {
		t.Fatal("no Alien token")
	}
	if p, tt := effectstest.PtOf(t, g, alien); p != 1 || tt != 1 {
		t.Fatalf("the Alien is %d/%d, want 1/1 (a 0/0 with a +1/+1 counter)", p, tt)
	}
	if c := effectstest.CardByID(g, alien); !c.IsCreature() || len(c.Colors) != 1 || c.Colors[0] != "U" {
		t.Errorf("the Alien is %+v, want a blue creature token", c)
	}
	_ = craft

	// Attack: proliferate adds another counter to the Alien (and the opponent's
	// -1/-1 counters would grow too; none here).
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	car := effectstest.B12Push(g2, me2.ID, "Recon Craft Theta", "Artifact — Vehicle", s82GapReconOracle, 4, 4)
	marked := effectstest.B12Creature(g2, me2.ID, "Marked", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g2, opp2.ID, "Theirs", "Creature — Bear", 2, 2)
	g2.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g2, marked).Counters = map[string]int{game.CounterPlusOne: 1}
		effectstest.FindBattlefieldCardForTest(g2, theirs).Counters = map[string]int{game.CounterPlusOne: 1}
	})
	if err := g2.ActivateCatalogAbility(me2.ID, car, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{marked}}); err != nil {
		t.Fatalf("crew 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.DeclareAttack(t, g2, opp2.ID, car)
	if n := effectstest.TriggersOnStackFrom(g2, car); n != 1 {
		t.Fatalf("%d proliferate triggers, want 1", n)
	}
	_ = opp
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerProliferate(t, g2, marked)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.CountersOn(g2, marked, game.CounterPlusOne); n != 2 {
		t.Errorf("the chosen permanent has %d +1/+1 counters after proliferate, want 2", n)
	}
	if n := effectstest.CountersOn(g2, theirs, game.CounterPlusOne); n != 1 {
		t.Errorf("a permanent that was not chosen has %d counters, want 1", n)
	}
}

// --- Power Fist -----------------------------------------------------------

func TestS82GapB60PowerFistGrantsTrampleAndCountersEqualToCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fist := effectstest.PushCatalogPermanent(g, me.ID, "Power Fist", "Artifact — Equipment", s82GapPowerFistOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.Pipb1Activate(t, g, me, fist, 0, "{C}{C}", game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, fist); host.ID != bear {
		t.Fatalf("Power Fist is on %v, want the Bear", host)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") || effectstest.HasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("only the equipped creature has trample")
	}
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-5)
	}
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 3 {
		t.Errorf("the equipped Bear has %d +1/+1 counters, want 3 (the damage it dealt)", n)
	}
	if n := effectstest.CountersOn(g, plain, game.CounterPlusOne); n != 0 {
		t.Errorf("the unequipped Bear has %d counters", n)
	}
}

// --- Silver Shroud Costume ------------------------------------------------

func TestS82GapB60SilverShroudCostumeAttachesOnEntryGivesShroudAndCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	if spec, _ := effects.Lookup(s82GapCostumeOracle); !slicesContains(spec.PrintedKeywords, "flash") {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
	costume := effectstest.CastCatalogSpell(t, g, "Silver Shroud Costume", "Artifact — Equipment", s82GapCostumeOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			break
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, costume); host.ID != bear {
		t.Fatalf("the Costume is on %v, want the Bear", host)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("the creature did not gain shroud until end of turn")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear, other)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if effectstest.BlockRefusal(t, g, wall, bear) == "" {
		t.Error("the equipped creature was blockable")
	}
	if effectstest.BlockRefusal(t, g, wall, other) != "" {
		t.Error("the unequipped creature became unblockable")
	}
}

// --- Junk Jet -------------------------------------------------------------

func TestS82GapB60JunkJetMakesAJunkAndDoublesTheEquippedPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	jet := effectstest.CastCatalogSpell(t, g, "Junk Jet", "Artifact — Equipment", s82GapJunkJetOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	junk := effectstest.AvB1Named(g, me.ID, "Junk")
	if junk == uuid.Nil {
		t.Fatal("no Junk token when the Equipment entered")
	}
	// Doubling with nothing equipped does nothing; the sacrifice is still paid.
	spare := effectstest.B12Permanent(g, me.ID, "Spare Rock", "Artifact")
	effectstest.Pipb1Activate(t, g, me, jet, 1, "{C}", game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, jet); host.ID != bear {
		t.Fatalf("Junk Jet is on %v, want the Bear", host)
	}
	if err := g.ActivateCatalogAbility(me.ID, jet, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{jet}}); err == nil {
		t.Fatal("Junk Jet sacrificed itself to its own \"another artifact\" cost")
	}
	effectstest.Pipb1Activate(t, g, me, jet, 0, "{C}{C}{C}", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{spare}})
	if g.Battlefield.Contains(spare) {
		t.Error("the sacrificed artifact is still on the battlefield")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 6 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	boy := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pip-Boy 3000", TypeLine: "Artifact — Equipment", OracleID: s82GapPipBoyOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	_ = boy
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, mode)
	return g, me, bear, boy
}

func TestS82GapB60PipBoyPickAPerkPutsACounterOnTheAttacker(t *testing.T) {
	g, _, bear, _ := s82GapPipBoyAttack(t, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the attacker has %d +1/+1 counters, want 1", n)
	}
}

func TestS82GapB60PipBoySortInventoryDrawsThenDiscards(t *testing.T) {
	g, me, _, _ := s82GapPipBoyAttack(t, 0)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	// The draw is in; the discard is the player's choice, asked next.
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
	discard := me.Hand.Cards[0].InstanceID
	effectstest.AnswerDiscard(t, g, me.ID, discard)
	if me.Hand.Size() != hand || !me.Graveyard.Contains(discard) {
		t.Errorf("hand %d, want %d after the discard", me.Hand.Size(), hand)
	}
}

func TestS82GapB60PipBoyCheckMapUntapsUpToTwoTargetLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pip-Boy 3000", TypeLine: "Artifact — Equipment", OracleID: s82GapPipBoyOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear},
	})
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		id := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = true })
		lands = append(lands, id)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the untap mode")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: lands[0]}, {Kind: game.TargetCard, ID: lands[1]},
	}); err != nil {
		t.Fatalf("two lands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, lands[0]) || effectstest.TappedNow(g, lands[1]) {
		t.Error("a chosen land stayed tapped")
	}
	if !effectstest.TappedNow(g, lands[2]) {
		t.Error("the unchosen land untapped")
	}
}

// --- C.A.M.P. -------------------------------------------------------------

func s82GapCampTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	camp := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "C.A.M.P.", TypeLine: "Artifact — Fortification", OracleID: s82GapCampOracle,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: forest},
	})
	return g, me, forest, camp
}

func TestS82GapB60CampFortifiesALandAndStaysAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	camp := effectstest.PushCatalogPermanent(g, me.ID, "C.A.M.P.", "Artifact — Fortification", s82GapCampOracle, false)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, camp, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("Fortify targeted a creature")
	}
	effectstest.Pipb1Activate(t, g, me, camp, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardRefs(forest)})
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, camp); host.ID != forest {
		t.Fatalf("C.A.M.P. is on %v, want the Forest", host)
	}
	// A state-based check must not take it off a land (CR 301.6).
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, camp); host.ID != forest {
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
		creature := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Colors: tc.colors, Owner: me.ID, Controller: me.ID,
		})
		if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("%s: tap: %v", tc.name, err)
		}
		effectstest.PickCard(t, g, me.ID, creature)
		effectstest.PassPriorityAroundTable(t, g)
		if n := effectstest.CountersOn(g, creature, game.CounterPlusOne); n != 1 {
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
	other := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, other, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if n := effectstest.TriggersOnStackFrom(g, camp); n != 0 {
		t.Errorf("%d triggers for an unfortified land", n)
	}
}

// --- Expert-Level Safe ----------------------------------------------------

func s82GapSafeTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "First", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Second", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Third", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "Fourth", TypeLine: "Basic Land — Island"},
	)
	safe := effectstest.CastAndResolveCreature(t, g, "Expert-Level Safe", "Artifact", s82GapSafeOracle)
	effectstest.PassPriorityAroundTable(t, g)
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
	if got := s82GapExiledWith(g, safe); len(got) != 2 || !effectstest.HasID(got, ids[0]) || !effectstest.HasID(got, ids[1]) {
		t.Fatalf("exiled with the Safe: %v, want the top two cards", got)
	}
	if effectstest.LibraryIndex(me, ids[2]) != len(me.Library.Cards)-1 {
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
	effectstest.Pipb1Activate(t, g, me, safe, 0, "{C}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteFor(t, g, me.ID, mine)
	effectstest.VoteFor(t, g, opp.ID, theirs)
}

func TestS82GapB60ExpertLevelSafeMatchingChoicesSacrificeItAndReturnTheCards(t *testing.T) {
	g, me, opp, safe, ids := s82GapSafeTable(t)
	hand := me.Hand.Size()
	s82GapSafeChoose(t, g, me, opp, safe, "2", "2")
	if g.Battlefield.Contains(safe) {
		t.Error("the Safe was not sacrificed")
	}
	if me.Hand.Size() != hand+2 || !effectstest.HandHas(me, ids[0]) || !effectstest.HandHas(me, ids[1]) {
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
	if len(got) != 3 || !effectstest.HasID(got, ids[2]) {
		t.Errorf("exiled with the Safe: %v, want the third card added", got)
	}
}
