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
// or loses. true means heads. It is FlipCoinsPreferringForEffect with heads
// wanted on every coin, which is the face every caller that does not say
// otherwise is better off with (Ral Zarek's extra turns). Caller must hold
// g.mu.
func (g *Game) FlipCoinsForEffect(d RandomDraw, n int) ([]bool, error) {
	if n <= 0 {
		return nil, ErrInvalidParam
	}
	prefer := make([]bool, n)
	for i := range prefer {
		prefer[i] = true
	}
	return g.FlipCoinsPreferringForEffect(d, prefer)
}

// FlipCoinsPreferringForEffect flips one uncalled coin per entry of
// `prefer` (CR 705.2: only the face matters) and returns the faces, true
// for heads. `prefer` is the face the flipper wants for that coin, and it
// is read only when a FlipTwoIgnoreOne rule (Krark's Thumb, #6152) has the
// player flip several coins for one and keep one: they keep that face when
// any of the coins shows it. Rakdos, the Showstopper wants its controller's
// own creatures heads and everyone else's tails. Caller must hold g.mu.
func (g *Game) FlipCoinsPreferringForEffect(d RandomDraw, prefer []bool) ([]bool, error) {
	n := len(prefer)
	if n <= 0 {
		return nil, ErrInvalidParam
	}
	r := g.randForLocked(rngStream{kind: rngStreamFlip, player: d.Player, source: d.Source})
	per := g.coinsPerFlipLocked(d.Player)
	faces := make([]bool, n)
	for i := range faces {
		// All the coins are drawn, in order, whichever is kept, so the
		// stream advances the same way however the keep falls.
		kept := r.IntN(2) == 0
		for j := 1; j < per; j++ {
			if face := r.IntN(2) == 0; face == prefer[i] {
				kept = face
			}
		}
		faces[i] = kept
	}
	// A coin-flip rule (Edgar's Two-Headed Coin) makes the coins heads and
	// won. Read before this flip's events exist, so "the first time each
	// turn" sees an earlier flip only.
	forced := g.coinsComeUpHeadsLocked(d.Player)
	batch := g.eventSeq + 1
	for i := range faces {
		if forced {
			faces[i] = true
		}
		label := "tails"
		if faces[i] {
			label = "heads"
		}
		g.EmitEvent(Event{Kind: EventFlipCoin, Actor: d.Player, Source: d.Source, Label: label, Won: forced, BatchSeq: batch})
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

// ChooseIndexAtRandomForEffect is "chooses one of their piles at random"
// (Whims of the Fates, #6248): an index in [0, n), drawn from the same
// keyed, persisted pick stream ChooseAtRandomForEffect uses, so an undo
// rewinds it. -1 when n <= 0. Caller must hold g.mu.
func (g *Game) ChooseIndexAtRandomForEffect(d RandomDraw, n int) int {
	if n <= 0 {
		return -1
	}
	return g.randForLocked(rngStream{kind: rngStreamPick, player: d.Player, source: d.Source}).IntN(n)
}

// CoinFlipSpec describes a call-and-win/loss flip. Wins is carried through a
// chain for the client and bot stop hint; ThenKey may queue the next link.
type CoinFlipSpec struct {
	Flipper       uuid.UUID
	Source        uuid.UUID
	Question      string
	Coins         int
	AllowStop     bool
	MaxUsefulWins int
	Wins          int

	// ThenKey is the rest of the card as a registered continuation
	// (RegisterCoinFlipThen, #4253 slice 10b), and Carry, Nums, Words and
	// Asked the plain data it is handed (CoinFlipped). A call holds no
	// closure, so a table waiting on it is a restore point. Zero is a flip
	// with nothing hanging off it. The closure form is gone (#4361).
	ThenKey CoinFlipThen
	Carry   []uuid.UUID
	Nums    []int
	Words   []string
	Asked   *AskedItem
}

type CoinFlipResult struct {
	Call    string
	Faces   []bool
	Won     []bool
	Stopped bool
}

// FlipCoinForEffect queues a coin-call choice. The result is intentionally
// drawn only when answered, and as won/lost rather than as a face: changing a
// call after undo changes the face but cannot turn a loss into a win.
// Caller must hold g.mu.
func (g *Game) FlipCoinForEffect(spec CoinFlipSpec) uuid.UUID {
	if spec.Coins <= 0 {
		spec.Coins = 1
	}
	c := PendingChoice{
		Kind: PendingChoiceCoinCall, Chooser: spec.Flipper, FromPlayer: spec.Flipper,
		Source: spec.Source, Reason: spec.Question, CoinAllowStop: spec.AllowStop,
		CoinCount: spec.Coins, CoinMaxUsefulWins: spec.MaxUsefulWins, CoinWins: spec.Wins,
	}
	// #4253 slice 10b: a keyed call, or one with nothing hanging off it, is
	// data (coin_flip_keyed.go).
	spec.keyedCoinFlip(&c)
	return g.QueueChoiceForEffect(c)
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
	spec, ok := coinFlipSpecOf(choice)
	g.dequeueChoiceLocked(idx)
	if !ok {
		return nil
	}
	result := CoinFlipResult{Call: call, Stopped: call == "stop"}
	if result.Stopped {
		result.Call = ""
	}
	if !result.Stopped {
		r := g.randForLocked(rngStream{kind: rngStreamFlip, player: spec.Flipper, source: spec.Source})
		result.Faces = make([]bool, spec.Coins)
		result.Won = make([]bool, spec.Coins)
		forced := g.coinsComeUpHeadsLocked(spec.Flipper)
		batch := g.eventSeq + 1
		// Krark's Thumb (#6152): each coin is several, and the player
		// keeps a won one when any is won (CR 705.2). All are drawn, in
		// order, so the stream advances the same way whatever is kept.
		per := g.coinsPerFlipLocked(spec.Flipper)
		for i := range result.Faces {
			won := false
			for j := 0; j < per; j++ {
				if r.IntN(2) == 0 {
					won = true
				}
			}
			head := (call == "heads") == won
			if forced {
				won, head = true, true
			}
			result.Won[i] = won
			result.Faces[i] = head
			label := "tails"
			if head {
				label = "heads"
			}
			g.EmitEvent(Event{Kind: EventFlipCoin, Actor: spec.Flipper, Source: spec.Source, Label: label, Call: call, Won: won, BatchSeq: batch})
		}
	}
	if err := g.runCoinFlipThen(spec, result); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, Actor: chooserID, Source: spec.Source, ErrorMsg: err.Error()})
	}
	g.runStateChecksLocked()
	return nil
}
