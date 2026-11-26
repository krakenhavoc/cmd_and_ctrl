package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CopyTargetedSpellByItsController is "the controller of target instant or
// sorcery spell copies it. That player may choose new targets for the copy"
// (Meletis Charlatan): the copy belongs to the copied spell's controller
// (CR 707.10), who gets the CR 707.10c prompt. A spell that left the stack
// in response is nothing to copy (CR 608.2b).
func CopyTargetedSpellByItsController(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	id := item.Targets[0].ID
	spell := g.StackItemForEffect(id)
	if spell == nil {
		return nil
	}
	return CopySpell{
		StackID:          id,
		Controller:       spell.Controller,
		ChooseNewTargets: true,
	}.Apply(NewContext(g, item))
}

// PyxisOfPandemoniumExile is "each player exiles the top card of their library
// face down", each linked to the resolving Pyxis (CR 607.2a), starting with
// the active player.
func PyxisOfPandemoniumExile(g *game.Game, item *game.StackItem) error {
	var all []uuid.UUID
	for _, seat := range g.Seats {
		all = append(all, seat.ID)
	}
	for _, pid := range ApnapOrder(g, all) {
		p := g.PlayerByIDForEffect(pid)
		if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
			continue
		}
		top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		if _, err := g.ExileFaceDownLinkedForEffect(item.SourceCardID, pid, top); err != nil {
			return err
		}
	}
	return nil
}

// PyxisOfPandemoniumRelease is "each player turns face up all cards they own
// exiled with this artifact, then puts all permanent cards among them onto
// the battlefield": every linked card is turned face up, then the permanent
// ones enter together under their owners' control.
func PyxisOfPandemoniumRelease(g *game.Game, item *game.StackItem) error {
	ref, ok := NewContext(g, item).SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	var entries []game.BatchEntry
	for _, id := range ids {
		g.TurnExiledFaceUpForEffect(item.SourceCardID, id, "Pyxis of Pandemonium")
		if c, found := g.LookupCardForEffect(id); found && c.IsPermanent() {
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneExile})
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{PutByEnteringPlayer: true}, game.KeyedTail{})
}
