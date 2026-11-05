package gamestats

// winkind.go is ADR 0217 §3: how a game was won, how each seat left,
// and who dealt the last blow, all read from the full log.

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Kind is how a game was won (game_stats.win_kind) or how a seat left
// (seat_stats.out_kind). "" is none: a draw, a game with no outcome, or
// a seat that did not leave.
type Kind string

const (
	KindCombo           Kind = "combo"
	KindConcession      Kind = "concession"
	KindWinEffect       Kind = "win_effect"
	KindPoison          Kind = "poison"
	KindCommanderDamage Kind = "commander_damage"
	KindCombat          Kind = "combat"
	KindNoncombat       Kind = "noncombat"
	KindDecking         Kind = "decking"
)

// Kinds lists every kind in ADR 0217 §3's order, the order rules are
// tried in.
var Kinds = []Kind{KindCombo, KindConcession, KindWinEffect, KindPoison, KindCommanderDamage,
	KindCombat, KindNoncombat, KindDecking}

// ComboRule is which part of the combo rule fired (game_stats.combo_rule).
type ComboRule string

const (
	ComboLoop    ComboRule = "loop"
	ComboRepeats ComboRule = "repeats"
	ComboActions ComboRule = "actions"
	ComboCombats ComboRule = "combats"
)

// The combo thresholds (ADR 0217 §3). Version-2 constants: changing one
// is a Version bump and a backfill.
const (
	// ComboRepeatsMin: the winner's one source, by card name, cast,
	// activated or triggered this many times in the turn.
	ComboRepeatsMin = 10
	// ComboActionsMin: the winner's casts and activations in the turn.
	ComboActionsMin = 30
	// ComboCombatsMin: combat phases in the turn (begin_combat steps).
	ComboCombatsMin = 4
)

// LoopCallSuspected and LoopCallReplayed are the Call of the stats-only
// LogLoop entries protocol.FullLogOf adds with protocol.StatsLog.
const (
	LoopCallSuspected = protocol.LoopCallSuspected
	LoopCallReplayed  = protocol.LoopCallReplayed
)

// exit is one seat leaving the game.
type exit struct {
	seat  int
	turn  int
	batch uint64
	seq   uint64
	cause string
	// combat: the last damage or life loss before the exit was combat
	// damage.
	combat bool
	// by is the seat whose source dealt that last damage or loss; nil
	// when none, or the seat itself.
	by *int
}

// kind is the exit's own kind, the combo rule aside.
func (e exit) kind() Kind {
	switch e.cause {
	case "concede":
		return KindConcession
	case "effect":
		return KindWinEffect
	case "poison":
		return KindPoison
	case "commander_damage":
		return KindCommanderDamage
	case "life":
		if e.combat {
			return KindCombat
		}
		return KindNoncombat
	case "empty_draw":
		return KindDecking
	}
	return ""
}

// combatExit: a combat exit is a combo only by the loop or combats rule.
func (e exit) combatExit() bool {
	switch e.cause {
	case "life", "poison", "commander_damage":
		return e.combat
	}
	return false
}

// turnFacts is what the combo rule reads about one turn.
type turnFacts struct {
	loop    bool
	combats int
	// uses counts, per seat, each source's casts, activations and
	// triggers, by card name (or instance id when the name is unknown).
	uses map[int]map[string]int
	// actions counts, per seat, casts and activations.
	actions map[int]int
}

// logFacts indexes the log for the win-kind rules.
type logFacts struct {
	exits    []exit // in log order, one per seat
	gameOver *protocol.LogEvent
	turns    map[int]*turnFacts
}

func (f *logFacts) turn(t int) *turnFacts {
	tf := f.turns[t]
	if tf == nil {
		tf = &turnFacts{uses: map[int]map[string]int{}, actions: map[int]int{}}
		f.turns[t] = tf
	}
	return tf
}

// isLoopEntry: a demonstrated loop's replay (ADR 0193) or the loop
// breaker firing (ADR 0055). A replay's log carries no Call, so its one
// line is recognised by the words the engine writes ("repeats …").
func isLoopEntry(e protocol.LogEvent) bool {
	switch e.Call {
	case LoopCallSuspected, LoopCallReplayed:
		return true
	case "":
		return strings.HasPrefix(e.Label, "repeats ")
	}
	return false
}

// readLogFacts walks the log once. seatFor resolves the seat responsible
// for an entry, nameOf a card id's name.
func readLogFacts(log []protocol.LogEvent, seatFor func(protocol.LogEvent) (*Seat, int), nameOf func(string) string) *logFacts {
	f := &logFacts{turns: map[int]*turnFacts{}}
	left := map[int]bool{}
	for i, e := range log {
		switch e.Kind {
		case protocol.LogStep:
			if e.Step == "begin_combat" {
				f.turn(e.Turn).combats++
			}
		case protocol.LogLoop:
			if isLoopEntry(e) {
				f.turn(e.Turn).loop = true
			}
		case protocol.LogCast, protocol.LogActivate, protocol.LogTrigger:
			s, n := seatFor(e)
			if s == nil {
				continue
			}
			tf := f.turn(e.Turn)
			times := 1
			if e.Kind == protocol.LogTrigger {
				times = max(e.Amount, 1) // a run of one trigger is one line with a count
			} else {
				tf.actions[n]++
			}
			key := nameOf(e.CardID)
			if key == "" {
				key = e.CardID
			}
			if key == "" {
				continue
			}
			if tf.uses[n] == nil {
				tf.uses[n] = map[string]int{}
			}
			tf.uses[n][key] += times
		case protocol.LogEliminated:
			if left[e.Seat] {
				continue
			}
			left[e.Seat] = true
			x := exit{seat: e.Seat, turn: e.Turn, batch: e.Batch, seq: e.Seq, cause: e.Cause}
			x.combat, x.by = lastBlow(log[:i], e.Seat, seatFor)
			if e.Cause == "effect" && e.CardID != "" {
				// An "opponents lose" card: its controller left them.
				if _, n := seatFor(protocol.LogEvent{Seat: protocol.NoSeat, CardID: e.CardID}); n != protocol.NoSeat && n != e.Seat {
					x.by = ptr(n)
				}
			}
			f.exits = append(f.exits, x)
		case protocol.LogGameOver:
			g := e
			f.gameOver = &g
		}
	}
	return f
}

// lastBlow finds the last damage or life loss to seat in before, and
// says whether it was combat damage and whose source it was. A life
// loss that comes with a damage line in the same batch is that damage.
func lastBlow(before []protocol.LogEvent, seat int, seatFor func(protocol.LogEvent) (*Seat, int)) (combat bool, by *int) {
	for i := len(before) - 1; i >= 0; i-- {
		e := before[i]
		var hit *protocol.LogEvent
		switch {
		case e.Kind == protocol.LogDamage && e.TargetSeat != nil && *e.TargetSeat == seat:
			hit = &before[i]
		case e.Kind == protocol.LogLife && e.Seat == seat && e.Amount < 0:
			hit = &before[i]
			// The damage that caused this loss, if any, is just above it
			// in the same batch.
			for j := i - 1; j >= 0 && e.Batch != 0 && before[j].Batch == e.Batch; j-- {
				d := before[j]
				if d.Kind == protocol.LogDamage && d.TargetSeat != nil && *d.TargetSeat == seat {
					hit = &before[j]
					break
				}
			}
		default:
			continue
		}
		if _, n := seatFor(*hit); n != protocol.NoSeat && n != seat {
			by = ptr(n)
		}
		return hit.Kind == protocol.LogDamage && hit.Combat, by
	}
	return false, nil
}

// comboRule is the part of the combo rule that holds for these exits in
// turn t with actor's actions, or "". A combat exit is a combo only by
// the loop or combats rule.
func (f *logFacts) comboRule(t int, actor *int, exits []exit) ComboRule {
	tf := f.turns[t]
	if tf == nil {
		return ""
	}
	if tf.loop {
		return ComboLoop
	}
	nonCombat := len(exits) == 0
	for _, e := range exits {
		if !e.combatExit() {
			nonCombat = true
		}
	}
	if nonCombat && actor != nil {
		for _, n := range tf.uses[*actor] {
			if n >= ComboRepeatsMin {
				return ComboRepeats
			}
		}
		if tf.actions[*actor] >= ComboActionsMin {
			return ComboActions
		}
	}
	if tf.combats >= ComboCombatsMin {
		return ComboCombats
	}
	return ""
}

// classify is ADR 0217 §3's ordered list over one batch of exits in turn
// t: the first rule that matches any of them, with concession only when
// every one of them conceded.
func (f *logFacts) classify(t int, actor *int, exits []exit) (Kind, ComboRule) {
	if rule := f.comboRule(t, actor, exits); rule != "" {
		return KindCombo, rule
	}
	if len(exits) == 0 {
		return "", ""
	}
	rest := make([]exit, 0, len(exits))
	for _, e := range exits {
		if e.cause != "concede" {
			rest = append(rest, e)
		}
	}
	if len(rest) == 0 {
		return KindConcession, ""
	}
	for _, k := range Kinds[2:] {
		for _, e := range rest {
			if e.kind() == k {
				return k, ""
			}
		}
	}
	return "", ""
}

// winKind is the game's kind, its turn and the combo rule that fired.
func (f *logFacts) winKind(outcome string, winner *int) (Kind, *int, ComboRule) {
	if outcome != "win" || winner == nil {
		return "", nil, ""
	}
	if f.gameOver != nil && f.gameOver.Cause == "effect" {
		t := f.gameOver.Turn
		k, rule := f.classify(t, winner, nil)
		if k == "" {
			k = KindWinEffect
		}
		return k, ptr(t), rule
	}
	var opp []exit
	for _, e := range f.exits {
		if e.seat != *winner {
			opp = append(opp, e)
		}
	}
	if len(opp) == 0 {
		return "", nil, ""
	}
	last := opp[len(opp)-1]
	batch := []exit{last}
	if last.batch != 0 {
		batch = batch[:0]
		for _, e := range opp {
			if e.batch == last.batch {
				batch = append(batch, e)
			}
		}
	}
	k, rule := f.classify(last.turn, winner, batch)
	return k, ptr(last.turn), rule
}

// outKind is how one seat left: its exit alone, with the seat that dealt
// the last blow (or else the winner) as the combo rule's actor.
func (f *logFacts) outKind(seat int, winner *int) (Kind, *int) {
	for _, e := range f.exits {
		if e.seat != seat {
			continue
		}
		actor := e.by
		if actor == nil && winner != nil && *winner != seat {
			actor = winner
		}
		k, _ := f.classify(e.turn, actor, []exit{e})
		return k, e.by
	}
	return "", nil
}
