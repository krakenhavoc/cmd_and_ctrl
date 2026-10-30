package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_1_cards_test.go — Lorwyn Eclipsed, first batch (#4458). Each test
// plays the real catalog entry and asserts what its text says.

const (
	ecl1CatharsisOracle   = "9e3303aa-2484-4832-86ac-f5ae290d02f1"
	ecl1EmptinessOracle   = "dc5dd7f7-375e-4d51-85cb-1abf16ca3854"
	ecl1VibranceOracle    = "52867cdd-8bf2-4cae-bb8d-0f16a18a455b"
	ecl1WistfulnessOracle = "e5e8fce4-c97a-4ae4-9583-003c565e3bd6"
	ecl1SafewrightOracle  = "ce19b337-9266-471d-bec3-49eb7f0e62ac"
	ecl1WatershaperOracle = "602cce3e-7033-47d6-aec1-d755f234f270"
	ecl1BattlerOracle     = "592689f2-f17c-40d6-b880-431f4b6a6b0c"
	ecl1NavigatorOracle   = "812df594-75af-494f-be09-5f9633aca46a"
	ecl1DisruptorOracle   = "e76ae81f-729c-4def-ae83-6870a696cbe8"
	ecl1SpeakerOracle     = "5bd06f9e-9a42-455d-ad26-24bc04e51c7c"
	ecl1GuardianOracle    = "97febf6b-302c-4c41-8838-eef7c7353bdc"
	ecl1GloomRipperOracle = "78e1beca-840d-47b2-8759-cc31454fdf5a"
	ecl1CrescendoOracle   = "b93cd38a-1d0c-40e6-aaf1-621a2c88bf46"
	ecl1KinbindingOracle  = "da35de40-6fde-4018-bda8-b0ae91ed872a"
	ecl1NurseryOracle     = "8bcbae66-1854-480e-a449-33e5389610ab"
	ecl1TarfireOracle     = "ff5502d2-37f5-4ad5-bee2-3a03ec973268"

	ecl1Incarnation = "Creature — Elemental Incarnation"
)

func TestEcl1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		ecl1CatharsisOracle:   "Catharsis",
		ecl1EmptinessOracle:   "Emptiness",
		ecl1VibranceOracle:    "Vibrance",
		ecl1WistfulnessOracle: "Wistfulness",
		ecl1SafewrightOracle:  "Selfless Safewright",
		ecl1WatershaperOracle: "Adept Watershaper",
		ecl1BattlerOracle:     "Bristlebane Battler",
		ecl1NavigatorOracle:   "Deepway Navigator",
		ecl1DisruptorOracle:   "Disruptor of Currents",
		ecl1SpeakerOracle:     "Formidable Speaker",
		ecl1GuardianOracle:    "Glen Elendra Guardian",
		ecl1GloomRipperOracle: "Gloom Ripper",
		ecl1CrescendoOracle:   "Harmonized Crescendo",
		ecl1KinbindingOracle:  "Kinbinding",
		ecl1NurseryOracle:     "Sapling Nursery",
		ecl1TarfireOracle:     "Lasting Tarfire",
	}
	for oracle, name := range want {
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

// ecl1CastIncarnation casts an Incarnation out of a floated pool under
// the strict gate, so the record of what paid for it is exact, and passes
// priority until the stack is empty or a prompt opens. alt is "" or
// "evoke".
func ecl1CastIncarnation(t *testing.T, g *game.Game, me *game.Player, name, cost, oracle, pool, alt string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	floatForTest(g, me, pool)
	id := castFromHandForTest(t, g, me, name, ecl1Incarnation, cost, oracle,
		game.CastSpellParams{Strict: true, AlternativeCost: alt})
	for i := 0; i < 12 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return id
}

// ecl1Named counts the permanents `controller` controls with the name.
func ecl1Named(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// --- Catharsis ------------------------------------------------------------

func TestEcl1CatharsisReadsTheManaSpentToCastIt(t *testing.T) {
	for _, tc := range []struct {
		name, pool string
		kithkin    int
		pump       bool
	}{
		{"white and white", "WWCCCC", 2, false},
		{"red and red", "RRCCCC", 0, true},
		{"one of each", "RWCCCC", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			cath := ecl1CastIncarnation(t, g, me, "Catharsis", "{4}{R/W}{R/W}", ecl1CatharsisOracle, tc.pool, "")
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(cath) {
				t.Fatal("Catharsis did not resolve")
			}
			if got := ecl1Named(g, me.ID, "Kithkin"); got != tc.kithkin {
				t.Errorf("Kithkin tokens = %d, want %d", got, tc.kithkin)
			}
			if tc.kithkin > 0 {
				k := findBattlefieldByName(g, "Kithkin")
				if p, tt := ptOf(t, g, k); p != 1 || tt != 1 {
					t.Errorf("a Kithkin token is %d/%d, want 1/1", p, tt)
				}
			}
			p, tt := ptOf(t, g, bear)
			haste := hasEffectiveKeyword(t, g, bear, "haste")
			if tc.pump && (p != 3 || tt != 3 || !haste) {
				t.Errorf("{R}{R}: the Bear is %d/%d, haste %v; want 3/3 with haste", p, tt, haste)
			}
			if !tc.pump && (p != 2 || tt != 2 || haste) {
				t.Errorf("no {R}{R}: the Bear is %d/%d, haste %v; want a plain 2/2", p, tt, haste)
			}
		})
	}
}

// Evoked for {W}{W}, the Kithkin trigger still fires, and Catharsis is
// sacrificed (CR 702.74a).
func TestEcl1CatharsisEvokedForWhiteMakesKithkinAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cath := ecl1CastIncarnation(t, g, me, "Catharsis", "{4}{R/W}{R/W}", ecl1CatharsisOracle, "WW", "evoke")
	answerTriggerOrderLastQueuedFirst(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cath) || !me.Graveyard.Contains(cath) {
		t.Error("an evoked Catharsis is sacrificed")
	}
	if got := ecl1Named(g, me.ID, "Kithkin"); got != 2 {
		t.Errorf("Kithkin tokens = %d, want 2", got)
	}
}

// --- Emptiness ------------------------------------------------------------

func TestEcl1EmptinessWhiteReturnsASmallCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := uuid.New()
	big := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: small, Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{1}{W}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big", TypeLine: "Creature — Beast", ManaCost: "{3}{G}",
			Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	})
	ecl1CastIncarnation(t, g, me, "Emptiness", "{4}{W/B}{W/B}", ecl1EmptinessOracle, "WWCCCC", "")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the {W}{W} trigger")
	}
	if !hasID(pick.PickTargetCards, small) || hasID(pick.PickTargetCards, big) {
		t.Fatalf("offered %v: want the mana value 2 card and not the mana value 4 one", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the creature card did not return to the battlefield")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the {B}{B} trigger fired without {B}{B} spent")
	}
}

func TestEcl1EmptinessBlackPutsThreeCountersOnUpToOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Giant", 5, 5)
	ecl1CastIncarnation(t, g, me, "Emptiness", "{4}{W/B}{W/B}", ecl1EmptinessOracle, "BBCCCC", "")
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, victim, game.CounterMinusOne); got != 3 {
		t.Errorf("-1/-1 counters on the target = %d, want 3", got)
	}
	if p, tt := ptOf(t, g, victim); p != 2 || tt != 2 {
		t.Errorf("the target is %d/%d, want 2/2", p, tt)
	}
}

// --- Vibrance -------------------------------------------------------------

func TestEcl1VibranceRedDealsThreeToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life, mine := opp.Life, me.Life
	ecl1CastIncarnation(t, g, me, "Vibrance", "{3}{R/G}{R/G}", ecl1VibranceOracle, "RRCCC", "")
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	if me.Life != mine {
		t.Errorf("no {G}{G}: life %d -> %d, want no change", mine, me.Life)
	}
}

func TestEcl1VibranceGreenFindsALandAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushLibraryCardForTest(me, game.Card{Name: "Desert", TypeLine: "Land — Desert"})
	life := me.Life
	ecl1CastIncarnation(t, g, me, "Vibrance", "{3}{R/G}{R/G}", ecl1VibranceOracle, "GGCCC", "")
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card was not searched into the hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

// --- Wistfulness ----------------------------------------------------------

func TestEcl1WistfulnessGreenExilesAnOpponentsArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	mine := b12Permanent(g, me.ID, "My Rock", "Artifact")
	hand := me.Hand.Size()
	ecl1CastIncarnation(t, g, me, "Wistfulness", "{3}{G/U}{G/U}", ecl1WistfulnessOracle, "GGCCC", "")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the {G}{G} trigger")
	}
	if hasID(pick.PickTargetCards, mine) {
		t.Error("an artifact you control is not one an opponent controls")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(rock) {
		t.Error("the opponent's artifact was not exiled")
	}
	if me.Hand.Size() != hand {
		t.Errorf("no {U}{U}: the hand went %d -> %d", hand, me.Hand.Size())
	}
}

func TestEcl1WistfulnessBlueDrawsTwoThenDiscardsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	ecl1CastIncarnation(t, g, me, "Wistfulness", "{3}{G/U}{G/U}", ecl1WistfulnessOracle, "UUCCC", "")
	passPriorityAroundTable(t, g)
	if c := discardChoiceFor(g, me.ID); c == nil {
		t.Fatal("no discard prompt after the draw")
	}
	if me.Hand.Size() != hand+2 {
		t.Fatalf("drew %d before the discard, want 2", me.Hand.Size()-hand)
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 (draw two, discard one)", hand, me.Hand.Size())
	}
}

// --- Selfless Safewright --------------------------------------------------

func TestEcl1SelflessSafewrightProtectsTheOtherPermanentsOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elf := pushTribalCreature(g, me.ID, "Elf", "Creature — Elf Druid", 1, 1)
	changeling := pushTribalCreature(g, me.ID, "Changeling", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	goblin := pushTribalCreature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	theirs := pushTribalCreature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	sw := rfdCast(t, g, "Selfless Safewright", "Creature — Elf Warrior", ecl1SafewrightOracle, 4, 2)
	answerCreatureType(t, g, me.ID, "Elf")
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{elf, changeling} {
		if !hasEffectiveKeyword(t, g, id, "hexproof") || !hasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("%s: an Elf you control gains hexproof and indestructible", id)
		}
	}
	for _, id := range []uuid.UUID{goblin, theirs, sw} {
		if hasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("%s: only OTHER Elves you control are protected", id)
		}
	}
}

// --- Adept Watershaper ----------------------------------------------------

func TestEcl1AdeptWatershaperGivesOtherTappedCreaturesIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ws := b12Push(g, me.ID, "Adept Watershaper", "Creature — Merfolk Cleric", ecl1WatershaperOracle, 3, 4)
	tapped := b12Creature(g, me.ID, "Tapped", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, me.ID, "Untapped", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	tapForTest(t, g, tapped, ws, theirs)
	if !hasEffectiveKeyword(t, g, tapped, "indestructible") {
		t.Error("a tapped creature you control has indestructible")
	}
	for _, id := range []uuid.UUID{untapped, ws, theirs} {
		if hasEffectiveKeyword(t, g, id, "indestructible") {
			t.Errorf("%s: untapped, the Watershaper itself, or an opponent's: no indestructible", id)
		}
	}
	if err := g.TapCard(tapped, false); err != nil {
		t.Fatalf("untap: %v", err)
	}
	if hasEffectiveKeyword(t, g, tapped, "indestructible") {
		t.Error("untapped again, it loses indestructible")
	}
}

// --- Bristlebane Battler --------------------------------------------------

func TestEcl1BristlebaneBattlerShedsACounterWhenAnotherCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	battler := rfdCast(t, g, "Bristlebane Battler", "Creature — Kithkin Soldier", ecl1BattlerOracle, 6, 6)
	if got := countersOn(g, battler, game.CounterMinusOne); got != 5 {
		t.Fatalf("entered with %d -1/-1 counters, want 5", got)
	}
	if p, tt := ptOf(t, g, battler); p != 1 || tt != 1 {
		t.Errorf("Battler is %d/%d, want 1/1", p, tt)
	}
	if !hasEffectiveKeyword(t, g, battler, "trample") {
		t.Error("no trample")
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, battler, game.CounterMinusOne); got != 4 {
		t.Errorf("after another creature entered: %d counters, want 4", got)
	}
	if p, tt := ptOf(t, g, battler); p != 2 || tt != 2 {
		t.Errorf("Battler is %d/%d, want 2/2", p, tt)
	}
}

// With no -1/-1 counter left, the trigger condition is false and nothing
// goes on the stack.
func TestEcl1BristlebaneBattlerDoesNotTriggerWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	battler := rfdCast(t, g, "Bristlebane Battler", "Creature — Kithkin Soldier", ecl1BattlerOracle, 6, 6)
	setCountersOn(g, battler, map[string]int{})
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority: %v", err)
	}
	if triggerOnStack(g, battler) != nil {
		t.Error("the Battler triggered with no -1/-1 counter on it")
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, battler); p != 6 || tt != 6 {
		t.Errorf("Battler is %d/%d, want 6/6", p, tt)
	}
}

// --- Deepway Navigator ----------------------------------------------------

func TestEcl1DeepwayNavigatorUntapsMerfolkAndRewardsThreeMerfolkAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "Merfolk A", "Creature — Merfolk Soldier", 2, 2)
	b := b12Creature(g, me.ID, "Merfolk B", "Creature — Merfolk Soldier", 2, 2)
	c := b12Creature(g, me.ID, "Merfolk C", "Creature — Merfolk Soldier", 2, 2)
	home := b12Creature(g, me.ID, "Merfolk Home", "Creature — Merfolk Soldier", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)
	tapForTest(t, g, a, b, bear, theirs)

	nav := rfdCast(t, g, "Deepway Navigator", "Creature — Merfolk Wizard", ecl1NavigatorOracle, 2, 2)
	if b16Tapped(t, g, a) || b16Tapped(t, g, b) {
		t.Error("each other Merfolk you control untaps")
	}
	if !b16Tapped(t, g, bear) || !b16Tapped(t, g, theirs) {
		t.Error("a non-Merfolk and an opponent's Merfolk stay tapped")
	}
	if p, _ := ptOf(t, g, home); p != 2 {
		t.Fatalf("before any attack a Merfolk has power %d, want 2", p)
	}

	declareAttack(t, g, opp.ID, a, b, c)
	if p, _ := ptOf(t, g, home); p != 3 {
		t.Errorf("three Merfolk attacked: a Merfolk at home has power %d, want 3", p)
	}
	if p, _ := ptOf(t, g, nav); p != 3 {
		t.Errorf("the Navigator is a Merfolk too: power %d, want 3", p)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Errorf("a Bear is not a Merfolk: power %d, want 2", p)
	}
	if p, _ := ptOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Merfolk: power %d, want 2", p)
	}
}

func TestEcl1DeepwayNavigatorTwoMerfolkAttackersAreNotEnough(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	b12Push(g, me.ID, "Deepway Navigator", "Creature — Merfolk Wizard", ecl1NavigatorOracle, 2, 2)
	a := b12Creature(g, me.ID, "Merfolk A", "Creature — Merfolk Soldier", 2, 2)
	b := b12Creature(g, me.ID, "Merfolk B", "Creature — Merfolk Soldier", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b, bear)
	if p, _ := ptOf(t, g, a); p != 2 {
		t.Errorf("two Merfolk and a Bear attacked: power %d, want 2", p)
	}
}

// --- Disruptor of Currents ------------------------------------------------

func TestEcl1DisruptorOfCurrentsBouncesUpToOneOtherNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	d := rfdCast(t, g, "Disruptor of Currents", "Creature — Merfolk Wizard", ecl1DisruptorOracle, 3, 3)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if hasID(pick.PickTargetCards, d) || hasID(pick.PickTargetCards, land) {
		t.Error("the Disruptor itself and a land are not legal targets")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
	if !hasEffectiveKeyword(t, g, d, "flash") {
		t.Error("no flash")
	}
}

// --- Formidable Speaker ---------------------------------------------------

func TestEcl1FormidableSpeakerDiscardsToFindACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beast := pushLibraryCardForTest(me, game.Card{Name: "Beast", TypeLine: "Creature — Beast"})
	rock := pushLibraryCardForTest(me, game.Card{Name: "Rock", TypeLine: "Artifact"})
	fodder := handCard(me, "Fodder", "Instant")
	rfdCast(t, g, "Formidable Speaker", "Creature — Elf Druid", ecl1SpeakerOracle, 2, 4)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt")
	}
	answerDiscard(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !me.Hand.Contains(beast) || me.Hand.Contains(rock) {
		t.Error("the search finds a creature card, and only a creature card")
	}
}

func TestEcl1FormidableSpeakerDeclinesTheDiscardAndSearchesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beast := pushLibraryCardForTest(me, game.Card{Name: "Beast", TypeLine: "Creature — Beast"})
	handCard(me, "Fodder", "Instant")
	rfdCast(t, g, "Formidable Speaker", "Creature — Elf Druid", ecl1SpeakerOracle, 2, 4)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if me.Hand.Contains(beast) {
		t.Error("searched although no card was discarded")
	}
}

func TestEcl1FormidableSpeakerUntapsAnotherPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	speaker := b12Push(g, me.ID, "Formidable Speaker", "Creature — Elf Druid", ecl1SpeakerOracle, 2, 4)
	theirs := b12Permanent(g, opp.ID, "Their Land", "Land")
	tapForTest(t, g, theirs)
	if err := g.ActivateCatalogAbility(me.ID, speaker, 0, game.ActivateAbilityParams{Targets: cardRefsOf(speaker)}); err == nil {
		t.Error("the Speaker targeted itself: it untaps ANOTHER permanent")
	}
	if err := g.ActivateCatalogAbility(me.ID, speaker, 0, game.ActivateAbilityParams{Targets: cardRefsOf(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, theirs) {
		t.Error("an opponent's permanent untaps too")
	}
	if !b16Tapped(t, g, speaker) {
		t.Error("{T} is part of the cost")
	}
}

// --- Glen Elendra Guardian ------------------------------------------------

func TestEcl1GlenElendraGuardianCountersANoncreatureSpellAndItsControllerDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	guardian := rfdCast(t, g, "Glen Elendra Guardian", "Creature — Faerie Wizard", ecl1GuardianOracle, 3, 4)
	if got := countersOn(g, guardian, game.CounterMinusOne); got != 1 {
		t.Fatalf("entered with %d -1/-1 counters, want 1", got)
	}
	if p, tt := ptOf(t, g, guardian); p != 2 || tt != 3 {
		t.Errorf("Guardian is %d/%d, want 2/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, guardian, "flying") {
		t.Error("no flying")
	}

	life := opp.Life
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{
		Targets: cardRefsOf(bolt), CounterKind: game.CounterMinusOne,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("the Bolt resolved: opponent %d -> %d", life, opp.Life)
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the countered Bolt is in its owner's graveyard")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("the countered spell's controller draws: hand %d -> %d", hand, me.Hand.Size())
	}
	if got := countersOn(g, guardian, game.CounterMinusOne); got != 0 {
		t.Errorf("the -1/-1 counter paid the cost: %d left", got)
	}
	// No counter left: the cost can't be paid.
	bolt2 := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{
		Targets: cardRefsOf(bolt2), CounterKind: game.CounterMinusOne,
	}); err == nil {
		t.Error("activated with no counter to remove")
	}
}

func TestEcl1GlenElendraGuardianCantCounterACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	guardian := rfdCast(t, g, "Glen Elendra Guardian", "Creature — Faerie Wizard", ecl1GuardianOracle, 3, 4)
	bear := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{
		Targets: cardRefsOf(bear), CounterKind: game.CounterMinusOne,
	}); err == nil {
		t.Error("a creature spell is not a noncreature spell")
	}
}

// --- Gloom Ripper ---------------------------------------------------------

func TestEcl1GloomRipperCountsElvesOnTheBoardAndInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 1, 1)
	batch01GraveyardCard(me, "Dead Elf", "Creature — Elf")
	batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Giant", 2, 5)
	rfdCast(t, g, "Gloom Ripper", "Creature — Elf Assassin", ecl1GloomRipperOracle, 4, 4)
	for i := 0; i < 4; i++ {
		p := latestPickTarget(g, me.ID)
		if p == nil {
			break
		}
		switch {
		case hasID(p.PickTargetCards, elf):
			pickCard(t, g, me.ID, elf)
		case hasID(p.PickTargetCards, victim):
			pickCard(t, g, me.ID, victim)
		default:
			t.Fatalf("unexpected target prompt %v", p.PickTargetCards)
		}
	}
	passPriorityAroundTable(t, g)
	// X = 2 Elves you control (the Elf and the Ripper) + 1 Elf card.
	if p, tt := ptOf(t, g, elf); p != 4 || tt != 1 {
		t.Errorf("your Elf is %d/%d, want 4/1 (+3/+0)", p, tt)
	}
	if p, tt := ptOf(t, g, victim); p != 2 || tt != 2 {
		t.Errorf("the opponent's creature is %d/%d, want 2/2 (-0/-3)", p, tt)
	}
}

// --- Harmonized Crescendo -------------------------------------------------

func TestEcl1HarmonizedCrescendoDrawsPerPermanentOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTribalCreature(g, me.ID, "Merfolk A", "Creature — Merfolk", 1, 1)
	pushTribalCreature(g, me.ID, "Merfolk B", "Creature — Merfolk", 1, 1)
	pushTribalCreature(g, me.ID, "Shapeshifter", "Creature — Shapeshifter", 1, 1, game.KeywordChangeling)
	pushTribalCreature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	pushTribalCreature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	castAndPause(t, g, "Harmonized Crescendo", "Instant", ecl1CrescendoOracle)
	before := handSize(me)
	answerCreatureType(t, g, me.ID, "Merfolk")
	if got := handSize(me) - before; got != 3 {
		t.Errorf("drew %d, want 3 (two Merfolk and a changeling)", got)
	}
	if spec, _ := Lookup(ecl1CrescendoOracle); spec.TapCost == nil {
		t.Error("Harmonized Crescendo has convoke")
	}
}

// --- Kinbinding -----------------------------------------------------------

func TestEcl1KinbindingGrowsWithEachCreatureThatEntersThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kinbinding", "Enchantment", ecl1KinbindingOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if p, tt := ptOf(t, g, bear); p != 2 || tt != 2 {
		t.Fatalf("nothing has entered this turn: the Bear is %d/%d, want 2/2", p, tt)
	}
	castCatalogSpell(t, g, "Cub", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("one creature entered: the Bear is %d/%d, want 3/3", p, tt)
	}
	if p, tt := ptOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("an opponent's creature gets nothing: %d/%d", p, tt)
	}

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	kithkin := findBattlefieldByName(g, "Kithkin")
	if kithkin == uuid.Nil {
		t.Fatal("no Kithkin token at the beginning of combat")
	}
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the Kithkin entered too: the Bear is %d/%d, want 4/4", p, tt)
	}
	if p, tt := ptOf(t, g, kithkin); p != 3 || tt != 3 {
		t.Errorf("the Kithkin is %d/%d, want 3/3", p, tt)
	}
}

// --- Sapling Nursery ------------------------------------------------------

func TestEcl1SaplingNurseryCostsLessPerForest(t *testing.T) {
	for _, tc := range []struct {
		forests int
		casts   bool
	}{{6, true}, {5, false}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		for i := 0; i < tc.forests; i++ {
			b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
		}
		tapForTest(t, g, forestsOf(g, me.ID)...)
		floatForTest(g, me, "GG")
		id := handCardOf(g, me, "Sapling Nursery", "Enchantment", "{6}{G}{G}", ecl1NurseryOracle)
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if tc.casts && err != nil {
			t.Errorf("%d Forests take {%d} off, so {G}{G} casts it: %v", tc.forests, tc.forests, err)
		}
		if !tc.casts && err == nil {
			t.Errorf("%d Forests leave {1}{G}{G} to pay, and {G}{G} is not enough", tc.forests)
		}
	}
}

func forestsOf(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Forest" {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

func TestEcl1SaplingNurseryMakesTreefolkAndProtectsThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nursery := pushCatalogPermanent(g, me.ID, "Sapling Nursery", "Enchantment", ecl1NurseryOracle, false)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	tree := findBattlefieldByName(g, "Treefolk")
	if tree == uuid.Nil {
		t.Fatal("no Treefolk token after a land drop")
	}
	if p, tt := ptOf(t, g, tree); p != 3 || tt != 4 || !hasEffectiveKeyword(t, g, tree, "reach") {
		t.Errorf("the Treefolk is %d/%d, reach %v; want a 3/4 with reach", p, tt, hasEffectiveKeyword(t, g, tree, "reach"))
	}
	forest := b12Permanent(g, me.ID, "Snow Forest", "Basic Snow Land — Forest")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, nursery, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(nursery) {
		t.Error("exiling the enchantment is part of the cost")
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, tree, "indestructible") {
		t.Error("the Treefolk gains indestructible")
	}
	if !hasEffectiveKeyword(t, g, forest, "indestructible") {
		t.Error("the Forest gains indestructible")
	}
	if hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("a Bear is neither a Treefolk nor a Forest")
	}
}

// --- Lasting Tarfire ------------------------------------------------------

func TestEcl1LastingTarfireBurnsOnlyAfterYouPutACounterOnACreature(t *testing.T) {
	for _, tc := range []struct {
		name    string
		counter bool
	}{{"a counter put", true}, {"no counter", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			b12Push(g, me.ID, "Lasting Tarfire", "Enchantment", ecl1TarfireOracle, 0, 0)
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			if tc.counter {
				g.WithWriteLock(func() {
					if err := g.AddCounterByForEffect(me.ID, bear, game.CounterPlusOne, 1); err != nil {
						t.Fatalf("AddCounterByForEffect: %v", err)
					}
				})
			}
			before := lifeOfOpponents(g)
			mine := me.Life
			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
			for i, b := range before {
				want := b
				if tc.counter {
					want = b - 2
				}
				if got := g.Seats[i+1].Life; got != want {
					t.Errorf("opponent %d: %d -> %d, want %d", i+1, b, got, want)
				}
			}
			if me.Life != mine {
				t.Errorf("you are not an opponent: %d -> %d", mine, me.Life)
			}
		})
	}
}
