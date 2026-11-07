package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dsk_hl_cards_test.go — Duskmourn: House of Horror, cards H–L and the
// D–G cards the D–G batch left (#4466). Each test plays the real catalog
// entry and asserts what its text says.

const (
	dskHLFailedTestsOracle  = "34fd461b-a913-4a60-9477-ce2bd5c6fc5a"
	dskHLImmobilityOracle   = "fb2600a8-0ba6-4888-ac8a-f206469d4c71"
	dskHLDarkOracle         = "9285cdf1-ad25-4aee-a865-1bf0cc905183"
	dskHLBurrowerOracle     = "04c4250e-014a-4067-80e8-4134069354f7"
	dskHLGlimmerburstOracle = "a26820b9-3bec-41b2-9812-99e4a2211297"
	dskHLGlimmerlightOracle = "8fd88b98-9899-4bc0-a4b7-7ef6fba36dfd"
	dskHLSeekerOracle       = "cbb5526b-3aa1-43d9-acdb-bd230ce3fe63"
	dskHLLongneckOracle     = "bc509d30-ceca-47ef-a574-62c021812014"
	dskHLEscortOracle       = "2c6ef9e8-6b60-436c-8391-ae0baabb3a1b"
	dskHLVigorOracle        = "bef3f0a9-4d5c-4469-81df-adbeedeed768"
	dskHLShackOracle        = "c56cd2ec-5907-4282-9162-d93b7dfd63b5"
	dskHLHopeOracle         = "c7ee79c3-a273-45b4-b3a4-548d9b2b883c"
	dskHLLionheartOracle    = "bd63bfda-e20a-4476-8a61-cdb51355a227"
	dskHLLiveOrDieOracle    = "dc367b23-d1f4-4c5c-b300-a44a726e3f74"
	dskHLPhoneOracle        = "da554d71-1f13-49ae-9eb2-c778a23df08b"

	dskHLNightmareTypeLine = "Enchantment Creature — Nightmare"
)

func TestDskHLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dskHLFailedTestsOracle:  "Fear of Failed Tests",
		dskHLImmobilityOracle:   "Fear of Immobility",
		dskHLDarkOracle:         "Fear of the Dark",
		dskHLBurrowerOracle:     "Flesh Burrower",
		dskHLGlimmerburstOracle: "Glimmerburst",
		dskHLGlimmerlightOracle: "Glimmerlight",
		dskHLSeekerOracle:       "Glimmer Seeker",
		dskHLLongneckOracle:     "Grasping Longneck",
		dskHLEscortOracle:       "Hardened Escort",
		dskHLVigorOracle:        "Horrid Vigor",
		dskHLShackOracle:        "Lakeside Shack",
		dskHLHopeOracle:         "Leyline of Hope",
		dskHLLionheartOracle:    "Lionheart Glimmer",
		dskHLLiveOrDieOracle:    "Live or Die",
		dskHLPhoneOracle:        "Living Phone",
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

// dskHLTarget is a single card target.
func dskHLTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// dskHLSurvival puts a catalog creature onto the active seat's
// battlefield, taps it when asked, and runs the table to the second main
// phase's Survival trigger.
func dskHLSurvival(t *testing.T, g *game.Game, name, typeLine, oracle string, tapped bool) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	if tapped {
		dskB1Tap(g, id)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
}

// --- Fear of Failed Tests ------------------------------------------------------

func TestDskHLFearOfFailedTestsDrawsAsManyAsTheCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fear := b12Push(g, me.ID, "Fear of Failed Tests", dskHLNightmareTypeLine, dskHLFailedTestsOracle, 2, 7)
	before := me.Hand.Size()
	attackWith(t, g, opp.ID, fear)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d cards for 2 combat damage, want 2", got)
	}
}

// --- Fear of Immobility -----------------------------------------------------------

func TestDskHLFearOfImmobilityStunsOnlyAnOpponentsCreature(t *testing.T) {
	for _, theirs := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		owner := me.ID
		if theirs {
			owner = opp.ID
		}
		victim := b12Creature(g, owner, "Some Bear", "Creature — Bear", 2, 2)
		castFromHand(t, g, game.Card{
			Name: "Fear of Immobility", TypeLine: dskHLNightmareTypeLine, OracleID: dskHLImmobilityOracle,
			ManaCost: "{4}{W}", Power: 4, Toughness: 4,
		}, game.CastSpellParams{})
		passPriorityAroundTable(t, g)
		dskMPPickIfAsked(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim})
		passPriorityAroundTable(t, g)
		if !b13Tapped(t, g, victim) {
			t.Errorf("theirs=%v: the target was not tapped", theirs)
		}
		want := 0
		if theirs {
			want = 1
		}
		if n := countersOn(g, victim, game.CounterStun); n != want {
			t.Errorf("theirs=%v: %d stun counters, want %d", theirs, n, want)
		}
	}
}

// --- Fear of the Dark -----------------------------------------------------------------

func TestDskHLFearOfTheDarkGainsMenaceAndDeathtouchUnlessTheDefenderHasAGlimmer(t *testing.T) {
	for _, glimmer := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		fear := b12Push(g, me.ID, "Fear of the Dark", dskHLNightmareTypeLine, dskHLDarkOracle, 5, 5)
		// Someone else's Glimmer does not count: only the defender's.
		b12Creature(g, g.Seats[2].ID, "Bystander Glimmer", "Enchantment Creature — Glimmer", 1, 1)
		if glimmer {
			b12Creature(g, opp.ID, "Their Glimmer", "Enchantment Creature — Glimmer", 1, 1)
		}
		declareAttack(t, g, opp.ID, fear)
		passPriorityAroundTable(t, g)
		for _, kw := range []string{"menace", "deathtouch"} {
			if got := hasKeywordOf(g, fear, kw); got == glimmer {
				t.Errorf("defender has a Glimmer=%v: %s is %v", glimmer, kw, got)
			}
		}
	}
}

// --- Flesh Burrower, Hardened Escort ----------------------------------------------------

func TestDskHLAttackTriggersGrantAnotherCreatureYouControl(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keyword                string
		bonus                  int
	}{
		{"Flesh Burrower", "Creature — Insect", dskHLBurrowerOracle, 2, 2, "deathtouch", 0},
		{"Hardened Escort", "Creature — Human Soldier", dskHLEscortOracle, 2, 4, "indestructible", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			bear := b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
			b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			declareAttack(t, g, opp.ID, src)
			pick := pickTargetPrompt(t, g)
			if slices.Contains(pick.PickTargetCards, src) {
				t.Error("the source offered itself: the clause says another creature")
			}
			answerPickTarget(t, g, bear)
			passPriorityAroundTable(t, g)
			if !hasKeywordOf(g, bear, tc.keyword) {
				t.Errorf("the Bear, staying home, did not gain %s", tc.keyword)
			}
			if p, _ := effectivePTOf(t, g, bear); p != 2+tc.bonus {
				t.Errorf("the Bear has power %d, want %d", p, 2+tc.bonus)
			}
		})
	}
	g := newCatalogGame(t)
	burrower := b12Push(g, g.Seats[0].ID, "Flesh Burrower", "Creature — Insect", dskHLBurrowerOracle, 2, 2)
	if !hasKeywordOf(g, burrower, "deathtouch") {
		t.Error("Flesh Burrower does not have deathtouch")
	}
}

// --- Glimmerburst ------------------------------------------------------------------------

func TestDskHLGlimmerburstDrawsTwoAndMakesAGlimmer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	if err := dskB1CastAtDrawStep(t, g, "Glimmerburst", "Instant", dskHLGlimmerburstOracle); err != nil {
		t.Fatalf("cast at instant speed: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if n := tokensNamedControlledBy(g, "Glimmer", me.ID); n != 1 {
		t.Errorf("%d Glimmer tokens, want 1", n)
	}
}

// --- Glimmerlight -----------------------------------------------------------------------

func TestDskHLGlimmerlightMakesAGlimmerAndEquipsForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	light := castFromHand(t, g, game.Card{
		Name: "Glimmerlight", TypeLine: "Artifact — Equipment", OracleID: dskHLGlimmerlightOracle, ManaCost: "{2}",
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Glimmer", me.ID); n != 1 {
		t.Fatalf("%d Glimmer tokens, want 1", n)
	}
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := hobB4Activate(t, g, me.ID, light, 0, "{C}", dskHLTarget(bear)); err != nil {
		t.Fatalf("equip for {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tough)
	}
}

// --- Glimmer Seeker ---------------------------------------------------------------------

func TestDskHLGlimmerSeekerMakesAGlimmerOrDrawsWithOne(t *testing.T) {
	const typeLine = "Creature — Human Survivor"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	dskHLSurvival(t, g, "Glimmer Seeker", typeLine, dskHLSeekerOracle, true)
	if n := tokensNamedControlledBy(g, "Glimmer", me.ID); n != 1 {
		t.Errorf("no Glimmer: %d Glimmer tokens, want 1", n)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("no Glimmer: hand %d, want %d (no draw)", got, hand)
	}

	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "My Glimmer", "Enchantment Creature — Glimmer", 1, 1)
	hand = me.Hand.Size()
	dskHLSurvival(t, g, "Glimmer Seeker", typeLine, dskHLSeekerOracle, true)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("with a Glimmer: hand %d, want %d", got, hand+1)
	}
	if n := tokensNamedControlledBy(g, "Glimmer", me.ID); n != 0 {
		t.Errorf("with a Glimmer: %d Glimmer tokens, want 0", n)
	}

	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	dskHLSurvival(t, g, "Glimmer Seeker", typeLine, dskHLSeekerOracle, false)
	if n := tokensNamedControlledBy(g, "Glimmer", me.ID); n != 0 {
		t.Errorf("untapped: %d Glimmer tokens, want 0", n)
	}
}

// --- Grasping Longneck, Leyline of Hope ----------------------------------------------------

func TestDskHLGraspingLongneckGainsTwoWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	neck := b12Push(g, me.ID, "Grasping Longneck", "Enchantment Creature — Horror", dskHLLongneckOracle, 4, 2)
	if !hasKeywordOf(g, neck, "reach") {
		t.Error("Grasping Longneck does not have reach")
	}
	life := me.Life
	killCreature(t, g, me.ID, neck)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestDskHLLeylineOfHopeAddsOneToEachGainAndPumpsAboveTheLine(t *testing.T) {
	spec, _ := Lookup(dskHLHopeOracle)
	if spec.OpeningHand == nil {
		t.Error("Leyline of Hope has no opening-hand action")
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Leyline of Hope", "Enchantment", dskHLHopeOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	start := me.Life

	neck := b12Push(g, me.ID, "Grasping Longneck", "Enchantment Creature — Horror", dskHLLongneckOracle, 4, 2)
	killCreature(t, g, me.ID, neck)
	passPriorityAroundTable(t, g)
	if me.Life != start+3 {
		t.Fatalf("gaining 2: life %d, want %d", me.Life, start+3)
	}
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("3 above the start: the Bear has power %d, want 2", p)
	}

	// Gaining 3 more is 4 with the Leyline: 7 above the start.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	if me.Life != start+7 {
		t.Fatalf("life %d, want %d", me.Life, start+7)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("7 above the start: your Bear is %d/%d, want 4/4", p, tough)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Bear has power %d, want 2", p)
	}
	if opp.Life != start {
		t.Errorf("opponent life %d, want %d (no gain)", opp.Life, start)
	}
}

// --- Horrid Vigor -------------------------------------------------------------------------

func TestDskHLHorridVigorGrantsDeathtouchAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// The table starts in the draw step: an instant may be cast there.
	id := pushCatalogHandCard(me, "Horrid Vigor", "Instant", dskHLVigorOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: dskHLTarget(bear)}); err != nil {
		t.Fatalf("cast in the draw step: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"deathtouch", "indestructible"} {
		if !hasKeywordOf(g, bear, kw) {
			t.Errorf("the Bear did not gain %s", kw)
		}
	}
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want an unpumped 2/2", p, tough)
	}
}

// --- Lakeside Shack ---------------------------------------------------------------------------

func TestDskHLLakesideShackEntersTappedUnlessAPlayerIsAtThirteen(t *testing.T) {
	g := newCatalogGame(t)
	land := b12PlayFromHand(t, g, "Lakeside Shack", "Land", dskHLShackOracle, game.CastSpellParams{})
	if !b13Tapped(t, g, land) {
		t.Error("everyone at 40: the land entered untapped")
	}
	if got := ltProduced(t, g, land); !slices.Equal(got, []string{"{G|U}"}) {
		t.Errorf("the land adds %v, want {G|U}", got)
	}
	g.Seats[3].Life = 13
	if b13Tapped(t, g, b12PlayFromHand(t, g, "Lakeside Shack", "Land", dskHLShackOracle, game.CastSpellParams{})) {
		t.Error("an opponent at 13: the land entered tapped")
	}
}

// --- Lionheart Glimmer ----------------------------------------------------------------------

func TestDskHLLionheartGlimmerPumpsYourTeamWhenYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lion := b12Push(g, me.ID, "Lionheart Glimmer", "Enchantment Creature — Cat Glimmer", dskHLLionheartOracle, 2, 5)
	attacker := b12Creature(g, me.ID, "Attacker A", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Attacker B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// The Glimmer stays home; two attackers are one attack, one trigger.
	declareAttack(t, g, opp.ID, attacker, other)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{attacker, other} {
		if p, tough := effectivePTOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("an attacker is %d/%d, want 3/3 (one trigger)", p, tough)
		}
	}
	if p, tough := effectivePTOf(t, g, lion); p != 3 || tough != 6 {
		t.Errorf("the Glimmer at home is %d/%d, want 3/6", p, tough)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

// --- Live or Die -------------------------------------------------------------------------------

func TestDskHLLiveOrDieReturnsOrDestroys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	castFromHand(t, g, game.Card{Name: "Live or Die", TypeLine: "Instant", OracleID: dskHLLiveOrDieOracle, ManaCost: "{3}{B}{B}"},
		game.CastSpellParams{Modes: []int{0}, Targets: dskHLTarget(dead)})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Error("the first mode did not return the creature card")
	}

	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Live or Die", TypeLine: "Instant", OracleID: dskHLLiveOrDieOracle, ManaCost: "{3}{B}{B}"},
		game.CastSpellParams{Modes: []int{1}, Targets: dskHLTarget(victim)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the second mode did not destroy the creature")
	}
}

func TestDskHLLiveOrDieReturnsOnlyFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := gyCard(opp, "Their Dead Bear", "Creature — Bear", "{1}{G}", "")
	advanceToMain(t, g)
	if _, err := castFromHandErr(g, g.Seats[0], "Live or Die", "Instant", "{3}{B}{B}", dskHLLiveOrDieOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: dskHLTarget(theirs)}); err == nil {
		t.Error("the return mode targeted an opponent's graveyard")
	}
}

// --- Living Phone --------------------------------------------------------------------------------

func TestDskHLLivingPhoneTakesASmallCreatureFromTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := uuid.New()
	big := uuid.New()
	for _, c := range []game.Card{
		{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"},
		{InstanceID: big, Name: "Big Beast", TypeLine: "Creature — Beast", Power: 5, Toughness: 5},
		{InstanceID: small, Name: "Small Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant"},
		{InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact"},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	phone := b12Push(g, me.ID, "Living Phone", "Artifact Creature — Toy", dskHLPhoneOracle, 2, 1)
	killCreature(t, g, me.ID, phone)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("the dies trigger raised no prompt")
	}
	if !slices.Equal(prompt.ChooseCards, []uuid.UUID{small}) {
		t.Fatalf("offered %v, want only the power-2 creature card", prompt.ChooseCards)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{small}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(small) {
		t.Error("the chosen card is not in hand")
	}
	if !me.Library.Contains(big) {
		t.Error("the big creature left the library")
	}
}
