package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidden Stag — Enchantment {1}{G}:
//
//	"Whenever an opponent plays a land, if this permanent is an enchantment, it
//	 becomes a 3/2 Elk Beast creature.
//	 Whenever you play a land, if this permanent is a creature, it becomes an
//	 enchantment."
//
// Two triggers, each with an intervening "if" (CR 603.4). "Becomes an
// enchantment" sets the card type again (CR 205.1a), so it stops being a
// creature and loses the Elk Beast subtypes; its later effect wins over the
// earlier one by timestamp (CR 613.7), so the cycle can repeat.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70c32155-dc07-45cb-bceb-e19d3ebc915d",
		Name:         "Hidden Stag",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return source.HasCardType("enchantment") && effects.OldB720AnOpponentPlayedALand(nil)(ev, source, ch, g)
			}, "Hidden Stag — it becomes a 3/2 Elk Beast creature",
				effects.OldB717BecomesACreature([]string{"Elk", "Beast"}, 3, 2, "Hidden Stag — a 3/2 Elk Beast creature")),
			effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source.IsCreature() && effects.B20LandPlayed(ev, g) && ev.Actor == source.Controller
			}, "Hidden Stag — it becomes an enchantment", hiddenStagBecomesAnEnchantment),
		},
	})
}

func hiddenStagBecomesAnEnchantment(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	if !effects.OnBattlefield(g, self) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(self)
	if !ok || !c.IsCreature() {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   self,
		Mods:     []game.Mod{game.SetTypesMod([]string{"Enchantment"}, nil)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Hidden Stag — an enchantment",
	}.Apply(effects.NewContext(g, item))
}
