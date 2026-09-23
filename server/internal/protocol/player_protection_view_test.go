package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// player_protection_view_test.go — #1197, the projection half of
// player protection and player hexproof.
//
// Two things reach the client, and they are the two the seat badge
// and the picker need: the player's own ability tokens, and a
// `legal_targets` stamp that already excludes a seat the announce
// gate would refuse.
//
// The second is the one this seam must not get wrong. The view
// computes it from legalTargetsLocked, the same function the announce
// gate and the bot's enumerator read, so the three cannot disagree —
// but "cannot disagree" is a claim about a call graph, and this is
// the test that keeps it one.

// TestHexproofPlayerIsMissingFromAnOpponentsLegalTargets is the
// coherence assertion, written PER VIEWER: the same spell in two
// different hands produces two different player lists, because
// hexproof asks who is casting.
func TestHexproofPlayerIsMissingFromAnOpponentsLegalTargets(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	const bolt = "test-view-player-hexproof-bolt"
	const leyline = "test-view-player-hexproof-leyline"

	prevSpec := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != bolt {
			return nil
		}
		return &game.TargetSpec{Mode: "player", Label: "target player", Players: true, Min: 1, Max: 1}
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prevSpec })

	prevKw := game.CatalogPlayerKeywords
	game.CatalogPlayerKeywords = func(id string) []string {
		if id == leyline {
			return []string{"hexproof"}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogPlayerKeywords = prevKw })

	wall := game.NewCard("Leyline of Sanctity", me.ID)
	wall.TypeLine = "Enchantment"
	wall.OracleID = leyline
	wall.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.Battlefield.PushTop(wall)

	// The stamp is computed for the card's OWNER's view, so each seat
	// gets its own copy of the spell.
	stampFor := func(caster *game.Player) []string {
		t.Helper()
		spell := game.NewCard("Test Bolt", caster.ID)
		spell.TypeLine = "Instant"
		spell.OracleID = bolt
		spell.KnownBy = map[uuid.UUID]bool{caster.ID: true}
		caster.Hand.PushTop(spell)
		defer func() { _, _ = caster.Hand.Remove(spell.InstanceID) }()

		view := ViewOfGameFor(g, caster.ID.String())
		for _, seat := range view.Seats {
			if seat.ID != caster.ID.String() {
				continue
			}
			for _, c := range seat.Hand.Cards {
				if c.InstanceID == spell.InstanceID.String() {
					if c.LegalTargets == nil {
						t.Fatal("no legal_targets stamped on a targeted spell")
					}
					return c.LegalTargets.Players
				}
			}
		}
		t.Fatal("the spell is not in its caster's own hand view")
		return nil
	}

	mine := stampFor(me)
	theirs := stampFor(opp)

	if !hasPlayerID(mine, me.ID) {
		t.Error("the hexproof player is missing from their OWN legal_targets — they must be able to target themselves")
	}
	if hasPlayerID(theirs, me.ID) {
		t.Error("an opponent's legal_targets still offers the hexproof player; the picker and the announce gate disagree")
	}
	if !hasPlayerID(theirs, opp.ID) {
		t.Error("the opponent lost themselves from their own list; the gate is refusing too much")
	}
}

// TestPlayerKeywordsAreProjectedToEveryViewer — protection and
// hexproof on a seat are facts about the board, so every viewer sees
// them. A viewer who could see the refusal (no legal target) but not
// its reason is strictly worse off than one who sees neither.
func TestPlayerKeywordsAreProjectedToEveryViewer(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	g.WithWriteLock(func() {
		g.GrantPlayerStaticForEffect(me.ID, "protection from everything",
			"Teferi's Protection", uuid.Nil, g.UntilYourNextTurnDuration(me.ID))
	})

	for _, viewer := range []*game.Player{me, opp} {
		view := ViewOfGameFor(g, viewer.ID.String())
		var got []string
		for _, seat := range view.Seats {
			if seat.ID == me.ID.String() {
				got = seat.Keywords
			}
		}
		if len(got) != 1 || got[0] != "protection from everything" {
			t.Errorf("viewer %s sees keywords %v on the protected seat, want [protection from everything]",
				viewer.Name, got)
		}
	}

	// And a seat with nothing carries nothing, so `omitempty` keeps
	// the field off the wire for nearly every seat in nearly every
	// game.
	view := ViewOfGameFor(g, me.ID.String())
	for _, seat := range view.Seats {
		if seat.ID == opp.ID.String() && len(seat.Keywords) != 0 {
			t.Errorf("an unprotected seat carries keywords %v", seat.Keywords)
		}
	}
}

func hasPlayerID(ids []string, want uuid.UUID) bool {
	for _, id := range ids {
		if id == want.String() {
			return true
		}
	}
	return false
}
