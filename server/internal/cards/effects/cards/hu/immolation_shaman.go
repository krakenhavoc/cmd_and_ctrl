package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Immolation Shaman — Creature — Lizard Shaman {1}{R}, 1/3 (Ravnica Allegiance):
//
//	"Whenever an opponent activates an ability of an artifact, creature, or land
//	 that isn't a mana ability, this creature deals 1 damage to that player.
//	 {3}{R}{R}: This creature gets +3/+3 and gains menace until end of turn."
//
// Harsh Mentor's trigger at one damage: it reads the activating player off the
// event as it resolves and does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf681c9e-a644-44c7-b744-34faafc67d50",
		Name:         "Immolation Shaman",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnOpponentActivates(
				"Immolation Shaman — 1 damage to that player",
				effects.Or(effects.Artifact(), effects.Creature(), effects.Land()),
				false,
				func(activator uuid.UUID) effects.Effect {
					return func(g *game.Game, item *game.StackItem) error {
						return effects.DealDamage{Source: item.SourceCardID, Target: activator, Amount: 1}.Apply(effects.NewContext(g, item))
					}
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{3}{R}{R}: This creature gets +3/+3 and gains menace until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump | game.AnswerCombatGrant,
				Pump:    &game.Pump{Power: 3, Toughness: 3, Keywords: []string{"menace"}},
			},
			Cost:   effects.ManaCost("{3}{R}{R}"),
			Effect: effects.ThisCreatureUntilEOT("Immolation Shaman — +3/+3 and menace until end of turn", 3, 3, "menace"),
		}},
	})
}
