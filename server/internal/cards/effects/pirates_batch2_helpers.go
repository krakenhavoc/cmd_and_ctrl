package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pirates_batch2_helpers.go — shared shapes for the second Pirates
// batch. Kept in their own file rather than appended to helpers.go,
// which every parallel card pass edits at the same anchor.

// IsPirateCard is the CardPredicate form of the Pirate check, for
// target clauses ("target Pirate you control"). Reads the printed
// type line for the same reason Corsair Captain's lord does: a
// predicate that calls Effective() inside a layer rebuild recurses
// into the computation being performed. Cost: a creature *turned
// into* a Pirate isn't a legal target.
func IsPirateCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return isPirate(c) }
}

// enteredUnderYourControl reports whether ev is an ETB for a card
// controlled by source's controller, and returns it. The `another`
// flag excludes the source itself.
func enteredUnderYourControl(ev game.Event, source *game.Card, g *game.Game, another bool) (game.Card, bool) {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return game.Card{}, false
	}
	if another && ev.CardID == source.InstanceID {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller {
		return game.Card{}, false
	}
	return c, true
}

// discardWholeHand discards every card in playerID's hand and
// returns how many left. Used by the "discard your hand" spells,
// where the choice of what to pitch doesn't exist — it's all of it,
// so the random-order discard primitive is exactly right.
func discardWholeHand(g *game.Game, playerID uuid.UUID) (int, error) {
	p := g.PlayerByIDForEffect(playerID)
	if p == nil {
		return 0, nil
	}
	n := p.Hand.Size()
	if n == 0 {
		return 0, nil
	}
	return n, g.DiscardRandomForEffect(playerID, n)
}

// tablePlayers returns the controller of `item` followed by their
// opponents — every seated player, in a stable order, for the
// "each player" spells.
func tablePlayers(ctx *Context) []uuid.UUID {
	out := []uuid.UUID{ctx.Controller()}
	return append(out, ctx.Opponents()...)
}

// untapUpToLands untaps up to n tapped lands playerID controls.
//
// Sandbox simplification: the card says "untap up to three lands",
// which is a choice. Untapping the player's own tapped lands is
// what that choice is for in every deck that plays these cards, and
// a prompt for it would be three clicks of ceremony. If a card ever
// makes the choice interesting — untapping an opponent's land, or
// choosing between land types — this needs a real picker.
func untapUpToLands(g *game.Game, playerID uuid.UUID, n int) error {
	for _, c := range g.BattlefieldCardsForEffect() {
		if n <= 0 {
			return nil
		}
		if c.Controller != playerID || !c.IsLand() || !c.Tapped {
			continue
		}
		if err := g.UntapTargetForEffect(c.InstanceID); err != nil {
			return err
		}
		n--
	}
	return nil
}
