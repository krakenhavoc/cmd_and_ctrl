package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kessig Forgemaster // Flameheart Werewolf — Creature — Human Shaman
// Werewolf {1}{R}, 2/1 // Creature — Werewolf 3/2 (Shadows over Innistrad,
// S82 OLD):
//
//	Front: "Whenever this creature blocks or becomes blocked by a creature,
//	        this creature deals 1 damage to that creature.
//	        At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Whenever this creature blocks or becomes blocked by a creature,
//	        this creature deals 2 damage to that creature.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// Ornery Goblin's block-pair trigger (one per creature blocked or blocking it,
// CR 509.3b, 509.3d) and Villagers of Estwald's pair of werewolf triggers.
//
// No simplification.
func init() {
	const oracle = "b0d66546-c6a3-4052-b86e-98c3c2ab44b3"
	pair := func(label string, n int) game.TriggeredAbility {
		return effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
		}, label, effects.OldB287PokesFoe(n))
	}
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Kessig Forgemaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			pair("Kessig Forgemaster — 1 damage to that creature", 1),
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Kessig Forgemaster"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Flameheart Werewolf",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			pair("Flameheart Werewolf — 2 damage to that creature", 2),
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Flameheart Werewolf"),
		},
	})
}
