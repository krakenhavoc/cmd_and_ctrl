package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_player.go — the card-facing shape of "choose a player" /
// "choose an opponent" at resolution (#929).
//
// The engine side is game/choose_player.go, which explains why this
// rides the option_pick kind and why the answer lives on the item's
// payload. What this file adds is the vocabulary a card writes in:
// the POOL the clause admits, the seats to leave out, and a branch
// that is an ordinary ctx-taking effect.
//
// Anything printed after the choice goes in Then, not after Apply
// returns — Apply only queues the prompt, exactly as Scry.Then and
// MayChoice.OnYes exist.
//
// "Choose a player" is not "target player". A card that prints
// "target" declares a TargetSpec and reads item.Targets; this
// primitive is for the clause that does not, and nothing can respond
// to the pick because it is made mid-resolution.

// PlayerPool is the set of seats a ChoosePlayer admits, read off the
// card's own words. The zero value is Players.
type PlayerPool struct {
	kind poolKind
	of   uuid.UUID
}

type poolKind uint8

const (
	poolPlayers poolKind = iota
	poolOpponents
)

var (
	// Players is every seated player, the chooser included — "choose
	// a player" (Gluntch, the Bestower). The zero value of
	// PlayerPool, so a ChoosePlayer that names no pool offers the
	// whole table.
	Players = PlayerPool{kind: poolPlayers}

	// Opponents is every opponent of the CHOOSER — "choose an
	// opponent" (Slithermuse, Skullwinder).
	Opponents = PlayerPool{kind: poolOpponents}
)

// OpponentsOf is every opponent of `id` rather than of the chooser —
// "target player chooses an opponent", where the two are different
// seats. Nothing in the catalog prints it yet; it exists because the
// pool and the chooser are genuinely independent and a primitive that
// conflated them would have to be re-cut the first time a card
// separated them.
func OpponentsOf(id uuid.UUID) PlayerPool {
	return PlayerPool{kind: poolOpponents, of: id}
}

// seats resolves the pool against the live game. Eligibility (a seat
// that has left) and ordering are the engine's, not this function's —
// see game.QueueChoosePlayerForEffect.
func (p PlayerPool) seats(ctx *Context, chooser uuid.UUID) []uuid.UUID {
	var all []uuid.UUID
	for _, seat := range ctx.Game.Seats {
		if seat == nil {
			continue
		}
		all = append(all, seat.ID)
	}
	if p.kind != poolOpponents {
		return all
	}
	me := p.of
	if me == uuid.Nil {
		me = chooser
	}
	out := all[:0:0]
	for _, id := range all {
		if id != me {
			out = append(out, id)
		}
	}
	return out
}

// ChoosePlayer asks Chooser to name a player while an effect resolves,
// records the answer on the resolving item, and runs Then.
//
// Then reads the answer with ctx.ChosenPlayer(). It runs even when no
// question could be asked — an empty pool, or a chooser who has left
// (CR 800.4a) — and ctx.ChosenPlayer() is uuid.Nil in that case, so a
// branch always checks before it acts. That is the #544 rule: a
// continuation that is silently never called is a card that stops
// halfway.
type ChoosePlayer struct {
	// Chooser is asked. Zero means the resolving effect's controller,
	// which is what every printed "choose a player" means.
	Chooser uuid.UUID

	// Among is the pool the clause admits. Zero value is Players.
	Among PlayerPool

	// Only, when set, is exactly the seats the clause admits, in place of
	// Among — for a clause whose pool is computed from the board:
	// "choose an opponent with the most life among your opponents" (The
	// Master, Gallifrey's End) is the tied opponents. Except still
	// applies. The engine drops any seat that has left the game.
	Only []uuid.UUID

	// Except are seats the clause rules out — "choose a SECOND
	// player" is Players minus the players already chosen, which is
	// ctx.ChosenPlayers().
	Except []uuid.UUID

	// Question is the prompt's header, written the way the card is.
	Question string

	// Then runs once the answer is in (or once it is settled that
	// there is none). Runs with g.mu held; may queue further choices.
	Then func(ctx *Context) error

	// ThenKey is Then as a registered continuation (OptionPickThen, #4253),
	// and Carry the IDs it is handed: a prompt queued with a key holds no
	// closure, so a table waiting on the answer is a restore point. The
	// answer is r.Option.Player (r.Option is nil when no question could be
	// asked, and the continuation still runs), and the Context is a
	// stand-in built from the controller and the source card, not the
	// resolving item, so ctx.ChosenPlayer() reads nothing: read the
	// answer. Set ThenKey or Then, not both.
	ThenKey game.OptionPickThen
	Carry   []uuid.UUID
}

func (c ChoosePlayer) Apply(ctx *Context) error {
	chooser := c.Chooser
	if chooser == uuid.Nil {
		chooser = ctx.Controller()
	}
	item := ctx.Item
	then := c.Then
	pool := c.Only
	if pool == nil {
		pool = c.Among.seats(ctx, chooser)
	}
	if c.ThenKey.Key() != "" {
		if then != nil {
			return fmt.Errorf("%w: ChoosePlayer %q sets both ThenKey and Then", game.ErrInvalidParam, c.ThenKey.Key())
		}
		carry := append([]uuid.UUID{ctx.Controller()}, c.Carry...)
		if ctx.Game.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
			Chooser:  chooser,
			Among:    exceptPlayers(pool, c.Except),
			Source:   ctx.Source(),
			Question: c.Question,
			ThenKey:  c.ThenKey,
			Carry:    carry,
		}) != uuid.Nil {
			return nil
		}
		// Nothing could be asked: the rest of the card still happens, with
		// nobody chosen.
		return ctx.Game.RunOptionPickThenForEffect(c.ThenKey, game.OptionPicked{
			Chooser: chooser, Source: ctx.Source(), Index: game.NoChoiceIndex, Carry: carry,
		})
	}
	queued := ctx.Game.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  chooser,
		Among:    exceptPlayers(pool, c.Except),
		Source:   ctx.Source(),
		Question: c.Question,
		Item:     item,
		Then: func(g *game.Game, _ uuid.UUID) error {
			return resumeClause(g, item, then)
		},
	})
	if queued != uuid.Nil || then == nil {
		return nil
	}
	// Nothing could be asked. The engine has already recorded the
	// absence on the item, so ctx.ChosenPlayer() reads uuid.Nil here
	// rather than the previous clause's answer.
	return then(ctx)
}

// ChoosePlayerAsEnters builds the `Spec.AsEnters` for "As this
// permanent enters, choose a player" (CR 614.12) — True-Name Nemesis.
// `label` is the prompt header, normally the card's name.
//
// THE THIRD FORM OF THE QUESTION, and the one that lasts. ChoosePlayer
// above is a choice made while an effect RESOLVES: its answer lives on
// the stack item and is gone when the item is. This one is made as the
// permanent enters and is stored on it (game.Card.ChosenPlayer), read
// for the rest of that permanent's life by whatever printed clause
// asked — on True-Name Nemesis, "protection from the chosen player".
// It is the shape ChooseColorAsEnters already has for colours, down to
// the declared simplification: the prompt is queued from the AsEnters
// hook rather than by pausing the CR 614 pipeline, so the permanent is
// briefly on the battlefield with nobody chosen. Nothing can act in
// that window, and an unchosen player is nobody rather than everybody.
//
// `Among` is the pool the clause admits, zero value Players — "choose
// a player" includes the controller, and on the Nemesis that is a legal
// if pointless answer.
func ChoosePlayerAsEnters(label string, among PlayerPool) func(*game.Card, *Context) error {
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueChoosePlayerAsEntersForEffect(
			card.Controller,
			card.InstanceID,
			label+" — choose a player",
			among.seats(ctx, card.Controller),
		)
		return nil
	}
}

// ChoosePlayersAsEnters is "As this enters, choose two players" (CR
// 614.12), Sower of Discord (#3389): two prompts in a row, the second
// leaving out the first, stored on game.Card.ChosenPlayer and
// ChosenSecondPlayer. Read the pair with ChosenPlayersOf. `among` is
// the pool the clause admits, zero value Players.
func ChoosePlayersAsEnters(label string, among PlayerPool) func(*game.Card, *Context) error {
	return func(card *game.Card, ctx *Context) error {
		ctx.Game.QueueChooseTwoPlayersAsEntersForEffect(
			card.Controller,
			card.InstanceID,
			label+" — choose two players",
			among.seats(ctx, card.Controller),
		)
		return nil
	}
}

// ChosenPlayersOf is the pair stored on the permanent `source` by a
// two-player as-enters choice; either is uuid.Nil while unchosen.
// Read-only; safe under either lock.
func ChosenPlayersOf(g *game.Game, source uuid.UUID) (first, second uuid.UUID) {
	return g.ChosenPlayersOf(source)
}

// ChosenPlayerOf is the player stored on the permanent `source` by an
// as-enters choice, or uuid.Nil while none has been made. Read-only;
// safe under either lock.
func ChosenPlayerOf(g *game.Game, source uuid.UUID) uuid.UUID {
	return g.ChosenPlayerOf(source)
}

// OpponentsWithTheMostLife is "the opponents with the most life among
// your opponents": every opponent of the effect's controller still in
// the game whose life total is the highest of theirs, in seat order.
// One seat is the answer to "the opponent with the most life"; several
// is the tie the controller breaks (CR 101.4 says nothing, the card
// says "choose"). Empty when the controller has no opponents left.
// Pass the result to ChoosePlayer.Only only when it is not empty: an
// empty list there falls back to the Among pool.
func OpponentsWithTheMostLife(ctx *Context) []uuid.UUID {
	var best []uuid.UUID
	top := 0
	for _, id := range ctx.Opponents() {
		p := ctx.Game.PlayerByIDForEffect(id)
		if p == nil || p.Eliminated {
			continue
		}
		switch {
		case len(best) == 0 || p.Life > top:
			best, top = []uuid.UUID{id}, p.Life
		case p.Life == top:
			best = append(best, id)
		}
	}
	return best
}

// exceptPlayers drops `except` from `ids`, keeping order.
func exceptPlayers(ids, except []uuid.UUID) []uuid.UUID {
	if len(except) == 0 {
		return ids
	}
	drop := make(map[uuid.UUID]bool, len(except))
	for _, id := range except {
		drop[id] = true
	}
	out := ids[:0:0]
	for _, id := range ids {
		if !drop[id] {
			out = append(out, id)
		}
	}
	return out
}
