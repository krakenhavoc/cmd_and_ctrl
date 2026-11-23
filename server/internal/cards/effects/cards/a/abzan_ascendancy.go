package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abzan Ascendancy — Enchantment {W}{B}{G} (S82 GAP, #5948):
//
//	"When this enchantment enters, put a +1/+1 counter on each creature you
//	 control.
//	 Whenever a nontoken creature you control dies, create a 1/1 white Spirit
//	 creature token with flying."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b7d4206-0d88-407b-ace5-d5d28881bf95",
		Name:         "Abzan Ascendancy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Abzan Ascendancy — put a +1/+1 counter on each creature you control",
				effects.B11PutCounterOnEachCreatureYouControl),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B19AnotherNontokenCreatureYouControlDied(ev, source, g)
			},
				"Abzan Ascendancy — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
	})
}
