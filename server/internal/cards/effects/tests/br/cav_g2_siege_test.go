package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// A Siege defeated on someone else's turn is no longer lost: the grant has
// flash timing, so its controller casts the back face in whatever step the
// defeat happened in, and it closes when they pass priority.
func TestSiegeDefeatedOnAnotherPlayersTurnCanStillBeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[(seat+1)%len(g.Seats)]
	protector := g.Seats[(seat+2)%len(g.Seats)]

	row := effectstest.SiegeRow(effects.InvasionOfKarsusOracleID, "Invasion of Karsus", "Refraction Elemental",
		"{2}{R}{R}", "4", "Creature — Elemental", "4", "4")
	id := effectstest.ImportToHand(row, owner)
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
		game.ZoneRef{Kind: game.ZoneBattlefield}, id,
	); err != nil {
		t.Fatalf("move battle to battlefield: %v", err)
	}
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, 4); err != nil {
			t.Fatalf("damage the battle: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	grant := effectstest.ExileGrantFor(g, id)
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
	effectstest.PassPriorityAroundTable(t, g)
	if landed := effectstest.BattlefieldCardFor(g, id); landed == nil || landed.Name != "Refraction Elemental" {
		t.Fatalf("the back face did not land: %+v", landed)
	}
}

// Invasion of New Phyrexia's Knights come from an entry trigger that reads
// the remembered X: the spell's resolution makes none, the trigger does.
func TestInvasionOfNewPhyrexiaKnightsComeFromTheEntryTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]

	id := effectstest.ImportToHand(effectstest.NewPhyrexiaRow(), owner)
	effectstest.ToMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast the Siege for X = 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
	if n := effectstest.CountNamed(g, "Knight"); n != 0 {
		t.Fatalf("%d Knights before the entry trigger resolved, want 0", n)
	}
	if effectstest.TriggerOnStack(g, id) == nil {
		t.Fatal("the Siege's enters trigger should be on the stack, a response window")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Knight"); n != 2 {
		t.Errorf("%d Knights after the trigger, want X = 2", n)
	}
}
