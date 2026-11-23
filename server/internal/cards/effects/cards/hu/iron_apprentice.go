package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Apprentice — Artifact Creature — Construct {1}, 0/0 (Kamigawa: Neon
// Dynasty):
//
//	"This creature enters with a +1/+1 counter on it.
//	 When this creature dies, if it had counters on it, put those counters on
//	 target creature you control."
//
// The entry counter is a CR 614.1c self-replacement (Afiya Grove's). The
// death trigger carries an intervening "if" (CR 603.4) read from the
// counters it had as it last existed (CR 603.10a, Buzzard-Wasp Colony's
// LastKnownCountersForEffect), and puts each kind on the target as it
// resolves, Essence Channeler's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd0c8db0-67ee-483b-8298-455cce35fef7",
		Name:         "Iron Apprentice",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 1, "Iron Apprentice enters with a +1/+1 counter on it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CardDied(ev, source) && len(g.LastKnownCountersForEffect(ev.CardID)) > 0
			}, "Iron Apprentice — if it had counters on it, put those counters on target creature you control",
				effects.PutItsCountersOnTargetCreature),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
