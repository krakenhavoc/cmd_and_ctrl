package game

import "github.com/google/uuid"

// RandomDraw names the keyed stream used by a random effect.
type RandomDraw struct{ Player, Source uuid.UUID }

// RollDiceForEffect rolls n natural d-sides results and emits one event per
// die. Caller must hold g.mu.
func (g *Game) RollDiceForEffect(d RandomDraw, sides, n int) ([]int, error) {
	if sides <= 0 || n <= 0 {
		return nil, ErrInvalidParam
	}
	r := g.randForLocked(rngStream{kind: rngStreamRoll, player: d.Player, source: d.Source})
	results := make([]int, n)
	for i := range results {
		results[i] = r.IntN(sides) + 1
	}
	batch := g.eventSeq + 1
	for _, result := range results {
		g.EmitEvent(Event{Kind: EventRollDie, Actor: d.Player, Source: d.Source, Amount: result, Sides: sides, BatchSeq: batch})
	}
	return results, nil
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
