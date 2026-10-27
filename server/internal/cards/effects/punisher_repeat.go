package effects

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// punisher_repeat.go — the repeated three-way punisher: "each opponent
// loses N life unless that player sacrifices a nonland permanent of
// their choice or discards a card", asked some number of times. Torment
// of Hailfire repeats it X times (#568); Rottenmouth Viper once for each
// blight counter on it (ADR 0100 sub-PR 4). Moved out of Torment's file
// unchanged, with the name and the life amount made parameters.
// Append-only.
//
// # The option list is built from what the player can do
//
// CR 608.2 resolves as much as possible: an opponent with no nonland
// permanent is not offered the sacrifice, and one with an empty hand is
// not offered the discard. "Lose N life" is offered always and is
// FIRST, which is the option-pick kind's own contract — the enumerator
// marks the first answer always-legal, so it has to be one that never
// fails. It is also the default the cards print, so the printed order
// and the engine's requirement agree.
//
// # One question at a time
//
// The repetitions are strictly sequential and so are the opponents
// within a repetition (APNAP order): every question is queued by the
// ANSWER to the one before it. That is the whole reason this is a chain
// (#552) rather than repetitions × opponents prompts queued at once,
// and it is observable — a player who sacrificed their last creature to
// repetition one has a different option list in repetition two, and a
// player knocked low by repetition two decides repetition three knowing
// it.

// punisherRepeat is one card's punisher: the name its questions carry
// and the life each refusal costs. A value, so every continuation below
// captures scalars only and an undo that replays an answer resolves it
// against the restored game (the delayed-trigger contract,
// docs/adding-cards.md).
type punisherRepeat struct {
	Name string
	Life int
}

// start builds the whole run of questions — `times` repetitions × the
// opponents, in APNAP order within each repetition — and asks the first
// one. Everything after it is queued by an answer.
//
// Caller holds g.mu.
func (p punisherRepeat) start(ctx *Context, times int) error {
	if times <= 0 {
		return nil
	}
	opponents := ctx.Opponents()
	if len(opponents) == 0 {
		return nil
	}
	queue := make([]uuid.UUID, 0, times*len(opponents))
	for i := 0; i < times; i++ {
		queue = append(queue, opponents...)
	}
	return p.step(ctx, queue)
}

// step asks the first question still owed and hands the rest of the
// queue to its answer.
//
// Caller holds g.mu.
func (p punisherRepeat) step(ctx *Context, remaining []uuid.UUID) error {
	keys, ok := punisherKeysOf[p.Name]
	if !ok {
		return fmt.Errorf("%w: punisherRepeat %q has no registered answers", game.ErrInvalidParam, p.Name)
	}
	for len(remaining) > 0 {
		victim := remaining[0]
		rest := remaining[1:]
		pl := ctx.Game.PlayerByIDForEffect(victim)
		if pl == nil || pl.Eliminated {
			// A player who has left the game is skipped rather than
			// prompted (CR 800.4a); the rest of the run continues.
			remaining = rest
			continue
		}
		queued := PickOption{
			Player:   victim,
			Question: p.Name + " — lose " + strconv.Itoa(p.Life) + " life, or...",
			Options:  p.options(ctx, victim),
			// The opponents (and repetitions) still to ask are the carry
			// (#4253); the branch is the option chosen, read by its label.
			ThenKey: keys.answer,
			Carry:   rest,
		}
		return queued.Apply(ctx)
	}
	return nil
}

// options builds this player's option list. The branch an answer took is
// read off the option it names (the label), so the list and the branches
// cannot drift.
//
// Caller holds g.mu.
func (p punisherRepeat) options(ctx *Context, victim uuid.UUID) []game.ChoiceOption {
	options := []game.ChoiceOption{{
		Label:    "Lose " + strconv.Itoa(p.Life) + " life",
		LifeCost: p.Life,
	}}
	if len(NonlandPermanentsControlledBy(ctx.Game, victim)) > 0 {
		options = append(options, game.ChoiceOption{Label: punisherSacrificeLabel})
	}
	if pl := ctx.Game.PlayerByIDForEffect(victim); pl != nil && pl.Hand != nil && pl.Hand.Size() > 0 {
		options = append(options, game.ChoiceOption{Label: punisherDiscardLabel})
	}
	return options
}

// The two branch labels; the first option is the life, always.
const (
	punisherSacrificeLabel = "Sacrifice a nonland permanent"
	punisherDiscardLabel   = "Discard a card"
)

// punisherKeys are one card's two registered continuations (#4253): the
// answer to the three-way question and the discard. A card has its own
// because the name and the life ride the body, not the prompt.
type punisherKeys struct {
	answer    game.OptionPickThen
	discard   game.DiscardThen
	sacrifice game.TailRef
}

// punisherKeysOf is each punisher's keys by its card name, filled by init.
var punisherKeysOf = map[string]punisherKeys{}

func init() {
	registerPunisher(tormentOfHailfire, "torment-of-hailfire")
	registerPunisher(rottenmouthViper, "rottenmouth-viper")
}

// registerPunisher registers a punisher's continuations under keys that
// name its card ("option-pick/<slug>-punisher" and
// "choose-cards/<slug>-punisher-discarded"): on-disk identities, never
// renamed, never reused. The rest of the queue is the carry of each, and the
// body is run against a Context whose controller is the asking effect's.
func registerPunisher(p punisherRepeat, slug string) {
	next := func(ctx *Context, rest []uuid.UUID) error { return p.step(ctx, rest) }
	keys := punisherKeys{}
	keys.discard = game.RegisterDiscardThen("choose-cards/"+slug+"-punisher-discarded", func(g *game.Game, r game.DiscardsDone) error {
		return next(keyedContext(g, r.Carry[0], r.Source), r.Carry[1:])
	})
	// The sacrifice's rest is a keyed tail (SacrificeChoice.ThenTail): the
	// rest of the queue is its carry, the asking effect's controller and
	// source the tail's own.
	keys.sacrifice = CardTail("then/"+slug+"-punisher-sacrificed", func(ctx *Context, o game.TailOutcome) error {
		return next(ctx, o.Tail.Carry)
	})
	keys.answer = OptionPickThen("option-pick/"+slug+"-punisher", func(ctx *Context, r game.OptionPicked) error {
		victim, rest := r.Chooser, r.Carry
		if r.Option == nil {
			// Nothing could be asked: the player left between the question
			// being built and it being queued. The rest of the run still
			// happens.
			return next(ctx, rest)
		}
		switch r.Option.Label {
		case punisherSacrificeLabel:
			return SacrificeChoice{
				Player:     victim,
				Candidates: NonlandPermanentsControlledBy(ctx.Game, victim),
				Question:   p.Name + " — sacrifice a nonland permanent",
				ThenTail:   withCarry(ctx.Tail(keys.sacrifice), rest...),
			}.Apply(ctx)
		case punisherDiscardLabel:
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   ctx.Source(),
				N:        1,
				Question: p.Name + " — discard a card",
			}, keys.discard, append([]uuid.UUID{ctx.Controller()}, rest...))
		default:
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -p.Life); err != nil {
				return err
			}
			return next(ctx, rest)
		}
	})
	punisherKeysOf[p.Name] = keys
}
