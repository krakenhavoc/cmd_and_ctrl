package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Cultist — {1}{B}{R} Creature — Human Wizard 1/1 (Shards of Alara, S82
// OLD b514):
//
//	"{T}: This creature deals 1 damage to target creature.
//	 Whenever a creature dealt damage by this creature this turn dies, put a
//	 +1/+1 counter on this creature."
//
// Sengir Vampire's trigger (damage dealt by this object this turn, read from
// the turn's damage record) beside a tap ability that pings.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "031ddf1e-afcf-4b09-8859-4df0f0ab5dcf",
		Name:         "Blood Cultist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to target creature.",
			Cost:    game.AbilityCost{Tap: true},
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Blood Cultist — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
