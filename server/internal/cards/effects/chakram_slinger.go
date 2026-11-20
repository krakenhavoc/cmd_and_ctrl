package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chakram Slinger — Creature — Human Warrior {4}{R}, 2/4 (Battlebond,
// #5529):
//
//	"Partner with Chakram Retriever (When this creature enters, target
//	 player may put Chakram Retriever into their hand from their library,
//	 then shuffle.)
//	 {R}, {T}: This creature deals 2 damage to target player or
//	 planeswalker."
//
// Not legendary, so it can't be a commander: of CR 702.124j's two
// abilities only the entry search does anything (PartnerWith). The ping is
// Fervent Paincaster's with a {R} beside the tap, so summoning sickness
// stops it the turn the Slinger arrives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a2568fd-4c22-413e-87e2-a384002ded38",
		Name:         "Chakram Slinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Chakram Slinger", "Chakram Retriever"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: This creature deals 2 damage to target player or planeswalker.",
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Targets: targetPlayerOrPlaneswalker(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
