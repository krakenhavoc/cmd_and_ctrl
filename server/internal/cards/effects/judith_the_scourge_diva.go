package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c01516e0-eec3-4370-b935-7674982e8850",
		Name:         "Judith, the Scourge Diva",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					dead, ok := diedCreature(ev, g)
					return ok && leftUnderControlOf(ev, dead) == source.Controller && !IsToken(dead)
				},
				Key:    "Judith, the Scourge Diva — 1 damage to any target",
				Effect: sourceDealsDamageToEachLegalTarget(1),
			}, TargetAny()), ForTargets(DamageToTarget(0, 1))),
		},
	})
}
