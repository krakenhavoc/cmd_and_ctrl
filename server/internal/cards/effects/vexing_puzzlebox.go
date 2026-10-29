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
			if zone := g.FindCardZoneForEffect(item.SourceCardID); zone == nil || zone.Kind != game.ZoneBattlefield || sourceIsNewObject(g, item) { // #1432
				return nil
			}
			return g.AddCounterForEffect(item.SourceCardID, "charge", total)
		})},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}",
			Label: "Add one mana of any color. Roll a d20.",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				// #4307: through the one dice door, so a replaced roll asks
				// its chooser; nothing of this ability follows the roll.
				return g.RollDiceThenForEffect(game.DiceRoll{Player: controller, Source: source, Sides: 20, N: 1})
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove 100 charge counters from this artifact: Search your library for an artifact card, put that card onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), RemoveCountersFromThis("charge", 100)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return (SearchLibrary{Player: item.Controller, Predicate: func(c game.Card) bool { return c.IsArtifact() }, Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, Reason: "Vexing Puzzlebox — an artifact"}).Apply(NewContext(g, item))
			},
		}},
	})
}
