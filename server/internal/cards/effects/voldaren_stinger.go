package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Voldaren Stinger — Creature — Vampire Warrior {R}, 1/1 (Innistrad:
// Midnight Hunt, #5497):
//
//	"This creature has first strike as long as it's attacking.
//	 {2}{R}: This creature gets +2/+0 until end of turn."
//
// Soltari Lancer's attacking-only first strike with Yahenni-style pump
// from Aquatic Alchemist's BoostUntilEOT.
//
// No simplification.
func init() {
	firstStrike := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "first strike")
	firstStrike.Label = "This creature has first strike as long as it's attacking."
	firstStrike.DependsOnAttackingStatus = true
	Register(Spec{
		OracleID:     "5bd66394-6e1e-4b13-8824-e83768fb533e",
		Name:         "Voldaren Stinger",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{firstStrike},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Voldaren Stinger — +2/+0 until end of turn"}.Apply(NewContext(g, item))
			},
		}},
	})
}
