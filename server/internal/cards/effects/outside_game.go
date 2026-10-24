package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// outside_game.go — "a card you own from outside the game" (CR 400.11,
// ADR 0191). A deck's sideboard is that pool; this file is the effect
// side of it. The only way a card leaves the sideboard is an effect
// that brings one in (CR 400.11c).
//
// Every pick here is KEYED (#4256, game/outside_pick_keyed.go): each
// card registers its filter and continuation under one key, from a
// package-level var, and a table waiting on the pick is a restore point
// on either branch, the sideboard's choose_cards pick and the card
// database's typed-name pick. The key is an on-disk identity: never
// renamed, never reused, ledgered in game/testdata/effect_keys.txt.

// outsideTake is one card's "put a card you own from outside the game
// into your hand" pick: the registered pick, and the filter, kept so a
// caller that offers the same kind of card from another zone (Karn, the
// Great Creator's exile) filters it the same way.
type outsideTake struct {
	pick  game.OutsidePick
	match game.OutsideMatchFunc
}

// registerOutsideTake registers a card's take-from-outside pick under
// `key`. `match` narrows the pool (nil offers every card) and must
// capture nothing but constants of the card's text. `reveal` shows the
// chosen card to the table before it moves: every printed wish reveals,
// Mastermind's Acquisition does not. Call it once, from a package-level
// var.
func registerOutsideTake(key string, match func(game.Card) bool, reveal bool) outsideTake {
	var filter game.OutsideMatchFunc
	if match != nil {
		filter = func(c game.Card, _ string) bool { return match(c) }
	}
	return registerOutsideTakeOfWord(key, filter, reveal)
}

// registerOutsideTakeOfWord is registerOutsideTake for a filter that
// reads one word of the card's text a text-changing effect may have
// changed (CR 612.2): Coax from the Blind Eternities' "Eldrazi". The
// word rides on the prompt (TakeFromOutside.Word), so the restored
// table's filter reads the same word.
func registerOutsideTakeOfWord(key string, match game.OutsideMatchFunc, reveal bool) outsideTake {
	return outsideTake{
		match: match,
		pick: game.RegisterOutsidePick(key, match, func(g *game.Game, r game.CardsChosen) error {
			if !r.Answered {
				return nil
			}
			// A keyed pick has no live-zone re-check of its own: a card
			// that has left the container is not brought in.
			picked := g.StillOutsideForEffect(r.Chooser, r.Picked)
			if len(picked) == 0 {
				return nil
			}
			if reveal {
				g.RevealOutsideForEffect(game.RevealSpec{
					Player: r.Chooser,
					Source: r.Source,
					Reason: r.Question,
					Cards:  picked[:1],
				})
			}
			return g.BringFromOutsideForEffect(r.Chooser, picked[0])
		}),
	}
}

// matches is the take's filter over one card, with `word` bound.
func (t outsideTake) matches(c game.Card, word string) bool {
	return t.match == nil || t.match(c, word)
}

// TakeFromOutside is "You may reveal a <kind> card you own from outside
// the game and put it into your hand" (the wishes) and its unrevealed,
// mandatory sibling "Put a card you own from outside the game into your
// hand" (Mastermind's Acquisition).
//
// It asks the owner to choose from their sideboard (a choose_cards
// prompt over cards only they can see) and brings the choice into their
// hand. A player whose deck seated no sideboard chooses from the card
// database instead (ADR 0198): the same filter, minted into a real card
// they own, and the same reveal and move. With nothing to choose from it does nothing and asks nothing: a
// clause that finds no card is not an error (CR 608.2c's "as much as
// possible"). While another player controls the owner's turn the pool is
// closed (ADR 0167 §5, CR 723.4) and it finds nothing.
type TakeFromOutside struct {
	// Player owns the cards and chooses. Zero means the resolving
	// effect's controller.
	Player uuid.UUID
	// Pick is the card's registered pick (registerOutsideTake): its
	// filter, whether it reveals, and its key. Required.
	Pick outsideTake
	// Word is the one word of the card's text the filter reads, when it
	// reads one (registerOutsideTakeOfWord). Plain data on the prompt.
	Word string
	// Question is the prompt's header, written the way the card is.
	Question string
	// Required is "put a card" rather than "you may reveal a card": the
	// owner may not decline when something matches.
	Required bool
}

func (t TakeFromOutside) Apply(ctx *Context) error {
	player := t.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	takeFromOutside(ctx.Game, player, ctx.Source(), t)
	return nil
}

// takeFromOutside queues t's pick for `player`, with `source` as the card
// asking. It is Apply without a Context, so a continuation that runs
// after the resolving frame (learn's rummage-or-Lesson choice) can ask
// the same question the same way. Caller must hold g.mu.
func takeFromOutside(g *game.Game, player, source uuid.UUID, t TakeFromOutside) {
	floor := 0
	if t.Required {
		floor = 1
	}
	g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:   player,
		Source:   source,
		Question: t.Question,
		Min:      floor,
		Max:      1,
		Keyed:    t.Pick.pick,
		MatchArg: t.Word,
	})
}

// outsideShuffle is one card's "shuffle up to N cards you own from
// outside the game into your library" pick, keyed.
type outsideShuffle struct{ pick game.OutsidePick }

// registerOutsideShuffle registers a card's shuffle-from-outside pick
// under `key`, with `match` narrowing the pool (nil offers every card).
// Call it once, from a package-level var.
func registerOutsideShuffle(key string, match func(game.Card) bool) outsideShuffle {
	var filter game.OutsideMatchFunc
	if match != nil {
		filter = func(c game.Card, _ string) bool { return match(c) }
	}
	return outsideShuffle{pick: game.RegisterOutsidePick(key, filter, func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered {
			// The owner left with the prompt open, and their library
			// with them.
			return nil
		}
		if _, err := g.PutFromOutsideIntoLibraryForEffect(r.Chooser, r.Picked); err != nil {
			return err
		}
		return g.ShuffleLibraryForEffect(r.Chooser)
	})}
}

// ShuffleFromOutside is "Shuffle up to N cards you own from outside the
// game into your library" (Research, #3525): the move the other way.
//
// It asks the owner to choose up to Max cards from their sideboard (the
// card database for a deck that seated none, ADR 0198), puts them on the
// bottom of the library unrevealed, and shuffles it. The library is
// shuffled whether or not any card was chosen or any was there to
// choose: an effect that would shuffle a set of objects into a library
// shuffles it even when the set is empty (CR 701.24d), and a pool that
// is closed to a controller of the owner's turn is an empty set.
type ShuffleFromOutside struct {
	// Player owns the cards and the library. Zero means the resolving
	// effect's controller.
	Player uuid.UUID
	// Pick is the card's registered pick (registerOutsideShuffle).
	// Required.
	Pick outsideShuffle
	// Max is the ceiling of the pick ("up to four").
	Max int
	// Question is the prompt's header, written the way the card is.
	Question string
}

func (s ShuffleFromOutside) Apply(ctx *Context) error {
	player := s.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	queued := ctx.Game.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:   player,
		Source:   ctx.Source(),
		Question: s.Question,
		Min:      0,
		Max:      s.Max,
		Keyed:    s.Pick.pick,
	})
	if queued {
		return nil
	}
	return ctx.Game.ShuffleLibraryForEffect(player)
}
