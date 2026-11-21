package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b08_helpers.go — shared pieces of S82 GAP batch 8 (Innistrad
// creatures and double-faced artifacts, #5948).

// gapB08APlayerHasNoCardsInHand is "a player has no cards in hand": any player
// still in the game whose hand is empty (Lupine Prototype).
func gapB08APlayerHasNoCardsInHand(g *game.Game) bool {
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.Hand != nil && p.Hand.Size() == 0 {
			return true
		}
	}
	return false
}

// gapB08TuckTheTargetCard is "put target card from your graveyard on the bottom
// of your library" (Epitaph Golem): the one target, if it is still legal as the
// ability resolves, goes to the bottom of its owner's library.
func gapB08TuckTheTargetCard(g *game.Game, item *game.StackItem) error {
	t, ok := firstLegalTarget(NewContext(g, item))
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return g.TuckToLibraryForEffect(t.ID, true)
}

// gapB08CondEachPlayerHasAtMostLife is "each player has N or less life" as an
// intervening-if Condition (Cryptolith Fragment): every player still in the game,
// read as the trigger goes on the stack and again as it resolves (CR 603.4). It
// declares no footprint, so its row keeps the trigger-order prompt.
type gapB08CondEachPlayerHasAtMostLife struct{ n int }

func (c gapB08CondEachPlayerHasAtMostLife) Holds(ctx *Context) bool {
	any := false
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		any = true
		if p.Life > c.n {
			return false
		}
	}
	return any
}

func (gapB08CondEachPlayerHasAtMostLife) condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
