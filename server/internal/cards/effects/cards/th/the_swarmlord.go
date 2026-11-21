package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Swarmlord — Legendary Creature — Tyranid {3}{G}{U}{R}, 5/5 (Warhammer
// 40,000 Commander, S82 GAP batch 193, #5948):
//
//	"Rapid Regeneration — The Swarmlord enters with two +1/+1 counters on it
//	 for each time you've cast your commander from the command zone this game.
//	 Xenos Cunning — Whenever a creature you control with a counter on it
//	 dies, draw a card."
//
// The entry counters are Prime Speaker Zegana's CR 614.1c clause counted once,
// as it enters, off the tally of commander casts from the command zone
// (Jyoti, Moag Ancient's reading). The dies trigger reads the counters of any
// kind the creature last had (CR 603.10a), the Swarmlord itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ead75ea4-025f-443a-824f-26fef9e424db",
		Name:         "The Swarmlord",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				return 2 * effects.SocB3CommanderCastsFromCommandZone(g, src.Controller)
			}, "The Swarmlord: enters with two +1/+1 counters for each time you've cast your commander from the command zone this game"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.GapB193ACreatureWithACounterYouControlDied,
				"The Swarmlord — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
