package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vexing Puzzlebox accumulates the total of each dice instruction, adds
// mana and rolls a d20, and spends 100 charge counters to tutor an artifact.
func init() {
	Register(Spec{
		OracleID:     "7267ab16-2157-4b86-93ea-ca2c0fee064a",
		Name:         "Vexing Puzzlebox",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{WheneverYouRollDice("Vexing Puzzlebox — put charge counters on it", func(g *game.Game, item *game.StackItem, total int) error {
			if zone := g.FindCardZoneForEffect(item.SourceCardID); zone == nil || zone.Kind != game.ZoneBattlefield {
				return nil
			}
			return g.AddCounterForEffect(item.SourceCardID, "charge", total)
		})},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}",
			Label: "Add one mana of any color. Roll a d20.", IgnoreCommanderIdentity: true,
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				_, err := g.RollDiceForEffect(game.RandomDraw{Player: controller, Source: source}, 20, 1)
				return err
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "Remove 100 charge counters: search for an artifact",
			Cost:  Plus(TapCost(), RemoveCountersFromThis("charge", 100)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return (SearchLibrary{Player: item.Controller, Predicate: func(c game.Card) bool { return c.IsArtifact() }, Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, Reason: "Vexing Puzzlebox — an artifact"}).Apply(NewContext(g, item))
			},
		}},
	})
}
