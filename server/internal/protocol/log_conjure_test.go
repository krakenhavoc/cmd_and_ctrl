package protocol

// log_conjure_test.go — ADR 0099 amendment 2026-11-24 (#5948). A
// conjured card is announced by its own line; a conjure into a hidden
// zone names the zone and not the card, as a spawn does.

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestConjureOntoTheBattlefieldNamesTheCard(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventConjured, Actor: me.ID, Target: me.ID,
			NewZone: game.ZoneBattlefield, Label: "Blood Artist", Amount: 1,
		})
	})
	e := findLog(t, ViewOfGame(g).Log, LogConjure)
	want := me.Name + " conjured Blood Artist onto the battlefield"
	if e.Text != want || e.Label != "Blood Artist" {
		t.Errorf("entry %+v, want text %q", e, want)
	}
}

func TestConjureIntoAHiddenZoneDoesNotNameTheCard(t *testing.T) {
	for _, zone := range []game.ZoneKind{game.ZoneHand, game.ZoneLibrary} {
		g := buildActiveGame(t)
		me := g.Seats[0]
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{
				Kind: game.EventConjured, Actor: me.ID, Target: me.ID,
				NewZone: zone, Label: "Mox Jet", Amount: 1,
			})
		})
		e := findLog(t, ViewOfGame(g).Log, LogConjure)
		if e.Label != "" || strings.Contains(e.Text, "Mox Jet") {
			t.Errorf("%s: the entry names the card: %+v", zone, e)
		}
		if want := me.Name + " conjured a card into their " + spawnZoneWord(string(zone)); e.Text != want {
			t.Errorf("%s: text %q, want %q", zone, e.Text, want)
		}
	}
}
