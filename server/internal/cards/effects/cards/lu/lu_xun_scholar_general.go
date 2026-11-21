package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lu Xun, Scholar General — "Horsemanship (This creature can't be blocked
// except by creatures with horsemanship.) Whenever Lu Xun deals damage to an
// opponent, you may draw a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea658352-abef-4201-b20c-f5c5809d1d3e",
		Name:            "Lu Xun, Scholar General",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"horsemanship"},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && ev.Amount > 0 && ev.Target != source.Controller && g.PlayerByIDForEffect(ev.Target) != nil
			}, "Lu Xun, Scholar General — draw a card", effects.Do(effects.DrawCards{N: 1})), "Lu Xun, Scholar General — draw a card?", 1),
		},
	})
}
