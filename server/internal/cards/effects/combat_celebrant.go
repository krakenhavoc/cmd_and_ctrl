package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Combat Celebrant — Creature — Human Warrior {2}{R}, 4/1:
//
//	"If this creature hasn't been exerted this turn, you may exert it
//	 as it attacks. When you do, untap all other creatures you control
//	 and after this phase, there is an additional combat phase. (An
//	 exerted creature won't untap during your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) unless this object has
// been exerted this turn (TurnTally.Exerts, per object, CR 400.7), with
// a linked "when you do" (CR 607.2h). The extra combat goes after the
// combat phase the trigger resolves in (CR 500.8). Several Celebrants
// exerted in one combat give that many extra combats, each Celebrant
// can be exerted only once a turn, and the phase is added even if
// Celebrant has died by then (the Amonkhet rulings).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "5e15ff93-99a0-4000-918e-4bd2c257188d",
		Name:          "Combat Celebrant",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacksUnless(ExertedThisTurn),
		Triggered: []game.TriggeredAbility{
			WhenExerted("Combat Celebrant — untap all other creatures you control; an additional combat phase after this phase",
				func(g *game.Game, item *game.StackItem) error {
					if err := untapAllOtherCreaturesYouControl(g, item); err != nil {
						return err
					}
					return ExtraCombatAfterThisPhase().Apply(NewContext(g, item))
				}),
		},
	})
}
