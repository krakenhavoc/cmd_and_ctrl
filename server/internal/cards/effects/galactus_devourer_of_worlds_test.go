package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// galactus_devourer_of_worlds_test.go — #2744: a creature that must
// attack an opponent with the most life among its controller's
// opponents (CR 508.1d).

const galactusOracle = "4232995b-68c0-4514-8cde-bc62b9d1cbaa"

// galactusAtDeclareAttackers puts Galactus under seat 0, walks to seat
// 0's NEXT declare-attackers step (so it is not summoning sick) and sets
// the three opponents' life totals.
func galactusAtDeclareAttackers(t *testing.T, lives ...int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	galactus := pushDiesCreatureForTest(g, me.ID, "Galactus, Devourer of Worlds", galactusOracle,
		"Legendary Creature — Elder Alien", 12, 12)
	advanceToUpkeepOf(t, g, 1)
	advanceToDeclareAttackersOf(t, g, 0)
	for i, life := range lives {
		g.Seats[i+1].Life = life
	}
	return g, galactus
}

// TestGalactusMustAttackTheOpponentWithTheMostLife — with Galactus
// home, the pass is refused and the sentence says why; an attack on an
// opponent with less life is refused too; an attack on the opponent
// with the most life is accepted, and so is the pass after it.
func TestGalactusMustAttackTheOpponentWithTheMostLife(t *testing.T) {
	g, galactus := galactusAtDeclareAttackers(t, 40, 30, 20)
	me, most, less := g.Seats[0], g.Seats[1], g.Seats[2]

	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != galactus {
		t.Fatalf("pass with Galactus home = %v, want a refusal naming it", err)
	}
	if got, want := re.Sentence(me.ID), "Galactus, Devourer of Worlds must attack an opponent with the most life if able."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	var must map[uuid.UUID][]uuid.UUID
	g.ReadSnapshot(func() { must = g.MustAttackForEffect() })
	if got := must[galactus]; len(got) != 1 || got[0] != most.ID {
		t.Errorf("must attack %v, want only the opponent with the most life %v", got, most.ID)
	}
	if err := g.DeclareAttacker(galactus, less.ID); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("attacking an opponent with less life = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(galactus, most.ID); err != nil {
		t.Fatalf("attacking the opponent with the most life: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("pass with Galactus attacking the most-life opponent: %v", err)
	}
}

// TestGalactusMayAttackEitherOpponentTiedForTheMostLife — a tie offers
// each of the tied opponents.
func TestGalactusMayAttackEitherOpponentTiedForTheMostLife(t *testing.T) {
	g, galactus := galactusAtDeclareAttackers(t, 40, 40, 20)
	if err := g.DeclareAttacker(galactus, g.Seats[2].ID); err != nil {
		t.Fatalf("attacking the second opponent tied for the most life: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("pass: %v", err)
	}
}

// TestGalactusNeedNotAttackWhileYouControlSilverSurfer — the "unless"
// clause switches the requirement off.
func TestGalactusNeedNotAttackWhileYouControlSilverSurfer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Galactus, Devourer of Worlds", galactusOracle,
		"Legendary Creature — Elder Alien", 12, 12)
	surfer := pushVanillaCreature(g, me.ID, "Silver Surfer, Galactus's Herald", 6, 6)
	advanceToUpkeepOf(t, g, 1)
	advanceToDeclareAttackersOf(t, g, 0)
	g.Seats[1].Life = 40
	unmet := func() bool {
		var e *game.AttackRequirementError
		g.ReadSnapshot(func() { e = g.AttackRequirementsUnmetForEffect() })
		return e != nil
	}
	if unmet() {
		t.Fatal("Galactus owes an attack beside Silver Surfer")
	}
	// Without the Surfer, the same board owes the attack.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(surfer); err != nil {
			t.Fatalf("destroy the Surfer: %v", err)
		}
	})
	if !unmet() {
		t.Error("with Silver Surfer gone, Galactus owes no attack")
	}
}
