package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trueheart Twins — Creature — Jackal Warrior {4}{R}, 4/4:
//
//	"You may exert this creature as it attacks. (It won't untap during
//	 your next untap step.)
//	 Whenever you exert a creature, creatures you control get +1/+0
//	 until end of turn."
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with no linked
// trigger, and a "whenever you exert a creature" payoff that sees this
// creature and any other creature its controller exerts (the Amonkhet
// ruling). Each exert is its own trigger and its own +1/+0, and the set
// of creatures is locked as each one resolves (CR 611.2c).
//
// No simplification.
func init() {
	const label = "Trueheart Twins — creatures you control get +1/+0 until end of turn"
	Register(Spec{
		OracleID:      "5d3be0d2-0940-4904-9f12-de592c990bdb",
		Name:          "Trueheart Twins",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			WheneverYouExert(label, func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: label}.Apply(NewContext(g, item))
			}),
		},
	})
}
