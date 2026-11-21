package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "16923d37-bac0-4e6a-bd02-f3aa9360deeb",
		Name:         "Implement of Malice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{B}, Sacrifice this artifact: Target player discards a card. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1}),
			Cost:         effects.Plus(effects.ManaCost("{B}"), effects.SacrificeThis()),
			Targets:      effects.TargetPlayer("target player"),
			SorcerySpeed: true,
			Effect:       effects.OnTheTargetPlayer(aerImplementOfMaliceDiscard),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Implement of Malice — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// aerImplementOfMaliceDiscard makes the target player discard one card of
// their choice.
func aerImplementOfMaliceDiscard(ctx *effects.Context, p uuid.UUID) error {
	return effects.DiscardOne(ctx.Game, ctx.Item, p)
}
