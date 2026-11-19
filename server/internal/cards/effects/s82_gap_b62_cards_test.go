package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b62_cards_test.go — GAP batch 62 (#5948, S82): Fallout
// creatures. Each test plays the real catalog entry.

const (
	gapb62MrHouseOracle        = "d4e54b48-956f-4467-8485-a5fdb44779fd"
	gapb62RaulOracle           = "d247a86d-917d-4833-82c4-876943d54e2c"
	gapb62ColonelAutumnOracle  = "5c6ce876-781b-437d-b20e-db0b6e2769f4"
	gapb62ArthurMaxsonOracle   = "d95cd013-6f32-4a8b-9c46-a28695eea4f1"
	gapb62EnforcerOracle       = "8e953251-aa5c-4aa3-8e69-1ff1f609f3ff"
	gapb62MacCreadyOracle      = "f40a1a41-d340-4256-8f97-77499cd511a8"
	gapb62GourmandOracle       = "3771c0d7-cea3-4c03-933d-f236f32ef2db"
	gapb62AyulaOracle          = "a79d2fc8-467f-464d-9e73-2dae4b059628"
	gapb62CathedralAcolyteOrcl = "85fa4ed3-14e3-4eb0-b31b-62e2088321d0"
	gapb62GemrazerOracle       = "3dfb0c0a-b68f-43b9-8475-28d0192fc4ed"
	gapb62GunnerConscriptOracl = "f3d3ed94-f32c-4a67-8e2a-765e454a80df"
	gapb62LilyBowenOracle      = "6b2d92f5-b0a1-4f30-9140-da490bf01bda"
)

func TestGapb62CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb62MrHouseOracle:        "Mr. House, President and CEO",
		gapb62RaulOracle:           "Raul, Trouble Shooter",
		gapb62ColonelAutumnOracle:  "Colonel Autumn",
		gapb62ArthurMaxsonOracle:   "Elder Arthur Maxson",
		gapb62EnforcerOracle:       "General's Enforcer",
		gapb62MacCreadyOracle:      "MacCready, Lamplight Mayor",
		gapb62GourmandOracle:       "White Glove Gourmand",
		gapb62AyulaOracle:          "Ayula, Queen Among Bears",
		gapb62CathedralAcolyteOrcl: "Cathedral Acolyte",
		gapb62GemrazerOracle:       "Gemrazer",
		gapb62GunnerConscriptOracl: "Gunner Conscript",
		gapb62LilyBowenOracle:      "Lily Bowen, Raging Grandma",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// --- Gemrazer ----------------------------------------------------------------------

func gapb62GemrazerInHand(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.PlayerByID(owner).Hand.PushTop(game.Card{
		InstanceID: id, Name: "Gemrazer", TypeLine: "Creature — Beast",
		OracleID: gapb62GemrazerOracle, ManaCost: "{3}{G}", Colors: []string{"G"},
		Keywords: []string{"reach", "trample", "mutate"}, Power: 4, Toughness: 4,
		Owner: owner, Controller: owner,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return id
}

func TestGapb62GemrazerMutateDestroysAnOpponentsArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elves := llanowarOn(g, me.ID)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	myRock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	gem := gapb62GemrazerInHand(t, g, me.ID)
	if err := castMutating(g, me.ID, gem, elves, false); err != nil {
		t.Fatalf("cast mutating: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the mutate trigger asked for no target")
	}
	if hasID(p.PickTargetCards, myRock) || hasID(p.PickTargetCards, bear) {
		t.Errorf("illegal targets offered: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the opponent's artifact survived")
	}
	if !g.Battlefield.Contains(myRock) {
		t.Error("my own artifact was destroyed")
	}
	f := readMerged(t, g, elves)
	if f.name != "Gemrazer" || f.power != 4 || !f.tapsForMana {
		t.Errorf("merged creature %+v, want Gemrazer 4/4 with the Elves' mana ability", f)
	}
}

func TestGapb62GemrazerCastNormallyDestroysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	gem := gapb62GemrazerInHand(t, g, me.ID)
	floatMana(t, g, me, "{C}{C}{C}{G}")
	if err := g.CastSpell(me.ID, gem, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || !g.Battlefield.Contains(rock) {
		t.Error("an ordinary cast (nothing mutated) destroyed or asked about a permanent")
	}
	if !g.Battlefield.Contains(gem) {
		t.Error("Gemrazer did not enter")
	}
}

// --- General's Enforcer -------------------------------------------------------------

func TestGapb62EnforcerGivesLegendaryHumansIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	enforcer := gapb62Creature(g, me.ID, "General's Enforcer", "Creature — Human Soldier", gapb62EnforcerOracle, 2, 3)
	legend := gapb62Creature(g, me.ID, "Legend", "Legendary Creature — Human Soldier", "", 2, 2)
	plain := gapb62Creature(g, me.ID, "Plain Human", "Creature — Human Soldier", "", 2, 2)
	legendBear := gapb62Creature(g, me.ID, "Legend Bear", "Legendary Creature — Bear", "", 2, 2)
	theirs := gapb62Creature(g, opp.ID, "Their Legend", "Legendary Creature — Human", "", 2, 2)
	for name, want := range map[string]struct {
		id   uuid.UUID
		want bool
	}{"legendary Human": {legend, true}, "plain Human": {plain, false}, "legendary Bear": {legendBear, false},
		"opponent's legendary Human": {theirs, false}, "Enforcer itself (not legendary)": {enforcer, false}} {
		if got := hasString(effectiveAbilities(t, g, want.id), "indestructible"); got != want.want {
			t.Errorf("%s indestructible: %v, want %v", name, got, want.want)
		}
	}
}

func TestGapb62EnforcerExilesACardAndMakesAHumanSoldierOnlyForACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	enforcer := gapb62Creature(g, me.ID, "General's Enforcer", "Creature — Human Soldier", gapb62EnforcerOracle, 2, 3)
	dead := gfDead(t, g, me, "Dead Bear", "Creature — Bear", "{1}{G}")
	spell := gfDead(t, g, me, "Dead Bolt", "Instant", "{R}")
	for i, tc := range []struct {
		id     uuid.UUID
		tokens int
	}{{dead, 1}, {spell, 1}} {
		floatMana(t, g, me, "{C}{C}{W}{B}")
		if err := g.ActivateCatalogAbility(me.ID, enforcer, 0, game.ActivateAbilityParams{Targets: cardRefs(tc.id), Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if !g.Exile.Contains(tc.id) {
			t.Errorf("card %d was not exiled", i)
		}
		if n := len(gapb62Named(g, me.ID, "Human Soldier")); n != tc.tokens {
			t.Errorf("after exiling card %d: %d Human Soldier tokens, want %d", i, n, tc.tokens)
		}
	}
}

// --- White Glove Gourmand -----------------------------------------------------------

func TestGapb62GourmandMakesTwoSoldiersAndAFoodWhenAnotherHumanDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "White Glove Gourmand", TypeLine: "Creature — Human Noble",
		OracleID: gapb62GourmandOracle, Owner: me.ID, Controller: me.ID})
	floatMana(t, g, me, "{C}{C}{W}{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	soldiers := gapb62Named(g, me.ID, "Human Soldier")
	if len(soldiers) != 2 {
		t.Fatalf("%d Human Soldier tokens, want 2", len(soldiers))
	}
	// A Human token dying counts, and the Gourmand's own death never does.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(soldiers[0]) })
	passPriorityAroundTable(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := len(gapb62Named(g, me.ID, "Food")); n != 1 {
		t.Errorf("%d Food tokens after another Human died, want 1", n)
	}
}

func TestGapb62GourmandNoFoodWithoutAnotherHumanDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gourmand := gapb62Creature(g, me.ID, "White Glove Gourmand", "Creature — Human Noble", gapb62GourmandOracle, 2, 2)
	bear := gapb62Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := len(gapb62Named(g, me.ID, "Food")); n != 0 {
		t.Errorf("%d Food tokens after a Bear died, want 0", n)
	}
	// The Gourmand dying itself is not "another Human".
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	self := gapb62Creature(g2, me2.ID, "White Glove Gourmand", "Creature — Human Noble", gapb62GourmandOracle, 2, 2)
	g2.WithWriteLock(func() { _ = g2.DestroyPermanentForEffect(self) })
	advanceToEndStep(t, g2)
	passPriorityAroundTable(t, g2)
	if n := len(gapb62Named(g2, me2.ID, "Food")); n != 0 {
		t.Errorf("%d Food tokens after the Gourmand itself died, want 0", n)
	}
	_ = gourmand
}

// gapb62Named lists the battlefield permanents with a name under a player.
func gapb62Named(g *game.Game, owner uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == owner {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

// gapb62Creature pushes a catalog creature onto the battlefield.
func gapb62Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// --- Colonel Autumn -----------------------------------------------------------------

func TestGapb62ColonelAutumnExploitPutsACounterOnEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	fodder := gapb62Creature(g, me.ID, "Fodder", "Creature — Goblin", "", 1, 1)
	other := gapb62Creature(g, me.ID, "Other", "Creature — Bear", "", 2, 2)
	theirs := gapb62Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", "", 2, 2)
	colonel := exploitCast(t, g, me, "Colonel Autumn", "Legendary Creature — Human Soldier", gapb62ColonelAutumnOracle)
	answerOwnPermanents(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the exploited creature was not sacrificed")
	}
	for name, id := range map[string]uuid.UUID{"Colonel Autumn": colonel, "the other creature": other} {
		if n := countersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("%s has %d +1/+1 counters after an exploit, want 1", name, n)
		}
	}
	if n := countersOn(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("an opponent's creature got %d counters", n)
	}
}

func TestGapb62ColonelAutumnGivesOtherLegendaryCreaturesExploit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	colonel := gapb62Creature(g, me.ID, "Colonel Autumn", "Legendary Creature — Human Soldier", gapb62ColonelAutumnOracle, 2, 3)
	fodder := gapb62Creature(g, me.ID, "Fodder", "Creature — Goblin", "", 1, 1)
	legend := exploitCast(t, g, me, "Some Legend", "Legendary Creature — Elf", "")
	answerOwnPermanents(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the legendary creature entered with no exploit: the fodder is still there")
	}
	// Colonel Autumn's counter trigger saw the granted exploit.
	for name, id := range map[string]uuid.UUID{"Colonel Autumn": colonel, "the exploiter": legend} {
		if n := countersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", name, n)
		}
	}
}

func TestGapb62ColonelAutumnDoesNotGiveANonlegendaryCreatureExploit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	gapb62Creature(g, me.ID, "Colonel Autumn", "Legendary Creature — Human Soldier", gapb62ColonelAutumnOracle, 2, 3)
	fodder := gapb62Creature(g, me.ID, "Fodder", "Creature — Goblin", "", 1, 1)
	handCardForTest(me, "Plain Elf", "Creature — Elf", "")
	elf := me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil || !g.Battlefield.Contains(fodder) {
		t.Error("a nonlegendary creature was given exploit")
	}
}

// --- Elder Arthur Maxson ------------------------------------------------------------

func TestGapb62MaxsonGivesCreatureTokensTraining(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb62Creature(g, me.ID, "Elder Arthur Maxson", "Legendary Creature — Human Knight", gapb62ArthurMaxsonOracle, 4, 2)
	token := gapb62Creature(g, me.ID, "Goblin", "Token Creature — Goblin", "", 1, 1)
	card := gapb62Creature(g, me.ID, "Card Goblin", "Creature — Goblin", "", 1, 1)
	big := gapb62Creature(g, me.ID, "Big", "Creature — Ogre", "", 5, 5)
	pr7bAttack(t, g, opp.ID, token, card, big)
	if n := countersOn(g, token, game.CounterPlusOne); n != 1 {
		t.Errorf("the token has %d counters after attacking beside a creature with greater power, want 1", n)
	}
	if n := countersOn(g, card, game.CounterPlusOne); n != 0 {
		t.Errorf("a nontoken creature got training (%d counters)", n)
	}
}

func TestGapb62MaxsonSacrificesAnotherCreatureForIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	maxson := gapb62Creature(g, me.ID, "Elder Arthur Maxson", "Legendary Creature — Human Knight", gapb62ArthurMaxsonOracle, 4, 2)
	// Alone, Maxson has nothing to sacrifice: he is not "another creature".
	if err := g.ActivateCatalogAbility(me.ID, maxson, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("Maxson activated with no other creature to sacrifice")
	}
	fodder := gapb62Creature(g, me.ID, "Fodder", "Creature — Goblin", "", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, maxson, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{maxson}, Strict: true}); err == nil {
		t.Fatal("Maxson sacrificed himself to his own ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, maxson, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still there")
	}
	if !hasString(effectiveAbilities(t, g, maxson), "indestructible") {
		t.Error("Maxson is not indestructible")
	}
}

// --- Lily Bowen, Raging Grandma -----------------------------------------------------

func TestGapb62LilyBowenEntersWithTwoCountersAndDoublesEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lily := rfdCast(t, g, "Lily Bowen, Raging Grandma", "Legendary Creature — Mutant Warrior", gapb62LilyBowenOracle, 0, 0)
	if n := countersOn(g, lily, game.CounterPlusOne); n != 2 {
		t.Fatalf("enters with %d counters, want 2", n)
	}
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, lily, game.CounterPlusOne); n != 4 {
		t.Errorf("after one upkeep: %d counters, want 4", n)
	}
	_ = me
}

func TestGapb62LilyBowenAboveSixteenPowerRemovesAllButOneAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lily := gapb62Creature(g, me.ID, "Lily Bowen, Raging Grandma", "Legendary Creature — Mutant Warrior", gapb62LilyBowenOracle, 0, 0)
	if err := g.AddCounter(lily, game.CounterPlusOne, 17); err != nil {
		t.Fatal(err)
	}
	life := me.Life
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, lily, game.CounterPlusOne); n != 1 {
		t.Errorf("a 17-power Lily has %d counters after upkeep, want 1", n)
	}
	if me.Life != life+16 {
		t.Errorf("life %d, want %d (one per counter removed)", me.Life, life+16)
	}
}

func TestGapb62LilyBowenAtExactlySixteenPowerStillDoubles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lily := gapb62Creature(g, me.ID, "Lily Bowen, Raging Grandma", "Legendary Creature — Mutant Warrior", gapb62LilyBowenOracle, 0, 0)
	if err := g.AddCounter(lily, game.CounterPlusOne, 16); err != nil {
		t.Fatal(err)
	}
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, lily, game.CounterPlusOne); n != 32 {
		t.Errorf("a 16-power Lily has %d counters after upkeep, want 32", n)
	}
}

// --- Cathedral Acolyte -------------------------------------------------------------

func TestGapb62CathedralAcolyteCountersTheCreatureThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	acolyte := gapb62Creature(g, me.ID, "Cathedral Acolyte", "Creature — Human Cleric", gapb62CathedralAcolyteOrcl, 1, 2)
	older := gapb62Creature(g, me.ID, "Older", "Creature — Bear", "", 2, 2)
	// Both have been here since an earlier turn: walk to my next main phase.
	advanceToUpkeepOfSeat(t, g, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	fresh := rfdCast(t, g, "Fresh Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Targets: cardRefs(older), Strict: true}); err == nil {
		t.Fatal("the Acolyte targeted a creature that did not enter this turn")
	}
	if err := g.ActivateCatalogAbility(me.ID, acolyte, 0, game.ActivateAbilityParams{Targets: cardRefs(fresh), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, fresh, game.CounterPlusOne); n != 1 {
		t.Errorf("the new creature has %d counters, want 1", n)
	}
}

func TestGapb62CathedralAcolyteGivesCreaturesWithCountersWard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb62Creature(g, me.ID, "Cathedral Acolyte", "Creature — Human Cleric", gapb62CathedralAcolyteOrcl, 1, 2)
	counted := gapb62Creature(g, me.ID, "Counted Bear", "Creature — Bear", "", 2, 2)
	if err := g.AddCounter(counted, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	advanceToMain(t, g)
	g.PassPriority()
	blade := castAtWardedCreature(t, g, opp, counted)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("a creature with a counter on it was not warded")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(counted) || !opp.Graveyard.Contains(blade) {
		t.Error("the unpaid ward did not counter the spell")
	}
}

func TestGapb62CathedralAcolyteDoesNotWardACreatureWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb62Creature(g, me.ID, "Cathedral Acolyte", "Creature — Human Cleric", gapb62CathedralAcolyteOrcl, 1, 2)
	bare := gapb62Creature(g, me.ID, "Bare Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	g.PassPriority()
	castAtWardedCreature(t, g, opp, bare)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, opp.ID) {
		t.Error("a creature with no counter was warded")
	}
	if g.Battlefield.Contains(bare) {
		t.Error("the removal spell did not resolve against an unwarded creature")
	}
}

// --- Ayula, Queen Among Bears -------------------------------------------------------

func gapb62AyulaModes(t *testing.T, g *game.Game, me *game.Player, mode int) {
	t.Helper()
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestGapb62AyulaPutsTwoCountersOnATargetBearWhenAnotherBearEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ayula := gapb62Creature(g, me.ID, "Ayula, Queen Among Bears", "Legendary Creature — Bear", gapb62AyulaOracle, 2, 2)
	theirBear := gapb62Creature(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	gapb62Creature(g, me.ID, "My Wolf", "Creature — Wolf", "", 2, 2)
	rfdCast(t, g, "New Bear", "Creature — Bear", "", 2, 2)
	gapb62AyulaModes(t, g, me, 0)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, theirBear)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, theirBear, game.CounterPlusOne); n != 2 {
		t.Errorf("the targeted Bear (an opponent's) has %d counters, want 2", n)
	}
	_ = ayula
}

func TestGapb62AyulaOnlyTriggersForAnotherBearYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	gapb62Creature(g, me.ID, "Ayula, Queen Among Bears", "Legendary Creature — Bear", gapb62AyulaOracle, 2, 2)
	rfdCast(t, g, "New Wolf", "Creature — Wolf", "", 2, 2)
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("a Wolf entering triggered Ayula")
	}
	// Ayula's own entry is not "another Bear".
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	rfdCast(t, g2, "Ayula, Queen Among Bears", "Legendary Creature — Bear", gapb62AyulaOracle, 2, 2)
	if latestChoiceOfKindFor(g2, game.PendingChoiceModePick, me2.ID) != nil {
		t.Error("Ayula entering triggered herself")
	}
	_ = opp
}

func TestGapb62AyulaFightsAnOpposingCreatureWithABearYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ayula := gapb62Creature(g, me.ID, "Ayula, Queen Among Bears", "Legendary Creature — Bear", gapb62AyulaOracle, 2, 2)
	victim := gapb62Creature(g, opp.ID, "Victim", "Creature — Rat", "", 1, 1)
	rfdCast(t, g, "New Bear", "Creature — Bear", "", 2, 2)
	gapb62AyulaModes(t, g, me, 1)
	for i := 0; i < 2; i++ {
		p := latestPickTarget(g, me.ID)
		if p == nil {
			t.Fatalf("target prompt %d missing", i)
		}
		// First clause: a Bear I control; second: a creature I don't.
		want := ayula
		if i == 1 {
			want = victim
		}
		pickCard(t, g, me.ID, want)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the fought creature survived 2 damage")
	}
	if !g.Battlefield.Contains(ayula) {
		t.Error("Ayula died fighting a 1/1")
	}
}

// --- MacCready, Lamplight Mayor -----------------------------------------------------

func TestGapb62MacCreadyGivesSkulkToSmallAttackersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb62Creature(g, me.ID, "MacCready, Lamplight Mayor", "Legendary Creature — Human Advisor", gapb62MacCreadyOracle, 1, 3)
	small := gapb62Creature(g, me.ID, "Small", "Creature — Goblin", "", 2, 2)
	big := gapb62Creature(g, me.ID, "Big", "Creature — Ogre", "", 3, 3)
	pr7bAttack(t, g, opp.ID, small, big)
	if !hasString(effectiveAbilities(t, g, small), "skulk") {
		t.Errorf("the 2-power attacker has no skulk: %v", effectiveAbilities(t, g, small))
	}
	if hasString(effectiveAbilities(t, g, big), "skulk") {
		t.Error("the 3-power attacker got skulk")
	}
}

func TestGapb62MacCreadyDrainsForABigCreatureAttackingYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb62Creature(g, me.ID, "MacCready, Lamplight Mayor", "Legendary Creature — Human Advisor", gapb62MacCreadyOracle, 1, 3)
	big := gapb62Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", "", 4, 4)
	small := gapb62Creature(g, opp.ID, "Their Goblin", "Creature — Goblin", "", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	advanceToDeclareAttackersOf(t, g, 1)
	life, oppLife := me.Life, opp.Life
	for _, id := range []uuid.UUID{big, small} {
		if err := g.DeclareAttacker(id, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != life+2 {
		t.Errorf("life: attacker's controller %d→%d, mine %d→%d; want -2 and +2 for the 4-power creature only",
			oppLife, opp.Life, life, me.Life)
	}
}

// --- Gunner Conscript ---------------------------------------------------------------

func TestGapb62GunnerConscriptGrowsForEachAuraAndEquipmentAttachedToIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	conscript := gapb62Creature(g, me.ID, "Gunner Conscript", "Creature — Human Mercenary", gapb62GunnerConscriptOracl, 2, 2)
	if p, tt := ptOf(t, g, conscript); p != 2 || tt != 2 {
		t.Fatalf("bare Conscript is %d/%d", p, tt)
	}
	aura := gapb62Creature(g, me.ID, "Test Aura", "Enchantment — Aura", "", 0, 0)
	sword := gapb62Creature(g, me.ID, "Test Sword", "Artifact — Equipment", "", 0, 0)
	other := gapb62Creature(g, me.ID, "Other Sword", "Artifact — Equipment", "", 0, 0)
	g.WithWriteLock(func() {
		for _, a := range []uuid.UUID{aura, sword} {
			if err := g.AttachForEffect(a, game.TargetRef{Kind: game.TargetCard, ID: conscript}); err != nil {
				t.Fatalf("attach: %v", err)
			}
		}
	})
	if p, tt := ptOf(t, g, conscript); p != 4 || tt != 4 {
		t.Errorf("Conscript with an Aura and an Equipment is %d/%d, want 4/4", p, tt)
	}
	_ = other
}

func TestGapb62GunnerConscriptMakesJunkOnlyForWhatWasAttachedWhenItDied(t *testing.T) {
	for _, tc := range []struct {
		name        string
		aura, sword bool
		wantJunk    int
	}{
		{"bare", false, false, 0},
		{"enchanted", true, false, 1},
		{"equipped", false, true, 1},
		{"enchanted and equipped", true, true, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			conscript := gapb62Creature(g, me.ID, "Gunner Conscript", "Creature — Human Mercenary", gapb62GunnerConscriptOracl, 2, 2)
			var attach []uuid.UUID
			if tc.aura {
				attach = append(attach, gapb62Creature(g, me.ID, "Test Aura", "Enchantment — Aura", "", 0, 0))
			}
			if tc.sword {
				attach = append(attach, gapb62Creature(g, me.ID, "Test Sword", "Artifact — Equipment", "", 0, 0))
			}
			g.WithWriteLock(func() {
				for _, a := range attach {
					if err := g.AttachForEffect(a, game.TargetRef{Kind: game.TargetCard, ID: conscript}); err != nil {
						t.Fatalf("attach: %v", err)
					}
				}
				_ = g.DestroyPermanentForEffect(conscript)
			})
			passPriorityAroundTable(t, g)
			// Two triggers of one controller at once are ordered by that controller.
			for i := 0; i < 4; i++ {
				ch := triggerOrderPrompt(g)
				if ch == nil {
					break
				}
				if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				passPriorityAroundTable(t, g)
			}
			if n := len(gapb62Named(g, me.ID, "Junk")); n != tc.wantJunk {
				t.Errorf("%d Junk tokens, want %d", n, tc.wantJunk)
			}
		})
	}
}

// --- Mr. House, President and CEO ---------------------------------------------------

// gapb62RollKey finds an RNG key under which `dice` six-sided dice rolled
// for the source `id` by seat 0 total something `want` accepts, by peeking at
// the roll on a throwaway table.
func gapb62RollKey(t *testing.T, id uuid.UUID, dice int, want func(sum int) bool) byte {
	t.Helper()
	for n := 0; n < 256; n++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(n)))
		var kept []int
		g.WithWriteLock(func() {
			kept, _ = g.RollDiceIgnoringLowestForEffect(game.RandomDraw{Player: g.Seats[0].ID, Source: id}, 6, dice, 0)
		})
		sum := 0
		for _, k := range kept {
			sum += k
		}
		if len(kept) == dice && want(sum) {
			return byte(n)
		}
	}
	t.Fatal("no key in 0..255 gives the wanted roll")
	return 0
}

// gapb62House rolls Mr. House's ability under a key that gives the total
// `want` accepts and returns the table, the player and the card.
func gapb62House(t *testing.T, want func(sum int) bool, treasureMana int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	id := uuid.New()
	dice := 1 + treasureMana
	key := gapb62RollKey(t, id, dice, want)
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(randomModesKey(key))
	me := g.Seats[0]
	advanceToMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: id, Name: "Mr. House, President and CEO", TypeLine: "Legendary Artifact Creature — Human",
		OracleID: gapb62MrHouseOracle, Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
		for i := 0; i < 4-treasureMana; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "C", Source: uuid.New()})
		}
		for i := 0; i < treasureMana; i++ {
			me.ManaPool.AddMana(game.ManaToken{Color: "C", Source: uuid.New(), SourceKinds: game.ManaSourceTreasure})
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	return g, me, id
}

func TestGapb62MrHouseSixOrMoreMakesARobotAndATreasure(t *testing.T) {
	g, me, _ := gapb62House(t, func(sum int) bool { return sum == 6 }, 0)
	robots := gapb62Named(g, me.ID, "Robot")
	if len(robots) != 1 || len(gapb62Named(g, me.ID, "Treasure")) != 1 {
		t.Fatalf("a 6: %d Robots and %d Treasures, want 1 and 1", len(robots), len(gapb62Named(g, me.ID, "Treasure")))
	}
	if p, tt := ptOf(t, g, robots[0]); p != 3 || tt != 3 {
		t.Errorf("the Robot is %d/%d, want 3/3", p, tt)
	}
}

func TestGapb62MrHouseFourOrFiveMakesOnlyARobot(t *testing.T) {
	g, me, _ := gapb62House(t, func(sum int) bool { return sum == 4 || sum == 5 }, 0)
	if n := len(gapb62Named(g, me.ID, "Robot")); n != 1 {
		t.Errorf("a 4 or 5: %d Robots, want 1", n)
	}
	if n := len(gapb62Named(g, me.ID, "Treasure")); n != 0 {
		t.Errorf("a 4 or 5 also made %d Treasures", n)
	}
}

func TestGapb62MrHouseThreeOrLessMakesNothingAndPutsNothingOnTheStack(t *testing.T) {
	g, me, _ := gapb62House(t, func(sum int) bool { return sum <= 3 }, 0)
	if len(gapb62Named(g, me.ID, "Robot")) != 0 || len(gapb62Named(g, me.ID, "Treasure")) != 0 {
		t.Error("a roll of 3 or less made a token")
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && ev.Label != "" && len(ev.Label) > 8 && ev.Label[:8] == "Mr. Hous" {
			t.Errorf("a trigger was put on the stack for a roll under 4: %q", ev.Label)
		}
	}
}

func TestGapb62MrHouseRollsAnExtraDieForEachManaFromATreasure(t *testing.T) {
	g, me, _ := gapb62House(t, func(sum int) bool { return sum >= 2 }, 2)
	if n := len(rollEventsOf(g)); n != 3 {
		t.Errorf("%d dice rolled with two mana from Treasures, want 3", n)
	}
	// The three dice are one roll: one result, one trigger at most.
	if n := len(gapb62Named(g, me.ID, "Robot")); n > 1 {
		t.Errorf("%d Robots from one roll", n)
	}
}

// --- Raul, Trouble Shooter ----------------------------------------------------------

func TestGapb62RaulMillsEachPlayerAndOffersOnlyThisTurnsMilledCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	raul := gapb62Creature(g, me.ID, "Raul, Trouble Shooter", "Legendary Creature — Zombie Mutant Rogue", gapb62RaulOracle, 1, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	milledBear := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: milledBear, Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirTop := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: theirTop, Name: "Their Milled Card", TypeLine: "Creature — Bear", ManaCost: "{G}",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	discarded := gfDead(t, g, me, "Discarded Bear", "Creature — Bear", "{G}")

	if err := g.ActivateCatalogAbility(me.ID, raul, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(milledBear) || !opp.Graveyard.Contains(theirTop) {
		t.Fatal("each player did not mill a card")
	}
	// A card put into the graveyard any other way is not offered.
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, discarded, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a discarded card was cast from the graveyard")
	}
	// An opponent's milled card is not "your graveyard".
	floatMana(t, g, me, "{C}{G}")
	if err := g.CastSpell(me.ID, theirTop, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a card from an opponent's graveyard was cast")
	}
	if err := g.CastSpell(me.ID, milledBear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("the milled card could not be cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(milledBear) {
		t.Error("the milled Bear did not resolve onto the battlefield")
	}
}

func TestGapb62RaulAllowsOnlyOneCastEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gapb62Creature(g, me.ID, "Raul, Trouble Shooter", "Legendary Creature — Zombie Mutant Rogue", gapb62RaulOracle, 1, 4)
	var ids []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		ids = append(ids, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{G}",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 2) })
	for _, id := range ids {
		if !me.Graveyard.Contains(id) {
			t.Fatal("the cards were not milled")
		}
	}
	floatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("first cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, ids[1], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("a second spell was cast from among the milled cards in one turn")
	}
}

func TestGapb62RaulOffersNothingOnAnotherPlayersTurnOrALaterTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gapb62Creature(g, me.ID, "Raul, Trouble Shooter", "Legendary Creature — Zombie Mutant Rogue", gapb62RaulOracle, 1, 4)
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{G}",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	// Next turn of mine: the card was milled on an earlier turn.
	advanceToUpkeepOfSeat(t, g, 0)
	advanceTo(t, g, game.StepPrecombatMain)
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("a card milled on an earlier turn was cast")
	}
}
