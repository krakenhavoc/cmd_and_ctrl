package effects

import (
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
		options, branches := p.options(ctx, victim)
		queued := PickOption{
			Player:   victim,
			Question: p.Name + " — lose " + strconv.Itoa(p.Life) + " life, or...",
			Options:  options,
			Then:     p.answered(victim, branches, rest),
		}
		return queued.Apply(ctx)
	}
	return nil
}

// options builds this player's option list and the matching branch
// keys. The keys are plain strings rather than closures so the two
// lists cannot drift: the index the chooser answers with indexes both.
//
// Caller holds g.mu.
func (p punisherRepeat) options(ctx *Context, victim uuid.UUID) ([]game.ChoiceOption, []string) {
	options := []game.ChoiceOption{{
		Label:    "Lose " + strconv.Itoa(p.Life) + " life",
		LifeCost: p.Life,
	}}
	branches := []string{punisherBranchLife}

	if len(NonlandPermanentsControlledBy(ctx.Game, victim)) > 0 {
		options = append(options, game.ChoiceOption{Label: "Sacrifice a nonland permanent"})
		branches = append(branches, punisherBranchSacrifice)
	}
	if pl := ctx.Game.PlayerByIDForEffect(victim); pl != nil && pl.Hand != nil && pl.Hand.Size() > 0 {
		options = append(options, game.ChoiceOption{Label: "Discard a card"})
		branches = append(branches, punisherBranchDiscard)
	}
	return options, branches
}

// The three branch keys.
const (
	punisherBranchLife      = "life"
	punisherBranchSacrifice = "sacrifice"
	punisherBranchDiscard   = "discard"
)

// answered runs the chosen branch and then the next question.
// Everything it captures is a scalar or a frozen slice.
func (p punisherRepeat) answered(victim uuid.UUID, branches []string, rest []uuid.UUID) func(ctx *Context, index int) error {
	return func(ctx *Context, index int) error {
		next := func(ctx *Context) error { return p.step(ctx, rest) }
		if index < 0 || index >= len(branches) {
			// Nothing could be asked — the player left between the
			// question being built and it being queued. The rest of
			// the run still happens.
			return next(ctx)
		}
		switch branches[index] {
		case punisherBranchSacrifice:
			return SacrificeChoice{
				Player:     victim,
				Candidates: NonlandPermanentsControlledBy(ctx.Game, victim),
				Question:   p.Name + " — sacrifice a nonland permanent",
				Then:       next,
			}.Apply(ctx)
		case punisherBranchDiscard:
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   ctx.Source(),
				N:        1,
				Question: p.Name + " — discard a card",
				Then: func(g *game.Game, _ uuid.UUID, _ []uuid.UUID) error {
					return p.step(NewContext(g, ctx.Item), rest)
				},
			})
			return nil
		default:
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -p.Life); err != nil {
				return err
			}
			return next(ctx)
		}
	}
}
