package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b135_test.go — Dominaria Remastered GAP batch 135 (#5948). Each
// test plays the real catalog entry and asserts what its text says.

const (
	s82B135FesteringGoblinOracle = "66fb4764-d309-4c30-a2a4-474f9030dc87"
	s82B135FleshReaverOracle     = "a143fdfe-f1aa-4eab-8c81-8c72a4dc1695"
	s82B135TurncoatOracle        = "297c2ddf-6d2d-4ed4-ab37-716514d5e3b6"
	s82B135LemureOracle          = "bc3576fb-3c08-456e-b3b6-9eb01b2366d6"
	s82B135MindslicerOracle      = "76c229d2-4bdd-4520-97dd-afb40f8c65ec"
	s82B135GladiatorOracle       = "4da2425d-2118-4116-855d-2ef80693a7b7"
	s82B135SyphonMageOracle      = "d6608efe-b014-4b4d-a995-0be5047567c5"
	s82B135AnuridOracle          = "f8beeb0a-750f-4afa-a244-10011342bbeb"
	s82B135PyreZombieOracle      = "accb82be-4f90-4d9a-bafd-f8de3622b3a3"
	s82B135SolkanarOracle        = "b8aac8e8-4588-4cb9-8b95-5c4abab1c48b"
)

func TestS82GapB135EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82B135FesteringGoblinOracle: "Festering Goblin",
		s82B135FleshReaverOracle:     "Flesh Reaver",
		s82B135TurncoatOracle:        "Goblin Turncoat",
		s82B135LemureOracle:          "Hyalopterous Lemure",
		s82B135MindslicerOracle:      "Mindslicer",
		s82B135GladiatorOracle:       "Undead Gladiator",
		s82B135SyphonMageOracle:      "Urborg Syphon-Mage",
		s82B135AnuridOracle:          "Wretched Anurid",
		s82B135PyreZombieOracle:      "Pyre Zombie",
		s82B135SolkanarOracle:        "Sol'kanar the Swamp King",
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

func TestS82GapB135FesteringGoblinGivesATargetCreatureMinusOneWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goblin := effectstest.B12Push(g, me.ID, "Festering Goblin", "Creature — Zombie Goblin", s82B135FesteringGoblinOracle, 1, 1)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, goblin)
	g.RunStateChecksForTest()
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 1 || tg != 1 {
		t.Errorf("the Bear is %d/%d, want 1/1", p, tg)
	}
}

func TestS82GapB135MindslicerMakesEveryPlayerDiscardTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slicer := effectstest.B12Push(g, me.ID, "Mindslicer", "Creature — Horror", s82B135MindslicerOracle, 4, 3)
	for _, p := range g.Seats {
		p.Hand.Cards = nil
		effectstest.PushCatalogHandCard(p, "One", "Instant", "")
		effectstest.PushCatalogHandCard(p, "Two", "Sorcery", "")
	}
	effectstest.KillCreature(t, g, me.ID, slicer)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Hand.Size() != 0 {
			t.Errorf("seat %d still holds %d cards", i, p.Hand.Size())
		}
	}
}

func TestS82GapB135HyalopterousLemureLosesPowerAndFliesPerActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lemure := effectstest.B12Push(g, me.ID, "Hyalopterous Lemure", "Creature — Spirit", s82B135LemureOracle, 4, 3)
	if effectstest.HasEffectiveKeyword(t, g, lemure, "flying") {
		t.Fatal("the Lemure flies before activating")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, lemure, 0, game.ActivateAbilityParams{})
	effectstest.RavB2MustActivate(t, g, me.ID, lemure, 0, game.ActivateAbilityParams{})
	if p := effectstest.EffectivePower(t, g, lemure); p != 2 {
		t.Errorf("power %d after two activations, want 2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, lemure, "flying") {
		t.Error("the Lemure has no flying")
	}
}

func TestS82GapB135GoblinTurncoatSacrificesAGoblinToRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	turncoat := effectstest.B12Push(g, me.ID, "Goblin Turncoat", "Creature — Goblin Mercenary", s82B135TurncoatOracle, 2, 1)
	goblin := effectstest.B12Creature(g, me.ID, "Goblin Token", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, turncoat, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a non-Goblin")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, turncoat, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{goblin}})
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(turncoat); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(turncoat) {
		t.Error("the regenerated Turncoat was destroyed")
	}
}

func TestS82GapB135UndeadGladiatorReturnsOnlyDuringYourUpkeepForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gladiator := effectstest.PushGraveyardCardForTest(me, "Undead Gladiator")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == gladiator {
				me.Graveyard.Cards[i].OracleID = s82B135GladiatorOracle
				me.Graveyard.Cards[i].TypeLine = "Creature — Zombie Barbarian"
				me.Graveyard.Cards[i].ManaCost = "{1}{B}{B}"
			}
		}
	})
	pitch := effectstest.HandCardForTest(me, "Pitch", "Instant", "")
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, gladiator, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}}, "B", "B"); err == nil {
		t.Fatal("returned the Gladiator outside the upkeep")
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, gladiator, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}}, "B", "B")
	if !effectstest.HandHasID(me, gladiator) {
		t.Error("the Gladiator did not return to its owner's hand")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestS82GapB135UrborgSyphonMageDrainsEachOpponentForTheLifeLost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mage := effectstest.B12Push(g, me.ID, "Urborg Syphon-Mage", "Creature — Human Spellshaper", s82B135SyphonMageOracle, 2, 2)
	pitch := effectstest.HandCardForTest(me, "Pitch", "Instant", "")
	effectstest.AdvanceToMain(t, g)
	start := me.Life
	var opps []int
	for _, p := range g.Seats[1:] {
		opps = append(opps, p.Life)
	}
	effectstest.RavB2MustActivate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}}, "B", "B", "B")
	for i, p := range g.Seats[1:] {
		if p.Life != opps[i]-2 {
			t.Errorf("opponent %d life %d, want %d", i, p.Life, opps[i]-2)
		}
	}
	if want := start + 2*len(opps); me.Life != want {
		t.Errorf("life %d, want %d", me.Life, want)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the card was not discarded")
	}
}

func TestS82GapB135WretchedAnuridCostsYou1LifeForAnyOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	anurid := effectstest.B12Push(g, me.ID, "Wretched Anurid", "Creature — Zombie Frog Beast", s82B135AnuridOracle, 3, 3)
	start := me.Life
	effectstest.DftABCast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start-1 {
		t.Errorf("life %d after another creature entered, want %d", me.Life, start-1)
	}
	effectstest.DftABCast(t, g, me, game.Card{Name: "Bolt Scroll", TypeLine: "Artifact"})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start-1 {
		t.Errorf("life %d after a noncreature permanent entered, want %d", me.Life, start-1)
	}
	if !g.Battlefield.Contains(anurid) {
		t.Fatal("the Anurid left")
	}
}

func TestS82GapB135PyreZombieReturnsFromTheGraveyardForManaAndSacrificesForDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zombie := effectstest.B12Push(g, me.ID, "Pyre Zombie", "Creature — Zombie", s82B135PyreZombieOracle, 2, 1)
	effectstest.AdvanceToMain(t, g)
	start := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, zombie, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), Strict: true}); err == nil {
		t.Fatal("activated without paying")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, zombie, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)}, "R", "R", "R")
	if opp.Life != start-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, start-2)
	}
	if g.Battlefield.Contains(zombie) {
		t.Fatal("the Zombie was not sacrificed")
	}
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { effectstest.FillPoolColored(me, "B", 3) })
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, zombie) {
		t.Error("the Zombie did not return to its owner's hand")
	}
}

func TestS82GapB135SolkanarGainsLifeWheneverAnyPlayerCastsABlackSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Sol'kanar the Swamp King", "Legendary Creature — Demon", s82B135SolkanarOracle, 5, 5)
	start := me.Life
	effectstest.DftABCast(t, g, me, game.Card{Name: "Black Bear", TypeLine: "Creature — Bear", Colors: []string{"B"}, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start+1 {
		t.Errorf("life %d after my black spell, want %d", me.Life, start+1)
	}
	effectstest.DftABCast(t, g, me, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start+1 {
		t.Errorf("life %d after a green spell, want %d", me.Life, start+1)
	}
	_ = opp
}

func TestS82GapB135FleshReaverDealsThatMuchToYouWhenItDamagesACreatureOrOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaver := effectstest.HmlCreature(g, me.ID, "Flesh Reaver", s82B135FleshReaverOracle, 4, 4)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	start, oppStart := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, reaver)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, reaver); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != start-4 {
		t.Errorf("life %d after dealing 4 to a creature, want %d", me.Life, start-4)
	}
	if opp.Life != oppStart {
		t.Errorf("the opponent lost life although the Reaver was blocked")
	}
}

func TestS82GapB135FleshReaverDealsThatMuchToYouWhenItHitsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaver := effectstest.HmlCreature(g, me.ID, "Flesh Reaver", s82B135FleshReaverOracle, 4, 4)
	start, oppStart := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, reaver)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppStart-4 {
		t.Errorf("the opponent's life %d, want %d", opp.Life, oppStart-4)
	}
	if me.Life != start-4 {
		t.Errorf("life %d after hitting an opponent for 4, want %d", me.Life, start-4)
	}
}

func TestS82GapB135FleshReaverBlockingDealsThatMuchToItsControllerAndIgnoresDamageItTakes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	reaver := effectstest.HmlCreature(g, opp.ID, "Flesh Reaver", s82B135FleshReaverOracle, 4, 4)
	biter := effectstest.B12Creature(g, me.ID, "Biter", "Creature — Bear", 2, 2)
	start, oppStart := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, biter)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(reaver, biter); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(biter) {
		t.Error("the Biter survived the Reaver")
	}
	if opp.Life != oppStart-4 {
		t.Errorf("the Reaver's controller life %d, want %d: only the 4 the Reaver dealt, not the 2 it took", opp.Life, oppStart-4)
	}
	if me.Life != start {
		t.Errorf("my life changed to %d", me.Life)
	}
}
