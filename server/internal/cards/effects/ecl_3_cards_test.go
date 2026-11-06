package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_3_cards_test.go — Lorwyn Eclipsed, third batch (#4458). Each test
// plays the real catalog entry through the engine and asserts what its
// text says.

const (
	ecl3BoulderDashOracle  = "30a9ad0e-1b23-41ce-bacc-3b27f78ef628"
	ecl3BruteOracle        = "eb25920a-3162-4838-b092-09cbd43332fd"
	ecl3OutriderOracle     = "48f7217e-ae11-44bd-ac42-7c8d9dd5dc36"
	ecl3StonebackOracle    = "8a5e17a0-6f3c-49ed-b328-ed90a46ed6c5"
	ecl3WayfinderOracle    = "93409d98-93c8-4742-a635-2a4408e66c32"
	ecl3GrasplingOracle    = "930035c6-cde7-40ac-98ef-0840812037db"
	ecl3ChompingOracle     = "d740dbd9-8e90-4121-8d53-c6ddf5178d58"
	ecl3FestivalOracle     = "5396045d-744b-435e-8324-a1bfbb03449d"
	ecl3CreakwoodOracle    = "f9b7bbea-ea20-400b-a5b7-4d6993d8777b"
	ecl3CribSwapOracle     = "2987c385-011a-4032-a516-a46d1e9dc9e8"
	ecl3GiantWarrior       = "Creature — Giant Warrior"
	ecl3Shapeshifter       = "Creature — Shapeshifter"
	ecl3FestivalTypeLine   = "Kindred Enchantment — Kithkin"
	ecl3CribSwapTypeLine   = "Kindred Instant — Shapeshifter"
	ecl3CreakwoodTypeLine  = "Creature — Elf Warrior"
	ecl3OutriderTypeLine   = "Creature — Kithkin Knight"
	ecl3ShapeshifterTokens = "Shapeshifter"
)

func TestEcl3EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		ecl3BoulderDashOracle: "Boulder Dash",
		ecl3BruteOracle:       "Brambleback Brute",
		ecl3OutriderOracle:    "Bristlebane Outrider",
		ecl3StonebackOracle:   "Burdened Stoneback",
		ecl3WayfinderOracle:   "Changeling Wayfinder",
		ecl3GrasplingOracle:   "Chitinous Graspling",
		ecl3ChompingOracle:    "Chomping Changeling",
		ecl3FestivalOracle:    "Clachan Festival",
		ecl3CreakwoodOracle:   "Creakwood Safewright",
		ecl3CribSwapOracle:    "Crib Swap",
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

// --- Boulder Dash ---------------------------------------------------------

func TestEcl3BoulderDashDealsTwoAndOneToDifferentTargets(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Squire", "Creature — Human", 1, 1)
	before := opp.Life
	castCatalogSpell(t, g, "Boulder Dash", "Sorcery", ecl3BoulderDashOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: small, Slot: 1}})
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("the first target is dealt 2: life %d -> %d", before, opp.Life)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the second target was dealt 1 and should have died")
	}
}

func TestEcl3BoulderDashRefusesTheSameTargetTwice(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	err := castCatalogSpellErr(t, g, "Boulder Dash", "Sorcery", ecl3BoulderDashOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}})
	if err == nil {
		t.Fatal("the same player was accepted for \"any other target\"")
	}
}

// --- Brambleback Brute ----------------------------------------------------

func TestEcl3BramblebackBruteStopsATargetFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brute := rfdCast(t, g, "Brambleback Brute", ecl3GiantWarrior, ecl3BruteOracle, 4, 5)
	if got := countersOn(g, brute, game.CounterMinusOne); got != 2 {
		t.Fatalf("entered with %d -1/-1 counters, want 2", got)
	}
	if p, tt := ptOf(t, g, brute); p != 2 || tt != 3 {
		t.Errorf("Brute is %d/%d, want 2/3", p, tt)
	}
	attacker := b12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, brute, 0, game.ActivateAbilityParams{
		Targets: cardRefs(wall), CounterKind: game.CounterMinusOne,
	})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, brute, game.CounterMinusOne); got != 1 {
		t.Errorf("the cost removed a counter: %d left, want 1", got)
	}
	brAttack(t, g, attacker)
	brRefusal(t, g.DeclareBlocker(wall, attacker), game.BlockReasonCantBlock)
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Fatalf("an untargeted creature still blocks: %v", err)
	}
}

func TestEcl3BramblebackBruteActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brute := rfdCast(t, g, "Brambleback Brute", ecl3GiantWarrior, ecl3BruteOracle, 4, 5)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, brute, 0, game.ActivateAbilityParams{
		Targets: cardRefs(wall), CounterKind: game.CounterMinusOne,
	}); err == nil {
		t.Error("activated outside a main phase")
	}
	if got := countersOn(g, brute, game.CounterMinusOne); got != 2 {
		t.Errorf("a refused activation paid its cost: %d counters left", got)
	}
}

// --- Bristlebane Outrider -------------------------------------------------

func TestEcl3BristlebaneOutriderGetsTwoWhenAnotherCreatureEntered(t *testing.T) {
	g := newCatalogGame(t)
	rider := rfdCast(t, g, "Bristlebane Outrider", ecl3OutriderTypeLine, ecl3OutriderOracle, 3, 5)
	if p, tt := ptOf(t, g, rider); p != 3 || tt != 5 {
		t.Fatalf("only the Outrider itself entered: it is %d/%d, want 3/5", p, tt)
	}
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, rider); p != 5 || tt != 5 {
		t.Errorf("another creature entered: the Outrider is %d/%d, want 5/5", p, tt)
	}
}

func TestEcl3BristlebaneOutriderIgnoresAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := b12Push(g, me.ID, "Bristlebane Outrider", ecl3OutriderTypeLine, ecl3OutriderOracle, 3, 5)
	if p, tt := ptOf(t, g, rider); p != 3 || tt != 5 {
		t.Fatalf("nothing entered: the Outrider is %d/%d, want 3/5", p, tt)
	}
	// An opponent's creature entering under their control is not "under
	// your control".
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		if _, err := g.CreateTokensForEffect(opp.ID, game.Card{Name: "Theirs", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2}, 1, game.TokenEntryOptions{}); err != nil {
			t.Errorf("CreateTokensForEffect: %v", err)
		}
	})
	if p, tt := ptOf(t, g, rider); p != 3 || tt != 5 {
		t.Errorf("an opponent's creature entered: the Outrider is %d/%d, want 3/5", p, tt)
	}
}

func TestEcl3BristlebaneOutriderCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := b12Push(g, me.ID, "Bristlebane Outrider", ecl3OutriderTypeLine, ecl3OutriderOracle, 3, 5)
	small := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttack(t, g, rider)
	brRefusal(t, g.DeclareBlocker(small, rider), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, rider); err != nil {
		t.Fatalf("a power 3 creature blocks: %v", err)
	}
}

// --- Burdened Stoneback ---------------------------------------------------

func TestEcl3BurdenedStonebackGrantsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := rfdCast(t, g, "Burdened Stoneback", ecl3GiantWarrior, ecl3StonebackOracle, 4, 4)
	if got := countersOn(g, stone, game.CounterMinusOne); got != 2 {
		t.Fatalf("entered with %d -1/-1 counters, want 2", got)
	}
	if p, tt := ptOf(t, g, stone); p != 2 || tt != 2 {
		t.Errorf("Stoneback is %d/%d, want 2/2", p, tt)
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear), CounterKind: game.CounterMinusOne,
	})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Fatal("the target did not gain indestructible")
	}
	if got := countersOn(g, stone, game.CounterMinusOne); got != 1 {
		t.Errorf("the cost removed a counter: %d left, want 1", got)
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{bear}) })
	if !g.Battlefield.Contains(bear) {
		t.Error("an indestructible creature was destroyed")
	}
}

func TestEcl3BurdenedStonebackActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stone := rfdCast(t, g, "Burdened Stoneback", ecl3GiantWarrior, ecl3StonebackOracle, 4, 4)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{
		Targets: cardRefs(stone), CounterKind: game.CounterMinusOne,
	}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Changeling Wayfinder -------------------------------------------------

func TestEcl3ChangelingWayfinderFindsABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nonbasic := pushLibraryCardForTest(me, game.Card{Name: "Desert", TypeLine: "Land — Desert"})
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	way := rfdCast(t, g, "Changeling Wayfinder", ecl3Shapeshifter, ecl3WayfinderOracle, 1, 2)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land card was not searched into the hand")
	}
	if me.Hand.Contains(nonbasic) {
		t.Error("a nonbasic land was found")
	}
	if !hasEffectiveKeyword(t, g, way, game.KeywordChangeling) {
		t.Error("no changeling")
	}
	if c, ok := battlefieldCard(g, way); !ok || !c.HasSubtype("Elf") {
		t.Error("the Wayfinder is not every creature type")
	}
}

// --- Chitinous Graspling --------------------------------------------------

func TestEcl3ChitinousGrasplingHasChangelingAndReach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grasp := b12Push(g, opp.ID, "Chitinous Graspling", ecl3Shapeshifter, ecl3GrasplingOracle, 3, 4)
	if !hasEffectiveKeyword(t, g, grasp, game.KeywordChangeling) || !hasEffectiveKeyword(t, g, grasp, "reach") {
		t.Fatal("changeling and reach are printed")
	}
	if c, ok := battlefieldCard(g, grasp); !ok || !c.HasSubtype("Goblin") {
		t.Error("the Graspling is not every creature type")
	}
	flier := b12Creature(g, me.ID, "Bird", "Creature — Bird", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	brAttack(t, g, flier)
	if err := g.DeclareBlocker(grasp, flier); err != nil {
		t.Fatalf("reach blocks a flier: %v", err)
	}
}

// --- Chomping Changeling --------------------------------------------------

func TestEcl3ChompingChangelingDestroysUpToOneArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	chomp := castHoldCreature(t, g, "Chomping Changeling", ecl3Shapeshifter, ecl3ChompingOracle, 1, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if hasID(pick.PickTargetCards, land) {
		t.Error("a land is not an artifact or enchantment")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Graveyard.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if !hasEffectiveKeyword(t, g, chomp, game.KeywordChangeling) {
		t.Error("no changeling")
	}
}

func TestEcl3ChompingChangelingWithNothingToDestroy(t *testing.T) {
	g := newCatalogGame(t)
	chomp := rfdCast(t, g, "Chomping Changeling", ecl3Shapeshifter, ecl3ChompingOracle, 1, 2)
	if !g.Battlefield.Contains(chomp) {
		t.Error("the Changeling did not resolve with no artifact or enchantment around")
	}
}

// --- Clachan Festival -----------------------------------------------------

func TestEcl3ClachanFestivalMakesKithkin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fest := castCatalogSpell(t, g, "Clachan Festival", ecl3FestivalTypeLine, ecl3FestivalOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fest) {
		t.Fatal("Clachan Festival did not resolve")
	}
	if got := rfdCountNamed(g, me.ID, "Kithkin"); got != 2 {
		t.Fatalf("Kithkin tokens on entry = %d, want 2", got)
	}
	k := findBattlefieldByName(g, "Kithkin")
	if p, tt := ptOf(t, g, k); p != 1 || tt != 1 {
		t.Errorf("a Kithkin token is %d/%d, want 1/1", p, tt)
	}
	if c, ok := battlefieldCard(g, k); !ok || !c.HasColor("G") || !c.HasColor("W") {
		t.Error("the Kithkin token is green and white")
	}
	b16Activate(t, g, me.ID, fest, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := rfdCountNamed(g, me.ID, "Kithkin"); got != 3 {
		t.Errorf("Kithkin tokens after one activation = %d, want 3", got)
	}
}

// --- Creakwood Safewright -------------------------------------------------

func TestEcl3CreakwoodSafewrightShedsACounterWithAnElfInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := rfdCast(t, g, "Creakwood Safewright", ecl3CreakwoodTypeLine, ecl3CreakwoodOracle, 5, 5)
	if got := countersOn(g, elf, game.CounterMinusOne); got != 3 {
		t.Fatalf("entered with %d -1/-1 counters, want 3", got)
	}
	if p, tt := ptOf(t, g, elf); p != 2 || tt != 2 {
		t.Errorf("Safewright is %d/%d, want 2/2", p, tt)
	}
	pushGraveyardCardTyped(me, "Fallen Elf", "Creature — Elf Druid")
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elf, game.CounterMinusOne); got != 2 {
		t.Errorf("after your end step: %d counters, want 2", got)
	}
}

func TestEcl3CreakwoodSafewrightNeedsAnElfCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := rfdCast(t, g, "Creakwood Safewright", ecl3CreakwoodTypeLine, ecl3CreakwoodOracle, 5, 5)
	pushGraveyardCardTyped(me, "Fallen Bear", "Creature — Bear")
	advanceTo(t, g, game.StepEnd)
	if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority: %v", err)
	}
	if triggerOnStack(g, elf) != nil {
		t.Error("the Safewright triggered with no Elf card in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elf, game.CounterMinusOne); got != 3 {
		t.Errorf("no Elf card: %d counters, want 3", got)
	}
}

func TestEcl3CreakwoodSafewrightNeedsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := rfdCast(t, g, "Creakwood Safewright", ecl3CreakwoodTypeLine, ecl3CreakwoodOracle, 5, 5)
	setCountersOn(g, elf, map[string]int{})
	pushGraveyardCardTyped(me, "Fallen Elf", "Creature — Elf Druid")
	advanceTo(t, g, game.StepEnd)
	if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority: %v", err)
	}
	if triggerOnStack(g, elf) != nil {
		t.Error("the Safewright triggered with no -1/-1 counter on it")
	}
}

// --- Crib Swap ------------------------------------------------------------

func TestEcl3CribSwapExilesAndGivesItsControllerAShapeshifter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := b12Creature(g, opp.ID, "Dragon", "Creature — Dragon", 6, 6)
	castCatalogSpell(t, g, "Crib Swap", ecl3CribSwapTypeLine, ecl3CribSwapOracle, cardRefs(dragon))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dragon) || !g.Exile.Contains(dragon) {
		t.Fatal("the target was not exiled")
	}
	if got := rfdCountNamed(g, opp.ID, ecl3ShapeshifterTokens); got != 1 {
		t.Fatalf("its controller has %d Shapeshifter tokens, want 1", got)
	}
	if got := rfdCountNamed(g, me.ID, ecl3ShapeshifterTokens); got != 0 {
		t.Errorf("the caster got %d Shapeshifter tokens, want 0", got)
	}
	tok := findBattlefieldByName(g, ecl3ShapeshifterTokens)
	if p, tt := ptOf(t, g, tok); p != 1 || tt != 1 {
		t.Errorf("the token is %d/%d, want 1/1", p, tt)
	}
	if c, ok := battlefieldCard(g, tok); !ok || !c.HasSubtype("Goblin") {
		t.Error("the token has changeling")
	}
}

func TestEcl3CribSwapIsEveryCreatureTypeInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	swap := castCatalogSpell(t, g, "Crib Swap", ecl3CribSwapTypeLine, ecl3CribSwapOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if got := rfdCountNamed(g, me.ID, ecl3ShapeshifterTokens); got != 1 {
		t.Errorf("your own creature's token goes to you: %d, want 1", got)
	}
	if !me.Graveyard.Contains(swap) {
		t.Fatal("Crib Swap is not in the graveyard")
	}
	found := false
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == swap {
			found = c.HasSubtype("Elf")
		}
	}
	if !found {
		t.Error("Crib Swap in the graveyard is not an Elf")
	}
}
