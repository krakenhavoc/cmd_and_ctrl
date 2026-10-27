package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selective Obliteration — Sorcery {3}{C}{C}:
//
//	"Each player chooses a color. Then exile each permanent unless it's
//	 colorless or it's only the color its controller chose."
//
// "Each player chooses" is a chain of #742's resolution-time colour
// prompts in APNAP order (CR 101.4): the active player first, then the
// others in turn order, each seeing the earlier answers, which is how
// the choices are made at a real table. Each answer's continuation asks
// the next player; the last one exiles.
//
// A permanent survives if it is colourless, or if its colours are
// exactly the one colour its controller named — a white-blue permanent
// is exiled even under a player who chose white. A permanent whose
// controller made no choice (they left the game mid-chain) is exiled.
// The exile is one simultaneous event (ExileAllMatching).
//
// The chain is data (#4253): the players in asking order ride the colour
// prompt's tail as its Carry and the answers so far as its Words, one per
// player asked, so a table waiting on any player's colour is a restore
// point and an undo back into the chain sees only its own answers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "9f95027a-5f04-48b4-99a0-8913be3a520a",
		Name:     "Selective Obliteration",
		// ADR 0126 §6: a colorless permanent, or one only of the color its controller chose, stays.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepExile, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return selectiveObliterationAsk(ctx, apnapPlayers(ctx.Game), nil)
		},
	})
}

// selectiveObliterationAsk asks the next player in `order` (the one after
// the len(chosen) who have answered), or, once every player has answered,
// exiles. The exile's error is returned, so from the last answer's tail it
// reaches ResolveColorChoice, which logs it as EventEffectError instead of
// losing it.
func selectiveObliterationAsk(ctx *Context, order []uuid.UUID, chosen []string) error {
	if len(chosen) >= len(order) {
		// Every player has chosen.
		picks := make(map[uuid.UUID]string, len(order))
		for i, p := range order {
			picks[p] = chosen[i]
		}
		return ExileAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			if c.IsColorless() {
				return false
			}
			colors := c.EffectiveColors()
			pick, ok := picks[c.Controller]
			return !(ok && len(colors) == 1 && colors[0] == pick)
		}}.Apply(ctx)
	}
	then := ctx.ItemTail(selectiveObliterationChosen)
	then.Carry = append([]uuid.UUID(nil), order...)
	then.Words = append([]string(nil), chosen...)
	ChooseColorTail(game.ColorForBenefit, ctx.Game, order[len(chosen)], ctx.Source(), "Selective Obliteration — choose a color", then)
	return nil
}

// selectiveObliterationChosen records one player's colour (the tail's Word)
// and asks the next. An on-disk key, registered in init because its body
// asks the next player with it.
var selectiveObliterationChosen game.TailRef

func init() {
	selectiveObliterationChosen = CardTail("then/selective-obliteration-color-chosen", func(ctx *Context, o game.TailOutcome) error {
		chosen := append(append([]string(nil), o.Tail.Words...), o.Tail.Word)
		return selectiveObliterationAsk(ctx, o.Tail.Carry, chosen)
	})
}

// apnapPlayers lists the players still in the game in APNAP order
// (CR 101.4): the active player, then each other player in turn order.
func apnapPlayers(g *game.Game) []uuid.UUID {
	n := len(g.Seats)
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		p := g.Seats[(g.Turn.ActiveSeat+i)%n]
		if p == nil || p.Eliminated {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}
