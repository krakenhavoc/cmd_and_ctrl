package game

import "github.com/google/uuid"

// RandomDraw names the keyed stream used by a random effect.
type RandomDraw struct{ Player, Source uuid.UUID }

// RollDiceForEffect rolls n natural d-sides results and emits one event per
// die. It is the UNREPLACEABLE roll: no replacement sees it (#4307). A
// card's instruction to roll goes through RollDiceThenForEffect
// (dice_roll.go) instead, which a replacement of the roll can change and
// pause; this door is the starting player's roll (opening_roll.go) and the
// tests that pin the roll stream. Caller must hold g.mu.
func (g *Game) RollDiceForEffect(d RandomDraw, sides, n int) ([]int, error) {
	return g.RollDiceIgnoringLowestForEffect(d, sides, n, 0)
}

// RollDiceIgnoringLowestForEffect is "roll n dice and ignore the lowest
// `ignore` rolls" (Berserker's Frenzy's "roll two d20 and ignore the
// lower roll"; #4174). CR 706.6: an ignored roll "is considered to have
// never happened. No abilities trigger because of the ignored roll, and
// no effects apply to that roll."
//
// All n dice are drawn from the roll stream first, in order, so the
// draw count (and every later draw on the stream) is the same whether
// or not a roll is ignored and a restore replays it exactly. The lowest
// `ignore` are then dropped, and only the rest are announced: one
// EventRollDie each, in one batch, so "whenever you roll one or more
// dice" triggers once, off the kept rolls. The returned slice is the
// kept rolls in the order they were rolled.
//
// When several results are tied for the lowest, the player chooses which
// to ignore (CR 706.6). Tied results are the same number, so which one
// is dropped changes nothing anyone can see and no prompt is asked.
//
// `ignore` must be at least 0 and less than n: ignoring every roll is
// not a roll at all.
//
// Caller must hold g.mu.
func (g *Game) RollDiceIgnoringLowestForEffect(d RandomDraw, sides, n, ignore int) ([]int, error) {
	if sides <= 0 || n <= 0 || ignore < 0 || ignore >= n {
		return nil, ErrInvalidParam
	}
	r := g.randForLocked(rngStream{kind: rngStreamRoll, player: d.Player, source: d.Source})
	rolled := make([]int, n)
	for i := range rolled {
		rolled[i] = r.IntN(sides) + 1
	}
	results := rolled
	if ignore > 0 {
		results = keepAllButLowest(rolled, ignore)
	}
	batch := g.eventSeq + 1
	for _, result := range results {
		g.EmitEvent(Event{Kind: EventRollDie, Actor: d.Player, Source: d.Source, Amount: result, Sides: sides, BatchSeq: batch})
	}
	return results, nil
}

// keepAllButLowest drops the `drop` lowest values of rolled and returns
// the rest in their original order. On a tie for the last place dropped
// the earlier roll goes, which is unobservable: the values are equal.
func keepAllButLowest(rolled []int, drop int) []int {
	ignored := make([]bool, len(rolled))
	for k := 0; k < drop; k++ {
		low := -1
		for i, v := range rolled {
			if !ignored[i] && (low < 0 || v < rolled[low]) {
				low = i
			}
		}
		ignored[low] = true
	}
	kept := make([]int, 0, len(rolled)-drop)
	for i, v := range rolled {
		if !ignored[i] {
			kept = append(kept, v)
		}
	}
	return kept
}

// FlipCoinsForEffect flips n coins whose faces matter but which nobody wins
// or loses. true means heads. Caller must hold g.mu.
func (g *Game) FlipCoinsForEffect(d RandomDraw, n int) ([]bool, error) {
	if n <= 0 {
		return nil, ErrInvalidParam
	}
	r := g.randForLocked(rngStream{kind: rngStreamFlip, player: d.Player, source: d.Source})
	faces := make([]bool, n)
	for i := range faces {
		faces[i] = r.IntN(2) == 0
	}
	batch := g.eventSeq + 1
	for _, head := range faces {
		label := "tails"
		if head {
			label = "heads"
		}
		g.EmitEvent(Event{Kind: EventFlipCoin, Actor: d.Player, Source: d.Source, Label: label, BatchSeq: batch})
	}
	return faces, nil
}

// ChooseAtRandomForEffect returns up to k distinct IDs in random order.
// Caller must hold g.mu.
func (g *Game) ChooseAtRandomForEffect(d RandomDraw, ids []uuid.UUID, k int) []uuid.UUID {
	if k <= 0 {
		return nil
	}
	unique := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if !seen[id] {
			unique = append(unique, id)
			seen[id] = true
		}
	}
	return g.pickAtRandomLocked(rngStream{kind: rngStreamPick, player: d.Player, source: d.Source}, unique, k)
}

// CoinFlipSpec describes a call-and-win/loss flip. Wins is carried through a
// chain for the client and bot stop hint; Then may queue the next link.
type CoinFlipSpec struct {
	Flipper       uuid.UUID
	Source        uuid.UUID
	Question      string
	Coins         int
	AllowStop     bool
	MaxUsefulWins int
	Wins          int
	Then          func(*Game, CoinFlipResult) error
}

type CoinFlipResult struct {
	Call    string
	Faces   []bool
	Won     []bool
	Stopped bool
}

type coinFlipFrame struct{ spec CoinFlipSpec }

// FlipCoinForEffect queues a coin-call choice. The result is intentionally
// drawn only when answered, and as won/lost rather than as a face: changing a
// call after undo changes the face but cannot turn a loss into a win.
// Caller must hold g.mu.
func (g *Game) FlipCoinForEffect(spec CoinFlipSpec) uuid.UUID {
	if spec.Coins <= 0 {
		spec.Coins = 1
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind: PendingChoiceCoinCall, Chooser: spec.Flipper, FromPlayer: spec.Flipper,
		Source: spec.Source, Reason: spec.Question, CoinAllowStop: spec.AllowStop,
		CoinCount: spec.Coins, CoinMaxUsefulWins: spec.MaxUsefulWins, CoinWins: spec.Wins,
		coinFlipResume: &coinFlipFrame{spec: spec},
	})
}

// ResolveCoinCall settles a coin-call prompt. call is heads, tails, or stop
// when the prompt permits it. Caller must not hold g.mu.
func (g *Game) ResolveCoinCall(choiceID, chooserID uuid.UUID, call string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx, choice := g.findChoiceLocked(choiceID)
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	if choice.Kind != PendingChoiceCoinCall || choice.Chooser != chooserID {
		if choice.Chooser != chooserID {
			return ErrNotTheChooser
		}
		return ErrInvalidParam
	}
	if call != "heads" && call != "tails" && call != "stop" {
		return ErrInvalidParam
	}
	if call == "stop" && !choice.CoinAllowStop {
		return ErrInvalidParam
	}
	frame := choice.coinFlipResume
	g.dequeueChoiceLocked(idx)
	if frame == nil {
		return nil
	}
	result := CoinFlipResult{Call: call, Stopped: call == "stop"}
	if result.Stopped {
		result.Call = ""
	}
	if !result.Stopped {
		r := g.randForLocked(rngStream{kind: rngStreamFlip, player: frame.spec.Flipper, source: frame.spec.Source})
		result.Faces = make([]bool, frame.spec.Coins)
		result.Won = make([]bool, frame.spec.Coins)
		batch := g.eventSeq + 1
		for i := range result.Faces {
			won := r.IntN(2) == 0
			result.Won[i] = won
			head := (call == "heads") == won
			result.Faces[i] = head
			label := "tails"
			if head {
				label = "heads"
			}
			g.EmitEvent(Event{Kind: EventFlipCoin, Actor: frame.spec.Flipper, Source: frame.spec.Source, Label: label, Call: call, Won: won, BatchSeq: batch})
		}
	}
	if frame.spec.Then != nil {
		if err := frame.spec.Then(g, result); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, Actor: chooserID, Source: frame.spec.Source, ErrorMsg: err.Error()})
		}
	}
	g.runStateChecksLocked()
	return nil
}
