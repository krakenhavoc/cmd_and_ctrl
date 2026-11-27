package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b718_helpers.go — shared pieces of the S82 OLD batch 718 cards
// (Urza's Legacy, #5623).

// OldB718ControlsMoreLandsThanEachOtherPlayer is "that player controls more
// lands than each other player" (Rivalry): strictly more than every other
// player still in the game. A player alone at the table has no other player
// to beat and passes.
func OldB718ControlsMoreLandsThanEachOtherPlayer(g *game.Game, player uuid.UUID) bool {
	lands := map[uuid.UUID]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() {
			lands[c.Controller]++
		}
	}
	for _, p := range LivePlayers(g) {
		if p.ID != player && lands[p.ID] >= lands[player] {
			return false
		}
	}
	return true
}

// OldB718SpellYouCastWasCountered is "whenever a spell you've cast is
// countered": the countered object is a spell the source's controller cast
// (its latest cast this turn), not a copy of one and not an ability.
func OldB718SpellYouCastWasCountered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCounterSpell || ev.Target != ev.CardID || ev.CardID == uuid.Nil {
		return false
	}
	log := g.EventsThisTurn()
	for i := len(log) - 1; i >= 0; i-- {
		if log[i].Kind == game.EventCast && log[i].CardID == ev.CardID {
			return log[i].Actor == source.Controller
		}
	}
	return false
}

// OldB718TuckEachOwnerOrdersTop is "each owner puts the cards they own among
// them on top of their library in any order" after a tuck to the top
// (Harmonic Convergence, CR 401.4): an owner with one card has nothing to
// order, and each owner with more is asked to arrange their own.
var oldB718TuckEachOwnerOrdersTop = CardTail("tuck/old-b718-each-owner-orders-top", func(ctx *Context, o game.TailOutcome) error {
	byOwner := map[uuid.UUID][]uuid.UUID{}
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			byOwner[c.Owner] = append(byOwner[c.Owner], id)
		}
	}
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated || len(byOwner[p.ID]) < 2 {
			continue
		}
		if err := (PutInLibraryInAnyOrder{
			Chooser:   p.ID,
			Cards:     byOwner[p.ID],
			From:      game.ZoneLibrary,
			Placement: game.LibraryPlaceTop,
			Label:     o.Tail.Word,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// OldB718TuckOnTopByOwner puts every card in ids on top of its owner's
// library as one simultaneous exit, from any zone, then asks each owner with
// more than one to order theirs.
func OldB718TuckOnTopByOwner(ctx *Context, ids []uuid.UUID, label string) error {
	tail := WithWord(ctx.ItemTail(oldB718TuckEachOwnerOrdersTop), label)
	return ctx.Game.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{}, tail)
}

// OldB718UntapTheTarget is "Untap target <permanent>" as a spell's whole
// OnResolve (Burst of Energy): the first target still legal as it resolves.
func OldB718UntapTheTarget(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return UntapTarget{Target: id}.Apply(ctx)
}

// OldB718AnyNumberOfYourCreatures is the candidate list of "Sacrifice any
// number of creatures" (Last-Ditch Effort): the chooser's creatures, with no
// floor and no ceiling.
func OldB718AnyNumberOfYourCreatures(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out, 0, 0
}
