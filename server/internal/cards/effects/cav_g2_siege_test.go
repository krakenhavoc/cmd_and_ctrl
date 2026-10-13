package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A Siege defeated on someone else's turn is no longer lost: the grant has
// flash timing, so its controller casts the back face in whatever step the
// defeat happened in, and it closes when they pass priority.
func TestSiegeDefeatedOnAnotherPlayersTurnCanStillBeCast(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[(seat+1)%len(g.Seats)]
	protector := g.Seats[(seat+2)%len(g.Seats)]

	row := siegeRow(invasionOfKarsusOracleID, "Invasion of Karsus", "Refraction Elemental",
		"{2}{R}{R}", "4", "Creature — Elemental", "4", "4")
	id := importToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	answerProtectorPrompt(t, g, owner.ID, protector.ID)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, 4); err != nil {
			t.Fatalf("damage the battle: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	grant := exileGrantFor(g, id)
	if grant.Timing != game.TimingFlash {
		t.Errorf("grant timing = %q, want flash: the cast ignores timing", grant.Timing)
	}
	if grant.LapseOnPass != game.LapseStaysInExile {
		t.Errorf("grant lapse = %q, want the window to close on a pass and leave the card exiled", grant.LapseOnPass)
	}
	if g.Turn.Step != game.StepBeginCombat {
		t.Fatalf("expected to still be in the beginning of combat step, got %v", g.Turn.Step)
	}
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the back face off turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if landed := battlefieldCardFor(g, id); landed == nil || landed.Name != "Refraction Elemental" {
		t.Fatalf("the back face did not land: %+v", landed)
	}
}

// Invasion of New Phyrexia's Knights come from an entry trigger that reads
// the remembered X: the spell's resolution makes none, the trigger does.
func TestInvasionOfNewPhyrexiaKnightsComeFromTheEntryTrigger(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	id := importToHand(newPhyrexiaRow(), owner)
	toMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast the Siege for X = 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProtectorPrompt(t, g, owner.ID, protector.ID)
	if n := countNamed(g, "Knight"); n != 0 {
		t.Fatalf("%d Knights before the entry trigger resolved, want 0", n)
	}
	if triggerOnStack(g, id) == nil {
		t.Fatal("the Siege's enters trigger should be on the stack, a response window")
	}
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Knight"); n != 2 {
		t.Errorf("%d Knights after the trigger, want X = 2", n)
	}
}
