package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Koth of the Hammer — Legendary Planeswalker — Koth {2}{R}{R}, loyalty 3:
//
//	"+1: Untap target Mountain. It becomes a 4/4 red Elemental creature
//	 until end of turn. It's still a land.
//	 −2: Add {R} for each Mountain you control.
//	 −5: You get an emblem with "Mountains you control have '{T}: This land
//	 deals 1 damage to any target.'""
//
// The emblem is Liliana of the Dark Realms' shape: a bundle on the card, a
// layer-6 grant in the emblem's statics. "Mountain" is a word of the text
// (CR 612.2), read off the source's text reader.
//
// No simplification.
func init() {
	const grant = "koth-of-the-hammer/ping"
	const emblemText = "Mountains you control have \"{T}: This land deals 1 damage to any target.\""
	effects.Register(effects.Spec{
		OracleID:        "589186e0-3ebe-46a3-a9e7-798b1712fa93",
		Name:            "Koth of the Hammer",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This land deals 1 damage to any target.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			}},
			Text: "{T}: This land deals 1 damage to any target.",
		}},
		Emblem: &effects.EmblemSpec{
			Label: "Koth of the Hammer emblem",
			Text:  emblemText,
			Static: []game.StaticAbility{
				effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Mountain"))
				}, grant),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Untap target Mountain. It becomes a 4/4 red Elemental creature until end of turn. It's still a land.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPermanent("target Mountain", effects.HasSubtype("Mountain")),
				Effect:  kothOfTheHammerAnimate,
			},
			{
				Label: "−2: Add {R} for each Mountain you control.",
				Cost:  effects.LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.S82OneB20LandsControlled(g, item.Controller, ctx.TextSubtype("Mountain"))
					return effects.AddMana{Amount: effects.ManaN(n, "{R}")}.Apply(ctx)
				},
			},
			{
				Label: "−5: You get an emblem with \"" + emblemText + "\"",
				Cost:  effects.LoyaltyCost(-5),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

func kothOfTheHammerAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return effects.ScopedEffectFor{
			Target: t.ID,
			Mods: []game.Mod{
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
				game.SetBasePowerMod(4),
				game.SetBaseToughnessMod(4),
				game.SetColorsMod(ctx.TextColor("R")),
			},
			Duration: effects.DurationUntilEndOfTurn(ctx),
			Label:    "Koth of the Hammer — becomes a 4/4 red Elemental creature until end of turn",
		}.Apply(ctx)
	}
	return nil
}
