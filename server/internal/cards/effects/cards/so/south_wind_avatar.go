package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// South Wind Avatar — {3}{B} Creature — Snake Spirit Avatar 3/4 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Deathtouch
//	 Whenever another creature you control dies, you gain life equal to
//	 its toughness.
//	 Whenever you gain life, each opponent loses 1 life."
//
// The death trigger is Erebos, Bleak-Hearted's "another creature you control
// dies" with Conclave Mentor's GainLifeEqualToSubject: the toughness is the
// creature's as it last existed on the battlefield (CR 608.2h), anthems and
// counters included, and a toughness of zero or less gains nothing. The
// second trigger is Marauding Blight-Priest's, once per life-gain event, so
// the death trigger's own gain sets it off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "003e5dc6-8b26-4a0a-a50b-5f0806a7bacd",
		Name:            "South Wind Avatar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B39AnotherCreatureYouControlDied(ev, source, g)
			}, "South Wind Avatar — gain life equal to its toughness", effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectToughness})),
			effects.WheneverYouGainLife("South Wind Avatar — each opponent loses 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
