package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Implement of Malice — Artifact {2} (Aether Revolt,
// #5536):
//
//	"{B}, Sacrifice this artifact: Target player discards a card. Activate
//	 only as a sorcery.
//	 When this artifact is put into a graveyard from the battlefield,
//	 draw a card."
//
// Dimir Guildmage's sorcery-speed discard row, paid with the
// Implement itself. The player chooses the card.
// The draw is Chromatic Star's leaves-for-the-graveyard trigger: it
// fires however the Implement got there, sacrificed to its own ability
// or destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16923d37-bac0-4e6a-bd02-f3aa9360deeb",
		Name:         "Implement of Malice",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{B}, Sacrifice this artifact: Target player discards a card. Activate only as a sorcery.",
			Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:         Plus(ManaCost("{B}"), SacrificeThis()),
			Targets:      TargetPlayer("target player"),
			SorcerySpeed: true,
			Effect:       onTheTargetPlayer(aerImplementOfMaliceDiscard),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Implement of Malice — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// aerImplementOfMaliceDiscard makes the target player discard one card of
// their choice.
func aerImplementOfMaliceDiscard(ctx *Context, p uuid.UUID) error {
	return discardOne(ctx.Game, ctx.Item, p)
}
