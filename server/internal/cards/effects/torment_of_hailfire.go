package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torment of Hailfire — Sorcery {X}{B}{B}:
//
//	"Repeat the following process X times. Each opponent loses 3 life
//	 unless that player sacrifices a nonland permanent of their choice
//	 or discards a card."
//
// The card that forced the option pick (#568). Each repetition puts a
// THREE-WAY question in front of every opponent, and none of the three
// prompts that existed could ask it: pay_unless is a mana payment,
// confirm is two branches, and a trigger's "you may" belongs to the
// trigger's controller.
//
// # The option list is built from what the player can do
//
// CR 608.2 resolves as much as possible: an opponent with no nonland
// permanent is not offered the sacrifice, and one with an empty hand
// is not offered the discard. "Lose 3 life" is offered always and is
// FIRST, which is the option-pick kind's own contract — the enumerator
// marks the first answer always-legal, so it has to be one that never
// fails. It is also the default the card prints, so the printed order
// and the engine's requirement agree.
//
// # One question at a time
//
// The repetitions are strictly sequential and so are the opponents
// within a repetition (APNAP order): every question is queued by the
// ANSWER to the one before it. That is the whole reason this is a
// chain (#552) rather than X × opponents prompts queued at once, and
// it is observable — a player who sacrificed their last creature to
// repetition one has a different option list in repetition two, and a
// player who has been knocked to 3 life by repetition two is deciding
// repetition three knowing it.
//
// X is the announced X (CR 601.2b), read off the stack item.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30a0a6c2-1fbb-4784-ab96-22611d57e62c",
		Name:         "Torment of Hailfire",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return tormentOfHailfireStart(ctx)
		},
	})
}

// tormentOfHailfireLifeLoss is the printed 3.
const tormentOfHailfireLifeLoss = 3

// tormentOfHailfireStart builds the whole run of questions — X
// repetitions × the opponents, in APNAP order within each repetition —
// and asks the first one. Everything after it is queued by an answer.
//
// Caller holds g.mu.
func tormentOfHailfireStart(ctx *Context) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	opponents := ctx.Opponents()
	if len(opponents) == 0 {
		return nil
	}
	queue := make([]uuid.UUID, 0, x*len(opponents))
	for i := 0; i < x; i++ {
		queue = append(queue, opponents...)
	}
	return tormentOfHailfireStep(ctx, queue)
}

// tormentOfHailfireStep asks the first question still owed and hands
// the rest of the queue to its answer.
//
// A package-level function over a frozen slice of seat IDs: no *Game
// and no player pointer is captured, so an undo that replays an answer
// resolves it against the restored game (the delayed-trigger contract,
// AGENTS.md §7).
//
// Caller holds g.mu.
func tormentOfHailfireStep(ctx *Context, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim := remaining[0]
		rest := remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(victim)
		if p == nil || p.Eliminated {
			// A player who has left the game is skipped rather than
			// prompted (CR 800.4a); the rest of the run continues.
			remaining = rest
			continue
		}
		options, branches := tormentOfHailfireOptions(ctx, victim)
		queued := PickOption{
			Player:   victim,
			Question: "Torment of Hailfire — lose " + strconv.Itoa(tormentOfHailfireLifeLoss) + " life, or...",
			Options:  options,
			Then:     tormentOfHailfireAnswered(victim, branches, rest),
		}
		return queued.Apply(ctx)
	}
	return nil
}

// tormentOfHailfireOptions builds this player's option list and the
// matching branch keys. The keys are plain strings rather than
// closures so the two lists cannot drift: the index the chooser
// answers with indexes both.
//
// Caller holds g.mu.
func tormentOfHailfireOptions(ctx *Context, victim uuid.UUID) ([]game.ChoiceOption, []string) {
	options := []game.ChoiceOption{{
		Label:    "Lose " + strconv.Itoa(tormentOfHailfireLifeLoss) + " life",
		LifeCost: tormentOfHailfireLifeLoss,
	}}
	branches := []string{tormentBranchLife}

	if len(NonlandPermanentsControlledBy(ctx.Game, victim)) > 0 {
		options = append(options, game.ChoiceOption{Label: "Sacrifice a nonland permanent"})
		branches = append(branches, tormentBranchSacrifice)
	}
	if p := ctx.Game.PlayerByIDForEffect(victim); p != nil && p.Hand != nil && p.Hand.Size() > 0 {
		options = append(options, game.ChoiceOption{Label: "Discard a card"})
		branches = append(branches, tormentBranchDiscard)
	}
	return options, branches
}

// The three branch keys.
const (
	tormentBranchLife      = "life"
	tormentBranchSacrifice = "sacrifice"
	tormentBranchDiscard   = "discard"
)

// tormentOfHailfireAnswered runs the chosen branch and then the next
// question. Everything it captures is a scalar or a frozen slice.
func tormentOfHailfireAnswered(victim uuid.UUID, branches []string, rest []uuid.UUID) func(ctx *Context, index int) error {
	return func(ctx *Context, index int) error {
		next := func(ctx *Context) error { return tormentOfHailfireStep(ctx, rest) }
		if index < 0 || index >= len(branches) {
			// Nothing could be asked — the player left between the
			// question being built and it being queued. The rest of
			// the run still happens.
			return next(ctx)
		}
		switch branches[index] {
		case tormentBranchSacrifice:
			return SacrificeChoice{
				Player:     victim,
				Candidates: NonlandPermanentsControlledBy(ctx.Game, victim),
				Question:   "Torment of Hailfire — sacrifice a nonland permanent",
				Then:       next,
			}.Apply(ctx)
		case tormentBranchDiscard:
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   ctx.Source(),
				N:        1,
				Question: "Torment of Hailfire — discard a card",
				Then:     func(g *game.Game) error { return tormentOfHailfireStep(NewContext(g, ctx.Item), rest) },
			})
			return nil
		default:
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -tormentOfHailfireLifeLoss); err != nil {
				return err
			}
			return next(ctx)
		}
	}
}
