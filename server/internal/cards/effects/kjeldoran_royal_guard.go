package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kjeldoran Royal Guard — Creature — Human Soldier {3}{W}{W}, 2/5:
//
//	"{T}: All combat damage that would be dealt to you by unblocked
//	 creatures this turn is dealt to this creature instead."
//
// ADR 0108 §9, amendment 2026-10-16 (#2056): a redirection for the rest
// of the turn of combat damage to the ability's controller from a source
// that is an unblocked creature as the damage would be dealt (CR 509.1h,
// 609.7b): #2026's source filter, Combat: unblocked, beside the creature
// query. A blocked creature's damage, trample included, is not
// redirected. The damage goes to the Guard as the object it was when the
// ability resolved; once it has left the battlefield the redirection
// does nothing (CR 614.9, 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "8137e765-4df9-469f-a527-dee91d58fb7e",
		Name:         "Kjeldoran Royal Guard",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{T}: All combat damage that would be dealt to you by unblocked creatures this turn is dealt to this creature instead.",
				TapCost(), nil,
				RedirectDamage{
					Queries:    []game.PermanentQuery{QueryTypes("creature")},
					Filter:     game.DamageSourceFilter{Combat: game.SourceCombatUnblocked},
					Protect:    ShieldYou,
					CombatOnly: true,
					To:         RedirectToThis,
				}),
		},
	})
}
