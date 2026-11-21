package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b134_test.go — Dominaria Remastered GAP batch 134 (#5948). Each
// test plays the real catalog entry and asserts what its text says.

const (
	s82B134CrypticGatewayOracle = "abdc1c92-1434-47d3-abaa-debe047903b9"
	s82B134GauntletOracle       = "1182591d-8660-4a2e-8d77-2fc544b02b66"
	s82B134HelmOracle           = "f2fcef65-4873-4b1f-a257-b8648097c990"
	s82B134IcyOracle            = "3608f1f7-8dc5-4dd1-ae91-c830e1de9529"
	s82B134JestersCapOracle     = "8783a94a-d0b2-4291-a550-dbcf8e16bf4b"
	s82B134LotusBlossomOracle   = "32e66ea0-5f35-4df8-8d00-580403c51bb5"
	s82B134UmbilicusOracle      = "c9132e72-d016-4cb7-8662-a16b7d932bef"
	s82B134ThranGolemOracle     = "fe55f366-2996-4a4f-9b30-fc494cc80f87"
	s82B134WallOfJunkOracle     = "22b0de2d-47a8-4269-ba21-4bc2b4b10b40"
	s82B134BodySnatcherOracle   = "e04688e0-672b-407f-a2c9-fd7292926483"
	s82B134EvilEyeOracle        = "3f79780a-accd-4782-94a8-a6e71fb3ada7"
)

func TestS82GapB134EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82B134CrypticGatewayOracle: "Cryptic Gateway",
		s82B134GauntletOracle:       "Gauntlet of Power",
		s82B134HelmOracle:           "Helm of Awakening",
		s82B134IcyOracle:            "Icy Manipulator",
		s82B134JestersCapOracle:     "Jester's Cap",
		s82B134LotusBlossomOracle:   "Lotus Blossom",
		s82B134UmbilicusOracle:      "Umbilicus",
		s82B134ThranGolemOracle:     "Thran Golem",
		s82B134WallOfJunkOracle:     "Wall of Junk",
		s82B134BodySnatcherOracle:   "Body Snatcher",
		s82B134EvilEyeOracle:        "Evil Eye of Orms-by-Gore",
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

// --- Gauntlet of Power ----------------------------------------------------------------

func TestS82GapB134GauntletPumpsEveryCreatureOfTheChosenColorAndDoublesBasicLands(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	redMine := b31Push(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	greenMine := b31Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	redTheirs := b31Push(g, them.ID, "Raging Goblin", "Creature — Goblin Berserker", "", "{R}", 1, 1, "R")
	pushChosenColorPermanent(t, g, me.ID, "Gauntlet of Power", "Artifact", s82B134GauntletOracle, "R")

	if got := effectivePower(t, g, redMine); got != 3 {
		t.Errorf("your red creature has power %d, want 3", got)
	}
	if got := effectivePower(t, g, redTheirs); got != 2 {
		t.Errorf("an opponent's red creature has power %d, want 2: the anthem has no controller clause", got)
	}
	if got := effectivePower(t, g, greenMine); got != 2 {
		t.Errorf("a green creature has power %d, want 2", got)
	}

	mountain := b31Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	tapForMana(t, g, me.ID, mountain)
	if got := sortedPool(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("your pool = %v, want two red", got)
	}
	// An opponent's basic land of the chosen colour doubles for its controller.
	theirMountain := b31Push(g, them.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	tapForMana(t, g, them.ID, theirMountain)
	if got := sortedPool(them); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("their pool = %v, want two red: the extra mana goes to the land's controller", got)
	}
	// A basic land of another colour makes no extra mana.
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	tapForMana(t, g, me.ID, forest)
	if got := sortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool after a Forest = %v, want just {G}", got)
	}
}

func TestS82GapB134GauntletIgnoresANonbasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushChosenColorPermanent(t, g, me.ID, "Gauntlet of Power", "Artifact", s82B134GauntletOracle, "R")
	taiga := seedPermanentWithOracle(g, me.ID, "Taiga", "Land — Mountain Forest", cagedSunTaigaOracle)
	tapForMana(t, g, me.ID, taiga)
	pick := pendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for Taiga's {R} or {G}")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice(R): %v", err)
	}
	if got := sortedPool(me); len(got) != 1 {
		t.Errorf("a nonbasic land made %v, want one mana", got)
	}
}

// --- Helm of Awakening ----------------------------------------------------------------

func TestS82GapB134HelmOfAwakeningMakesEverySpellCostOneLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, opp.ID, "Helm of Awakening", s82B134HelmOracle, "Artifact")
	advanceToMain(t, g)
	card := game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
	if err := socB19CastWithPool(g, me, card, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for {1}{G} through an opponent's Helm: %v", err)
	}
	passPriorityAroundTable(t, g)
	second := game.Card{Name: "Elvish Mystic", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1}
	if err := socB19CastWithPool(g, me, second, "G", 0, game.CastSpellParams{}); err == nil {
		t.Error("a {G} spell cost nothing: the Helm may only reduce generic mana")
	}
}

// --- Icy Manipulator ------------------------------------------------------------------

func TestS82GapB134IcyManipulatorTapsAnArtifactCreatureOrLandButNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	icy := pushCatalogPermanent(g, me.ID, "Icy Manipulator", "Artifact", s82B134IcyOracle, false)
	creature := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	rock := b12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	ench := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	advanceToMain(t, g)

	fillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, icy, 0, game.ActivateAbilityParams{Targets: cardRefs(ench), Strict: true}); err == nil {
		t.Fatal("tapped an enchantment")
	}
	for _, target := range []uuid.UUID{creature, land, rock} {
		untap(g, icy)
		fillPool(me, 1)
		if err := g.ActivateCatalogAbility(me.ID, icy, 0, game.ActivateAbilityParams{Targets: cardRefs(target), Strict: true}); err != nil {
			t.Fatalf("Icy Manipulator on %s: %v", target, err)
		}
		passPriorityAroundTable(t, g)
		if !dftQRUZTapped(t, g, target) {
			t.Errorf("%s is not tapped", target)
		}
	}
}

// --- Jester's Cap ---------------------------------------------------------------------

func TestS82GapB134JestersCapExilesThreeCardsFromTheTargetsLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := pushCatalogPermanent(g, me.ID, "Jester's Cap", "Artifact", s82B134JestersCapOracle, false)
	seedLibrary(opp, "A", "B", "C", "D", "E")
	before := opp.Library.Size()
	advanceToMain(t, g)
	fillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, cap, 0, game.ActivateAbilityParams{Targets: playerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("Jester's Cap: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := searchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not asked to search the library")
	}
	var ids []uuid.UUID
	for _, id := range pick.SearchCards {
		if len(ids) < 3 {
			ids = append(ids, id)
		}
	}
	answerSearchByID(t, g, me.ID, ids...)
	passPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != before-3 {
		t.Errorf("the library has %d cards, want %d", got, before-3)
	}
	for _, id := range ids {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if g.Battlefield.Contains(cap) {
		t.Error("the Cap was not sacrificed")
	}
}

// --- Lotus Blossom --------------------------------------------------------------------

func TestS82GapB134LotusBlossomGrowsOnYourUpkeepAndPaysOutItsPetals(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lotus := pushCatalogPermanent(g, me.ID, "Lotus Blossom", "Artifact", s82B134LotusBlossomOracle, false)
	walkToUpkeepOfSeat0(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := pr8Counters(g, lotus, "petal"); got != 1 {
		t.Fatalf("%d petal counters after accepting, want 1", got)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == lotus {
				g.Battlefield.Cards[i].Counters["petal"] = 3
			}
		}
	})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, lotus, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("Lotus Blossom: %v", err)
	}
	if got := sortedPool(me); len(got) != 3 || got[0] != "U" || got[2] != "U" {
		t.Errorf("pool = %v, want three blue", got)
	}
	if g.Battlefield.Contains(lotus) {
		t.Error("the Blossom was not sacrificed")
	}
}

func TestS82GapB134LotusBlossomDeclinedAddsNoPetal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lotus := pushCatalogPermanent(g, me.ID, "Lotus Blossom", "Artifact", s82B134LotusBlossomOracle, false)
	walkToUpkeepOfSeat0(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := pr8Counters(g, lotus, "petal"); got != 0 {
		t.Errorf("%d petal counters after declining, want 0", got)
	}
}

// --- Umbilicus ------------------------------------------------------------------------

func TestS82GapB134UmbilicusMakesTheUpkeepPlayerPayLifeOrBounce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Umbilicus", "Artifact", s82B134UmbilicusOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	start := opp.Life
	prompt := pendingOfKind(g, game.PendingChoicePayUnless)
	if prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("the upkeep player was not asked: %+v", prompt)
	}
	answerPayUnless(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != start-2 {
		t.Errorf("life %d after paying, want %d", opp.Life, start-2)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the permanent was returned although the life was paid")
	}

	advanceToUpkeepOf(t, g, 2)
	passPriorityAroundTable(t, g)
	third := g.Seats[2]
	mine := b12Creature(g, third.ID, "Ox", "Creature — Ox", 2, 2)
	start = third.Life
	answerPayUnless(t, g, third.ID, false)
	passPriorityAroundTable(t, g)
	own := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, third.ID)
	if own == nil {
		t.Fatal("the declining player was not asked which permanent to return")
	}
	if err := g.ResolveOwnPermanents(own.ID, third.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if third.Life != start {
		t.Errorf("life %d after declining, want %d", third.Life, start)
	}
	if g.Battlefield.Contains(mine) || !handHasID(third, mine) {
		t.Error("the declining player did not return their permanent to their hand")
	}
}

// --- Thran Golem ----------------------------------------------------------------------

func TestS82GapB134ThranGolemGrowsOnlyWhileEnchanted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	golem := b12Push(g, me.ID, "Thran Golem", "Artifact Creature — Golem", s82B134ThranGolemOracle, 3, 3)
	aura := b12Push(g, me.ID, "Pacifism", "Enchantment — Aura", "", 0, 0)
	if p, tg := effectivePower(t, g, golem), effectiveToughness(t, g, golem); p != 3 || tg != 3 {
		t.Fatalf("unenchanted Golem is %d/%d, want 3/3", p, tg)
	}
	if hasEffectiveKeyword(t, g, golem, "flying") {
		t.Fatal("an unenchanted Golem flies")
	}
	brAttach(t, g, aura, golem)
	if p, tg := effectivePower(t, g, golem), effectiveToughness(t, g, golem); p != 5 || tg != 5 {
		t.Errorf("enchanted Golem is %d/%d, want 5/5", p, tg)
	}
	for _, kw := range []string{"flying", "first strike", "trample"} {
		if !hasEffectiveKeyword(t, g, golem, kw) {
			t.Errorf("the enchanted Golem has no %s", kw)
		}
	}
}

// --- Wall of Junk ---------------------------------------------------------------------

func TestS82GapB134WallOfJunkReturnsToHandAtEndOfCombatAfterBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := hmlCreature(g, me.ID, "Attacker", "", 1, 5)
	wall := hmlCreature(g, opp.ID, "Wall of Junk", s82B134WallOfJunkOracle, 0, 7)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Fatal("the Wall left before end of combat")
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) || !handHasID(opp, wall) {
		t.Error("the Wall was not returned to its owner's hand at end of combat")
	}
}

func TestS82GapB134WallOfJunkStaysWhenItDoesNotBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := hmlCreature(g, me.ID, "Attacker", "", 1, 5)
	wall := hmlCreature(g, opp.ID, "Wall of Junk", s82B134WallOfJunkOracle, 0, 7)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Error("the Wall returned although it never blocked")
	}
}

// --- Body Snatcher --------------------------------------------------------------------

func TestS82GapB134BodySnatcherStaysIfYouDiscardACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	pitch := handCardForTest(me, "Bear", "Creature — Bear", "")
	handCardForTest(me, "Bolt", "Instant", "")
	snatcher := dftABCast(t, g, me, game.Card{Name: "Body Snatcher", TypeLine: "Creature — Phyrexian Minion",
		OracleID: s82B134BodySnatcherOracle, Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(snatcher) {
		t.Error("the Snatcher was exiled although a creature card was discarded")
	}
	if me.Hand.Contains(pitch) {
		t.Error("the creature card was not discarded")
	}
}

func TestS82GapB134BodySnatcherIsExiledWithoutADiscard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	handCardForTest(me, "Bolt", "Instant", "")
	snatcher := dftABCast(t, g, me, game.Card{Name: "Body Snatcher", TypeLine: "Creature — Phyrexian Minion",
		OracleID: s82B134BodySnatcherOracle, Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snatcher) || !g.Exile.Contains(snatcher) {
		t.Error("the Snatcher was not exiled with no creature card to discard")
	}
}

func TestS82GapB134BodySnatcherDiesExilesItselfAndReanimates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	snatcher := b12Push(g, me.ID, "Body Snatcher", "Creature — Phyrexian Minion", s82B134BodySnatcherOracle, 2, 2)
	dead := pushGraveyardCardForTest(me, "Fallen Bear")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == dead {
				me.Graveyard.Cards[i].TypeLine = "Creature — Bear"
				me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 2, 2
			}
		}
	})
	killCreature(t, g, me.ID, snatcher)
	g.RunStateChecksForTest()
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(snatcher) {
		t.Error("the Snatcher was not exiled")
	}
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card was not returned to the battlefield")
	}
}

// --- Evil Eye of Orms-by-Gore ---------------------------------------------------------

func TestS82GapB134EvilEyeStopsNonEyesAttackingAndIsBlockedOnlyByWalls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eye := hmlCreature(g, me.ID, "Evil Eye of Orms-by-Gore", s82B134EvilEyeOracle, 3, 6)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, eye).TypeLine = "Creature — Eye" })
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Fatal("a non-Eye creature attacked")
	}
	if err := g.DeclareAttacker(eye, opp.ID); err != nil {
		t.Fatalf("the Eye cannot attack: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	brRefusal(t, g.DeclareBlocker(theirs, eye), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(wall, eye); err != nil {
		t.Fatalf("a Wall blocks the Eye: %v", err)
	}
}

func TestS82GapB134EvilEyeDoesNotRestrictTheOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hmlCreature(g, opp.ID, "Evil Eye of Orms-by-Gore", s82B134EvilEyeOracle, 3, 6)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("an opponent's Eye stopped my creature attacking: %v", err)
	}
}

// --- Cryptic Gateway ------------------------------------------------------------------

func TestS82GapB134CrypticGatewayPutsACreatureSharingATypeWithEachTappedOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gateway := pushCatalogPermanent(g, me.ID, "Cryptic Gateway", "Artifact", s82B134CrypticGatewayOracle, false)
	elfWarrior := b12Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior", 2, 2)
	elfShaman := b12Creature(g, me.ID, "Elf Shaman", "Creature — Elf Shaman", 1, 1)
	both := handCardForTest(me, "Elf Sharer", "Creature — Elf", "")
	onlyOne := handCardForTest(me, "Warrior Only", "Creature — Human Warrior", "")
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, gateway, 0, game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{elfWarrior}, Strict: true}); err == nil {
		t.Fatal("activated with one creature tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, gateway, 0, game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{elfWarrior, elfShaman}, Strict: true}); err != nil {
		t.Fatalf("Cryptic Gateway: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not offered a creature card")
	}
	for _, id := range pick.ChooseCards {
		if id == onlyOne {
			t.Error("a card sharing a type with only one tapped creature was offered")
		}
	}
	answerChooseCards(t, g, me.ID, both)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(both) {
		t.Error("the Elf did not enter the battlefield")
	}
}
