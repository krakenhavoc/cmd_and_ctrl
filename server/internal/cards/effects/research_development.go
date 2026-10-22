package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Research // Development — split card, oracle
// fa38394b-6ac4-460c-ae3a-54d50037715b (EDHREC rank 29699):
//
//	Research — Instant {G}{U}: "Shuffle up to four cards you own from
//	                            outside the game into your library."
//	Development — Instant {3}{U}{R}: "Create a 3/1 red Elemental
//	                            creature token unless any opponent has
//	                            you draw a card. Repeat this process two
//	                            more times."
//
// Both halves register as ADR 0034 faces (ADR 0103): Research under the
// bare oracle ID, Development under "#1".
//
// RESEARCH is the move the other way on the outside-the-game seam (ADR
// 0191): ShuffleFromOutside asks for up to four cards from the
// sideboard (the card database for a deck that seated none, ADR 0198),
// puts them in the library unrevealed and shuffles it. The library is
// shuffled even when no card is chosen (CR 701.24d).
//
// DEVELOPMENT runs three rounds. In each, the opponents are asked in
// turn order, starting with the one after the caster and going round
// the table (a seat that has left is skipped), whether to have the
// caster draw a card. The first to say yes ends the round with a draw;
// if every opponent says no, the caster creates the token. Each
// opponent's choice is the only thing that matters about it, so asking
// them one at a time and stopping at the first yes is the same outcome
// as asking them all, and it lets the choice ride one prompt. The
// question is a keyed option pick, so a table waiting on an opponent's
// answer is still a restore point; the round and the opponents still to
// be asked are plain data on it (Carry), one registered key per round.
// A round's draw runs its tail through a draw continuation, so the next
// round does not start until a dredge offer or an ordering prompt on
// the draw has been answered (CR 608.2c).
//
// No simplification.
func init() {
	const id = "fa38394b-6ac4-460c-ae3a-54d50037715b"
	Register(Spec{
		OracleID:     id,
		Name:         "Research",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ShuffleFromOutside{
				Max:      4,
				Question: "Research — shuffle up to four cards you own from outside the game into your library",
			}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Development",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return developmentAsk(ctx, 1, opponentsInTurnOrder(ctx.Game, ctx.Controller()))
		},
	})

	for round := 1; round <= developmentRounds; round++ {
		developmentAnswer[round-1] = OptionPickThen(
			"option-pick/development-opponent-round-"+string(rune('0'+round)),
			func(ctx *Context, r game.OptionPicked) error { return developmentAnswered(ctx, round, r) })
	}
	developmentAfterDraw = game.RegisterDrawThen("development-after-draw", func(g *game.Game, d game.DrawThen) error {
		ctx := NewContext(g, &game.StackItem{
			Kind:         game.StackItemSpell,
			Controller:   d.Player,
			Owner:        d.Player,
			SourceCardID: d.Source,
		})
		return developmentNextRound(ctx, d.N)
	})
}

// developmentRounds is "once, then repeat two more times".
const developmentRounds = 3

var (
	// developmentAnswer[i] is the continuation of round i+1's question.
	developmentAnswer [developmentRounds]game.OptionPickThen
	// developmentAfterDraw starts the next round once a round's draw has
	// finished; the DrawThen's N is the round that just ended.
	developmentAfterDraw game.DrawThenRef
)

// developmentAsk asks the first of `ask` who is still in the game
// whether to have the caster draw, with the rest riding the prompt. With
// none left to ask, every opponent has declined: the round makes the
// token.
func developmentAsk(ctx *Context, round int, ask []uuid.UUID) error {
	for len(ask) > 0 {
		opp := ask[0]
		ask = ask[1:]
		if p := ctx.PlayerByID(opp); p == nil || p.Eliminated {
			continue
		}
		name := "the caster"
		if p := ctx.PlayerByID(ctx.Controller()); p != nil && p.Name != "" {
			name = p.Name
		}
		return PickOption{
			Player:   opp,
			Question: "Development — " + name + " creates a 3/1 red Elemental creature token unless you have them draw a card",
			Options: []game.ChoiceOption{
				{Label: "Have " + name + " draw a card"},
				{Label: "Let " + name + " create a 3/1 red Elemental creature token"},
			},
			ThenKey: developmentAnswer[round-1],
			Carry:   ask,
		}.Apply(ctx)
	}
	if err := (CreateToken{Template: TokenCard("3/1 red Elemental"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return developmentNextRound(ctx, round)
}

// developmentAnswered is an opponent's answer in round `round`: the first
// option has the caster draw and ends the round; anything else (the
// second option, or an opponent who left before answering) asks the next
// opponent.
func developmentAnswered(ctx *Context, round int, r game.OptionPicked) error {
	if r.Option != nil && r.Index == 0 {
		caster := ctx.Controller()
		return ctx.Game.DrawNThenForEffect(caster, 1, game.DrawThen{
			Ref: developmentAfterDraw, Player: caster, Source: ctx.Source(), N: round,
		})
	}
	return developmentAsk(ctx, round, r.Carry)
}

// developmentNextRound starts the round after `round`, if there is one.
func developmentNextRound(ctx *Context, round int) error {
	if round >= developmentRounds {
		return nil
	}
	return developmentAsk(ctx, round+1, opponentsInTurnOrder(ctx.Game, ctx.Controller()))
}
