package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Netter en-Dal — Creature — Human Spellshaper {W}, 1/1
// (Nemesis, #5618):
//
//	"{W}, {T}, Discard a card: Target creature can't attack this turn."
//
// The spellshaper cost (nemSpellshaperCost) on Change of Heart's
// restriction, as an activated ability's body
// (nemTargetRestrictedThisTurn with game.CantAttack). It ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8cd65dcd-e36d-47b1-b011-961c79004cb0",
		Name:         "Netter en-Dal",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, {T}, Discard a card: Target creature can't attack this turn.",
			Cost:    nemSpellshaperCost("{W}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(HarmUntilEOT(0, 0, 0)),
			Effect:  nemTargetRestrictedThisTurn(game.CantAttack, "Netter en-Dal — can't attack this turn"),
		}},
	})
}
