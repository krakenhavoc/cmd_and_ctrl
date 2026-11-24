package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// secret_choices.go — the card-facing shape of a secret choice round
// (#5948 Q3, game/secret_choice_round.go, ADR 0180's amendment of
// 2026-11-24): "each opponent secretly chooses a number", "each player
// secretly chooses a creature they control and a creature they don't
// control", "secretly choose a number … an opponent guesses which". Each
// ask is one question to one player; every answer stays hidden from every
// other seat until the last one is in, then they are revealed together
// and Then runs with them (VoteResult.AnswerTo reads each ask's answer).
//
// The bot hints (ForController, ForOpponents) ride the wire to every seat:
// build them from what the table already knows, never from a hidden card.

// SecretChoices starts a secret choice round as the effect resolves.
type SecretChoices struct {
	// Question is every ask's header unless the ask has its own.
	Question string
	// Options are the choices option asks pick from; Words is the same
	// for words. Numbers(lo, hi) builds a list of numbers.
	Options []game.ChoiceOption
	Words   []string
	// ForController and ForOpponents are bot hints, one per option.
	ForController, ForOpponents []int
	// Asks are the questions, in the order they are put.
	Asks []game.VoteAsk
	// Then is the rest of the card (VoteResultThen), and Carry the IDs it
	// is handed.
	Then  game.VoteThen
	Carry []uuid.UUID
}

// Apply starts the round. With no ask to put, Then runs at once with no
// answers.
func (s SecretChoices) Apply(ctx *Context) error {
	opts := s.Options
	if len(s.Words) > 0 {
		opts = make([]game.ChoiceOption, len(s.Words))
		for i, w := range s.Words {
			opts[i] = game.ChoiceOption{Label: w}
		}
	}
	if len(s.Asks) == 0 {
		return ctx.Game.RunVoteThenForEffect(s.Then, game.VoteResult{
			Controller: ctx.Controller(), Source: ctx.Source(), Options: opts, Carry: s.Carry,
		})
	}
	return ctx.Game.StartSecretChoicesForEffect(game.SecretChoicePrompt{
		Controller:    ctx.Controller(),
		Source:        ctx.Source(),
		Question:      s.Question,
		Options:       opts,
		ForController: s.ForController,
		ForOpponents:  s.ForOpponents,
		Asks:          s.Asks,
		Then:          s.Then,
		Carry:         s.Carry,
	})
}

// Numbers is the words "lo" … "hi", for a choice of a number from a short
// printed range ("a number between 1 and 5"): option i is the number lo+i.
func Numbers(lo, hi int) []string {
	var out []string
	for n := lo; n <= hi; n++ {
		out = append(out, strconv.Itoa(n))
	}
	return out
}
