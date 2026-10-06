package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// commander_return_view_test.go — ADR 0115 decision 2: a
// commander_return prompt carries `playable_from_zone`, computed on
// every view from the cast path's own functions. Set for a commander
// its owner could cast from the graveyard it is in (here, a flashback
// declared by the catalog), absent for one it could not.
func TestCommanderReturnCarriesPlayableFromZone(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-commander-flashback"

	prevZones := game.CatalogCastableZones
	game.CatalogCastableZones = func(id string) []game.ZoneKind {
		if id == oracle {
			return []game.ZoneKind{game.ZoneGraveyard}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastableZones = prevZones })
	prevAlts := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id != oracle {
			return nil
		}
		return []game.AlternativeCost{
			{Key: "flashback", Label: "Flashback {2}{G}", ManaCost: "{2}{G}", FromZone: game.ZoneGraveyard},
		}
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prevAlts })

	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	castable := game.NewCommander("Castable Commander", me.ID)
	castable.TypeLine = "Legendary Creature — Elf"
	castable.OracleID = oracle
	castable.KnownBy = seen
	me.Graveyard.PushTop(castable)
	plain := game.NewCommander("Plain Commander", me.ID)
	plain.TypeLine = "Legendary Creature — Elf"
	plain.KnownBy = seen
	me.Graveyard.PushTop(plain)

	ids := map[uuid.UUID]uuid.UUID{}
	g.WithWriteLock(func() {
		for _, c := range []game.Card{castable, plain} {
			ids[c.InstanceID] = g.QueueChoiceForEffect(game.PendingChoice{
				Kind:    game.PendingChoiceCommanderReturn,
				Chooser: me.ID,
				Count:   1,
				Source:  c.InstanceID,
				Reason:  c.Name + " — put it into the command zone?",
			})
		}
	})

	v := ViewOfGameFor(g, me.ID.String())
	got := map[string]bool{}
	for _, ch := range v.PendingChoices {
		if ch.Kind != string(game.PendingChoiceCommanderReturn) {
			continue
		}
		got[ch.ID] = ch.PlayableFromZone
		if ch.Source == "" {
			t.Errorf("prompt %s has no source", ch.ID)
		}
	}
	if !got[ids[castable.InstanceID].String()] {
		t.Error("a commander castable from its graveyard is not playable_from_zone")
	}
	if p, ok := got[ids[plain.InstanceID].String()]; !ok || p {
		t.Errorf("a plain commander: playable_from_zone = %v (present %v), want false", p, ok)
	}
}

// TestCommanderReturnIsVisibleToTheOtherSeats — ADR 0115 PR 5: the
// seats that are not deciding see whose question it is and which card
// it is about, so the client can say "X is deciding" and name the
// commander. The chooser and the source are public on every view; the
// source sits in the owner's graveyard, which is public too.
func TestCommanderReturnIsVisibleToTheOtherSeats(t *testing.T) {
	g := buildActiveGame(t)
	me, other := g.Seats[0], g.Seats[1]

	cmd := game.NewCommander("Public Commander", me.ID)
	cmd.TypeLine = "Legendary Creature — Elf"
	cmd.KnownBy = map[uuid.UUID]bool{me.ID: true, other.ID: true}
	me.Graveyard.PushTop(cmd)
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:    game.PendingChoiceCommanderReturn,
			Chooser: me.ID,
			Count:   1,
			Source:  cmd.InstanceID,
			Reason:  cmd.Name + " — put it into the command zone?",
		})
	})

	v := ViewOfGameFor(g, other.ID.String())
	var found *PendingChoiceView
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoiceCommanderReturn) {
			found = &v.PendingChoices[i]
		}
	}
	if found == nil {
		t.Fatal("the other seat's view has no commander_return prompt")
	}
	if found.Chooser != me.ID.String() {
		t.Errorf("chooser = %s, want the owner %s", found.Chooser, me.ID)
	}
	if found.Source != cmd.InstanceID.String() {
		t.Errorf("source = %s, want the commander %s", found.Source, cmd.InstanceID)
	}
	named := false
	for _, s := range v.Seats {
		if s.ID != me.ID.String() {
			continue
		}
		for _, c := range s.Graveyard.Cards {
			if c.InstanceID == cmd.InstanceID.String() && c.Name == cmd.Name {
				named = true
			}
		}
	}
	if !named {
		t.Error("the commander is not named in the owner's graveyard on the other seat's view")
	}
}
