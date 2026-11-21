package cardsco

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Covert Technician — Artifact Creature — Human Ninja {2}{U}, 2/4 (Neon Dynasty):
//
//	"Ninjutsu {1}{U}
//	 Whenever this creature deals combat damage to a player, you may put an
//	 artifact card with mana value less than or equal to that damage from
//	 your hand onto the battlefield."
//
// Broodcaller Scourge's optional hand put, narrowed to artifact cards; "that
// damage" is the amount of the damage event that fired the trigger (a card
// in hand with {X} in its cost counts X as 0).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b713ee3a-b6cd-4cd7-b403-f0bbc683ea99",
		Name:         "Covert Technician",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Covert Technician — you may put an artifact card with mana value less than or equal to that damage from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					dmg := item.Trigger.Event.Amount
					return effects.PutFromHandOntoBattlefield{
						Optional: true,
						Label:    "Covert Technician — put an artifact card with mana value " + strconv.Itoa(dmg) + " or less from your hand onto the battlefield",
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.IsArtifact() && c.ManaValue() <= dmg
						},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
