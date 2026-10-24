package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_cast_view_test.go — #2947's wire: a resolution_cast prompt
// shows the copies still on offer, face up to every viewer (they sit in
// exile or a graveyard and are printed on a face-up card), and lists in
// `eligible` only the ones that can still be cast.
func TestResolutionCastPromptShowsTheCopiesOnOffer(t *testing.T) {
	g := buildThreeSeatGame(t)
	chooser, other := g.Seats[0], g.Seats[1]
	var origs []uuid.UUID
	for _, name := range []string{"Fixture A", "Fixture B"} {
		id := uuid.New()
		chooser.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Instant", ManaCost: "{R}",
			Owner: chooser.ID, Controller: chooser.ID,
		})
		origs = append(origs, id)
	}
	var copies []uuid.UUID
	g.WithWriteLock(func() {
		copies = g.CopyCardsAndOfferCastsForEffect(game.CardCopyOffer{
			Player: chooser.ID, SourceName: "Fixture copier", Cost: "{0}",
			Timing: game.TimingFlash, Lapse: game.LapseStaysInExile,
		}, origs)
		// A third id that is no copy on offer is not shown.
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind: game.PendingChoiceResolutionCast, Chooser: chooser.ID, Count: 1,
			Reason: "Fixture copier", DiscardOptions: append(append([]uuid.UUID(nil), copies...), origs[0]),
		})
	})
	if len(copies) != 2 {
		t.Fatalf("copies = %v", copies)
	}
	for name, viewer := range map[string]string{"the chooser": chooser.ID.String(), "another seat": other.ID.String()} {
		c := choiceFor(t, ViewOfGameFor(g, viewer), string(game.PendingChoiceResolutionCast))
		got := optionIDs(c)
		if len(c.Options) != 2 || !got[copies[0].String()] || !got[copies[1].String()] {
			t.Errorf("%s sees options %+v, want the two copies", name, c.Options)
		}
		for _, o := range c.Options {
			if o.Name == "" {
				t.Errorf("%s sees a copy face down: %+v", name, o)
			}
		}
		if len(c.Eligible) != 2 {
			t.Errorf("%s: eligible = %v, want the two copies", name, c.Eligible)
		}
	}
}

// #4243: cascade's and discover's cast is a resolution_cast over a real card
// in exile, not a copy. The wire shows that card face up, and carries the
// keyword and the words of the two answers, as the may_cast it replaces did.
// A card in exile that no permission of the chooser's names is not shown.
func TestResolutionCastPromptForACascadeHit(t *testing.T) {
	g := buildThreeSeatGame(t)
	chooser, other := g.Seats[0], g.Seats[1]
	hit := game.Card{InstanceID: uuid.New(), Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{U}", Owner: chooser.ID, Controller: chooser.ID}
	stray := game.Card{InstanceID: uuid.New(), Name: "Stray", TypeLine: "Instant", ManaCost: "{U}", Owner: chooser.ID, Controller: chooser.ID}
	// Both cards are exiled face up, as cascade's walk leaves them: known
	// to every seat.
	hit.KnownBy = map[uuid.UUID]bool{}
	stray.KnownBy = map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		hit.KnownBy[s.ID] = true
		stray.KnownBy[s.ID] = true
	}
	g.Exile.PushTop(hit)
	g.Exile.PushTop(stray)
	g.WithWriteLock(func() {
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player: chooser.ID, Zone: game.ZoneExile, Duration: g.UntilEndOfTurnDuration(),
			Cost: "{0}", Timing: game.TimingFlash, CastOnly: true, LapseOnPass: game.LapseToLibraryBottom,
			Label: "Cascade — cast it without paying its mana cost",
		}, []game.Card{hit})
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind: game.PendingChoiceResolutionCast, Chooser: chooser.ID, Count: 1,
			Reason: "Cascade", DiscardOptions: []uuid.UUID{hit.InstanceID, stray.InstanceID},
			CastPile:       []uuid.UUID{uuid.New()},
			MayCastKeyword: game.MayCastKeywordCascade, AcceptLabel: "Cast it free", DeclineLabel: "Put it on the bottom",
		})
	})
	for name, viewer := range map[string]string{"the chooser": chooser.ID.String(), "another seat": other.ID.String()} {
		c := choiceFor(t, ViewOfGameFor(g, viewer), string(game.PendingChoiceResolutionCast))
		if len(c.Options) != 1 || c.Options[0].InstanceID != hit.InstanceID.String() || c.Options[0].Name != "Cheap Spell" {
			t.Errorf("%s sees options %+v, want the hit alone, face up", name, c.Options)
		}
		if len(c.Eligible) != 1 || c.Eligible[0] != hit.InstanceID.String() {
			t.Errorf("%s: eligible = %v, want the hit", name, c.Eligible)
		}
		if c.MayCastKeyword != "cascade" || c.AcceptLabel != "Cast it free" || c.DeclineLabel != "Put it on the bottom" {
			t.Errorf("%s: words = %q / %q / %q, want cascade's", name, c.MayCastKeyword, c.AcceptLabel, c.DeclineLabel)
		}
	}
}
