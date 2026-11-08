package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reigning Victor — Creature — Orc Warrior {2/R}{2/W}{2/B}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Mobilize 1 (Whenever this creature attacks, create a tapped and
//	 attacking 1/1 red Warrior creature token. Sacrifice it at the
//	 beginning of the next end step.)
//	 When this creature enters, target creature gets +1/+0 and gains
//	 indestructible until end of turn. (Damage and effects that say
//	 "destroy" don't destroy it.)"
//
// Mobilize is mobilize.go (CR 702.181a). The enters trigger targets any
// creature, the Victor itself included, and a target gone by resolution
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Reigning Victor — target creature gets +1/+0 and gains indestructible until end of turn"
	Register(Spec{
		OracleID:     "b8c5e481-371c-4999-bb2c-347a2aa33171",
		Name:         "Reigning Victor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Mobilize("Reigning Victor", 1),
			TriggerWithPurpose(Targeting(WhenThisEnters(label, trekB41PumpAndGrantEachTarget(1, 0, label, "indestructible")),
				TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 1, 0, "indestructible"))),
		},
	})
}
