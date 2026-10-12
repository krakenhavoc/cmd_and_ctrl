package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// attrition_commander_test.go pins #2474's board model: a commander of
// mine that dies in an exchange and is cast again (CR 903.9a, 903.8),
// and the turn the line waits for it to lose summoning sickness.

// commanderTradeState is the full seed-1409 Wurm mirror's last board,
// as the attrition search sees it: my 3/3 commander ("mine") and a
// Bear against their 3/3 commander ("theirs"), me on myLife, and two
// lands of mine, enough to cast a {1}{R} commander once more. extra is
// added to the defender.
func commanderTradeState(myLife int, extra ...protocol.CardView) (*state, *SeatEval) {
	const me, def = "me", "def"
	cmdr := func(id, seat string) protocol.CardView {
		return protocol.CardView{InstanceID: id, Owner: seat, Controller: seat, TypeLine: "Legendary Creature — Bear",
			ManaCost: "{1}{R}", Power: 3, Toughness: 3, IsCommander: true}
	}
	cards := []protocol.CardView{
		cmdr("mine", me),
		{InstanceID: "bear", Owner: me, Controller: me, TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		cmdr("theirs", def),
		{InstanceID: "l1", Owner: me, Controller: me, TypeLine: "Basic Land — Mountain"},
		{InstanceID: "l2", Owner: me, Controller: me, TypeLine: "Basic Land — Mountain"},
	}
	cards = append(cards, extra...)
	view := &protocol.GameView{Seats: []protocol.PlayerView{{ID: me}, {ID: def}}}
	view.Battlefield.Cards = cards
	d := &SeatEval{ID: def, Life: 3}
	return &state{w: DefaultWeights(), me: me, view: view, myEval: &SeatEval{ID: me, Life: myLife}, opps: []*SeatEval{d}}, d
}

func cardIn(st *state, id string) *protocol.CardView {
	for i := range st.view.Battlefield.Cards {
		if st.view.Battlefield.Cards[i].InstanceID == id {
			return &st.view.Battlefield.Cards[i]
		}
	}
	return nil
}

// TestAfterExchangeCastsMyCommanderAgain: the trade kills my commander,
// and I cast it again. It is back on the next board, summoning sick, so
// the line can block with it and count it for a race's next turn but
// not attack with it; its cast count goes up by one, so a second trade
// in the same line needs {2} more than my two lands make. The board the
// search started from is not touched.
func TestAfterExchangeCastsMyCommanderAgain(t *testing.T) {
	st, def := commanderTradeState(6)
	o := outcome{mineDead: map[string]bool{"mine": true}, mineBack: map[string]bool{"mine": true}, theirDead: map[string]bool{}}
	next, nd := st.afterExchange(def, o, exchange{sure: 2, crack: 3})

	c := cardIn(next, "mine")
	if c == nil || !c.SummoningSick {
		t.Fatalf("my commander on the next board = %+v; want it back, summoning sick", c)
	}
	if nd.Life != 1 || next.myEval.Life != 3 {
		t.Errorf("lives = %d / %d, want the defender on 1 and me on 3", nd.Life, next.myEval.Life)
	}
	if got := next.view.Seats[0].CommanderCasts["mine"]; got != 1 {
		t.Errorf("cast count = %d, want 1", got)
	}
	if next.canRecast("me", c) {
		t.Error("two lands pay for a second cast; the tax is {2} more after the first")
	}
	if !next.waitingOn() {
		t.Error("waitingOn = false with my commander summoning sick")
	}
	for _, a := range simAttackers(next, nd) {
		if a.InstanceID == "mine" {
			t.Error("simAttackers offers the summoning-sick commander")
		}
	}
	if orig := cardIn(st, "mine"); orig.SummoningSick || st.view.Seats[0].CommanderCasts["mine"] != 0 {
		t.Error("afterExchange changed the board it started from")
	}
}

// TestAfterWait: on the turn the line waits, every opponent swings
// everything into every creature I have. Their commander is held by
// mine or the Bear, so I lose nothing and my commander can attack
// after. A 4/1 menace Raider is held by nothing (the crack-back counts
// menace as unblockable), and on 4 life or less that kills me.
func TestAfterWait(t *testing.T) {
	st, def := commanderTradeState(6)
	cardIn(st, "mine").SummoningSick = true
	next, ok := st.afterWait()
	if !ok {
		t.Fatal("afterWait refused a turn their commander cannot hurt me on")
	}
	if next.myEval.Life != 6 || next.waitingOn() || len(simAttackers(next, def)) != 2 {
		t.Errorf("after the wait: life %d, waiting %v, %d attackers; want 6, false, 2",
			next.myEval.Life, next.waitingOn(), len(simAttackers(next, def)))
	}
	if !cardIn(st, "mine").SummoningSick {
		t.Error("afterWait changed the board it started from")
	}

	raider := protocol.CardView{InstanceID: "raider", Owner: "def", Controller: "def", TypeLine: "Creature — Orc",
		Power: 4, Toughness: 1, Abilities: []string{"menace"}}
	for _, tc := range []struct {
		life int
		ok   bool
	}{{5, true}, {4, false}} {
		st, _ := commanderTradeState(tc.life, raider)
		cardIn(st, "mine").SummoningSick = true
		next, ok := st.afterWait()
		if ok != tc.ok {
			t.Errorf("on %d life: afterWait ok = %v, want %v", tc.life, ok, tc.ok)
		}
		if ok && next.myEval.Life != tc.life-4 {
			t.Errorf("on %d life: %d after the wait, want %d", tc.life, next.myEval.Life, tc.life-4)
		}
	}
}

// TestOutcomeOfTheCommanderTrade reads the outcome the line plays
// forward on the issue's board, swinging both: blockToSurvive's model
// only guesses (it blocks the Bear with their commander, which lives,
// and has nothing left for mine). With Config.HorizonPlaysFreeAnswer
// the line plays the free answer, commander for commander, and in the
// pass that casts my commander again it is not lost. Only both together
// make the exchange acceptable.
func TestOutcomeOfTheCommanderTrade(t *testing.T) {
	for _, tc := range []struct {
		name               string
		free, recast       bool
		mineDead, mineBack int
	}{
		{"both off: the guess, both of mine dead", false, false, 2, 0},
		{"the free answer", true, false, 1, 0},
		{"the guess, my commander cast again", false, true, 2, 1},
		{"both on", true, true, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st, def := commanderTradeState(6)
			st.bf = map[string]*protocol.CardView{}
			for i := range st.view.Battlefield.Cards {
				st.bf[st.view.Battlefield.Cards[i].InstanceID] = &st.view.Battlefield.Cards[i]
			}
			cfg := BaselineConfig()
			cfg.HorizonPlaysFreeAnswer = tc.free
			p := NewWithConfig(cfg)
			swing := []*protocol.CardView{cardIn(st, "mine"), cardIn(st, "bear")}
			e := p.raceNumbers(st, def, swing, defenderBlockers(st, def.ID))
			if !e.d.guessed {
				t.Fatal("blockToSurvive found the blocks; the board no longer pins the guess")
			}
			o := p.outcomeOf(st, def, e, tc.recast)
			if len(o.mineDead) != tc.mineDead || len(o.mineBack) != tc.mineBack || len(o.theirDead) != 0 {
				t.Errorf("outcome %+v; want %d of mine dead, %d cast again, none of theirs", o, tc.mineDead, tc.mineBack)
			}
			x := st.exchangeOf(def, e, o)
			if want := tc.free && tc.recast; x.acceptable(st) != want {
				t.Errorf("exchange %+v acceptable = %v, want %v", x, !want, want)
			}
		})
	}
}
