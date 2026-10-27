package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// take_keyed.go — TakeFromLibraryToHand and EachPlayerTakesFromLibrary
// with their rules and their rest of the card REGISTERED rather than held
// in closures (#4290, #4294), so a table waiting on the pick of a look is
// a restore point.
//
// What a take's pick needs after it is answered is the looked-at cards
// (the prompt's carry), the player and the source (the prompt's chooser
// and source), and three things that are card text and so live in the
// registration: whether the taken cards are revealed, the set rule over
// what is taken (slots, and a free rule), and the player's own "the rest".
// The registration is one key (ledgered "choose <key>", never renamed,
// never reused) that names a keyed choose_cards continuation and, for a
// fan-out, the run and its legs.

// TakeDef is the registered part of a take.
type TakeDef struct {
	// Reveal is the printed "you may REVEAL a card from among them":
	// the taken cards become public before they move. Label is the
	// reason shown with the reveal (the printed clause); empty reads the
	// prompt's header.
	Reveal bool
	Label  string

	// Slots and Validate are the rule over the picked SET (TakeSlot,
	// TakeFromLibraryToHand.Validate). Both are read with no *Game: a
	// slot's Match must judge a card by its own characteristics, and
	// Validate is a function of the cards alone, because the rule runs
	// under the enumerator's read lock on whichever table it is handed.
	Slots    []TakeSlot
	Validate func(picked []game.Card) bool

	// Then is the player's own "the rest" (TakeRestOnBottomInRandomOrder,
	// a rider on the result). It must capture nothing.
	Then func(g *game.Game, res TakeFromLibraryResult) error
}

// TakeThen names a registered take. Hold one only from RegisterTake.
type TakeThen struct{ key string }

// Key is the registration's on-disk key.
func (t TakeThen) Key() string { return t.key }

// takeEntry is what a registered take is: its definition and the keyed
// choose_cards continuation that finishes it.
type takeEntry struct {
	def  TakeDef
	pick game.ChooseCardsThen
}

// takeDefs are the registered takes, by key. Written by package
// initialisation only.
var takeDefs = map[string]takeEntry{}

// RegisterTake registers a take under key. Call it once, from a
// package-level var. A TakeFromLibraryToHand names it as Key, and its
// Slots, Validate, Reveal and Then must then be unset on the struct: they
// are the registration's.
func RegisterTake(key string, def TakeDef) TakeThen {
	pick := game.RegisterChooseCardsThenRule(key, takeRule(def), func(g *game.Game, r game.CardsChosen) error {
		var picked []uuid.UUID
		if r.Answered {
			picked = r.Picked
		}
		label := def.Label
		if label == "" {
			label = r.Question
		}
		// r.Carry is the looked-at cards: "the rest" is read off them.
		return finishTake(g, r.Source, r.Chooser, def.Reveal, label, takeMovedTail(key, r.Source, r.Chooser, r.Carry), picked)
	})
	takeDefs[key] = takeEntry{def: def, pick: pick}
	return TakeThen{key: key}
}

// takeMovedTail is a registered take's keyed rest (#4253): the take's key
// as Word, who took and for which clause, and the looked-at cards as Carry.
func takeMovedTail(key string, source, player uuid.UUID, looked []uuid.UUID) game.KeyedTail {
	return game.KeyedTail{Ref: takeMoved, Player: player, Source: source, Word: key, Carry: append([]uuid.UUID(nil), looked...)}
}

// takeMoved runs a registered take's own Then once its move to hand has
// settled, told the cards that reached a hand. An on-disk key.
var takeMoved = game.RegisterTail("take/moved", func(g *game.Game, o game.TailOutcome) error {
	e, ok := takeDefs[o.Tail.Word]
	if !ok || e.def.Then == nil {
		return nil
	}
	return takeThenErr(e.def.Then(g, takeResult(g, o.Tail.Source, o.Tail.Player, o.Tail.Carry, o.IDs)))
})

// takeRule is the set rule a registered take is answered against: the
// slots hold every picked card at once, and Validate agrees. Nil when the
// take has neither.
func takeRule(def TakeDef) game.SetRuleFunc {
	if len(def.Slots) == 0 {
		return def.Validate
	}
	return func(picked []game.Card) bool {
		r := &takeSlotRule{caps: make([]int, len(def.Slots)), fits: make(map[uuid.UUID][]int, len(picked))}
		for i, s := range def.Slots {
			r.caps[i] = s.Max
			if s.Max <= 0 {
				r.caps[i] = len(picked)
			}
		}
		for _, c := range picked {
			for i, s := range def.Slots {
				if s.matcher()(nil, uuid.Nil, c) {
					r.fits[c.InstanceID] = append(r.fits[c.InstanceID], i)
				}
			}
		}
		return r.validate(def.Validate)(picked)
	}
}

// applyKeyed is Apply for a take that names a registered definition.
func (p TakeFromLibraryToHand) applyKeyed(ctx *Context, player, source uuid.UUID) error {
	e, ok := takeDefs[p.Key.key]
	if !ok {
		return fmt.Errorf("%w: take %q", game.ErrUnknownEffectKey, p.Key.key)
	}
	if p.Validate != nil || len(p.Slots) > 0 || p.Reveal {
		return fmt.Errorf("%w: keyed take %q also sets a closure or a rule the registration owns", game.ErrInvalidParam, p.Key.key)
	}
	p.Slots, p.Validate, p.Reveal = e.def.Slots, e.def.Validate, e.def.Reveal
	label := e.def.Label
	if label == "" {
		label = p.Label
	}
	prompt, answer := p.ask(ctx.Game, player, source)
	if prompt == nil {
		return finishTake(ctx.Game, source, player, e.def.Reveal, label, takeMovedTail(p.Key.key, source, player, p.Cards), answer)
	}
	// The registered continuation re-checks nothing but the live zone, and
	// holds the set rule: the prompt carries the looked-at cards.
	prompt.Validate, prompt.Zone = nil, ""
	prompt.LiveZone = game.ZoneLibrary
	prompt.ThenKey = e.pick
	prompt.Carry = append([]uuid.UUID(nil), p.Cards...)
	ctx.Game.QueueChooseCardsForEffect(*prompt)
	return nil
}

// EachTakeThen names a registered fan-out of takes. Hold one only from
// RegisterEachTake.
type EachTakeThen struct {
	run game.PromptRunThen
	leg game.ChooseCardsThen
	def TakeDef
}

// RegisterEachTake registers EachPlayerTakesFromLibrary's rest under two
// keys: runKey is the run's continuation (ledgered "run <runKey>") and
// legKey each player's pick, which holds the set rule (ledgered "choose
// <legKey>"). def is every player's take, and `then` the rest of the card
// after every take has finished, handed a Context rebuilt from the
// resolving item and one result per player in Players order (nil for
// none). Call it once, from a package-level var.
func RegisterEachTake(runKey, legKey string, def TakeDef, then func(ctx *Context, res []TakeFromLibraryResult) error) EachTakeThen {
	leg := game.RegisterChooseCardsRunLegRule(legKey, takeRule(def))
	run := RunThen(runKey, func(ctx *Context, r game.RunDone) error {
		return eachTakeFinish(ctx, r, runKey, def)
	})
	eachTakeDefs[runKey] = eachTakeEntry{def: def, then: then}
	return EachTakeThen{run: run, leg: leg, def: def}
}

// eachTakeEntry is a registered fan-out's definition and its rest of the
// card, read by its keyed moves (eachTakeMoved).
type eachTakeEntry struct {
	def  TakeDef
	then func(ctx *Context, res []TakeFromLibraryResult) error
}

// eachTakeDefs are the registered fan-outs, by run key. Written by package
// initialisation only.
var eachTakeDefs = map[string]eachTakeEntry{}

// applyKeyed is Apply for a fan-out that names a registered definition.
//
// The run's carry is the players in order, then each player's looked-at
// cards, then, for each player who was not asked (no choice to make), a
// list led by the player and followed by what they take.
func (e EachPlayerTakesFromLibrary) applyKeyed(ctx *Context) error {
	g := ctx.Game
	source := ctx.Source()
	players := e.Players
	if players == nil {
		players = apnapPlayers(g)
	}
	if e.Take.Validate != nil || len(e.Take.Slots) > 0 || e.Take.Reveal {
		return fmt.Errorf("%w: keyed fan-out of takes also sets a closure or a rule the registration owns", game.ErrInvalidParam)
	}
	def := e.Keyed.def
	lists := [][]uuid.UUID{players}
	looked := make([][]uuid.UUID, len(players))
	var autos [][]uuid.UUID
	var legs []game.ChooseCardsPrompt
	for i, player := range players {
		t := e.Take
		t.Player = player
		t.Cards = g.LookAtTopOfLibraryForEffect(player, e.N)
		t.Slots, t.Validate, t.Reveal = def.Slots, def.Validate, def.Reveal
		looked[i] = t.Cards
		prompt, answer := t.ask(g, player, source)
		if prompt == nil {
			autos = append(autos, append([]uuid.UUID{player}, answer...))
			continue
		}
		prompt.Validate, prompt.Zone = nil, ""
		prompt.LiveZone = game.ZoneLibrary
		prompt.ThenKey = e.Keyed.leg
		legs = append(legs, *prompt)
	}
	lists = append(append(lists, looked...), autos...)
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, e.Keyed.run, source, ctx.RunData(packIDLists(lists...)))
	return err
}

// eachTakeFinish is the fan-out's continuation: every player's take, in
// order, then the card's rest.
func eachTakeFinish(ctx *Context, r game.RunDone, runKey string, def TakeDef) error {
	g := ctx.Game
	lists := unpackIDLists(r.Carry)
	var players []uuid.UUID
	if len(lists) > 0 {
		players = lists[0]
	}
	n := len(players)
	auto := map[uuid.UUID][]uuid.UUID{}
	if len(lists) > 1+n {
		for _, l := range lists[1+n:] {
			if len(l) > 0 {
				auto[l[0]] = l[1:]
			}
		}
	}
	picks := r.Picks()
	looked := make([][]uuid.UUID, n)
	chosen := make([][]uuid.UUID, n)
	for i, player := range players {
		looked[i] = lists[1+i]
		if a, ok := auto[player]; ok {
			chosen[i] = a
		} else {
			chosen[i] = picks.By(player)
		}
	}
	if def.Reveal {
		for i, player := range players {
			revealTaken(g, player, r.Source, def.Label, chosen[i])
		}
	}
	step := eachTakeStep{runKey: runKey, source: r.Source, controller: ctx.Controller(), asked: game.NewAskedItem(ctx.Item),
		players: players, looked: looked, chosen: chosen}
	return step.move(g)
}

// eachTakeStep is the fan-out's moves as data (#4253): player len(done)'s
// picks move to their hand, and the move's keyed tail (eachTakeMoved) runs
// that player's own Then and moves the next player's, so a CR 903.9b pause
// on any player's move is a restore point and holds the rest of the
// table's moves behind it. The final rest of the card runs once, with every
// result.
type eachTakeStep struct {
	runKey     string
	source     uuid.UUID
	controller uuid.UUID
	asked      *game.AskedItem
	players    []uuid.UUID
	looked     [][]uuid.UUID
	chosen     [][]uuid.UUID
	done       []TakeFromLibraryResult
}

// move moves the next player's picks, or runs the card's rest when every
// player is done.
func (s eachTakeStep) move(g *game.Game) error {
	i := len(s.done)
	if i >= len(s.players) {
		e, ok := eachTakeDefs[s.runKey]
		if !ok || e.then == nil {
			return nil
		}
		ctx := tailContext(g, game.KeyedTail{Player: s.controller, Source: s.source, Asked: s.asked})
		return e.then(ctx, s.done)
	}
	// Every player's take was revealed together, before any card moved.
	return finishTake(g, s.source, s.players[i], false, "", s.tail(), s.chosen[i])
}

// tail writes the step as eachTakeMoved's tail: the run key as Word, the
// asking item, and in Carry (packIDLists) the players, each one's look,
// each one's picks, and for each player done the cards they took and the
// rest of their look as it was when their take finished.
func (s eachTakeStep) tail() game.KeyedTail {
	lists := [][]uuid.UUID{s.players}
	lists = append(lists, s.looked...)
	lists = append(lists, s.chosen...)
	for _, d := range s.done {
		lists = append(lists, d.Taken, d.Rest)
	}
	return game.KeyedTail{Ref: eachTakeMoved, Player: s.controller, Source: s.source, Word: s.runKey, Asked: s.asked, Carry: packIDLists(lists...)}
}

// eachTakeStepOf reads a step back from eachTakeMoved's tail; false when the
// carry does not hold one.
func eachTakeStepOf(t game.KeyedTail) (eachTakeStep, bool) {
	s := eachTakeStep{runKey: t.Word, source: t.Source, controller: t.Player, asked: t.Asked}
	lists := unpackIDLists(t.Carry)
	if len(lists) == 0 {
		return s, false
	}
	s.players = lists[0]
	n := len(s.players)
	if len(lists) < 1+2*n || (len(lists)-1-2*n)%2 != 0 {
		return s, false
	}
	s.looked = lists[1 : 1+n]
	s.chosen = lists[1+n : 1+2*n]
	for j, rest := 0, lists[1+2*n:]; j+1 < len(rest); j += 2 {
		k := len(s.done)
		s.done = append(s.done, TakeFromLibraryResult{
			Source: s.source, Player: s.players[k], Taken: rest[j], Rest: rest[j+1],
			Looked: append([]uuid.UUID(nil), s.looked[k]...),
		})
	}
	return s, len(s.done) < n
}

// eachTakeMoved is one fan-out step's rest, told the cards that reached the
// player's hand: that player's own Then, then the next player's move. The
// player's own Then's error is returned after the rest of the table's moves,
// as each player's take is its own. An on-disk key. Registered from init:
// the body moves the next player, whose tail names this one.
var eachTakeMoved game.TailRef

func init() { eachTakeMoved = game.RegisterTail("take/each-moved", eachTakeMovedBody) }

func eachTakeMovedBody(g *game.Game, o game.TailOutcome) error {
	s, ok := eachTakeStepOf(o.Tail)
	if !ok {
		return nil
	}
	i := len(s.done)
	res := takeResult(g, s.source, s.players[i], s.looked[i], o.IDs)
	var ownErr error
	if e, ok := eachTakeDefs[s.runKey]; ok && e.def.Then != nil {
		ownErr = e.def.Then(g, res)
	}
	s.done = append(append(make([]TakeFromLibraryResult, 0, i+1), s.done...), res)
	if err := unwrapTakeThen(s.move(g)); err != nil {
		return takeThenErr(err)
	}
	return takeThenErr(ownErr)
}
