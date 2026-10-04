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
