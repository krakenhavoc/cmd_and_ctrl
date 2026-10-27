package game

import (
	"github.com/google/uuid"
)

// pile_split_keyed.go — #4253 slice 10b: a pile split (Fact or Fiction's
// "an opponent separates those cards into two piles", and the face-down
// split of #2147) whose two links and the rest of the card are data, so a
// table waiting on the split, or on the pick of a pile, is a restore point.
//
// The closure form chains a reveal pick into an option pick through two
// closures and hands the card's Then the two piles. The keyed form is two
// registered links and the card's tail:
//
//   - the split is a keyed reveal pick (pileSplitSplit) carrying the
//     chooser (Carry), the face-down flag (PickNums), the pick's header
//     (PickMatchArg) and the card's tail (pickTail);
//   - the pick is a keyed option pick (pileSplitTake) carrying the two
//     piles, first option's first (Carry), the first option's size
//     (PickNums) and the card's tail;
//   - the card's tail is told the pile taken and then the pile left, as
//     its IDs, with how many of them were taken as its Amount
//     (PileSplitPrompt.ThenTail).

// pileSplitSplit is the keyed answer to the split: the splitter's cards are
// the first pile (or, face down, the face-down pile), and the chooser is
// asked which to take. An on-disk key: never renamed, never reused.
// Registered from init: the pick names the next link.
var (
	pileSplitSplit ChooseCardsThen
	pileSplitTake  OptionPickThen
)

func init() {
	pileSplitSplit = RegisterChooseCardsThen("choose-cards/pile-split", pileSplitSplitAnswered)
	pileSplitTake = RegisterOptionPickThen("option-pick/pile-split-take", pileSplitTakeAnswered)
}

func pileSplitSplitAnswered(g *Game, r CardsChosen) error {
	chooser := r.Chooser
	if len(r.Carry) > 0 {
		chooser = r.Carry[0]
	}
	faceDown := len(r.Nums) > 0 && r.Nums[0] == 1
	return g.queuePilePickKeyedLocked(chooser, r.From, r.Source, r.Word, r.Candidates, r.Picked, faceDown, r.Tail)
}

func pileSplitTakeAnswered(g *Game, r OptionPicked) error {
	n := 0
	if len(r.Nums) > 0 {
		n = r.Nums[0]
	}
	if n > len(r.Carry) {
		n = len(r.Carry)
	}
	first, second := r.Carry[:n], r.Carry[n:]
	taken, left := first, second
	if r.Option != nil && sameUUIDSet(r.Option.Cards, second) && !sameUUIDSet(r.Option.Cards, first) {
		taken, left = second, first
	}
	return g.runPileTailLocked(r.Tail, taken, left)
}

// sameUUIDSet reports whether a and b name the same IDs, in any order.
func sameUUIDSet(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	in := make(map[uuid.UUID]bool, len(a))
	for _, id := range a {
		in[id] = true
	}
	for _, id := range b {
		if !in[id] {
			return false
		}
	}
	return true
}

// runPileTailLocked tells the card's tail the pile taken and the pile left.
//
// Caller must hold g.mu.
func (g *Game) runPileTailLocked(t KeyedTail, taken, left []uuid.UUID) error {
	ids := append(copyUUIDs(taken), left...)
	return t.RunTold(g, ids, len(taken))
}

// queuePileSplitKeyedLocked is QueuePileSplitForEffect for a prompt with a
// keyed tail and no closure.
//
// Caller must hold g.mu.
func (g *Game) queuePileSplitKeyedLocked(p PileSplitPrompt) {
	cards := copyUUIDs(p.Cards)
	tail := copyKeyedTail(p.ThenTail)
	report := func(err error) {
		if err != nil {
			g.emitChoiceEffectErrorLocked(p.Chooser, p.Source, err)
		}
	}
	if len(cards) == 0 {
		report(g.runPileTailLocked(tail, nil, nil))
		return
	}
	owner := p.Owner
	if owner == uuid.Nil {
		owner = p.Chooser
	}
	if who := g.playerByIDLocked(p.Splitter); who == nil || who.Eliminated {
		// The splitter has left (CR 800.4a): the whole reveal is one pile.
		report(g.runPileTailLocked(tail, cards, nil))
		return
	}
	faceDown := 0
	if p.FaceDown {
		faceDown = 1
	}
	if g.QueueChoiceForEffect(PendingChoice{
		Kind:         PendingChoiceRevealPick,
		Chooser:      p.Splitter,
		FromPlayer:   owner,
		Count:        len(cards),
		Source:       p.Source,
		Reason:       p.SplitQuestion,
		ChooseCards:  cards,
		ChooseMin:    0,
		ChooseMax:    len(cards),
		PickThen:     pileSplitSplit.key,
		OptionCarry:  []uuid.UUID{p.Chooser},
		PickNums:     []int{faceDown},
		PickMatchArg: p.PickQuestion,
		pickTail:     tail,
	}) != uuid.Nil {
		return
	}
	// Nobody could be asked to split: an empty first pile, as the closure
	// form's run settles it.
	report(g.queuePilePickKeyedLocked(p.Chooser, owner, p.Source, p.PickQuestion, cards, nil, p.FaceDown, tail))
}

// queuePilePickKeyedLocked is the keyed second link: the chooser takes one
// of the two piles. `first` is the splitter's pick (the face-down pile of a
// face-down split, whose face-up pile is revealed here first).
//
// Caller must hold g.mu.
func (g *Game) queuePilePickKeyedLocked(chooser, owner, source uuid.UUID, question string, all, first []uuid.UUID, faceDown bool, tail KeyedTail) error {
	one, two := partitionPiles(all, first)
	var options []ChoiceOption
	if faceDown {
		down, up := one, two
		g.RevealForEffect(RevealSpec{Player: owner, Source: source, Reason: "face-up pile", Cards: up})
		one, two = up, down
		options = []ChoiceOption{
			{Label: facePileLabel("face-up", up), Cards: up},
			{Label: facePileLabel("face-down", down), Cards: down},
		}
	} else {
		options = []ChoiceOption{
			{Label: pileLabel(1, one), Cards: one},
			{Label: pileLabel(2, two), Cards: two},
		}
	}
	queued := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:    chooser,
		FromPlayer: owner,
		Source:     source,
		Question:   question,
		Options:    options,
		ThenKey:    pileSplitTake,
		Carry:      append(copyUUIDs(one), two...),
		Nums:       []int{len(one)},
		Tail:       tail,
	})
	if queued != uuid.Nil {
		return nil
	}
	// The chooser has left between the split and the pick: the first
	// option's pile, so the rest of the card still resolves.
	return g.runPileTailLocked(tail, one, two)
}
