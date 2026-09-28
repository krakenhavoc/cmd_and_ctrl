package effects

import (
	"testing"

	"github.com/google/uuid"

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

// mbCapacity is game.BlockCapacity of a battlefield card, layers fresh.
func mbCapacity(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c := e2Card(t, g, id)
	return game.BlockCapacity(&c)
}

// mbBlocked is the attackers `id` blocks.
func mbBlocked(t *testing.T, g *game.Game, id uuid.UUID) []uuid.UUID {
	t.Helper()
	c := e2Card(t, g, id)
	return c.BlockedAttackers()
}

// mbAttack has `me` attack `opp` with n fresh 2/2s and moves to
// declare blockers.
func mbAttack(t *testing.T, g *game.Game, me, opp *game.Player, n int) []uuid.UUID {
	t.Helper()
	var atks []uuid.UUID
	for i := 0; i < n; i++ {
		atks = append(atks, pushVanillaCreature(g, me.ID, "Bear", 2, 2))
	}
	declareAttack(t, g, opp.ID, atks...)
	advanceTo(t, g, game.StepDeclareBlockers)
	return atks
}

// mbBlock declares `blocker` on every attacker in one declaration.
func mbBlock(g *game.Game, blocker uuid.UUID, atks ...uuid.UUID) error {
	decls := make([]game.BlockDeclaration, len(atks))
	for i, a := range atks {
		decls[i] = game.BlockDeclaration{Blocker: blocker, Attacker: a}
	}
	return g.DeclareBlockers(decls)
}

func TestPalaceGuardBlocksEveryAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := pushMonster(g, opp.ID, "Palace Guard", palaceGuardOracle, 1, 4)
	if got := mbCapacity(t, g, guard); got != 0 {
		t.Fatalf("capacity = %d, want 0 (any number)", got)
	}
	atks := mbAttack(t, g, me, opp, 4)
	if err := mbBlock(g, guard, atks...); err != nil {
		t.Fatalf("blocking all four: %v", err)
	}
	if got := mbBlocked(t, g, guard); len(got) != 4 {
		t.Fatalf("blocks %d attackers, want 4", len(got))
	}
}

func TestHighGroundLetsYourCreaturesBlockTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "High Ground", "Enchantment", highGroundOracle, false)
	bear := pushMonster(g, opp.ID, "Their Bear", "", 2, 4)
	mine := pushMonster(g, me.ID, "My Bear", "", 2, 2)
	if got := mbCapacity(t, g, bear); got != 2 {
		t.Errorf("your creature's capacity = %d, want 2", got)
	}
	if got := mbCapacity(t, g, mine); got != 1 {
		t.Errorf("an opponent's creature's capacity = %d, want 1", got)
	}
	atks := mbAttack(t, g, me, opp, 3)
	if err := mbBlock(g, bear, atks[:2]...); err != nil {
		t.Fatalf("blocking two: %v", err)
	}
	if err := mbBlock(g, bear, atks[2]); err == nil {
		t.Fatal("a third block was accepted")
	}
}

func TestTwoHighGroundsAddUp(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "High Ground", "Enchantment", highGroundOracle, false)
	pushCatalogPermanent(g, opp.ID, "High Ground", "Enchantment", highGroundOracle, false)
	bear := pushMonster(g, opp.ID, "Their Bear", "", 2, 4)
	if got := mbCapacity(t, g, bear); got != 3 {
		t.Errorf("capacity under two High Grounds = %d, want 3", got)
	}
}

// Guardian of the Gateless blocks two, its "whenever this blocks"
// triggers ONCE (CR 509.3a), and it resolves for +2/+2.
func TestGuardianOfTheGatelessPumpsOncePerCreatureItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := pushMonster(g, opp.ID, "Guardian of the Gateless", guardianOfTheGatelessOracle, 3, 3)
	if !hasString(effectiveAbilities(t, g, guardian), "flying") {
		t.Error("no flying")
	}
	atks := mbAttack(t, g, me, opp, 2)
	if err := mbBlock(g, guardian, atks...); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := triggersOnStackFrom(g, guardian); n != 1 {
		t.Fatalf("%d triggers from blocking two, want one (CR 509.3a)", n)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, guardian), effectiveToughness(t, g, guardian); p != 5 || tg != 5 {
		t.Errorf("Guardian is %d/%d after blocking two, want 5/5", p, tg)
	}
}

func TestTwoHeadedGiantOfForiysBlocksTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	giant := pushMonster(g, opp.ID, "Two-Headed Giant of Foriys", twoHeadedGiantOfForiysOracle, 4, 4)
	if got := mbCapacity(t, g, giant); got != 2 {
		t.Errorf("capacity = %d, want 2", got)
	}
	if !hasString(effectiveAbilities(t, g, giant), "trample") {
		t.Error("no trample")
	}
}

func TestWallOfGlareBlocksAnyNumber(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := pushMonster(g, opp.ID, "Wall of Glare", wallOfGlareOracle, 0, 5)
	if got := mbCapacity(t, g, wall); got != 0 {
		t.Errorf("capacity = %d, want any number", got)
	}
	if !hasString(effectiveAbilities(t, g, wall), "defender") {
		t.Error("no defender")
	}
	atks := mbAttack(t, g, me, opp, 3)
	if err := mbBlock(g, wall, atks...); err != nil {
		t.Fatal(err)
	}
	// Power 0: nothing to divide, so no prompt holds the table.
	advanceTo(t, g, game.StepCombatDamage)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceDamageAssignment {
			t.Fatal("a 0-power wall was asked to divide its damage")
		}
	}
}

func TestEchoCircletLetsTheEquippedCreatureBlockTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	circlet := pushCatalogPermanent(g, me.ID, "Echo Circlet", "Artifact — Equipment", echoCircletOracle, false)
	host := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if got := mbCapacity(t, g, host); got != 1 {
		t.Fatalf("unequipped capacity = %d, want 1", got)
	}
	advanceTo(t, g, game.StepPrecombatMain)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, circlet, host)
	if got := mbCapacity(t, g, host); got != 2 {
		t.Errorf("equipped capacity = %d, want 2", got)
	}
}

func TestWatcherInTheWebBlocksEight(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	spider := pushMonster(g, opp.ID, "Watcher in the Web", watcherInTheWebOracle, 2, 5)
	if got := mbCapacity(t, g, spider); got != 8 {
		t.Errorf("capacity = %d, want 8", got)
	}
	if !hasString(effectiveAbilities(t, g, spider), "reach") {
		t.Error("no reach")
	}
}

// Hundred-Handed One: one before it is monstrous, a hundred after.
func TestHundredHandedOneBlocksAHundredWhenMonstrous(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giant := pushMonster(g, me.ID, "Hundred-Handed One", hundredHandedOneOracle, 3, 5)
	if got := mbCapacity(t, g, giant); got != 1 {
		t.Fatalf("capacity before monstrosity = %d, want 1", got)
	}
	activateMonstrosity(t, g, me.ID, giant, 0)
	passPriorityAroundTable(t, g)
	if got := mbCapacity(t, g, giant); got != 100 {
		t.Errorf("capacity once monstrous = %d, want 100", got)
	}
}

// Savvy Hunter's "whenever this creature attacks or blocks" triggers
// once when High Ground lets it block two (CR 509.3a).
func TestSavvyHunterBlockingTwoMakesOneFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "High Ground", "Enchantment", highGroundOracle, false)
	hunter := pushMonster(g, opp.ID, "Savvy Hunter", b30SavvyHunterOracle, 3, 6)
	atks := mbAttack(t, g, me, opp, 2)
	if err := mbBlock(g, hunter, atks...); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if n := triggersOnStackFrom(g, hunter); n != 1 {
		t.Fatalf("%d Savvy Hunter triggers from blocking two, want one", n)
	}
}
