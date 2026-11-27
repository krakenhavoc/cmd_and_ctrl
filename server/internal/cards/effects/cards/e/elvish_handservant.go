package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Handservant — Creature — Elf Warrior {G}, 1/1 (Lorwyn, S82 OLD,
// #5593):
//
//	"Whenever a player casts a Giant spell, you may put a +1/+1 counter on this
//	 creature."
//
// Witchstalker's cast trigger over any player's Giant spell (the type is a word
// of the card's text, CR 612.2), optional at resolution (CR 603.5). One counter
// for each such spell, and none once the Handservant has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16fb84e1-0bc7-469c-bd73-40747be56d9e",
		Name:         "Elvish Handservant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.S82OldB556AGiantSpellWasCast,
				"Elvish Handservant — you may put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				"Elvish Handservant — put a +1/+1 counter on it?"),
		},
	})
}
