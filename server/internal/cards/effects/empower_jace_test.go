package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// empower_jace_test.go — ADR 0139 (#2796): the Jace planeswalker token
// and the Empower Jace keyword action (CR 701.71a), then one test per
// proof card.

const (
	oracleNoAdmittance          = "d53ae01b-87e1-48e1-9fc3-23a734ce6622"
	oracleCountersculpt         = "fc3353e9-fe24-4ca7-bff8-9da767f2903a"
	oracleRewriteRegrets        = "42565432-3a17-4274-87fc-016ab900e313"
	oracleCampusCrier           = "c1a9d60f-a469-48f9-8f1c-3ca5ae86bd64"
	oracleSolveForDisappoint    = "3241a3f5-7ae5-4497-9dd2-a55f3370a055"
	oraclePlanForAllOutcomes    = "7fd2cac5-cd29-4a47-9e89-9c6be39a0b67"
	oracleMindseekerOculus      = "aa9547b3-e4e4-4557-9dd1-3aa9a7cb9937"
	oracleJaceRealitySculptor   = "bf0e9e00-fba5-448a-9052-ae730aea4bf8"
	oracleAcademicAscent        = "406e853b-713f-4c89-9efb-aca74a11183f"
	oracleProtegesAwakening     = "07617648-6490-4719-acec-7705671f7e01"
	oracleTamsResistance        = "7500b7dc-588d-46bd-acc5-885dcd464668"
	oracleVraskasFinalMercy     = "7799413d-e313-4a05-ad28-f4c926c26d28"
	oracleArcaneAmphisbaena     = "5c151e3b-fcf2-4adb-a60b-9eff0c360705"
	oracleKeeperOfTheQuietHour  = "ecdfee75-15fd-41ea-89d4-4bfc1fa4d032"
	oracleRepurposedEnforcer    = "b775a404-ec0a-47ff-bb9d-eae2273d945b"
	jaceRealitySculptorTypeLine = "Legendary Planeswalker — Jace"
)

// jaceTokensOf lists the Jace planeswalker tokens a player controls.
func jaceTokensOf(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.WithWriteLock(func() { out = JaceTokensControlledBy(g, player) })
	return out
}

// onlyJaceToken returns the one Jace token a player controls, failing
// the test when there is not exactly one.
func onlyJaceToken(t *testing.T, g *game.Game, player uuid.UUID) uuid.UUID {
	t.Helper()
	jaces := jaceTokensOf(g, player)
	if len(jaces) != 1 {
		t.Fatalf("player controls %d Jace tokens, want 1", len(jaces))
	}
	return jaces[0]
}

// seedJaceToken creates a Jace token for player through the ordinary
// creation path and puts `loyalty` counters on it, in one locked call
// so no state-based action check sees it empty.
func seedJaceToken(t *testing.T, g *game.Game, player uuid.UUID, loyalty int) uuid.UUID {
	t.Helper()
	var made []uuid.UUID
	var err error
	g.WithWriteLock(func() {
		made, err = g.CreateTokensForEffect(player, JaceToken(), 1, game.TokenEntryOptions{})
		if err == nil && len(made) == 1 {
			err = g.AddCounterForEffect(made[0], game.CounterLoyalty, loyalty)
		}
	})
	if err != nil || len(made) != 1 {
		t.Fatalf("seed a Jace token: made %v, err %v", made, err)
	}
	return made[0]
}

// pushJaceRealitySculptor seats Jace, Reality Sculptor with his printed
// type line — "Legendary Planeswalker — Jace", so he IS a Jace for "among
// Jaces you control" — and 5 loyalty.
func pushJaceRealitySculptor(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Jace, Reality Sculptor", TypeLine: jaceRealitySculptorTypeLine,
		OracleID: oracleJaceRealitySculptor, Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: 5},
	})
	return id
}

// castProtegesAwakening casts Protege's Awakening (Empower Jace 6, then
// draw a card), the plainest card on the seam.
func castProtegesAwakening(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Protege's Awakening", "Sorcery", oracleProtegesAwakening, nil)
}

// --- the token -----------------------------------------------------

// The token is what CR 701.71a says it is: named Jace, a blue
// non-legendary "Token Planeswalker — Jace" with 0 printed loyalty and
// the two loyalty abilities, reached through the catalog like a printed
// planeswalker's.
func TestJaceTokenIsABlueJacePlaneswalkerTokenWithTwoLoyaltyAbilities(t *testing.T) {
	tok := JaceToken()
	if tok.Name != "Jace" || tok.TypeLine != "Token Planeswalker — Jace" {
		t.Fatalf("token printed as %q / %q", tok.Name, tok.TypeLine)
	}
	if len(tok.Colors) != 1 || tok.Colors[0] != "U" {
		t.Errorf("token colours %v, want [U]", tok.Colors)
	}
	if tok.StartingLoyalty != 0 {
		t.Errorf("printed loyalty %d, want 0", tok.StartingLoyalty)
	}
	if tok.IsLegendary() {
		t.Error("the Jace token is not legendary")
	}
	if !IsJacePlaneswalkerToken(tok) {
		t.Error("the template is not a Jace planeswalker token")
	}
	abs := game.ActivatedAbilitiesForCard(tok)
	if len(abs) != 2 {
		t.Fatalf("token has %d activated abilities, want 2", len(abs))
	}
	for i, want := range []int{-1, -3} {
		if abs[i].Cost.Loyalty == nil || *abs[i].Cost.Loyalty != want {
			t.Errorf("ability %d loyalty cost %v, want %d", i, abs[i].Cost.Loyalty, want)
		}
	}
	if game.TokenTextForCard(tok) == "" {
		t.Error("the token prints no ability text for the board")
	}
}

// Its loyalty abilities are ordinary CR 606 activations: the −3 draws,
// the turn's one activation is spent, a −3 at 2 loyalty is refused
// (CR 606.6), and the −1 surveils.
func TestJaceTokenLoyaltyAbilitiesActivate(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := seedJaceToken(t, g, me.ID, 4)

	before := handSize(me)
	b16Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{})
	if got := handSize(me); got != before+1 {
		t.Fatalf("−3: hand %d, want %d", got, before+1)
	}
	if got := loyaltyCount(g, jace); got != 1 {
		t.Fatalf("loyalty after −3 from 4: %d, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, jace, 0, game.ActivateAbilityParams{}); err != game.ErrLoyaltyAlreadyActivated {
		t.Fatalf("a second activation the same turn: %v, want ErrLoyaltyAlreadyActivated", err)
	}

	other := seedJaceToken(t, g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, other, 1, game.ActivateAbilityParams{}); err != game.ErrInsufficientLoyalty {
		t.Fatalf("−3 at 2 loyalty: %v, want ErrInsufficientLoyalty", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, other, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c == nil {
		t.Fatal("the −1 asked no surveil question")
	}
	if got := loyaltyCount(g, other); got != 1 {
		t.Fatalf("loyalty after −1 from 2: %d, want 1", got)
	}
}

// --- the keyword action ------------------------------------------------

// With no Jace token, the action creates one and puts N counters on it,
// all in one instruction: it is on the battlefield with N loyalty after
// the resolution and the state-based check that follows it.
func TestEmpowerJaceCreatesTheTokenAndPutsNCountersOnIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := handSize(me)
	castProtegesAwakening(t, g)
	passPriorityAroundTable(t, g)

	jace := onlyJaceToken(t, g, me.ID)
	if got := loyaltyCount(g, jace); got != 6 {
		t.Fatalf("Jace token loyalty %d, want 6", got)
	}
	if got := handSize(me); got != before+1 {
		t.Fatalf("hand %d after \"Draw a card.\", want %d", got, before+1)
	}
	c, _ := battlefieldCard(g, jace)
	if c.Controller != me.ID || !c.IsToken() {
		t.Fatalf("token controller %s token=%v", c.Controller, c.IsToken())
	}
}

// With a Jace token, the action makes no second one: the counters go on
// the one you have.
func TestEmpowerJaceAddsToTheJaceTokenYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jace := seedJaceToken(t, g, me.ID, 3)
	castProtegesAwakening(t, g)
	passPriorityAroundTable(t, g)

	if got := len(jaceTokensOf(g, me.ID)); got != 1 {
		t.Fatalf("%d Jace tokens after empowering an existing one, want 1", got)
	}
	if got := loyaltyCount(g, jace); got != 9 {
		t.Fatalf("Jace token loyalty %d, want 3+6", got)
	}
}

// An OPPONENT's Jace token is not "a Jace token you control": the
// action creates your own.
func TestEmpowerJaceIgnoresAnOpponentsJaceToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := seedJaceToken(t, g, opp.ID, 2)
	castProtegesAwakening(t, g)
	passPriorityAroundTable(t, g)

	if got := loyaltyCount(g, theirs); got != 2 {
		t.Fatalf("the opponent's Jace went to %d loyalty", got)
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 6 {
		t.Fatalf("my new Jace token has %d loyalty, want 6", got)
	}
}

// Two Jace tokens: the controller chooses which one gets the counters,
// and "Draw a card." waits for the answer.
func TestEmpowerJaceAsksWhichJaceWhenThereAreSeveral(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := seedJaceToken(t, g, me.ID, 1)
	second := seedJaceToken(t, g, me.ID, 1)
	before := handSize(me)
	castProtegesAwakening(t, g)
	passPriorityAroundTable(t, g)

	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no prompt asked which Jace gets the counters")
	}
	if len(c.ChooseCards) != 2 || c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Fatalf("prompt offers %v (%d..%d), want both Jaces, exactly one", c.ChooseCards, c.ChooseMin, c.ChooseMax)
	}
	if got := handSize(me); got != before {
		t.Fatal("\"Draw a card.\" ran before the Jace was chosen")
	}
	answerOwnPermanents(t, g, me.ID, second)
	passPriorityAroundTable(t, g)

	if got := loyaltyCount(g, second); got != 7 {
		t.Fatalf("chosen Jace loyalty %d, want 7", got)
	}
	if got := loyaltyCount(g, first); got != 1 {
		t.Fatalf("the other Jace went to %d loyalty", got)
	}
	if got := handSize(me); got != before+1 {
		t.Fatalf("hand %d after the answer, want %d", got, before+1)
	}
}

// Doubling Season doubles the token AND the counters (CR 701.7b, CR
// 614.1): two Jaces are created, the controller picks one, it gets 2N,
// and the other — 0 loyalty — goes at the state-based check (CR
// 704.5i), which is held until the answer (CR 704.3).
func TestEmpowerJaceUnderDoublingSeasonPicksOneOfTheTwoJaces(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	castProtegesAwakening(t, g)
	passPriorityAroundTable(t, g)

	jaces := jaceTokensOf(g, me.ID)
	if len(jaces) != 2 {
		t.Fatalf("%d Jace tokens under Doubling Season, want 2", len(jaces))
	}
	answerOwnPermanents(t, g, me.ID, jaces[0])
	passPriorityAroundTable(t, g)

	if got := loyaltyCount(g, jaces[0]); got != 12 {
		t.Fatalf("chosen Jace loyalty %d, want 6 doubled", got)
	}
	if _, ok := battlefieldCard(g, jaces[1]); ok {
		t.Fatal("the unchosen 0-loyalty Jace survived the state-based check")
	}
}

// "Empower Jace 0" still creates the token, and a 0-loyalty planeswalker
// then dies to the state-based check: the rules, not a special case.
// Jace, Reality Sculptor's +1 with no Islands is exactly that.
func TestEmpowerJaceZeroCreatesATokenThatDies(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jrs := pushJaceRealitySculptor(g, me.ID)
	b16Activate(t, g, me.ID, jrs, 0, game.ActivateAbilityParams{})
	created := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventTokenCreated && ev.Actor == me.ID {
			created = true
		}
	}
	if !created {
		t.Fatal("Empower Jace 0 created no token")
	}
	if got := len(jaceTokensOf(g, me.ID)); got != 0 {
		t.Fatalf("a 0-loyalty Jace token survived the state-based check (%d left)", got)
	}
}

// --- proof cards -----------------------------------------------------

func TestNoAdmittanceDealsThreeAndEmpowersJaceOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	castCatalogSpell(t, g, "No Admittance", "Sorcery", oracleNoAdmittance,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("opponent life %d, want %d", opp.Life, life-3)
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 1 {
		t.Fatalf("Jace loyalty %d, want 1", got)
	}
}

// Countersculpt beholds the Jace token an earlier Empower Jace made,
// counters the spell and empowers that same Jace.
func TestCountersculptBeholdsAJaceTokenCountersAndEmpowers(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	caster := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jace := seedJaceToken(t, g, caster.ID, 2)

	boltID := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: boltID, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: "4457ed35-7c10-48c8-9776-456485fdf070", Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	counterID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: counterID, Name: "Countersculpt", TypeLine: "Instant",
		OracleID: oracleCountersculpt, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, counterID, game.CastSpellParams{
		Targets:    []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
		CostBranch: branch(0),
		RevealIDs:  []uuid.UUID{jace},
	}); err != nil {
		t.Fatalf("cast Countersculpt beholding the Jace token: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(boltID) {
		t.Fatal("the Bolt was not countered")
	}
	if got := loyaltyCount(g, jace); got != 3 {
		t.Fatalf("Jace loyalty %d, want 2+1", got)
	}
}

func TestRewriteRegretsReanimatesAndEmpowersJaceTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Six Drop", TypeLine: "Creature — Giant", ManaCost: "{5}{B}",
		Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Rewrite Regrets", "Sorcery", oracleRewriteRegrets,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead}})
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, dead); !ok || c.Controller != me.ID {
		t.Fatal("the creature card did not return under my control")
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 2 {
		t.Fatalf("Jace loyalty %d, want 2", got)
	}
}

// Mana value 7 is not "6 or less": the cast is refused at announce.
func TestRewriteRegretsRefusesAManaValueSevenCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: big, Name: "Seven Drop", TypeLine: "Creature — Giant", ManaCost: "{6}{B}",
		Owner: me.ID, Controller: me.ID,
	})
	if err := castCatalogSpellErr(t, g, "Rewrite Regrets", "Sorcery", oracleRewriteRegrets,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}}); err == nil {
		t.Fatal("a mana value 7 creature card was a legal target")
	}
}

func TestCampusCrierExilesFromTheGraveyardToEmpowerJaceTwo(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	crier := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: crier, Name: "Campus Crier", TypeLine: "Creature — Human Advisor",
		OracleID: oracleCampusCrier, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { me.ManaPool.AddMana(game.ManaToken{Color: "C"}) })
	b16Activate(t, g, me.ID, crier, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(crier) {
		t.Fatal("Campus Crier was not exiled to pay the cost")
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 2 {
		t.Fatalf("Jace loyalty %d, want 2", got)
	}
}

func TestSolveForDisappointmentDiscardsThenEmpowersJaceOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp.Hand.Cards = nil
	bear := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Solve for Disappointment", "Sorcery", oracleSolveForDisappoint,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if len(jaceTokensOf(g, me.ID)) != 0 {
		t.Fatal("Jace was empowered before the card was chosen")
	}
	pick := openVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("choose the Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bear) {
		t.Fatal("the chosen nonland permanent card was not discarded")
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 1 {
		t.Fatalf("Jace loyalty %d, want 1", got)
	}
}

func TestPlanForAllOutcomesFirstNoncreatureSpellEmpowersJaceOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Plan for All Outcomes", "Enchantment", oraclePlanForAllOutcomes, false)
	castProtegesAwakening(t, g) // a noncreature spell: its own Empower Jace 6, plus the trigger's 1
	passPriorityAroundTable(t, g)
	jace := onlyJaceToken(t, g, me.ID)
	if got := loyaltyCount(g, jace); got != 7 {
		t.Fatalf("Jace loyalty %d after the first noncreature spell, want 6+1", got)
	}
	castProtegesAwakening(t, g) // the second: no trigger
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, jace); got != 13 {
		t.Fatalf("Jace loyalty %d after the second noncreature spell, want 7+6", got)
	}
}

// The enters trigger: the target's OWNER answers top or bottom.
func TestPlanForAllOutcomesAsksTheOwnerTopOrBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Plan for All Outcomes", "Enchantment", oraclePlanForAllOutcomes, nil)
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatal("the owner was not asked top or bottom")
	}
}

func TestMindseekerOculusEntersAndEmpowersJaceFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Mindseeker Oculus", "Creature — Homunculus", oracleMindseekerOculus, nil)
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 4 {
		t.Fatalf("Jace loyalty %d, want 4", got)
	}
}

func TestKeeperAndAmphisbaenaEnterAndEmpowerJaceTwo(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Keeper of the Quiet Hour", "Artifact Creature — Chimera", oracleKeeperOfTheQuietHour},
		{"Arcane Amphisbaena", "Creature — Snake", oracleArcaneAmphisbaena},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 2 {
				t.Fatalf("Jace loyalty %d, want 2", got)
			}
		})
	}
}

func TestArcaneAmphisbaenaHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	snake := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Arcane Amphisbaena", TypeLine: "Creature — Snake",
		OracleID: oracleArcaneAmphisbaena, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if !effectiveAbilitiesContain(t, g, snake, "deathtouch") {
		t.Fatal("no deathtouch")
	}
}

// X is counted as the attack trigger resolves: three creatures, three
// counters.
func TestRepurposedEnforcerEmpowersJaceByCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	enforcer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Repurposed Enforcer", TypeLine: "Creature — Human Soldier",
		OracleID: oracleRepurposedEnforcer, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 2; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
	}
	declareAttack(t, g, opp.ID, enforcer)
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 3 {
		t.Fatalf("Jace loyalty %d, want 3 (creatures you control)", got)
	}
}

func TestAcademicAscentPumpsGrantsFlyingAndEmpowersJaceTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Academic Ascent", "Instant", oracleAcademicAscent,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 4 {
		t.Fatalf("power %d, want 4", got)
	}
	if !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Fatal("no flying")
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 2 {
		t.Fatalf("Jace loyalty %d, want 2", got)
	}
}

func TestTamsResistanceWithNoTargetStillEmpowersJaceFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Tam's Resistance", "Sorcery", oracleTamsResistance, nil)
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 4 {
		t.Fatalf("Jace loyalty %d, want 4", got)
	}
}

func TestTamsResistanceCountersAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Tam's Resistance", "Sorcery", oracleTamsResistance,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, bear)
	if c.Counters["+1/+1"] != 1 {
		t.Fatalf("+1/+1 counters %d, want 1", c.Counters["+1/+1"])
	}
	if !effectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Fatal("no vigilance")
	}
}

func TestVraskasFinalMercyEmpowerBulletLosesTwoAndEmpowersSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpellWithModes(t, g, "Vraska's Final Mercy", "Sorcery", oracleVraskasFinalMercy, []int{1})
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Fatalf("life %d, want %d", me.Life, life-2)
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 6 {
		t.Fatalf("Jace loyalty %d, want 6", got)
	}
}

// --- Jace, Reality Sculptor ---------------------------------------

// The +1 counts Islands as it resolves and empowers a Jace TOKEN; Jace
// himself is a Jace but not a token, and only his +1's own cost moves
// his loyalty.
func TestJaceRealitySculptorPlusOneEmpowersATokenByIslands(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jrs := pushJaceRealitySculptor(g, me.ID)
	for i := 0; i < 3; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID,
		})
	}
	b16Activate(t, g, me.ID, jrs, 0, game.ActivateAbilityParams{})
	if got := loyaltyCount(g, jrs); got != 6 {
		t.Fatalf("Jace, Reality Sculptor loyalty %d, want 5+1", got)
	}
	if got := loyaltyCount(g, onlyJaceToken(t, g, me.ID)); got != 3 {
		t.Fatalf("Jace token loyalty %d, want 3 (Islands)", got)
	}
}

// The 0 needs twenty-five loyalty among Jaces you control — Jace and
// the tokens together — and exiles all but the bottom card of each
// opponent's library.
func TestJaceRealitySculptorZeroNeedsTwentyFiveAmongJaces(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jrs := pushJaceRealitySculptor(g, me.ID)
	tok := seedJaceToken(t, g, me.ID, 19)
	if err := g.ActivateCatalogAbility(me.ID, jrs, 2, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the 0 activated with 24 loyalty among Jaces")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tok, game.CounterLoyalty, 1) })
	bottom := opp.Library.Cards[0].InstanceID
	b16Activate(t, g, me.ID, jrs, 2, game.ActivateAbilityParams{})
	if len(opp.Library.Cards) != 1 || opp.Library.Cards[0].InstanceID != bottom {
		t.Fatalf("opponent library %d cards, want only the bottom one", len(opp.Library.Cards))
	}
	if len(me.Library.Cards) < 2 {
		t.Fatal("the controller's own library was exiled")
	}
}

// The −3: until your next turn, each creature that attacks you or a
// planeswalker you control gets -5/-0 — and only those.
func TestJaceRealitySculptorMinusThreeShrinksAttackersUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	jrs := pushJaceRealitySculptor(g, me.ID)
	b16Activate(t, g, me.ID, jrs, 1, game.ActivateAbilityParams{})

	advanceToNextSeatsTurn(t, g)
	opp := g.Seats[g.Turn.ActiveSeat]
	brute := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Brute", TypeLine: "Creature — Ogre",
		Power: 6, Toughness: 6, Owner: opp.ID, Controller: opp.ID,
	})
	bystander := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if bystander.ID == me.ID {
		bystander = g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	}
	other := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Other Brute", TypeLine: "Creature — Ogre",
		Power: 6, Toughness: 6, Owner: opp.ID, Controller: opp.ID,
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(brute, me.ID); err != nil {
		t.Fatalf("attack me: %v", err)
	}
	if err := g.DeclareAttacker(other, bystander.ID); err != nil {
		t.Fatalf("attack someone else: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, brute); got != 1 {
		t.Fatalf("attacker power %d, want 6-5", got)
	}
	if got := effectivePower(t, g, other); got != 6 {
		t.Fatalf("a creature attacking another player went to power %d", got)
	}

	// "Until your next turn": the trigger is gone once my turn begins.
	for g.Seats[g.Turn.ActiveSeat].ID != me.ID {
		advanceToNextSeatsTurn(t, g)
	}
	for _, dt := range g.DelayedTriggers {
		if dt.Controller == me.ID && dt.Repeats {
			t.Fatalf("the −3's trigger outlived the start of my next turn: %+v", dt.Label)
		}
	}
}
