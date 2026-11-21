package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB34UntapUpToYourLands is "untap up to N lands you control": the
// controller chooses, as the effect resolves, up to N of the tapped lands
// they control (no target, CR 115.1). It is UntapUpToLands' prompt and keyed
// answer narrowed to the controller's own lands.
func GapB34UntapUpToYourLands(ctx *Context, n int, question string) error {
	var tapped []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Tapped && c.Controller == ctx.Controller() {
			tapped = append(tapped, c.InstanceID)
		}
	}
	if len(tapped) == 0 || n <= 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: question,
		Cards:    tapped,
		Min:      0,
		Max:      n,
		ThenKey:  untapLandsChosen,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// gapB34DrawEvents counts the draws a player has made so far this game, as
// the engine logged them (one EventDrawCard per card that arrived).
func GapB34DrawEvents(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventDrawCard && ev.Actor == player {
			n++
		}
	}
	return n
}

// gapB34TableInTurnOrder is every player still in the game, the active
// player first and then in turn order (APNAP, CR 101.4).
func GapB34TableInTurnOrder(ctx *Context) []uuid.UUID {
	g := ctx.Game
	var out []uuid.UUID
	for i := range g.Seats {
		p := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
		if p != nil && !p.Eliminated {
			out = append(out, p.ID)
		}
	}
	return out
}
