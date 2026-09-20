package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prismari_command_test.go — #1112. Prismari Command is a "choose
// two" with per-mode targets, Kolaghan's Command's shape (#764):
// each chosen bullet gets its own target group and is re-checked
// against ITS OWN clause at resolution (CR 608.2b).

const prismariCommandOracle = "fa3e28b1-131c-4223-81e0-18dfbab22c26"

// --- two different targeted bullets, damage + destroy -------------

func TestPrismariCommandBurnsAndDestroysDifferentTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	before := opp.Life

	castModal(t, g, "Prismari Command", "Instant", prismariCommandOracle,
		[]int{0, 3},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, opp.ID, 0, 0), // occurrence 0 = "deals 2 damage to any target"
			modeRef(game.TargetCard, rock, 1, 0),     // occurrence 1 = "destroy target artifact"
		})
	passPriorityAroundTable(t, g)

	if opp.Life != before-2 {
		t.Errorf("2 damage to any target (the player): %d → %d, want %d", before, opp.Life, before-2)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the destroy bullet should have destroyed the artifact")
	}
	_ = me
}

// A creature named in the "target player" or "target artifact" slot
// is illegal — each bullet's clause is checked on its own.
func TestPrismariCommandRefusesATargetForTheWrongMode(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)

	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Prismari Command", TypeLine: "Instant",
		OracleID: prismariCommandOracle, Owner: active.ID, Controller: active.ID,
	})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Modes: []int{2, 3}, // "target player creates a Treasure" / "destroy target artifact"
		Targets: []game.TargetRef{
			modeRef(game.TargetCard, bear, 0, 0),
			modeRef(game.TargetCard, bear, 1, 0),
		},
	})
	if err != game.ErrIllegalTarget {
		t.Errorf("a creature in the player/artifact slots: err %v, want ErrIllegalTarget", err)
	}
}

// --- draw two, then discard two ------------------------------------

func TestPrismariCommandDrawsTwoThenDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	before := opp.Hand.Size()

	castModal(t, g, "Prismari Command", "Instant", prismariCommandOracle,
		[]int{1, 3},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, opp.ID, 0, 0), // occurrence 0 = the draw-then-discard bullet
			modeRef(game.TargetCard, rock, 1, 0),     // occurrence 1 = "destroy target artifact"
		})
	passPriorityAroundTable(t, g)

	if got := opp.Hand.Size(); got != before+2 {
		t.Fatalf("draws two first: hand %d, want %d", got, before+2)
	}
	c := discardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("then asks the target player to discard two")
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("bounds [%d,%d], want [2,2]", c.ChooseMin, c.ChooseMax)
	}
	discardFromHand(t, g, opp.ID)
	if got := opp.Hand.Size(); got != before {
		t.Errorf("hand %d, want back to %d", got, before)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the destroy bullet should still have resolved")
	}
	_ = me
}

// --- creates a Treasure for the target player ----------------------

func TestPrismariCommandCreatesATreasureForTheTargetPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)

	castModal(t, g, "Prismari Command", "Instant", prismariCommandOracle,
		[]int{0, 2},
		[]game.TargetRef{
			modeRef(game.TargetCard, bear, 0, 0),     // occurrence 0 = "deals 2 damage to any target"
			modeRef(game.TargetPlayer, opp.ID, 1, 0), // occurrence 1 = "target player creates a Treasure"
		})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Error("2 damage should not kill a 3/3")
	}
	if opp.Life != before {
		t.Error("the damage bullet targeted the creature, not the player — opponent life should be unchanged")
	}
	treasures := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && c.Name == "Treasure" {
			treasures++
		}
	}
	if treasures != 1 {
		t.Errorf("the target player should control one Treasure token, found %d", treasures)
	}
	_ = me
}
