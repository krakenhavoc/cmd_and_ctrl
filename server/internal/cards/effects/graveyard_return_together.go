package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_return_together.go — #1867. Every card that returns more
// than one card from a graveyard to the battlefield does it through
// ReturnFromGraveyardTogether, one simultaneous entry, rather than a
// loop over ReturnFromGraveyard. Append-only: the helpers below are
// the shapes those cards share (the clone gate's rule).

// ReturnFromGraveyardTogether puts several cards from graveyards onto
// the battlefield as ONE simultaneous entry (#1867) — Reveillark's "up
// to two target creature cards", Replenish's "all enchantment cards",
// Exhume's one card per player. Every card's CR 614 window runs against
// the board as it was before any of them entered (CR 614.12), all of
// them land, and only then is any announced, so each newcomer sees the
// others enter (CR 603.6a). A loop over ReturnFromGraveyard is not the
// same thing: it announces the first card before the second moves, so
// "whenever another creature enters" on the second never sees the
// first.
//
// Cards that are no longer in a graveyard (a target exiled in
// response, CR 608.2b), tokens and nonpermanent cards are skipped
// rather than refusing the whole entry, which is what the engine's
// batch door does with them.
//
// Controller is ReturnFromGraveyard's: zero is "under its owner's
// control", and ctx.Controller() is "under your control". Tapped is
// the effect's own "to the battlefield tapped", riding each card's
// entry event.
//
// Then, when set, is the rest of the effect, and it runs exactly once
// with the IDs that entered (a graveyard card keeps its ID) — after
// the last entry question has been answered when one was asked, so it
// must not be written on the line after Apply. Its context is rebuilt
// from the live *Game, the contract every continuation follows.
type ReturnFromGraveyardTogether struct {
	Targets    []uuid.UUID
	Controller uuid.UUID
	Tapped     bool
	Then       func(ctx *Context, entered []uuid.UUID) error
}

func (r ReturnFromGraveyardTogether) Apply(ctx *Context) error {
	entries := make([]game.BatchEntry, 0, len(r.Targets))
	for _, id := range r.Targets {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsToken() || !c.IsPermanent() {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneGraveyard})
	}
	var then func(*game.Game, []uuid.UUID) error
	if r.Then != nil {
		item, next := ctx.Item, r.Then
		then = func(g *game.Game, entered []uuid.UUID) error {
			return next(NewContext(g, item), entered)
		}
	}
	return ctx.Game.PutOntoBattlefieldTogetherThenForEffect(entries,
		game.ZoneEntryOptions{Controller: r.Controller, Tapped: r.Tapped}, then)
}

// legalTargetCardIDs is the spell's or ability's still-legal card
// targets (CR 608.2b), in announce order — the list a "return up to N
// target cards" hands to ReturnFromGraveyardTogether.
func legalTargetCardIDs(ctx *Context) []uuid.UUID {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

// graveyardCardIDs is the cards in `player`'s graveyard that match, read
// once before anything moves — "return all land cards from your
// graveyard". Nil when the player or graveyard is gone.
func graveyardCardIDs(ctx *Context, player uuid.UUID, match func(game.Card) bool) []uuid.UUID {
	p := ctx.PlayerByID(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if match(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// allGraveyardsCardIDs is graveyardCardIDs over every seat, in seat
// order — "all creature cards from all graveyards".
func allGraveyardsCardIDs(ctx *Context, match func(game.Card) bool) []uuid.UUID {
	var ids []uuid.UUID
	for _, p := range ctx.Game.Seats {
		if p == nil {
			continue
		}
		ids = append(ids, graveyardCardIDs(ctx, p.ID, match)...)
	}
	return ids
}
