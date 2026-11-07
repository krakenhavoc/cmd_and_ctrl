package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_hl_cards_test.go — Aetherdrift (DFT), cards from H to L (#4464).
// Each test plays the real catalog entry and asserts what its text says.
// The fixtures are dft_ab_cards_test.go's, dft_st_cards_test.go's,
// dft_mp_cards_test.go's and dft_cg_cards_test.go's.

const (
	dftHLHellrideOracle    = "47944bf1-b954-4143-b479-45c10071e35a"
	dftHLHauntOracle       = "7e074b14-0ec8-439b-9b83-5f7c1a904384"
	dftHLHazardOracle      = "799664fc-279c-40c1-8e40-44a68c38801b"
	dftHLHowlersOracle     = "8abeaa77-d455-47e7-bf60-aefb94ab83fb"
	dftHLHulldrifterOracle = "a0a7319d-9441-4aed-a351-9ff3fb3b1d9e"
	dftHLTacticsOracle     = "121bd1af-da21-4937-a911-377c5b430796"
	dftHLJibbirikOracle    = "9cda6cb8-c254-413a-babd-0e24e7b89cc4"
	dftHLKalakscionOracle  = "d85f5edb-60a8-4abd-82b0-f9a58f3d4185"
	dftHLBuccaneerOracle   = "8cadf9c0-9a63-4cde-a6be-0ae86890c831"
	dftHLParryOracle       = "ecef9e3d-dce8-405f-bba3-14732444258f"
	dftHLLotusguardOracle  = "1139ee31-cb6e-4ae6-a08b-252c46b2db97"
)

func TestDftHLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftHLHellrideOracle:    "Haunted Hellride",
		dftHLHauntOracle:       "Haunt the Network",
		dftHLHazardOracle:      "Hazard of the Dunes",
		dftHLHowlersOracle:     "Howler's Heavy",
		dftHLHulldrifterOracle: "Hulldrifter",
		dftHLTacticsOracle:     "Intimidation Tactics",
		dftHLJibbirikOracle:    "Jibbirik Omnivore",
		dftHLKalakscionOracle:  "Kalakscion, Hunger Tyrant",
		dftHLBuccaneerOracle:   "Keen Buccaneer",
		dftHLParryOracle:       "Lightshield Parry",
		dftHLLotusguardOracle:  "Lotusguard Disciple",
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

// dftHLTokensNamed counts the permanents named name that p controls.
func dftHLTokensNamed(g *game.Game, p uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p && c.Name == name {
			n++
		}
	}
	return n
}

// --- Haunted Hellride -------------------------------------------------------------

func TestDftHLHauntedHellridePumpsGrantsDeathtouchAndUntapsWhenYouAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hellride := dftABPush(g, me.ID, "Haunted Hellride", dftABVehicle, dftHLHellrideOracle, 3, 3)
	attacker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The uncrewed Hellride does not attack; "whenever you attack" fires
	// anyway.
	declareAttack(t, g, opp.ID, attacker)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if hasID(pick.PickTargetCards, hellride) {
		t.Error("the uncrewed Hellride is not a creature, so not a legal target")
	}
	dftCGPickRefused(t, g, me.ID, theirs)
	pickCard(t, g, me.ID, attacker)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, attacker)
	if !ok {
		t.Fatal("the attacker left the battlefield")
	}
	if c.Tapped {
		t.Error("the attacker was not untapped")
	}
	if p, tough := effectivePTOf(t, g, attacker); p != 3 || tough != 2 {
		t.Errorf("the attacker is %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, attacker, "deathtouch") {
		t.Error("the attacker has no deathtouch")
	}
	dftMPPassTurn(t, g)
	if p, tough := effectivePTOf(t, g, attacker); p != 2 || tough != 2 {
		t.Errorf("the attacker is %d/%d after the turn ended, want 2/2", p, tough)
	}
	if hasEffectiveKeyword(t, g, attacker, "deathtouch") {
		t.Error("deathtouch outlived the turn")
	}
}

func TestDftHLHauntedHellrideCrewsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hellride := dftABPush(g, me.ID, "Haunted Hellride", dftABVehicle, dftHLHellrideOracle, 3, 3)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	advanceToMain(t, g)
	dftABCrew(t, g, me, hellride, 0, one)
	if p, tough := effectivePTOf(t, g, hellride); p != 3 || tough != 3 {
		t.Errorf("the crewed Hellride is %d/%d, want 3/3", p, tough)
	}
}

// --- Haunt the Network ------------------------------------------------------------

func TestDftHLHauntTheNetworkMakesTwoThoptersThenDrainsByArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	b12Permanent(g, opp.ID, "Their Stone", "Artifact")
	advanceToMain(t, g)
	mine, theirs := me.Life, opp.Life
	castAt(t, g, "Haunt the Network", "Sorcery", dftHLHauntOracle, pbPlayer(opp.ID))
	if got := dftHLTokensNamed(g, me.ID, "Thopter"); got != 2 {
		t.Fatalf("%d Thopters, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Thopter" && !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
			t.Error("a Thopter has no flying")
		}
	}
	// X counts the two new Thopters and the Mind Stone, not the
	// opponent's artifact.
	if opp.Life != theirs-3 || me.Life != mine+3 {
		t.Errorf("life: opponent %d -> %d, you %d -> %d; want -3 and +3", theirs, opp.Life, mine, me.Life)
	}
}

func TestDftHLHauntTheNetworkTargetsAnOpponentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := castCatalogSpellErr(t, g, "Haunt the Network", "Sorcery", dftHLHauntOracle,
		[]game.TargetRef{pbPlayer(me.ID)}); err == nil {
		t.Fatal("Haunt the Network targeted its caster; it wants an opponent")
	}
}

// --- Hazard of the Dunes ----------------------------------------------------------

func TestDftHLHazardOfTheDunesExhaustsForThreeCountersOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wurm := dftABPush(g, me.ID, "Hazard of the Dunes", "Creature — Wurm", dftHLHazardOracle, 4, 4)
	for _, kw := range []string{"reach", "trample"} {
		if !hasEffectiveKeyword(t, g, wurm, kw) {
			t.Errorf("Hazard of the Dunes has no %s", kw)
		}
	}
	advanceToMain(t, g)
	fillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, wurm, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("exhausted for colourless mana; it costs {6}{G}")
	}
	me.ManaPool = nil
	fillPool(me, 6)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, wurm, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, wurm, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3", got)
	}
	if p, tough := ptOf(t, g, wurm); p != 7 || tough != 7 {
		t.Errorf("the Wurm is %d/%d, want 7/7", p, tough)
	}
	fillPool(me, 6)
	fillPoolColored(me, "G", 1)
	if err := g.ActivateCatalogAbility(me.ID, wurm, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Howler's Heavy ---------------------------------------------------------------

func TestDftHLHowlersHeavyCyclesAndShrinksAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	seedLibrary(me, "Drawn")
	hand := me.Hand.Size()
	id, _ := cycleFromHand(t, g, "Howler's Heavy", "Creature — Seal Pirate", dftHLHowlersOracle, "{C}{U}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Heavy is not in the graveyard")
	}
	dftCGPickRefused(t, g, me.ID, mine)
	pickTriggerTarget(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d after cycling, want %d (the Heavy out, a card in)", got, hand+1)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 4 {
		t.Errorf("the target is %d/%d, want 1/4", p, tough)
	}
	dftMPPassTurn(t, g)
	if p, tough := effectivePTOf(t, g, theirs); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d after the turn ended, want 4/4", p, tough)
	}
}

func TestDftHLHowlersHeavyCyclingWantsBlue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Howler's Heavy", "Creature — Seal Pirate", dftHLHowlersOracle)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("cycled for colourless mana; cycling costs {1}{U}")
	}
}

// --- Hulldrifter ------------------------------------------------------------------

func TestDftHLHulldrifterDrawsTwoOnEntryAndCrewsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	ship := dftABCast(t, g, me, game.Card{Name: "Hulldrifter", TypeLine: dftABVehicle,
		OracleID: dftHLHulldrifterOracle, Power: 3, Toughness: 2})
	if !g.Battlefield.Contains(ship) {
		t.Fatal("Hulldrifter did not enter")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want two cards drawn", hand, got)
	}
	if !hasEffectiveKeyword(t, g, ship, "flying") {
		t.Error("Hulldrifter has no flying")
	}
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{two}}); err == nil {
		t.Fatal("crewed with total power 2; crew 3 wants 3")
	}
	dftABCrew(t, g, me, ship, 0, two, one)
	if p, tough := effectivePTOf(t, g, ship); p != 3 || tough != 2 {
		t.Errorf("the crewed Hulldrifter is %d/%d, want 3/2", p, tough)
	}
}

// --- Intimidation Tactics ---------------------------------------------------------

func TestDftHLIntimidationTacticsExilesAnArtifactOrCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := revealHand(victim, rhForest(), rhBolt(),
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	castAt(t, g, "Intimidation Tactics", "Sorcery", dftHLTacticsOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, ids[2:]) || pick.PickDestination != game.PickExile || pick.PickOptional {
		t.Fatalf("pick = options %v destination %q optional %v; want the creature and the artifact, exile, mandatory",
			pick.DiscardOptions, pick.PickDestination, pick.PickOptional)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[3:]); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !g.Exile.Contains(ids[3]) || victim.Graveyard.Contains(ids[3]) {
		t.Error("the artifact card was not exiled")
	}
	if victim.Hand.Size() != 3 {
		t.Errorf("victim's hand is %d, want 3", victim.Hand.Size())
	}
}

func TestDftHLIntimidationTacticsTargetsAnOpponentAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := castCatalogSpellErr(t, g, "Intimidation Tactics", "Sorcery", dftHLTacticsOracle,
		[]game.TargetRef{pbPlayer(me.ID)}); err == nil {
		t.Fatal("Intimidation Tactics targeted its caster; it wants an opponent")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	seedLibrary(me2, "Drawn")
	hand := me2.Hand.Size()
	id, _ := cycleFromHand(t, g2, "Intimidation Tactics", "Sorcery", dftHLTacticsOracle, "{C}{C}{C}")
	passPriorityAroundTable(t, g2)
	if !me2.Graveyard.Contains(id) || me2.Hand.Size() != hand+1 {
		t.Errorf("cycling did not discard the card for a card: graveyard %v, hand %d want %d",
			me2.Graveyard.Contains(id), me2.Hand.Size(), hand+1)
	}
}

// --- the vanillas -----------------------------------------------------------------

func TestDftHLTheVanillasResolveAsPlainCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Jibbirik Omnivore", "Creature — Beast", dftHLJibbirikOracle, 3, 2},
		{"Kalakscion, Hunger Tyrant", "Legendary Creature — Crocodile", dftHLKalakscionOracle, 7, 2},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		id := dftABCast(t, g, me, game.Card{Name: tc.name, TypeLine: tc.typeLine, OracleID: tc.oracle,
			Power: tc.power, Toughness: tc.toughness})
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not enter", tc.name)
			continue
		}
		if p, tough := effectivePTOf(t, g, id); p != tc.power || tough != tc.toughness {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.power, tc.toughness)
		}
		if len(g.PendingChoices) != 0 || g.Stack.Size() != 0 {
			t.Errorf("%s left something behind: choices %v, stack %d", tc.name, g.PendingChoices, g.Stack.Size())
		}
	}
}

// --- Keen Buccaneer ---------------------------------------------------------------

func TestDftHLKeenBuccaneerLootsThenGrowsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pirate := dftABPush(g, me.ID, "Keen Buccaneer", "Creature — Octopus Pirate", dftHLBuccaneerOracle, 2, 3)
	if !hasEffectiveKeyword(t, g, pirate, "vigilance") {
		t.Error("Keen Buccaneer has no vigilance")
	}
	advanceToMain(t, g)
	seedLibrary(me, "Drawn", "Next")
	junk := pushCatalogHandCard(me, "Forest", "Basic Land — Forest", "")
	hand := me.Hand.Size()
	fillPool(me, 1)
	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, pirate, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt after the draw")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d before the discard, want one card drawn", hand, got)
	}
	// The counter waits for the discard (CR 608.2c).
	if got := countersOn(g, pirate, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters before the discard, want 0", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{junk}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(junk) || me.Hand.Size() != hand {
		t.Errorf("the loot did not discard the Forest: graveyard %v, hand %d want %d",
			me.Graveyard.Contains(junk), me.Hand.Size(), hand)
	}
	if got := countersOn(g, pirate, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	fillPool(me, 1)
	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, pirate, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Lightshield Parry ------------------------------------------------------------

func TestDftHLLightshieldParryPumpsTwoTwoUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	castAt(t, g, "Lightshield Parry", "Instant", dftHLParryOracle, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	dftMPPassTurn(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear is %d/%d after the turn ended, want 2/2", p, tough)
	}
}

func TestDftHLLightshieldParryCycles(t *testing.T) {
	dftSTCycles(t, "Lightshield Parry", "Instant", dftHLParryOracle, 0)
}

// --- Lotusguard Disciple ----------------------------------------------------------

func TestDftHLLotusguardDiscipleGrantsLifelinkAndIndestructibleToACreatureOrVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stone := b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	wagon := b12Permanent(g, opp.ID, "Their Wagon", dftABVehicle)
	advanceToMain(t, g)
	disciple := dftABCast(t, g, me, game.Card{Name: "Lotusguard Disciple", TypeLine: "Creature — Bird Cleric",
		OracleID: dftHLLotusguardOracle, Power: 2, Toughness: 2})
	if !hasEffectiveKeyword(t, g, disciple, "flying") {
		t.Error("Lotusguard Disciple has no flying")
	}
	dftCGPickRefused(t, g, me.ID, stone)
	pickTriggerTarget(t, g, me.ID, wagon)
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"lifelink", "indestructible"} {
		if !hasEffectiveKeyword(t, g, wagon, kw) {
			t.Errorf("the Vehicle has no %s", kw)
		}
	}
	dftMPPassTurn(t, g)
	if hasEffectiveKeyword(t, g, wagon, "indestructible") {
		t.Error("indestructible outlived the turn")
	}
}

func TestDftHLLotusguardDiscipleCanTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	disciple := dftABCast(t, g, me, game.Card{Name: "Lotusguard Disciple", TypeLine: "Creature — Bird Cleric",
		OracleID: dftHLLotusguardOracle, Power: 2, Toughness: 2})
	pickTriggerTarget(t, g, me.ID, disciple)
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"lifelink", "indestructible"} {
		if !hasEffectiveKeyword(t, g, disciple, kw) {
			t.Errorf("the Disciple has no %s", kw)
		}
	}
}
