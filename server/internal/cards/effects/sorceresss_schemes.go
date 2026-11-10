package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sorceress's Schemes — Sorcery {3}{R} (Final Fantasy, #4462):
//
//	"Return target instant or sorcery card from your graveyard or exiled
//	 card with flashback you own to your hand. Add {R}.
//	 Flashback {4}{R} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// One target clause over two zones: an instant or sorcery card in your
// graveyard, or a card you own in exile that has flashback (the card's own
// printed flashback, read from the catalog, so a card that was only granted
// flashback by Snapcaster Mage does not count). The legality is re-judged as
// the spell resolves (CR 608.2b): a target that left its zone is illegal and
// the spell does nothing, mana included. The card goes to its owner's hand
// from whichever zone it is in. The {R} is added as the spell resolves and
// empties at the end of the step (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "945e56b9-709f-4bb3-8a6b-0a04d9202a03",
		Name:             "Sorceress's Schemes",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{R}")},
		Targets: &game.TargetSpec{
			Mode:  "card_in_graveyard",
			Label: "target instant or sorcery card from your graveyard or exiled card with flashback you own",
			Zones: []game.ZoneKind{game.ZoneGraveyard, game.ZoneExile},
			CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, zone game.ZoneKind) bool {
				if c.Owner != caster {
					return false
				}
				switch zone {
				case game.ZoneGraveyard:
					return Or(Instant(), Sorcery())(g, caster, c)
				case game.ZoneExile:
					return game.AlternativeCostByKey(game.CatalogKey(c), "flashback") != nil
				}
				return false
			},
			Min: 1, Max: 1,
		},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					// The mana is owed whenever the spell resolved with its target
					// legal, wherever the card then goes.
					ctx.RecordStep(sorceresssSchemesResolved, []uuid.UUID{t.ID})
					_, err := ctx.Game.ReturnFromAnyZoneToHandForEffect(ctx.Controller(), ctx.Source(), t.ID)
					return err
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if len(ctx.StepRecord(sorceresssSchemesResolved).IDs) == 0 {
					return nil
				}
				return AddMana{Produced: "{R}"}.Apply(ctx)
			}),
		),
	})
}

// sorceresssSchemesResolved records that the spell resolved with a legal
// target, so the mana follows the return.
const sorceresssSchemesResolved = "sorceresss-schemes/resolved"
