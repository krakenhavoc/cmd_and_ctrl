package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// pet_sehlat_test.go — #4973: "This creature can't attack alone."
// (CR 506.5, 508.1c-d; game.AttackLimitNotAlone).

const petSehlatOracle = "509f255f-c3ee-4836-a583-d018060d0c4d"

func petSehlat(g *game.Game) uuid.UUID {
	me := g.Seats[0]
	return apaPush(g, me.ID, me.ID, apaCreature("Pet Sehlat", petSehlatOracle, 4, 3))
}

func petSehlatAttacking(g *game.Game, id uuid.UUID) bool {
	c := findBattlefieldCardForTest(g, id)
	return c != nil && c.AttackingTarget != uuid.Nil
}

// petSehlatMoves is the active seat's legal moves: whether the pass is
// offered, and the attack moves by attacker with their AlwaysLegal flag.
func petSehlatMoves(g *game.Game) (pass bool, attacks map[uuid.UUID]bool) {
	attacks = map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, g.Seats[0].ID) {
		switch m.Kind {
		case legal.KindPass:
			pass = true
		case legal.KindAttack:
			attacks[m.Source] = attacks[m.Source] || m.AlwaysLegal
		}
	}
	return pass, attacks
}

// Declared alone, the declaration is refused at the active player's
// pass, the pass is withheld and the companion is the owed answer; with
// a second attacker it is legal.
func TestPetSehlatCantAttackAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sehlat := petSehlat(g)
	bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)

	if err := g.DeclareAttacker(sehlat, opp.ID); err != nil {
		t.Fatalf("declaring the Sehlat first was refused: %v", err)
	}
	pass, attacks := petSehlatMoves(g)
	if pass {
		t.Error("the pass is offered while the Sehlat attacks alone")
	}
	if !attacks[bear] {
		t.Errorf("the Bear is not the owed (always-legal) answer: %v", attacks)
	}
	err := g.PassPriority()
	if !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Sehlat alone: %v, want ErrAttackLimit", err)
	}
	var le *game.AttackLimitError
	if !errors.As(err, &le) || le.Sentence(me.ID) != "Pet Sehlat can't attack alone." || le.Attacker != sehlat {
		t.Errorf("refusal %+v", le)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrAttackLimit) {
		t.Errorf("leaving the step with the Sehlat alone: %v, want ErrAttackLimit", err)
	}

	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	if pass, _ := petSehlatMoves(g); !pass {
		t.Error("the pass is withheld with two attackers")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("passing with two attackers: %v", err)
	}
	if !petSehlatAttacking(g, sehlat) || !petSehlatAttacking(g, bear) {
		t.Error("both should be attacking")
	}
}

// Joining an attack already under way is fine.
func TestPetSehlatJoinsAnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sehlat := petSehlat(g)
	bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker bear: %v", err)
	}
	if err := g.DeclareAttacker(sehlat, g.Seats[2].ID); err != nil {
		t.Fatalf("the Sehlat beside another attacker was refused: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

// With nothing that could join it, the verb refuses the Sehlat outright,
// the enumerator never offers it, and the pass is free.
func TestPetSehlatWithNoCompanionIsNotOffered(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	sehlat := petSehlat(g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, attacks := petSehlatMoves(g); len(attacks) != 0 {
		t.Errorf("attacks offered with no companion: %v", attacks)
	}
	err := g.DeclareAttacker(sehlat, opp.ID)
	var le *game.AttackLimitError
	if !errors.As(err, &le) || le.Scope != game.AttackLimitNotAlone {
		t.Fatalf("declaring the Sehlat alone: %v, want the can't-attack-alone refusal", err)
	}
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: sehlat, Target: opp.ID}}); !errors.As(err, &le) {
		t.Fatalf("the bulk verb with the Sehlat alone: %v", err)
	}
	if petSehlatAttacking(g, sehlat) || findBattlefieldCardForTest(g, sehlat).Tapped {
		t.Error("a refused declaration changed the board")
	}
	if err := g.PassPriority(); err != nil {
		t.Errorf("pass: %v", err)
	}
}

// CR 508.1c's example: two creatures that each can't attack alone may
// attack together.
func TestTwoPetSehlatsAttackTogether(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a, b := petSehlat(g), petSehlat(g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a, Target: opp.ID}, {Attacker: b, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if !petSehlatAttacking(g, a) || !petSehlatAttacking(g, b) {
		t.Error("both Sehlats should be attacking")
	}
}

// The companion leaves after the Sehlat is declared: the declaration can
// no longer be made legal, so the Sehlat is taken out of it (untapped,
// its tap was the declaration's) rather than attacking alone, and the
// table moves on.
func TestPetSehlatLosesItsCompanion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sehlat := petSehlat(g)
	bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sehlat, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if pass, _ := petSehlatMoves(g); !pass {
		t.Fatal("the pass is withheld with nothing left to answer it — a wedge")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	c := findBattlefieldCardForTest(g, sehlat)
	if c.AttackingTarget != uuid.Nil || c.Tapped {
		t.Errorf("the Sehlat attacks alone (target %v, tapped %v)", c.AttackingTarget, c.Tapped)
	}
}

// CR 508.1d with "can't attack alone": a requirement on the Sehlat is
// obeyed only beside a companion. Goaded with a free creature beside it,
// the empty declaration is refused and the Sehlat must attack with
// company; goaded with nothing to join it, or under Silent Arbiter, it
// owes nothing.
func TestPetSehlatAgainstRequirements(t *testing.T) {
	t.Run("goaded with a companion", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		sehlat := petSehlat(g)
		bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
		g.WithWriteLock(func() { g.GoadForEffect(sehlat, opp.ID) })
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
			t.Fatalf("empty declaration: %v, want ErrAttackRequirement", err)
		}
		if err := g.DeclareAttacker(sehlat, third.ID); err != nil {
			t.Fatalf("DeclareAttacker sehlat: %v", err)
		}
		if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
			t.Fatalf("the goaded Sehlat alone: %v, want ErrAttackLimit", err)
		}
		if err := g.DeclareAttacker(bear, third.ID); err != nil {
			t.Fatalf("DeclareAttacker bear: %v", err)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	})
	t.Run("goaded alone", func(t *testing.T) {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		sehlat := petSehlat(g)
		g.WithWriteLock(func() { g.GoadForEffect(sehlat, opp.ID) })
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("a goaded Sehlat with nothing to join it owes nothing: %v", err)
		}
		if petSehlatAttacking(g, sehlat) {
			t.Error("the Sehlat attacks alone")
		}
	})
	t.Run("goaded under Silent Arbiter", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		sehlat := petSehlat(g)
		apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
		apaPush(g, opp.ID, opp.ID, apaCreature("Silent Arbiter", silentArbiterOracle, 1, 5))
		g.WithWriteLock(func() { g.GoadForEffect(sehlat, opp.ID) })
		advanceTo(t, g, game.StepDeclareAttackers)
		if _, attacks := petSehlatMoves(g); attacks[sehlat] {
			t.Error("the Sehlat is offered where nothing could join it")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("one attacker at most, so the Sehlat owes nothing: %v", err)
		}
	})
	t.Run("goaded companion", func(t *testing.T) {
		// A goaded Bear and a free Sehlat: the Bear owes the attack;
		// the Sehlat may join it, and the pass is legal either way.
		g := newCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		sehlat := petSehlat(g)
		bear := apaPush(g, me.ID, me.ID, apaCreature("Bear", "", 2, 2))
		g.WithWriteLock(func() { g.GoadForEffect(bear, opp.ID) })
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, third.ID); err != nil {
			t.Fatalf("DeclareAttacker bear: %v", err)
		}
		if err := g.DeclareAttacker(sehlat, third.ID); err != nil {
			t.Fatalf("DeclareAttacker sehlat: %v", err)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	})
}
