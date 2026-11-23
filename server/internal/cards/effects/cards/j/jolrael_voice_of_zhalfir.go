package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jolrael, Voice of Zhalfir — Legendary Creature — Human Druid {2}{G}{U},
// 3/3 (March of the Machine: The Aftermath, #5476):
//
//	"At the beginning of combat on your turn, up to one target land you
//	 control becomes an X/X green and blue Bird creature with flying and
//	 haste until end of turn, where X is the number of cards in your hand.
//	 It's still a land.
//	 Whenever a land creature you control deals combat damage to a player,
//	 draw a card."
//
// The animation is Creeping Tar Pit's one continuous effect (CR 611.2a)
// pinned to the object (CR 400.7): creature type and Bird subtype
// (layer 4), green and blue (layer 5), flying and haste (layer 6), base
// power and toughness X (layer 7b), all at one timestamp (CR 613.7). X is
// counted once as the ability resolves and stays fixed (CR 611.2c), so
// drawing or casting afterwards does not change it. The type is added to
// the land's, so it is still a land. The target is chosen as the trigger
// goes on the stack (CR 603.3d) and checked again as it resolves
// (CR 608.2b). The draw is one trigger per land creature that connects,
// read through the layers at the moment it deals the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65e4fbd1-4935-495e-9897-bad5a0cfd935",
		Name:         "Jolrael, Voice of Zhalfir",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Jolrael, Voice of Zhalfir — up to one target land you control becomes an X/X green and blue Bird creature with flying and haste until end of turn",
					jolraelAnimate),
				effects.TargetPermanent("up to one target land you control", effects.Land(), effects.YouControl()).WithCount(0, 1)),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Source)
				return ok && c.IsLand() && c.IsCreature()
			}, "Jolrael, Voice of Zhalfir — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

func jolraelAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		x := 0
		if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
			x = len(p.Hand.Cards)
		}
		return effects.ScopedEffectFor{
			Target: t.ID,
			Mods: append([]game.Mod{
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(ctx.TextSubtypes("Bird")...),
				game.SetColorsMod(ctx.TextColors("G", "U")...),
				game.AddKeywordsMod("flying", "haste"),
			}, game.SetBasePTMods(x, x)...),
			Duration: effects.DurationUntilEndOfTurn(ctx),
			Label:    "Jolrael, Voice of Zhalfir — an X/X green and blue Bird creature with flying and haste until end of turn",
		}.Apply(ctx)
	}
	return nil
}
