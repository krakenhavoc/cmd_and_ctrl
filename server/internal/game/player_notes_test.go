package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// player_notes_test.go — #6097: the words an object's abilities record for
// a player (Archangel of Strife's war or peace, Diseased Vermin's damaged
// players).

func TestPlayerNotesAreKeptPerPlayerAndOnlyOnTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := pushBattlefieldForTest(g, me.ID, "Archangel", "Creature — Angel", "")
	g.WithWriteLock(func() {
		if !g.SetPlayerNoteForEffect(angel, me.ID, "war") || !g.SetPlayerNoteForEffect(angel, opp.ID, "peace") {
			t.Fatal("a choice was not recorded")
		}
		// One answer per player: a new choice replaces the old one.
		g.SetPlayerNoteForEffect(angel, opp.ID, " War ")
		if !g.AddPlayerNoteForEffect(angel, opp.ID, "damaged") {
			t.Fatal("a mark was not recorded")
		}
		if g.AddPlayerNoteForEffect(angel, opp.ID, "DAMAGED") || g.AddPlayerNoteForEffect(angel, opp.ID, " ") {
			t.Error("a mark already recorded, or an empty one, was recorded again")
		}
	})
	c := findBattlefieldCard(g, angel)
	if !c.PlayerNoted(opp.ID, "war") || c.PlayerNoted(opp.ID, "peace") {
		t.Error("the opponent's second choice did not replace the first")
	}
	if got := c.PlayersNoted("war"); !slices.Equal(got, []uuid.UUID{me.ID, opp.ID}) {
		t.Errorf("players who chose war = %v, want me and the opponent", got)
	}
	if c.PlayerNoted(me.ID, "damaged") || !c.PlayerNoted(opp.ID, "damaged") {
		t.Error("a mark landed on the wrong player")
	}

	// A copy taken now does not share the list with the live card.
	clone := g.Clone()
	g.WithWriteLock(func() { g.SetPlayerNoteForEffect(angel, me.ID, "peace") })
	clone.WithWriteLock(func() {
		if !findBattlefieldCard(clone, angel).PlayerNoted(me.ID, "war") {
			t.Error("a later choice on the original reached the clone")
		}
	})

	// Carried by the snapshot.
	_, restored := roundTrip(t, g)
	restored.WithWriteLock(func() {
		got := restored.PlayerNotesForEffect(angel)
		if len(got) != 3 || !findBattlefieldCard(restored, angel).PlayerNoted(me.ID, "peace") {
			t.Errorf("restored player notes = %+v", got)
		}
	})

	// CR 400.7: the permanent that leaves keeps nothing, though what it
	// had is still readable as it last existed (CR 608.2h).
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, angel); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if got := g.PlayerNotesForEffect(angel); len(got) != 3 {
			t.Errorf("last-known player notes = %+v", got)
		}
		if g.AddPlayerNoteForEffect(angel, me.ID, "damaged") || g.SetPlayerNoteForEffect(angel, me.ID, "war") {
			t.Error("a card off the battlefield was given a note")
		}
	})
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == angel && len(c.PlayerNotes) != 0 {
			t.Errorf("the card in the graveyard still has player notes %+v", c.PlayerNotes)
		}
	}
}
