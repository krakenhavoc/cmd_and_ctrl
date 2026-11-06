package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_b1_cards_test.go — Bloomburrow batch 1 (#4467). Each test plays
// the real catalog entry and asserts what its text says.

const (
	blbB1AgateOracle      = "381a3e8e-71dd-48e4-ab62-53478bde4a14"
	blbB1CrierOracle      = "f2f3be8a-5ea1-47c2-8104-747f4589f5e6"
	blbB1BraveKinOracle   = "1f5b7f4c-a3d2-4472-80f5-8504562fc7fd"
	blbB1CaptainOracle    = "104baa4e-a3f3-45dc-979d-4c58b659ba1a"
	blbB1BakersbaneOracle = "ecd9a0a4-1dfa-4d91-8339-062e2be4e5d9"
	blbB1BoxerOracle      = "08e173e3-4c28-49d2-bb57-ccf533f20daf"
	blbB1TideOracle       = "d605d9d1-c01a-4ef6-89f9-06fae8cd57e6"
	blbB1ConductOracle    = "0369627e-3b83-4e5b-a8af-a99f944443b3"
	blbB1CrumbOracle      = "93596258-bb44-4aa3-833d-9a0ab16b5ec1"
	blbB1BlastOracle      = "2e5b8973-4f6e-4abe-bbbe-81fdca784997"
)

func TestBlbB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbB1AgateOracle:      "Agate-Blade Assassin",
		blbB1CrierOracle:      "Bellowing Crier",
		blbB1BraveKinOracle:   "Brave-Kin Duo",
		blbB1CaptainOracle:    "Brambleguard Captain",
		blbB1BakersbaneOracle: "Bakersbane Duo",
		blbB1BoxerOracle:      "Bark-Knuckle Boxer",
		blbB1TideOracle:       "Calamitous Tide",
		blbB1ConductOracle:    "Conduct Electricity",
		blbB1CrumbOracle:      "Crumb and Get It",
		blbB1BlastOracle:      "Blooming Blast",
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

// blbB1Push puts a catalog creature on the battlefield, able to attack
// and tap this turn.
func blbB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// blbB1Seats is the active player and the next two seats.
func blbB1Seats(g *game.Game) (me, opp, other *game.Player) {
	n := len(g.Seats)
	a := g.Turn.ActiveSeat
	return g.Seats[a], g.Seats[(a+1)%n], g.Seats[(a+2)%n]
}

// --- Agate-Blade Assassin ----------------------------------------------------

func TestBlbB1AgateBladeAssassinDrainsTheDefendingPlayerWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	assassin := blbB1Push(g, me.ID, "Agate-Blade Assassin", "Creature — Lizard Assassin", blbB1AgateOracle, 1, 3)
	myLife, oppLife, otherLife := lifeOf(g, me.ID), lifeOf(g, opp.ID), lifeOf(g, other.ID)
	declareAttack(t, g, opp.ID, assassin)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("the defending player is at %d, want %d", got, oppLife-1)
	}
	if got := lifeOf(g, me.ID); got != myLife+1 {
		t.Errorf("I am at %d, want %d", got, myLife+1)
	}
	if got := lifeOf(g, other.ID); got != otherLife {
		t.Errorf("a player who is not defending lost life: %d, want %d", got, otherLife)
	}
}

func TestBlbB1AgateBladeAssassinIgnoresAnotherCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Agate-Blade Assassin", "Creature — Lizard Assassin", blbB1AgateOracle, 1, 3)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	myLife, oppLife := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if lifeOf(g, me.ID) != myLife || lifeOf(g, opp.ID) != oppLife {
		t.Error("the Assassin triggered on another creature's attack; it says \"this creature\"")
	}
}

// --- Bellowing Crier ---------------------------------------------------------

func TestBlbB1BellowingCrierLootsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	before := me.Hand.Size()
	yard := me.Graveyard.Size()
	crier := castHoldCreature(t, g, "Bellowing Crier", "Creature — Frog Advisor", blbB1CrierOracle, 2, 1)
	discardFromHand(t, g, me.ID)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(crier) {
		t.Fatal("the Crier is not on the battlefield")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (draw one, discard one)", got, before)
	}
	if got := me.Graveyard.Size(); got != yard+1 {
		t.Errorf("graveyard %d, want %d", got, yard+1)
	}
}

// --- Brave-Kin Duo -----------------------------------------------------------

func TestBlbB1BraveKinDuoPumpsAtSorcerySpeedOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	duo := blbB1Push(g, me.ID, "Brave-Kin Duo", "Creature — Rabbit Mouse", blbB1BraveKinOracle, 1, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	params := game.ActivateAbilityParams{Targets: cardTarget(bear)}

	// The harness parks on the draw step: not a main phase.
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, duo, 0, params); err == nil {
		t.Fatal("the ability was activated in the draw step; it is sorcery speed")
	}

	advanceToMain(t, g)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, duo, 0, params); err != nil {
		t.Fatalf("activate in my main phase: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
	if c := findBattlefieldCardByID(g, duo); c == nil || !c.Tapped {
		t.Error("the Duo should be tapped by its cost")
	}
}

// --- Brambleguard Captain ----------------------------------------------------

func TestBlbB1BrambleguardCaptainGivesItsPowerAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	captain := riotPush(g, me.ID, "Brambleguard Captain", "Creature — Mouse Soldier", blbB1CaptainOracle, 2, 3, 1)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("an opponent's creature was accepted; the target is a creature you control")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	// The Captain is 3/4 with its counter: X is 3.
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 5/2", p, tough)
	}
	if p, _ := ptOf(t, g, captain); p != 3 {
		t.Errorf("the Captain's power is %d, want 3 (it pumped another creature)", p)
	}
}

func TestBlbB1BrambleguardCaptainMayTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	captain := blbB1Push(g, me.ID, "Brambleguard Captain", "Creature — Mouse Soldier", blbB1CaptainOracle, 2, 3)
	advanceTo(t, g, game.StepBeginCombat)
	answerPickTarget(t, g, captain)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, captain); p != 4 {
		t.Errorf("the Captain's power is %d, want 4", p)
	}
}

// --- Bakersbane Duo ----------------------------------------------------------

func TestBlbB1BakersbaneDuoMakesAFoodAndGrowsOnExpendFour(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	duo := castHoldCreature(t, g, "Bakersbane Duo", "Creature — Squirrel Raccoon", blbB1BakersbaneOracle, 2, 2)
	if n, _ := countTokensNamed(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Food tokens, want 1", n)
	}
	castPaid(t, g, me, "Three", "{2}{G}", "CCG")
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, duo); p != 2 || tough != 2 {
		t.Fatalf("three mana spent: the Duo is %d/%d, want 2/2", p, tough)
	}
	castPaid(t, g, me, "One", "{G}", "G")
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, duo); p != 3 || tough != 3 {
		t.Errorf("the fourth mana was spent: the Duo is %d/%d, want 3/3", p, tough)
	}
	castPaid(t, g, me, "Four More", "{3}{G}", "CCCG")
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, duo); p != 3 {
		t.Errorf("expend 4 triggers once a turn: the Duo's power is %d, want 3", p)
	}
}

// --- Bark-Knuckle Boxer ------------------------------------------------------

func TestBlbB1BarkKnuckleBoxerGainsIndestructibleOnExpendFour(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	boxer := blbB1Push(g, me.ID, "Bark-Knuckle Boxer", "Creature — Raccoon Berserker", blbB1BoxerOracle, 3, 2)
	toMainForCost(t, g)
	castPaid(t, g, me, "Three", "{2}{G}", "CCG")
	passPriorityAroundTable(t, g)
	if hasKeywordOnBattlefield(t, g, boxer, "indestructible") {
		t.Fatal("three mana spent: no expend 4 yet")
	}
	castPaid(t, g, me, "Two", "{1}{G}", "CG")
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, boxer, "indestructible") {
		t.Error("the fourth mana was spent: the Boxer should be indestructible")
	}
}

// --- Calamitous Tide ---------------------------------------------------------

func TestBlbB1CalamitousTideReturnsTwoCreaturesThenLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	a := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	b := blbB1Push(g, other.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Calamitous Tide", "Sorcery", blbB1TideOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !other.Hand.Contains(b) {
		t.Error("each creature should be in its owner's hand")
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (draw two, discard one)", got, before+1)
	}
}

func TestBlbB1CalamitousTideWithNoTargetStillLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Calamitous Tide", "Sorcery", blbB1TideOracle, nil)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (draw two, discard one)", got, before+1)
	}
}

// --- Conduct Electricity -----------------------------------------------------

func TestBlbB1ConductElectricityHitsACreatureAndAToken(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	big := blbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 7, 7)
	tok := TokenCard("1/1 blue Fish")
	tok.Power, tok.Toughness = 3, 3
	fish := pushToken(g, opp.ID, tok)
	castCatalogSpell(t, g, "Conduct Electricity", "Instant", blbB1ConductOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: big, Slot: 0}, {Kind: game.TargetCard, ID: fish, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, big); got != 6 {
		t.Errorf("the creature has %d damage, want 6", got)
	}
	if got := damageOn(t, g, fish); got != 2 {
		t.Errorf("the token has %d damage, want 2", got)
	}
}

func TestBlbB1ConductElectricitySecondTargetMustBeAToken(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	a := blbB1Push(g, opp.ID, "Big", "Creature — Beast", "", 7, 7)
	b := blbB1Push(g, opp.ID, "Not A Token", "Creature — Beast", "", 7, 7)
	if err := castCatalogSpellErr(t, g, "Conduct Electricity", "Instant", blbB1ConductOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1},
	}); err == nil {
		t.Error("a nontoken creature was accepted as the creature token")
	}
}

func TestBlbB1ConductElectricityMayHitOneTokenTwice(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	tok := TokenCard("1/1 blue Fish")
	tok.Power, tok.Toughness = 9, 9
	fish := pushToken(g, opp.ID, tok)
	castCatalogSpell(t, g, "Conduct Electricity", "Instant", blbB1ConductOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: fish, Slot: 0}, {Kind: game.TargetCard, ID: fish, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, fish); got != 8 {
		t.Errorf("the token has %d damage, want 8 (CR 115.3)", got)
	}
}

// --- Crumb and Get It --------------------------------------------------------

func TestBlbB1CrumbAndGetItPromisedGivesFoodAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	mustCastWithGift(t, g, "Crumb and Get It", "Instant", blbB1CrumbOracle, cardTarget(bear), opp.ID)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, opp.ID, "Food"); n != 1 {
		t.Errorf("the opponent has %d Food tokens, want 1", n)
	}
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the bear is %d/%d, want 4/4", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("the promised gift should make the bear indestructible")
	}
}

func TestBlbB1CrumbAndGetItUnpromisedOnlyPumps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if _, err := castWithGift(t, g, "Crumb and Get It", "Instant", blbB1CrumbOracle, cardTarget(theirs), uuid.Nil); err == nil {
		t.Fatal("an opponent's creature was accepted; the target is a creature you control")
	}
	mustCastWithGift(t, g, "Crumb and Get It", "Instant", blbB1CrumbOracle, cardTarget(bear), uuid.Nil)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, opp.ID, "Food"); n != 0 {
		t.Errorf("the opponent has %d Food tokens, want 0", n)
	}
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the bear is %d/%d, want 4/4", p, tough)
	}
	if hasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("without the gift the bear must not be indestructible")
	}
}

// --- Blooming Blast ----------------------------------------------------------

func TestBlbB1BloomingBlastPromisedHitsTheCreaturesController(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, other := blbB1Seats(g)
	// The gift goes to one opponent; the creature is another's.
	victim := blbB1Push(g, other.ID, "Their Big", "Creature — Beast", "", 5, 5)
	oppLife, otherLife := lifeOf(g, opp.ID), lifeOf(g, other.ID)
	mustCastWithGift(t, g, "Blooming Blast", "Instant", blbB1BlastOracle, cardTarget(victim), opp.ID)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, opp.ID, "Treasure"); n != 1 {
		t.Errorf("the gift's opponent has %d Treasure tokens, want 1", n)
	}
	if got := damageOn(t, g, victim); got != 2 {
		t.Errorf("the creature has %d damage, want 2", got)
	}
	if got := lifeOf(g, other.ID); got != otherLife-3 {
		t.Errorf("the creature's controller is at %d, want %d", got, otherLife-3)
	}
	if got := lifeOf(g, opp.ID); got != oppLife {
		t.Errorf("the gift's opponent is at %d, want %d", got, oppLife)
	}
}

func TestBlbB1BloomingBlastUnpromisedOnlyHitsTheCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	victim := blbB1Push(g, opp.ID, "Their Big", "Creature — Beast", "", 5, 5)
	oppLife := lifeOf(g, opp.ID)
	mustCastWithGift(t, g, "Blooming Blast", "Instant", blbB1BlastOracle, cardTarget(victim), uuid.Nil)
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, victim); got != 2 {
		t.Errorf("the creature has %d damage, want 2", got)
	}
	if got := lifeOf(g, opp.ID); got != oppLife {
		t.Errorf("the creature's controller is at %d, want %d", got, oppLife)
	}
	if n, _ := countTokensNamed(g, opp.ID, "Treasure"); n != 0 {
		t.Errorf("%d Treasure tokens without the gift, want 0", n)
	}
}
