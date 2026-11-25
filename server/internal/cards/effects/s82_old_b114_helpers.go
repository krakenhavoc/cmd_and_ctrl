package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b114_helpers.go — shared pieces of Ikoria: Lair of Behemoths
// batch 114 (#5512): the counter-pick chain and the mutate trigger; the
// "enters with your choice of …" counters and the mentors reuse batch 113.

// S82OldB114CounterPick asks `chooser` which of `candidates` (creatures on
// the battlefield) gets a counter, for each kind of `kinds` in turn, at most
// two: the first is put on the pick, then the next is asked (Call of the
// Death-Dweller's "a deathtouch counter on either of them. Then a menace
// counter on either of them"; Hunted Nightmare's "target opponent puts a
// deathtouch counter on a creature they control"). A kind with one candidate
// left is put on it without a question, and none left ends the chain (CR
// 608.2b). Caller holds g.mu.
func S82OldB114CounterPick(g *game.Game, chooser, source uuid.UUID, name string, candidates []uuid.UUID, kinds ...string) error {
	live := s82OldB114Live(g, candidates)
	switch len(live) {
	case 0:
		return nil
	case 1:
		for _, kind := range kinds {
			if err := NothingIfGone(g.AddCounterForEffect(live[0], kind, 1)); err != nil {
				return err
			}
		}
		return nil
	}
	key := s82OldB114CounterPickLast
	if len(kinds) > 1 {
		key = s82OldB114CounterPickChain
	}
	s82OldB114QueuePick(g, key, chooser, source, name, live, kinds)
	return nil
}

func s82OldB114Live(g *game.Game, ids []uuid.UUID) []uuid.UUID {
	var live []uuid.UUID
	for _, id := range ids {
		if OnBattlefield(g, id) {
			live = append(live, id)
		}
	}
	return live
}

func s82OldB114QueuePick(g *game.Game, key game.ChooseCardsThen, chooser, source uuid.UUID, name string, live []uuid.UUID, kinds []string) {
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   chooser,
		Source:    source,
		Question:  name + " — choose a creature to put a " + kinds[0] + " counter on",
		Cards:     live,
		Min:       1,
		Max:       1,
		Carry:     append([]uuid.UUID{source}, live...),
		CarryWord: name + "|" + strings.Join(kinds, ","),
		ThenKey:   key,
		LiveZone:  game.ZoneBattlefield,
	})
}

// s82OldB114CounterPickLast puts the one kind its prompt names on the
// creature picked. The name and the kinds ride the prompt as its word. An
// on-disk key: never renamed, never reused.
var s82OldB114CounterPickLast = game.RegisterChooseCardsThen("choose-cards/s82-old-b114-counter-pick-last",
	func(g *game.Game, r game.CardsChosen) error {
		_, kindList, ok := strings.Cut(r.Word, "|")
		if len(r.Picked) != 1 || !ok {
			return nil
		}
		return NothingIfGone(g.AddCounterForEffect(r.Picked[0], strings.Split(kindList, ",")[0], 1))
	})

// s82OldB114CounterPickChain puts the first of its two kinds on the creature
// picked, then asks for the second among the same candidates. An on-disk key:
// never renamed, never reused.
var s82OldB114CounterPickChain = game.RegisterChooseCardsThen("choose-cards/s82-old-b114-counter-pick-chain",
	func(g *game.Game, r game.CardsChosen) error {
		name, kindList, ok := strings.Cut(r.Word, "|")
		if len(r.Picked) != 1 || len(r.Carry) < 2 || !ok {
			return nil
		}
		kinds := strings.Split(kindList, ",")
		if len(kinds) != 2 {
			return nil
		}
		if err := NothingIfGone(g.AddCounterForEffect(r.Picked[0], kinds[0], 1)); err != nil {
			return err
		}
		live := s82OldB114Live(g, r.Carry[1:])
		switch len(live) {
		case 0:
			return nil
		case 1:
			return NothingIfGone(g.AddCounterForEffect(live[0], kinds[1], 1))
		}
		s82OldB114QueuePick(g, s82OldB114CounterPickLast, r.Chooser, r.Carry[0], name, live, kinds[1:])
		return nil
	})

// S82OldB114WillOfTheAllHunter is "Target creature gets +2/+2 until end of
// turn. If it's blocking, instead put two +1/+1 counters on it." The
// blocking test is read as the spell resolves (CR 506.4).
func S82OldB114WillOfTheAllHunter() Instruction {
	return OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if ok && c.BlockingTarget != uuid.Nil {
			return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
		}
		return BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Will of the All-Hunter — +2/+2"}.Apply(ctx)
	})
}

// S82OldB114MutatedCreatureYouControl is "whenever a creature you control
// mutates": the merged permanent is controlled by the source's controller as
// the event happens (CR 730.2c, CR 702.140d).
func S82OldB114MutatedCreatureYouControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventMutate {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller
}
