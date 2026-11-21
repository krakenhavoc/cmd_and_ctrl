package testsms

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// multi_block_cards_test.go — #1706's proof cards: each can block the
// number of attackers it prints, and nothing more. The engine behaviour
// (adding, refusing past capacity, the CR 510.1d division, menace,
// Lure, limits, snapshots) is pinned in game/multi_block_test.go.

const (
	palaceGuardOracle            = "5c92f375-ae6a-4be5-a499-ab87d6bdc49b"
	highGroundOracle             = "b59053d5-b5f2-44aa-a29c-7f3908d86577"
	guardianOfTheGatelessOracle  = "e133605f-9227-42c4-afe9-6613ab095433"
	twoHeadedGiantOfForiysOracle = "38aa31bd-7145-43b9-9409-463d9ad6cd69"
	wallOfGlareOracle            = "a59897cb-7ead-494a-a1d0-c9baf79bf18b"
	echoCircletOracle            = "17134950-fdb3-48d0-b541-419db673f4c9"
	watcherInTheWebOracle        = "232a4128-133e-446d-9f1c-dd3875094473"
)

func TestPalaceGuardBlocksEveryAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.PushMonster(g, opp.ID, "Palace Guard", palaceGuardOracle, 1, 4)
	if got := effectstest.MbCapacity(t, g, guard); got != 0 {
		t.Fatalf("capacity = %d, want 0 (any number)", got)
	}
	atks := effectstest.MbAttack(t, g, me, opp, 4)
	if err := effectstest.MbBlock(g, guard, atks...); err != nil {
		t.Fatalf("blocking all four: %v", err)
	}
	if got := effectstest.MbBlocked(t, g, guard); len(got) != 4 {
		t.Fatalf("blocks %d attackers, want 4", len(got))
	}
}

func TestHighGroundLetsYourCreaturesBlockTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "High Ground", "Enchantment", highGroundOracle, false)
	bear := effectstest.PushMonster(g, opp.ID, "Their Bear", "", 2, 4)
	mine := effectstest.PushMonster(g, me.ID, "My Bear", "", 2, 2)
	if got := effectstest.MbCapacity(t, g, bear); got != 2 {
		t.Errorf("your creature's capacity = %d, want 2", got)
	}
	if got := effectstest.MbCapacity(t, g, mine); got != 1 {
		t.Errorf("an opponent's creature's capacity = %d, want 1", got)
	}
	atks := effectstest.MbAttack(t, g, me, opp, 3)
	if err := effectstest.MbBlock(g, bear, atks[:2]...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if err := effectstest.MbBlock(g, bear, atks[2]); err == nil {
		t.Fatal("a third block was accepted")
	}
}

func TestTwoHighGroundsAddUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "High Ground", "Enchantment", highGroundOracle, false)
	effectstest.PushCatalogPermanent(g, opp.ID, "High Ground", "Enchantment", highGroundOracle, false)
	bear := effectstest.PushMonster(g, opp.ID, "Their Bear", "", 2, 4)
	if got := effectstest.MbCapacity(t, g, bear); got != 3 {
		t.Errorf("capacity under two High Grounds = %d, want 3", got)
	}
}

// Guardian of the Gateless blocks two, its "whenever this blocks"
// triggers ONCE (CR 509.3a), and it resolves for +2/+2.
func TestGuardianOfTheGatelessPumpsOncePerCreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := effectstest.PushMonster(g, opp.ID, "Guardian of the Gateless", guardianOfTheGatelessOracle, 3, 3)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, guardian), "flying") {
		t.Error("no flying")
	}
	atks := effectstest.MbAttack(t, g, me, opp, 2)
	if err := effectstest.MbBlock(g, guardian, atks...); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := effectstest.TriggersOnStackFrom(g, guardian); n != 1 {
		t.Fatalf("%d triggers from blocking two, want one (CR 509.3a)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, guardian), effectstest.EffectiveToughness(t, g, guardian); p != 5 || tg != 5 {
		t.Errorf("Guardian is %d/%d after blocking two, want 5/5", p, tg)
	}
}

func TestTwoHeadedGiantOfForiysBlocksTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	giant := effectstest.PushMonster(g, opp.ID, "Two-Headed Giant of Foriys", twoHeadedGiantOfForiysOracle, 4, 4)
	if got := effectstest.MbCapacity(t, g, giant); got != 2 {
		t.Errorf("capacity = %d, want 2", got)
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, giant), "trample") {
		t.Error("no trample")
	}
}

func TestWallOfGlareBlocksAnyNumber(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushMonster(g, opp.ID, "Wall of Glare", wallOfGlareOracle, 0, 5)
	if got := effectstest.MbCapacity(t, g, wall); got != 0 {
		t.Errorf("capacity = %d, want any number", got)
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, wall), "defender") {
		t.Error("no defender")
	}
	atks := effectstest.MbAttack(t, g, me, opp, 3)
	if err := effectstest.MbBlock(g, wall, atks...); err != nil {
		t.Fatal(err)
	}
	// Power 0: nothing to divide, so no prompt holds the table.
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDamageAssignment {
			t.Fatal("a 0-power wall was asked to divide its damage")
		}
	}
}

func TestEchoCircletLetsTheEquippedCreatureBlockTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	circlet := effectstest.PushCatalogPermanent(g, me.ID, "Echo Circlet", "Artifact — Equipment", echoCircletOracle, false)
	host := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if got := effectstest.MbCapacity(t, g, host); got != 1 {
		t.Fatalf("unequipped capacity = %d, want 1", got)
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.EquipTo(t, g, me.ID, circlet, host)
	if got := effectstest.MbCapacity(t, g, host); got != 2 {
		t.Errorf("equipped capacity = %d, want 2", got)
	}
}

func TestWatcherInTheWebBlocksEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	spider := effectstest.PushMonster(g, opp.ID, "Watcher in the Web", watcherInTheWebOracle, 2, 5)
	if got := effectstest.MbCapacity(t, g, spider); got != 8 {
		t.Errorf("capacity = %d, want 8", got)
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, spider), "reach") {
		t.Error("no reach")
	}
}

// Hundred-Handed One: one before it is monstrous, a hundred after.
func TestHundredHandedOneBlocksAHundredWhenMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	giant := effectstest.PushMonster(g, me.ID, "Hundred-Handed One", effectstest.HundredHandedOneOracle, 3, 5)
	if got := effectstest.MbCapacity(t, g, giant); got != 1 {
		t.Fatalf("capacity before monstrosity = %d, want 1", got)
	}
	effectstest.ActivateMonstrosity(t, g, me.ID, giant, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.MbCapacity(t, g, giant); got != 100 {
		t.Errorf("capacity once monstrous = %d, want 100", got)
	}
}

// Savvy Hunter's "whenever this creature attacks or blocks" triggers
// once when High Ground lets it block two (CR 509.3a).
func TestSavvyHunterBlockingTwoMakesOneFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "High Ground", "Enchantment", highGroundOracle, false)
	hunter := effectstest.PushMonster(g, opp.ID, "Savvy Hunter", effectstest.B30SavvyHunterOracle, 3, 6)
	atks := effectstest.MbAttack(t, g, me, opp, 2)
	if err := effectstest.MbBlock(g, hunter, atks...); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := effectstest.TriggersOnStackFrom(g, hunter); n != 1 {
		t.Fatalf("%d Savvy Hunter triggers from blocking two, want one", n)
	}
}
