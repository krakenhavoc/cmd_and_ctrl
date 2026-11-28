package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b722_helpers.go — Urza's Saga card draw and library, S82 OLD batch
// 722 (#5624).

// S82OldB722TimeSpiral is Time Spiral's resolution: "Exile Time Spiral. Each
// player shuffles their hand and graveyard into their library, then draws seven
// cards. You untap up to six lands." The spell exiles itself as it resolves (so
// it is not among the graveyard cards it shuffles away), the shuffle and draw
// are Timetwister's, and the untap is the controller's pick of up to six lands.
func S82OldB722TimeSpiral(item *game.StackItem, ctx *Context) error {
	if err := ctx.Game.ExileCardForEffect(item.SourceCardID); err != nil {
		return err
	}
	if err := EchoOfEonsResolve(item, ctx); err != nil {
		return err
	}
	return UntapUpToLands{N: 6, Question: "Time Spiral — untap up to six lands"}.Apply(ctx)
}

// S82OldB722DiscardYourHand is "discard your hand" as a trigger's effect: every
// card, none chosen (Grafted Skullcap).
func S82OldB722DiscardYourHand(g *game.Game, item *game.StackItem) error {
	_, err := DiscardWholeHand(g, item.Controller)
	return err
}

// S82OldB722YourNontokenCreatureDied is "a nontoken creature you control
// dies": the creature left the battlefield for a graveyard under your control
// (Midnight Reaper's condition).
func S82OldB722YourNontokenCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	return ok && LeftUnderControlOf(ev, dead) == source.Controller && !IsToken(dead)
}

// S82OldB722SearchForTheSameName is Remembrance's effect: "you may search your
// library for a card with the same name as that creature, reveal it, put it
// into your hand, then shuffle". The name is the dead creature's as it last
// existed (CR 603.10), read from the trigger.
func S82OldB722SearchForTheSameName(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	name := item.Trigger.Object.Name
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return game.CardNameMatches(c, name) },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Optional:  true,
		Reason:    "Remembrance — you may search for a card named " + name,
	}.Apply(NewContext(g, item))
}
