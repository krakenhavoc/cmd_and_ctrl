package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// --- Gauntlet of Power ----------------------------------------------------------------

func TestS82GapB134GauntletPumpsEveryCreatureOfTheChosenColorAndDoublesBasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	redMine := effectstest.B31Push(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	greenMine := effectstest.B31Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	redTheirs := effectstest.B31Push(g, them.ID, "Raging Goblin", "Creature — Goblin Berserker", "", "{R}", 1, 1, "R")
	effectstest.PushChosenColorPermanent(t, g, me.ID, "Gauntlet of Power", "Artifact", s82B134GauntletOracle, "R")

	if got := effectstest.EffectivePower(t, g, redMine); got != 3 {
		t.Errorf("your red creature has power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, redTheirs); got != 2 {
		t.Errorf("an opponent's red creature has power %d, want 2: the anthem has no controller clause", got)
	}
	if got := effectstest.EffectivePower(t, g, greenMine); got != 2 {
		t.Errorf("a green creature has power %d, want 2", got)
	}

	mountain := effectstest.B31Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	effectstest.TapForMana(t, g, me.ID, mountain)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("your pool = %v, want two red", got)
	}
	// An opponent's basic land of the chosen colour doubles for its controller.
	theirMountain := effectstest.B31Push(g, them.ID, "Mountain", "Basic Land — Mountain", "", "", 0, 0)
	effectstest.TapForMana(t, g, them.ID, theirMountain)
	if got := effectstest.SortedPool(them); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("their pool = %v, want two red: the extra mana goes to the land's controller", got)
	}
	// A basic land of another colour makes no extra mana.
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.TapForMana(t, g, me.ID, forest)
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool after a Forest = %v, want just {G}", got)
	}
}

func TestS82GapB134GauntletIgnoresANonbasicLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushChosenColorPermanent(t, g, me.ID, "Gauntlet of Power", "Artifact", s82B134GauntletOracle, "R")
	taiga := effectstest.SeedPermanentWithOracle(g, me.ID, "Taiga", "Land — Mountain Forest", effectstest.CagedSunTaigaOracle)
	effectstest.TapForMana(t, g, me.ID, taiga)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil {
		t.Fatal("no mana pick for Taiga's {R} or {G}")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatalf("ResolveManaChoice(R): %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 1 {
		t.Errorf("a nonbasic land made %v, want one mana", got)
	}
}

// --- Helm of Awakening ----------------------------------------------------------------

func TestS82GapB134HelmOfAwakeningMakesEverySpellCostOneLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, opp.ID, "Helm of Awakening", s82B134HelmOracle, "Artifact")
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
	if err := effectstest.SocB19CastWithPool(g, me, card, "G", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for {1}{G} through an opponent's Helm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	second := game.Card{Name: "Elvish Mystic", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1}
	if err := effectstest.SocB19CastWithPool(g, me, second, "G", 0, game.CastSpellParams{}); err == nil {
		t.Error("a {G} spell cost nothing: the Helm may only reduce generic mana")
	}
}

// --- Icy Manipulator ------------------------------------------------------------------

func TestS82GapB134IcyManipulatorTapsAnArtifactCreatureOrLandButNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	icy := effectstest.PushCatalogPermanent(g, me.ID, "Icy Manipulator", "Artifact", s82B134IcyOracle, false)
	creature := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	rock := effectstest.B12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	effectstest.AdvanceToMain(t, g)

	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, icy, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ench), Strict: true}); err == nil {
		t.Fatal("tapped an enchantment")
	}
	for _, target := range []uuid.UUID{creature, land, rock} {
		effectstest.Untap(g, icy)
		effectstest.FillPool(me, 1)
		if err := g.ActivateCatalogAbility(me.ID, icy, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target), Strict: true}); err != nil {
			t.Fatalf("Icy Manipulator on %s: %v", target, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.TappedForTest(t, g, target) {
			t.Errorf("%s is not tapped", target)
		}
	}
}

// --- Jester's Cap ---------------------------------------------------------------------

func TestS82GapB134JestersCapExilesThreeCardsFromTheTargetsLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := effectstest.PushCatalogPermanent(g, me.ID, "Jester's Cap", "Artifact", s82B134JestersCapOracle, false)
	effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E")
	before := opp.Library.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, cap, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}); err != nil {
		t.Fatalf("Jester's Cap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.SearchChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not asked to search the library")
	}
	var ids []uuid.UUID
	for _, id := range pick.SearchCards {
		if len(ids) < 3 {
			ids = append(ids, id)
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids...)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lotus := effectstest.PushCatalogPermanent(g, me.ID, "Lotus Blossom", "Artifact", s82B134LotusBlossomOracle, false)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr8Counters(g, lotus, "petal"); got != 1 {
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
	if got := effectstest.SortedPool(me); len(got) != 3 || got[0] != "U" || got[2] != "U" {
		t.Errorf("pool = %v, want three blue", got)
	}
	if g.Battlefield.Contains(lotus) {
		t.Error("the Blossom was not sacrificed")
	}
}

func TestS82GapB134LotusBlossomDeclinedAddsNoPetal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lotus := effectstest.PushCatalogPermanent(g, me.ID, "Lotus Blossom", "Artifact", s82B134LotusBlossomOracle, false)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr8Counters(g, lotus, "petal"); got != 0 {
		t.Errorf("%d petal counters after declining, want 0", got)
	}
}

// --- Umbilicus ------------------------------------------------------------------------

func TestS82GapB134UmbilicusMakesTheUpkeepPlayerPayLifeOrBounce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Umbilicus", "Artifact", s82B134UmbilicusOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	start := opp.Life
	prompt := effectstest.PendingOfKind(g, game.PendingChoicePayUnless)
	if prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("the upkeep player was not asked: %+v", prompt)
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start-2 {
		t.Errorf("life %d after paying, want %d", opp.Life, start-2)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the permanent was returned although the life was paid")
	}

	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	third := g.Seats[2]
	mine := effectstest.B12Creature(g, third.ID, "Ox", "Creature — Ox", 2, 2)
	start = third.Life
	effectstest.AnswerPayUnless(t, g, third.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	own := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, third.ID)
	if own == nil {
		t.Fatal("the declining player was not asked which permanent to return")
	}
	if err := g.ResolveOwnPermanents(own.ID, third.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if third.Life != start {
		t.Errorf("life %d after declining, want %d", third.Life, start)
	}
	if g.Battlefield.Contains(mine) || !effectstest.HandHasID(third, mine) {
		t.Error("the declining player did not return their permanent to their hand")
	}
}

// --- Thran Golem ----------------------------------------------------------------------

func TestS82GapB134ThranGolemGrowsOnlyWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golem := effectstest.B12Push(g, me.ID, "Thran Golem", "Artifact Creature — Golem", s82B134ThranGolemOracle, 3, 3)
	aura := effectstest.B12Push(g, me.ID, "Pacifism", "Enchantment — Aura", "", 0, 0)
	if p, tg := effectstest.EffectivePower(t, g, golem), effectstest.EffectiveToughness(t, g, golem); p != 3 || tg != 3 {
		t.Fatalf("unenchanted Golem is %d/%d, want 3/3", p, tg)
	}
	if effectstest.HasEffectiveKeyword(t, g, golem, "flying") {
		t.Fatal("an unenchanted Golem flies")
	}
	effectstest.BrAttach(t, g, aura, golem)
	if p, tg := effectstest.EffectivePower(t, g, golem), effectstest.EffectiveToughness(t, g, golem); p != 5 || tg != 5 {
		t.Errorf("enchanted Golem is %d/%d, want 5/5", p, tg)
	}
	for _, kw := range []string{"flying", "first strike", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, golem, kw) {
			t.Errorf("the enchanted Golem has no %s", kw)
		}
	}
}

// --- Wall of Junk ---------------------------------------------------------------------

func TestS82GapB134WallOfJunkReturnsToHandAtEndOfCombatAfterBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.HmlCreature(g, me.ID, "Attacker", "", 1, 5)
	wall := effectstest.HmlCreature(g, opp.ID, "Wall of Junk", s82B134WallOfJunkOracle, 0, 7)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Fatal("the Wall left before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) || !effectstest.HandHasID(opp, wall) {
		t.Error("the Wall was not returned to its owner's hand at end of combat")
	}
}

func TestS82GapB134WallOfJunkStaysWhenItDoesNotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.HmlCreature(g, me.ID, "Attacker", "", 1, 5)
	wall := effectstest.HmlCreature(g, opp.ID, "Wall of Junk", s82B134WallOfJunkOracle, 0, 7)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Error("the Wall returned although it never blocked")
	}
}

// --- Body Snatcher --------------------------------------------------------------------

func TestS82GapB134BodySnatcherStaysIfYouDiscardACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	pitch := effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	effectstest.HandCardForTest(me, "Bolt", "Instant", "")
	snatcher := effectstest.DftABCast(t, g, me, game.Card{Name: "Body Snatcher", TypeLine: "Creature — Phyrexian Minion",
		OracleID: s82B134BodySnatcherOracle, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(snatcher) {
		t.Error("the Snatcher was exiled although a creature card was discarded")
	}
	if me.Hand.Contains(pitch) {
		t.Error("the creature card was not discarded")
	}
}

func TestS82GapB134BodySnatcherIsExiledWithoutADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	effectstest.HandCardForTest(me, "Bolt", "Instant", "")
	snatcher := effectstest.DftABCast(t, g, me, game.Card{Name: "Body Snatcher", TypeLine: "Creature — Phyrexian Minion",
		OracleID: s82B134BodySnatcherOracle, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(snatcher) || !g.Exile.Contains(snatcher) {
		t.Error("the Snatcher was not exiled with no creature card to discard")
	}
}

func TestS82GapB134BodySnatcherDiesExilesItselfAndReanimates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	snatcher := effectstest.B12Push(g, me.ID, "Body Snatcher", "Creature — Phyrexian Minion", s82B134BodySnatcherOracle, 2, 2)
	dead := effectstest.PushGraveyardCardForTest(me, "Fallen Bear")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == dead {
				me.Graveyard.Cards[i].TypeLine = "Creature — Bear"
				me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 2, 2
			}
		}
	})
	effectstest.KillCreature(t, g, me.ID, snatcher)
	g.RunStateChecksForTest()
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(snatcher) {
		t.Error("the Snatcher was not exiled")
	}
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card was not returned to the battlefield")
	}
}

// --- Evil Eye of Orms-by-Gore ---------------------------------------------------------

func TestS82GapB134EvilEyeStopsNonEyesAttackingAndIsBlockedOnlyByWalls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eye := effectstest.HmlCreature(g, me.ID, "Evil Eye of Orms-by-Gore", s82B134EvilEyeOracle, 3, 6)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, eye).TypeLine = "Creature — Eye" })
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Fatal("a non-Eye creature attacked")
	}
	if err := g.DeclareAttacker(eye, opp.ID); err != nil {
		t.Fatalf("the Eye cannot attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(theirs, eye), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(wall, eye); err != nil {
		t.Fatalf("a Wall blocks the Eye: %v", err)
	}
}

func TestS82GapB134EvilEyeDoesNotRestrictTheOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.HmlCreature(g, opp.ID, "Evil Eye of Orms-by-Gore", s82B134EvilEyeOracle, 3, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("an opponent's Eye stopped my creature attacking: %v", err)
	}
}

// --- Cryptic Gateway ------------------------------------------------------------------

func TestS82GapB134CrypticGatewayPutsACreatureSharingATypeWithEachTappedOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gateway := effectstest.PushCatalogPermanent(g, me.ID, "Cryptic Gateway", "Artifact", s82B134CrypticGatewayOracle, false)
	elfWarrior := effectstest.B12Creature(g, me.ID, "Elf Warrior", "Creature — Elf Warrior", 2, 2)
	elfShaman := effectstest.B12Creature(g, me.ID, "Elf Shaman", "Creature — Elf Shaman", 1, 1)
	both := effectstest.HandCardForTest(me, "Elf Sharer", "Creature — Elf", "")
	onlyOne := effectstest.HandCardForTest(me, "Warrior Only", "Creature — Human Warrior", "")
	effectstest.AdvanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, gateway, 0, game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{elfWarrior}, Strict: true}); err == nil {
		t.Fatal("activated with one creature tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, gateway, 0, game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{elfWarrior, elfShaman}, Strict: true}); err != nil {
		t.Fatalf("Cryptic Gateway: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the controller was not offered a creature card")
	}
	for _, id := range pick.ChooseCards {
		if id == onlyOne {
			t.Error("a card sharing a type with only one tapped creature was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, both)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(both) {
		t.Error("the Elf did not enter the battlefield")
	}
}
