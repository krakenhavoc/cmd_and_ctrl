package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Judith, the Scourge Diva — Legendary Creature — Human Shaman {1}{B}{R},
// 2/2:
//
//	"Other creatures you control get +1/+0.
//	 Whenever a nontoken creature you control dies, Judith deals 1 damage
//	 to any target."
//
// The anthem is a Layer 7c static that leaves Judith herself out. The
// trigger looks back in time (CR 603.10a): the dead creature's controller
// and its token-ness are read as it last existed, and the creature may be
// Judith herself, in which case her damage is dealt by her last-known
// self (CR 608.2h). The target is chosen as the trigger goes on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c01516e0-eec3-4370-b935-7674982e8850",
		Name:         "Judith, the Scourge Diva",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					dead, ok := effects.DiedCreature(ev, g)
					return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && !effects.IsToken(dead)
				},
				Key:    "Judith, the Scourge Diva — 1 damage to any target",
				Effect: effects.SourceDealsDamageToEachLegalTarget(1),
			}, effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
