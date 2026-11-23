package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tatyova, Steward of Tides — Legendary Creature — Merfolk Druid
// {G}{G}{U}, 3/3 (Dominaria United):
//
//	"Land creatures you control have flying.
//	 Whenever a land you control enters, if you control seven or more
//	 lands, up to one target land you control becomes a 3/3 Elemental
//	 creature with haste. It's still a land."
//
// The first line is Blossoming Tortoise's land-creature filter (read
// through the layers, so an animated land counts) granting flying
// (KeywordGrant). The landfall trigger carries a CR 603.4 intervening
// "if": the seven lands, counting the one that just entered, are checked
// as it triggers and again as it resolves (InterveningIf). The animation
// has no duration, so it is one indefinite continuous effect pinned to the
// land (CR 611.2a): it stays a land, becomes a creature with its creature
// type replaced by Elemental (CR 205.1a), has base power and toughness 3/3
// and gains haste. "Up to one" lets the controller pick nothing.
//
// No simplification.
func init() {
	const label = "Tatyova, Steward of Tides — up to one target land you control becomes a 3/3 Elemental creature with haste"
	sevenLands := effects.CondBoard{
		Name: "you control seven or more lands",
		Test: func(ctx *effects.Context) bool {
			return effects.B02CountLandsControlledBy(ctx.Game, ctx.Controller()) >= 7
		},
	}
	animate := func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.ScopedEffectFor{
			Target: id,
			Mods: []game.Mod{
				game.AddTypesMod("Creature"),
				game.SetCreatureTypesMod(ctx.TextSubtypes("Elemental")...),
				game.SetBasePowerMod(3),
				game.SetBaseToughnessMod(3),
				game.AddKeywordsMod("haste"),
			},
			Duration: g.PinnedTo(game.IndefiniteDuration(), id),
			Label:    label + ". It's still a land.",
		}.Apply(ctx)
	}
	landfall := effects.On(game.EventETB, effects.AllOf(effects.LandEnteredUnderYourControl,
		func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return effects.B02CountLandsControlledBy(g, source.Controller) >= 7
		}), label, animate)
	effects.Register(effects.Spec{
		OracleID:     "7d5e734f-3efc-4124-b1d8-4fe3e9d3db28",
		Name:         "Tatyova, Steward of Tides",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.KeywordGrant(effects.B31LandCreaturesYouControl, "flying")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(sevenLands, landfall),
				effects.TargetPermanent("up to one target land you control", effects.And(effects.Land(), effects.YouControl())).WithCount(0, 1)),
		},
	})
}
