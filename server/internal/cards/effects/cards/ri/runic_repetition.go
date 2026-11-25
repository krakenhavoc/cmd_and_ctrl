package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Runic Repetition — Sorcery {2}{U} (Commander 2019):
//
//	"Return target exiled card with flashback you own to your hand."
//
// Sorceress's Schemes' exile half: a card you own in exile that has flashback
// (the card's own printed flashback, read from the catalog, so one that was
// only granted flashback by Snapcaster Mage does not count). The legality is
// judged as the spell is cast and again as it resolves (CR 608.2b); the card
// goes to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ab99b5e-c9e0-45c2-8df2-864ab0ad1caf",
		Name:         "Runic Repetition",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: &game.TargetSpec{
			Mode:  "card_in_graveyard",
			Label: "target exiled card with flashback you own",
			Zones: []game.ZoneKind{game.ZoneExile},
			CardOK: func(_ *game.Game, caster uuid.UUID, c game.Card, zone game.ZoneKind) bool {
				return zone == game.ZoneExile && c.Owner == caster && game.AlternativeCostByKey(game.CatalogKey(c), "flashback") != nil
			},
			Min: 1, Max: 1,
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				_, err := ctx.Game.ReturnFromAnyZoneToHandForEffect(ctx.Controller(), ctx.Source(), t.ID)
				return err
			}
			return nil
		},
	})
}
