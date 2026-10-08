package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oketra's Avenger — Creature — Human Warrior {1}{W}, 3/1:
//
//	"You may exert this creature as it attacks. When you do, prevent
//	 all combat damage that would be dealt to it this turn. (An exerted
//	 creature won't untap during your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h). The shield is ADR 0108's ModPreventFromSource,
// combat damage only and pinned to this object, so a creature that has
// left and come back is a new object and is not shielded (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8f064160-3afe-408a-85b4-b335eae8571c",
		Name:          "Oketra's Avenger",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				WhenExerted("Oketra's Avenger — prevent all combat damage that would be dealt to it this turn",
					func(g *game.Game, item *game.StackItem) error {
						return shieldThis(g, item, PreventDamageFromSource{CombatOnly: true})
					}),
				game.Purpose{PreventCombatDamageToSelf: true}),
		},
	})
}
