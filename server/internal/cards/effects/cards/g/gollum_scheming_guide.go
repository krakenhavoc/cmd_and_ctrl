package cardsg

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gollum, Scheming Guide — Legendary Creature — Halfling Horror 2/1 {1}{B}
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"Whenever Gollum attacks, look at the top two cards of your library,
//	 put them back in any order, then choose land or nonland. An opponent
//	 guesses whether the top card of your library is the chosen kind.
//	 Reveal that card. If they guessed right, remove Gollum from combat.
//	 Otherwise, you draw a card and Gollum can't be blocked this turn."
//
// The look and the reorder are Portent's (LookAtLibraryThenPlace). Then
// the controller names the opponent who guesses (only when there is more
// than one), and a secret choice round (#5948, ADR 0180 amendment
// 2026-11-24) asks the controller for land or nonland and the opponent,
// who has not seen the library, whether the top card is that kind. Both
// answers are revealed together with the top card. A right guess removes
// Gollum from combat; anything else (a wrong guess, or no guess because
// the guesser left) draws the controller a card and makes Gollum
// unblockable this turn. Naming the guesser before the kind changes
// nothing anyone can learn: the kind is the controller's own secret.
//
// The bot hints are flat on purpose: they ride the wire to every seat,
// and a hint built from the top card would tell the guesser what it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fb5c9a6-3fab-4109-875b-83c92518c5eb",
		Name:         "Gollum, Scheming Guide",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Gollum, Scheming Guide — look at the top two cards of your library, then an opponent guesses the top card's kind",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.LookAtLibraryThenPlace{
						N:         2,
						Placement: game.LibraryPlaceTop,
						Label:     "Gollum, Scheming Guide — put them back in any order",
						ThenTail:  ctx.ItemTail(gollumSchemingGuideLooked),
					}.Apply(ctx)
				}),
		},
	})
}

// The round's words: the controller's two kinds, then the guesser's two
// answers.
const (
	gollumLand = iota
	gollumNonland
	gollumYes
	gollumNo
)

// gollumSchemingGuideLooked names the guesser once the cards are back. An
// on-disk key.
var gollumSchemingGuideLooked = effects.CardTail("gollum-scheming-guide/looked", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Gollum, Scheming Guide — choose an opponent to guess",
		ThenKey:  gollumSchemingGuideGuesser,
	}.Apply(ctx)
})

// gollumSchemingGuideGuesser starts the round. An on-disk key.
var gollumSchemingGuideGuesser = effects.ChoosePlayerThen("option-pick/gollum-scheming-guide-guesser", func(ctx *effects.Context, opp uuid.UUID, _ []uuid.UUID) error {
	asks := []game.VoteAsk{{
		Voter:    ctx.Controller(),
		Allowed:  []int{gollumLand, gollumNonland},
		Question: "Gollum, Scheming Guide — choose land or nonland",
	}}
	var carry []uuid.UUID
	if opp != uuid.Nil {
		// #6175: an opponent guesses. With none left nobody does.
		effects.NotePileOrGuess(ctx)
		asks = append(asks, game.VoteAsk{
			Voter:    opp,
			Allowed:  []int{gollumYes, gollumNo},
			Question: "Gollum, Scheming Guide — guess: is the top card of their library the chosen kind (land or nonland)?",
		})
		carry = []uuid.UUID{opp}
	}
	return effects.SecretChoices{
		Question: "Gollum, Scheming Guide",
		Words:    []string{"Land", "Nonland", "It is the chosen kind", "It isn't the chosen kind"},
		Asks:     asks,
		Then:     gollumSchemingGuideRevealed,
		Carry:    carry,
	}.Apply(ctx)
})

// gollumSchemingGuideRevealed reveals the top card and settles the guess.
// An on-disk key.
var gollumSchemingGuideRevealed = effects.VoteResultThen("secret/gollum-scheming-guide", func(ctx *effects.Context, r game.VoteResult) error {
	g, gollum := ctx.Game, ctx.Source()
	kind, chose := r.AnswerTo(0)
	guess, guessed := r.AnswerTo(1)
	top := g.RevealTopOfLibraryForEffect(ctx.Controller(), gollum, 1, "Gollum, Scheming Guide")
	isKind := false
	if chose && len(top) == 1 {
		if c, ok := g.LookupCardForEffect(top[0]); ok {
			isKind = c.IsLand() == (kind.Option == gollumLand)
		}
	}
	right := chose && guessed && len(r.Carry) == 1 && guess.Voter == r.Carry[0] &&
		(guess.Option == gollumYes) == isKind
	if right {
		if !effects.OnBattlefield(g, gollum) {
			return nil
		}
		if err := g.RemoveFromCombatForEffect(gollum); err != nil && !errors.Is(err, game.ErrNotInCombat) {
			return err
		}
		return nil
	}
	if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	if !effects.OnBattlefield(g, gollum) {
		return nil
	}
	return effects.RestrictUntilEOT{Target: gollum, Restrictions: game.CantBeBlocked, Label: "Gollum, Scheming Guide — can't be blocked this turn"}.Apply(ctx)
})
