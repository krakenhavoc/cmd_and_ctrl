package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Halsin, Emerald Archdruid — Legendary Creature — Elf Druid {3}{G}, 2/4
// (Commander Legends: Battle for Baldur's Gate):
//
//	"{1}: Until end of turn, target token you control becomes a green Bear
//	 creature with base power and toughness 4/4 in addition to its other
//	 colors and types.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// The animation is Shadow Puppeteers': a layer 4 subtype add, a layer 5 colour
// add and a layer 7b base power and toughness set, all at one timestamp and
// pinned to the one object the ability targeted (CR 611.2c, CR 400.7); counters
// and other effects still apply on top. Choose a Background is a deck-building
// rule the deck validator applies (deck/partner.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01ab3603-6c2f-4d65-9bb2-da95efdbe622",
		Name:         "Halsin, Emerald Archdruid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Until end of turn, target token you control becomes a green Bear creature with base power and toughness 4/4 in addition to its other colors and types.",
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetPermanent("target token you control", effects.IsTokenPredicate(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.ScopedEffectFor{
						Target: t.ID,
						Mods: []game.Mod{
							game.SetBasePowerMod(4),
							game.SetBaseToughnessMod(4),
							game.AddColorsMod(ctx.TextColors("G")...),
							game.AddSubtypesMod(ctx.TextSubtypes("Bear")...),
						},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Halsin, Emerald Archdruid — becomes a green Bear with base power and toughness 4/4 until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
