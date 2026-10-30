package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b7_cards_test.go — Marvel Super Heroes, half B, seventh batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB7DeathlokOracle     = "d89aa87a-d59a-41e2-b92f-15c9bc6d3d13"
	mshB7PymOracle          = "636b44dc-f7e4-42d2-abf6-ae49432b6052"
	mshB7QuicksilverOracle  = "b76ce131-e51a-445e-8216-00df5904ab93"
	mshB7RapidRescueOracle  = "de1ad369-5aa3-4c54-8fb1-30b7f1d4d542"
	mshB7RedHulkOracle      = "a14a838a-49f4-45e3-9ec0-3c803af45a68"
	mshB7ReptilOracle       = "c6284597-1983-4a21-a461-724135220e0d"
	mshB7RewriteOracle      = "f91c860e-3feb-4a5a-b7ef-4e7d38d61c75"
	mshB7RobotDomOracle     = "4f69d2f9-c16a-4752-9026-794a29d73af8"
	mshB7RoxxonOracle       = "b92c2da9-eaa4-404b-8114-5527eec6390e"
	mshB7FlyingCarOracle    = "12d51534-a98d-487b-9ba9-c10861dff992"
	mshB7SpyKitOracle       = "6fd44fb9-3433-4cb0-95a4-4ac8170b2166"
	mshB7ScientistOracle    = "c09c424b-6aba-4190-9f55-f0307d746386"
	mshB7SerpentOracle      = "556f127f-6131-4171-b1ed-11420c21d345"
	mshB7QuicksilverTypes   = "Legendary Creature — Mutant Hero"
	mshB7ReptilTypes        = "Legendary Creature — Human Hero"
	mshB7RedHulkTypes       = "Legendary Creature — Gamma Berserker Villain"
	mshB7RoxxonTypes        = "Creature — Human Berserker Villain"
	mshB7ScientistTypes     = "Legendary Creature — Human Scientist Villain"
	mshB7DeathlokTypes      = "Artifact Creature — Zombie Soldier"
	mshB7SerpentTypes       = "Creature — Human Snake Villain"
	mshB7PlanTypes          = "Enchantment — Plan"
	mshB7RewriteHistoryText = "Rewrite History"
)

func TestMshB7EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB7DeathlokOracle:    "Project Deathlok Soldier",
		mshB7PymOracle:         "Pym Particles",
		mshB7QuicksilverOracle: "Quicksilver, Brash Blur",
		mshB7RapidRescueOracle: "Rapid Rescue",
		mshB7RedHulkOracle:     "Red Hulk",
		mshB7ReptilOracle:      "Reptil, Dinomorpher",
		mshB7RewriteOracle:     "Rewrite History",
		mshB7RobotDomOracle:    "Robot Domination",
		mshB7RoxxonOracle:      "Roxxon Brutes",
		mshB7FlyingCarOracle:   "S.H.I.E.L.D. Flying Car",
		mshB7SpyKitOracle:      "S.H.I.E.L.D. Spy Kit",
		mshB7ScientistOracle:   "Scientist Supreme of A.I.M.",
		mshB7SerpentOracle:     "Serpent Specialist",
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

func mshB7CardRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// --- Project Deathlok Soldier -----------------------------------------------

func TestMshB7DeathlokReturnsItselfFromTheGraveyardToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := pushCatalogGraveyardCard(me, "Project Deathlok Soldier", mshB7DeathlokTypes, mshB7DeathlokOracle, 1, 2)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}{C}{B}", nil); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
		t.Errorf("in hand %v, in graveyard %v; want it back in hand", me.Hand.Contains(id), me.Graveyard.Contains(id))
	}
}

func TestMshB7DeathlokCannotActivateFromTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Project Deathlok Soldier", mshB7DeathlokTypes, mshB7DeathlokOracle, 1, 2)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}{C}{B}", nil); err == nil {
		t.Error("the return ability was activated from the battlefield")
	}
}

// --- Pym Particles ------------------------------------------------------------

func TestMshB7PymParticlesGivesVigilanceUnblockableAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)

	castCatalogSpell(t, g, "Pym Particles", "Sorcery", mshB7PymOracle, []game.TargetRef{mshB7CardRef(bear)})
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, hand+1)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the Bear did not gain vigilance")
	}
	brAttack(t, g, bear)
	brRefusal(t, g.DeclareBlocker(wall, bear), game.BlockReasonCantBeBlocked)
}

// --- Quicksilver, Brash Blur ---------------------------------------------------

func TestMshB7QuicksilverMayBeginTheGameOnTheBattlefieldAndHasHaste(t *testing.T) {
	spec, ok := Lookup(mshB7QuicksilverOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.OpeningHand == nil {
		t.Error("no opening-hand clause")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	powerUpMain(t, g)
	id := pushPowerUpPermanent(g, me.ID, "Quicksilver, Brash Blur", mshB7QuicksilverTypes, "{R}", mshB7QuicksilverOracle)
	if !hasEffectiveKeyword(t, g, id, "haste") {
		t.Error("Quicksilver lacks haste")
	}
}

func TestMshB7QuicksilverPowerUpPutsACounterAndADoubleStrikeCounterOnce(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	qs := castPowerUpCreature(t, g, me, "Quicksilver, Brash Blur", mshB7QuicksilverTypes, "{R}", "{R}", mshB7QuicksilverOracle, 1)
	if got := chargedPowerUp(t, g, me.ID, qs); got != "{4}" {
		t.Fatalf("the turn he entered: %q, want {4}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, qs, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, qs); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if got := countersOn(g, qs, game.CounterDoubleStrike); got != 1 {
		t.Errorf("double strike counters = %d, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, qs, "double strike") {
		t.Error("the double strike counter did not give double strike")
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, qs, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

// --- Rapid Rescue ---------------------------------------------------------------

// mshB7SeedMill replaces the library with `top` (first card on top) over one
// filler card, and returns the ids of `top` in order.
func mshB7SeedMill(me *game.Player, top ...game.Card) []uuid.UUID {
	cards := append(append([]game.Card{}, top...), game.Card{Name: "Filler", TypeLine: "Sorcery"})
	return seedSearchLibrary(me, cards...)[:len(top)]
}

func TestMshB7RapidRescueTakesAMilledPermanentAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	old := game.Card{InstanceID: uuid.New(), Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(old)
	ids := mshB7SeedMill(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	bolt, bear := ids[0], ids[1]
	life := me.Life

	castCatalogSpell(t, g, "Rapid Rescue", "Instant", mshB7RapidRescueOracle, nil)
	passPriorityAroundTable(t, g)

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the milled permanent card")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != bear {
		t.Fatalf("offered %v, want only the milled Bear (not the Bolt %v or the old card %v)", pick.ChooseCards, bolt, old.InstanceID)
	}
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(bear) {
		t.Error("the Bear is not in hand")
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the Bolt should stay in the graveyard")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestMshB7RapidRescueAsksNothingWhenNoMilledCardIsAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mshB7SeedMill(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Deep", TypeLine: "Sorcery"},
	)
	life := me.Life
	castCatalogSpell(t, g, "Rapid Rescue", "Instant", mshB7RapidRescueOracle, nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a prompt opened with no permanent card among the milled cards")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

// --- Red Hulk ---------------------------------------------------------------------

func mshB7Hurt(t *testing.T, g *game.Game, source, target uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(source, target, n) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
}

func TestMshB7RedHulkGrowsAndDealsDamageEqualToItsCountersToAnotherTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulk := b12Push(g, me.ID, "Red Hulk", mshB7RedHulkTypes, mshB7RedHulkOracle, 6, 7)
	src := b12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)
	victim := b12Creature(g, opp.ID, "Troll", "Creature — Troll", 3, 9)

	mshB7Hurt(t, g, src, hulk, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the damage")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, mshB7CardRef(hulk)); err == nil {
		t.Error("Red Hulk may target himself; it is \"any other target\"")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hulk); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
	if got := damageMarkedOn(g, victim); got != 1 {
		t.Errorf("first hit: victim has %d damage, want 1 (one counter)", got)
	}

	mshB7Hurt(t, g, src, hulk, 1)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hulk); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}
	if got := damageMarkedOn(g, victim); got != 3 {
		t.Errorf("second hit: victim has %d damage, want 3 (1 + 2)", got)
	}
	for _, kw := range []string{"reach", "trample"} {
		if !hasEffectiveKeyword(t, g, hulk, kw) {
			t.Errorf("Red Hulk lacks %s", kw)
		}
	}
}

// --- Reptil, Dinomorpher ----------------------------------------------------------

func TestMshB7ReptilBecomesABrontosaurusUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Reptil, Dinomorpher", mshB7ReptilTypes, mshB7ReptilOracle, 1, 2)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("Brontosaurus: %v", err)
	}
	passPriorityAroundTable(t, g)

	if p, tough := effectivePTOf(t, g, id); p != 3 || tough != 5 {
		t.Errorf("Reptil is %d/%d, want 3/5", p, tough)
	}
	c := ringBFCard(t, g, id)
	if !c.HasSubtype("Dinosaur") || !c.HasSubtype("Hero") || c.HasSubtype("Human") {
		t.Errorf("types %q: want Dinosaur Hero and no longer Human", c.TypeLine)
	}
	for _, kw := range []string{"reach", "vigilance"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Reptil lacks %s", kw)
		}
	}
	if hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("the Brontosaurus form gave trample")
	}

	if err := g.PassTurn(); err != nil {
		t.Fatal(err)
	}
	if p, tough := effectivePTOf(t, g, id); p != 1 || tough != 2 {
		t.Errorf("after the turn Reptil is %d/%d, want 1/2", p, tough)
	}
	if !ringBFCard(t, g, id).HasSubtype("Human") {
		t.Error("Reptil did not become a Human again")
	}
}

func TestMshB7ReptilBecomesATyrannosaurusRex(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Reptil, Dinomorpher", mshB7ReptilTypes, mshB7ReptilOracle, 1, 2)
	if err := hobB4Activate(t, g, me.ID, id, 1, "{C}{C}{C}{C}{C}{C}", nil); err != nil {
		t.Fatalf("Tyrannosaurus Rex: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, id); p != 6 || tough != 6 {
		t.Errorf("Reptil is %d/%d, want 6/6", p, tough)
	}
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("Reptil lacks trample")
	}
	if hasEffectiveKeyword(t, g, id, "reach") {
		t.Error("the Tyrannosaurus form gave reach")
	}
}

// --- Rewrite History ----------------------------------------------------------------

func TestMshB7RewriteHistoryLootsAndAddsAPlanCounterOncePerBatchOfTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plan := b12Push(g, me.ID, mshB7RewriteHistoryText, mshB7PlanTypes, mshB7RewriteOracle, 0, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()

	tapForTest(t, g, a)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (drew one and discarded one)", got, hand)
	}
	if got := counterCount(g, plan, "plan"); got != 1 {
		t.Errorf("%d plan counters, want 1", got)
	}
	opp := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	tapForTest(t, g, opp)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, plan, "plan"); got != 1 {
		t.Errorf("an opponent's creature tapping made a plan counter (%d)", got)
	}

}

func TestMshB7RewriteHistoryFourthCounterSacrificesAndReturnsUpToTwoSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plan := b12Push(g, me.ID, mshB7RewriteHistoryText, mshB7PlanTypes, mshB7RewriteOracle, 0, 0)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	bolt := pushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
	ritual := pushGraveyardPermanent(me, "Ritual", "Sorcery", "{B}")
	other := pushGraveyardPermanent(me, "Ring", "Artifact", "{1}")
	third := pushGraveyardPermanent(me, "Fireball", "Sorcery", "{X}{R}")
	if err := g.AddCounterBy(me.ID, plan, "plan", 3); err != nil {
		t.Fatal(err)
	}

	tapForTest(t, g, a)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(plan) {
		t.Fatal("Rewrite History was not sacrificed on its fourth plan counter")
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the returned spells")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, mshB7CardRef(other)); err == nil {
		t.Error("an artifact card is not an instant or sorcery card")
	}
	three := []game.TargetRef{mshB7CardRef(bolt), mshB7CardRef(ritual), mshB7CardRef(third)}
	if err := g.ResolvePickTargets(pick.ID, me.ID, three); err == nil {
		t.Error("three targets were accepted; the limit is two")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{mshB7CardRef(bolt), mshB7CardRef(ritual)}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(bolt) || !me.Hand.Contains(ritual) {
		t.Errorf("Bolt in hand %v, Ritual in hand %v; want both", me.Hand.Contains(bolt), me.Hand.Contains(ritual))
	}
	if me.Hand.Contains(third) || !me.Graveyard.Contains(third) {
		t.Error("a third spell came back; the limit is two")
	}
}

// --- Robot Domination -----------------------------------------------------------------

func TestMshB7RobotDominationDrawsLosesLifeAndCountsOnYourCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	plan := b12Push(g, me.ID, "Robot Domination", mshB7PlanTypes, mshB7RobotDomOracle, 0, 0)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life

	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, plan, "plan"); got != 0 {
		t.Fatalf("an opponent's creature card made %d plan counters, want 0", got)
	}

	killCreature(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, plan, "plan"); got != 1 {
		t.Errorf("%d plan counters, want 1", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestMshB7RobotDominationThirdCounterSacrificesItAndMakesThreeRobots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	plan := b12Push(g, me.ID, "Robot Domination", mshB7PlanTypes, mshB7RobotDomOracle, 0, 0)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := g.AddCounterBy(me.ID, plan, "plan", 2); err != nil {
		t.Fatal(err)
	}

	killCreature(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(plan) {
		t.Error("Robot Domination was not sacrificed on its third plan counter")
	}
	if got := mshB1Tokens(g, me.ID, "Robot"); got != 3 {
		t.Fatalf("%d Robot tokens, want 3", got)
	}
	tok := mshB1Token(t, g, me.ID, "Robot")
	if tok.Power != 2 || tok.Toughness != 2 || !tok.IsArtifact() || !mshB1HasType(tok, "Villain") || len(tok.Colors) != 0 {
		t.Errorf("token %q %d/%d colors %v, want a colorless 2/2 artifact Robot Villain", tok.TypeLine, tok.Power, tok.Toughness, tok.Colors)
	}
}

// --- Roxxon Brutes ----------------------------------------------------------------------

func TestMshB7RoxxonBrutesPutsACounterOnTargetCreatureWhenYouDrawYourSecondCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Roxxon Brutes", mshB7RoxxonTypes, mshB7RoxxonOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	pick := mshB7CardRef(bear)
	hobA5Draw(t, g, me, &pick)
	if got := plusOnes(g, bear); got != 0 {
		t.Fatalf("the first draw made %d counters, want 0", got)
	}
	hobA5Draw(t, g, me, &pick)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("the second draw made %d counters, want 1", got)
	}
	hobA5Draw(t, g, me, &pick)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("the third draw made %d counters, want still 1", got)
	}
}

func TestMshB7RoxxonBrutesBasicLandcyclingFindsABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Roxxon Brutes", mshB7RoxxonTypes, mshB7RoxxonOracle)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("landcycling: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || !me.Graveyard.Contains(id) {
		t.Errorf("Forest in hand %v, Brutes in graveyard %v", me.Hand.Contains(ids[1]), me.Graveyard.Contains(id))
	}
}

// --- S.H.I.E.L.D. Flying Car -----------------------------------------------------------------

func TestMshB7FlyingCarFlickersYourCreatureAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "S.H.I.E.L.D. Flying Car", "Artifact — Vehicle", mshB7FlyingCarOracle, nil)
	passPriorityAroundTable(t, g)
	mshB2PickIfAsked(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear is still on the battlefield after the exile")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	// It returns as a new object (CR 400.7), so look it up by name.
	if findBattlefieldByName(g, "Bear") == uuid.Nil {
		t.Error("the Bear did not return at the beginning of the next end step")
	}
}

// --- S.H.I.E.L.D. Spy Kit -----------------------------------------------------------------------

func TestMshB7SpyKitPumpsAndUntapsAndScriesOnlyForALoneAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	kit := pushCatalogPermanent(g, me.ID, "S.H.I.E.L.D. Spy Kit", "Artifact — Equipment", mshB7SpyKitOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, kit, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Fatalf("equipped Bear is %d/%d, want 3/3", p, tough)
	}

	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	scry := scryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if ringBFCard(t, g, bear).Tapped {
		t.Error("the lone attacker was not untapped")
	}
}

func TestMshB7SpyKitDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	kit := pushCatalogPermanent(g, me.ID, "S.H.I.E.L.D. Spy Kit", "Artifact — Equipment", mshB7SpyKitOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, kit, bear)

	declareAttack(t, g, opp.ID, bear, other)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("the kit scried for a non-lone attacker")
	}
	if !ringBFCard(t, g, bear).Tapped {
		t.Error("the attacker was untapped though it did not attack alone")
	}
}

// --- Scientist Supreme of A.I.M. --------------------------------------------------------------------

func TestMshB7ScientistCopiesOnlyYourArtifactAbilitiesPaysLifeAndIsOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sci := pushCatalogPermanent(g, me.ID, "Scientist Supreme of A.I.M.", mshB7ScientistTypes, mshB7ScientistOracle, false)
	rock := pushCatalogPermanent(g, me.ID, "Source Rock", "Artifact", "", false)
	bear := pushCatalogPermanent(g, me.ID, "Source Bear", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)

	for _, a := range []struct {
		who  uuid.UUID
		src  uuid.UUID
		name string
	}{{me.ID, rock, "artifact trigger"}, {me.ID, bear, "creature trigger"}, {opp.ID, theirs, "their trigger"}} {
		if err := g.AnnounceTrigger(a.who, a.src, game.AbilityParams{Label: a.name}); err != nil {
			t.Fatalf("AnnounceTrigger %s: %v", a.name, err)
		}
	}
	trig := acItemLabelled(g, "artifact trigger")
	offered := acLegalAbilityTargets(g, me.ID, mshB7ScientistOracle, 0)
	if !offered[trig] {
		t.Errorf("an ability of yours from an artifact is not offered: %v", offered)
	}
	if offered[acItemLabelled(g, "creature trigger")] || offered[acItemLabelled(g, "their trigger")] {
		t.Errorf("a creature source or an opponent's ability is offered: %v", offered)
	}

	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, sci, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{mshB7CardRef(trig)}}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d (2 paid)", me.Life, life-2)
	}
	acPassUntilCopy(t, g)
	if n := acCopiesOnStack(g); n != 1 {
		t.Fatalf("%d copies, want 1", n)
	}
	if err := g.ActivateCatalogAbility(me.ID, sci, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{mshB7CardRef(trig)}}); err == nil {
		t.Error("a second activation in the same turn was accepted")
	}
}

func TestMshB7ScientistCannotBeActivatedOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	sci := pushCatalogPermanent(g, opp.ID, "Scientist Supreme of A.I.M.", mshB7ScientistTypes, mshB7ScientistOracle, false)
	rock := pushCatalogPermanent(g, opp.ID, "Source Rock", "Artifact", "", false)
	if err := g.AnnounceTrigger(opp.ID, rock, game.AbilityParams{Label: "artifact trigger"}); err != nil {
		t.Fatal(err)
	}
	trig := acItemLabelled(g, "artifact trigger")
	if err := g.ActivateCatalogAbility(opp.ID, sci, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{mshB7CardRef(trig)}}); err == nil {
		t.Error("the ability was activated during another player's turn")
	}
}

// --- Serpent Specialist -----------------------------------------------------------------------------

func TestMshB7SerpentSpecialistPowerUpPutsTwoCountersOnce(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	sp := castPowerUpCreature(t, g, me, "Serpent Specialist", mshB7SerpentTypes, "{G}", "{G}", mshB7SerpentOracle, 1)
	if !hasEffectiveKeyword(t, g, sp, "deathtouch") {
		t.Error("Serpent Specialist lacks deathtouch")
	}
	if got := chargedPowerUp(t, g, me.ID, sp); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sp, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, sp); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sp, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}
