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
// The answers are threaded through the chain as a fresh map per link
// rather than one shared map, so an undo back into the chain cannot
// see an answer from a branch that was undone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f95027a-5f04-48b4-99a0-8913be3a520a",
		Name:         "Selective Obliteration",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			selectiveObliterationAsk(ctx.Game, item, apnapPlayers(ctx.Game), map[uuid.UUID]string{})
			return nil
		},
	})
}

func selectiveObliterationAsk(g *game.Game, item *game.StackItem, order []uuid.UUID, chosen map[uuid.UUID]string) {
	if len(order) == 0 {
		// Every player has chosen.
		_ = ExileAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			if c.IsColorless() {
				return false
			}
			colors := c.EffectiveColors()
			pick, ok := chosen[c.Controller]
			return !(ok && len(colors) == 1 && colors[0] == pick)
		}}.Apply(NewContext(g, item))
		return
	}
	player := order[0]
	ChooseColorThen(g, player, item.SourceCardID, "Selective Obliteration — choose a color",
		func(g *game.Game, color string) error {
			next := make(map[uuid.UUID]string, len(chosen)+1)
			for k, v := range chosen {
				next[k] = v
			}
			next[player] = color
			selectiveObliterationAsk(g, item, order[1:], next)
			return nil
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
