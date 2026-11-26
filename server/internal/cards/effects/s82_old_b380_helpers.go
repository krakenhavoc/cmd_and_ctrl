package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB380TappedCreaturesOf counts the tapped creatures a player controls
// (Borrowing 100,000 Arrows), read through the layers as it resolves.
func S82OldB380TappedCreaturesOf(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && c.Tapped {
			n++
		}
	}
	return n
}

// S82OldB380PlayersWithTwoMoreLands is Surveyor's Scope's X: the number of
// players who control at least two more lands than `you`.
func S82OldB380PlayersWithTwoMoreLands(g *game.Game, you uuid.UUID) int {
	mine := B03LandsControlled(g, you)
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if B03LandsControlled(g, p.ID) >= mine+2 {
			n++
		}
	}
	return n
}

// S82OldB380SurveyorSearch is "Search your library for up to X basic land
// cards … Put those cards onto the battlefield, then shuffle." With X = 0 the
// search still happens and finds nothing, and the library is still shuffled.
func S82OldB380SurveyorSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := S82OldB380PlayersWithTwoMoreLands(g, item.Controller)
	if x <= 0 {
		return g.ShuffleLibraryForEffect(item.Controller)
	}
	return SearchLibrary{
		Player:    item.Controller,
		Source:    item.SourceCardID,
		Predicate: IsBasicLand,
		Dest:      game.ZoneBattlefield,
		Limit:     x,
		Shuffle:   true,
		Reason:    "Surveyor's Scope — up to X basic land cards onto the battlefield",
	}.Apply(ctx)
}

// S82OldB380DiscardHandsThenDraw is Incendiary Command's last bullet: each
// player discards all the cards in their hand, then draws that many cards.
// Every discard is done before any draw (the discards are simultaneous), then
// each player draws what they discarded.
func S82OldB380DiscardHandsThenDraw(ctx *Context) error {
	players := TablePlayers(ctx)
	counts := make([]int, len(players))
	for i, p := range players {
		n, err := DiscardWholeHand(ctx.Game, p)
		if err != nil {
			return err
		}
		counts[i] = n
	}
	for i, p := range players {
		if counts[i] <= 0 {
			continue
		}
		if err := (DrawCards{Player: p, N: counts[i]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// S82OldB380DivinerSpiritDraw is "you and that player each draw that many
// cards": the damage dealt, for the controller and for the player it hit.
func S82OldB380DivinerSpiritDraw(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	n := item.Trigger.Event.Amount
	if n <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
		return err
	}
	victim := item.Trigger.Event.Target
	if g.PlayerByIDForEffect(victim) == nil {
		return nil
	}
	return DrawCards{Player: victim, N: n}.Apply(ctx)
}
