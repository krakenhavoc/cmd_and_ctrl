package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combinedOffer attacks and returns the combined offer over id, held
// inside the trigger's resolution.
func (tb *envoyTable) CombinedOffer(t *testing.T, id uuid.UUID) *game.PendingChoice {
	t.Helper()
	tb.Attack(t)
	c := KeywordCastPromptFor(tb.G, tb.Me.ID)
	if c == nil || c.MayCastKeyword != game.MayCastKeywordPlayOrFreeOrHand || KeywordCastCard(c) != id {
		t.Fatalf("offer %+v (in hand %v), want the combined land play or cast", c, tb.Me.Hand.Contains(id))
	}
	if !tb.G.ResolutionPaused() {
		t.Fatal("the offer is not held inside the trigger's resolution")
	}
	return c
}

// settled checks the offer is answered and its permission gone.
func (tb *envoyTable) Settled(t *testing.T, id uuid.UUID) {
	t.Helper()
	if KeywordCastPromptFor(tb.G, tb.Me.ID) != nil || tb.G.ResolutionPaused() {
		t.Error("the offer is still open")
	}
	var perm *game.CastPermission
	tb.G.ReadSnapshot(func() { perm = tb.G.CastPermissionOnCardByIDForEffect(id) })
	if perm != nil {
		t.Error("the free cast outlived the answer")
	}
}

// playFace plays the card as face through the offer.
func (tb *envoyTable) PlayFace(t *testing.T, c *game.PendingChoice, id uuid.UUID, face int, name string) {
	t.Helper()
	if err := tb.G.CastForResolutionChoice(tb.Me.ID, c.ID, id, game.CastSpellParams{FromZone: "library", Face: face}); err != nil {
		t.Fatalf("play face %d: %v", face, err)
	}
	land, ok := BattlefieldCard(tb.G, id)
	if !ok || land.ActiveFace != face || !land.IsLand() || land.Name != name {
		t.Fatalf("on the battlefield: %+v, want %s face up", land, name)
	}
	if got := tb.G.LandsPlayedThisTurnFor(tb.Me.ID); got != 1 {
		t.Errorf("lands played this turn = %d, want 1", got)
	}
	tb.Settled(t, id)
}

// castFace casts face through the offer.
func (tb *envoyTable) CastFace(t *testing.T, c *game.PendingChoice, id uuid.UUID, face int) game.Card {
	t.Helper()
	if err := tb.G.CastForResolutionChoice(tb.Me.ID, c.ID, id, game.CastSpellParams{FromZone: "library", Face: face}); err != nil {
		t.Fatalf("cast face %d: %v", face, err)
	}
	spell, ok := StackCardOf(tb.G, id)
	if !ok || spell.ActiveFace != face || spell.IsLand() {
		t.Fatalf("on the stack: %+v, want face %d as a spell", spell, face)
	}
	if got := tb.G.LandsPlayedThisTurnFor(tb.Me.ID); got != 0 {
		t.Errorf("lands played this turn = %d, want 0 after a cast", got)
	}
	tb.Settled(t, id)
	return spell
}

// decline answers the offer "neither": the card goes into the hand.
func (tb *envoyTable) Decline(t *testing.T, c *game.PendingChoice, id uuid.UUID) {
	t.Helper()
	if err := tb.G.ResolvePendingChoice(c.ID, tb.Me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !tb.Me.Hand.Contains(id) || tb.Me.Library.Contains(id) {
		t.Fatal("the declined card is not in the hand")
	}
	if c, _ := BattlefieldCard(tb.G, id); c.InstanceID == id {
		t.Error("a declined card reached the battlefield")
	}
	if got := tb.G.LandsPlayedThisTurnFor(tb.Me.ID); got != 0 {
		t.Errorf("lands played this turn = %d, want 0 after a decline", got)
	}
	tb.Settled(t, id)
}
