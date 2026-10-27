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
		return finishTake(g, r.Source, r.Chooser, r.Carry, def.Reveal, label, def.Then, picked)
	})
	takeDefs[key] = takeEntry{def: def, pick: pick}
	return TakeThen{key: key}
}

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
	if p.Then != nil || p.Validate != nil || len(p.Slots) > 0 || p.Reveal {
		return fmt.Errorf("%w: keyed take %q also sets a closure or a rule the registration owns", game.ErrInvalidParam, p.Key.key)
	}
	p.Slots, p.Validate, p.Reveal = e.def.Slots, e.def.Validate, e.def.Reveal
	label := e.def.Label
	if label == "" {
		label = p.Label
	}
	prompt, answer := p.ask(ctx.Game, player, source)
	if prompt == nil {
		return finishTake(ctx.Game, source, player, p.Cards, e.def.Reveal, label, e.def.Then, answer)
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
		return eachTakeFinish(ctx, r, def, then)
	})
	return EachTakeThen{run: run, leg: leg, def: def}
}

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
	if e.Take.Then != nil || e.Take.Validate != nil || len(e.Take.Slots) > 0 || e.Take.Reveal {
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
func eachTakeFinish(ctx *Context, r game.RunDone, def TakeDef, then func(ctx *Context, res []TakeFromLibraryResult) error) error {
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
	takes := make([]TakeFromLibraryToHand, n)
	chosen := make([][]uuid.UUID, n)
	for i, player := range players {
		takes[i] = TakeFromLibraryToHand{Player: player, Cards: lists[1+i], Then: def.Then}
		if a, ok := auto[player]; ok {
			chosen[i] = a
		} else {
			chosen[i] = picks.By(player)
		}
	}
	if def.Reveal {
		for i, t := range takes {
			revealTaken(g, t.Player, r.Source, def.Label, chosen[i])
		}
	}
	item := ctx.Item
	var rest func(*game.Game, []TakeFromLibraryResult) error
	if then != nil {
		rest = func(g *game.Game, res []TakeFromLibraryResult) error { return then(NewContext(g, item), res) }
	}
	return eachPlayerTakeMoves(g, r.Source, takes, chosen, nil, rest)
}
