package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const reprieveOracle = "f0449cbd-855c-4c20-a1f5-f76a395d8d39"

func castReprieveOn(t *testing.T, g *game.Game, caster *game.Player, target uuid.UUID) {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Reprieve", TypeLine: "Instant",
		OracleID: reprieveOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("CastSpell Reprieve: %v", err)
	}
}

// TestReprieveReturnsSpellAndDraws is the plain shape: the targeted
// spell ends up in its owner's hand, off the stack, and the caster
// draws a card.
func TestReprieveReturnsSpellAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]

	shockID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: shockID, Name: "Shock", TypeLine: "Instant",
		OracleID: b5ShockOracle, Owner: opponent.ID, Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, shockID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Opponent CastSpell Shock: %v", err)
	}

	handBefore := len(caster.Hand.Cards)
	castReprieveOn(t, g, caster, shockID)
	passPriorityAroundTable(t, g)

	if g.Stack.Contains(shockID) {
		t.Error("Shock should be off the stack")
	}
	if !opponent.Hand.Contains(shockID) {
		t.Error("Shock should be back in its owner's hand, not countered into the graveyard")
	}
	if opponent.Graveyard.Contains(shockID) {
		t.Error("Reprieve does not counter — Shock must not land in a graveyard")
	}
	// castReprieveOn pushes Reprieve to hand then casts it (net zero),
	// so the only change left is the draw.
	if got := len(caster.Hand.Cards); got != handBefore+1 {
		t.Errorf("caster hand size %d -> %d, want +1 from the draw", handBefore, got)
	}
}

// TestReprieveIgnoresCantBeCountered is the whole reason
// ReturnSpellToHandForEffect exists beside counterSpellLocked: CR
// 701.6 never applies to Reprieve, so Supreme Verdict's printed
// "can't be countered" does not stop the bounce.
func TestReprieveIgnoresCantBeCountered(t *testing.T) {
	g := newCatalogGame(t)
	active, responder := g.Seats[0], g.Seats[1]
	// Supreme Verdict is a sorcery: it needs the active seat's own
	// main phase and an empty stack. Reprieve, an instant, answers it
	// from the other seat.
	toMain(t, g)

	verdictID := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: verdictID, Name: "Supreme Verdict", TypeLine: "Sorcery",
		OracleID: supremeVerdictOracle, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, verdictID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Supreme Verdict: %v", err)
	}

	eventsBefore := countEventsOf(g, game.EventCounterSpell, verdictID)
	castReprieveOn(t, g, responder, verdictID)
	passPriorityAroundTable(t, g)

	if g.Stack.Contains(verdictID) {
		t.Error("Supreme Verdict should be off the stack")
	}
	if !active.Hand.Contains(verdictID) {
		t.Error("a spell that can't be countered is still bounced by Reprieve")
	}
	if got := countEventsOf(g, game.EventCounterSpell, verdictID); got != eventsBefore {
		t.Errorf("EventCounterSpell fired %d times, want 0 — Reprieve is not a counter", got-eventsBefore)
	}
}
