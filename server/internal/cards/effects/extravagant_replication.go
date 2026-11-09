package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extravagant Replication — Enchantment {4}{U}{U} (Foundations, #4465):
//
//	"At the beginning of your upkeep, create a token that's a copy of
//	 another target nonland permanent you control."
//
// The trigger is a targeted upkeep trigger: "another" is object identity
// (Another), so the enchantment cannot copy itself, and the pick is
// re-read as the trigger resolves (CR 608.2b), a permanent that left in
// response making nothing. The token is a copy of the permanent's
// copiable values (CR 707.2) through CreateTokenCopy, so a copied
// creature's "when this enters" abilities trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a646245-b8b7-4f91-a312-d5eea9a9e49a",
		Name:         "Extravagant Replication",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourUpkeep("Extravagant Replication — create a token that's a copy of another target nonland permanent you control",
					extravagantReplicationCopy),
				Another(TargetPermanent("another target nonland permanent you control", Nonland(), YouControl())),
			),
		},
	})
}

// extravagantReplicationCopy makes the token copy of the still-legal
// target.
func extravagantReplicationCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	copyOf := FirstLegalBattlefieldTarget(ctx)
	if copyOf == uuid.Nil {
		return nil
	}
	return CreateTokenCopy{Controller: item.Controller, Copy: copyOf, N: 1}.Apply(ctx)
}
