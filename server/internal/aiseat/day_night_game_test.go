package aiseat_test

import (
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// day_night_game_test.go is the whole-game half of #2561 (ADR 0132): a
// bot with The Celestus on the battlefield is OFFERED the sorcery-speed
// toggle by the enumerator, the heuristic prices it above passing, the
// dispatcher accepts it, and the table survives the flip trigger it
// puts on the stack (the life, and the optional loot the bot answers).
//
// The legal-package test pins the offer and the heuristic package the
// price; what only a game can say is that the whole loop turns over
// without stalling on the prompt the trigger opens.

const oracleCelestusBot = "c0ad2b5f-066b-424b-bddf-d3014731e599"

func TestABotWithTheCelestusTogglesDayAndNight(t *testing.T) {
	requireGameTests(t)

	room := newGraveyardRoom(t, 25610)
	g := room.Game
	bot := g.Seats[0]

	celestus := seedBattlefield(g, bot, game.Card{
		Name: "The Celestus", TypeLine: "Legendary Artifact", OracleID: oracleCelestusBot,
	})
	// Four lands: {3} for the toggle and spare mana so the cost is
	// never what stops it.
	for i := 0; i < 4; i++ {
		seedBattlefield(g, bot, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationDay })
	startLife := bot.Life

	res := playGameIn(t, room, 25610, []aiseat.Policy{heuristic.New(), passPolicy{}}, 6, 120*time.Second)

	activations, flips := 0, 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			switch {
			case ev.Kind == game.EventActivateAbility && ev.Source == celestus:
				activations++
			case game.DayNightFlipped(ev):
				flips++
			}
		}
	})
	if activations == 0 {
		t.Fatalf("the bot never toggled day and night in %d turns (%d flips in all)", res.turns, flips)
	}
	if flips < activations {
		t.Errorf("%d toggles but only %d flips: a toggle always flips a game that has a designation", activations, flips)
	}
	// Every flip puts the Celestus's ability on the stack: "you gain 1
	// life", then the optional loot the bot has to answer. Life is the
	// proof it resolved rather than stalled behind its prompt.
	if bot.Life <= startLife {
		t.Errorf("life %d -> %d across %d flips: the Celestus's ability never resolved", startLife, bot.Life, flips)
	}
	if got := res.totals().Rejected; got != 0 {
		t.Errorf("%d moves were rejected: the enumerator offered something the dispatcher refused", got)
	}
}
