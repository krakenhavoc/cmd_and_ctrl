package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_drain_test.go — ADR 0018's #3013 amendment,
// sub-PR 4: EachOpponentLosesLife is a footprinted step, so the drain
// cards written with it join the "Ask only when the order matters"
// check.

const (
	corpseKnightOracle   = "0c77b805-728d-4ae4-88a9-223f4b7b52e0"
	waywardServantOracle = "2916f041-f080-450f-ae79-dc09076d971a"
)

func TestTheRegistryDeclaresTheDrainsFootprints(t *testing.T) {
	loses := game.FootprintStep{Kind: game.FootprintLoseLifeEachOpponent, N: 1}
	gains := game.FootprintStep{Kind: game.FootprintGainLife, N: 1}
	for _, c := range []struct {
		name   string
		oracle string
		want   []game.FootprintStep
	}{
		{"Corpse Knight", corpseKnightOracle, []game.FootprintStep{loses}},
		{"Wayward Servant", waywardServantOracle, []game.FootprintStep{loses, gains}},
		{"Zulaport Cutthroat", zulaportOracle, []game.FootprintStep{loses, gains}},
	} {
		rows := game.CatalogTriggers(c.oracle)
		if len(rows) != 1 {
			t.Fatalf("%s: %d rows", c.name, len(rows))
		}
		got := rows[0].Footprint
		if len(got) != len(c.want) {
			t.Errorf("%s: footprint %+v, want %+v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i].Kind != c.want[i].Kind || got[i].N != c.want[i].N {
				t.Errorf("%s: step %d is %+v, want %+v", c.name, i, got[i], c.want[i])
			}
		}
	}
}

// drainBoard seats Corpse Knight and Wayward Servant for me.
func drainBoard(g *game.Game, me *game.Player) {
	pushCatalogPermanent(g, me.ID, "Corpse Knight", "Creature — Zombie Knight", corpseKnightOracle, false)
	pushCatalogPermanent(g, me.ID, "Wayward Servant", "Creature — Zombie", waywardServantOracle, false)
}

// A Zombie cast beside both drains triggers each of them; every order
// gives the same game, so the batch goes on with no prompt.
func TestTwoDrainsOnOneEntryDoNotAsk(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drainBoard(g, me)
	life, oppLife := me.Life, opp.Life

	b20CastCreature(t, g, me, "Walking Corpse", "Creature — Zombie", "", 2, 2)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("two drains asked for an order: %+v", ch)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d → %d, want -2", oppLife, opp.Life)
	}
}

func TestTwoDrainsStillAskWhenTheOrderCanShow(t *testing.T) {
	t.Run("a lifegain payoff watches the gain", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		drainBoard(g, me)
		pushCatalogPermanent(g, me.ID, "Ajani's Pridemate", "Creature — Cat Soldier", b24AjanisPridemateOracle, false)
		before := promptsByReason(t, game.TriggerOrderReasonReactor)
		b20CastCreature(t, g, me, "Walking Corpse", "Creature — Zombie", "", 2, 2)
		if ch := settleUntilOrderPrompt(t, g); ch == nil {
			t.Fatal("a lifegain payoff on the board must keep the prompt")
		}
		if got := promptsByReason(t, game.TriggerOrderReasonReactor) - before; got != 1 {
			t.Errorf("counted %v under reactor, want 1", got)
		}
	})
}
