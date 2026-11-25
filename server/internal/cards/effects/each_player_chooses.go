package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// each_player_chooses.go — "As this creature enters, each player chooses
// war or peace." (Archangel of Strife, #6097).
//
// Each player's answer is recorded on the permanent as a player note
// (game.Card.PlayerNotes, game/player_notes.go), which a static reads
// for as long as the permanent stays: "Creatures controlled by players
// who chose war get +3/+0" is `source.PlayerNoted(c.Controller, "war")`
// in the static's AppliesTo. The players choose in APNAP order (CR 101.4:
// the active player first, then the others in turn order), one prompt
// after another, each keyed, so a table waiting on one is a restore point.
//
// It is an AsEnters hook, as ChooseOptionAsEnters is (game/choose_option.go
// explains why the as-enters choice is not a paused CR 614 replacement),
// and the answers land on the permanent as they come. A permanent that
// left before every player answered records nothing more (CR 400.7).

// EachPlayerChoosesAsEnters builds the Spec.AsEnters for "As this
// permanent enters, each player chooses <options[0]> or <options[1]>".
// `label` is the prompt header, normally the card's name; `options` are
// the words in printed order, and each is the note recorded for the
// player who chose it.
func EachPlayerChoosesAsEnters(label string, options ...string) func(*game.Card, *Context) error {
	words := append([]string(nil), options...)
	return func(card *game.Card, ctx *Context) error {
		first := card.Controller
		if p := ctx.Game.Seats[ctx.Game.Turn.ActiveSeat]; p != nil {
			first = p.ID
		}
		players := SeatsStartingWith(ctx.Game, first)
		askEachPlayerToChoose(ctx.Game, card.InstanceID, label+" — choose "+joinOr(words), words, players)
		return nil
	}
}

// askEachPlayerToChoose asks players[0], carrying the rest. The carry is
// [source, players...]; Words are [question, options...].
func askEachPlayerToChoose(g *game.Game, source uuid.UUID, question string, words []string, players []uuid.UUID) {
	for len(players) > 0 {
		opts := make([]game.ChoiceOption, len(words))
		for i, w := range words {
			opts[i] = game.ChoiceOption{Label: w}
		}
		if g.QueueOptionPickForEffect(game.OptionPickPrompt{
			Chooser:  players[0],
			Source:   source,
			Question: question,
			Options:  opts,
			ThenKey:  eachPlayerChose,
			Carry:    append([]uuid.UUID{source}, players...),
			Words:    append([]string{question}, words...),
		}) != uuid.Nil {
			return
		}
		// A player who has left is not asked; the next one is.
		players = players[1:]
	}
}

// eachPlayerChose records one answer and asks the next player. It is set
// by the registration below rather than by its own initializer, because
// the continuation queues the next question under the same key and a
// self-referencing initializer is an initialization cycle.
var eachPlayerChose game.OptionPickThen

var _ = registerEachPlayerChose()

func registerEachPlayerChose() bool {
	eachPlayerChose = game.RegisterOptionPickThen("option/each-player-chooses-as-enters", eachPlayerChoseAnswer)
	return true
}

func eachPlayerChoseAnswer(g *game.Game, r game.OptionPicked) error {
	if len(r.Carry) < 2 || len(r.Words) < 1 {
		return nil
	}
	source, players := r.Carry[0], r.Carry[1:]
	if r.Option != nil {
		g.SetPlayerNoteForEffect(source, players[0], r.Option.Label)
	}
	if !OnBattlefield(g, source) {
		return nil
	}
	askEachPlayerToChoose(g, source, r.Words[0], r.Words[1:], players[1:])
	return nil
}
