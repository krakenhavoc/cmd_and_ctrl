package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thaumaton Torpedo — Artifact {1} (Edge of Eternities, #4461):
//
//	"{6}, {T}, Sacrifice this artifact: Destroy target nonland permanent.
//	 This ability costs {3} less to activate if you attacked with a
//	 Spacecraft this turn."
//
// Unstable Obelisk's sacrifice-to-destroy, narrowed to a nonland
// permanent, with Flying Drone's own-row discount (CostModifiers, CR
// 601.2f via 602.2b). The condition reads the turn's attack tally
// (AttackedWithSubtypeThisTurn), which records each attacker's subtypes
// as it was declared, so a Spacecraft that attacked and has since left
// the battlefield still counts, and one an opponent attacked with does
// not. The target is re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f9ae015-dc59-4ecb-a2bb-668496b48a48",
		Name:         "Thaumaton Torpedo",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{6}, {T}, Sacrifice this artifact: Destroy target nonland permanent. This ability costs {3} less to activate if you attacked with a Spacecraft this turn.",
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{6}"},
			Targets: TargetPermanent("target nonland permanent", Not(Land())),
			CostModifiers: []game.CostModifier{
				ActivationCostsLess(3, "This ability costs {3} less to activate if you attacked with a Spacecraft this turn.",
					youAttackedWithASpacecraftThisTurn()),
			},
			Effect: destroyFirstLegalCardTarget,
		}},
	})
}

// youAttackedWithASpacecraftThisTurn passes while the activator declared
// a Spacecraft as an attacker this turn.
func youAttackedWithASpacecraftThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.AttackedWithSubtypeThisTurn(q.Controller, q.Source.TextSubtype("Spacecraft")) > 0
	}
}
