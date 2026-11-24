package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unnatural Selection — Enchantment {1}{U}:
//
//	"{1}: Choose a creature type other than Wall. Target creature becomes
//	 that type until end of turn."
//
// The type is chosen as the ability resolves (CR 608.2d, the Kindred spells'
// ChooseCreatureTypeTail), and Wall is not offered. "Becomes that type" sets
// the creature's creature types (CR 205.1a, 613.1d): it stops being any other
// type it had, a changeling's every creature type included, and keeps its card
// types and abilities. One record until end of turn per activation (CR 611.2c),
// pinned to the creature as the ability resolved (CR 400.7), so a target that
// has left the battlefield by the time the type is named is not affected.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0be10525-09dc-4970-b495-2db7d4d3c3e7",
		Name:         "Unnatural Selection",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Choose a creature type other than Wall. Target creature becomes that type until end of turn.",
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				g.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Unnatural Selection — choose a creature type other than Wall",
					Except:   []string{"Wall"},
					ThenTail: effects.WithCarry(ctx.ItemTail(unnaturalSelectionTypeChosen), id),
				})
				return nil
			},
		}},
	})
}

// unnaturalSelectionTypeChosen makes the carried creature the chosen type until
// end of turn. An on-disk key: never renamed, never reused.
var unnaturalSelectionTypeChosen = effects.CreatureTypeChosen("then/unnatural-selection-type-chosen", func(ctx *effects.Context, t string, o game.TailOutcome) error {
	target := effects.TailCarry0(o)
	if c, ok := ctx.Game.LookupCardForEffect(target); !ok || !effects.OnBattlefield(ctx.Game, c.InstanceID) {
		return nil
	}
	return effects.UntilEndOfTurn(ctx, target, nil, "Unnatural Selection — becomes a "+t+" until end of turn",
		game.SetCreatureTypesMod(t))
})
