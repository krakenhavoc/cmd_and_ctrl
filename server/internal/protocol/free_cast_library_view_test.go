package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_cast_library_view_test.go — Aetherworks Marvel's look (#2110): a
// card the looker only LOOKED at on top of their library, picked for a
// free cast. Nobody else may learn the card, the pick or the permission.

func TestLibraryFreeCastPickPromptDeclaresTheCastToItsChooserOnly(t *testing.T) {
	g := newTwoSeatGame(t)
	me, them := g.Seats[0], g.Seats[1]
	top := game.NewCard("Looked At", me.ID)
	top.TypeLine = "Instant"
	top.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Library.PushTop(top)
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:     me.ID,
			Question:    "cast a spell from among the top cards of your library without paying its mana cost?",
			Cards:       []uuid.UUID{top.InstanceID},
			Max:         1,
			Zone:        game.ZoneLibrary,
			Destination: game.ChooseCastFreeFromLibrary,
		})
	})
	mine := FilterViewFor(ViewOfGame(g), me.ID.String()).PendingChoices[0]
	if mine.ChooseDestination != "cast_free_library" || len(mine.Options) != 1 {
		t.Errorf("chooser reads destination %q with %d options, want cast_free_library with 1", mine.ChooseDestination, len(mine.Options))
	}
	theirs := FilterViewFor(ViewOfGame(g), them.ID.String()).PendingChoices
	if len(theirs) > 0 && (theirs[0].ChooseDestination != "" || len(theirs[0].Options) != 0) {
		t.Errorf("an opponent reads destination %q with %d options, want both withheld", theirs[0].ChooseDestination, len(theirs[0].Options))
	}
}

func TestALookedAtLibraryPickIsCastableForItsOwnerOnly(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	const oracle = "test-view-library-free-cast"
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})
	withTopVisibility(t, "unused", game.LibraryTopHidden)

	me.Library.Cards = nil
	pick := game.NewCard("Looked At Pick", me.ID)
	pick.TypeLine = "Sorcery"
	pick.ManaCost = "{5}{R}"
	pick.OracleID = oracle
	pick.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Library.PushTop(pick)

	g.WithWriteLock(func() {
		g.FreeCastFromLibraryForEffect(me.ID, uuid.Nil, pick.InstanceID, "Aetherworks Marvel")
	})

	card := cardInZone(ViewOfGameFor(g, me.ID.String()).Seats[0].Library, pick.InstanceID)
	if card == nil || card.Name == "" || !card.CastableHere {
		t.Fatalf("the owner sees the looked-at pick, castable: %+v", card)
	}
	var theirs *CardView
	for _, s := range ViewOfGameFor(g, them.ID.String()).Seats {
		if s.ID == me.ID.String() {
			theirs = cardInZone(s.Library, pick.InstanceID)
		}
	}
	if theirs != nil && (theirs.Name != "" || theirs.CastableHere) {
		t.Errorf("an opponent sees the looked-at card or its permission: %+v", theirs)
	}
}
