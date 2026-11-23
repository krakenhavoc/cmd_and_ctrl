package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b1_cards_test.go — Edge of Eternities leftovers, batches 1 and 2
// (#4461). Each test plays the real catalog entry and asserts what its
// text says.

const (
	stdB1DroneOracle       = "aa704b26-729a-4d0f-b92d-56aadb1e20b9"
	stdB1HoundOracle       = "ae791b61-5825-4987-a8c2-80813758d3f7"
	stdB1RiddlerOracle     = "58582fac-4c42-4a1c-9ab1-4a892b500da4"
	stdB1InquisitorOracle  = "57012049-06d7-4354-8125-9ea9e742f14d"
	stdB1PulsarOracle      = "8019e879-cfea-40a2-ac0d-86b47ec02b48"
	stdB1RaybladeOracle    = "792da586-a687-428c-a10e-75431249df5b"
	stdB1RerouteOracle     = "26d2762b-4346-40d0-b93d-b0dbfeb6e2bc"
	stdB1SkiffOracle       = "112aaaf0-3301-4f79-9640-d36c75a0fb30"
	stdB1DownloadOracle    = "f27889b9-40d9-4cc7-b8af-8716c53c44f5"
	stdB1CryoOracle        = "865715e9-6759-4e01-803d-702c959e386e"
	stdB1BlastOracle       = "2d80ee9c-55d4-4115-8a58-63e8460f9795"
	stdB1DivertOracle      = "76fdbc2c-d2b1-466f-9230-06e187bcc9e6"
	stdB1ShieldmateOracle  = "d0025d29-d652-4af8-b367-13142d5532df"
	stdB1MouthOracle       = "832fbb1f-0ff5-4d6c-b1ee-2231923a6f64"
	stdB1SentinelOracle    = "07155470-c5ab-41ae-ade3-bdc06d357379"
	stdB1SelfcraftOracle   = "4ba4cb4b-f421-4abd-8fcf-a01174447790"
	stdB1OperatorOracle    = "770b2097-5178-42d2-be6c-393b4bfe3727"
	stdB1WeftwalkingOracle = "08af6751-5848-47fb-95f5-57215e899460"
	stdB1CrawlerOracle     = "0c1b6305-8863-43e0-b50d-ae1c5e31be55"
	stdB1FeederOracle      = "de70a6bb-32ed-4a89-a522-c2a71f08fb02"
)

func stdB1Hand(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

func stdB1Draw(g *game.Game, p uuid.UUID, n int) error {
	var err error
	g.WithWriteLock(func() { err = g.DrawNForEffect(p, n) })
	return err
}

func stdB1EmptyHand(p *game.Player) {
	for p.Hand.Size() > 0 {
		p.Hand.PopTop()
	}
}

func TestStdB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB1DroneOracle:       "Dockworker Drone",
		stdB1HoundOracle:       "Codecracker Hound",
		stdB1RiddlerOracle:     "Quantum Riddler",
		stdB1InquisitorOracle:  "Lightstall Inquisitor",
		stdB1PulsarOracle:      "Pulsar Squadron Ace",
		stdB1RaybladeOracle:    "Rayblade Trooper",
		stdB1RerouteOracle:     "Reroute Systems",
		stdB1SkiffOracle:       "Rescue Skiff",
		stdB1DownloadOracle:    "Cerebral Download",
		stdB1CryoOracle:        "Cryoshatter",
		stdB1BlastOracle:       "Desculpting Blast",
		stdB1DivertOracle:      "Divert Disaster",
		stdB1ShieldmateOracle:  "Mechan Shieldmate",
		stdB1MouthOracle:       "Mouth of the Storm",
		stdB1SentinelOracle:    "Nanoform Sentinel",
		stdB1SelfcraftOracle:   "Selfcraft Mechan",
		stdB1OperatorOracle:    "Steelswarm Operator",
		stdB1WeftwalkingOracle: "Weftwalking",
		stdB1CrawlerOracle:     "Comet Crawler",
		stdB1FeederOracle:      "Monoist Circuit-Feeder",
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

// --- Codecracker Hound ---

func TestStdB1CodecrackerHoundTakesOneOfTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.SeedLibrary(me, "One", "Two", "Three")
	effectstest.B12PlayFromHand(t, g, "Codecracker Hound", "Creature — Dog", stdB1HoundOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCards(g, me.ID)
	if c == nil || len(c.ChooseCards) != 2 {
		t.Fatalf("want a prompt over the top two, got %+v", c)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[1]) || !me.Graveyard.Contains(top[0]) || !me.Library.Contains(top[2]) {
		t.Error("picked card to hand, the other to the graveyard, the third untouched")
	}
}

// --- Dockworker Drone ---

func TestStdB1DockworkerDroneEntersWithACounterAndHandsItOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	drone := effectstest.B12PlayFromHand(t, g, "Dockworker Drone", "Artifact Creature — Robot", stdB1DroneOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, drone, game.CounterPlusOne); got != 1 {
		t.Fatalf("enters with %d counters, want 1", got)
	}
	effectstest.KillCreature(t, g, me.ID, drone)
	effectstest.EoeABPickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the bear has %d counters, want the Drone's 1", got)
	}
}

// --- Quantum Riddler ---

func TestStdB1QuantumRiddlerDrawsOneMoreWithASmallHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Quantum Riddler", "Creature — Sphinx", stdB1RiddlerOracle, 4, 6)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	stdB1EmptyHand(me)
	if err := stdB1Draw(g, me.ID, 2); err != nil {
		t.Fatal(err)
	}
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("drew to %d cards, want 3 (two plus one)", got)
	}
	// hand is now 3: no bonus
	if err := stdB1Draw(g, me.ID, 2); err != nil {
		t.Fatal(err)
	}
	if got := me.Hand.Size(); got != 5 {
		t.Errorf("hand %d, want 5: no bonus with three cards in hand", got)
	}
}

func TestStdB1QuantumRiddlerEntersAndDrawsTwoFromAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToMain(t, g)
	stdB1EmptyHand(me)
	// The Riddler is in hand while it is cast, so one card is in hand
	// before it resolves; once on the battlefield the ETB draw is
	// "one or more", so it draws two.
	effectstest.B12PlayFromHand(t, g, "Quantum Riddler", "Creature — Sphinx", stdB1RiddlerOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2: the enters draw is a single draw plus one", got)
	}
}

func TestStdB1QuantumRiddlerTwoOfThemAddTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Quantum Riddler", "Creature — Sphinx", stdB1RiddlerOracle, 4, 6)
	effectstest.B12Push(g, me.ID, "Quantum Riddler", "Creature — Sphinx", stdB1RiddlerOracle, 4, 6)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	stdB1EmptyHand(me)
	if err := stdB1Draw(g, me.ID, 1); err != nil {
		t.Fatal(err)
	}
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3: each Riddler adds one", got)
	}
}

func stdB1Tgt(ids ...uuid.UUID) []game.TargetRef { return effectstest.CardRefs(ids...) }

// --- Reroute Systems ---

func TestStdB1RerouteSystemsGrantsIndestructibleToAnArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastModal(t, g, "Reroute Systems", "Instant", stdB1RerouteOracle, []int{0}, stdB1Tgt(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, rock, "indestructible") {
		t.Error("the artifact is not indestructible")
	}
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	if _, err := effectstest.CastModalErr(t, g, "Reroute Systems", "Instant", stdB1RerouteOracle, []int{0}, stdB1Tgt(land)); err == nil {
		t.Error("a land was accepted as a target")
	}
}

func TestStdB1RerouteSystemsDealsTwoToATappedCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tapped := effectstest.B12Creature(g, opp.ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, tapped)
	effectstest.AdvanceToMain(t, g)
	if _, err := effectstest.CastModalErr(t, g, "Reroute Systems", "Instant", stdB1RerouteOracle, []int{1}, stdB1Tgt(untapped)); err == nil {
		t.Error("an untapped creature was accepted as a target")
	}
	effectstest.CastModal(t, g, "Reroute Systems", "Instant", stdB1RerouteOracle, []int{1}, stdB1Tgt(tapped))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped 2/2 survived 2 damage")
	}
	_ = me
}

// --- Cerebral Download ---

func TestStdB1CerebralDownloadSurveilsXThenDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Other Rock", "Artifact Creature — Robot")
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	hand := me.Hand.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Cerebral Download", "Instant", stdB1DownloadOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("want a surveil of 2 (two artifacts), got %+v", c)
	}
	if me.Hand.Size() != hand {
		t.Fatal("drew before the surveil was answered")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ids[0]) {
		t.Error("the surveilled card is not in the graveyard")
	}
	if me.Hand.Size() != hand+3 || !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[2]) || !me.Hand.Contains(ids[3]) {
		t.Errorf("hand %d (want %d): draws the kept card and the next two", me.Hand.Size(), hand+3)
	}
}

func TestStdB1CerebralDownloadWithNoArtifactsJustDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Cerebral Download", "Instant", stdB1DownloadOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+3)
	}
}

// --- Desculpting Blast ---

func TestStdB1DesculptingBlastMakesADroneOnlyForAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.CastFromSeat(t, g, me, "Desculpting Blast", "Instant", stdB1BlastOracle, stdB1Tgt(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Fatal("the attacker was not returned to its owner's hand")
	}
	if effectstest.B16CountNamed(g, "Drone") != 1 {
		t.Fatalf("want one Drone, have %d", effectstest.B16CountNamed(g, "Drone"))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Drone" {
			if c.Controller != me.ID || c.Power != 1 || c.Toughness != 1 || !c.IsArtifact() || !c.IsCreature() {
				t.Errorf("Drone is %+v", c)
			}
			effectstest.AssertKeywords(t, g, c.InstanceID, "flying")
		}
	}
}

func TestStdB1DesculptingBlastMakesNoDroneForAnIdlePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Desculpting Blast", "Instant", stdB1BlastOracle, stdB1Tgt(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Drone") != 0 || !me.Hand.Contains(rock) {
		t.Error("an idle permanent was bounced with a Drone, or not bounced")
	}
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	if _, err := effectstest.CastModalErr(t, g, "Desculpting Blast", "Instant", stdB1BlastOracle, nil, stdB1Tgt(land)); err == nil {
		t.Error("a land was accepted as a target")
	}
}

// --- Cryoshatter ---

func stdB1Cryo(t *testing.T, g *game.Game, host uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12PlayFromHand(t, g, "Cryoshatter", "Enchantment — Aura", stdB1CryoOracle, game.CastSpellParams{Targets: stdB1Tgt(host)})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestStdB1CryoshatterShrinksAndDestroysWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 6, 6)
	stdB1Cryo(t, g, bear)
	if p := effectstest.LayeredCard(t, g, bear).CurrentPower(); p != 1 {
		t.Errorf("power %d, want 1 (-5/-0)", p)
	}
	effectstest.TapForTest(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the tapped creature was not destroyed")
	}
}

func TestStdB1CryoshatterDestroysWhenDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 6, 6)
	stdB1Cryo(t, g, bear)
	_ = me
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the damaged creature was not destroyed")
	}
}

// --- Divert Disaster ---

// stdB1DivertAtTheirBear has seat 0 cast a creature spell and seat 1 answer
// with Divert Disaster, then passes until seat 0 is asked for the {2}.
func stdB1DivertAtTheirBear(t *testing.T, g *game.Game) (victim uuid.UUID) {
	t.Helper()
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim = effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, them, "Divert Disaster", "Instant", stdB1DivertOracle, stdB1Tgt(victim))
	effectstest.PassUntilTaxed(t, g, me.ID)
	return victim
}

func TestStdB1DivertDisasterMakesALanderForTheCasterWhenTheyPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	victim := stdB1DivertAtTheirBear(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, victim) {
		t.Error("a paid spell must resolve")
	}
	if effectstest.B16CountNamed(g, "Lander") != 1 {
		t.Fatalf("want one Lander, have %d", effectstest.B16CountNamed(g, "Lander"))
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Lander" && c.Controller != them.ID {
			t.Errorf("the Lander belongs to %s, want the Divert Disaster caster", c.Controller)
		}
	}
}

func TestStdB1DivertDisasterCountersAnUnpaidSpellWithNoLander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	victim := stdB1DivertAtTheirBear(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, victim) || g.Stack.Contains(victim) {
		t.Error("an unpaid spell was not countered")
	}
	if effectstest.B16CountNamed(g, "Lander") != 0 {
		t.Error("a Lander was made without the payment")
	}
}

// --- Mechan Shieldmate ---

func TestStdB1MechanShieldmateAttacksOnlyAfterAnArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mate := effectstest.B12Push(g, me.ID, "Mechan Shieldmate", "Artifact Creature — Robot Soldier", stdB1ShieldmateOracle, 3, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mate, opp.ID); err == nil {
		t.Fatal("a defender attacked with no artifact having entered")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	mate2 := effectstest.B12Push(g2, me2.ID, "Mechan Shieldmate", "Artifact Creature — Robot Soldier", stdB1ShieldmateOracle, 3, 2)
	effectstest.AdvanceToMain(t, g2)
	effectstest.B12PlayFromHand(t, g2, "Rock", "Artifact", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(mate2, opp2.ID); err != nil {
		t.Fatalf("after an artifact entered, the Shieldmate could not attack: %v", err)
	}
	effectstest.AssertKeywords(t, g2, mate2, "defender")
}

// --- Mouth of the Storm ---

func TestStdB1MouthOfTheStormShrinksOpposingCreaturesUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12PlayFromHand(t, g, "Mouth of the Storm", "Creature — Elemental", stdB1MouthOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LayeredCard(t, g, theirs).CurrentPower(); p != 2 {
		t.Errorf("their power %d, want 2 (-3/-0)", p)
	}
	if p := effectstest.LayeredCard(t, g, mine).CurrentPower(); p != 2 {
		t.Errorf("my power %d: your own creatures are not shrunk", p)
	}
	late := effectstest.B12Creature(g, opp.ID, "Late Ogre", "Creature — Ogre", 4, 4)
	if p := effectstest.LayeredCard(t, g, late).CurrentPower(); p != 4 {
		t.Errorf("a creature that entered later has power %d, want its own 4", p)
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.LayeredCard(t, g, theirs).CurrentPower(); p != 2 {
		t.Errorf("power %d on their turn, want the shrink to last until my next turn", p)
	}
	effectstest.AdvanceToMainOf(t, g, 0)
	if p := effectstest.LayeredCard(t, g, theirs).CurrentPower(); p != 5 {
		t.Errorf("power %d on my next turn, want 5 restored", p)
	}
}

// --- Nanoform Sentinel ---

func TestStdB1NanoformSentinelUntapsAnotherPermanentOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sentinel := effectstest.B12Push(g, me.ID, "Nanoform Sentinel", "Artifact Creature — Robot", stdB1SentinelOracle, 3, 2)
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	land2 := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.AdvanceToMain(t, g)
	effectstest.TapForTest(t, g, land, land2)
	effectstest.TapForTest(t, g, sentinel)
	effectstest.AnswerPickTargetIfOpen(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, land) {
		t.Error("the targeted land was not untapped")
	}
	if !effectstest.B16Tapped(t, g, land2) {
		t.Error("an untargeted land was untapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(sentinel) })
	effectstest.TapForTest(t, g, land)
	effectstest.TapForTest(t, g, sentinel)
	effectstest.AnswerPickTargetIfOpen(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the ability triggered a second time in the same turn")
	}
}

// --- Selfcraft Mechan ---

func TestStdB1SelfcraftMechanSacrificesAnArtifactForACounterAndACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.B12PlayFromHand(t, g, "Selfcraft Mechan", "Artifact Creature — Robot Artificer", stdB1SelfcraftOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.AnswerChooseCardsOrPick(t, g, me.ID, rock)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the bear has %d counters, want 1", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: a card drawn", me.Hand.Size(), hand+1)
	}
}

func TestStdB1SelfcraftMechanDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.B12PlayFromHand(t, g, "Selfcraft Mechan", "Artifact Creature — Robot Artificer", stdB1SelfcraftOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 0 || me.Hand.Size() != hand {
		t.Error("declining the sacrifice still did something")
	}
}

// --- Steelswarm Operator ---

func TestStdB1SteelswarmOperatorPaysOnlyForArtifactSpellsAndArtifactAbilities(t *testing.T) {
	art := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{U}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{U}"}
	one, _ := game.ParseCost("{U}")
	two, _ := game.ParseCost("{U}{U}")
	for _, tc := range []struct {
		ability int
		cost    game.ParsedCost
		ok      []game.ManaSpendContext
		no      []game.ManaSpendContext
	}{
		{0, one, []game.ManaSpendContext{game.ManaSpendForCast(art)},
			[]game.ManaSpendContext{game.ManaSpendForCast(bear), game.ManaSpendForAbility(art)}},
		{1, two, []game.ManaSpendContext{game.ManaSpendForAbility(art)},
			[]game.ManaSpendContext{game.ManaSpendForAbility(bear), game.ManaSpendForCast(art)}},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		op := effectstest.B12Push(g, me.ID, "Steelswarm Operator", "Artifact Creature — Robot Soldier", stdB1OperatorOracle, 1, 1)
		if err := g.ActivateManaAbility(me.ID, op, tc.ability, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ability %d: %v", tc.ability, err)
		}
		for _, ctx := range tc.ok {
			if !me.ManaPool.CanPayFor(tc.cost, 0, ctx) {
				t.Errorf("ability %d refused a legal spend %+v", tc.ability, ctx)
			}
		}
		for _, ctx := range tc.no {
			if me.ManaPool.CanPayFor(tc.cost, 0, ctx) {
				t.Errorf("ability %d paid for a spend it must refuse %+v", tc.ability, ctx)
			}
		}
	}
}

// --- Comet Crawler ---

func TestStdB1CometCrawlerSacrificesForPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crawler := effectstest.B12Push(g, me.ID, "Comet Crawler", "Creature — Insect Horror", stdB1CrawlerOracle, 2, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.DeclareAttack(t, g, opp.ID, crawler)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrPick(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if p := effectstest.LayeredCard(t, g, crawler).CurrentPower(); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
}

// --- Monoist Circuit-Feeder ---

func TestStdB1MonoistCircuitFeederScalesBothHalvesWithArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Rock Two", "Artifact")
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 6)
	effectstest.B12PlayFromHand(t, g, "Monoist Circuit-Feeder", "Artifact Creature — Nautilus", stdB1FeederOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, mine)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Rocks and the Circuit-Feeder itself: X = 3.
	if p := effectstest.LayeredCard(t, g, mine).CurrentPower(); p != 5 {
		t.Errorf("my power %d, want 5 (+3/+0)", p)
	}
	if th := effectstest.LayeredCard(t, g, theirs).CurrentToughness(); th != 3 {
		t.Errorf("their toughness %d, want 3 (-0/-3)", th)
	}
}

// --- Rayblade Trooper ---

func TestStdB1RaybladeTrooperCountersAndMakesHumanSoldiersWhenCounteredCreaturesDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12PlayFromHand(t, g, "Rayblade Trooper", "Creature — Human Soldier", stdB1RaybladeOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("the bear has %d counters, want 1", got)
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Human Soldier") != 1 {
		t.Errorf("want one Human Soldier token, have %d", effectstest.B16CountNamed(g, "Human Soldier"))
	}
}

func TestStdB1RaybladeTrooperIgnoresTokensAndCounterlessCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Rayblade Trooper", "Creature — Human Soldier", stdB1RaybladeOracle, 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16CountNamed(g, "Human Soldier") != 0 {
		t.Error("a creature with no +1/+1 counter made a token")
	}
}

// --- Rescue Skiff ---

func TestStdB1RescueSkiffReturnsACreatureOrEnchantmentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "")
	ench := effectstest.PushGraveyardPermanent(me, "Old Ench", "Enchantment", "")
	rock := effectstest.PushGraveyardPermanent(me, "Old Rock", "Artifact", "")
	_ = rock
	effectstest.B12PlayFromHand(t, g, "Rescue Skiff", "Artifact — Spacecraft", stdB1SkiffOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ench) {
		t.Error("the enchantment card did not return")
	}
	if g.Battlefield.Contains(dead) || me.Graveyard.Contains(rock) == false {
		t.Error("only the chosen card may return, and an artifact card is no legal choice")
	}
}

// --- Pulsar Squadron Ace ---

func TestStdB1PulsarSquadronAceGetsACounterWhenNoSpacecraftIsTaken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	hand, libBefore := me.Hand.Size(), me.Library.Size()
	ace := effectstest.B12PlayFromHand(t, g, "Pulsar Squadron Ace", "Creature — Human Pilot", stdB1PulsarOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, ace, game.CounterPlusOne); got != 1 {
		t.Errorf("the Ace has %d counters, want 1: no card went to the hand", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand)
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library %d, want %d: the five looked-at cards go back on the bottom", me.Library.Size(), libBefore)
	}
}

func TestStdB1PulsarSquadronAceTakesASpacecraftAndGetsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	craft := effectstest.TopLibraryCard(me, "Skiff", "Artifact — Spacecraft")
	ace := effectstest.B12PlayFromHand(t, g, "Pulsar Squadron Ace", "Creature — Human Pilot", stdB1PulsarOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, craft)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(craft) {
		t.Error("the Spacecraft is not in hand")
	}
	if got := effectstest.CountersOn(g, ace, game.CounterPlusOne); got != 0 {
		t.Errorf("the Ace has %d counters, want none: a card went to the hand", got)
	}
}

// --- Lightstall Inquisitor ---

func stdB1Inquisitor(t *testing.T, g *game.Game, opp *game.Player, cards ...game.Card) {
	t.Helper()
	for _, seat := range g.Seats {
		if seat.ID != g.Seats[0].ID {
			stdB1EmptyHand(seat)
		}
	}
	for _, c := range cards {
		c.Owner, c.Controller = opp.ID, opp.ID
		opp.Hand.PushTop(c)
	}
	effectstest.B12PlayFromHand(t, g, "Lightstall Inquisitor", "Creature — Angel Wizard", stdB1InquisitorOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestStdB1LightstallInquisitorExilesASpellThatCostsOneMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := uuid.New()
	stdB1Inquisitor(t, g, opp,
		game.Card{InstanceID: spell, Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	effectstest.AnswerChooseCards(t, g, opp.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) || opp.Hand.Contains(spell) {
		t.Fatal("their only card was not exiled")
	}
	if !effectstest.B12ExiledPlayableBy(t, g, spell, opp.ID) {
		t.Fatal("the owner has no permission to play the exiled card")
	}
	if effectstest.B12ExiledPlayableBy(t, g, spell, me.ID) {
		t.Error("the Inquisitor's controller may play their card")
	}
	var price game.CastPrice
	var err error
	g.ReadSnapshot(func() {
		c, _ := effectstest.ExileCard(g, spell)
		price, err = g.PriceCastForEffect(opp.ID, c, game.CastSpellParams{FromZone: "exile"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := price.Total.String(); got != "{2}{G}" {
		t.Errorf("price %s, want {2}{G}: it costs {1} more", got)
	}
	// The permission outlives the Inquisitor.
	effectstest.B15Destroy(t, g, firstNamed(g, "Lightstall Inquisitor"))
	if !effectstest.B12ExiledPlayableBy(t, g, spell, opp.ID) {
		t.Error("the permission ended with the Inquisitor")
	}
}

func TestStdB1LightstallInquisitorLandEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := uuid.New()
	stdB1Inquisitor(t, g, opp,
		game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.AnswerChooseCards(t, g, opp.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(land) {
		t.Fatal("their only card was not exiled")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := g.CastSpell(opp.ID, land, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("the owner could not play the exiled land: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not enter")
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("a land played this way must enter tapped")
	}
}

func TestStdB1LightstallInquisitorAsksEachOpponentWithMoreThanOneCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a, b := uuid.New(), uuid.New()
	stdB1Inquisitor(t, g, opp,
		game.Card{InstanceID: a, Name: "A", TypeLine: "Sorcery"}, game.Card{InstanceID: b, Name: "B", TypeLine: "Sorcery"})
	c := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if c == nil || len(c.ChooseCards) != 2 {
		t.Fatalf("want a prompt for their hand, got %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, nil); err == nil {
		t.Error("exiling no card was accepted: it is not a \"may\"")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(b) || !opp.Hand.Contains(a) {
		t.Error("the chosen card is exiled and the other stays in hand")
	}
}

// --- Weftwalking ---

func TestStdB1WeftwalkingShufflesHandAndGraveyardThenDrawsSevenWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F", "G", "H", "I", "J")
	stdB1EmptyHand(me)
	stdB1Hand(me, 2)
	gy := effectstest.PushGraveyardPermanent(me, "Old", "Creature — Bear", "")
	effectstest.B12PlayFromHand(t, g, "Weftwalking", "Enchantment", stdB1WeftwalkingOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("hand %d, want 7", got)
	}
	if me.Graveyard.Contains(gy) && me.Graveyard.Size() > 0 && !me.Library.Contains(gy) && !me.Hand.Contains(gy) {
		t.Error("the graveyard was not shuffled into the library")
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard %d, want 0", me.Graveyard.Size())
	}
}

func TestStdB1WeftwalkingDoesNothingWhenItIsNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stdB1Hand(me, 2)
	hand := me.Hand.Size()
	effectstest.PushCatalogPermanent(g, me.ID, "Weftwalking", "Enchantment", stdB1WeftwalkingOracle, false)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: firstNamed(g, "Weftwalking")})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: a Weftwalking put onto the battlefield does nothing", hand, me.Hand.Size())
	}
}

func TestStdB1WeftwalkingFreeCastsTheFirstSpellEachPlayerCastsOnTheirTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Weftwalking", "Enchantment", stdB1WeftwalkingOracle, false)
	effectstest.AdvanceToMain(t, g)
	first, second := uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: first, Name: "Expensive", TypeLine: "Creature — Ogre", ManaCost: "{5}{U}", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: second, Name: "Expensive Two", TypeLine: "Creature — Ogre", ManaCost: "{5}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast a six-mana spell for nothing without claiming the offer")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFirstSpellFree}); err != nil {
		t.Fatalf("the first spell could not be cast for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(first) {
		t.Fatal("the free spell did not resolve")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFirstSpellFree}); err == nil {
		t.Error("a second spell the same turn was cast for free")
	}
	// An opponent gets the same on their own turn.
	third := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: third, Name: "Their Expensive", TypeLine: "Creature — Ogre", ManaCost: "{5}{U}", Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := g.CastSpell(opp.ID, third, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFirstSpellFree}); err != nil {
		t.Fatalf("the opponent's first spell on their turn could not be cast for free: %v", err)
	}
}

func TestStdB1WeftwalkingOffersNothingOutsideTheCastersOwnTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Weftwalking", "Enchantment", stdB1WeftwalkingOracle, false)
	effectstest.AdvanceToMain(t, g)
	flash := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: flash, Name: "Their Flash", TypeLine: "Instant", ManaCost: "{5}{U}", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, flash, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostFirstSpellFree}); err == nil {
		t.Error("an opponent cast a spell for free on my turn")
	}
}

// firstNamed is the instance ID of the first battlefield permanent named name.
func firstNamed(g *game.Game, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}
