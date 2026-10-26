package game

import (
	"github.com/google/uuid"
)

// proliferate_keyed.go — #4253 slice 10b: a player's proliferate (CR
// 701.34a) whose rest of the sentence is a registered tail, run as data
// from the prompt to the last counter, so a table waiting on the choice, or
// on a CR 616 prompt over one of its counters, is a restore point.
//
// ProliferateChoosingForEffect's closure form holds the rest of the card on
// the prompt's chooseCardsResume frame and walks the counters through
// closures. The keyed form is three registered pieces:
//
//   - the choice is a keyed card-set pick (proliferatePicked): the prompt
//     carries the proliferates still owed (PickNums) and the rest of the
//     sentence (pickTail);
//   - the walk is one engine tail (proliferateStepTail) that places one
//     counter and names itself as that counter's tail, its position as
//     data: the chosen permanents and players (Carry, permanents first),
//     how many of them are permanents, the target in turn, the counter kind
//     in turn and the proliferates still owed (Nums), and the kinds the
//     target had when its turn came (Words), so a replacement that adds a
//     new kind mid-placement does not cascade;
//   - the rest of the sentence is the walk's Next.

// ProliferateChoosingKeyedForEffect is ProliferateChoosingForEffect with
// the rest of the sentence as a registered tail (#4253 slice 10b): run once
// every proliferate of the settled count has been answered and every
// counter has landed, or at once when there was nothing to do. A zero tail
// is "nothing more".
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) ProliferateChoosingKeyedForEffect(player, source uuid.UUID, then KeyedTail) error {
	_, err := g.runKeywordActionLocked(&ReplacementEvent{
		Kind:               RepEventKeywordAction,
		Actor:              player,
		Source:             source,
		KeywordAction:      KeywordActionProliferate,
		KeywordActionCount: 1,
		keywordAction:      &keywordActionTail{ask: true, keyed: copyKeyedTail(then)},
	})
	return err
}

// proliferateAskKeyedLocked is proliferateAskLocked with the rest as data.
//
// Caller must hold g.mu.
func (g *Game) proliferateAskKeyedLocked(actor, source uuid.UUID, remaining int, rest KeyedTail) error {
	if remaining <= 0 {
		return rest.Run(g, nil)
	}
	cards, players := g.ProliferateCandidatesForEffect()
	if len(cards)+len(players) == 0 {
		return rest.Run(g, nil)
	}
	sc, sp := g.ProliferateSuggestionForEffect(actor)
	n := len(cards) + len(players)
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:            PendingChoiceProliferate,
		Chooser:         actor,
		FromPlayer:      actor,
		Count:           n,
		Source:          source,
		Reason:          proliferateQuestion,
		ChooseCards:     cards,
		ChoosePlayers:   players,
		ChooseSuggested: append(sc, sp...),
		ChooseMin:       0,
		ChooseMax:       n,
		PickThen:        proliferatePicked.key,
		PickNums:        []int{remaining - 1},
		pickTail:        copyKeyedTail(rest),
	})
	if id == uuid.Nil {
		// The proliferating player is gone: nothing to give, and the rest
		// of the sentence still runs.
		return rest.Run(g, nil)
	}
	return nil
}

// proliferatePicked is the keyed answer to a proliferate prompt: the picks
// split into permanents and players, and the walk starts. An unanswered
// prompt gives nothing and goes on. An on-disk key: never renamed, never
// reused. Registered from init: the walk asks the next proliferate, which
// names this key.
var proliferatePicked ChooseCardsThen

func proliferateAnswered(g *Game, r CardsChosen) error {
	var cards, players []uuid.UUID
	for _, id := range r.Picked {
		if g.playerByIDLocked(id) != nil {
			players = append(players, id)
		} else {
			cards = append(cards, id)
		}
	}
	remaining := 0
	if len(r.Nums) > 0 {
		remaining = r.Nums[0]
	}
	return g.proliferateStepLocked(proliferateRun(r.From, r.Source, cards, players, remaining, r.Tail))
}

// proliferateRun is the walk's first position.
func proliferateRun(placer, source uuid.UUID, cards, players []uuid.UUID, remaining int, rest KeyedTail) KeyedTail {
	run := KeyedTail{
		Ref: proliferateStepTail, Player: placer, Source: source,
		Carry: append(copyUUIDs(cards), players...),
		Nums:  []int{len(cards), 0, 0, remaining},
	}
	if !rest.IsZero() {
		next := copyKeyedTail(rest)
		run.Next = &next
	}
	return run
}

// proliferateStepTail is the walk itself, told nothing it needs: each
// counter placed names it again with the position after. An on-disk key:
// never renamed, never reused.
var proliferateStepTail TailRef

func init() {
	proliferatePicked = RegisterChooseCardsThen("choose-cards/proliferate", proliferateAnswered)
	proliferateStepTail = RegisterTail("engine/proliferate-step", func(g *Game, o TailOutcome) error {
		return g.proliferateStepLocked(o.Tail)
	})
}

// proliferateStepLocked places the counter `run` is at, with the position
// after it as that placement's tail, or, past the last target, asks for the
// next proliferate owed or runs the rest of the sentence.
//
// Caller must hold g.mu.
func (g *Game) proliferateStepLocked(run KeyedTail) error {
	if len(run.Nums) < 4 {
		return nil
	}
	nCards, i, k, remaining := run.Nums[0], run.Nums[1], run.Nums[2], run.Nums[3]
	kinds := copyStrings(run.Words)
	placer := run.Player
	for i < len(run.Carry) {
		id := run.Carry[i]
		if len(kinds) == 0 {
			// The target's turn: its kinds as they are now.
			if i < nCards {
				if c, ok := g.LookupCardForEffect(id); ok {
					kinds = sortedCounterKinds(c.Counters)
				}
			} else if p := g.playerByIDLocked(id); p != nil && !p.Eliminated {
				kinds = sortedCounterKinds(p.Counters)
			}
			k = 0
		}
		if k >= len(kinds) {
			i, k, kinds = i+1, 0, nil
			continue
		}
		next := copyKeyedTail(run)
		next.Nums = []int{nCards, i, k + 1, remaining}
		next.Words = kinds
		if i < nCards {
			return g.AddCounterByKeyedForEffect(placer, id, kinds[k], 1, next)
		}
		return g.addPlayerCounterByKeyedLocked(placer, id, kinds[k], 1, next)
	}
	rest := nextTailValue(run.Next)
	if remaining > 0 {
		return g.proliferateAskKeyedLocked(placer, run.Source, remaining, rest)
	}
	return rest.Run(g, nil)
}
