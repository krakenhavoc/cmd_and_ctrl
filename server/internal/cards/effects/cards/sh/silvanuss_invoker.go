package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silvanus's Invoker — Creature — Dragon Druid {2}{G}, 3/2 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Conjure Elemental — {8}: Untap target land you control. It becomes an 8/8
//	 Elemental creature with trample and haste until end of turn. It's still a
//	 land."
//
// The untap is the first instruction and the animation the second, both on the
// one land the ability targeted. The animation is a layer 4 creature type and
// Elemental subtype, a layer 6 trample and haste and a layer 7b base 8/8, all
// at one timestamp, pinned to that object (CR 611.2c, CR 400.7); "still a
// land" is the type add keeping the land type. A land that came under the
// player's control this turn can attack only with the haste it is given.
// "Conjure Elemental" is an ability word with no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a25fab1-cabb-44b9-8a69-01fc1fc46bd4",
		Name:         "Silvanus's Invoker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: Untap target land you control. It becomes an 8/8 Elemental creature with trample and haste until end of turn. It's still a land.",
			Cost:    effects.ManaCost("{8}"),
			Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.ScopedEffectFor{
						Target: t.ID,
						Mods: []game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
							game.SetBasePowerMod(8),
							game.SetBaseToughnessMod(8),
							game.AddKeywordsMod("trample", "haste"),
						},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Silvanus's Invoker — becomes an 8/8 Elemental creature with trample and haste until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
