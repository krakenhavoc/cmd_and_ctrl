package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trophy Hunter — Creature — Human Archer {2}{G}, 2/3 (Ravnica: City of
// Guilds, S82 OLD b582):
//
//	"{1}{G}: This creature deals 1 damage to target creature with flying.
//	 Whenever a creature with flying dealt damage by this creature this turn
//	 dies, put a +1/+1 counter on this creature."
//
// Blood Cultist's shape: the ping is a damage ability on one target, and the
// trigger is Sengir Vampire's, reading the turn's damage record for damage
// dealt by this object. The flying is the target's at the time it was
// damaged, which the ability's own target restriction enforces; the
// trigger asks only that the creature was dealt damage by the Hunter and
// had flying when it died, read from its last-known characteristics.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2c68c15-761e-4aea-89d8-f7fe8f44905c",
		Name:         "Trophy Hunter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: This creature deals 1 damage to target creature with flying.",
			Cost:    effects.ManaCost("{1}{G}"),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.DealDamageToTheTarget(1),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, died := effects.DiedCreature(ev, g)
				return died && ev.CardID != source.InstanceID && effects.LeftWithKeyword(ev, c, "flying") && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
			}, "Trophy Hunter — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
