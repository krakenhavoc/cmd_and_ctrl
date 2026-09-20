package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const vaultbornTyrantOracle = "e8d0accc-b320-4c07-8a71-a09db860351e"

// castVaultbornTyrant casts the real printed 6/6 from hand.
// castCatalogSpell seeds only Name / TypeLine / OracleID, which would
// leave the entering permanent a 0/0 — wrong for a card whose own ETB
// ability reads its own power.
func castVaultbornTyrant(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Vaultborn Tyrant", TypeLine: "Creature — Dinosaur",
		OracleID: vaultbornTyrantOracle, Power: 6, Toughness: 6,
		Owner: p.ID, Controller: p.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Vaultborn Tyrant: %v", err)
	}
	return id
}

// TestVaultbornTyrantEntersItselfAndTriggersTheLifeDraw — "this
// creature OR another creature you control" includes itself, with no
// exclusion.
func TestVaultbornTyrantEntersItselfAndTriggersTheLifeDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beforeLife, beforeHand := me.Life, me.Hand.Size()
	tyrant := castVaultbornTyrant(t, g, me)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(tyrant) {
		t.Fatal("Vaultborn Tyrant is on the battlefield")
	}
	if got := me.Life; got != beforeLife+3 {
		t.Errorf("life %d, want %d (its own ETB triggers the ability)", got, beforeLife+3)
	}
	if got := me.Hand.Size(); got != beforeHand+1 {
		t.Errorf("hand %d, want %d", got, beforeHand+1)
	}
	if !hasEffectiveKeyword(t, g, tyrant, "trample") {
		t.Error("trample is printed")
	}
}

// TestVaultbornTyrantTriggersForAFourPowerCreatureButNotATwoPower —
// the power-4 threshold, read at the entering creature's CURRENT
// power.
func TestVaultbornTyrantTriggersForAFourPowerCreatureButNotATwoPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castVaultbornTyrant(t, g, me)
	passPriorityAroundTable(t, g)

	hand, life := me.Hand.Size(), me.Life
	top100CastCreature(t, g, me, "Small Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("power 2 does not meet the threshold: hand %d → %d, life %d → %d",
			hand, me.Hand.Size(), life, me.Life)
	}

	hand, life = me.Hand.Size(), me.Life
	top100CastCreature(t, g, me, "Big Beast", 4, 4)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+3 {
		t.Errorf("power 4 or greater triggers: hand %d → %d, life %d → %d",
			hand, me.Hand.Size(), life, me.Life)
	}
}

// TestVaultbornTyrantDiesIntoAnArtifactTokenCopyThatDoesNotChain
// covers the dies trigger end to end: the token is a copy that's also
// an artifact, and — because it's a token — its own death does not
// create a second copy.
func TestVaultbornTyrantDiesIntoAnArtifactTokenCopyThatDoesNotChain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tyrant := castVaultbornTyrant(t, g, me)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tyrant) })
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(tyrant) {
		t.Fatal("the original Vaultborn Tyrant left the battlefield")
	}
	if !me.Graveyard.Contains(tyrant) {
		t.Fatal("Vaultborn Tyrant is in the graveyard")
	}

	var token game.Card
	found := false
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := g.Battlefield.Cards[i]
			if c.OracleID == vaultbornTyrantOracle && c.InstanceID != tyrant {
				token, found = c, true
			}
		}
	})
	if !found {
		t.Fatal("no token copy was created when Vaultborn Tyrant died")
	}
	if !token.IsToken() {
		t.Error("the copy should be a token")
	}
	if !token.IsArtifact() {
		t.Error("the token should be an artifact in addition to its other types")
	}
	if !token.IsCreature() || !token.HasSubtype("Dinosaur") {
		t.Errorf("the token should keep its other types: %q", token.TypeLine)
	}

	// The token dying must NOT create another copy — it is a token,
	// so the "if it's not a token" clause refuses the trigger.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token.InstanceID) })
	passPriorityAroundTable(t, g)

	var secondCopies int
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].OracleID == vaultbornTyrantOracle {
				secondCopies++
			}
		}
	})
	if secondCopies != 0 {
		t.Errorf("the token's own death must not chain into another copy: %d Vaultborn Tyrants remain", secondCopies)
	}
}
