package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b7_helpers.go — Star Trek, batch 7 (#4973): the pieces the
// batch's cards share, and the ones a later Star Trek card will want.

// YouFacedADilemma — "whenever you face a dilemma" (Star Trek): the
// source's controller chose one or more modes for a spell or ability
// (game.EventFacedDilemma). Watch game.EventFacedDilemma.
func YouFacedADilemma(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventFacedDilemma && ev.Actor != uuid.Nil && ev.Actor == source.Controller
}

// trekB7CreatureTypesAmongNonBorgYouControl is Hoshi Sato's X: the
// number of distinct creature types among the creatures the resolving
// item's controller controls that are not Borg. A changeling has every
// creature type (and so is a Borg, and counts for nothing). "Borg" is a
// word of the source's text (CR 612.2), read through its lens while the
// source is still on the battlefield.
func trekB7CreatureTypesAmongNonBorgYouControl(g *game.Game, item *game.StackItem) int {
	var reader *game.Card
	if src, ok := g.LookupCardForEffect(item.SourceCardID); ok && onBattlefield(g, item.SourceCardID) {
		reader = &src
	}
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller || !c.IsCreature() {
			continue
		}
		if game.ThroughTextOf(reader, c).HasSubtypeWord("Borg") {
			continue
		}
		for _, t := range game.CreatureTypesOf(&c) {
			seen[t] = true
		}
	}
	return len(seen)
}

// trekB7YouControlA is "you control a <creature type>" at resolution:
// a permanent the resolving item's controller controls with that
// subtype in the layered view (a changeling has every creature type).
// The type is a word of the source's text (CR 612.2), read through its
// lens while the source is on the battlefield, as its last-known text
// is not kept.
func trekB7YouControlA(g *game.Game, item *game.StackItem, subtype string) bool {
	var reader *game.Card
	if src, ok := g.LookupCardForEffect(item.SourceCardID); ok && onBattlefield(g, item.SourceCardID) {
		reader = &src
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && game.ThroughTextOf(reader, c).HasSubtypeWord(subtype) {
			return true
		}
	}
	return false
}

// createTokenAttachedToSource is "create [an Equipment token] … Attach
// it to [this creature]" (U.S.Agent's Sturdy Shield, Gwyndala's
// Neuroflux Heirloom): one token of `template` is created through the
// keyed creation, and the attach-to-host tail attaches it to the
// ability's source. A source that has left the battlefield, or come back
// as a new object (CR 400.7), gets nothing attached and the token stays
// unattached (CR 701.3b).
func createTokenAttachedToSource(g *game.Game, item *game.StackItem, template game.Card) error {
	if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
		return CreateToken{Template: template, N: 1}.Apply(NewContext(g, item))
	}
	ctx := NewContext(g, item)
	tail := ctx.Tail(roleAttachToHost)
	tail.Carry = []uuid.UUID{item.SourceCardID}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(template),
			Count:    1,
		}},
	}, tail)
}
