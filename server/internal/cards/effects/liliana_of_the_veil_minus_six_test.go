package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana of the Veil −6 (#2154): the controller separates every
// permanent the target player controls into two piles, the target
// player picks the pile they sacrifice. Permanents of any type are
// separated, and the controller's own are not.
func TestLilianaMinusSixYouSeparateTheyChooseAndSacrificeThePile(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	lili := pushWalkerForTest(g, owner.ID, "Liliana of the Veil", lilianaOfTheVeilOracle, 6)
	a := b12Creature(g, them.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, them.ID, "B", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, them.ID, "Rock", "Artifact")
	mine := b12Creature(g, owner.ID, "Mine", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, lili, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	passPriorityAroundTable(t, g)

	split := latestRevealPickFor(g, owner.ID)
	if split == nil {
		t.Fatalf("Liliana's controller separates the piles: %+v", g.PendingChoices)
	}
	if len(split.ChooseCards) < 3 {
		t.Fatalf("the target's three permanents are separated, got %d", len(split.ChooseCards))
	}
	if err := g.ResolveRevealPick(split.ID, owner.ID, []uuid.UUID{a}); err != nil {
		t.Fatal(err)
	}
	pick := latestOptionPickFor(g, them.ID)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("the target player chooses a pile: %+v", g.PendingChoices)
	}
	// Piles are {A} and everything else; they sacrifice the larger one.
	answerOptionPick(t, g, them.ID, 1)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(a) {
		t.Error("A was in the pile the player did not choose")
	}
	if g.Battlefield.Contains(b) || g.Battlefield.Contains(rock) {
		t.Errorf("the chosen pile is sacrificed: b=%v rock=%v", g.Battlefield.Contains(b), g.Battlefield.Contains(rock))
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the controller's own permanents are not separated")
	}
}

// A target who controls nothing is asked nothing.
func TestLilianaMinusSixAgainstAnEmptyBoardAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	lili := pushWalkerForTest(g, owner.ID, "Liliana of the Veil", lilianaOfTheVeilOracle, 6)
	advanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(owner.ID, lili, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("−6: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestRevealPickFor(g, owner.ID) != nil || latestOptionPickFor(g, them.ID) != nil {
		t.Errorf("no piles to separate, but a prompt is open: %+v", g.PendingChoices)
	}
}
