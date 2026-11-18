package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_mutate_test.go — ADR 0223: a mutated permanent's components on
// the wire (merged_from, topmost first) and the mutate log line.

func TestAMutatedPermanentShowsItsComponentsAndLogsTheMutate(t *testing.T) {
	g, me, opp := attachTable(t)
	host := uuid.New()
	phoenix := uuid.New()
	known := map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: host, Name: "Everquill Phoenix", ScryfallID: "phx",
			TypeLine: "Creature — Phoenix", Power: 4, Toughness: 4,
			Owner: me.ID, Controller: me.ID, KnownBy: known,
			MergedFrom: []game.Card{
				{InstanceID: phoenix, Name: "Everquill Phoenix", ScryfallID: "phx", TypeLine: "Creature — Phoenix", ManaCost: "{2}{R}{R}", Power: 4, Toughness: 4, Owner: me.ID},
				{InstanceID: host, Name: "Llanowar Elves", ScryfallID: "elf", TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me.ID},
			},
		})
		g.EmitEvent(game.Event{Kind: game.EventMutate, CardID: host, Source: phoenix, Actor: me.ID, Label: "Everquill Phoenix"})
		g.EmitEvent(game.Event{Kind: game.EventMutate, CardID: host, Source: phoenix, Actor: me.ID, Label: "Everquill Phoenix", Amount: 1})
	})
	for _, viewer := range allViewers(g) {
		c := fieldCardView(t, g, viewer, host)
		if len(c.MergedFrom) != 2 || c.MergedFrom[0].Name != "Everquill Phoenix" || c.MergedFrom[1].Name != "Llanowar Elves" {
			t.Fatalf("viewer %q: merged_from %+v, want the Phoenix over the Elves", viewer, c.MergedFrom)
		}
		if c.MergedFrom[1].Image != "/cards/elf/image" || c.MergedFrom[1].ManaCost != "{G}" {
			t.Errorf("viewer %q: a component carries its art and cost: %+v", viewer, c.MergedFrom[1])
		}
		lines := logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogMutate)
		if len(lines) != 2 {
			t.Fatalf("viewer %q: mutate lines %+v, want two", viewer, lines)
		}
		if lines[0].Text != "Everquill Phoenix mutated: Everquill Phoenix went on top" {
			t.Errorf("viewer %q: over text %q", viewer, lines[0].Text)
		}
		if lines[1].Text != "Everquill Phoenix mutated: Everquill Phoenix went under it" {
			t.Errorf("viewer %q: under text %q", viewer, lines[1].Text)
		}
	}
}
