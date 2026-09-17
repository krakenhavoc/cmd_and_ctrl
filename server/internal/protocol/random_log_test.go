package protocol

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestRandomLogGroupsInterleavedBatchesAndRedactsSource(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	source := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Secret source", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
		g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: me.ID, Source: source, Amount: 2, Sides: 6, BatchSeq: 100})
		g.EmitEvent(game.Event{Kind: game.EventChangeLife, Target: me.ID, Amount: 1})
		g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: me.ID, Source: source, Amount: 5, Sides: 6, BatchSeq: 100})
		g.EmitEvent(game.Event{Kind: game.EventRollDie, Actor: me.ID, Source: source, Amount: 4, Sides: 20, BatchSeq: 101})
		g.EmitEvent(game.Event{Kind: game.EventFlipCoin, Actor: me.ID, Source: source, Label: "heads", Call: "heads", Won: true, BatchSeq: 102})
		g.EmitEvent(game.Event{Kind: game.EventFlipCoin, Actor: me.ID, Source: source, Label: "tails", Call: "heads", Won: false, BatchSeq: 102})
	})
	view := ViewOfGame(g)
	roll := findLog(t, view.Log, LogRoll)
	if !reflect.DeepEqual(roll.Results, []int{2, 5}) || roll.Sides != 6 {
		t.Fatalf("roll = %+v", roll)
	}
	if !strings.Contains(roll.Text, "2d6") || !strings.Contains(roll.Text, "Secret source") {
		t.Fatal(roll.Text)
	}
	flip := findLog(t, view.Log, LogFlip)
	if !reflect.DeepEqual(flip.Faces, []string{"heads", "tails"}) || flip.Wins != 1 || flip.Call != "heads" {
		t.Fatalf("flip = %+v", flip)
	}
	filtered := FilterViewFor(view, g.Seats[1].ID.String())
	for _, entry := range filtered.Log {
		if strings.Contains(entry.Text, "Secret source") {
			t.Fatalf("hidden source leaked: %s", entry.Text)
		}
	}
	if len(findLog(t, filtered.Log, LogRoll).Results) != 2 {
		t.Fatal("public results redacted")
	}
}

func TestCoinCallViewAndRandomWireNeverExposeRNG(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.FlipCoinForEffect(game.CoinFlipSpec{Flipper: me.ID, Coins: 5, AllowStop: true, Wins: 2, MaxUsefulWins: 3})
	})
	view := ViewOfGame(g)
	if len(view.PendingChoices) != 1 {
		t.Fatalf("choices = %v", view.PendingChoices)
	}
	ch := view.PendingChoices[0]
	if ch.Kind != "coin_call" || ch.Coins != 5 || !ch.AllowStop || ch.Wins != 2 || ch.MaxUsefulWins != 3 {
		t.Fatalf("choice = %+v", ch)
	}
	if err := g.ResolveCoinCall(uuid.MustParse(ch.ID), me.ID, "heads"); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(ViewOfGame(g))
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"rngKey", "rngCounters", "sourceOrdinals", "coinFlipResume"} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("wire contains %s", secret)
		}
	}
	flip := findLog(t, ViewOfGame(g).Log, LogFlip)
	if len(flip.Faces) != 5 {
		t.Fatalf("flip batch = %+v", flip)
	}
}
