package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Careless Celebrant — Creature — Satyr Shaman {1}{R}, 2/1 (Theros Beyond Death, #5514):
//
//	"When this creature dies, it deals 2 damage to target creature or
//	 planeswalker an opponent controls."
//
// Bogardan Firefiend's dies trigger with this card's narrower clause.
// The Celebrant deals the damage from the graveyard as it last existed
// on the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56afad48-c5e5-4fbe-8889-15bcab234c31",
		Name:         "Careless Celebrant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisDies("Careless Celebrant — 2 damage to target creature or planeswalker an opponent controls",
					sourceDealsDamageToEachLegalTarget(2)),
				TargetPermanent("target creature or planeswalker an opponent controls",
					Or(Creature(), Planeswalker()), OpponentControls())),
				ForTargets(DamageToTarget(0, 2))),
		},
	})
}
