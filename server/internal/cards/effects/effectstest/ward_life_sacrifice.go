package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ward_life_sacrifice_test.go — the two non-mana ward costs.
//
//	ward—pay N life        Refraction Elemental, Sedgemoor Witch
//	ward—sacrifice a creature   Vein Ripper
//
// ward_test.go pins what every ward shares (a legal target, the caster
// pays, own spells don't trigger it). These pin what the cost changes:
// who is asked, what the prompt charges, what "can't pay" means for
// that cost, and that paying actually takes the life or the creature.

const (
	RefractionElementalOracle = effects.InvasionOfKarsusOracleID + "#1"
	SedgemoorWitchOracle      = "25dce517-ac0a-4577-89ed-04296c7c4069"
	VeinRipperOracle          = "f83a768c-162d-46f0-8ebd-d9c6b2c69322"
)

// wardTable puts a catalog ward creature on seat 0's battlefield, moves
// to a main phase and hands priority to seat 1, the caster.
func WardTable(t *testing.T, name, typeLine, oracle string) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warded := PushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	return g, me, opp, warded
}

// passUntilConfirmFor passes priority until the ward trigger resolves
// and its prompt reaches chooser, or the stack settles without one.
func PassUntilConfirmFor(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 8; i++ {
		if c := ConfirmChoiceFor(g, chooser); c != nil {
			return c
		}
		if StackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return ConfirmChoiceFor(g, chooser)
}

// settleDrainingAt passes priority to an empty stack, answering every
// Vein Ripper drain target prompt addressed to controller with victim.
func SettleDrainingAt(t *testing.T, g *game.Game, controller, victim uuid.UUID) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if p := LatestPickTarget(g, controller); p != nil {
			if err := g.ResolvePickTarget(p.ID, controller, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("stack did not settle")
}
