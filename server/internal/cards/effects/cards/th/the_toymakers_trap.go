package cardsth

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Toymaker's Trap — Enchantment {2}{B} (Doctor Who):
//
//	"At the beginning of your upkeep, secretly choose a number between 1
//	 and 5 that hasn't been chosen. If you do, an opponent guesses which
//	 number you chose, then you reveal the number you chose. If they
//	 guessed wrong, they lose life equal to the number they guessed and
//	 you draw a card. If they guessed right, sacrifice this enchantment."
//
// The numbers already chosen are noted on this enchantment (Card.Notes,
// #6097), so "hasn't been chosen" spans its upkeeps and a Trap that
// leaves and returns is a new object with none chosen (CR 400.7). With
// all five chosen there is nothing to choose, and nothing else happens.
//
// Otherwise the controller names the opponent who guesses (only when
// there is more than one), then a secret choice round (#5948, ADR 0180
// amendment 2026-11-24) asks the controller for a number from those not
// yet chosen and then the opponent for a guess from 1 to 5. The guess is
// made without seeing the number; both are revealed together. A wrong
// guess costs the guesser that much life and draws the controller a card;
// a right one sacrifices the Trap (only while its controller still
// controls it, CR 701.21a). Choosing the opponent first changes nothing
// anyone can learn: the number is the controller's own secret either way.
//
// A bot hides the highest number left and guesses the lowest one left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e93ea21-5e3f-448e-a175-96c8438c328e",
		Name:         "The Toymaker's Trap",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("The Toymaker's Trap — secretly choose a number between 1 and 5 that hasn't been chosen; an opponent guesses it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if len(toymakersTrapUnchosen(g, ctx.Source())) == 0 {
						return nil
					}
					return effects.ChoosePlayer{
						Among:    effects.Opponents,
						Question: "The Toymaker's Trap — choose an opponent to guess the number",
						ThenKey:  toymakersTrapGuesser,
					}.Apply(ctx)
				}),
		},
	})
}

// toymakersTrapUnchosen is the numbers 1–5 not yet noted for the Trap
// `id`, ascending.
func toymakersTrapUnchosen(g *game.Game, id uuid.UUID) []int {
	notes := g.NotesForEffect(id)
	var out []int
	for n := 1; n <= 5; n++ {
		if !game.NotedIn(notes, strconv.Itoa(n)) {
			out = append(out, n)
		}
	}
	return out
}

// toymakersTrapGuesser starts the round once the guesser is named. An
// on-disk key.
var toymakersTrapGuesser = effects.ChoosePlayerThen("option-pick/toymakers-trap-guesser", func(ctx *effects.Context, opp uuid.UUID, _ []uuid.UUID) error {
	unchosen := toymakersTrapUnchosen(ctx.Game, ctx.Source())
	if len(unchosen) == 0 {
		return nil
	}
	allowed := make([]int, len(unchosen))
	hide := make([]int, 5)
	guess := make([]int, 5)
	for i, n := range unchosen {
		allowed[i] = n - 1
		hide[n-1] = n
		guess[n-1] = 10 - n
	}
	asks := []game.VoteAsk{{
		Voter:    ctx.Controller(),
		Allowed:  allowed,
		Question: "The Toymaker's Trap — secretly choose a number between 1 and 5 that hasn't been chosen",
	}}
	if opp != uuid.Nil {
		asks = append(asks, game.VoteAsk{
			Voter:    opp,
			Question: "The Toymaker's Trap — guess which number was chosen (a wrong guess loses you that much life)",
		})
	}
	var carry []uuid.UUID
	if opp != uuid.Nil {
		// #6175: an opponent guesses the number.
		effects.NotePileOrGuess(ctx)
		carry = []uuid.UUID{opp}
	}
	return effects.SecretChoices{
		Question:      "The Toymaker's Trap",
		Words:         effects.Numbers(1, 5),
		ForController: hide,
		ForOpponents:  guess,
		Asks:          asks,
		Then:          toymakersTrapRevealed,
		Carry:         carry,
	}.Apply(ctx)
})

// toymakersTrapRevealed settles the guess. An on-disk key.
var toymakersTrapRevealed = effects.VoteResultThen("secret/toymakers-trap", func(ctx *effects.Context, r game.VoteResult) error {
	chosen, ok := r.AnswerTo(0)
	if !ok {
		return nil
	}
	number := chosen.Option + 1
	ctx.Game.AddNoteForEffect(ctx.Source(), strconv.Itoa(number))
	guessed, ok := r.AnswerTo(1)
	if !ok || len(r.Carry) == 0 || guessed.Voter != r.Carry[0] {
		return nil
	}
	guess := guessed.Option + 1
	if guess != number {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), guessed.Voter, -guess); err != nil {
			return err
		}
		return effects.DrawCards{N: 1}.Apply(ctx)
	}
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || !effects.OnBattlefield(ctx.Game, ctx.Source()) || src.Controller != ctx.Controller() {
		return nil
	}
	return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})
