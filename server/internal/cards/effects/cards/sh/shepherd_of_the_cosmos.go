package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shepherd of the Cosmos — Creature — Angel Warrior {4}{W}{W}, 3/3 (Kaldheim):
//
//	"Flying
//	 When this creature enters, return target permanent card with mana value 2
//	 or less from your graveyard to the battlefield.
//	 Foretell {3}{W} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell cost.)"
//
// The target is a permanent card (not an instant or sorcery) in your own
// graveyard whose mana value is 2 or less, chosen as the trigger goes on the
// stack and returned under your control as it resolves (CR 608.2b re-check).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c5aa155-646c-47f6-ba5c-6e1b41fd4bb2",
		Name:            "Shepherd of the Cosmos",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{effects.Foretell("{3}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Shepherd of the Cosmos — return target permanent card with mana value 2 or less from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnFirstLegalGraveyardTargetToBattlefield(g, item)
				}),
				effects.TargetCardInGraveyard("target permanent card with mana value 2 or less in your graveyard", effects.YouOwn(), effects.ManaValueLE(2),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsPermanent() })),
		},
	})
}
