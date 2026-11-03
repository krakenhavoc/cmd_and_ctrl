package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// TargetPlayerShufflesTheirCards is "target player shuffles up to N target
// cards from their graveyard into their library" (#4862), over one target
// group: the player picked in playerSlot and the cards picked in cardSlot,
// a FromTargetPlayersGraveyard clause.
//
// The rulings decide the edges. An illegal player performs no action, so
// nothing moves and nothing is shuffled, even if the cards are still legal
// (Memory's Journey). A legal player shuffles their library even when no
// card was picked or every picked card has left the graveyard (Quandrix
// Command, Gaea's Blessing). A card that has left the graveyard is an
// illegal target and stays where it is (CR 608.2b).
func TargetPlayerShufflesTheirCards(ctx *Context, group []game.TargetRef, playerSlot, cardSlot int) error {
	var player game.TargetRef
	found := false
	for _, t := range group {
		if t.Slot == playerSlot && t.Kind == game.TargetPlayer {
			player, found = t, true
			break
		}
	}
	if !found || !ctx.IsTargetLegal(player) {
		return nil
	}
	for _, t := range group {
		if t.Slot != cardSlot || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
			return err
		}
	}
	return ShuffleLibrary{Player: player.ID}.Apply(ctx)
}
