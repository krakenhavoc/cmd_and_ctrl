package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goro-Goro, Disciple of Ryusei — Legendary Creature — Goblin Samurai
// {1}{R}, 2/2 (EDHREC rank 2856):
//
//	"{R}: Creatures you control gain haste until end of turn.
//	 {3}{R}{R}: Create a 5/5 red Dragon Spirit creature token with
//	 flying. Activate only if you control an attacking modified
//	 creature. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// A Kamigawa: Neon Dynasty modified-matters commander. The haste is a
// one-shot grant to the creatures you control as it resolves (CR
// 611.2c), so a creature that arrives afterwards needs another {R}.
// The Dragon's "Activate only if you control an attacking modified
// creature" is its activation condition (CR 602.1b, #743): a creature
// you control that is currently declared as an attacker and is
// modified by Kodama of the West Tree's reading (b07IsModified).
// Neither ability taps Goro-Goro, so neither cares about summoning
// sickness.
//
// The Dragon ability works in the end of combat step, as printed:
// creatures leave combat as that step ENDS (CR 511.3, #785), so an
// attacking modified creature is still attacking while the step runs.
func init() {
	Register(Spec{
		OracleID:     "b80df711-af51-4f9a-9a3f-f7a4cddc9c2c",
		Name:         "Goro-Goro, Disciple of Ryusei",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{R}: Creatures you control gain haste until end of turn.",
				Cost:  ManaCost("{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"haste"},
						Label:    "Goro-Goro — haste",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:     "{3}{R}{R}: Create a 5/5 red Dragon Spirit creature token with flying. Activate only if you control an attacking modified creature.",
				Cost:      ManaCost("{3}{R}{R}"),
				Condition: goroGoroAttackingModifiedCreature,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("5/5 red Dragon Spirit with flying"), N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// goroGoroAttackingModifiedCreature is Goro-Goro's Dragon condition.
func goroGoroAttackingModifiedCreature(g *game.Game, controller, _ uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.AttackingTarget != uuid.Nil && b07IsModified(g, c) {
			return true
		}
	}
	return false
}
