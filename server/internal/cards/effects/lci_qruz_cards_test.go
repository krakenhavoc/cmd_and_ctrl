package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_qruz_cards_test.go — The Lost Caverns of Ixalan, cards Q, R and U–Z
// (#4470). Each test plays the real catalog entry and asserts what its
// text says.

const (
	lciQRUZWhirlpoolOracle   = "43fb0f35-8eea-4344-9415-6ecb8aa46491"
	lciQRUZSpiketailOracle   = "41f6d6d9-3fe5-44b0-904d-1cb90d36f2b6"
	lciQRUZRayOfRuinOracle   = "e36b78ed-4e3e-406d-85b3-fe0deb493418"
	lciQRUZPrairieOracle     = "c071257a-63e7-48d0-a677-0b396a09b624"
	lciQRUZRidgelineOracle   = "4c0f4a63-586a-4dde-9621-b0dd9118b2e5"
	lciQRUZVentsOracle       = "696e7ddb-bdc7-40ee-bc5c-59e98f4a7401"
	lciQRUZHeraldGuideOracle = "b9687fef-4cb2-4bf1-91ad-49dada07f4d6"
	lciQRUZHeraldScoutOracle = "021428b1-005c-40ed-9d47-fb5aa970598b"
	lciQRUZLurkerBatOracle   = "e6f379ad-96b1-46c3-bbc6-a8cb57156a7f"
	lciQRUZRockslideOracle   = "0909cbce-b4a0-42bb-8a97-47eeb99606c0"
	lciQRUZBoulderOracle     = "57f15861-82cf-4faf-827c-9bc7f4844d2a"
	lciQRUZUnluckyDropOracle = "98ebde8c-b47d-440b-93cb-a1407b807074"
	lciQRUZInquisitorOracle  = "03ee49c9-95a3-44f1-b010-f7c4b2f86a7c"
	lciQRUZWanderglyphOracle = "ebc3ca0f-35eb-4606-90ab-eb64556f81fe"
)

func TestLciQRUZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciQRUZWhirlpoolOracle:   "Quicksand Whirlpool",
		lciQRUZSpiketailOracle:   "Rampaging Spiketail",
		lciQRUZRayOfRuinOracle:   "Ray of Ruin",
		lciQRUZPrairieOracle:     "Restless Prairie",
		lciQRUZRidgelineOracle:   "Restless Ridgeline",
		lciQRUZVentsOracle:       "Restless Vents",
		lciQRUZHeraldGuideOracle: "River Herald Guide",
		lciQRUZHeraldScoutOracle: "River Herald Scout",
		lciQRUZLurkerBatOracle:   "Ruin-Lurker Bat",
		lciQRUZRockslideOracle:   "Rumbling Rockslide",
		lciQRUZBoulderOracle:     "Runaway Boulder",
		lciQRUZUnluckyDropOracle: "Unlucky Drop",
		lciQRUZInquisitorOracle:  "Vito's Inquisitor",
		lciQRUZWanderglyphOracle: "Volatile Wanderglyph",
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

// lciQRUZLandsControlledBy counts the lands a player controls.
func lciQRUZLandsControlledBy(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.IsLand() {
			n++
		}
	}
	return n
}

// lciQRUZRestless puts a Restless land on the battlefield under the
// active seat and animates it with its row-0 ability.
func lciQRUZRestless(t *testing.T, g *game.Game, name, oracle string, generic int, colours ...string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	land := s58p6Push(g, me.ID, name, "Land", oracle, "", 0, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	if c := s58p6View(t, g, land); c.IsCreature() {
		t.Fatalf("an idle %s is a creature", name)
	}
	ngheAnimate(t, g, me, land, generic, colours...)
	return land
}

// lciQRUZAssertAnimated checks a Restless land's animated characteristics.
func lciQRUZAssertAnimated(t *testing.T, g *game.Game, land uuid.UUID, subtype string, colours []string, power, toughness int) {
	t.Helper()
	c := s58p6View(t, g, land)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	if !slices.Contains(c.Effective().Subtypes, subtype) {
		t.Errorf("subtypes %v, want %s", c.Effective().Subtypes, subtype)
	}
	if !slices.Equal(c.EffectiveColors(), colours) {
		t.Errorf("colours %v, want %v", c.EffectiveColors(), colours)
	}
	if c.CurrentPower() != power || c.CurrentToughness() != toughness {
		t.Errorf("%d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), power, toughness)
	}
}

// --- Quicksand Whirlpool -----------------------------------------------------------

func TestLciQRUZQuicksandWhirlpoolCostsThreeLessAgainstATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	tapped := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { b16Tap(g, tapped) })
	advanceToMain(t, g)
	card := game.Card{Name: "Quicksand Whirlpool", TypeLine: "Instant", ManaCost: "{5}{W}", OracleID: lciQRUZWhirlpoolOracle}
	if err := socB19CastWithPool(g, me, card, "W", 3, game.CastSpellParams{Targets: cardRefs(untapped)}); err == nil {
		t.Fatal("cast with three mana against an untapped creature")
	}
	card.InstanceID = uuid.New()
	if err := socB19CastWithPool(g, me, card, "W", 3, game.CastSpellParams{Targets: cardRefs(tapped)}); err != nil {
		t.Fatalf("cast for {2}{W} against a tapped creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !g.Exile.Contains(tapped) {
		t.Error("the tapped creature was not exiled")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature left the battlefield")
	}
}

func TestLciQRUZQuicksandWhirlpoolAtFullPriceExilesAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	advanceToMain(t, g)
	card := game.Card{Name: "Quicksand Whirlpool", TypeLine: "Instant", ManaCost: "{5}{W}", OracleID: lciQRUZWhirlpoolOracle}
	if err := socB19CastWithPool(g, me, card, "W", 6, game.CastSpellParams{Targets: cardRefs(ox)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ox) {
		t.Error("the creature was not exiled")
	}
}

// --- Rampaging Spiketail -----------------------------------------------------------

func TestLciQRUZRampagingSpiketailPumpsAndShieldsACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spiketail := castHoldCreature(t, g, "Rampaging Spiketail", "Creature — Dinosaur", lciQRUZSpiketailOracle, 5, 6)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no enters-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, theirs) || !hasID(pick.PickTargetCards, bear) || !hasID(pick.PickTargetCards, spiketail) {
		t.Errorf("offered %v; want the creatures you control only", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the targeted bear is %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the targeted bear did not gain indestructible")
	}
	if hasEffectiveKeyword(t, g, spiketail, "indestructible") {
		t.Error("the Spiketail itself gained indestructible")
	}
	endTurn(t, g)
	if p := effectivePower(t, g, bear); p != 2 || hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Errorf("the effect outlasted the turn: power %d", p)
	}
}

func TestLciQRUZRampagingSpiketailSwampcyclesForASwamp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ids := seedSearchLibrary(me,
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	card := pushCatalogHandCard(me, "Rampaging Spiketail", "Creature — Dinosaur", lciQRUZSpiketailOracle)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("swampcycle: %v", err)
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the Spiketail was not discarded to swampcycle")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search offers %v, want only the Swamp", c.SearchCards)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Swamp did not go to the hand")
	}
}

// --- Ray of Ruin -------------------------------------------------------------------

func TestLciQRUZRayOfRuinTargetsACreatureVehicleOrNonbasicLand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	basic := b12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	if err := castCatalogSpellErr(t, g, "Ray of Ruin", "Sorcery", lciQRUZRayOfRuinOracle, cardRefs(basic)); err == nil {
		t.Error("a basic land was accepted as the target")
	}
	for _, tc := range []struct {
		name, typeLine string
		power          int
	}{
		{"a creature", "Creature — Bear", 2},
		{"a noncreature Vehicle", "Artifact — Vehicle", 0},
		{"a nonbasic land", "Land", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			var target uuid.UUID
			if tc.power > 0 {
				target = b12Creature(g, opp.ID, "Their Thing", tc.typeLine, tc.power, tc.power)
			} else {
				target = b12Permanent(g, opp.ID, "Their Thing", tc.typeLine)
			}
			seedLibrary(me, "Top")
			castCatalogSpell(t, g, "Ray of Ruin", "Sorcery", lciQRUZRayOfRuinOracle, cardRefs(target))
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(target) || !g.Exile.Contains(target) {
				t.Error("the target was not exiled")
			}
			sc := scryChoiceFor(g, me.ID)
			if sc == nil {
				t.Fatal("no scry 1 after the exile")
			}
			if len(sc.ScryCards) != 1 {
				t.Errorf("scry looks at %d cards, want 1", len(sc.ScryCards))
			}
		})
	}
}

// --- Restless Prairie --------------------------------------------------------------

func TestLciQRUZRestlessPrairieEntersTappedAndBecomesALlama(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Restless Prairie", "Land", lciQRUZPrairieOracle, game.CastSpellParams{})
	if !isTapped(g, id) {
		t.Error("Restless Prairie entered untapped")
	}
	g2 := newCatalogGame(t)
	prairie := lciQRUZRestless(t, g2, "Restless Prairie", lciQRUZPrairieOracle, 2, "G", "W")
	lciQRUZAssertAnimated(t, g2, prairie, "Llama", []string{"G", "W"}, 3, 3)
}

func TestLciQRUZRestlessPrairieAttackPumpsOtherCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	prairie := lciQRUZRestless(t, g, "Restless Prairie", lciQRUZPrairieOracle, 2, "G", "W")
	declareAttack(t, g, opp.ID, prairie)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("your other creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePower(t, g, prairie), effectiveToughness(t, g, prairie); p != 3 || tough != 3 {
		t.Errorf("the Prairie is %d/%d, want 3/3 (it is not an other creature)", p, tough)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

// --- Restless Ridgeline ------------------------------------------------------------

func TestLciQRUZRestlessRidgelineBecomesADinosaur(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Restless Ridgeline", "Land", lciQRUZRidgelineOracle, game.CastSpellParams{})
	if !isTapped(g, id) {
		t.Error("Restless Ridgeline entered untapped")
	}
	g2 := newCatalogGame(t)
	ridge := lciQRUZRestless(t, g2, "Restless Ridgeline", lciQRUZRidgelineOracle, 2, "R", "G")
	lciQRUZAssertAnimated(t, g2, ridge, "Dinosaur", []string{"R", "G"}, 3, 4)
}

func TestLciQRUZRestlessRidgelinePumpsAndUntapsAnotherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "My Idle Bear", "Creature — Bear", 2, 2)
	ridge := lciQRUZRestless(t, g, "Restless Ridgeline", lciQRUZRidgelineOracle, 2, "R", "G")
	declareAttack(t, g, opp.ID, ridge, bear)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no attack-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, ridge) || hasID(pick.PickTargetCards, idle) || !hasID(pick.PickTargetCards, bear) {
		t.Errorf("offered %v; want another attacking creature only", pick.PickTargetCards)
	}
	if !tappedNow(g, bear) {
		t.Fatal("the attacking bear is not tapped")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
	if tappedNow(g, bear) {
		t.Error("the target was not untapped")
	}
	if !tappedNow(g, ridge) {
		t.Error("the Ridgeline itself untapped")
	}
}

// --- Restless Vents ----------------------------------------------------------------

func TestLciQRUZRestlessVentsBecomesAnInsectWithMenace(t *testing.T) {
	g := newCatalogGame(t)
	id := b12PlayFromHand(t, g, "Restless Vents", "Land", lciQRUZVentsOracle, game.CastSpellParams{})
	if !isTapped(g, id) {
		t.Error("Restless Vents entered untapped")
	}
	g2 := newCatalogGame(t)
	vents := lciQRUZRestless(t, g2, "Restless Vents", lciQRUZVentsOracle, 1, "B", "R")
	lciQRUZAssertAnimated(t, g2, vents, "Insect", []string{"B", "R"}, 2, 3)
	if !hasEffectiveKeyword(t, g2, vents, "menace") {
		t.Error("the animated Vents has no menace")
	}
	endTurn(t, g2)
	if c := s58p6View(t, g2, vents); c.IsCreature() || hasEffectiveKeyword(t, g2, vents, "menace") {
		t.Error("the animation outlasted the turn")
	}
}

func TestLciQRUZRestlessVentsAttackMayRummage(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		seedLibrary(me, "Drawn")
		discard := pushHandCard(g, me)
		vents := lciQRUZRestless(t, g, "Restless Vents", lciQRUZVentsOracle, 1, "B", "R")
		hand, library := me.Hand.Size(), me.Library.Size()
		declareAttack(t, g, opp.ID, vents)
		passPriorityAroundTable(t, g)
		if accept {
			answerDiscard(t, g, me.ID, discard)
		} else {
			answerDiscard(t, g, me.ID)
		}
		passPriorityAroundTable(t, g)
		wantDrawn := 0
		if accept {
			wantDrawn = 1
		}
		if got := library - me.Library.Size(); got != wantDrawn {
			t.Errorf("accept=%v: drew %d, want %d", accept, got, wantDrawn)
		}
		if me.Hand.Size() != hand {
			t.Errorf("accept=%v: hand %d -> %d, want unchanged", accept, hand, me.Hand.Size())
		}
		if accept && !me.Graveyard.Contains(discard) {
			t.Error("the chosen card was not discarded")
		}
	}
}

// --- River Herald Guide and Scout ----------------------------------------------------

func TestLciQRUZRiverHeraldGuideHasVigilanceAndExploresALandToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	top := putOnTopForTest(me, "Forest", "Basic Land — Forest")
	guide := msc17Play(t, g, me, game.Card{Name: "River Herald Guide", TypeLine: "Creature — Merfolk Scout", OracleID: lciQRUZHeraldGuideOracle, Power: 3, Toughness: 1})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the revealed land did not go to the hand")
	}
	if got := countersOn(g, guide, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters after revealing a land, want 0", got)
	}
	if !hasEffectiveKeyword(t, g, guide, "vigilance") {
		t.Error("River Herald Guide has no vigilance")
	}
}

func TestLciQRUZRiverHeraldScoutExploresANonlandForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	top := putOnTopForTest(me, "Some Sorcery", "Sorcery")
	scout := msc17Play(t, g, me, game.Card{Name: "River Herald Scout", TypeLine: "Creature — Merfolk Scout", OracleID: lciQRUZHeraldScoutOracle, Power: 1, Toughness: 2})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, scout, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	answerExploreQuestion(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) || me.Library.Contains(top) {
		t.Error("the revealed card did not go to the graveyard when the player chose to")
	}
}

// --- Ruin-Lurker Bat ---------------------------------------------------------------

func TestLciQRUZRuinLurkerBatScriesAtYourEndStepIfYouDescended(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		scry     bool
	}{
		{"a permanent card milled", "Artifact", true},
		{"only an instant card milled", "Instant", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			bat := b12Push(g, me.ID, "Ruin-Lurker Bat", "Creature — Bat", lciQRUZLurkerBatOracle, 1, 1)
			if !hasEffectiveKeyword(t, g, bat, "flying") || !hasEffectiveKeyword(t, g, bat, "lifelink") {
				t.Error("Ruin-Lurker Bat is missing flying or lifelink")
			}
			lciDGMillOne(t, g, me, tc.typeLine)
			seedLibrary(me, "Top")
			advanceToEndStepOf(t, g, 0)
			if !tc.scry {
				if triggersOnStackFrom(g, bat) != 0 {
					t.Fatal("the end-step trigger went on the stack without a descent")
				}
				return
			}
			passPriorityAroundTable(t, g)
			sc := scryChoiceFor(g, me.ID)
			if sc == nil {
				t.Fatal("no scry 1 after a descent")
			}
			if len(sc.ScryCards) != 1 {
				t.Errorf("scry looks at %d cards, want 1", len(sc.ScryCards))
			}
		})
	}
}

// --- Rumbling Rockslide ------------------------------------------------------------

func TestLciQRUZRumblingRockslideDealsDamageEqualToYourLands(t *testing.T) {
	for _, tc := range []struct {
		name  string
		extra int // toughness above the land count
		dies  bool
	}{
		{"toughness equal to your lands dies", 0, true},
		{"one more toughness survives", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			for i := 0; i < 3; i++ {
				b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
			}
			// An opponent's lands are not yours: these would kill the
			// survivor if they were counted.
			for i := 0; i < 4; i++ {
				b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
			}
			lands := lciQRUZLandsControlledBy(g, me.ID)
			target := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 1, lands+tc.extra)
			castCatalogSpell(t, g, "Rumbling Rockslide", "Sorcery", lciQRUZRockslideOracle, cardRefs(target))
			passPriorityAroundTable(t, g)
			if died := !g.Battlefield.Contains(target); died != tc.dies {
				t.Errorf("with %d lands against toughness %d: died=%v, want %v", lands, lands+tc.extra, died, tc.dies)
			}
		})
	}
}

// --- Runaway Boulder ---------------------------------------------------------------

func TestLciQRUZRunawayBoulderDealsSixToACreatureAnOpponentControls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 6, 6)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)
	castHoldCreature(t, g, "Runaway Boulder", "Artifact", lciQRUZBoulderOracle, 0, 0)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no enters-trigger target prompt")
	}
	if hasID(pick.PickTargetCards, mine) || !hasID(pick.PickTargetCards, theirs) {
		t.Errorf("offered %v; want only a creature an opponent controls", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 6/6 survived 6 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature died")
	}
}

func TestLciQRUZRunawayBoulderHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepBeginCombat)
	// Cartographer's Companion, an artifact without flash, is refused here.
	slow := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: slow, Name: "Cartographer's Companion", TypeLine: "Artifact Creature — Gnome",
		OracleID: "460c0f0b-bbf4-4307-9b44-350175d83bfe", Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, slow, game.CastSpellParams{}); err == nil {
		t.Fatal("an artifact without flash was cast at the beginning of combat")
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Runaway Boulder", TypeLine: "Artifact", OracleID: lciQRUZBoulderOracle,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast at the beginning of combat: %v", err)
	}
}

func TestLciQRUZRunawayBoulderCycles(t *testing.T) {
	dftSTCycles(t, "Runaway Boulder", "Artifact", lciQRUZBoulderOracle, 0)
}

// --- Unlucky Drop ------------------------------------------------------------------

func TestLciQRUZUnluckyDropLetsTheOwnerChooseTopOrBottom(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bottom bool
	}{{"top", false}, {"bottom", true}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rock := pushArtifactForTest(g, opp.ID, "Their Mind Stone")
			castCatalogSpell(t, g, "Unlucky Drop", "Instant", lciQRUZUnluckyDropOracle, cardRefs(rock))
			passPriorityAroundTable(t, g)
			if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
				t.Fatal("the caster was asked; the owner chooses")
			}
			c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
			if c == nil {
				t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
			}
			var top, bottom []uuid.UUID
			if tc.bottom {
				bottom = []uuid.UUID{rock}
			} else {
				top = []uuid.UUID{rock}
			}
			if err := g.ResolvePutInLibrary(c.ID, opp.ID, bottom, top); err != nil {
				t.Fatalf("ResolvePutInLibrary: %v", err)
			}
			if g.Battlefield.Contains(rock) {
				t.Fatal("the artifact is still on the battlefield")
			}
			cards := opp.Library.Cards
			// The bottom of a library is Cards[0], the top the last entry.
			if tc.bottom && cards[0].InstanceID != rock {
				t.Errorf("the bottom card is %q, want the artifact", cards[0].Name)
			}
			if !tc.bottom && cards[len(cards)-1].InstanceID != rock {
				t.Errorf("the top card is %q, want the artifact", cards[len(cards)-1].Name)
			}
		})
	}
}

func TestLciQRUZUnluckyDropCantTargetALand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	if err := castCatalogSpellErr(t, g, "Unlucky Drop", "Instant", lciQRUZUnluckyDropOracle, cardRefs(land)); err == nil {
		t.Error("a land was accepted as the target")
	}
}

// --- Vito's Inquisitor -------------------------------------------------------------

func TestLciQRUZVitosInquisitorSacrificesAnotherForACounterAndMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	inq := b12Push(g, me.ID, "Vito's Inquisitor", "Creature — Vampire Knight", lciQRUZInquisitorOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Mana Rock", "Artifact")
	shrine := b12Permanent(g, me.ID, "Shrine", "Enchantment")
	advanceToMain(t, g)
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, inq, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{inq}}); err == nil {
		t.Fatal("the Inquisitor sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, inq, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{shrine}}); err == nil {
		t.Fatal("an enchantment is neither a creature nor an artifact")
	}
	b16Activate(t, g, me.ID, inq, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := countersOn(g, inq, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, inq, "menace") {
		t.Error("the Inquisitor did not gain menace")
	}
	fillPoolColored(me, "B", 1)
	b16Activate(t, g, me.ID, inq, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if got := countersOn(g, inq, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters after sacrificing an artifact too, want 2", got)
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, inq, "menace") {
		t.Error("menace outlasted the turn")
	}
	if got := countersOn(g, inq, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters after the turn, want 2 (counters stay)", got)
	}
}

// --- Volatile Wanderglyph ----------------------------------------------------------

func TestLciQRUZVolatileWanderglyphMayRummageWhenItBecomesTapped(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		seedLibrary(me, "Drawn")
		discard := pushHandCard(g, me)
		glyph := b12Push(g, me.ID, "Volatile Wanderglyph", "Artifact Creature — Golem", lciQRUZWanderglyphOracle, 2, 2)
		advanceToMain(t, g)
		if discardChoiceFor(g, me.ID) != nil {
			t.Fatal("a rummage is offered before the Wanderglyph has tapped")
		}
		hand, library := me.Hand.Size(), me.Library.Size()
		declareAttack(t, g, opp.ID, glyph)
		passPriorityAroundTable(t, g)
		if accept {
			answerDiscard(t, g, me.ID, discard)
		} else {
			answerDiscard(t, g, me.ID)
		}
		passPriorityAroundTable(t, g)
		wantDrawn := 0
		if accept {
			wantDrawn = 1
		}
		if got := library - me.Library.Size(); got != wantDrawn {
			t.Errorf("accept=%v: drew %d, want %d", accept, got, wantDrawn)
		}
		if me.Hand.Size() != hand {
			t.Errorf("accept=%v: hand %d -> %d, want unchanged", accept, hand, me.Hand.Size())
		}
	}
}
