package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang, A Lot to Learn — Legendary Creature — Human Avatar Ally {2}{G/W},
// 3/2 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Aang has vigilance as long as there's a Lesson card in your
//	 graveyard.
//	 Whenever another creature you control dies, put a +1/+1 counter on
//	 Aang."
//
// The vigilance is a layer 6 grant (CR 613.1f) re-read on every
// recompute, so it comes and goes with the graveyard. As the ruling says,
// gaining or losing it after Aang is declared as an attacker doesn't tap
// or untap him: vigilance matters only as attackers are declared
// (CR 702.20b). The dies trigger is Garna's condition
// (anotherCreatureYouControlDied) with a CounterOnThis body; each
// creature that dies is its own trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "854278bc-2bcf-45dd-abd5-6aca7a1e5fc9",
		Name:         "Aang, A Lot to Learn",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && avB4LessonsInGraveyard(g, source.Controller) > 0
			}, "vigilance"), "Aang has vigilance as long as there's a Lesson card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return anotherCreatureYouControlDied(ev, source, g)
			}, "Aang, A Lot to Learn — put a +1/+1 counter on Aang", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
