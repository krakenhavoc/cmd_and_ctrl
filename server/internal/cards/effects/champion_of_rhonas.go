package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Champion of Rhonas — Creature — Jackal Warrior {3}{G}, 3/3:
//
//	"You may exert this creature as it attacks. When you do, you may
//	 put a creature card from your hand onto the battlefield. (An
//	 exerted creature won't untap during your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h). The put is Quicksilver Amulet's: the creature
// card is put, not cast, so it triggers no cast trigger, and the "may"
// is a real decline. It enters untapped and not attacking (CR 508.4),
// before blockers.
//
// No simplification.
func init() {
	const label = "Champion of Rhonas — you may put a creature card from your hand onto the battlefield"
	Register(Spec{
		OracleID:      "517502e4-4588-4327-b2c0-451f4531fecc",
		Name:          "Champion of Rhonas",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			WhenExerted(label, func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    Creature(),
					Optional: true,
					Label:    label,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
