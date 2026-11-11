package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const magmaticInsightOracle = "6a9031b6-3f4d-4a36-90e1-ac77b8518cee"

// magmaticInsightInHand puts the spell in the active seat's otherwise
// empty hand at a main phase with {R} floating.
func magmaticInsightInHand(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Magmatic Insight", TypeLine: "Sorcery", ManaCost: "{R}",
		OracleID: magmaticInsightOracle, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	floatMana(t, g, me, "{R}")
	return g, me, id
}

// The land card is discarded as the spell is cast (CR 601.2h), and the
// spell draws two on resolution.
func TestMagmaticInsightDiscardsALandCardAndDrawsTwo(t *testing.T) {
	g, me, id := magmaticInsightInHand(t)
	land := handCard(me, "Mountain", "Basic Land — Mountain")
	before := me.Library.Size()

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !me.Graveyard.Contains(land) {
		t.Fatalf("the land card should be discarded at announce, not on resolution")
	}
	if got := before - me.Library.Size(); got != 0 {
		t.Fatalf("drew %d cards before the spell resolved", got)
	}
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

// Only a land card pays the cost: a nonland card is refused, and with no
// land card in hand the spell cannot be cast at all (CR 118.3).
func TestMagmaticInsightRefusesANonlandDiscard(t *testing.T) {
	g, me, id := magmaticInsightInHand(t)
	spell := handCard(me, "Shock", "Instant")

	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, DiscardIDs: []uuid.UUID{spell}})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("discarding a nonland card: err = %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatalf("cast with no discard named was accepted")
	}
	if !me.Hand.Contains(id) || !me.Hand.Contains(spell) || g.Stack.Contains(id) {
		t.Fatalf("a refused cast moved a card")
	}
	// An artifact land is a land card.
	land := handCard(me, "Darksteel Citadel", "Artifact Land")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("discarding an artifact land: %v", err)
	}
}
