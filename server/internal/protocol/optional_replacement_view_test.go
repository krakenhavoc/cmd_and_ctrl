package protocol

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Stinkweed Imp
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// optional_replacement_view_test.go — #2390. An optional_replacement
// prompt carries the facts its yes turns on: `dredge` (N) on a dredge
// offer, and `playable_from_zone` on CR 903.9b's commander question
// when the commander is headed for a hand.

func onlyChoiceView(t *testing.T, g *game.Game) PendingChoiceView {
	t.Helper()
	v := ViewOfGame(g)
	if len(v.PendingChoices) != 1 {
		t.Fatalf("expected one prompt on the wire, got %d", len(v.PendingChoices))
	}
	return v.PendingChoices[0]
}

func TestDredgeOfferCarriesItsCount(t *testing.T) {
	g := buildActiveGame(t)
	p := g.Seats[0]
	imp := game.Card{
		InstanceID: uuid.New(), Name: "Stinkweed Imp", TypeLine: "Creature — Imp",
		OracleID: "e005bc76-4985-4ec6-b9f6-cf6d0d9f5df4", Owner: p.ID, Controller: p.ID,
	}
	g.WithWriteLock(func() {
		for i := range 8 {
			p.Library.PushTop(game.NewCard(fmt.Sprintf("Library %d", i), p.ID))
		}
		p.Graveyard.PushTop(imp)
	})
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	ch := onlyChoiceView(t, g)
	if ch.Kind != string(game.PendingChoiceOptionalReplacement) {
		t.Fatalf("prompt kind %q, want optional_replacement", ch.Kind)
	}
	if ch.Dredge != 5 || ch.Source != imp.InstanceID.String() {
		t.Errorf("dredge %d from %q, want Dredge 5 from the Imp %s", ch.Dredge, ch.Source, imp.InstanceID)
	}
	if ch.PlayableFromZone {
		t.Error("a dredge offer is not the commander question")
	}
}

func TestCommanderHeadedForAHandIsPlayableFromZone(t *testing.T) {
	for _, tc := range []struct {
		name string
		move func(g *game.Game, id uuid.UUID) error
		want bool
	}{
		{"bounced to hand", func(g *game.Game, id uuid.UUID) error { return g.BounceToHandForEffect(id) }, true},
		{"tucked into library", func(g *game.Game, id uuid.UUID) error { return g.TuckToLibraryForEffect(id, false) }, false},
	} {
		g := buildActiveGame(t)
		owner := g.Seats[0]
		cmd := game.NewCommander("Bounced Commander", owner.ID)
		cmd.TypeLine = "Legendary Creature — Elf"
		cmd.Controller = owner.ID
		var err error
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(cmd)
			err = tc.move(g, cmd.InstanceID)
		})
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		ch := onlyChoiceView(t, g)
		if ch.Kind != string(game.PendingChoiceOptionalReplacement) || ch.Source != cmd.InstanceID.String() {
			t.Fatalf("%s: prompt %q about %q, want the CR 903.9b question about the commander", tc.name, ch.Kind, ch.Source)
		}
		if ch.PlayableFromZone != tc.want {
			t.Errorf("%s: playable_from_zone = %v, want %v", tc.name, ch.PlayableFromZone, tc.want)
		}
		if ch.Dredge != 0 {
			t.Errorf("%s: dredge = %d on the commander question", tc.name, ch.Dredge)
		}
	}
}

// #2056: a "may" redirection (Blood of the Martyr) carries the damage,
// the creature it would be dealt to and the player the yes deals it to.
func TestOptionalRedirectionCarriesItsOffer(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 4,
		Owner: opp.ID, Controller: opp.ID}
	rod := game.Card{InstanceID: uuid.New(), Name: "Rod", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(bear)
		g.Battlefield.PushTop(rod)
		g.RecomputeLayersIfStaleLocked()
		if !g.RedirectDamageThisTurnForEffect(game.DamageRedirection{
			Controller: me.ID, To: me.ID, Optional: true, Label: "Blood of the Martyr",
			ProtectSet: game.DamageRecipientFilter{Permanents: true, Match: []game.PermanentQuery{{Types: []string{"creature"}}}},
		}) {
			t.Fatal("no redirection")
		}
		if err := g.DealDamageToCreatureForEffect(rod.InstanceID, bear.InstanceID, 3); err != nil {
			t.Fatal(err)
		}
	})
	ch := onlyChoiceView(t, g)
	if ch.Kind != string(game.PendingChoiceOptionalReplacement) || ch.Chooser != me.ID.String() {
		t.Fatalf("prompt %q for %q, want the may for its controller", ch.Kind, ch.Chooser)
	}
	if ch.RedirectAmount != 3 || ch.RedirectFrom != bear.InstanceID.String() || ch.RedirectTo != me.ID.String() {
		t.Errorf("offer %d from %q to %q, want 3 from the Bear to me", ch.RedirectAmount, ch.RedirectFrom, ch.RedirectTo)
	}
	if ch.Dredge != 0 || ch.PlayableFromZone {
		t.Error("a redirection is neither a dredge nor the commander question")
	}
}
