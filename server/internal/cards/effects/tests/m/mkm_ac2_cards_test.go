package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMkmAC2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.MkmAC2AlquistOracle:      "Alquist Proft, Master Sleuth",
		effectstest.MkmAC2AudienceOracle:     "Audience with Trostani",
		effectstest.MkmAC2BuriedOracle:       "Buried in the Garden",
		effectstest.MkmAC2WitnessOracle:      "Call a Surprise Witness",
		effectstest.MkmAC2CandlestickOracle:  "Candlestick",
		effectstest.MkmAC2GatewayOracle:      "Case of the Gateway Express",
		effectstest.MkmAC2RansackedLabOracle: "Case of the Ransacked Lab",
		effectstest.MkmAC2TrampledOracle:     "Case of the Trampled Garden",
		effectstest.MkmAC2ChalkOracle:        "Chalk Outline",
		effectstest.MkmAC2CoercedOracle:      "Coerced to Kill",
		effectstest.MkmAC2CadaverOracle:      "Curious Cadaver",
		effectstest.MkmAC2InquiryOracle:      "Curious Inquiry",
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

// mkmAC2RunUntilSolved walks the game forward until the Case solves,
// failing if it never does.
func mkmAC2RunUntilSolved(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 400 && !effectstest.MkmB1Solved(g, id); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.MkmB1Solved(g, id) {
		t.Fatal("the Case never solved")
	}
}

// --- Alquist Proft, Master Sleuth ------------------------------------------------

func TestMkmAC2AlquistProftInvestigatesOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastHoldCreature(t, g, "Alquist Proft, Master Sleuth", "Legendary Creature — Human Detective", effectstest.MkmAC2AlquistOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, id, "vigilance")
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues on entering, want 1", got)
	}
}

func TestMkmAC2AlquistProftSacrificesAClueToDrawXAndGainX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AuraSeedLibrary(me, 4)
	effectstest.AdvanceToMain(t, g)
	alquist := effectstest.B12Push(g, me.ID, "Alquist Proft, Master Sleuth", "Legendary Creature — Human Detective", effectstest.MkmAC2AlquistOracle, 3, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())

	effectstest.MkmAC2AddMana(t, g, me, "{C}{C}{W}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, alquist, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{rock}, XValue: 2,
	}); err == nil {
		t.Fatal("a plain artifact paid the Clue sacrifice")
	}
	hand, life := me.Hand.Size(), me.Life
	effectstest.P7Activate(t, g, me, alquist, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{clue}, XValue: 2})

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AuraSeedLibrary(me, 5)
	// Two Detectives share a name; a Clue is not a creature; the
	// opponent's token is not mine. With the new Plant: two names.
	effectstest.PushToken(g, me.ID, effects.TokenCard("2/2 white and blue Detective"))
	effectstest.PushToken(g, me.ID, effects.TokenCard("2/2 white and blue Detective"))
	effectstest.PushToken(g, me.ID, effects.ClueToken())
	effectstest.PushToken(g, opp.ID, effects.TokenCard("0/1 green Plant"))
	effectstest.PushToken(g, opp.ID, effects.TokenCard("2/2 red and white Spirit"))

	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Audience with Trostani", "Sorcery", effectstest.MkmAC2AudienceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.TmtB1Count(g, me.ID, "Plant"); got != 1 {
		t.Fatalf("%d Plants, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (Detective and Plant)", got)
	}
}

// --- Buried in the Garden --------------------------------------------------------

func TestMkmAC2BuriedInTheGardenExilesUntilItLeavesAndAddsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")

	aura := effectstest.CastCatalogSpell(t, g, "Buried in the Garden", effectstest.AuraTypeLine, effectstest.MkmAC2BuriedOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || effectstest.OnBattlefieldNamed(g, "Their Relic") != 0 {
		t.Fatal("the permanent was not exiled")
	}

	effectstest.TapForMana(t, g, me.ID, forest)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no pick for the additional mana of any color")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "W"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "W" {
		t.Errorf("pool = %v, want {G} and the chosen {W}", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if effectstest.TmtB1Count(g, opp.ID, "Their Relic") != 1 {
		t.Error("the permanent did not return when the Aura left")
	}
}

func TestMkmAC2BuriedInTheGardenRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	theirLand := effectstest.B31Push(g, opp.ID, "Island", "Basic Land — Island", "", "", 0, 0)
	effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Buried in the Garden", effectstest.AuraTypeLine, effectstest.MkmAC2BuriedOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: forest}})
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirLand}); err == nil {
			t.Error("a land was accepted as the nonland target")
		}
	}
}

// --- Call a Surprise Witness -----------------------------------------------------

func TestMkmAC2CallASurpriseWitnessReturnsAFlyingSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{W}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
	})
	effectstest.CastCatalogSpell(t, g, "Call a Surprise Witness", "Sorcery", effectstest.MkmAC2WitnessOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature card did not return")
	}
	if got := effectstest.CountersOn(g, bear, game.CounterFlying); got != 1 {
		t.Errorf("%d flying counters, want 1", got)
	}
	effectstest.AssertKeywords(t, g, bear, "flying")
	c := effectstest.LayeredCard(t, g, bear)
	if !c.HasSubtype("Spirit") || !c.HasSubtype("Bear") {
		t.Errorf("the creature is not a Spirit Bear: %v", c.Effective().Subtypes)
	}
}

func TestMkmAC2CallASurpriseWitnessRefusesManaValueFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := effectstest.PushGraveyardCreature(g, me.ID, "Big Bear", "{2}{W}{W}")
	if err := effectstest.CastCatalogSpellErr(t, g, "Call a Surprise Witness", "Sorcery", effectstest.MkmAC2WitnessOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}}); err == nil {
		t.Error("a mana value 4 creature card was accepted as the target")
	}
}

// --- Candlestick -----------------------------------------------------------------

func TestMkmAC2CandlestickEquipsPumpsSurveilsAndCracks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	stick := effectstest.B12Push(g, me.ID, "Candlestick", "Artifact — Clue Equipment", effectstest.MkmAC2CandlestickOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	effectstest.MkmAC2AddMana(t, g, me, "{C}{C}")
	effectstest.P7Activate(t, g, me, stick, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !effectstest.AttachedTo(t, g, stick, bear) {
		t.Fatal("Candlestick is not attached")
	}
	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped creature is %d/%d, want 3/3", p, tough)
	}

	effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
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
	effectstest.MkmAC2AddMana(t, g, me, "{C}{C}")
	effectstest.P7Activate(t, g, me, stick, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(stick) {
		t.Error("Candlestick was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Case of the Gateway Express -------------------------------------------------

func TestMkmAC2GatewayExpressEachCreatureDealsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, name := range []string{"A", "B", "C"} {
		effectstest.B12Creature(g, me.ID, "My "+name, "Creature — Bear", 2, 2)
	}
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 5)

	effectstest.CastCatalogSpell(t, g, "Case of the Gateway Express", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2GatewayOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, ogre); got != 3 {
		t.Errorf("%d damage on the target, want 3 (one from each creature)", got)
	}
}

func TestMkmAC2GatewayExpressSolvesOnThreeAttackersAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushClass(g, me.ID, "Case of the Gateway Express", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2GatewayOracle)
	var bears []uuid.UUID
	for _, name := range []string{"A", "B", "C"} {
		bears = append(bears, effectstest.B12Creature(g, me.ID, "My "+name, "Creature — Bear", 2, 2))
	}
	if p, _ := effectstest.MkmB1PT(t, g, bears[0]); p != 2 {
		t.Fatalf("an unsolved Case pumps: power %d", p)
	}

	effectstest.DeclareAttack(t, g, opp.ID, bears...)
	mkmAC2RunUntilSolved(t, g, id)
	if p, tough := effectstest.MkmB1PT(t, g, bears[0]); p != 3 || tough != 2 {
		t.Errorf("solved, a creature is %d/%d, want 3/2", p, tough)
	}
}

func TestMkmAC2GatewayExpressDoesNotSolveOnTwoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushClass(g, me.ID, "Case of the Gateway Express", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2GatewayOracle)
	a := effectstest.B12Creature(g, me.ID, "My A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "My B", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	if effectstest.MkmAC2EndStepSolves(t, g, id) {
		t.Error("the Case solved with two attackers")
	}
}

// --- Case of the Ransacked Lab ---------------------------------------------------

func TestMkmAC2RansackedLabDiscountsInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushClass(g, me.ID, "Case of the Ransacked Lab", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2RansackedLabOracle)
	if got := effectstest.PriceInHand(t, g, me, "Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("instant: %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Rite", "Sorcery", "{2}{R}"); got != 2 {
		t.Errorf("sorcery: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature: %d, want 2 (no discount)", got)
	}
}

func TestMkmAC2RansackedLabSolvesOnFourSpells(t *testing.T) {
	for _, tc := range []struct {
		spells int
		solved bool
	}{{3, false}, {4, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := effectstest.PushClass(g, me.ID, "Case of the Ransacked Lab", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2RansackedLabOracle)
		for i := 0; i < tc.spells; i++ {
			typeLine := "Instant"
			if i%2 == 1 {
				typeLine = "Sorcery"
			}
			effectstest.CastCatalogSpell(t, g, "Any Spell", typeLine, "", nil)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if got := effectstest.MkmAC2EndStepSolves(t, g, id); got != tc.solved {
			t.Errorf("%d spells: solved %v, want %v", tc.spells, got, tc.solved)
		}
	}
}

func TestMkmAC2RansackedLabDrawsOnlyOnceSolved(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AuraSeedLibrary(me, 3)
	id := effectstest.PushClass(g, me.ID, "Case of the Ransacked Lab", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2RansackedLabOracle)

	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("an unsolved Case drew: hand %d -> %d", hand, me.Hand.Size())
	}

	effectstest.MkmAC2Solve(t, g, id)
	effectstest.CastCatalogSpell(t, g, "Any Bolt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on an instant once solved, want 1", got)
	}
	effectstest.CastCatalogSpell(t, g, "Any Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("a creature spell drew: %d drawn in all, want 1", got)
	}
}

// --- Case of the Trampled Garden -------------------------------------------------

func TestMkmAC2TrampledGardenDistributesTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "My A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "My B", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Case of the Trampled Garden", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2TrampledOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 1}, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.PlusOnes(g, id); got != 1 {
			t.Errorf("%d +1/+1 counters on %s, want 1", got, id)
		}
	}
}

func TestMkmAC2TrampledGardenSolvesAtTotalPowerEight(t *testing.T) {
	for _, tc := range []struct {
		power  int
		solved bool
	}{{7, false}, {8, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := effectstest.PushClass(g, me.ID, "Case of the Trampled Garden", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2TrampledOracle)
		effectstest.B12Creature(g, me.ID, "Big", "Creature — Beast", tc.power-3, 4)
		effectstest.B12Creature(g, me.ID, "Small", "Creature — Beast", 3, 3)
		if got := effectstest.MkmAC2EndStepSolves(t, g, id); got != tc.solved {
			t.Errorf("total power %d: solved %v, want %v", tc.power, got, tc.solved)
		}
	}
}

func TestMkmAC2TrampledGardenSolvedCountersAnAttackerAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushClass(g, me.ID, "Case of the Trampled Garden", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2TrampledOracle)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.MkmAC2Solve(t, g, id)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.MscM2PickTargets(t, g, bear)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters on the attacker, want 1", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the attacker has no trample")
	}
}

func TestMkmAC2TrampledGardenUnsolvedDoesNotTriggerOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushClass(g, me.ID, "Case of the Trampled Garden", effectstest.MkmAC2CaseTypeLine, effectstest.MkmAC2TrampledOracle)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.PlusOnes(g, bear) != 0 {
		t.Error("an unsolved Case triggered on the attack")
	}
}

// --- Chalk Outline ---------------------------------------------------------------

func TestMkmAC2ChalkOutlineMakesOneDetectiveAndClueForABatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Chalk Outline", "Enchantment", effectstest.MkmAC2ChalkOracle, 0, 0)
	a := effectstest.MkmAC2PushGraveyard(g, me, "Dead A", "Creature — Zombie", "", "{1}")
	b := effectstest.MkmAC2PushGraveyard(g, me, "Dead B", "Creature — Zombie", "", "{1}")
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{a, b}) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives, want 1", got)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestMkmAC2ChalkOutlineIgnoresANoncreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Chalk Outline", "Enchantment", effectstest.MkmAC2ChalkOracle, 0, 0)
	spell := effectstest.MkmAC2PushGraveyard(g, me, "Old Spell", "Sorcery", "", "{1}")
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{spell}) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Detective"); got != 0 {
		t.Errorf("%d Detectives for a sorcery card, want 0", got)
	}
}

// --- Coerced to Kill -------------------------------------------------------------

func TestMkmAC2CoercedToKillStealsAndMakesAOneOneDeathtouchAssassin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	aura := effectstest.AuraCast(t, g, "Coerced to Kill", effectstest.MkmAC2CoercedOracle, ogre)

	if got := effectstest.ControllerOf(t, g, ogre); got != me.ID {
		t.Errorf("controller %s, want the Aura's controller %s", got, me.ID)
	}
	if p, tough := effectstest.MkmB1PT(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the creature is %d/%d, want 1/1", p, tough)
	}
	effectstest.AssertKeywords(t, g, ogre, "deathtouch")
	if c := effectstest.LayeredCard(t, g, ogre); !c.HasSubtype("Assassin") || !c.HasSubtype("Ogre") {
		t.Errorf("the creature is not an Ogre Assassin: %v", c.Effective().Subtypes)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	if got := effectstest.ControllerOf(t, g, ogre); got != opp.ID {
		t.Errorf("control did not revert when the Aura left: %s", got)
	}
	if p, tough := effectstest.MkmB1PT(t, g, ogre); p != 4 || tough != 4 {
		t.Errorf("the creature is %d/%d after the Aura left, want 4/4", p, tough)
	}
}

// --- Curious Cadaver -------------------------------------------------------------

func TestMkmAC2CuriousCadaverReturnsWhenYouSacrificeAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cadaver := effectstest.MkmAC2PushGraveyard(g, me, "Curious Cadaver", "Creature — Zombie Detective", effectstest.MkmAC2CadaverOracle, "{2}{U}{B}")

	theirClue := effectstest.PushToken(g, opp.ID, effects.ClueToken())
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirClue) })
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(cadaver) {
		t.Fatal("an opponent's Clue sacrifice returned the Cadaver")
	}

	myClue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(myClue) })
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cadaver) {
		t.Error("the Cadaver did not return to hand when I sacrificed a Clue")
	}
}

func TestMkmAC2CuriousCadaverFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B12Push(g, me.ID, "Curious Cadaver", "Creature — Zombie Detective", effectstest.MkmAC2CadaverOracle, 3, 1)
	effectstest.AssertKeywords(t, g, id, "flying")
}

// --- Curious Inquiry -------------------------------------------------------------

func TestMkmAC2CuriousInquiryPumpsAndInvestigatesOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Curious Inquiry", effectstest.MkmAC2InquiryOracle, bear)
	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("the enchanted creature is %d/%d, want 3/3", p, tough)
	}
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestMkmAC2CuriousInquiryClueGoesToTheCreaturesController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Curious Inquiry", effectstest.MkmAC2InquiryOracle, theirs)
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, opp.ID, "Clue"); got != 1 {
		t.Errorf("the creature's controller has %d Clues, want 1", got)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 0 {
		t.Errorf("the Aura's controller has %d Clues, want 0", got)
	}
}
