package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dsk_rt_cards_test.go — Duskmourn: House of Horror, the R–T cards
// (#4466). Each test plays the real catalog entry and asserts what its
// text says.

const (
	dskRTCarnivalOracle     = "24a97436-ba61-4ebc-a560-a6c027ccfdf3"
	dskRTGorgeOracle        = "8f69bd3a-244e-42d8-bfac-5a426f4b54b4"
	dskRTCemeteryOracle     = "3a5b3405-a1e3-4aad-ab4e-1b8db2d1f3a8"
	dskRTShepherdingOracle  = "e8ff3e99-eca9-44a5-baac-99ee34df3c36"
	dskRTBranchsnapperOrcl  = "29326487-6d95-40a8-9bcc-f94ada4e27c3"
	dskRTSnatcherOracle     = "371762f6-e07b-4c0e-a4ee-6c35a3809eb6"
	dskRTSurveyorOracle     = "7aae54a6-3011-4867-b5ee-21a2111aecc6"
	dskRTHordecallerOracle  = "3fc4bdf1-7da3-4f9f-806f-f82cd18ea65f"
	dskRTStorytellerOracle  = "ec238819-81f6-4904-922d-298a17cbac7e"
	dskRTSaviorOracle       = "15f2085f-d5e6-4bc6-870f-00fcaccd981b"
	dskRTShroudstomperOrcl  = "36415fe3-6bf3-4fec-909c-6c16584c44eb"
	dskRTRipchainOracle     = "ce000c0b-db42-4569-855f-f4eae0431c09"
	dskRTSeizedOracle       = "7ac608bc-dc2a-43c9-9b3b-1b985c121d60"
	dskRTStartingLife       = 40
	dskRTShroudstomperDrain = 2
)

func TestDskRTEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dskRTCarnivalOracle:    "Raucous Carnival",
		dskRTGorgeOracle:       "Razortrap Gorge",
		dskRTCemeteryOracle:    "Strangled Cemetery",
		dskRTShepherdingOracle: "Shepherding Spirits",
		dskRTBranchsnapperOrcl: "Slavering Branchsnapper",
		dskRTSnatcherOracle:    "Spectral Snatcher",
		dskRTSurveyorOracle:    "Tunnel Surveyor",
		dskRTHordecallerOracle: "Razorkin Hordecaller",
		dskRTStorytellerOracle: "Shrewd Storyteller",
		dskRTSaviorOracle:      "Savior of the Small",
		dskRTShroudstomperOrcl: "Shroudstomper",
		dskRTRipchainOracle:    "Ripchain Razorkin",
		dskRTSeizedOracle:      "Seized from Slumber",
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

// --- Raucous Carnival, Razortrap Gorge, Strangled Cemetery ----------------------

func TestDskRTTaplandsEnterTappedUnlessAPlayerIsAtThirteen(t *testing.T) {
	for _, tc := range []struct{ name, oracle, produced string }{
		{"Raucous Carnival", dskRTCarnivalOracle, "{R|W}"},
		{"Razortrap Gorge", dskRTGorgeOracle, "{B|R}"},
		{"Strangled Cemetery", dskRTCemeteryOracle, "{B|G}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			land := b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})
			if !b13Tapped(t, g, land) {
				t.Error("everyone at 40: the land entered untapped")
			}
			if got := ltProduced(t, g, land); !slices.Equal(got, []string{tc.produced}) {
				t.Errorf("the land adds %v, want %s", got, tc.produced)
			}
			g.Seats[3].Life = 13
			if b13Tapped(t, g, b12PlayFromHand(t, g, tc.name, "Land", tc.oracle, game.CastSpellParams{})) {
				t.Error("an opponent at 13: the land entered tapped")
			}
		})
	}
}

// --- Shepherding Spirits, Slavering Branchsnapper, Spectral Snatcher ------------

func TestDskRTLandcyclersFindOnlyTheirLandType(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword, wantLand, otherLand string
	}{
		{"Shepherding Spirits", "Creature — Spirit", dskRTShepherdingOracle, "flying", "Plains", "Island"},
		{"Slavering Branchsnapper", "Creature — Lizard", dskRTBranchsnapperOrcl, "trample", "Forest", "Mountain"},
		{"Spectral Snatcher", "Creature — Spirit", dskRTSnatcherOracle, "", "Swamp", "Forest"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			if tc.keyword != "" {
				id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 4, 4)
				if !hasKeywordOf(g, id, tc.keyword) {
					t.Errorf("%s does not have %s", tc.name, tc.keyword)
				}
			}
			g.WithWriteLock(func() { me.Hand.Cards = nil })
			ids := seedSearchLibrary(me,
				searchTestLand(tc.wantLand, "Basic Land — "+tc.wantLand),
				searchTestLand(tc.otherLand, "Basic Land — "+tc.otherLand),
				game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
			)
			id, _ := cycleFromHand(t, g, tc.name, tc.typeLine, tc.oracle, "{C}{C}")
			if !me.Graveyard.Contains(id) {
				t.Fatal("the cycled card is not in the graveyard")
			}
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
				t.Fatalf("search = %+v, want only the %s", c, tc.wantLand)
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
				t.Fatalf("ResolveSearchLibrary: %v", err)
			}
			if !me.Hand.Contains(ids[0]) {
				t.Errorf("the %s did not reach the hand", tc.wantLand)
			}
		})
	}
}

func TestDskRTSpectralSnatcherHasWardDiscardACard(t *testing.T) {
	g, _, opp, warded := wardTable(t, "Spectral Snatcher", "Creature — Spirit", dskRTSnatcherOracle)
	hand := fillHand(opp, "Filler A", "Filler B")
	bolt := boltWardedCreature(t, g, opp, warded)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	answerDiscard(t, g, opp.ID, hand[0])
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(hand[0]) {
		t.Error("the ward's discard did not happen")
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the Lightning Bolt should have resolved and gone to the graveyard")
	}
}

// --- Tunnel Surveyor --------------------------------------------------------------

func TestDskRTTunnelSurveyorMakesAGlimmerOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castFromHand(t, g, game.Card{
		Name: "Tunnel Surveyor", TypeLine: "Creature — Human Detective", OracleID: dskRTSurveyorOracle,
		ManaCost: "{2}{U}", Power: 2, Toughness: 2,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Glimmer", me.ID); n != 1 {
		t.Errorf("%d Glimmer tokens, want 1", n)
	}
}

// --- Razorkin Hordecaller ---------------------------------------------------------

func TestDskRTRazorkinHordecallerMakesOneGremlinPerAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	caller := b12Push(g, me.ID, "Razorkin Hordecaller", "Creature — Human Clown Berserker", dskRTHordecallerOracle, 4, 4)
	if !hasKeywordOf(g, caller, "haste") {
		t.Error("Razorkin Hordecaller does not have haste")
	}
	a := b12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	// The Hordecaller stays home; two attackers are one attack, one Gremlin.
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Gremlin", me.ID); n != 1 {
		t.Errorf("%d Gremlin tokens, want 1", n)
	}
}

// --- Shrewd Storyteller -----------------------------------------------------------

func TestDskRTShrewdStorytellerCountersATargetOnlyIfTapped(t *testing.T) {
	for _, tapped := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		teller := pushCatalogPermanent(g, me.ID, "Shrewd Storyteller", "Creature — Human Survivor", dskRTStorytellerOracle, false)
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if tapped {
			dskB1Tap(g, teller)
		}
		advanceTo(t, g, game.StepPostcombatMain)
		if pick := latestPickTarget(g, me.ID); pick != nil {
			if !tapped {
				t.Fatal("untapped: the Survival trigger asked for a target")
			}
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		passPriorityAroundTable(t, g)
		want := 0
		if tapped {
			want = 1
		}
		if n := countersOn(g, theirs, game.CounterPlusOne); n != want {
			t.Errorf("tapped=%v: the target has %d +1/+1 counters, want %d", tapped, n, want)
		}
	}
}

// --- Savior of the Small ----------------------------------------------------------

func TestDskRTSaviorOfTheSmallReturnsACheapCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	savior := pushCatalogPermanent(g, me.ID, "Savior of the Small", "Creature — Kor Survivor", dskRTSaviorOracle, false)
	cheap := gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	big := gyCard(me, "Old Giant", "Creature — Giant", "{3}{R}", "")
	sorcery := gyCard(me, "Old Ritual", "Sorcery", "{B}", "")
	dskB1Tap(g, savior)
	advanceTo(t, g, game.StepPostcombatMain)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the Survival trigger did not ask for a target")
	}
	for _, bad := range []uuid.UUID{big, sorcery} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatal("a card that is not a creature card with mana value 3 or less was accepted as the target")
		}
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: cheap}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) {
		t.Error("the creature card did not return to the hand")
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the mana value 4 creature card left the graveyard")
	}
}

// --- Shroudstomper ----------------------------------------------------------------

// dskRTDrainedTimes checks that the Shroudstomper's controller gained 2,
// every opponent lost 2 and the controller drew one card per trigger.
func dskRTDrainedTimes(t *testing.T, g *game.Game, me *game.Player, handBefore, times int) {
	t.Helper()
	if want := dskRTStartingLife + times*dskRTShroudstomperDrain; me.Life != want {
		t.Errorf("your life %d, want %d", me.Life, want)
	}
	for _, p := range g.Seats {
		if want := dskRTStartingLife - times*dskRTShroudstomperDrain; p.ID != me.ID && p.Life != want {
			t.Errorf("an opponent's life %d, want %d", p.Life, want)
		}
	}
	if got := me.Hand.Size() - handBefore; got != times {
		t.Errorf("drew %d, want %d", got, times)
	}
}

func TestDskRTShroudstomperDrainsAndDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	stomper := castFromHand(t, g, game.Card{
		Name: "Shroudstomper", TypeLine: "Creature — Elemental", OracleID: dskRTShroudstomperOrcl,
		ManaCost: "{3}{W}{W}{B}{B}", Power: 5, Toughness: 5,
	}, game.CastSpellParams{})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	dskRTDrainedTimes(t, g, me, hand, 1)
	if !hasKeywordOf(g, stomper, "deathtouch") {
		t.Error("Shroudstomper does not have deathtouch")
	}
}

func TestDskRTShroudstomperDrainsAndDrawsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C")
	stomper := b12Push(g, me.ID, "Shroudstomper", "Creature — Elemental", dskRTShroudstomperOrcl, 5, 5)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, stomper)
	passPriorityAroundTable(t, g)
	dskRTDrainedTimes(t, g, me, hand, 1)
}

// --- Ripchain Razorkin ------------------------------------------------------------

func TestDskRTRipchainRazorkinSacrificesALandToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	razorkin := pushCatalogPermanent(g, me.ID, "Ripchain Razorkin", "Creature — Human Berserker", dskRTRipchainOracle, true)
	if !hasKeywordOf(g, razorkin, "reach") {
		t.Error("Ripchain Razorkin does not have reach")
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, razorkin, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no land to sacrifice")
	}
	land := b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	before := me.Hand.Size()
	// No {T} in the cost: a creature that just arrived may activate it.
	if err := g.ActivateCatalogAbility(me.ID, razorkin, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Seized from Slumber ----------------------------------------------------------

func TestDskRTSeizedFromSlumberCostsThreeLessAgainstATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { b16Tap(g, tapped) })
	advanceToMain(t, g)
	card := game.Card{Name: "Seized from Slumber", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: dskRTSeizedOracle}

	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(untapped)}); err == nil {
		t.Fatal("cast at the full price with two mana against an untapped creature")
	}
	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {1}{W} against a tapped creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature died")
	}

	if err := socB19CastWithPool(g, me, card, "W", 5, game.CastSpellParams{Targets: cardRefs(untapped)}); err != nil {
		t.Fatalf("cast at full price: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature survived the full-price cast")
	}
}
