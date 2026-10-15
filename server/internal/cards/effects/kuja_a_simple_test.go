package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_a_simple_test.go — Queen Brahne, Lindblum, Neheb and Deadpool
// of slice kuja-a (#3213).

// --- Queen Brahne --------------------------------------------------

func TestQueenBrahneAttackMakesAWizardAndHasProwess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brahne := pushDiesCreatureForTest(g, me.ID, "Queen Brahne", queenBrahneOracle, "Legendary Creature — Human Noble", 2, 1)
	if !hasEffectiveKeyword(t, g, brahne, "prowess") {
		t.Error("Queen Brahne should have prowess")
	}
	attackWith(t, g, opp.ID, brahne)
	passPriorityAroundTable(t, g)
	toks := wizardTokens(g, me.ID)
	if len(toks) != 1 {
		t.Fatalf("Wizard tokens after the attack = %d, want 1", len(toks))
	}
	if toks[0].Tapped {
		t.Error("the Wizard is created untapped, not attacking")
	}
	if toks[0].Power != 0 || toks[0].Toughness != 1 {
		t.Errorf("Wizard is %d/%d, want 0/1", toks[0].Power, toks[0].Toughness)
	}
}

func TestQueenBrahneMakesNoWizardWithoutAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Queen Brahne", queenBrahneOracle, "Legendary Creature — Human Noble", 2, 1)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := len(wizardTokens(g, me.ID)); got != 0 {
		t.Errorf("Wizard tokens = %d, want 0 without an attack", got)
	}
}

// --- Lindblum, Industrial Regency // Mage Siege ----------------------

func TestLindblumEntersTappedAndMageSiegeMakesAWizard(t *testing.T) {
	g, me, id := seatLandAdventure(t, lindblumOracleID, "Lindblum, Industrial Regency", "Mage Siege", "{2}{R}", []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Mage Siege: %v", err)
	}
	passPriorityAroundTable(t, g)
	toks := wizardTokens(g, me.ID)
	if len(toks) != 1 {
		t.Fatalf("Wizard tokens = %d, want 1", len(toks))
	}
	assertLandPlayableFromExile(t, g, me, id)
}

func TestLindblumIsAPlainTappedLandFromHand(t *testing.T) {
	g, me, id := seatLandAdventure(t, lindblumOracleID, "Lindblum, Industrial Regency", "Mage Siege", "{2}{R}", []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Lindblum: %v", err)
	}
	c := battlefieldCardFor(g, id)
	if c == nil || !c.Tapped {
		t.Fatalf("Lindblum = %+v, want a tapped land on the battlefield", c)
	}
}

// --- Neheb, the Eternal ---------------------------------------------

func TestNehebAfflictsThenAddsRedPerLifeLostThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neheb := pushDiesCreatureForTest(g, me.ID, "Neheb, the Eternal", nehebEternalOracle, "Legendary Creature — Zombie Minotaur Warrior", 4, 6)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 9)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(neheb, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, neheb); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g) // afflict resolves in the declare-blockers step
	if opp.Life != 40-3 {
		t.Fatalf("afflict 3: life = %d, want 37", opp.Life)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	red := 0
	for _, m := range me.ManaPool.Units() {
		if m.Color == "R" {
			red++
		}
	}
	if red != 3 {
		t.Errorf("mana pool = %+v, want {R}{R}{R} for the 3 life lost", me.ManaPool)
	}
}

func TestNehebAddsNothingWhenNoOpponentLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Neheb, the Eternal", nehebEternalOracle, "Legendary Creature — Zombie Minotaur Warrior", 4, 6)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("mana pool = %+v, want empty", me.ManaPool)
	}
}

func TestNehebCountsLifeLostByEveryOpponentButNotYours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Neheb, the Eternal", nehebEternalOracle, "Legendary Creature — Zombie Minotaur Warrior", 4, 6)
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, -2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[2].ID, -1)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5)
	})
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 3 {
		t.Errorf("mana pool = %+v, want 3 red (2 + 1, not your own 5)", me.ManaPool)
	}
}

// --- Deadpool, Trading Card ------------------------------------------

func TestDeadpoolCostsThreeLifeOnMyUpkeepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dp := pushDiesCreatureForTest(g, me.ID, "Deadpool, Trading Card", deadpoolOracle, "Legendary Creature — Mutant Mercenary Hero", 5, 3)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Life != 40 {
		t.Fatalf("life = %d after an opponent's upkeep, want 40", me.Life)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != 37 {
		t.Fatalf("life = %d after my upkeep, want 37", me.Life)
	}
	if !g.Battlefield.Contains(dp) {
		t.Fatal("Deadpool should still be on the battlefield")
	}
}

func TestDeadpoolSacrificeGivesEachOtherPlayerACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dp := pushDiesCreatureForTest(g, me.ID, "Deadpool, Trading Card", deadpoolOracle, "Legendary Creature — Mutant Mercenary Hero", 5, 3)
	for i := 0; i < 3; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	hands := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		hands[i] = p.Hand.Size()
	}
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, dp, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dp) {
		t.Error("Deadpool is sacrificed as a cost")
	}
	for i, p := range g.Seats {
		want := hands[i]
		if i != 0 {
			want++
		}
		if p.Hand.Size() != want {
			t.Errorf("seat %d hand = %d, want %d", i, p.Hand.Size(), want)
		}
	}
}
