package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b13_helpers.go holds the shapes batches 13–15 of the plane cards
// (ADR 0216 slice 4, #4870, the gold-bordered Unknown Event planes) share.
// Append-only.

// b13CreaturesOf lists the creatures `seat` controls, in battlefield order:
// the candidate set behind "chooses a creature they control".
func b13CreaturesOf(g *game.Game, seat uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range PermanentsControlledBy(g, seat) {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
			out = append(out, id)
		}
	}
	return out
}

// b13PlayersFrom lists the players still in the game in turn order starting
// with `first` (the "each player" fan-out, active player first, CR 101.4).
func b13PlayersFrom(g *game.Game, first uuid.UUID) []uuid.UUID {
	out := []uuid.UUID{}
	if p := g.PlayerByIDForEffect(first); p != nil && !p.Eliminated {
		out = append(out, first)
	}
	return append(out, opponentsInTurnOrder(g, first)...)
}

// b13EachPlayerChoosesACreature is "each player chooses a creature they
// control": the seats with a creature are asked one at a time in turn order
// from `first` (ADR 0175's ordered collector), each answering for their own
// board, and `then` is handed every choice (RunDone.Picks). `data` is
// Context.RunData.
func b13EachPlayerChoosesACreature(ctx *Context, first uuid.UUID, question string, then game.PromptRunThen, carry []uuid.UUID) error {
	var legs []game.ChooseCardsPrompt
	for _, seat := range b13PlayersFrom(ctx.Game, first) {
		cards := b13CreaturesOf(ctx.Game, seat)
		if len(cards) == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser: seat, FromPlayer: seat, Source: ctx.Source(), Question: question,
			Cards: cards, Min: 1, Max: 1, Zone: game.ZoneBattlefield,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	return ctx.Game.ChoosePermanentsInOrderKeyedForEffect(legs, then, ctx.Source(), ctx.RunData(carry))
}
