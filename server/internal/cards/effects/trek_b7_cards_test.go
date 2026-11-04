package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b7_cards_test.go — Star Trek, batch 7 (#4973). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	trekB7GwyndalaOracle  = "852c38fe-8bfa-420b-b961-7b91af9b38d1"
	trekB7IllogicalOracle = "38c43f30-0e01-4dcd-88fc-f1de318a5e29"
	trekB7BountyOracle    = "eaa47a9b-79ad-4f42-a98b-dcf76138207c"
	trekB7HoshiOracle     = "39426cd1-8c0d-4ec6-8747-f0693080f8f2"
	trekB7WhalesOracle    = "a7285536-6cc1-448e-839a-606d585fbd9d"
	trekB7EffigyOracle    = "76f52fad-8db2-42f0-8e12-f4341e4c8680"
	trekB7JettOracle      = "492ba01e-c0fa-414c-95fd-df5b59212d2a"
	trekB7KobayashiOracle = "2b7ac3e8-0a7d-4b8a-a6a6-4cacc9b4f9de"
	trekB7DataOracle      = "ebe988c1-96e0-4a99-86b5-d948837545de"
	trekB7GeordiOracle    = "cb3b4cfe-fd8d-4a26-905e-8ccd30f033a3"
)

func TestTrekB7EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB7GwyndalaOracle:  "Gwyndala, the Progeny",
		trekB7IllogicalOracle: "Highly Illogical",
		trekB7BountyOracle:    "H.M.S. Bounty, Unlikely Timeship",
		trekB7HoshiOracle:     "Hoshi Sato, Exolinguist",
		trekB7WhalesOracle:    "Humpback Whales",
		trekB7EffigyOracle:    "Interstellar Effigy",
		trekB7JettOracle:      "Jett Reno, Problem Solver",
		trekB7KobayashiOracle: "Kobayashi Maru Scenario",
		trekB7DataOracle:      "Lieutenant Commander Data",
		trekB7GeordiOracle:    "Lt. Cmdr. Geordi La Forge",
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

// trekB7Cast puts `c` in `p`'s hand and casts it (no mana check), in the
// active player's main phase.
func trekB7Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	if err := g.CastSpell(p.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return c.InstanceID
}

// trekB7Creature seeds a non-catalog creature that can attack.
func trekB7Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return b27Push(g, owner, name, typeLine, "", "", power, toughness)
}

func trekB7Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return layeredCard(t, g, id).CurrentPower()
}

// --- Interstellar Effigy -----------------------------------------------------

func TestTrekB7EffigyBouncesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := trekB7Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := trekB7Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB7Cast(t, g, me, game.Card{Name: "Interstellar Effigy", TypeLine: "Artifact Creature — Construct",
		ManaCost: "{4}{U}", OracleID: trekB7EffigyOracle, Power: 4, Toughness: 3}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for a target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Fatal("a creature you control is not 'a creature an opponent controls'")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Errorf("the opponent's creature returns to its owner's hand (zone %v)", zoneOf(g, theirs))
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your own creature stays")
	}
}

// --- Humpback Whales ---------------------------------------------------------

func TestTrekB7WhalesHaveVigilanceAndWard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	whales := b27Push(g, me.ID, "Humpback Whales", "Creature — Whale", trekB7WhalesOracle, "{4}{U}{U}", 5, 6, "U")
	if !hasEffectiveKeyword(t, g, whales, "vigilance") {
		t.Error("Humpback Whales have vigilance")
	}
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, g.Seats[1], whales)
	for i := 0; i < 8 && !hasPayUnlessFor(g, g.Seats[1].ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, g.Seats[1].ID) {
		t.Error("an opponent targeting the Whales is asked to pay ward {2}")
	}
}

func TestTrekB7WhalesBasicLandcycle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me, searchTestLand("Island", "Basic Land — Island"))
	id, _ := cycleFromHand(t, g, "Humpback Whales", "Creature — Whale", trekB7WhalesOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("basic landcycling discards the Whales")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("basic landcycling searches")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the basic land reaches the hand")
	}
}

// --- Highly Illogical --------------------------------------------------------

func TestTrekB7HighlyIllogicalUnkickedCountersOnlyCheapSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := castWithCost(t, g, "Three Drop", "Sorcery", "{2}{U}", "")
	if _, err := castWithOptionalCosts(t, g, "Highly Illogical", "Instant", trekB7IllogicalOracle,
		b16TargetCard(big), nil, nil); err == nil {
		t.Fatal("unkicked, a mana value 3 spell is not a legal target")
	}
	small := castWithCost(t, g, "Two Drop", "Instant", "{1}{U}", "")
	if _, err := castWithOptionalCosts(t, g, "Highly Illogical", "Instant", trekB7IllogicalOracle,
		b16TargetCard(small), nil, nil); err != nil {
		t.Fatalf("unkicked at a mana value 2 spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(small) {
		t.Error("the two-drop is countered")
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the three-drop resolved into the graveyard as a sorcery")
	}
}

func TestTrekB7HighlyIllogicalKickedCountersAnything(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := castWithCost(t, g, "Wurm", "Creature — Wurm", "{4}{G}{G}", "")
	if _, err := castWithOptionalCosts(t, g, "Highly Illogical", "Instant", trekB7IllogicalOracle,
		b16TargetCard(big), []int{0}, nil); err != nil {
		t.Fatalf("kicked at a six-drop: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !me.Graveyard.Contains(big) {
		t.Error("the kicked counter takes the six-drop")
	}
}

// --- Kobayashi Maru Scenario ---------------------------------------------------

func TestTrekB7KobayashiMaruRunsOutAndMustAbandonShip(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	b27Push(g, me.ID, "Kobayashi Maru Scenario", "Enchantment", trekB7KobayashiOracle, "{1}{U}{U}", 0, 0, "U")
	cheap := b27Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	pricey := b27Push(g, opp.ID, "Their Giant", "Creature — Giant", "", "{3}{R}", 4, 4, "R")

	// Upkeep 1: Sensors.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	noModePick(t, g, me.ID, "only YOUR upkeep")
	advanceToUpkeepOf(t, g, seat)
	hand := handSize(me)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 1)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+2 {
		t.Errorf("Sensors draws two: hand %d → %d", hand, handSize(me))
	}

	// Upkeep 2: Hail, which may name only a mana value 3 or less creature.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	advanceToUpkeepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{0, 2, 3}, 0)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("Hail asks for up to one target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: pricey}); err == nil {
		t.Error("a mana value 4 creature is not a legal Hail target")
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(cheap) {
		t.Error("Hail returns the two-drop to its owner's hand")
	}

	// Upkeep 3: Transport.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	advanceToUpkeepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{2, 3}, 2)
	passPriorityAroundTable(t, g)
	officers := trekB7TokensNamed(t, g, me.ID, "Officer")
	if len(officers) != 2 {
		t.Fatalf("Transport made %d Officers, want 2", len(officers))
	}
	for _, o := range officers {
		if o.CurrentPower() != 1 || o.CurrentToughness() != 1 || !o.HasColor("R") || !o.IsCreature() {
			t.Errorf("Officer %+v, want a 1/1 red creature", o)
		}
	}

	// Upkeep 4: only Abandon Ship is left, and the trigger must take it.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	advanceToUpkeepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{3}, 3)
	passPriorityAroundTable(t, g)
	if !me.Eliminated {
		t.Error("Abandon Ship: you lose the game")
	}
}

func trekB7TokensNamed(t *testing.T, g *game.Game, controller uuid.UUID, name string) []game.Card {
	t.Helper()
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			if v := layeredCard(t, g, c.InstanceID); v.IsToken() {
				out = append(out, v)
			}
		}
	}
	return out
}

// --- Jett Reno, Problem Solver ------------------------------------------------

func trekB7Jett(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Jett Reno, Problem Solver", "Legendary Creature — Human Artificer", trekB7JettOracle, "{1}{U}", 1, 2, "U")
}

// trekB7AttackWith declares one attacker at `defender` and locks the
// attack in, so the "becomes tapped" trigger is on the stack.
func trekB7AttackWith(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
}

func TestTrekB7JettSurveilsWhenSheBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stockLibraryFor(g, me, 5)
	jett := trekB7Jett(g, me.ID)
	hand := handSize(me)
	trekB7AttackWith(t, g, jett, opp.ID)
	if n := triggersOnStackFrom(g, jett); n != 1 {
		t.Fatalf("%d Jett triggers on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Fatal("power 1: surveil 1")
	}
	if handSize(me) != hand {
		t.Error("power 1 draws nothing")
	}
}

func TestTrekB7JettDrawsInsteadAtFourPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stockLibraryFor(g, me, 5)
	jett := trekB7Jett(g, me.ID)
	setCounters(g, jett, map[string]int{game.CounterPlusOne: 3})
	if p := trekB7Power(t, g, jett); p != 4 {
		t.Fatalf("Jett's power %d, want 4", p)
	}
	hand := handSize(me)
	trekB7AttackWith(t, g, jett, opp.ID)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("at power 4 she draws instead of surveilling")
	}
	if handSize(me) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, handSize(me))
	}
}

func TestTrekB7JettGrowsWhenYouFaceADilemma(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jett := trekB7Jett(g, me.ID)

	// A modal spell: the modes are chosen as it is cast.
	castCatalogSpellWithModes(t, g, "Jeska's Will", "Sorcery", jeskasWillOracle, []int{1})
	if n := triggersOnStackFrom(g, jett); n != 1 {
		t.Fatalf("%d Jett triggers after a modal cast, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, jett, game.CounterPlusOne); got != 1 {
		t.Fatalf("+1/+1 counters %d, want 1", got)
	}

	// A plain spell is no dilemma.
	castWithCost(t, g, "Plain Sorcery", "Sorcery", "{1}", "")
	if n := triggersOnStackFrom(g, jett); n != 0 {
		t.Errorf("%d Jett triggers off a non-modal spell, want 0", n)
	}
	passPriorityAroundTable(t, g)
}

func TestTrekB7JettGrowsOffAModalTriggerButNotAnOpponentsDilemma(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	jett := trekB7Jett(g, me.ID)
	b27Push(g, me.ID, "Kobayashi Maru Scenario", "Enchantment", trekB7KobayashiOracle, "{1}{U}{U}", 0, 0, "U")
	b27Push(g, opp.ID, "Kobayashi Maru Scenario", "Enchantment", trekB7KobayashiOracle, "{1}{U}{U}", 0, 0, "U")

	// The opponent's Scenario: they face the dilemma, not you.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	chooseModeNow(t, g, opp.ID, []int{0, 1, 2, 3}, 1)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, jett, game.CounterPlusOne); got != 0 {
		t.Fatalf("an opponent's mode choice gave Jett %d counters", got)
	}

	// Your Scenario's upkeep mode choice is your dilemma: Jett's trigger
	// joins the Scenario's, and you order the two.
	advanceToUpkeepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 1)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, jett, game.CounterPlusOne); got != 1 {
		t.Errorf("your modal trigger's choice: %d counters, want 1", got)
	}
}

// --- Lieutenant Commander Data ------------------------------------------------

func trekB7Data(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Lieutenant Commander Data", "Legendary Artifact Creature — Android Scientist", trekB7DataOracle, "{2}{U}", 2, 2, "U")
}

func TestTrekB7DataGrowsOnAnyPlayersNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	data := trekB7Data(g, me.ID)
	advanceToMain(t, g)
	castWithCost(t, g, "Plain Sorcery", "Sorcery", "{1}", "")
	passPriorityAroundTable(t, g)
	if got := counterOn(g, data, game.CounterPlusOne); got != 1 {
		t.Fatalf("your sorcery: %d counters, want 1", got)
	}
	castWithCost(t, g, "Bear", "Creature — Bear", "{1}{G}", "")
	passPriorityAroundTable(t, g)
	if got := counterOn(g, data, game.CounterPlusOne); got != 1 {
		t.Errorf("a creature spell is not a noncreature spell: %d counters", got)
	}
	// An opponent's instant counts too.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	opp := g.Seats[1]
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Instant", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent's instant: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, data, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent's instant: %d counters, want 2", got)
	}
}

func TestTrekB7DataRemovesThreeCountersToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stockLibraryFor(g, me, 5)
	data := trekB7Data(g, me.ID)
	other := trekB7Creature(g, me.ID, "Helper", "Creature — Human", 1, 1)
	setCounters(g, data, map[string]int{game.CounterPlusOne: 1})
	setCounters(g, other, map[string]int{game.CounterCharge: 1})
	advanceToMain(t, g)
	fillPool(me, 1)
	// Any kinds, from among any of your creatures.
	params := game.ActivateAbilityParams{
		Strict:           true,
		CounterSourceIDs: []uuid.UUID{data, other},
		CounterCounts:    []int{1, 1},
		CounterKinds:     []string{game.CounterPlusOne, game.CounterCharge},
	}
	if err := g.ActivateCatalogAbility(me.ID, data, 0, params); err == nil {
		t.Fatal("two counters among your creatures cannot pay 'remove three'")
	}
	setCounters(g, other, map[string]int{game.CounterCharge: 2})
	params.CounterCounts = []int{1, 2}
	hand := handSize(me)
	if err := g.ActivateCatalogAbility(me.ID, data, 0, params); err != nil {
		t.Fatalf("{1} and three counters: %v", err)
	}
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, handSize(me))
	}
	if counterOn(g, data, game.CounterPlusOne)+counterOn(g, other, game.CounterCharge) != 0 {
		t.Error("the three counters were removed")
	}
}

// --- Lt. Cmdr. Geordi La Forge ------------------------------------------------

func TestTrekB7GeordiDoublesChargeCountersOnYourArtifacts(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "The Seriema", "Legendary Artifact — Spacecraft", theSeriemaOracle, 2)
	b27Push(g, me.ID, "Lt. Cmdr. Geordi La Forge", "Legendary Creature — Human Artificer", trekB7GeordiOracle, "{2}{U}", 3, 2, "U")
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ship, game.CounterCharge); got != 4 {
		t.Errorf("a 2-power station under Geordi: %d charge counters, want 4", got)
	}
}

func TestTrekB7GeordiLeavesOtherCountersAndOtherPermanentsAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Lt. Cmdr. Geordi La Forge", "Legendary Creature — Human Artificer", trekB7GeordiOracle, "{2}{U}", 3, 2, "U")
	// Lieutenant Commander Data is an artifact, and his trigger puts a
	// +1/+1 counter on him: not a charge counter, so not doubled.
	data := trekB7Data(g, me.ID)
	advanceToMain(t, g)
	castWithCost(t, g, "Plain Sorcery", "Sorcery", "{1}", "")
	passPriorityAroundTable(t, g)
	if got := counterOn(g, data, game.CounterPlusOne); got != 1 {
		t.Errorf("a +1/+1 counter on an artifact is not doubled: %d", got)
	}
}

// --- Hoshi Sato, Exolinguist --------------------------------------------------

func TestTrekB7HoshiLooksAtOnePerCreatureTypeAmongNonBorg(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB7Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior", 1, 1)
	trekB7Creature(g, me.ID, "Second Elf", "Creature — Elf", 1, 1)
	trekB7Creature(g, me.ID, "Borg Drone", "Artifact Creature — Borg Drone", 1, 1)
	trekB7Creature(g, g.Seats[1].ID, "Their Goblin", "Creature — Goblin", 1, 1)
	var lib []game.Card
	for _, n := range []string{"A", "B", "C", "D", "E", "F", "G"} {
		lib = append(lib, game.Card{Name: n, TypeLine: "Sorcery"})
	}
	ids := znrStackLibrary(me, lib...)
	trekB7Cast(t, g, me, game.Card{Name: "Hoshi Sato, Exolinguist", TypeLine: "Legendary Creature — Human Scientist",
		ManaCost: "{4}{U}", OracleID: trekB7HoshiOracle, Power: 3, Toughness: 3}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	// Human, Scientist, Elf, Warrior: four types; the Borg Drone and the
	// opponent's Goblin add nothing.
	if len(pick.ChooseCards) != 4 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards bounds %d..%d, want the top 4, exactly one", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, ids[2])
	if !me.Hand.Contains(ids[2]) {
		t.Error("the chosen card reaches the hand")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[4] {
		t.Errorf("the rest went to the bottom: the top card is %v, want the fifth", top)
	}
}

// --- H.M.S. Bounty, Unlikely Timeship -----------------------------------------

func trekB7Bounty(g *game.Game, owner uuid.UUID, charge int) uuid.UUID {
	id := b27Push(g, owner, "H.M.S. Bounty, Unlikely Timeship", "Legendary Artifact — Spacecraft", trekB7BountyOracle, "{4}{U}", 3, 4, "U")
	if charge > 0 {
		setCounters(g, id, map[string]int{game.CounterCharge: charge})
	}
	return id
}

func TestTrekB7BountySearchesForAWhale(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Humpback Whales", TypeLine: "Creature — Whale"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	trekB7Cast(t, g, me, game.Card{Name: "H.M.S. Bounty, Unlikely Timeship", TypeLine: "Legendary Artifact — Spacecraft",
		ManaCost: "{4}{U}", OracleID: trekB7BountyOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters search asks")
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Whale is not offered")
	}
	answerSearchNamed(t, g, me.ID, "Humpback Whales")
	if !b02bHandHasNamed(me, "Humpback Whales") {
		t.Error("the Whale reaches the hand")
	}
}

func TestTrekB7BountyIsAFlyingCreatureAtEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	low := trekB7Bounty(g, me.ID, 7)
	if layeredCard(t, g, low).IsCreature() {
		t.Error("seven charge counters is below 8+")
	}
	high := trekB7Bounty(g, me.ID, 8)
	v := layeredCard(t, g, high)
	if !v.IsCreature() || v.CurrentPower() != 3 || v.CurrentToughness() != 4 {
		t.Errorf("at 8+ it is a 3/4 creature: creature=%v %d/%d", v.IsCreature(), v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, high, "flying") {
		t.Error("at 8+ it has flying")
	}
}

func TestTrekB7BountyTakesOneExtraTurnWithAWhale(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	stockLibraryFor(g, me, 10)
	bounty := trekB7Bounty(g, me.ID, 8)
	b27Push(g, me.ID, "Humpback Whales", "Creature — Whale", trekB7WhalesOracle, "{4}{U}{U}", 5, 6, "U")
	hand := handSize(me)
	attackWith(t, g, opp.ID, bounty)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Fatalf("hand %d → %d, want one card drawn", hand, handSize(me))
	}
	endTurn(t, g)
	assertTurnOf(t, g, seat, true, "after the Bounty connected with a Whale")
	if me.ExtraTurnsBegun != 1 {
		t.Fatalf("ExtraTurnsBegun = %d, want 1", me.ExtraTurnsBegun)
	}

	// In the extra turn it draws, but you have begun an extra turn.
	advanceToMain(t, g)
	hand = handSize(me)
	attackWith(t, g, opp.ID, bounty)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Errorf("the extra turn's hit still draws: hand %d → %d", hand, handSize(me))
	}
	endTurn(t, g)
	assertTurnOf(t, g, (seat+1)%4, false, "no second extra turn")
}

func TestTrekB7BountyWithoutAWhaleOnlyDraws(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	stockLibraryFor(g, me, 10)
	bounty := trekB7Bounty(g, me.ID, 8)
	hand := handSize(me)
	attackWith(t, g, opp.ID, bounty)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Fatalf("hand %d → %d, want one card drawn", hand, handSize(me))
	}
	endTurn(t, g)
	assertTurnOf(t, g, (seat+1)%4, false, "no Whale, no extra turn")
}

// --- Gwyndala, the Progeny -----------------------------------------------------

func TestTrekB7GwyndalaEntersWithAnHeirloomAttached(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gw := trekB7Cast(t, g, me, game.Card{Name: "Gwyndala, the Progeny", TypeLine: "Legendary Creature — Vau-N'Akat Advisor",
		ManaCost: "{1}{U}", OracleID: trekB7GwyndalaOracle, Power: 1, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	heirlooms := trekB7TokensNamed(t, g, me.ID, "Neuroflux Heirloom")
	if len(heirlooms) != 1 {
		t.Fatalf("%d Neuroflux Heirlooms, want 1", len(heirlooms))
	}
	h := heirlooms[0]
	if !h.IsArtifact() || !h.HasSubtype("Equipment") || len(h.Colors) != 0 {
		t.Errorf("the Heirloom is a colorless Equipment artifact: %+v", h)
	}
	if h.AttachedTo.ID != gw {
		t.Fatal("the Heirloom is attached to Gwyndala")
	}
	if p := trekB7Power(t, g, gw); p != 2 {
		t.Errorf("equipped Gwyndala power %d, want 2", p)
	}
}

func TestTrekB7GwyndalaTurnsEquipmentTokensIntoCopies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gw := trekB7Cast(t, g, me, game.Card{Name: "Gwyndala, the Progeny", TypeLine: "Legendary Creature — Vau-N'Akat Advisor",
		ManaCost: "{1}{U}", OracleID: trekB7GwyndalaOracle, Power: 1, Toughness: 1}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	heirloom := trekB7TokensNamed(t, g, me.ID, "Neuroflux Heirloom")[0].InstanceID
	blade := b27Push(g, me.ID, "Bonesplitter", "Artifact — Equipment", bonesplitterOracle, "{1}", 0, 0)
	theirs := b27Push(g, g.Seats[1].ID, "Bonesplitter", "Artifact — Equipment", bonesplitterOracle, "{1}", 0, 0)

	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, gw, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(theirs)}); err == nil {
		t.Fatal("an opponent's Equipment is not 'Equipment you control'")
	}
	if err := g.ActivateCatalogAbility(me.ID, gw, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetCard(blade)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	v := layeredCard(t, g, heirloom)
	if v.Name != "Bonesplitter" || !v.IsToken() {
		t.Fatalf("the Heirloom is now a Bonesplitter token: name %q token %v", v.Name, v.IsToken())
	}
	if v.AttachedTo.ID != gw {
		t.Error("the copy stays attached")
	}
	if p := trekB7Power(t, g, gw); p != 3 {
		t.Errorf("Gwyndala under a copied Bonesplitter: power %d, want 3", p)
	}
	if layeredCard(t, g, blade).Name != "Bonesplitter" {
		t.Error("the nontoken target is untouched")
	}

	// Until end of turn.
	advanceToUpkeepOf(t, g, 1)
	if v := layeredCard(t, g, heirloom); v.Name != "Neuroflux Heirloom" {
		t.Errorf("after the turn the token is %q again, want Neuroflux Heirloom", v.Name)
	}
}

// --- the dilemma seam ----------------------------------------------------------

func TestTrekB7NoDilemmaForAModeChosenUpToNone(t *testing.T) {
	// The seam's negative: a modal spell cast with no modes is refused
	// before anything is chosen, and so emits nothing.
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	jett := trekB7Jett(g, me.ID)
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Jeska's Will", "Sorcery", jeskasWillOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Fatal("Jeska's Will needs a mode")
	}
	if n := triggersOnStackFrom(g, jett); n != 0 {
		t.Errorf("a refused cast faced %d dilemmas", n)
	}
	if me.ExtraTurnsBegun != 0 {
		t.Error("nobody has begun an extra turn")
	}
}
