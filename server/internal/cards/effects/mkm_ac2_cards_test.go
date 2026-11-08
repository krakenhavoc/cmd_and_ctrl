package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_ac2_cards_test.go — Murders at Karlov Manor A–C, batch 2 (#4469).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mkmAC2AlquistOracle      = "dd79a994-7dad-49d4-8d8f-62dbedde4a38"
	mkmAC2AudienceOracle     = "cbbb5c0c-e953-410d-88c6-8fedb719e94c"
	mkmAC2BuriedOracle       = "f1f0eab3-681b-4c2d-b590-555951920880"
	mkmAC2WitnessOracle      = "9038a457-ce01-4f6d-9e2b-5220f08b4290"
	mkmAC2CandlestickOracle  = "2f7b46c1-3d82-4fd9-9882-589315cab47b"
	mkmAC2GatewayOracle      = "42a9aeac-5b9c-4675-8c96-7c8ce7e9098c"
	mkmAC2RansackedLabOracle = "f85eea8f-5887-4921-a943-d15e0ad1e028"
	mkmAC2TrampledOracle     = "575a0dc2-206a-406a-90b3-72a92f2ea87a"
	mkmAC2ChalkOracle        = "af335eb3-f60a-44b3-a1ed-1e165f3e3605"
	mkmAC2CoercedOracle      = "3a007dbd-18fd-44d1-a305-7cc8dd6231eb"
	mkmAC2CadaverOracle      = "cc5370e0-a0e6-4958-a6f2-093a3868b575"
	mkmAC2InquiryOracle      = "52fe8101-82af-4aed-a395-54e7b7639bd2"

	mkmAC2CaseTypeLine = "Enchantment — Case"
)

func TestMkmAC2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmAC2AlquistOracle:      "Alquist Proft, Master Sleuth",
		mkmAC2AudienceOracle:     "Audience with Trostani",
		mkmAC2BuriedOracle:       "Buried in the Garden",
		mkmAC2WitnessOracle:      "Call a Surprise Witness",
		mkmAC2CandlestickOracle:  "Candlestick",
		mkmAC2GatewayOracle:      "Case of the Gateway Express",
		mkmAC2RansackedLabOracle: "Case of the Ransacked Lab",
		mkmAC2TrampledOracle:     "Case of the Trampled Garden",
		mkmAC2ChalkOracle:        "Chalk Outline",
		mkmAC2CoercedOracle:      "Coerced to Kill",
		mkmAC2CadaverOracle:      "Curious Cadaver",
		mkmAC2InquiryOracle:      "Curious Inquiry",
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

// mkmAC2PushGraveyard puts a catalog card with a mana cost into a
// player's graveyard.
func mkmAC2PushGraveyard(g *game.Game, p *game.Player, name, typeLine, oracle, manaCost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
			Owner: p.ID, Controller: p.ID,
		})
	})
	return id
}

// mkmAC2AddMana puts mana into a player's pool.
func mkmAC2AddMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	})
}

// mkmAC2Solve solves a Case directly.
func mkmAC2Solve(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
	})
}

// mkmAC2RunUntilSolved walks the game forward until the Case solves,
// failing if it never does.
func mkmAC2RunUntilSolved(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 400 && !mkmB1Solved(g, id); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !mkmB1Solved(g, id) {
		t.Fatal("the Case never solved")
	}
}

// mkmAC2EndStepSolves reports whether the Case is solved once the
// active player's end step trigger has resolved.
func mkmAC2EndStepSolves(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	return mkmB1Solved(g, id)
}

// --- Alquist Proft, Master Sleuth ------------------------------------------------

func TestMkmAC2AlquistProftInvestigatesOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castHoldCreature(t, g, "Alquist Proft, Master Sleuth", "Legendary Creature — Human Detective", mkmAC2AlquistOracle, 3, 3)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, id, "vigilance")
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues on entering, want 1", got)
	}
}

func TestMkmAC2AlquistProftSacrificesAClueToDrawXAndGainX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	auraSeedLibrary(me, 4)
	advanceToMain(t, g)
	alquist := b12Push(g, me.ID, "Alquist Proft, Master Sleuth", "Legendary Creature — Human Detective", mkmAC2AlquistOracle, 3, 3)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	clue := pushToken(g, me.ID, ClueToken())

	mkmAC2AddMana(t, g, me, "{C}{C}{W}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, alquist, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock}, XValue: 2,
	}); err == nil {
		t.Fatal("a plain artifact paid the Clue sacrifice")
	}
	hand, life := me.Hand.Size(), me.Life
	p7Activate(t, g, me, alquist, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clue}, XValue: 2})

	if g.Battlefield.Contains(clue) {
		t.Error("the Clue was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
}

// --- Audience with Trostani ------------------------------------------------------

func TestMkmAC2AudienceWithTrostaniDrawsPerDifferentlyNamedCreatureToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	auraSeedLibrary(me, 5)
	// Two Detectives share a name; a Clue is not a creature; the
	// opponent's token is not mine. With the new Plant: two names.
	pushToken(g, me.ID, TokenCard("2/2 white and blue Detective"))
	pushToken(g, me.ID, TokenCard("2/2 white and blue Detective"))
	pushToken(g, me.ID, ClueToken())
	pushToken(g, opp.ID, TokenCard("0/1 green Plant"))
	pushToken(g, opp.ID, TokenCard("2/2 red and white Spirit"))

	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Audience with Trostani", "Sorcery", mkmAC2AudienceOracle, nil)
	passPriorityAroundTable(t, g)

	if got := tmtB1Count(g, me.ID, "Plant"); got != 1 {
		t.Fatalf("%d Plants, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (Detective and Plant)", got)
	}
}

// --- Buried in the Garden --------------------------------------------------------

func TestMkmAC2BuriedInTheGardenExilesUntilItLeavesAndAddsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")

	aura := castCatalogSpell(t, g, "Buried in the Garden", auraTypeLine, mkmAC2BuriedOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || onBattlefieldNamed(g, "Their Relic") != 0 {
		t.Fatal("the permanent was not exiled")
	}

	tapForMana(t, g, me.ID, forest)
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no pick for the additional mana of any color")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "W"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "W" {
		t.Errorf("pool = %v, want {G} and the chosen {W}", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if tmtB1Count(g, opp.ID, "Their Relic") != 1 {
		t.Error("the permanent did not return when the Aura left")
	}
}

func TestMkmAC2BuriedInTheGardenRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	theirLand := b31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	castCatalogSpell(t, g, "Buried in the Garden", auraTypeLine, mkmAC2BuriedOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirLand}); err == nil {
			t.Error("a land was accepted as the nonland target")
		}
	}
}

// --- Call a Surprise Witness -----------------------------------------------------

func TestMkmAC2CallASurpriseWitnessReturnsAFlyingSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{W}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
	})
	castCatalogSpell(t, g, "Call a Surprise Witness", "Sorcery", mkmAC2WitnessOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature card did not return")
	}
	if got := countersOn(g, bear, game.CounterFlying); got != 1 {
		t.Errorf("%d flying counters, want 1", got)
	}
	assertKeywords(t, g, bear, "flying")
	c := layeredCard(t, g, bear)
	if !c.HasSubtype("Spirit") || !c.HasSubtype("Bear") {
		t.Errorf("the creature is not a Spirit Bear: %v", c.Effective().Subtypes)
	}
}

func TestMkmAC2CallASurpriseWitnessRefusesManaValueFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := pushGraveyardCreature(g, me.ID, "Big Bear", "{2}{W}{W}")
	if err := castCatalogSpellErr(t, g, "Call a Surprise Witness", "Sorcery", mkmAC2WitnessOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}}); err == nil {
		t.Error("a mana value 4 creature card was accepted as the target")
	}
}

// --- Candlestick -----------------------------------------------------------------

func TestMkmAC2CandlestickEquipsPumpsSurveilsAndCracks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	stick := b12Push(g, me.ID, "Candlestick", "Artifact — Clue Equipment", mkmAC2CandlestickOracle, 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	mkmAC2AddMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, stick, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !attachedTo(t, g, stick, bear) {
		t.Fatal("Candlestick is not attached")
	}
	if p, tough := mkmB1PT(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped creature is %d/%d, want 3/3", p, tough)
	}

	seedLibrary(me, "One", "Two", "Three")
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("attacking did not surveil")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil looked at %d cards, want 2", len(c.ScryCards))
	}
	looked := append([]uuid.UUID(nil), c.ScryCards...)
	if err := g.ResolveSurveil(c.ID, me.ID, looked, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	for _, id := range looked {
		if !me.Graveyard.Contains(id) {
			t.Errorf("surveilled card %s is not in the graveyard", id)
		}
	}

	hand := me.Hand.Size()
	mkmAC2AddMana(t, g, me, "{C}{C}")
	p7Activate(t, g, me, stick, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(stick) {
		t.Error("Candlestick was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Case of the Gateway Express -------------------------------------------------

func TestMkmAC2GatewayExpressEachCreatureDealsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, name := range []string{"A", "B", "C"} {
		b12Creature(g, me.ID, "My "+name, "Creature — Bear", 2, 2)
	}
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 5)

	castCatalogSpell(t, g, "Case of the Gateway Express", mkmAC2CaseTypeLine, mkmAC2GatewayOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, ogre)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, ogre); got != 3 {
		t.Errorf("%d damage on the target, want 3 (one from each creature)", got)
	}
}

func TestMkmAC2GatewayExpressSolvesOnThreeAttackersAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	id := pushClass(g, me.ID, "Case of the Gateway Express", mkmAC2CaseTypeLine, mkmAC2GatewayOracle)
	var bears []uuid.UUID
	for _, name := range []string{"A", "B", "C"} {
		bears = append(bears, b12Creature(g, me.ID, "My "+name, "Creature — Bear", 2, 2))
	}
	if p, _ := mkmB1PT(t, g, bears[0]); p != 2 {
		t.Fatalf("an unsolved Case pumps: power %d", p)
	}

	declareAttack(t, g, opp.ID, bears...)
	mkmAC2RunUntilSolved(t, g, id)
	if p, tough := mkmB1PT(t, g, bears[0]); p != 3 || tough != 2 {
		t.Errorf("solved, a creature is %d/%d, want 3/2", p, tough)
	}
}

func TestMkmAC2GatewayExpressDoesNotSolveOnTwoAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	id := pushClass(g, me.ID, "Case of the Gateway Express", mkmAC2CaseTypeLine, mkmAC2GatewayOracle)
	a := b12Creature(g, me.ID, "My A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "My B", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	if mkmAC2EndStepSolves(t, g, id) {
		t.Error("the Case solved with two attackers")
	}
}

// --- Case of the Ransacked Lab ---------------------------------------------------

func TestMkmAC2RansackedLabDiscountsInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushClass(g, me.ID, "Case of the Ransacked Lab", mkmAC2CaseTypeLine, mkmAC2RansackedLabOracle)
	if got := priceInHand(t, g, me, "Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("instant: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Rite", "Sorcery", "{2}{R}"); got != 2 {
		t.Errorf("sorcery: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature: %d, want 2 (no discount)", got)
	}
}

func TestMkmAC2RansackedLabSolvesOnFourSpells(t *testing.T) {
	for _, tc := range []struct {
		spells int
		solved bool
	}{{3, false}, {4, true}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := pushClass(g, me.ID, "Case of the Ransacked Lab", mkmAC2CaseTypeLine, mkmAC2RansackedLabOracle)
		for i := 0; i < tc.spells; i++ {
			typeLine := "Instant"
			if i%2 == 1 {
				typeLine = "Sorcery"
			}
			castCatalogSpell(t, g, "Any Spell", typeLine, "", nil)
			passPriorityAroundTable(t, g)
		}
		if got := mkmAC2EndStepSolves(t, g, id); got != tc.solved {
			t.Errorf("%d spells: solved %v, want %v", tc.spells, got, tc.solved)
		}
	}
}

func TestMkmAC2RansackedLabDrawsOnlyOnceSolved(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	auraSeedLibrary(me, 3)
	id := pushClass(g, me.ID, "Case of the Ransacked Lab", mkmAC2CaseTypeLine, mkmAC2RansackedLabOracle)

	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("an unsolved Case drew: hand %d -> %d", hand, me.Hand.Size())
	}

	mkmAC2Solve(t, g, id)
	castCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on an instant once solved, want 1", got)
	}
	castCatalogSpell(t, g, "Any Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("a creature spell drew: %d drawn in all, want 1", got)
	}
}

// --- Case of the Trampled Garden -------------------------------------------------

func TestMkmAC2TrampledGardenDistributesTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b12Creature(g, me.ID, "My A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "My B", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Case of the Trampled Garden", mkmAC2CaseTypeLine, mkmAC2TrampledOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 1}, a, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := plusOnes(g, id); got != 1 {
			t.Errorf("%d +1/+1 counters on %s, want 1", got, id)
		}
	}
}

func TestMkmAC2TrampledGardenSolvesAtTotalPowerEight(t *testing.T) {
	for _, tc := range []struct {
		power  int
		solved bool
	}{{7, false}, {8, true}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := pushClass(g, me.ID, "Case of the Trampled Garden", mkmAC2CaseTypeLine, mkmAC2TrampledOracle)
		b12Creature(g, me.ID, "Big", "Creature — Beast", tc.power-3, 4)
		b12Creature(g, me.ID, "Small", "Creature — Beast", 3, 3)
		if got := mkmAC2EndStepSolves(t, g, id); got != tc.solved {
			t.Errorf("total power %d: solved %v, want %v", tc.power, got, tc.solved)
		}
	}
}

func TestMkmAC2TrampledGardenSolvedCountersAnAttackerAndGivesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	id := pushClass(g, me.ID, "Case of the Trampled Garden", mkmAC2CaseTypeLine, mkmAC2TrampledOracle)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mkmAC2Solve(t, g, id)

	declareAttack(t, g, opp.ID, bear)
	mscM2PickTargets(t, g, bear)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters on the attacker, want 1", got)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the attacker has no trample")
	}
}

func TestMkmAC2TrampledGardenUnsolvedDoesNotTriggerOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	pushClass(g, me.ID, "Case of the Trampled Garden", mkmAC2CaseTypeLine, mkmAC2TrampledOracle)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || plusOnes(g, bear) != 0 {
		t.Error("an unsolved Case triggered on the attack")
	}
}

// --- Chalk Outline ---------------------------------------------------------------

func TestMkmAC2ChalkOutlineMakesOneDetectiveAndClueForABatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Chalk Outline", "Enchantment", mkmAC2ChalkOracle, 0, 0)
	a := mkmAC2PushGraveyard(g, me, "Dead A", "Creature — Zombie", "", "{1}")
	b := mkmAC2PushGraveyard(g, me, "Dead B", "Creature — Zombie", "", "{1}")
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{a, b}) })
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives, want 1", got)
	}
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestMkmAC2ChalkOutlineIgnoresANoncreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Chalk Outline", "Enchantment", mkmAC2ChalkOracle, 0, 0)
	spell := mkmAC2PushGraveyard(g, me, "Old Spell", "Sorcery", "", "{1}")
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{spell}) })
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Detective"); got != 0 {
		t.Errorf("%d Detectives for a sorcery card, want 0", got)
	}
}

// --- Coerced to Kill -------------------------------------------------------------

func TestMkmAC2CoercedToKillStealsAndMakesAOneOneDeathtouchAssassin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	aura := auraCast(t, g, "Coerced to Kill", mkmAC2CoercedOracle, ogre)

	if got := controllerOf(t, g, ogre); got != me.ID {
		t.Errorf("controller %s, want the Aura's controller %s", got, me.ID)
	}
	if p, tough := mkmB1PT(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the creature is %d/%d, want 1/1", p, tough)
	}
	assertKeywords(t, g, ogre, "deathtouch")
	if c := layeredCard(t, g, ogre); !c.HasSubtype("Assassin") || !c.HasSubtype("Ogre") {
		t.Errorf("the creature is not an Ogre Assassin: %v", c.Effective().Subtypes)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if got := controllerOf(t, g, ogre); got != opp.ID {
		t.Errorf("control did not revert when the Aura left: %s", got)
	}
	if p, tough := mkmB1PT(t, g, ogre); p != 4 || tough != 4 {
		t.Errorf("the creature is %d/%d after the Aura left, want 4/4", p, tough)
	}
}

// --- Curious Cadaver -------------------------------------------------------------

func TestMkmAC2CuriousCadaverReturnsWhenYouSacrificeAClue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cadaver := mkmAC2PushGraveyard(g, me, "Curious Cadaver", "Creature — Zombie Detective", mkmAC2CadaverOracle, "{2}{U}{B}")

	theirClue := pushToken(g, opp.ID, ClueToken())
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirClue) })
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(cadaver) {
		t.Fatal("an opponent's Clue sacrifice returned the Cadaver")
	}

	myClue := pushToken(g, me.ID, ClueToken())
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(myClue) })
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(cadaver) {
		t.Error("the Cadaver did not return to hand when I sacrificed a Clue")
	}
}

func TestMkmAC2CuriousCadaverFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b12Push(g, me.ID, "Curious Cadaver", "Creature — Zombie Detective", mkmAC2CadaverOracle, 3, 1)
	assertKeywords(t, g, id, "flying")
}

// --- Curious Inquiry -------------------------------------------------------------

func TestMkmAC2CuriousInquiryPumpsAndInvestigatesOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Curious Inquiry", mkmAC2InquiryOracle, bear)
	if p, tough := mkmB1PT(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("the enchanted creature is %d/%d, want 3/3", p, tough)
	}
	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestMkmAC2CuriousInquiryClueGoesToTheCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := auraBear(g, opp.ID)
	auraCast(t, g, "Curious Inquiry", mkmAC2InquiryOracle, theirs)
	dealCombatDamageToPlayer(g, theirs, me.ID, 3)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, opp.ID, "Clue"); got != 1 {
		t.Errorf("the creature's controller has %d Clues, want 1", got)
	}
	if got := tmtB1Count(g, me.ID, "Clue"); got != 0 {
		t.Errorf("the Aura's controller has %d Clues, want 0", got)
	}
}
