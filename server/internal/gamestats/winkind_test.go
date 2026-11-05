package gamestats

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// winkind_test.go is ADR 0217 §3's classifier on hand-built logs.

// table is a four-seat view: Ana (seat 0) holds a creature, Bo (1) a
// burn spell and a commander, Cy (2) and Dee (3) nothing.
func table() protocol.GameView {
	return protocol.GameView{
		StartingSeat: 2,
		Battlefield: zone(
			card("ana-bear", "Grizzly Bears", ana),
			card("ana-engine", "Kiki-Jiki, Mirror Breaker", ana),
			card("bo-cmdr", "Krenko, Mob Boss", bo, commander),
		),
		Seats: []protocol.PlayerView{
			{ID: ana, Seat: 0, Life: 20, CommanderDamage: map[string]int{}},
			{ID: bo, Seat: 1, Life: 0, Eliminated: true, CommanderDamage: map[string]int{},
				Graveyard: zone(card("bo-bolt", "Lightning Bolt", bo))},
			{ID: cy, Seat: 2, Life: 0, Eliminated: true, CommanderDamage: map[string]int{"bo-cmdr": 21}},
			{ID: dee, Seat: 3, Life: 0, Eliminated: true, CommanderDamage: map[string]int{}},
		},
	}
}

func step(turn int, name string) protocol.LogEvent {
	return protocol.LogEvent{Kind: protocol.LogStep, Turn: turn, Step: name, Seat: protocol.NoSeat}
}

func out(turn, seat int, cause string, batch uint64) protocol.LogEvent {
	e := protocol.LogEvent{Kind: protocol.LogEliminated, Turn: turn, Seat: seat, Cause: cause, Batch: batch}
	if cause == "concede" {
		e.Amount = 1
	}
	return e
}

func hit(turn, from int, cardID string, to, amount int, combat bool, batch uint64) []protocol.LogEvent {
	return []protocol.LogEvent{
		{Kind: protocol.LogDamage, Turn: turn, Seat: from, CardID: cardID, TargetSeat: seatPtr(to), Amount: amount, Combat: combat, Batch: batch},
		{Kind: protocol.LogLife, Turn: turn, Seat: to, Amount: -amount, Batch: batch},
	}
}

func uses(turn, seat int, kind protocol.LogKind, cardID string, n int) []protocol.LogEvent {
	var l []protocol.LogEvent
	for i := 0; i < n; i++ {
		l = append(l, protocol.LogEvent{Kind: kind, Turn: turn, Seat: seat, CardID: cardID})
	}
	return l
}

func cat(parts ...[]protocol.LogEvent) []protocol.LogEvent {
	var l []protocol.LogEvent
	for _, p := range parts {
		l = append(l, p...)
	}
	return l
}

func won(log []protocol.LogEvent) Game {
	return Compute(Input{View: table(), Log: log, Outcome: "win", WinnerSeat: seatPtr(0)})
}

// The three opponents leave one at a time before the last exit, so each
// case is decided by the last one.
func earlyExits() []protocol.LogEvent {
	return []protocol.LogEvent{step(1, "upkeep"), out(1, 3, "concede", 1), out(2, 2, "commander_damage", 2)}
}

func TestWinKindByTheLastExit(t *testing.T) {
	for _, c := range []struct {
		name string
		log  []protocol.LogEvent
		want Kind
		rule ComboRule
	}{
		{"combat damage", cat(earlyExits(), []protocol.LogEvent{step(5, "begin_combat")},
			hit(5, 0, "ana-bear", 1, 20, true, 9), []protocol.LogEvent{out(5, 1, "life", 9)}), KindCombat, ""},
		{"burn", cat(earlyExits(), hit(5, 0, "ana-engine", 1, 3, false, 9), []protocol.LogEvent{out(5, 1, "life", 9)}),
			KindNoncombat, ""},
		{"a life payment with no damage line", cat(earlyExits(), hit(4, 0, "ana-bear", 1, 5, true, 8),
			[]protocol.LogEvent{{Kind: protocol.LogLife, Turn: 5, Seat: 1, Amount: -2, Batch: 9}, out(5, 1, "life", 9)}),
			KindNoncombat, ""},
		{"poison", cat(earlyExits(), []protocol.LogEvent{out(5, 1, "poison", 9)}), KindPoison, ""},
		{"commander damage", cat(earlyExits(), []protocol.LogEvent{out(5, 1, "commander_damage", 9)}), KindCommanderDamage, ""},
		{"decking", cat(earlyExits(), []protocol.LogEvent{out(5, 1, "empty_draw", 9)}), KindDecking, ""},
		{"an opponents-lose card", cat(earlyExits(), []protocol.LogEvent{out(5, 1, "effect", 9)}), KindWinEffect, ""},
		{"the last one concedes", cat(earlyExits(), []protocol.LogEvent{out(5, 1, "concede", 9)}), KindConcession, ""},
		{"a you-win card", cat(earlyExits(), []protocol.LogEvent{{Kind: protocol.LogGameOver, Turn: 5, Cause: "effect", CardID: "ana-engine"}}),
			KindWinEffect, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := won(c.log)
			if g.WinKind != c.want || g.ComboRule != c.rule {
				t.Errorf("win kind %q (rule %q), want %q (%q)", g.WinKind, g.ComboRule, c.want, c.rule)
			}
			if g.WinTurn == nil || *g.WinTurn != 5 {
				t.Errorf("win turn %v, want 5", g.WinTurn)
			}
		})
	}
}

// Two leave in one batch: the first kind in the list that matches any of
// them, and concession only when every one of them conceded.
func TestWinKindOfABatch(t *testing.T) {
	early := []protocol.LogEvent{step(1, "upkeep"), out(1, 3, "life", 1)}
	g := won(cat(early, hit(5, 0, "ana-bear", 1, 20, true, 9), []protocol.LogEvent{out(5, 1, "life", 9), out(5, 2, "poison", 9)}))
	if g.WinKind != KindPoison {
		t.Errorf("poison and combat in one batch: %q, want poison (earlier in the list)", g.WinKind)
	}
	g = won(cat(early, hit(5, 0, "ana-bear", 1, 20, true, 9), []protocol.LogEvent{out(5, 1, "life", 9), out(5, 2, "concede", 9)}))
	if g.WinKind != KindCombat {
		t.Errorf("one killed in combat and one conceding: %q, want combat", g.WinKind)
	}
	g = won(cat(early, []protocol.LogEvent{out(5, 1, "concede", 9), out(5, 2, "concede", 9)}))
	if g.WinKind != KindConcession {
		t.Errorf("both conceding: %q, want concession", g.WinKind)
	}
}

func TestComboRule(t *testing.T) {
	lastBurn := func(turn int) []protocol.LogEvent {
		return cat(hit(turn, 0, "ana-engine", 1, 3, false, 9), []protocol.LogEvent{out(turn, 1, "life", 9)})
	}
	lastSwing := func(turn int) []protocol.LogEvent {
		return cat(hit(turn, 0, "ana-bear", 1, 20, true, 9), []protocol.LogEvent{out(turn, 1, "life", 9)})
	}
	combats := func(turn, n int) []protocol.LogEvent {
		var l []protocol.LogEvent
		for i := 0; i < n; i++ {
			l = append(l, step(turn, "begin_combat"), step(turn, "end_combat"))
		}
		return l
	}
	for _, c := range []struct {
		name string
		log  []protocol.LogEvent
		want Kind
		rule ComboRule
	}{
		{"one source ten times, then burn", cat(earlyExits(), uses(5, 0, protocol.LogActivate, "ana-engine", 10), lastBurn(5)),
			KindCombo, ComboRepeats},
		{"one source nine times is not enough", cat(earlyExits(), uses(5, 0, protocol.LogActivate, "ana-engine", 9), lastBurn(5)),
			KindNoncombat, ""},
		{"a trigger line with a count", cat(earlyExits(), []protocol.LogEvent{{Kind: protocol.LogTrigger, Turn: 5, Seat: 0, CardID: "ana-engine", Amount: 12}}, lastBurn(5)),
			KindCombo, ComboRepeats},
		{"thirty casts and activations", cat(earlyExits(), uses(5, 0, protocol.LogCast, "ana-bear", 9), uses(5, 0, protocol.LogActivate, "x1", 7),
			uses(5, 0, protocol.LogActivate, "x2", 7), uses(5, 0, protocol.LogCast, "x3", 7), lastBurn(5)), KindCombo, ComboActions},
		{"the repeats were on another turn", cat(earlyExits(), uses(4, 0, protocol.LogActivate, "ana-engine", 12), lastBurn(5)),
			KindNoncombat, ""},
		{"someone else's repeats", cat(earlyExits(), uses(5, 1, protocol.LogActivate, "bo-cmdr", 12), lastBurn(5)),
			KindNoncombat, ""},
		{"a big attack after many spells stays combat", cat(earlyExits(), uses(5, 0, protocol.LogActivate, "ana-engine", 12), lastSwing(5)),
			KindCombat, ""},
		{"four combats", cat(earlyExits(), combats(5, 4), lastSwing(5)), KindCombo, ComboCombats},
		{"three combats is an extra-combat card", cat(earlyExits(), combats(5, 3), lastSwing(5)), KindCombat, ""},
		{"a demonstrated loop's replay", cat(earlyExits(), []protocol.LogEvent{{Kind: protocol.LogLoop, Turn: 5, Seat: 0, Label: "repeats Kiki-Jiki ×300: Ana: library -300"}}, lastSwing(5)),
			KindCombo, ComboLoop},
		{"the loop breaker", cat(earlyExits(), []protocol.LogEvent{{Kind: protocol.LogLoop, Turn: 5, Seat: 0, Call: LoopCallSuspected}}, lastBurn(5)),
			KindCombo, ComboLoop},
		{"a loop recording that was discarded is no loop", cat(earlyExits(), []protocol.LogEvent{{Kind: protocol.LogLoop, Turn: 5, Seat: 0, Label: "discards the loop recording"}}, lastBurn(5)),
			KindNoncombat, ""},
		{"a concession to a combo", cat(earlyExits(), uses(5, 0, protocol.LogCast, "ana-engine", 31), []protocol.LogEvent{out(5, 1, "concede", 9)}),
			KindCombo, ComboRepeats},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := won(c.log)
			if g.WinKind != c.want || g.ComboRule != c.rule {
				t.Errorf("win kind %q (rule %q), want %q (%q)", g.WinKind, g.ComboRule, c.want, c.rule)
			}
		})
	}
}

func TestNoWinKindWithoutAWinner(t *testing.T) {
	log := cat(earlyExits(), []protocol.LogEvent{out(5, 1, "life", 9)})
	for _, outcome := range []string{"draw", ""} {
		g := Compute(Input{View: table(), Log: log, Outcome: outcome})
		if g.WinKind != "" || g.WinTurn != nil || g.ComboRule != "" {
			t.Errorf("outcome %q: win kind %q turn %v rule %q, want none", outcome, g.WinKind, g.WinTurn, g.ComboRule)
		}
	}
}

func TestOutKindOutBySeatAndTurnPosition(t *testing.T) {
	log := cat(
		[]protocol.LogEvent{step(1, "upkeep"), out(1, 3, "concede", 1)},
		// Bo's commander deals Cy the last blow, in combat.
		hit(2, protocol.NoSeat, "bo-cmdr", 2, 7, true, 2), []protocol.LogEvent{out(2, 2, "commander_damage", 2)},
		hit(5, 0, "ana-bear", 1, 20, true, 9), []protocol.LogEvent{out(5, 1, "life", 9)},
	)
	g := won(log)
	want := []struct {
		kind Kind
		by   *int
		pos  int
	}{
		{"", nil, 3},                         // Ana: still standing
		{KindCombat, seatPtr(0), 4},          // Bo: Ana's Bears
		{KindCommanderDamage, seatPtr(1), 1}, // Cy: Bo's commander, through its controller
		{KindConcession, nil, 2},             // Dee: no blow before conceding
	}
	for i, w := range want {
		s := g.Seats[i]
		if s.OutKind != w.kind {
			t.Errorf("seat %d out kind %q, want %q", i, s.OutKind, w.kind)
		}
		switch {
		case w.by == nil && s.OutBySeat != nil:
			t.Errorf("seat %d out by %d, want none", i, *s.OutBySeat)
		case w.by != nil && (s.OutBySeat == nil || *s.OutBySeat != *w.by):
			t.Errorf("seat %d out by %v, want %d", i, s.OutBySeat, *w.by)
		}
		if s.TurnPosition == nil || *s.TurnPosition != w.pos {
			t.Errorf("seat %d turn position %v, want %d (seat 2 started)", i, s.TurnPosition, w.pos)
		}
	}
	// The spread: Ana hit only Bo, and Bo's commander dealt Cy 21.
	if tg := g.Seats[0].Targets; len(tg) != 3 || tg[0].TargetSeat != 1 || tg[0].Damage != 20 || tg[0].CombatDamage != 20 || tg[1].Damage != 0 {
		t.Errorf("Ana's targets = %+v", tg)
	}
	if tg := g.Seats[1].Targets; tg[1].TargetSeat != 2 || tg[1].CommanderDamage != 21 || tg[1].Damage != 7 {
		t.Errorf("Bo's targets = %+v", tg)
	}
	if g.Seats[1].LifeLostToCombat != 20 || g.Seats[2].LifeLostToCombat != 7 {
		t.Errorf("life lost to combat: Bo %d, Cy %d; want 20 and 7", g.Seats[1].LifeLostToCombat, g.Seats[2].LifeLostToCombat)
	}
}
