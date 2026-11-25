package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// each_player_exiles_board_and_hand.go — #6265: "each player exiles X
// permanents they control and/or cards from their hand" (Descent into
// Madness).
//
// One count, filled from two zones together: each player picks exactly X
// objects (or every one they have, when they have fewer) from among the
// permanents they control and the cards in their hand, in any split. The
// picks are one keyed choose_cards run (ChooseCardsRunThenKeyedForEffect,
// the Ill-Gotten Gains shape): one prompt per player, asked in APNAP order
// (CR 101.4), each over that player's own permanents and hand, and nothing
// moves until the last player has answered. Then every chosen object is
// exiled in one simultaneous event (CR 701.13a). A choose_cards prompt
// carries each candidate's card view, so a battlefield permanent and a hand
// card sit in one picker; the hand cards are redacted from every other seat
// as any choose_cards over a hand is.

// EachPlayerExilesPermanentsAndOrHandCards asks every player still in the
// game, in APNAP order, to choose `n` permanents they control and/or cards
// in their hand (all of them when they have fewer), then exiles every
// chosen object together. A player with nothing to choose is not asked;
// `n` of zero or less asks nobody.
//
// Caller must hold g.mu in write mode (a resolution frame).
func EachPlayerExilesPermanentsAndOrHandCards(ctx *Context, n int) error {
	if n <= 0 {
		return nil
	}
	g := ctx.Game
	players := ApnapOrder(g, TablePlayers(ctx))
	name := "Exile"
	if src, ok := g.LookupCardForEffect(ctx.Source()); ok && src.Name != "" {
		name = src.Name
	}
	var legs []game.ChooseCardsPrompt
	for _, id := range players {
		cards := append(PermanentsControlledBy(g, id), HandCardIDs(g, id)...)
		k := min(n, len(cards))
		if k == 0 {
			continue
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    id,
			FromPlayer: id,
			Source:     ctx.Source(),
			Question:   fmt.Sprintf("%s — exile %d %s you control and/or %s from your hand", name, k, plural(k, "permanent", "permanents"), plural(k, "card", "cards")),
			Cards:      cards,
			Min:        k,
			Max:        k,
		})
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, eachPlayerExilesBoardAndHand, ctx.Source(), ctx.RunData(players))
	return err
}

// eachPlayerExilesBoardAndHand exiles every object the players chose, in
// one simultaneous event, once the last of them has chosen. An on-disk key.
var eachPlayerExilesBoardAndHand = RunThen("prompt-run/each-player-exiles-board-and-hand", func(ctx *Context, r game.RunDone) error {
	picks := r.Picks()
	var all []uuid.UUID
	for _, id := range r.Carry {
		all = append(all, picks.By(id)...)
	}
	if len(all) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(all, game.KeyedTail{})
})

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
